package client

import (
	"fmt"
	"github.com/pkg/errors"
	"perun.network/go-perun/client"
	"perun.network/go-perun/watcher/local"
	icwire "perun.network/perun-icp-backend/wire"

	"perun.network/perun-icp-backend/channel"
	chanconn "perun.network/perun-icp-backend/channel/connector"
	icwallet "perun.network/perun-icp-backend/wallet"

	"perun.network/perun-icp-backend/wallet"
)

func SetupPaymentClient(
	w *wallet.FsWallet, // w is the wallet used to resolve addresses to accounts for channels.
	sharedComm *SharedComm,
	perunID string,
	ledgerID string,
	host string, // networkId is the identifier of the blockchain.
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
	waddr := icwallet.AsAddr(acc.Address())
	wireaddr := &icwire.Address{Address: waddr}

	perunClient, err := client.New(wireaddr, bus, funder, adj, w, watcher)
	if err != nil {
		return nil, errors.WithMessage(err, "creating client")
	}

	// Create client and start request handler.
	c := &PaymentClient{
		perunClient: perunClient,
		account:     waddr,
		currency:    channel.Asset,
		channels:    make(chan *PaymentChannel, 1),
		dfxConn:     perunConn,
	}

	go perunClient.Handle(c, c)
	return c, nil
}
