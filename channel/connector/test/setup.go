// SPDX-License-Identifier: Apache-2.0
package test

import (
	"github.com/stretchr/testify/require"
	"math/big"
	"sync"
	//"fmt"
	"crypto/rand"
	"io"
	mrand "math/rand"
	pchannel "perun.network/go-perun/channel"

	"perun.network/go-perun/wallet"

	"perun.network/perun-icp-backend/channel"
	chanconn "perun.network/perun-icp-backend/channel/connector"

	chtest "perun.network/perun-icp-backend/channel/test"
	"testing"
)

type PerunSetup struct {
	*chtest.Setup
	Deps    []*channel.Depositor
	Funders []*channel.Funder
	Adjs    []*channel.Adjudicator
}

func NewPerunSetup(t *testing.T) *PerunSetup {

	s := chtest.NewTestSetup(t)
	c := s.DfxConns

	ret := &PerunSetup{Setup: s}

	sharedMutex := &sync.Mutex{}

	for i := 0; i < len(s.L2Accs); i++ {
		c[i].Mutex = sharedMutex
		dep := channel.NewDepositor(c[i])
		ret.Deps = append(ret.Deps, dep)
		ret.Funders = append(ret.Funders, channel.NewFunder(s.L2Accs[i], c[i]))
		ret.Adjs = append(ret.Adjs, channel.NewAdjudicator(s.L2Accs[i], c[i]))
	}

	return ret
}

func (s *PerunSetup) SignState(state *pchannel.State) []wallet.Sig {
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

	sig1, err := s.L2Accs[0].SignData(stateBytes)
	require.NoError(s.T, err)
	sig2, err := s.L2Accs[1].SignData(stateBytes)
	require.NoError(s.T, err)

	return []wallet.Sig{sig1, sig2}
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

// func (*backend) Sign(acc pwallet.Account, state *pchannel.State) (pwallet.Sig, error) {
// 	// Provide signature to the state such that the canister can verify it on-chain.
// 	stateCan, err := chanconn.StateForChain(state)
// 	if err != nil {
// 		return nil, err
// 	}

// 	var stateBytes []byte

// 	stateBytes = append(stateBytes, stateCan.Channel[:]...)
// 	stateBytes = append(stateBytes, Uint64ToBytes(stateCan.Version)...)
// 	for _, a := range stateCan.Balances {
// 		myBigInt := big.NewInt(0).SetUint64(a)
// 		stateBytes = append(stateBytes, BigToLittleEndianBytes(myBigInt)...)
// 	}

// 	stateBytes = append(stateBytes, BoolToBytes(stateCan.Final)...)

// 	return acc.SignData(stateBytes)
// }
