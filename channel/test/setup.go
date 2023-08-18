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
	"context"
	cr "crypto/rand"
	"github.com/aviate-labs/agent-go/ic/icpledger"
	"github.com/aviate-labs/agent-go/principal"
	"math"
	"math/rand"

	"fmt"
	pchannel "perun.network/go-perun/channel"
	pchtest "perun.network/go-perun/channel/test"
	"perun.network/perun-icp-backend/channel"
	pkgtest "polycry.pt/poly-go/test"

	chanconn "perun.network/perun-icp-backend/channel/connector"
	"perun.network/perun-icp-backend/channel/connector/icperun"

	"path/filepath"
	"perun.network/perun-icp-backend/wallet"
	"testing"
	"time"
)

const (
	Host     = "http://127.0.0.1"
	Port     = 4943
	perunID  = "be2us-64aaa-aaaaa-qaabq-cai"
	ledgerID = "bkyz2-fmaaa-aaaaa-qaaaq-cai"
)

var DefaultTestTimeout = 200

func NewRandL2Account() (wallet.Account, error) {
	wlt, err := wallet.NewRAMWallet(cr.Reader)
	if err != nil {
		return nil, err
	}
	acc := wlt.NewAccount()

	return acc, nil
}

func NewTestSetup(t *testing.T) *Setup {

	basePath := chanconn.GetBasePath()

	aliceAccPath := filepath.Join(basePath, "./../../userdata/identities/usera_identity.pem")
	bobAccPath := filepath.Join(basePath, "./../../userdata/identities/userb_identity.pem")

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
		t.Fatal("Error generating random account:", err)
	}
	bobL2Acc, err := NewRandL2Account()
	if err != nil {
		t.Fatal("Error generating random account:", err)
	}
	accsL2 := []wallet.Account{aliceL2Acc, bobL2Acc}
	alicePrince := (*aliceL1Acc).Sender()
	bobPrince := (*bobL1Acc).Sender()

	accsL1 := []*principal.Principal{&alicePrince, &bobPrince}
	conn1 := chanconn.NewICConnector(perunID, ledgerID, aliceAccPath, Host, Port)
	conn2 := chanconn.NewICConnector(perunID, ledgerID, bobAccPath, Host, Port)

	conns := []*chanconn.Connector{conn1, conn2}
	return &Setup{t, pkgtest.Prng(t), accsL1, accsL2, conns}
}

type Setup struct {
	T       *testing.T
	Rng     *rand.Rand
	L1Accs  []*principal.Principal
	L2Accs  []wallet.Account
	ICConns []*chanconn.Connector
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
	timeout := time.Duration(float64(DefaultTestTimeout) * float64(time.Second))
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

func (s *Setup) GetL1Balances() ([]uint64, error) {

	bals := make([]uint64, len(s.L1Accs))

	for i, acc := range s.L1Accs {
		l1Ledger := s.ICConns[i].LedgerAgent

		accID := acc.AccountIdentifier(principal.DefaultSubAccount)

		bal, err := l1Ledger.AccountBalance(icpledger.AccountBalanceArgs{Account: accID.Bytes()})
		if err != nil {
			return nil, err
		}

		bals[i] = bal.E8s
		fmt.Println("bals getl1balances: ", bals)
	}

	return bals, nil
}

func (s *Setup) GetPerunBalances() ([]uint64, error) {

	bals := make([]uint64, len(s.L1Accs))

	for i := range s.L1Accs {
		l1Ledger := s.ICConns[i].LedgerAgent

		accID := s.ICConns[i].PerunID.AccountIdentifier((principal.DefaultSubAccount)) //acc.AccountIdentifier(principal.DefaultSubAccount)

		bal, err := l1Ledger.AccountBalance(icpledger.AccountBalanceArgs{Account: accID.Bytes()})
		if err != nil {
			return nil, err
		}

		bals[i] = bal.E8s
	}

	return bals, nil
}

func (s *Setup) GetChannelBalances(cid pchannel.ID) ([]uint64, error) {

	bals := make([]uint64, len(s.L2Accs))

	for i, acc := range s.L2Accs {
		ICConn := s.ICConns[i]
		l2Addr, err := acc.Address().MarshalBinary()
		if err != nil {
			return nil, err
		}
		queryBalArgs := icperun.Funding{
			Channel:     cid,
			Participant: l2Addr}
		balNat, err := ICConn.PerunAgent.QueryHoldings(queryBalArgs)
		if err != nil {
			return nil, err
		}

		bnn := *balNat

		if bnn == nil {
			bals[i] = uint64(0)
		} else {
			bn := bnn.BigInt()

			bals[i] = bn.Uint64()
		}

	}

	return bals, nil
}

func (s *Setup) AssertRegistered(cid pchannel.ID) {

	qs, err := s.ICConns[0].PerunAgent.QueryState(cid)
	if err != nil {
		s.T.Fatal(err)
	}
	if qs == nil {
		s.T.Fatal("no registered state on-chain found")
	}

}

func (s *Setup) AssertNoRegistered(cid pchannel.ID) {
	qs, err := s.ICConns[0].PerunAgent.QueryState(cid)
	if err != nil {
		s.T.Fatal(err)
	}
	if qs != nil {
		s.T.Fatal("registered state on-chain found")
	}
}
