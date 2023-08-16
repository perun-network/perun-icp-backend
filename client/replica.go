// Copyright 2023 - See NOTICE file for copyright holders.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package client

import (
	"github.com/aviate-labs/agent-go/ic/icpledger"
	"github.com/aviate-labs/agent-go/principal"
	"math/big"
)

func (p *PaymentClient) GetOwnBalance() *big.Int {

	l1Principal := p.ICConn.ICAgent.Sender()
	l1AccountId := l1Principal.AccountIdentifier(principal.DefaultSubAccount)
	balance, err := p.ICConn.LedgerAgent.AccountBalance(icpledger.AccountBalanceArgs{Account: l1AccountId.Bytes()})
	if err != nil {
		panic(err)
	}
	bal := balance.E8s
	return new(big.Int).SetUint64(bal)
}

func (p *PaymentClient) GetExtBalance(extPrince principal.Principal) uint64 {

	l1AccountId := extPrince.AccountIdentifier(principal.DefaultSubAccount)
	balance, err := p.ICConn.LedgerAgent.AccountBalance(icpledger.AccountBalanceArgs{Account: l1AccountId.Bytes()})
	if err != nil {
		panic(err)
	}
	return uint64(balance.E8s)
}
