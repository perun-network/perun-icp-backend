package client

import (
	"fmt"
	"github.com/pkg/errors"
	"math/big"
	"perun.network/go-perun/client"
	"perun.network/go-perun/watcher/local"
	"perun.network/go-perun/wire/net/simple"

	"perun.network/perun-icp-backend/channel"
	chanconn "perun.network/perun-icp-backend/channel/connector"

	"perun.network/perun-icp-backend/wallet"
)

func SetupPaymentClient(
	name string,
	w *wallet.FsWallet, // w is the wallet used to resolve addresses to accounts for channels.
	sharedComm *SharedComm,
	perunID string,
	ledgerID string,
	host string,
	port int,
	accountPath string,
) (*PaymentClient, error) {

	acc := w.NewAccount()

	// Connect to Perun pallet and get funder + adjudicator from it.
	perunConn := chanconn.NewDfxConnector(perunID, ledgerID, accountPath, host, port)
	perunConn.Mutex = sharedComm.mutex
	bus := sharedComm.bus

	funder := channel.NewFunder(acc, perunConn)
	adj := channel.NewAdjudicator(acc, perunConn)

	// Setup dispute watcher.
	watcher, err := local.NewWatcher(adj)
	if err != nil {
		return nil, fmt.Errorf("intializing watcher: %w", err)
	}

	// Setup Perun client.
	wireAddr := simple.NewAddress(acc.Address().String())
	perunClient, err := client.New(wireAddr, bus, funder, adj, w, watcher)
	if err != nil {
		return nil, errors.WithMessage(err, "creating client")
	}

	// Create client and start request handler.
	c := &PaymentClient{
		Name:        name,
		perunClient: perunClient,
		account:     &acc,
		currency:    channel.Asset,
		channels:    make(chan *PaymentChannel, 1),
		dfxConn:     perunConn,
		wAddr:       wireAddr,
		balance:     big.NewInt(0),
	}

	go c.PollBalances()
	go perunClient.Handle(c, c)
	return c, nil
}
