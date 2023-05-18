// SPDX-License-Identifier: Apache-2.0

package connector_test

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"log"
	"math"
	pchannel "perun.network/go-perun/channel"
	pchtest "perun.network/go-perun/channel/test"
	pwallet "perun.network/go-perun/wallet"
	"perun.network/perun-icp-backend/channel"
	chanconn "perun.network/perun-icp-backend/channel/connector"
	"perun.network/perun-icp-backend/utils"
	"testing"

	"perun.network/perun-icp-backend/channel/connector/test"
	chtest "perun.network/perun-icp-backend/channel/test"
)

func TestConcludeDfxCLI(t *testing.T) {
	s := test.NewSetup(t)

	err := s.Setup.DfxSetup.StartDeployDfx()
	require.NoError(t, err, "Failed to start and deploy DFX environment")
	defer func() {
		err := s.Setup.DfxSetup.StopDFX()
		assert.NoError(t, err, "Failed to stop DFX environment")
	}()

	params, state := s.NewRandomParamAndState()
	dSetup := chtest.NewDepositSetup(params, state)

	err = chtest.FundAll(context.Background(), s.Funders, dSetup.FReqs)
	require.NoError(t, err)

	wReq, err := channel.NewDepositReqFromPerun(dSetup.FReqs[0], s.Funders[0].GetAcc())
	require.NoError(t, err)
	dReqFunding := wReq.Funding

	dfxState, err := chanconn.NewState(state)
	sigs := s.SignState(dfxState)

	require.NoError(t, err)

	var nonceArray [32]byte
	copy(nonceArray[:], params.Nonce.Bytes())
	statefinal := state.IsFinal
	alloc := state.Allocation
	chanId := dReqFunding.Channel
	adj := s.Adjs[0]

	outpConclude, err := adj.ConcludeDfxCLI(nonceArray, params.Parts, params.ChallengeDuration, chanId, state.Version, &alloc, statefinal, sigs)
	if err != nil {
		log.Fatalf("Failed to conclude via DFX CLI: %v", err)
	}

	assert.Equal(t, "(opt \"successful concluding the channel\")\n", outpConclude)

	fmt.Println("Concluded channel via DFX CLI: ", outpConclude)

	err = s.Setup.DfxSetup.StopDFX()
	assert.NoError(t, err, "Failed to stop DFX environment")
}

func TestConcludeAgentGO(t *testing.T) {
	s := test.NewSetup(t)

	err := s.Setup.DfxSetup.StartDeployDfx()
	require.NoError(t, err, "Failed to start and deploy DFX environment")
	defer func() {
		err := s.Setup.DfxSetup.StopDFX()
		assert.NoError(t, err, "Failed to stop DFX environment")
	}()

	params, state := s.NewRandomParamAndState()
	dSetup := chtest.NewDepositSetup(params, state)

	err = chtest.FundAll(context.Background(), s.Funders, dSetup.FReqs)
	require.NoError(t, err)

	wReq, err := channel.NewDepositReqFromPerun(dSetup.FReqs[0], s.Funders[0].GetAcc())
	require.NoError(t, err)
	dReqFunding := wReq.Funding

	dfxState, err := chanconn.NewState(state)
	sigs := s.SignState(dfxState)

	require.NoError(t, err)

	var nonceArray [32]byte
	copy(nonceArray[:], params.Nonce.Bytes())
	statefinal := state.IsFinal
	alloc := state.Allocation
	chanId := dReqFunding.Channel
	adj := s.Adjs[0]

	outpConclude, err := adj.ConcludeAgentGo(nonceArray, params.Parts, params.ChallengeDuration, chanId, state.Version, &alloc, statefinal, sigs)
	require.NoError(t, err)
	assert.Equal(t, "(opt \"successful concluding\")\n", outpConclude)
}

