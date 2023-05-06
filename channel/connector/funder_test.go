// SPDX-License-Identifier: Apache-2.0
package connector_test

import (
	"context"
	"fmt"
	"perun.network/perun-icp-backend/channel"
	chanconn "perun.network/perun-icp-backend/channel/connector"
	"perun.network/perun-icp-backend/channel/connector/test"
	chtest "perun.network/perun-icp-backend/channel/test"

	"perun.network/perun-icp-backend/utils"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"testing"
)

func TestNewDepositSetupDfxCLI(t *testing.T) {
	s := test.NewSetup(t)

	err := s.Setup.DfxSetup.StartDeployDfx()
	require.NoError(t, err, "Failed to start and deploy DFX environment")

	params, state := s.NewRandomParamAndState()
	dSetup := chtest.NewDepositSetup(params, state)

	err = chtest.FundAll(context.TODO(), s.Funders, dSetup.FReqs) //s.NewCtx()
	require.NoError(t, err)

	wReq, err := channel.NewDepositReqFromPerun(dSetup.FReqs[0], s.Funders[0].GetAcc())
	require.NoError(t, err)
	dReqFunding := wReq.Funding

	execPathTyped := chanconn.NewExecPath("./../../test/testdata/")

	recipPerunID, err := utils.DecodePrincipal("r7inp-6aaaa-aaaaa-aaabq-cai")
	if err != nil {
		panic(err)
	}

	dfxState, err := chanconn.NewState(state)
	sigs := s.SignState(dfxState)

	require.NoError(t, err)

	var nonceArray [32]byte
	copy(nonceArray[:], params.Nonce.Bytes())
	statefinal := state.IsFinal
	alloc := state.Allocation
	chanId := dReqFunding.Channel
	adj := s.Adjs[0] // testing the adjudicator of the first perun client

	outpConclude, err := adj.ConcludeDfxCLI(nonceArray, params.Parts, params.ChallengeDuration, chanId, state.Version, &alloc, statefinal, sigs, *recipPerunID, execPathTyped)

	require.NoError(t, err)
	fmt.Println("outp: ", outpConclude)
	outpWithdraw, err := chanconn.Withdraw(dReqFunding, sigs[0], *recipPerunID, execPathTyped)
	fmt.Println("outp: ", outpWithdraw)
	require.NoError(t, err)
	fmt.Println("outp: ", outpWithdraw)
	// Check the on-chain balance.
	//s.AssertDeposits(dSetup.FIDs, dSetup.FinalBals)

	err = s.Setup.DfxSetup.StopDFX()
	assert.NoError(t, err, "Failed to stop DFX environment")
}

func TestNewDepositSetupAgentGO(t *testing.T) {
	s := test.NewSetup(t)

	err := s.Setup.DfxSetup.StartDeployDfx()
	require.NoError(t, err, "Failed to start and deploy DFX environment")

	params, state := s.NewRandomParamAndState()
	dSetup := chtest.NewDepositSetup(params, state)

	err = chtest.FundAll(context.TODO(), s.Funders, dSetup.FReqs) //s.NewCtx()
	require.NoError(t, err)

	wReq, err := channel.NewDepositReqFromPerun(dSetup.FReqs[0], s.Funders[0].GetAcc())
	require.NoError(t, err)
	dReqFunding := wReq.Funding

	execPathTyped := chanconn.NewExecPath("./../../test/testdata/")

	recipPerunID, err := utils.DecodePrincipal("r7inp-6aaaa-aaaaa-aaabq-cai")
	if err != nil {
		panic(err)
	}

	dfxState, err := chanconn.NewState(state)
	sigs := s.SignState(dfxState)

	require.NoError(t, err)

	var nonceArray [32]byte
	copy(nonceArray[:], params.Nonce.Bytes())
	statefinal := state.IsFinal
	alloc := state.Allocation
	chanId := dReqFunding.Channel
	adj := s.Adjs[0] // testing the adjudicator of the first perun client

	outpConclude, err := adj.ConcludeAgentGo(nonceArray, params.Parts, params.ChallengeDuration, chanId, state.Version, &alloc, statefinal, sigs, *recipPerunID)

	require.NoError(t, err)
	fmt.Println("outp: ", outpConclude)
	outpWithdraw, err := chanconn.Withdraw(dReqFunding, sigs[0], *recipPerunID, execPathTyped)
	require.NoError(t, err)
	fmt.Println("outp: ", outpWithdraw)
	// Check the on-chain balance.
	//s.AssertDeposits(dSetup.FIDs, dSetup.FinalBals)

	err = s.Setup.DfxSetup.StopDFX()
	assert.NoError(t, err, "Failed to stop DFX environment")
}
