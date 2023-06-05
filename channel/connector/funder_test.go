// SPDX-License-Identifier: Apache-2.0
package connector_test

import (
	"fmt"
	"github.com/aviate-labs/agent-go"

	"log"
	"net/url"

	"github.com/aviate-labs/agent-go/ic/icpledger"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	//"github.com/aviate-labs/agent-go/identity"

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

	ledgerPrincipal := "rrkah-fqaaa-aaaaa-aaaaq-cai"

	ledgerId, err := principal.Decode(ledgerPrincipal)
	if err != nil {
		t.Fatalf("Failed to decode principal: %v", err)
	}

	dfx := setup.NewDfxSetup(ledgerTestConfig)

	err = dfx.StartDeployDfx()
	if err != nil {
		t.Fatalf("Failed to start Dfx: %v", err)
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
	fromSubaccount := accID.Bytes()
	perunPrincipal := "r7inp-6aaaa-aaaaa-aaabq-cai"
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
		FromSubaccount: &fromSubaccount,
		To:             toAccount,
	}

	txRes, err := ldgAgent.Transfer(txArgs)

	if err != nil {
		t.Fatalf("Failed to transfer: %v", err)
	}
	fmt.Println("txRes: ", txRes)

	err = dfx.StopDFX()
	if err != nil {
		t.Fatalf("Failed to stop Dfx: %v", err)
	}
}

func TestQueryPerun(t *testing.T) {
	s := test.NewPerunSetup(t)
	err := s.Setup.DfxSetup.StartDeployDfx()
	require.NoError(t, err, "Failed to start and deploy DFX environment")

	defer func() {
		err := s.Setup.DfxSetup.StopDFX()
		assert.NoError(t, err, "Failed to stop DFX environment")
	}()

	perunID := "r7inp-6aaaa-aaaaa-aaabq-cai"
	err = channel.QueryCandidCLI("()", perunID, "./../../test/testdata/")
	require.NoError(t, err, "Failed to query Perun ID")

}

func TestDepositCLI(t *testing.T) {
	s := test.NewPerunSetup(t)

	err := s.Setup.DfxSetup.StartDeployDfx()
	require.NoError(t, err, "Failed to start and deploy DFX environment")

	defer func() {
		err := s.Setup.DfxSetup.StopDFX()
		assert.NoError(t, err, "Failed to stop DFX environment")
	}()

	params, state := s.NewRandomParamAndState()
	dSetup := chtest.NewDepositSetup(params, state)

	err = chtest.FundMtx(s.NewCtx(), s.Funders, dSetup.FReqs)
	require.NoError(t, err)
}

func TestDepositAG(t *testing.T) {
	s := test.NewPerunSetup(t)

	err := s.Setup.DfxSetup.StartDeployDfx()
	require.NoError(t, err, "Failed to start and deploy DFX environment")

	defer func() {
		err := s.Setup.DfxSetup.StopDFX()
		assert.NoError(t, err, "Failed to stop DFX environment")
	}()

	params, state := s.NewRandomParamAndState()
	dSetup := chtest.NewDepositSetup(params, state)

	err = chtest.FundAllAG(s.NewCtx(), s.Funders, dSetup.FReqs)
	require.NoError(t, err)
}

func TestValidateSig(t *testing.T) {
	s := test.NewPerunSetup(t)

	err := s.Setup.DfxSetup.StartDeployDfx()
	require.NoError(t, err, "Failed to start and deploy DFX environment")
	defer func() {
		err := s.Setup.DfxSetup.StopDFX()
		assert.NoError(t, err, "Failed to stop DFX environment")
	}()

	params, state := s.NewRandomParamAndState()
	dSetup := chtest.NewDepositSetup(params, state)
	dfxState, err := chanconn.NewState(state)
	if err != nil {
		panic(err)
	}
	wReq, err := channel.NewDepositReqFromPerun(dSetup.FReqs[0], s.Funders[0].GetAcc())
	require.NoError(t, err)
	dReqFunding := wReq.Funding
	sigs := s.SignState(dfxState)
	alloc := state.Allocation

	var nonceArray [32]byte
	copy(nonceArray[:], params.Nonce.Bytes())
	statefinal := state.IsFinal
	chanId := dReqFunding.Channel

	sigOut, err := s.Deps[0].VerifySig(nonceArray, params.Parts, params.ChallengeDuration, chanId, state.Version, &alloc, statefinal, sigs)
	require.NoError(t, err)

	log.Printf("Result of signature verification: %s", sigOut)
}