func TestConcludeWithdraw(t *testing.T) {
	s := test.NewSetup(t)

	err := s.Setup.DfxSetup.StartDeployDfx()
	require.NoError(t, err, "Failed to start and deploy DFX environment")
	defer func() {
		err := s.Setup.DfxSetup.StopDFX()
		assert.NoError(t, err, "Failed to stop DFX environment")
	}()

	params, state := s.NewRandomParamAndState()
	dSetup := chtest.NewDepositSetup(params, state)

	err = chtest.FundAll(context.Background(), s.Funders, dSetup.FReqs)
	require.NoError(t, err)

	wReq, err := channel.NewDepositReqFromPerun(dSetup.FReqs[0], s.Funders[0].GetAcc())
	require.NoError(t, err)
	dReqFunding := wReq.Funding

	dfxState, err := chanconn.NewState(state)
	sigs := s.SignState(dfxState)

	require.NoError(t, err)

	var nonceArray [32]byte
	copy(nonceArray[:], params.Nonce.Bytes())
	statefinal := state.IsFinal
	alloc := state.Allocation
	chanId := dReqFunding.Channel
	adj := s.Adjs[0]

	outpConclude, err := adj.ConcludeDfxCLI(nonceArray, params.Parts, params.ChallengeDuration, chanId, state.Version, &alloc, statefinal, sigs)
	if err != nil {
		log.Fatalf("Failed to conclude via DFX CLI: %v", err)
	}

	assert.Equal(t, "(opt \"successful concluding the channel\")\n", outpConclude)
	fmt.Println("outp: ", outpConclude)

	execPathTyped := chanconn.NewExecPath("./../../test/testdata/")

	recipPerunID, err := utils.DecodePrincipal("r7inp-6aaaa-aaaaa-aaabq-cai")
	if err != nil {
		panic(err)
	}

	// test withdrawal to usera with principal: exqrz-uemtb-qnd6t-mvbn7-mxjre-bodlr-jnqql-tnaxm-ur6uc-mmgb4-jqe
	require.NoError(t, err)
	outpWithdraw, err := chanconn.Withdraw(dReqFunding, sigs[0], *recipPerunID, execPathTyped)
	fmt.Println("outp: ", outpWithdraw)
	require.NoError(t, err)
	// Check the on-chain balance.
	//s.AssertDeposits(dSetup.FIDs, dSetup.FinalBals)

	err = s.Setup.DfxSetup.StopDFX()
	assert.NoError(t, err, "Failed to stop DFX environment")
}

func TestDispute(t *testing.T) {
	s := test.NewSetup(t)

	err := s.Setup.DfxSetup.StartDeployDfx()
	require.NoError(t, err, "Failed to start and deploy DFX environment")
	defer func() {
		err := s.Setup.DfxSetup.StopDFX()
		assert.NoError(t, err, "Failed to stop DFX environment")
	}()

	params, state := s.NewRandomParamAndState()
	dSetup := chtest.NewDepositSetup(params, state)

	err = chtest.FundAll(context.Background(), s.Funders, dSetup.FReqs)
	require.NoError(t, err)

	wReq, err := channel.NewDepositReqFromPerun(dSetup.FReqs[0], s.Funders[0].GetAcc())
	require.NoError(t, err)
	dReqFunding := wReq.Funding

	dfxState, err := chanconn.NewState(state)
	sigs := s.SignState(dfxState)

	require.NoError(t, err)

	var nonceArray [32]byte
	copy(nonceArray[:], params.Nonce.Bytes())
	statefinal := state.IsFinal
	alloc := state.Allocation
	chanId := dReqFunding.Channel
	adj := s.Adjs[0]

	outpDispute, err := adj.Dispute(nonceArray, params.Parts, params.ChallengeDuration, chanId, state.Version, &alloc, statefinal, sigs)
	if err != nil {
		log.Fatalf("Failed to dispute: %v", err)
	}
	assert.Equal(t, "(opt \"successful initialization of a dispute\")\n", outpDispute)
	fmt.Println("outp: ", outpDispute)

	err = s.Setup.DfxSetup.StopDFX()
	assert.NoError(t, err, "Failed to stop DFX environment")
}

func newAdjReq(s *test.Setup, final bool) (pchannel.AdjudicatorReq, *pchannel.Params, *pchannel.State) {
	var state *pchannel.State
	// make sure that Version is within int64 range
	for {
		state = pchtest.NewRandomState(s.Rng, chtest.DefaultRandomOpts())
		if state.Version <= uint64(math.MaxInt64) {
			break
		} else {
			fmt.Println("Version is not in uint64 range: ", state.Version)
		}
	}
	state.IsFinal = final
	var data [20]byte
	s.Rng.Read(data[:])
	nonce := pchannel.NonceFromBytes(data[:])
	params, err := pchannel.NewParams(60, []pwallet.Address{s.Alice.Address(), s.Bob.Address()}, pchannel.NoApp(), nonce, true, false)
	require.NoError(s.T, err)
	state.ID = params.ID()
	wState, err := chanconn.NewState(state)
	require.NoError(s.T, err)
	sigs := s.SignState(wState)
	req := pchannel.AdjudicatorReq{
		Params:    params,
		Acc:       s.Alice,
		Tx:        pchannel.Transaction{State: state, Sigs: sigs},
		Idx:       0,
		Secondary: false,
	}
	return req, params, state
}
