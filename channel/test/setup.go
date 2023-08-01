// SPDX-License-Identifier: Apache-2.0
package test

import (
	"context"
	cr "crypto/rand"
	"github.com/aviate-labs/agent-go/candid"
	"github.com/aviate-labs/agent-go/principal"
	"github.com/stretchr/testify/require"
	"math"
	"math/rand"

	pchannel "perun.network/go-perun/channel"
	pchtest "perun.network/go-perun/channel/test"
	"perun.network/perun-icp-backend/channel"

	"perun.network/perun-icp-backend/setup"
	"perun.network/perun-icp-backend/utils"

	pkgtest "polycry.pt/poly-go/test"

	pwallet "perun.network/go-perun/wallet"
	chanconn "perun.network/perun-icp-backend/channel/connector"
	"perun.network/perun-icp-backend/wallet"

	"testing"
	"time"
)

const BlockTime = 0.04

var DefaultTestTimeout = 20

// type FundingListParams struct {
// 	Users             []*client.DfxConnector
// 	ChallengeDuration uint64
// }

func NewRandL2Account() (wallet.Account, error) {
	wlt, err := wallet.NewRAMWallet(cr.Reader)
	if err != nil {
		return nil, err
	}
	acc := wlt.NewAccount()

	return acc, nil
}

func NewTransferSetup(t *testing.T) *Setup {
	testConfig := setup.DfxConfig{
		Host: "http://127.0.0.1",
		Port: 4943,
	}

	aliceAccPath := "./../../test/testdata/identities/usera_identity.pem"
	bobAccPath := "./../../test/testdata/identities/userb_identity.pem"
	//minterAccPath := "./../test/testdata/identities/minter_identity.pem"

	aliceL1Acc, err := chanconn.NewIdentity(aliceAccPath)
	if err != nil {
		panic(err)
	}
	bobL1Acc, err := chanconn.NewIdentity(bobAccPath)
	if err != nil {
		panic(err)
	}

	aliceL2Acc, err := NewRandL2Account()
	if err != nil {
		t.Fatal("Error generating random account 1:", err)
	}
	bobL2Acc, err := NewRandL2Account()
	if err != nil {
		t.Fatal("Error generating random account 2:", err)
	}
	accs := []wallet.Account{aliceL2Acc, bobL2Acc}
	alicePrince := (*aliceL1Acc).Sender()
	bobPrince := (*bobL1Acc).Sender()
	//minterPrince := (*minterAcc).Sender()

	perunID := "be2us-64aaa-aaaaa-qaabq-cai"
	ledgerID := "bkyz2-fmaaa-aaaaa-qaaaq-cai"

	// perunPrince, err := principal.Decode(perunID)
	// if err != nil {
	// 	panic(err)
	// }

	accsL1 := []*principal.Principal{&alicePrince, &bobPrince}
	conn1 := chanconn.NewDfxConnector(perunID, ledgerID, aliceAccPath, testConfig.Host, testConfig.Port)
	conn2 := chanconn.NewDfxConnector(perunID, ledgerID, bobAccPath, testConfig.Host, testConfig.Port)
	//connPerun := chanconn.NewConnector(perunID, ledgerID, minterAccPath, testConfig.Host, testConfig.Port)

	conns := []*chanconn.Connector{conn1, conn2}
	return &Setup{t, pkgtest.Prng(t), accs, accs[0], accs[1], accsL1, conns}
}

type Setup struct {
	T   *testing.T
	Rng *rand.Rand

	Accs       []wallet.Account
	Alice, Bob wallet.Account
	L1Accs     []*principal.Principal

	Conns []*chanconn.Connector
}

type DepositSetup struct {
	FReqs     []*pchannel.FundingReq
	FIDs      []uint64
	FinalBals []pchannel.Bal
	DReqs     []*channel.DepositReq
}

type TransferSetup struct {
	L1Accounts  []*principal.Principal
	MinterAcc   *principal.Principal
	Balances    []uint64
	AmountForTx []uint64
}

// NewRandomParamAndState generates compatible Params and State.
func (s *Setup) NewRandomParamAndState() (*pchannel.Params, *pchannel.State) {
	var params *pchannel.Params
	var state *pchannel.State

	for {
		params, state = pchtest.NewRandomParamsAndState(s.Rng, DefaultRandomOpts())

		// Check if ChallengeDuration and state.Version are within the valid range of int64
		if params.ChallengeDuration <= uint64(math.MaxInt64) && state.Version <= uint64(math.MaxInt64) {
			break
		} else {
			if params.ChallengeDuration > uint64(math.MaxInt64) {
				s.T.Logf("ChallengeDuration %v is not within the valid range of int64, generating new value...", params.ChallengeDuration)
			}
			if state.Version > uint64(math.MaxInt64) {
				s.T.Logf("state.Version %v is not within the valid range of int64, generating new value...", state.Version)
			}
		}
	}

	return params, state
}

// NewCtx returns a new context that will timeout after DefaultTestTimeout
// blocks and cancel on test cleanup.
func (s *Setup) NewCtx() context.Context {
	timeout := time.Duration(BlockTime * float64(time.Second) * float64(DefaultTestTimeout))
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	s.T.Cleanup(cancel)
	return ctx
}

func NewDepositSetup(params *pchannel.Params, state *pchannel.State, accs ...wallet.Account) *DepositSetup {
	reqAlice := pchannel.NewFundingReq(params, state, 0, state.Balances)
	reqBob := pchannel.NewFundingReq(params, state, 1, state.Balances)
	fReqAlice, _ := channel.MakeFundingReq(reqAlice)
	fReqBob, _ := channel.MakeFundingReq(reqBob)

	fidAlice, _ := fReqAlice.Memo()
	fidBob, _ := fReqBob.Memo()
	balAlice := state.Balances[0][reqAlice.Idx]
	balBob := state.Balances[0][reqBob.Idx]

	fReqs := []*pchannel.FundingReq{reqAlice, reqBob}
	dReqs := make([]*channel.DepositReq, len(accs))
	for i := range accs {
		dReqs[i], _ = channel.NewDepositReqFromPerun(fReqs[i], accs[i])
	}

	return &DepositSetup{
		FReqs:     []*pchannel.FundingReq{reqAlice, reqBob},
		FIDs:      []uint64{fidAlice, fidBob},
		FinalBals: []pchannel.Bal{balAlice, balBob},
		DReqs:     dReqs,
	}
}

// SignState returns the signatures for Alice and Bob on the state.
func (s *Setup) SignState(state *chanconn.State) []pwallet.Sig {

	stateArgs := utils.FormatStateArgs(state.Channel[:], state.Version, state.Balances, state.Final)

	data, err := candid.EncodeValueString(stateArgs)
	require.NoError(s.T, err)

	sig1, err := s.Alice.SignData(data)
	require.NoError(s.T, err)
	sig2, err := s.Bob.SignData(data)
	require.NoError(s.T, err)

	return []pwallet.Sig{sig1, sig2}
}
