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

package test

import (
	"crypto/rand"
	"io"
	"math/big"
	mrand "math/rand"
	"testing"

	"github.com/stretchr/testify/require"
	pchannel "perun.network/go-perun/channel"
	pwallet "perun.network/go-perun/wallet"
	"perun.network/perun-icp-backend/channel"
	chanconn "perun.network/perun-icp-backend/channel/connector"
	chtest "perun.network/perun-icp-backend/channel/test"
	"perun.network/perun-icp-backend/wallet"
)

type PerunSetup struct {
	*chtest.Setup
	Deps    []*channel.Depositor
	Funders []*channel.Funder
	Adjs    []*channel.Adjudicator
}

func NewPerunSetup(t *testing.T) *PerunSetup {

	s := chtest.NewTestSetup(t)
	c := s.ICConns

	ret := &PerunSetup{Setup: s}

	for i := 0; i < len(s.L2Accs); i++ {
		dep := channel.NewDepositor(c[i])
		ret.Deps = append(ret.Deps, dep)
		l2acc, ok := s.L2Accs[i][wallet.ICPBackendID].(wallet.Account)
		if !ok {
			panic("L2 account is not of type wallet.Account")
		}
		ret.Funders = append(ret.Funders, channel.NewFunder(l2acc, c[i]))
		ret.Adjs = append(ret.Adjs, channel.NewAdjudicator(l2acc, c[i]))
	}

	return ret
}

func (s *PerunSetup) SignState(state *pchannel.State) []pwallet.Sig {
	stateCan, err := chanconn.StateForChain(state)
	require.NoError(s.T, err)
	var stateBytes []byte

	stateBytes = append(stateBytes, stateCan.Channel[:]...)
	stateBytes = append(stateBytes, channel.Uint64ToBytes(stateCan.Version)...)
	for _, a := range stateCan.Balances {
		myBigInt := big.NewInt(0).SetUint64(a)
		stateBytes = append(stateBytes, channel.BigToLittleEndianBytes(myBigInt)...)
	}

	stateBytes = append(stateBytes, channel.BoolToBytes(stateCan.Final)...)

	sig1, err := s.L2Accs[0][wallet.ICPBackendID].SignData(stateBytes)
	require.NoError(s.T, err)
	sig2, err := s.L2Accs[1][wallet.ICPBackendID].SignData(stateBytes)
	require.NoError(s.T, err)

	return []pwallet.Sig{sig1, sig2}
}

func MixBals(rng io.Reader, bals []pchannel.Bal) {
	// Transfer a random amount between two entries `len(bals)` times.
	for i := 0; i < len(bals); i++ {
		from, to := mrand.Intn(len(bals)), mrand.Intn(len(bals))
		diff, err := rand.Int(rng, bals[from])
		if err != nil {
			panic(err)
		}
		bals[from].Sub(bals[from], diff)
		bals[to].Add(bals[to], diff)
	}
}
