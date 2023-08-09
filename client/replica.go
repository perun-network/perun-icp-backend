package client

import (
	"github.com/aviate-labs/agent-go/ic/icpledger"
	"github.com/aviate-labs/agent-go/principal"
	"math/big"
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
