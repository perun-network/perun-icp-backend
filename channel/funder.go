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

package channel

import (
	"context"
	"fmt"
	"github.com/pkg/errors"

	"math/big"
	pchannel "perun.network/go-perun/channel"
	"perun.network/go-perun/log"
	pwallet "perun.network/go-perun/wallet"
	chanconn "perun.network/perun-icp-backend/channel/connector"
	"perun.network/perun-icp-backend/channel/connector/icperun"

	"perun.network/perun-icp-backend/wallet"

	"time"
)

const DefaultMaxIters = 15
const DefaultPollInterval = 3 * time.Second

type DepositReq struct {
	Balance pchannel.Bal
	Fee     pchannel.Bal
	Account wallet.Account
	Funding chanconn.Funding
}

type Funder struct {
	acc  *wallet.Account
	log  log.Embedding
	conn *chanconn.Connector
}

type FundingEventSub struct {
	agent        *icperun.Agent
	address      wallet.Address
	idx          pchannel.Index
	chanId       [32]byte
	timestamp    uint64
	queryArgs    icperun.ChannelTime
	fundingReq   pchannel.FundingReq
	pollInterval time.Duration
	maxIters     int
}

func NewFundingEventSub(addr wallet.Address, starttime uint64, req pchannel.FundingReq, conn *chanconn.Connector) (*FundingEventSub, error) {
	userIdx := req.Idx
	a := conn.PerunAgent
	cid := req.Params.ID()

	queryArgs := icperun.ChannelTime{
		Channel:   cid,
		Timestamp: starttime,
	}

	return &FundingEventSub{
		fundingReq:   req,
		agent:        a,
		address:      addr,
		idx:          userIdx,
		chanId:       cid,
		timestamp:    starttime,
		queryArgs:    queryArgs,
		pollInterval: DefaultPollInterval,
		maxIters:     DefaultMaxIters,
	}, nil
}

func (f *Funder) GetAcc() *wallet.Account {
	return f.acc
}

func (f *FundingEventSub) QueryEvents() (string, error) {
	return f.agent.QueryEvents(f.queryArgs)
}

func (f *FundingEventSub) QueryFundingState(ctx context.Context) error {

	funderAddr := f.address
	funderIdx := f.idx
	fundingReq := f.fundingReq
	fundingReqAlloc := fundingReq.Agreement[0][funderIdx].Uint64()
	fundedTotal := uint64(0)

polling:
	for i := 0; i < f.maxIters; i++ {
		select {
		case <-ctx.Done():
			return makeTimeoutErr(funderIdx)
		case <-time.After(f.pollInterval):
			eventStr, err := f.QueryEvents()
			if err != nil {
				continue polling
			}

			parsedEvents, err := parseEvents(eventStr)
			if err != nil {
				return errors.Wrap(err, "failed to sort events")
			}

			funded, err := EvaluateFundedEvents(parsedEvents, funderAddr, fundingReqAlloc, fundedTotal)
			if err != nil {
				return errors.Wrap(err, "failed to evaluate events")
			}
			if funded {
				return nil
			}
		}
	}
	return makeTimeoutErr(funderIdx)
}

func makeTimeoutErr(remains pchannel.Index) error {
	indices := make([]pchannel.Index, 0, 1)
	indices = append(indices, remains)
	return pchannel.NewFundingTimeoutError(
		[]*pchannel.AssetFundingError{{
			Asset:         Asset.Index(),
			TimedOutPeers: indices,
		}},
	)
}

func (d *Depositor) TransferToPerun(req *DepositReq) (chanconn.BlockNum, error) {

	transferArgs, err := d.cnr.BuildTransfer(*d.cnr.L1Account, req.Balance, req.Fee, req.Funding, *d.cnr.PerunID)
	if err != nil {
		return 0, fmt.Errorf("failed to build transfer: %w", err)
	}
	blockNum, err := d.cnr.TransferIC(transferArgs)
	if blockNum.Ok == nil {
		return 0, fmt.Errorf("blockNum is nil")
	}

	return chanconn.BlockNum(*blockNum.Ok), err
}

func (d *Depositor) Deposit(ctx context.Context, req *DepositReq) error { //, cid [32]byte

	// Transfer IC Tokens to the Perun canister with a unique memo.

	blnm, err := d.TransferToPerun(req)
	if err != nil {
		return fmt.Errorf("failed to execute IC transfer during channel opening: %w", err)
	}
	_, err = d.cnr.NotifyTransferToPerun(blnm, *d.cnr.PerunID)

	if err != nil {
		return fmt.Errorf("failed to notify transfer to perun: %w", err)
	}

	addr := req.Account.L2Address()

	err = d.cnr.DepositToPerunChannel(addr, req.Funding.Channel)
	if err != nil {
		return fmt.Errorf("failed to deposit to perun channel: %w", err)
	}

	return nil
}

func (f *Funder) Fund(ctx context.Context, req pchannel.FundingReq) error {

	acc := f.acc
	conn := f.conn
	addr := acc.L2Address()

	tstamp := uint64(0)
	wReq, err := NewDepositReqFromPerun(&req, acc)
	if err != nil {
		return err
	}

	if err := NewDepositor(conn).Deposit(ctx, wReq); err != nil {
		return err
	}

	evSub, err := NewFundingEventSub(addr, tstamp, req, conn)
	if err != nil {
		return fmt.Errorf("failed to create event subscription: %w", err)
	}

	err = evSub.QueryFundingState(ctx)
	if err != nil {
		return err
	}

	return nil
}

func NewFunder(acc wallet.Account, c *chanconn.Connector) *Funder {
	return &Funder{
		acc:  &acc,
		conn: c,
		log:  log.MakeEmbedding(log.Default()),
	}
}

func NewDepositReqFromPerun(req *pchannel.FundingReq, acc pwallet.Account) (*DepositReq, error) {
	if !req.Agreement.Equal(req.State.Balances) && (len(req.Agreement) == 1) {
		return nil, ErrFundingReqIncompatible
	}
	// The funding request is for a single asset channel, so the only entry is 0th index with two balances.
	bal := req.Agreement[0][req.Idx]
	fee := big.NewInt(chanconn.ICTransferFee)
	fReq, err := MakeFundingReq(req)
	if err != nil {
		return nil, errors.WithMessage(ErrFundingReqIncompatible, err.Error())
	}
	convAcc := *acc.(*wallet.Account)
	return NewDepositReq(bal, fee, convAcc, fReq), err
}

func MakeFundingReq(req *pchannel.FundingReq) (chanconn.Funding, error) {
	ident, err := chanconn.MakeOffIdent(req.Params.Parts[req.Idx])
	cid := req.Params.ID()
	return chanconn.Funding{
		Channel: cid,
		Part:    ident,
	}, err
}

func NewDepositReq(bal, fee pchannel.Bal, acc wallet.Account, funding chanconn.Funding) *DepositReq {
	return &DepositReq{bal, fee, acc, funding}
}

// NewDepositor returns a new Depositor.
func NewDepositor(cnr *chanconn.Connector) *Depositor {
	return &Depositor{log.MakeEmbedding(log.Default()), cnr}
}

type Depositor struct {
	log.Embedding
	cnr *chanconn.Connector
}
