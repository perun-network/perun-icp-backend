// SPDX-License-Identifier: Apache-2.0
package connector_test

import (
	"log"
	"net/url"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/aviate-labs/agent-go/identity"

	"github.com/aviate-labs/agent-go/principal"

	"github.com/aviate-labs/agent-go/ledger"

	"perun.network/perun-icp-backend/channel"
	chanconn "perun.network/perun-icp-backend/channel/connector"
	"perun.network/perun-icp-backend/channel/connector/test"
	chtest "perun.network/perun-icp-backend/channel/test"
	"perun.network/perun-icp-backend/setup"

	"testing"
)

func TestLedgerAgent(t *testing.T) {
	ledgerTestConfig := setup.DfxConfig{
		Host:        "http://127.0.0.1",
		Port:        4943,
		ExecPath:    "./../../test/testdata/",
		AccountPath: "./test/testdata/identities/minter_identity.pem",
	}

	ledgerPrincipal := "rrkah-fqaaa-aaaaa-aaaaq-cai"

	canId, err := principal.Decode(ledgerPrincipal)
	if err != nil {
		t.Fatalf("Failed to decode principal: %v", err)
	}

	url, err := url.Parse("http://127.0.0.1:4943")
	if err != nil {
		t.Fatalf("Failed to parse URL: %v", err)
	}

	dfx := setup.NewDfxSetup(ledgerTestConfig)

	err = dfx.StartDeployDfx()
	if err != nil {
		t.Fatalf("Failed to start Dfx: %v", err)
	}

	id, err := identity.NewRandomSecp256k1Identity()
	if err != nil {
		t.Fatalf("Failed to create new identity: %v", err)
	}

	a := ledger.NewWithIdentity(canId, url, id)
	prID, err := principal.Decode("exqrz-uemtb-qnd6t-mvbn7-mxjre-bodlr-jnqql-tnaxm-ur6uc-mmgb4-jqe")
	if err != nil {
		t.Fatalf("Failed to decode principal: %v", err)
	}
	accID := prID.AccountIdentifier(principal.DefaultSubAccount)
	tokens, err := a.AccountBalance(ledger.AccountBalanceArgs{
		Account: accID,
	})
	if err != nil {
		t.Fatalf("Failed to get account balance: %v", err)
	}

	t.Logf("check tokens: %v", tokens)

	err = dfx.StopDFX()
	if err != nil {
		t.Fatalf("Failed to stop Dfx: %v", err)
	}
}

func TestQueryPerun(t *testing.T) {
	s := test.NewSetup(t)
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
	s := test.NewSetup(t)

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
	s := test.NewSetup(t)

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
	s := test.NewSetup(t)

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
