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

package connector_test

import (
	"github.com/stretchr/testify/require"
	"log"
	"math/big"
	pchannel "perun.network/go-perun/channel"
	pchtest "perun.network/go-perun/channel/test"
	pwallet "perun.network/go-perun/wallet"
	"perun.network/perun-icp-backend/channel"
	"testing"

	"perun.network/perun-icp-backend/channel/connector/test"

	chtest "perun.network/perun-icp-backend/channel/test"
)

func TestAdjudicator_Register(t *testing.T) {
	s := test.NewPerunSetup(t)
	chanAlloc := uint64(50000)
	userIdx := 0

	req, params, state := newAdjReq(s, chanAlloc, userIdx, 0, false)

	dSetup := chtest.NewDepositSetup(params, state)

	err := chtest.FundAll(s.NewCtx(), s.Funders, dSetup.FReqs)
	require.NoError(t, err)

	// initialize adjudicator as user with index 0
	adj := channel.NewAdjudicator(s.L2Accs[userIdx], s.ICConns[userIdx])

	ctx := s.NewCtx()

	// Channel is not yet registered
	s.AssertNoRegistered(state.ID)
	// Register the channel twice. Register should be idempotent.
	require.NoError(t, adj.Register(ctx, req, nil))
	// Check on-chain state for the register.
	require.NoError(t, err)
	s.AssertRegistered(state.ID)
}

func TestAdjudicator_ConcludeFinal(t *testing.T) {
	s := test.NewPerunSetup(t)

	chanAlloc := uint64(50000)

	userIdx := 0

	req, params, state := newAdjReq(s, chanAlloc, userIdx, 0, true)

	dSetup := chtest.NewDepositSetup(params, state)

	err := chtest.FundAll(s.NewCtx(), s.Funders, dSetup.FReqs)
	require.NoError(t, err)

	// Withdraw
	{
		// Alice
		adjIdx := 0
		adj := channel.NewAdjudicator(s.L2Accs[adjIdx], s.ICConns[adjIdx])
		ctx := s.NewCtx()
		require.NoError(t, adj.Withdraw(ctx, req, nil))
		withdrawerIdx := 1
		req.Idx = 1
		req.Acc = s.L2Accs[withdrawerIdx]
		adjWithdrawer := channel.NewAdjudicator(s.L2Accs[withdrawerIdx], s.ICConns[withdrawerIdx])
		require.NoError(t, adjWithdrawer.Withdraw(ctx, req, nil))
	}
}

func TestAdjudicator_Walkthrough(t *testing.T) {
	s := test.NewPerunSetup(t)
	chanAlloc := uint64(50000)
	userIdx := 0
	req, params, state := newAdjReq(s, chanAlloc, userIdx, 0, false)
	dSetup := chtest.NewDepositSetup(params, state)
	adjAliceIdx := 0
	adjAlice := channel.NewAdjudicator(s.L2Accs[adjAliceIdx], s.ICConns[adjAliceIdx])
	adjBobIdx := 1
	adjBob := channel.NewAdjudicator(s.L2Accs[adjBobIdx], s.ICConns[adjBobIdx])
	ctx := s.NewCtx()

	// Fund

	err := chtest.FundAll(ctx, s.Funders, dSetup.FReqs)
	require.NoError(t, err)

	// Dispute
	{
		// Register non-final state

		log.Println("Alice: Register non-final state")

		require.NoError(t, adjAlice.Register(ctx, req, nil))

		// Register non-final state with higher version
		next := req.Tx.State.Clone()
		next.Version++                        // increase version to allow progression
		test.MixBals(s.Rng, next.Balances[0]) // mix up the balances
		next.IsFinal = false
		sigs := s.SignState(next)
		req.Acc = s.L2Accs[adjBobIdx]
		req.Tx = pchannel.Transaction{State: next, Sigs: sigs}
		req.Idx = 1
		log.Println("Bob: Register some higher-version non-final state")

		require.NoError(t, adjBob.Register(ctx, req, nil))
		log.Println("Bob: Register/Conclude final state")
		// Register final state with higher version
		next = next.Clone()
		next.Version++ // increase version to allow progression
		next.IsFinal = true
		sigs = s.SignState(next)
		req.Tx = pchannel.Transaction{State: next, Sigs: sigs}
	}
	// Withdraw
	{
		// Bob
		log.Println("Bob: Withdraw")
		require.NoError(t, adjBob.Withdraw(ctx, req, nil))

		// Alice
		req.Idx = 0
		req.Acc = s.L2Accs[0]
		require.NoError(t, adjAlice.Withdraw(ctx, req, nil))
	}
}

func newAdjReq(s *test.PerunSetup, alloc uint64, userIdx int, version uint64, final bool) (pchannel.AdjudicatorReq, *pchannel.Params, *pchannel.State) {
	state := pchtest.NewRandomState(s.Rng, chtest.DefaultRandomOpts())
	state.IsFinal = final
	state.Version = version
	state.Allocation.Balances[0][0] = new(big.Int).SetUint64(alloc)
	state.Allocation.Balances[0][1] = new(big.Int).SetUint64(alloc)
	var data [20]byte
	s.Rng.Read(data[:])
	nonce := pchannel.NonceFromBytes(data[:])
	aliceAddr := s.Funders[0].GetAcc().L2Address()
	bobAddr := s.Funders[1].GetAcc().L2Address()

	params, err := pchannel.NewParams(60, []pwallet.Address{&aliceAddr, &bobAddr}, pchannel.NoApp(), nonce, true, false)
	require.NoError(s.T, err)
	state.ID = params.ID()
	sigs := s.SignState(state)

	require.NoError(s.T, err)

	chanIdx := pchannel.Index(userIdx)

	require.NoError(s.T, err)
	req := pchannel.AdjudicatorReq{
		Params:    params,
		Acc:       s.L2Accs[userIdx],
		Tx:        pchannel.Transaction{State: state, Sigs: sigs},
		Idx:       chanIdx,
		Secondary: false,
	}
	return req, params, state
}
