package client

import (
	"github.com/aviate-labs/agent-go/ic/icpledger"
	"github.com/aviate-labs/agent-go/principal"
	"math/big"
	//"perun.network/perun-icp-backend/setup"
)

func (p *PaymentClient) GetOwnBalance() *big.Int {

	l1Principal := p.dfxConn.DfxAgent.Sender()
	l1AccountId := l1Principal.AccountIdentifier(principal.DefaultSubAccount)
	balance, err := p.dfxConn.LedgerAgent.AccountBalance(icpledger.AccountBalanceArgs{Account: l1AccountId.Bytes()})
	if err != nil {
		panic(err)
	}
	bal := balance.E8s
	return new(big.Int).SetUint64(bal)
}

func (p *PaymentClient) GetExtBalance(extPrince principal.Principal) uint64 {

	l1AccountId := extPrince.AccountIdentifier(principal.DefaultSubAccount)
	balance, err := p.dfxConn.LedgerAgent.AccountBalance(icpledger.AccountBalanceArgs{Account: l1AccountId.Bytes()})
	if err != nil {
		panic(err)
	}
	return uint64(balance.E8s)
}

// const transferFee = 10000

// func NewReplica() *setup.DfxSetup {

// 	demoConfig := setup.DfxConfig{
// 		Host:        "http://127.0.0.1",
// 		Port:        4943,
// 		ExecPath:    "./test/testdata/",
// 		AccountPath: "./test/testdata/identities/minter_identity.pem",
// 	}

// 	dfx := setup.NewDfxSetup(demoConfig)

// 	return dfx
// }
