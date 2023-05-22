// SPDX-License-Identifier: Apache-2.0

package client

import (
	"context"
	"crypto/rand"
	"fmt"
	"github.com/aviate-labs/agent-go"
	"github.com/aviate-labs/agent-go/identity"
	"github.com/aviate-labs/agent-go/ledger"
	"github.com/aviate-labs/agent-go/principal"
	"net/url"
	"os"
	"path/filepath"
	"perun.network/go-perun/channel"
	"perun.network/go-perun/client"
	"perun.network/go-perun/wire"

	"math/big"
	"perun.network/go-perun/wallet"
	icchannel "perun.network/perun-icp-backend/channel"
	"perun.network/perun-icp-backend/setup"
	"perun.network/perun-icp-backend/utils"
	icwallet "perun.network/perun-icp-backend/wallet"
	icwire "perun.network/perun-icp-backend/wire"
)

type PerunUser struct {
	L2Account wallet.Account
	Agent     *agent.Agent
	Ledger    *ledger.Agent
}

// PaymentClient is a payment channel client.
type PaymentClient struct {
	perunClient *client.Client       // The core Perun client.
	account     wallet.Address       // The account we use for on-chain and off-chain transactions.
	currency    channel.Asset        // The currency we expect to get paid in.
	channels    chan *PaymentChannel // Accepted payment channels.
}

func (u *PerunUser) NewL2Account() (wallet.Account, error) {
	wlt, err := icwallet.NewRAMWallet(rand.Reader)
	if err != nil {
		return nil, err
	}
	acc := wlt.NewAccount()

	return acc, nil
}

func MakeLedger(accountPath, host string, canisterId principal.Principal) (*ledger.Agent, error) {
	data, err := os.ReadFile(accountPath)
	if err != nil {
		return nil, err
	}

	var agentID identity.Identity
	agentID, err = identity.NewSecp256k1IdentityFromPEM(data)
	if err != nil {
		return nil, err
	}

	hostURL, err := url.Parse(host)
	if err != nil {
		return nil, fmt.Errorf("error parsing host URL: %v", err)
	}

	a := ledger.NewWithIdentity(canisterId, hostURL, agentID)

	return &a, nil
}

func NewUserConfig(balance uint64, pemAccountName, host string, port int) (setup.UserConfig, error) {
	return setup.UserConfig{
		Host:        host,
		Port:        port,
		Balance:     balance,
		AccountPath: filepath.Join(utils.SetHomeDir(), ".config", "dfx", "identity", pemAccountName, "identity.pem"),
	}, nil
}

func NewPerunUser(config setup.UserConfig, ledgerAddr string) (*PerunUser, error) {

	ledgerPrincipal, err := principal.Decode(ledgerAddr)
	if err != nil {
		return nil, err
	}

	agent, err := NewUserAgent(config)
	if err != nil {
		return nil, err
	}
	perunUser := &PerunUser{
		Agent: agent,
	}

	perunUser.Ledger, err = MakeLedger(config.AccountPath, config.Host, ledgerPrincipal)
	if err != nil {
		return nil, err
	}

	perunUser.L2Account, err = perunUser.NewL2Account()
	if err != nil {
		return nil, err
	}

	return perunUser, nil
}

func NewUserAgent(config setup.UserConfig) (*agent.Agent, error) {
	data, err := os.ReadFile(config.AccountPath)
	if err != nil {
		return nil, err
	}
	var agentID identity.Identity
	agentID, err = identity.NewSecp256k1IdentityFromPEM(data)
	if err != nil {
		return nil, err
	}
	ic0, err := url.Parse(fmt.Sprintf("%s:%d", config.Host, config.Port))
	if err != nil {
		return nil, err
	}

	agent := agent.New(agent.Config{
		Identity: agentID,
		ClientConfig: &agent.ClientConfig{
			Host: ic0,
		}})

	return &agent, nil
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
func (c *PaymentClient) OpenChannel(peer wire.Address, amount float64) *PaymentChannel {
	// We define the channel participants. The proposer has always index 0. Here
	// we use the on-chain addresses as off-chain addresses, but we could also
	// use different ones.
	waddr := *icwallet.AsAddr(c.account)
	wireaddr := &icwire.Address{Address: &waddr}

	participants := []wire.Address{wireaddr, peer}

	// We create an initial allocation which defines the starting balances.
	initBal := big.NewInt(int64(amount))

	initAlloc := channel.NewAllocation(2, icchannel.Asset)
	initAlloc.SetAssetBalances(icchannel.Asset, []channel.Bal{
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

	return newPaymentChannel(ch, c.currency)
}

// WireAddress returns the wire address of the client.
func (c *PaymentClient) WireAddress() *icwire.Address {
	waddr := icwallet.AsAddr(c.account)
	return &icwire.Address{Address: waddr}
}
