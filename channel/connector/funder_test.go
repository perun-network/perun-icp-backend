// SPDX-License-Identifier: Apache-2.0
package connector_test

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"log"
	"perun.network/perun-icp-backend/channel"
	chanconn "perun.network/perun-icp-backend/channel/connector"
	"perun.network/perun-icp-backend/channel/connector/test"
	chtest "perun.network/perun-icp-backend/channel/test"

	"testing"
)

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

func TestDeposit(t *testing.T) {
	s := test.NewSetup(t)

	err := s.Setup.DfxSetup.StartDeployDfx()
	require.NoError(t, err, "Failed to start and deploy DFX environment")

	defer func() {
		err := s.Setup.DfxSetup.StopDFX()
		assert.NoError(t, err, "Failed to stop DFX environment")
	}()

	params, state := s.NewRandomParamAndState()
	dSetup := chtest.NewDepositSetup(params, state)

	err = chtest.FundAll(s.NewCtx(), s.Funders, dSetup.FReqs)
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
