package client

import (
	"fmt"
	"github.com/aviate-labs/agent-go/ic/icpledger"
	"github.com/aviate-labs/agent-go/principal"
	"perun.network/perun-icp-backend/setup"
)

// sdrA := userA.Agent.Sender()
// sdrB := userB.Agent.Sender()

// accidA := sdrA.AccountIdentifier(principal.DefaultSubAccount)
// accidB := sdrB.AccountIdentifier(principal.DefaultSubAccount)

// blncA, err := userA.Ledger.AccountBalance(icpledger.AccountBalanceArgs{Account: accidA.Bytes()})
// if err != nil {
// 	panic(err)
// }
// blncB, err := userA.Ledger.AccountBalance(icpledger.AccountBalanceArgs{Account: accidB.Bytes()})
// if err != nil {
// 	panic(err)
// }

func (p *PaymentClient) GetOwnBalance() uint64 {

	l1Principal := p.dfxConn.DfxAgent.Sender()
	l1AccountId := l1Principal.AccountIdentifier(principal.DefaultSubAccount)
	balance, err := p.dfxConn.LedgerAgent.AccountBalance(icpledger.AccountBalanceArgs{Account: l1AccountId.Bytes()})
	if err != nil {
		panic(err)
	}
	fmt.Println("balance: ", balance.E8s)
	return uint64(balance.E8s)
}

func (p *PaymentClient) GetExtBalance(extPrince principal.Principal) uint64 {

	//l1Principal := p.dfxConn.DfxAgent.Sender()
	l1AccountId := extPrince.AccountIdentifier(principal.DefaultSubAccount)
	balance, err := p.dfxConn.LedgerAgent.AccountBalance(icpledger.AccountBalanceArgs{Account: l1AccountId.Bytes()})
	if err != nil {
		panic(err)
	}
	fmt.Println("balance: ", balance.E8s)
	return uint64(balance.E8s)
}

const transferFee = 10000

func NewReplica() *setup.DfxSetup {

	demoConfig := setup.DfxConfig{
		Host:        "http://127.0.0.1",
		Port:        4943,
		ExecPath:    "./test/testdata/",
		AccountPath: "./test/testdata/identities/minter_identity.pem",
	}

	dfx := setup.NewDfxSetup(demoConfig)

	return dfx
}
