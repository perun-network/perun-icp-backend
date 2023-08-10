// SPDX-License-Identifier: Apache-2.0

package client

import (
	"context"
	"fmt"
	chanconn "github.com/perun-network/perun-icp-backend/channel/connector"
	"github.com/perun-network/perun-icp-backend/channel/connector/icperun"

	vc "perun.network/perun-demo-tui/client"
	"sync"

	pchannel "perun.network/go-perun/channel"
	"perun.network/go-perun/client"
	"perun.network/go-perun/wire"

	"math/big"
	//"perun.network/go-perun/wallet"
	"github.com/perun-network/perun-icp-backend/channel"
	icwallet "github.com/perun-network/perun-icp-backend/wallet"
	//icwire "github.com/perun-network/perun-icp-backend/wire"
)

type SharedComm struct {
	bus   wire.Bus
	mutex *sync.Mutex
}

// PaymentClient is a payment channel client.
type PaymentClient struct {
	perunClient   *client.Client       // The core Perun client.
	account       *icwallet.Account    // The account we use for on-chain and off-chain transactions.
	currency      pchannel.Asset       // The currency we expect to get paid in.
	channels      chan *PaymentChannel // Accepted payment channels.
	Channel       *PaymentChannel      // The current payment channel.
	dfxConn       *chanconn.Connector  // The connector to the Dfx blockchain
	observerMutex sync.Mutex
	observers     []vc.Observer
	balanceMutex  sync.Mutex
	Name          string
	wAddr         wire.Address
	balance       *big.Int
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

	participants := []wire.Address{c.WireAddress(), peer}

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
		c.account.Address(),
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
	c.Channel.ch.OnUpdate(c.NotifyAllState)
	c.NotifyAllState(nil, ch.State())

}

func (p *PaymentClient) WireAddress() wire.Address {
	return p.wAddr
}

// AcceptedChannel returns the next accepted channel.
func (c *PaymentClient) AcceptedChannel() *PaymentChannel {
	c.Channel = <-c.channels
	c.Channel.ch.OnUpdate(c.NotifyAllState)
	c.NotifyAllState(nil, c.Channel.ch.State())
	return c.Channel

}

// Shutdown gracefully shuts down the client.
func (c *PaymentClient) Shutdown() {
	c.perunClient.Close()
}

func (c *PaymentClient) GetChannelBalance() (*big.Int, error) {
	if c.Channel == nil {
		return big.NewInt(0), nil
	}

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
		return big.NewInt(0), nil
	}

	return (*balNat).BigInt(), nil
}

func (p *PaymentClient) Deregister(observer vc.Observer) {
	p.observerMutex.Lock()
	defer p.observerMutex.Unlock()
	for i, o := range p.observers {
		if o.GetID().String() == observer.GetID().String() {
			p.observers[i] = p.observers[len(p.observers)-1]
			p.observers = p.observers[:len(p.observers)-1]
		}

	}
}

func (p *PaymentClient) DisplayAddress() string {
	addr := p.account.Address().String()

	return addr
}

func (p *PaymentClient) DisplayName() string {
	return p.Name
}

func (p *PaymentClient) HasOpenChannel() bool {
	return p.Channel != nil
}

func (p *PaymentClient) NotifyAllBalance(bal int64) {
	str := FormatBalance(new(big.Int).SetInt64(bal))
	for _, o := range p.observers {
		o.UpdateBalance(str)
	}
}

func (p *PaymentClient) NotifyAllState(from, to *pchannel.State) {
	p.observerMutex.Lock()
	defer p.observerMutex.Unlock()
	str := FormatState(p.Channel, to)
	for _, o := range p.observers {
		o.UpdateState(str)
	}
}

func (p *PaymentClient) Register(observer vc.Observer) {
	p.observerMutex.Lock()
	defer p.observerMutex.Unlock()
	p.observers = append(p.observers, observer)
	if p.Channel != nil {
		observer.UpdateState(FormatState(p.Channel, p.Channel.GetChannelState()))
	}
	observer.UpdateBalance(FormatBalance(p.GetOwnBalance()))
}
