// SPDX-License-Identifier: Apache-2.0
package connector_test

import (
	"fmt"
	"github.com/aviate-labs/agent-go"
	"math/big"
	"net/url"

	"github.com/aviate-labs/agent-go/ic/icpledger"

	"github.com/stretchr/testify/require"

	"github.com/aviate-labs/agent-go/principal"

	"perun.network/perun-icp-backend/channel"
	chanconn "perun.network/perun-icp-backend/channel/connector"

	"perun.network/perun-icp-backend/channel/connector/test"
	chtest "perun.network/perun-icp-backend/channel/test"
	"perun.network/perun-icp-backend/setup"

	"testing"
)

func TestLedgerAgent(t *testing.T) {
	ledgerTestConfig := setup.DfxConfig{
		Host:     "http://127.0.0.1",
		Port:     4943,
		ExecPath: "./../../test/testdata/",
	}

	ledgerPrincipal := "bkyz2-fmaaa-aaaaa-qaaaq-cai"

	ledgerId, err := principal.Decode(ledgerPrincipal)
	if err != nil {
		t.Fatalf("Failed to decode principal: %v", err)
	}

	id, err := chanconn.NewIdentity("./../../test/testdata/identities/usera_identity.pem")
	if err != nil {
		t.Fatalf("Failed to create new identity: %v", err)
	}

	idPrince := (*id).Sender()
	ic0, err := url.Parse(fmt.Sprintf("%s:%d", ledgerTestConfig.Host, ledgerTestConfig.Port))
	if err != nil {
		t.Fatalf("Failed to create ic0: %v", err)
	}
	ldgConfig := agent.Config{
		Identity:     *id,
		ClientConfig: &agent.ClientConfig{Host: ic0},
		FetchRootKey: true,
	}

	ldgAgent, err := icpledger.NewAgent(ledgerId, ldgConfig)
	if err != nil {
		t.Fatalf("Failed to create ledger agent: %v", err)
	}

	accID := idPrince.AccountIdentifier(principal.DefaultSubAccount)

	resp, err := ldgAgent.AccountBalance(icpledger.AccountBalanceArgs{Account: accID.Bytes()})
	if err != nil {
		t.Fatalf("Failed to get account balance: %v", err)
	}
	fmt.Println("resp: ", resp.E8s)
	perunPrincipal := "be2us-64aaa-aaaaa-qaabq-cai"
	perunID, err := principal.Decode(perunPrincipal)
	require.NoError(t, err, "Failed to decode principal")
	perunaccountID := perunID.AccountIdentifier(principal.DefaultSubAccount)
	toAccount := perunaccountID.Bytes()

	txArgs := icpledger.TransferArgs{
		Memo: 2,
		Amount: struct {
			E8s uint64 "ic:\"e8s\""
		}{E8s: 50000},
		Fee: struct {
			E8s uint64 "ic:\"e8s\""
		}{E8s: 10000},
		To: toAccount,
	}

	txRes, err := ldgAgent.Transfer(txArgs)
	if err != nil {
		t.Fatalf("Failed to transfer: %v", err)
	}

	if txRes.Err != nil {
		if txRes.Err.BadFee != nil {
			fmt.Printf("BadFee error. Expected Fee: %v\n", txRes.Err.BadFee.ExpectedFee)
		} else if txRes.Err.InsufficientFunds != nil {
			fmt.Printf("InsufficientFunds error. Balance: %v\n", txRes.Err.InsufficientFunds.Balance)
		} else if txRes.Err.TxTooOld != nil {
			fmt.Printf("TxTooOld error. Allowed Window Nanos: %v\n", txRes.Err.TxTooOld.AllowedWindowNanos)
		} else if txRes.Err.TxCreatedInFuture != nil {
			fmt.Println("TxCreatedInFuture error.")
		} else if txRes.Err.TxDuplicate != nil {
			fmt.Printf("TxDuplicate error. Duplicate Of: %v\n", txRes.Err.TxDuplicate.DuplicateOf)
		} else {
			fmt.Println("Unknown error")
		}
	} else if txRes.Ok != nil {
		fmt.Println("BlockIndex: ", *txRes.Ok)
	} else {
		fmt.Println("Both BlockIndex and TransferError are nil")
	}
	// Print entire txRes for debugging
	fmt.Printf("txRes: %+v\n", txRes)
	respP, err := ldgAgent.AccountBalance(icpledger.AccountBalanceArgs{Account: accID.Bytes()})
	if err != nil {
		t.Fatalf("Failed to get account balance: %v", err)
	}
	fmt.Println("respP: ", respP.E8s)

	resp2, err := ldgAgent.AccountBalance(icpledger.AccountBalanceArgs{Account: toAccount})
	if err != nil {
		t.Fatalf("Failed to get account balance: %v", err)
	}
	fmt.Println("resp2: ", resp2.E8s)

}

func TestQueryPerun(t *testing.T) {

	perunID := "be2us-64aaa-aaaaa-qaabq-cai"
	err := channel.QueryCandidCLI("()", perunID, "./../../test/testdata/")
	require.NoError(t, err, "Failed to query Perun ID")

}

func TestDeposit(t *testing.T) {
	s := test.NewPerunSetup(t)

	params, state := s.NewRandomParamAndState()

	fmt.Println("params: ", params, "state: ", state)

	state.Allocation.Balances[0][0] = big.NewInt(200000)
	state.Allocation.Balances[0][1] = big.NewInt(200000)

	dSetup := chtest.NewDepositSetup(params, state)

	err := chtest.FundConc(s.NewCtx(), s.Funders, dSetup.FReqs)
	require.NoError(t, err)
}

// func TestValidateSig(t *testing.T) {
// 	s := test.NewPerunSetup(t)

// 	params, state := s.NewRandomParamAndState()
// 	dSetup := chtest.NewDepositSetup(params, state)
// 	dfxState, err := chanconn.StateForChain(state)
// 	if err != nil {
// 		panic(err)
// 	}
// 	wReq, err := channel.NewDepositReqFromPerun(dSetup.FReqs[0], s.Funders[0].GetAcc())
// 	require.NoError(t, err)
// 	dReqFunding := wReq.Funding
// 	sigs := s.SignState(dfxState)
// 	alloc := state.Allocation

// 	var nonceArray [32]byte
// 	copy(nonceArray[:], params.Nonce.Bytes())
// 	statefinal := state.IsFinal
// 	chanId := dReqFunding.Channel

// 	sigOut, err := s.Deps[0].VerifySig(nonceArray, params.Parts, params.ChallengeDuration, chanId, state.Version, &alloc, statefinal, sigs)
// 	require.NoError(t, err)

// 	log.Printf("Result of signature verification: %s", sigOut)
// }
