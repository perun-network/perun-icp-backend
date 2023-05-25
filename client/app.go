package client

import (
	"fmt"
	"github.com/pkg/errors"
	"perun.network/go-perun/client"
	"perun.network/go-perun/watcher/local"
	"perun.network/go-perun/wire"
	icwire "perun.network/perun-icp-backend/wire"

	icchannel "perun.network/perun-icp-backend/channel"
	chanconn "perun.network/perun-icp-backend/channel/connector"
	icwallet "perun.network/perun-icp-backend/wallet"
	"sync"

	"perun.network/perun-icp-backend/wallet"
)

func SetupPaymentClient(
	bus wire.Bus, // bus is used of off-chain communication.
	w *wallet.FsWallet, // w is the wallet used to resolve addresses to accounts for channels.
	mtx *sync.Mutex,
	perunID string,
	ledgerID string,
	host string, // networkId is the identifier of the blockchain.
	port int,
	accountPath string,
	execPath string,
	//perun *chanconn.Connector,
) (*PaymentClient, error) {

	acc := w.NewAccount()

	// Connect to Perun pallet and get funder + adjudicator from it.
	perun := chanconn.NewConnector(perunID, ledgerID, accountPath, execPath, host, port)
	perun.Mutex = mtx
	fmt.Println("perun.Mutex: ", perun.Mutex, &perun.Mutex)

	funder := icchannel.NewFunder(acc, perun)
	adj := icchannel.NewAdjudicator(acc, perun)

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
		currency:    icchannel.Asset,
		channels:    make(chan *PaymentChannel, 1),
	}

	go perunClient.Handle(c, c)
	return c, nil
}
