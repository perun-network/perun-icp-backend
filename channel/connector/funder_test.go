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
	"math/big"

	"github.com/stretchr/testify/require"

	"math/rand"
	chanconn "perun.network/perun-icp-backend/channel/connector"
	"perun.network/perun-icp-backend/channel/connector/test"
	chtest "perun.network/perun-icp-backend/channel/test"
	"time"

	"testing"
)

func TestFunding(t *testing.T) {
	s := test.NewPerunSetup(t)

	rand.Seed(time.Now().UnixNano())
	chanAlloc := uint64(rand.Intn(10000) + 1)

	params, state := s.NewRandomParamAndState()

	for i := 0; i < len(state.Allocation.Balances[0]); i++ {
		state.Allocation.Balances[0][i] = new(big.Int).SetUint64(chanAlloc)
		l2Address := s.Funders[i].GetAcc().L2Address()
		params.Parts[i] = &l2Address
	}

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
