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

package icperun

import (
	"math/big"

	"github.com/aviate-labs/agent-go"
	"github.com/aviate-labs/agent-go/candid/idl"
	"github.com/aviate-labs/agent-go/principal"
)

type Agent struct {
	a          *agent.Agent
	canisterId principal.Principal
}

type ChannelId = [32]byte
type Signature = Hash
type Amount = idl.Nat
type Duration = uint64
type Error = string
type Hash = []uint8
type L2Account = []uint8
type EventTxt = string
type Nonce = [32]byte
type Timestamp = uint64

type FullySignedState = struct {
	State State    `ic:"state"`
	Sigs  [][]byte `ic:"sigs"`
}

type FundMem = struct {
	Channel     ChannelId `ic:"channel"`
	Participant L2Account `ic:"participant"`
	Memo        uint64    `ic:"memo"`
}

type Funding = struct {
	Channel     ChannelId `ic:"channel"`
	Participant L2Account `ic:"participant"`
}

type Params = struct {
	Nonce             Nonce       `ic:"nonce"`
	Participants      []L2Account `ic:"participants"`
	ChallengeDuration Duration    `ic:"challenge_duration"`
}

type RegisteredState = struct {
	State   State     `ic:"state"`
	Timeout Timestamp `ic:"timeout"`
}

type State = struct {
	Channel    ChannelId `ic:"channel"`
	Version    uint64    `ic:"version"`
	Allocation []Amount  `ic:"allocation"`
	Finalized  bool      `ic:"finalized"`
}

type WithdrawalRequest = struct {
	Channel     ChannelId           `ic:"channel"`
	Participant L2Account           `ic:"participant"`
	Receiver    principal.Principal `ic:"receiver"`
	Sig         Signature           `ic:"signature"`
	Timestamp   Timestamp           `ic:"time"`
}

type AdjRequest = struct {
	Nonce             Nonce       `ic:"nonce"`
	Participants      []L2Account `ic:"participants"`
	ChallengeDuration Duration    `ic:"challenge_duration"`
	Channel           ChannelId   `ic:"channel"`
	Version           uint64      `ic:"version"`
	Allocation        []Amount    `ic:"allocation"`
	Finalized         bool        `ic:"finalized"`
	Sigs              []Signature `ic:"sigs"`
}

type Event = struct {
	who   L2Account `ic:"who"`
	total Amount    `ic:"total"`
}

type ChannelTime = struct {
	Channel   ChannelId `ic:"chanid"`
	Timestamp Timestamp `ic:"time"`
}

func NewBigNat(b *big.Int) Amount {
	return idl.NewBigNat(b)
}

func NewAgent(canisterId principal.Principal, config agent.Config) (*Agent, error) {
	a, err := agent.New(config)
	if err != nil {
		return nil, err
	}
	return &Agent{
		a:          a,
		canisterId: canisterId,
	}, nil
}

func (a Agent) Conclude(arg0 AdjRequest) (Error, error) {
	var r0 Error
	if err := a.a.Call(
		a.canisterId,
		"conclude",
		[]any{arg0},
		[]any{&r0},
	); err != nil {
		return "", err
	}
	return r0, nil
}

func (a Agent) Dispute(arg0 AdjRequest) (Error, error) {
	var r0 Error
	if err := a.a.Call(
		a.canisterId,
		"dispute",
		[]any{arg0},
		[]any{&r0},
	); err != nil {
		return "", err
	}
	return r0, nil
}

func (a Agent) QueryHoldings(arg0 Funding) (**Amount, error) {
	var r0 *Amount
	if err := a.a.Query(
		a.canisterId,
		"query_holdings",
		[]any{arg0},
		[]any{&r0},
	); err != nil {
		return nil, err
	}
	return &r0, nil
}

func (a Agent) QueryEvents(arg0 ChannelTime) (EventTxt, error) {
	var r0 EventTxt
	if err := a.a.Query(
		a.canisterId,
		"query_events",
		[]any{arg0},
		[]any{&r0},
	); err != nil {
		return "", err
	}

	return r0, nil
}

func (a Agent) QueryState(arg0 ChannelId) (*RegisteredState, error) {
	var r0 *RegisteredState
	if err := a.a.Call(
		a.canisterId,
		"query_state",
		[]any{arg0},
		[]any{&r0},
	); err != nil {
		return nil, err
	}
	return r0, nil
}

func (a Agent) TransactionNotification(arg0 uint64) (**Amount, error) {
	var r0 *Amount
	if err := a.a.Call(
		a.canisterId,
		"transaction_notification",
		[]any{arg0},
		[]any{&r0},
	); err != nil {
		return nil, err
	}
	return &r0, nil
}

func (a Agent) Withdraw(arg0 WithdrawalRequest) (Error, error) {
	var r0 Error
	if err := a.a.Call(
		a.canisterId,
		"withdraw",
		[]any{arg0},
		[]any{&r0},
	); err != nil {
		return "", err
	}
	return r0, nil
}

func (a Agent) Deposit(arg0 Funding) (*Error, error) {
	var r0 Error
	if err := a.a.Call(
		a.canisterId,
		"deposit",
		[]any{arg0},
		[]any{&r0},
	); err != nil {
		return nil, err
	}
	return &r0, nil
}

func (a Agent) RegisterEvent(arg0 FundMem) error {
	if err := a.a.Call(
		a.canisterId,
		"deposit_memo",
		[]any{arg0},
		[]any{},
	); err != nil {
		return err
	}
	return nil
}
