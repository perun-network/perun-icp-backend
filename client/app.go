package client

import (
	"fmt"
	"github.com/pkg/errors"
	"perun.network/go-perun/client"
	"perun.network/go-perun/watcher/local"
	"perun.network/go-perun/wire"
	icwire "perun.network/perun-icp-backend/wire"

	icchannel "perun.network/perun-icp-backend/channel"
	icwallet "perun.network/perun-icp-backend/wallet"

	chanconn "perun.network/perun-icp-backend/channel/connector"

	"perun.network/perun-icp-backend/wallet"
)

func SetupPaymentClient(
	bus wire.Bus, // bus is used of off-chain communication.
	w *wallet.FsWallet, // w is the wallet used to resolve addresses to accounts for channels.
	//acc wallet.Account, // acc is the account to be used for signing transactions.
	//nodeURL string, // nodeURL is the URL of the blockchain node.
	perunID string,
	ledgerID string,
	host string, // networkId is the identifier of the blockchain.
	port int,
	accountPath string,
	execPath string,
	//queryDepth types.BlockNumber, // queryDepth is the number of blocks being evaluated when looking for events.
) (*PaymentClient, error) {
	// Connect to backend.
	// api, err := dot.NewAPI(nodeURL, networkId)
	// if err != nil {
	// 	panic(err)
	// }

	acc := w.NewAccount()

	// Connect to Perun pallet and get funder + adjudicator from it.
	perun := chanconn.NewConnector(perunID, ledgerID, accountPath, execPath, host, port)
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

// func SetupPaymentClient(
// 	bus wire.Bus,
// 	nodeURL string,
// 	//networkId dot.NetworkID,
// 	//queryDepth types.BlockNumber,
// 	perunAcc wallet.Account,
// ) (*PaymentClient, error) {
// 	// Create wallet and account.

// 	// Create and start client.
// 	c, err := client.SetupPaymentClient(
// 		bus,
// 		perunAcc,
// 	)
// 	if err != nil {
// 		panic(err)
// 	}

// 	return c
// }
