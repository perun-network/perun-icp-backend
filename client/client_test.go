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

package client_test

import (
	"fmt"
	"github.com/aviate-labs/agent-go/ic/icpledger"
	"github.com/aviate-labs/agent-go/principal"
	"github.com/stretchr/testify/require"
	"math/rand"
	"path/filepath"
	chanconn "perun.network/perun-icp-backend/channel/connector"
	"testing"
	"time"
)

const (
	ledgerID = "bkyz2-fmaaa-aaaaa-qaaaq-cai"
	perunID  = "be2us-64aaa-aaaaa-qaabq-cai"
	Host     = "http://127.0.0.1"
	Port     = 4943
)

func TestPrincipalTransfers(t *testing.T) {
	s := SimpleTxSetup(t)

	userBalancesPreTx := make([]uint64, len(s.L1Users))

	perunBalancePreTx, err := s.PerunNode.GetBalance()
	require.NoError(t, err, "Failed to get balance")

	for i := 0; i < len(s.L1Users); i++ {
		bal, err := s.L1Users[i].GetBalance()
		require.NoError(t, err, "Failed to get balance")

		userBalancesPreTx[i] = bal
	}

	rand.Seed(time.Now().UnixNano())

	txBalances := make([]uint64, len(s.L1Users))
	for i := range txBalances {
		txBalances[i] = uint64(rand.Intn(4001) + 1000)
	}

	perunPrincipal, err := principal.Decode(perunID)
	require.NoError(t, err, "Failed to decode principal")
	ledgerPrincipal, err := principal.Decode(ledgerID)
	require.NoError(t, err, "Failed to decode principal")

	perunaccountID := perunPrincipal.AccountIdentifier(principal.DefaultSubAccount)
	txArgsList := make([]icpledger.TransferArgs, len(s.L1Users))

	for i := 0; i < len(s.L1Users); i++ {
		toAccount := perunaccountID.Bytes()
		txArgsList[i] = createTransferArgs(i, toAccount, txBalances[i])

		_, err := s.L1Users[i].TransferIC(txArgsList[i], ledgerPrincipal)
		require.NoError(t, err, "Failed to transfer")
		_, err = s.L1Users[i].GetBalance()
		require.NoError(t, err, "Failed to get balance")

	}

	balDiffr := uint64(0)

	for i := 0; i < len(s.L1Users); i++ {
		bal, err := s.L1Users[i].GetBalance()
		require.NoError(t, err, "Failed to get balance")
		balDiffr += bal - userBalancesPreTx[i]
	}

	perunBalancePostTx, err := s.PerunNode.GetBalance()
	require.NoError(t, err, "Failed to get balance")
	require.NoError(t, err, perunBalancePostTx-perunBalancePreTx, balDiffr)

}

func NewL1User(prince *principal.Principal, c *chanconn.Connector) *L1User {
	return &L1User{prince, c}
}

type L1Setup struct {
	T           *testing.T
	Accs        []*principal.Principal
	MinterAcc   *principal.Principal
	PerunPrince *principal.Principal
	Conns       []*chanconn.Connector
	ConnPerun   *chanconn.Connector
}

type OnChainSetup struct {
	*L1Setup
	L1Users   []*L1User
	PerunNode *L1User
}

type OnChainBareSetup struct {
	*L1Setup
	L1Users   []*L1User
	PerunNode *L1User
}

type L1User struct {
	Prince *principal.Principal
	Conn   *chanconn.Connector
}

func (u *L1User) GetBalance() (uint64, error) {

	accountID := u.Prince.AccountIdentifier(principal.DefaultSubAccount)
	ledgerAgent := u.Conn.LedgerAgent
	onChainBal, err := ledgerAgent.AccountBalance(icpledger.AccountBalanceArgs{Account: accountID.Bytes()})
	if err != nil {
		return 0, fmt.Errorf("failed to get balance: %v", err)
	}

	return onChainBal.E8s, nil
}

func (u *L1User) TransferIC(txArgs icpledger.TransferArgs, canID principal.Principal) (uint64, error) {
	ldg := u.Conn.LedgerAgent

	transferResult, err := ldg.Transfer(txArgs)
	if err != nil {
		return 0, fmt.Errorf("Transfer method in TransferIC failed: %v", err)
	}

	if transferResult.Err != nil {
		err := chanconn.HandleTransferError(transferResult.Err)
		if err != nil {
			return 0, err
		}
	}

	blnm := transferResult.Ok
	if blnm == nil {
		return 0, fmt.Errorf("blockNum is nil")
	}

	return *blnm, nil
}
func createTransferArgs(i int, toAccount []byte, txBalance uint64) icpledger.TransferArgs {
	return icpledger.TransferArgs{
		Memo: uint64(i),
		Amount: struct {
			E8s uint64 "ic:\"e8s\""
		}{E8s: txBalance},
		Fee: struct {
			E8s uint64 "ic:\"e8s\""
		}{E8s: chanconn.ICTransferFee},
		To: toAccount,
	}
}

func SimpleTxSetup(t *testing.T) *OnChainBareSetup {

	s := TransferSetup(t)
	c := s.Conns
	cp := s.ConnPerun
	pP := s.PerunPrince

	ret := &OnChainBareSetup{L1Setup: s}

	for i := 0; i < len(s.Accs); i++ {
		dep := NewL1User(s.Accs[i], c[i])
		ret.L1Users = append(ret.L1Users, dep)
	}
	pnode := NewL1User(pP, cp)

	ret.PerunNode = pnode
	return ret
}

func TransferSetup(t *testing.T) *L1Setup {

	basePath := chanconn.GetBasePath()

	aliceAccPath := filepath.Join(basePath, "./../userdata/identities/usera_identity.pem")
	bobAccPath := filepath.Join(basePath, "./../userdata/identities/userb_identity.pem")
	minterAccPath := filepath.Join(basePath, "./../userdata/identities/minter_identity.pem")

	aliceAcc, err := chanconn.NewIdentity(aliceAccPath)
	if err != nil {
		panic(err)
	}
	bobAcc, err := chanconn.NewIdentity(bobAccPath)
	if err != nil {
		panic(err)
	}

	minterAcc, err := chanconn.NewIdentity(minterAccPath)
	if err != nil {
		panic(err)
	}

	alicePrince := (*aliceAcc).Sender()
	bobPrince := (*bobAcc).Sender()
	minterPrince := (*minterAcc).Sender()

	perunPrince, err := principal.Decode(perunID)
	if err != nil {
		panic(err)
	}

	accs := []*principal.Principal{&alicePrince, &bobPrince}
	conn1 := chanconn.NewICConnector(perunID, ledgerID, aliceAccPath, Host, Port)
	conn2 := chanconn.NewICConnector(perunID, ledgerID, bobAccPath, Host, Port)
	connPerun := chanconn.NewICConnector(perunID, ledgerID, minterAccPath, Host, Port)

	conns := []*chanconn.Connector{conn1, conn2}
	return &L1Setup{t, accs, &minterPrince, &perunPrince, conns, connPerun}
}
