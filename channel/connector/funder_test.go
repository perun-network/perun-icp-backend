// SPDX-License-Identifier: Apache-2.0
package connector_test

import (
	"fmt"
	"github.com/aviate-labs/agent-go/ic/icpledger"
	"math/big"

	"github.com/stretchr/testify/require"

	"github.com/aviate-labs/agent-go/principal"

	chanconn "github.com/perun-network/perun-icp-backend/channel/connector"
	"github.com/perun-network/perun-icp-backend/channel/connector/test"

	chtest "github.com/perun-network/perun-icp-backend/channel/test"

	"testing"
)

func TestTransferToLedger(t *testing.T) {
	tSetup := chtest.NewTestSetup(t)

	aliceLedger := tSetup.DfxConns[0].LedgerAgent
	aliceL1ID := tSetup.L1Accs[0]

	accID := aliceL1ID.AccountIdentifier(principal.DefaultSubAccount)

	_, err := aliceLedger.AccountBalance(icpledger.AccountBalanceArgs{Account: accID.Bytes()})
	require.NoError(t, err, "Failed to get account balance")

	perunID := tSetup.DfxConns[0].PerunID
	perunaccountID := perunID.AccountIdentifier(principal.DefaultSubAccount)
	toAccount := perunaccountID.Bytes()

	txArgs := icpledger.TransferArgs{
		Memo: 1,
		Amount: struct {
			E8s uint64 "ic:\"e8s\""
		}{E8s: 50000},
		Fee: struct {
			E8s uint64 "ic:\"e8s\""
		}{E8s: 10000},
		To: toAccount,
	}

	txRes, err := aliceLedger.Transfer(txArgs)
	if err != nil {
		t.Fatalf("Failed to transfer: %v", err)
	}

	if txRes.Err != nil {
		t.Fatalf("Transfer failed with error: %v", chanconn.HandleTransferError(txRes.Err))
	} else if txRes.Ok != nil {
		fmt.Println("BlockIndex: ", *txRes.Ok)
	} else {
		fmt.Println("Both BlockIndex and TransferError are nil")
	}

	_, err = aliceLedger.AccountBalance(icpledger.AccountBalanceArgs{Account: accID.Bytes()})
	require.NoError(t, err, "Failed to get account balance")

	_, err = aliceLedger.AccountBalance(icpledger.AccountBalanceArgs{Account: toAccount})
	require.NoError(t, err, "Failed to get account balance")
}

func TestDeposit(t *testing.T) {
	s := test.NewPerunSetup(t)

	chanAlloc := uint64(50000)

	params, state := s.NewRandomParamAndState()
	state.Allocation.Balances[0][0] = new(big.Int).SetUint64(chanAlloc)
	state.Allocation.Balances[0][1] = new(big.Int).SetUint64(chanAlloc)

	funderaddr1 := s.Funders[0].GetAcc().L2Address()
	funderaddr2 := s.Funders[1].GetAcc().L2Address()

	params.Parts[0] = &funderaddr1
	params.Parts[1] = &funderaddr2

	dSetup := chtest.NewDepositSetup(params, state)
	balsPrev, err := s.GetL1Balances()
	require.NoError(t, err)
	perunBalsPrev, err := s.GetPerunBalances()
	require.NoError(t, err)

	err = chtest.FundAll(s.NewCtx(), s.Funders, dSetup.FReqs)

	require.NoError(t, err)
	balsPost, err := s.GetL1Balances()
	require.NoError(t, err)
	perunBalsPost, err := s.GetPerunBalances()
	require.NoError(t, err)

	// check if balances have arrived exactly

	cid := params.ID()
	chanBals, err := s.GetChannelBalances(cid)
	require.NoError(t, err)

	allocDiffrL1 := chanAlloc + 2*chanconn.DfxTransferFee
	allocDiffrChan := chanAlloc + chanconn.DfxTransferFee

	for i, el := range chanBals {
		// check that channel balances are exactly allocation plus withdrawal fees
		require.Equal(t, allocDiffrChan, el)
		// check that L1 balances pre funding are exactly L1 balances post funding plus funding and withdrawal fees
		require.Equal(t, balsPrev[i], balsPost[i]+allocDiffrL1)
		//check that balances present in the channel are exactly allocation plus withdrawal fees
		require.Equal(t, chanBals[i], allocDiffrChan)
		// check that balances in the perun canister is exactly the allocation from both users
		require.Equal(t, perunBalsPost[i]-perunBalsPrev[i], chanBals[0]+chanBals[1])
	}

}
