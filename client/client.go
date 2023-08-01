// SPDX-License-Identifier: Apache-2.0

package client

import (
	"context"
	"fmt"
	chanconn "perun.network/perun-icp-backend/channel/connector"
	"perun.network/perun-icp-backend/channel/connector/icperun"

	"sync"

	pchannel "perun.network/go-perun/channel"
	"perun.network/go-perun/client"
	"perun.network/go-perun/wire"

	"math/big"
	"perun.network/go-perun/wallet"
	"perun.network/perun-icp-backend/channel"
	icwallet "perun.network/perun-icp-backend/wallet"
	icwire "perun.network/perun-icp-backend/wire"
)

type SharedComm struct {
	bus   wire.Bus
	mutex *sync.Mutex
}

// PaymentClient is a payment channel client.
type PaymentClient struct {
	perunClient *client.Client       // The core Perun client.
	account     wallet.Address       // The account we use for on-chain and off-chain transactions.
	currency    pchannel.Asset       // The currency we expect to get paid in.
	channels    chan *PaymentChannel // Accepted payment channels.
	Channel     *PaymentChannel      // The current payment channel.
	dfxConn     *chanconn.Connector  // The connector to the Dfx blockchain
}

func InitSharedComm() *SharedComm {
	bus := wire.NewLocalBus()
	mutex := &sync.Mutex{}
	return &SharedComm{
		bus:   bus,
		mutex: mutex,
	}
}

// startWatching starts the dispute watcher for the specified channel.
func (c *PaymentClient) startWatching(ch *client.Channel) {
	go func() {
		err := ch.Watch(c)
		if err != nil {
			fmt.Printf("Watcher returned with error: %v", err)
		}
	}()
}

// OpenChannel opens a new channel with the specified peer and funding.
func (c *PaymentClient) OpenChannel(peer wire.Address, amount float64) { //*PaymentChannel
	// We define the channel participants. The proposer has always index 0. Here
	// we use the on-chain addresses as off-chain addresses, but we could also
	// use different ones.
	waddr := *icwallet.AsAddr(c.account)
	wireaddr := &icwire.Address{Address: &waddr}

	participants := []wire.Address{wireaddr, peer}

	// We create an initial allocation which defines the starting balances.
	initBal := big.NewInt(int64(amount))

	initAlloc := pchannel.NewAllocation(2, channel.Asset)
	initAlloc.SetAssetBalances(channel.Asset, []pchannel.Bal{
		initBal, // Our initial balance.
		initBal, // Peer's initial balance.
	})

	// Prepare the channel proposal by defining the channel parameters.
	challengeDuration := uint64(10) // On-chain challenge duration in seconds.
	proposal, err := client.NewLedgerChannelProposal(
		challengeDuration,
		c.account,
		initAlloc,
		participants,
	)
	if err != nil {
		panic(err)
	}

	// Send the proposal.
	ch, err := c.perunClient.ProposeChannel(context.TODO(), proposal)
	if err != nil {
		panic(err)
	}

	// Start the on-chain event watcher. It automatically handles disputes.
	c.startWatching(ch)
	c.Channel = newPaymentChannel(ch, c.currency)

}

// WireAddress returns the wire address of the client.
func (c *PaymentClient) WireAddress() *icwire.Address {
	waddr := icwallet.AsAddr(c.account)
	return &icwire.Address{Address: waddr}
}

// AcceptedChannel returns the next accepted channel.
func (c *PaymentClient) AcceptedChannel() { //*PaymentChannel
	c.Channel = <-c.channels
}

// Shutdown gracefully shuts down the client.
func (c *PaymentClient) Shutdown() {
	c.perunClient.Close()
}

func (c *PaymentClient) GetChannelBalance() *big.Int {
	chanParams := c.Channel.GetChannelParams()
	cid := chanParams.ID()
	addr := chanParams.Parts[c.Channel.ch.Idx()]
	addrBytes, err := addr.MarshalBinary()
	if err != nil {
		panic(err)
	}

	queryBalArgs := icperun.Funding{
		Channel:     cid,
		Participant: addrBytes,
	}
	balNat, err := c.dfxConn.PerunAgent.QueryHoldings(queryBalArgs)
	if err != nil {
		panic(err)
	}

	if (*balNat) == nil {
		return big.NewInt(0)
	}

	return (*balNat).BigInt()
}
