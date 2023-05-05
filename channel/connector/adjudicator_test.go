// SPDX-License-Identifier: Apache-2.0

package connector_test

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"math"
	pchannel "perun.network/go-perun/channel"
	pchtest "perun.network/go-perun/channel/test"
	pwallet "perun.network/go-perun/wallet"
	chanconn "perun.network/perun-icp-backend/channel/connector"
	"testing"

	"perun.network/perun-icp-backend/channel/connector/test"
	chtest "perun.network/perun-icp-backend/channel/test"
)

func TestAdjudicator_ConcludeFinal(t *testing.T) {
	s := test.NewSetup(t)
	_, params, state := newAdjReq(s, true)
	dSetup := chtest.NewDepositSetup(params, state)
	ctx := s.NewCtx()

	// Fund
	err := chtest.FundAll(ctx, s.Funders, dSetup.FReqs)
	assert.NoError(t, err)
	// Withdraw
	// {
	// 	// Alice
	// 	adj := chanconn.NewAdjudicator(s.Alice, s.Conns[0])
	// 	assert.NoError(t, adj.Withdraw(ctx, req, nil))
	// 	req.Idx = 1
	// 	req.Acc = s.Bob
	// 	adj = chanconn.NewAdjudicator(s.Bob, s.Conns[1])
	// 	assert.NoError(t, adj.Withdraw(ctx, req, nil))
	// }
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
