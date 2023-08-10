// SPDX-License-Identifier: Apache-2.0
package channel

//package connector

import (
	"bytes"
	"context"
	"fmt"
	chanconn "github.com/perun-network/perun-icp-backend/channel/connector"
	"github.com/perun-network/perun-icp-backend/channel/connector/icperun"
	"github.com/pkg/errors"
	pchannel "perun.network/go-perun/channel"
	"perun.network/go-perun/log"
	"time"

	"github.com/perun-network/perun-icp-backend/wallet"

	pwallet "perun.network/go-perun/wallet"
)

// Adjudicator implements the Perun Adjudicator interface.

type Adjudicator struct {
	acc          *wallet.Account
	log          log.Embedding
	conn         *chanconn.Connector
	pollInterval time.Duration
	maxIters     int
}

// NewAdjudicator returns a new Adjudicator.
func NewAdjudicator(acc wallet.Account, c *chanconn.Connector) *Adjudicator {
	return &Adjudicator{
		acc:          &acc,
		conn:         c,
		log:          log.MakeEmbedding(log.Default()),
		maxIters:     DefaultMaxIters,
		pollInterval: DefaultPollInterval,
	}
}

func (a *Adjudicator) Subscribe(ctx context.Context, cid pchannel.ID) (pchannel.AdjudicatorSubscription, error) {
	c := a.conn
	return chanconn.NewAdjudicatorSub(ctx, cid, c), nil
}

// Register registers and disputes a channel.
func (a *Adjudicator) Register(ctx context.Context, req pchannel.AdjudicatorReq, states []pchannel.SignedState) error {
	defer a.log.Log().Trace("Registering done")

	cid := req.Params.ID()

	if req.Tx.IsFinal {
		_, err := a.ensureConcluded(ctx, req, nil, cid)
		if err != nil {
			return err
		}

	} else {

		if err := a.checkRegister(req, states); err != nil {
			return err
		}
		return a.dispute(ctx, req, cid)
	}
	return nil
}

// Progress returns nil because app channels are currently not supported.
func (a *Adjudicator) Progress(ctx context.Context, req pchannel.ProgressReq) error {

	return nil
}

func (a *Adjudicator) waitForDisputed(ctx context.Context, evsub *chanconn.AdjEventSub, cid pchannel.ID, version chanconn.Version) error {
	a.log.Log().Tracef("Waiting for conclude event")
loop:
	for {

		select {
		case event := <-evsub.Events():
			_, ok := event.(*chanconn.DisputedEvent)
			if !ok {
				continue loop
			}

			disputedVersion := event.Version()

			if disputedVersion < version {
				// The disputed Version is lower or equal than the recent one.
				a.log.Log().Tracef("Discarded dispute event. Version: %d", disputedVersion)
				// discard the event
				continue loop
			}
			a.log.Log().Debugf("Accepted dispute event. Version: %d", disputedVersion)

			evsub.Close()
			return nil

		case <-ctx.Done():
			return ctx.Err()
		case err := <-evsub.PanicErr():
			return err
		default:
			continue loop
		}
	}
}

func (a *Adjudicator) waitForConcluded(ctx context.Context, evsub *chanconn.AdjEventSub, cid pchannel.ID) error {
	a.log.Log().Tracef("Waiting for conclude event")
loop:
	for {

		select {
		case event := <-evsub.Events():

			_, ok := event.(*chanconn.ConcludedEvent)

			if !ok {
				continue loop
			}

			evsub.Close()
			return nil

		case <-ctx.Done():
			return ctx.Err()
		case err := <-evsub.PanicErr():
			return err
		default:
			continue loop
		}

	}
}

func (a *Adjudicator) isConcluded(ctx context.Context, cid pchannel.ID, req pchannel.AdjudicatorReq) (bool, error) {

	evSub := chanconn.NewAdjudicatorSub(ctx, cid, a.conn)

	adjReq, err := MakeAdjReq(req)
	if err != nil {
		return false, fmt.Errorf("failed to build AdjudicatorRequest: %w", err)

	}

	concludeResp, err := a.conn.PerunAgent.Conclude(adjReq)
	if err != nil {
		return false, ErrFailConclude
	}

	if concludeResp == ResponseErrorConcludingChannel {
		// calling conclusion failed: Look for a conclusion event that has been emitted by the other participant

		matched, err := a.queryAndMatchEvents(ctx, cid, req.Tx.State.Version)
		if err != nil || matched {
			return matched, err
		}

		return false, ErrFailConclude
	}

	err = a.waitForConcluded(ctx, evSub, cid)
	if err != nil {
		return false, fmt.Errorf("failed to wait for conclude event: %w", err)
	}

	// here we wait for a conclusion event to arrive

	defer evSub.Close()

	return true, nil
}

func (a *Adjudicator) queryAndMatchEvents(ctx context.Context, cid pchannel.ID, reqVersion uint64) (bool, error) {
	queryEventsArgs := icperun.ChannelTime{
		Channel:   cid,
		Timestamp: 0,
	}
	evString, err := a.conn.PerunAgent.QueryEvents(queryEventsArgs)
	if err != nil {
		return false, fmt.Errorf("failed to call query_events: %w", err)
	}

	concEvs, err := chanconn.ParseEventsConcluded(evString)
	if err != nil {
		return false, fmt.Errorf("failed to parse event stream: %w", err)
	}

	if len(concEvs) == 0 {
		return false, ErrFailConclude
	}

	for _, ev := range concEvs {
		if ev.VersionV == reqVersion && bytes.Equal(ev.IDV[:], cid[:]) {
			return true, nil
		}
	}

	return false, nil // Return false if no match found
}

func (a *Adjudicator) checkDisputes(ctx context.Context, req pchannel.AdjudicatorReq, nonce chanconn.Nonce) error {

	// check for disputes: if there are disputes for a non-finalized state, we need to verify everything and then conclude
	cid := req.Params.ID()

	qs, err := a.conn.PerunAgent.QueryState(cid)
	if err != nil {
		return fmt.Errorf("failed to query state: %w", err)
	}

	if qs == nil {
		return fmt.Errorf("failed to fetch registered state (dispute) and channel is not concludable yet: %w", err)
	}

	if qs.State.Version > req.Tx.Version {
		return fmt.Errorf("dispute version higher than the requested version")
	}

	concludeArgs, err := func() (*icperun.AdjRequest, error) {
		timeoutDispState := qs.Timeout
		chTimeout := chanconn.MakeTimeout(timeoutDispState)
		if err := chTimeout.Wait(ctx); err != nil {
			return nil, err
		}

		parts, err := MakeParts(req.Params.Parts)
		if err != nil {
			return nil, err
		}
		alloc := MakeAlloc(req.Tx.State.Allocation.Balances)

		concludeArgs := &icperun.AdjRequest{
			Nonce:             nonce,
			Participants:      parts,
			Channel:           cid,
			Version:           req.Tx.State.Version,
			ChallengeDuration: req.Params.ChallengeDuration,
			Allocation:        alloc,
			Sigs:              req.Tx.Sigs,
			Finalized:         req.Tx.IsFinal,
		}

		return concludeArgs, nil

	}()
	if err != nil {
		return fmt.Errorf("failed to wait for dispute timeout to finish: %w", err)
	}

	_, err = a.conn.PerunAgent.Conclude(*concludeArgs)
	if err != nil {
		return ErrFailConclude
	}

	evSub := chanconn.NewAdjudicatorSub(ctx, cid, a.conn) //a.Subscribe(ctx, cid)

	err = a.waitForConcluded(ctx, evSub, cid)
	if err != nil {
		return fmt.Errorf("failed to wait for dispute event: %w", err)
	}

	defer evSub.Close()

	qConcluded, err := a.conn.PerunAgent.QueryState(cid)
	if err != nil {
		return err
	}
	// Check that our version was concluded.
	if req.Tx.Version != qConcluded.State.Version {
		return ErrConcludedDifferentVersion
	}

	return nil
}

func MakeParts(addrs []pwallet.Address) ([][]byte, error) {
	parts := make([][]byte, len(addrs))

	for i, part := range addrs {
		partMb, err := part.MarshalBinary()
		if err != nil {
			return nil, fmt.Errorf("failed to marshal address: %w", err)
		}
		parts[i] = partMb
	}

	return parts, nil
}

func MakeAlloc(bals pchannel.Balances) []icperun.Amount {
	alloc := make([]icperun.Amount, len(bals[0]))

	for i, balance := range bals[0] {
		alloc[i] = icperun.NewBigNat(balance)
	}
	return alloc
}

func (a *Adjudicator) ensureConcluded(ctx context.Context, req pchannel.AdjudicatorReq, smap pchannel.StateMap, cid pchannel.ID) (bool, error) {
	// If not concluded, then ensure concludable and/or conclude

	concludeFinal := req.Tx.State.IsFinal && fullySignedTx(req.Tx, req.Params.Parts) == nil

	nonce, err := chanconn.MakeNonce(req.Params.Nonce)
	if err != nil {
		return false, err
	}

	if concludeFinal {
		// if concludeFinal is true, then we can conclude the channel if it has not been concluded already
		chanConcluded, err := a.isConcluded(ctx, cid, req)
		if err != nil {
			return false, err
		}

		if chanConcluded {
			// channel has been concluded, so we can withdraw
			return true, nil
		}

	} else {
		// channel is not concludable, so we need to dispute and check disputed states
		err := a.checkDisputes(ctx, req, nonce)
		if err != nil {
			return false, err
		}

	}

	return true, nil
}

func (a *Adjudicator) MakeWithdrawalReq(req pchannel.AdjudicatorReq) (icperun.WithdrawalRequest, error) {
	cid := req.Params.ID()
	addr := a.acc.Address()
	addrSlice, err := addr.MarshalBinary()
	if err != nil {
		return icperun.WithdrawalRequest{}, fmt.Errorf("failed to marshal address: %w", err)
	}
	receiver := a.conn.L1Account

	receiverBytes := receiver.Raw
	var msgEnc []byte

	msgEnc = append(msgEnc, cid[:]...)
	msgEnc = append(msgEnc, addrSlice[:]...)
	msgEnc = append(msgEnc, receiverBytes...)

	sig, err := a.acc.SignData(msgEnc)
	if err != nil {
		return icperun.WithdrawalRequest{}, fmt.Errorf("failed to sign data: %w", err)
	}

	withdrawReq := icperun.WithdrawalRequest{
		Channel:     cid,
		Participant: addrSlice[:],
		Receiver:    *receiver,
		Sig:         sig[:],
		Timestamp:   uint64(time.Now().Unix()),
	}

	return withdrawReq, nil
}

func (a *Adjudicator) withdraw(ctx context.Context, req pchannel.AdjudicatorReq, finalized bool, cid pchannel.ID) error {
	// If not concluded, then ensure concludable and/or conclude

	withdrawReq, err := a.MakeWithdrawalReq(req)
	if err != nil {
		return fmt.Errorf("failed to withdraw: %w", err)
	}

	withdrawalResp, err := a.conn.PerunAgent.Withdraw(withdrawReq)

	if withdrawalResp != WithdrawalSuccessResponse {
		return ErrFailWithdrawal
	}

	if err != nil {
		return fmt.Errorf("failed to withdraw: %w", err)
	}

	return nil
}

func (a *Adjudicator) Withdraw(ctx context.Context, req pchannel.AdjudicatorReq, smap pchannel.StateMap) error {

	cid := req.Params.ID()

	finalized, err := a.ensureConcluded(ctx, req, smap, cid)

	if err != nil {
		return err
	}

	err = a.withdraw(ctx, req, finalized, cid)
	if err != nil {
		return err
	}

	return nil
}

func MakeAdjReq(req pchannel.AdjudicatorReq) (icperun.AdjRequest, error) {

	cid := req.Params.ID()

	parts, err := MakeParts(req.Params.Parts)
	if err != nil {
		return icperun.AdjRequest{}, fmt.Errorf("failed to make parts: %w", err)
	}

	nonce, err := chanconn.MakeNonce(req.Params.Nonce)

	if err != nil {
		return icperun.AdjRequest{}, fmt.Errorf("failed to make nonce: %w", err)
	}

	alloc := MakeAlloc(req.Tx.State.Allocation.Balances)

	AdjArgs := icperun.AdjRequest{
		Nonce:             nonce,
		Participants:      parts,
		Channel:           cid,
		Version:           req.Tx.State.Version,
		ChallengeDuration: req.Params.ChallengeDuration,
		Allocation:        alloc,
		Finalized:         req.Tx.IsFinal,
		Sigs:              req.Tx.Sigs,
	}

	return AdjArgs, nil

}

func (a *Adjudicator) dispute(ctx context.Context, req pchannel.AdjudicatorReq, cid pchannel.ID) error {
	defer a.log.Log().Trace("Dispute done")

	disputeArgs, err := MakeAdjReq(req)
	if err != nil {
		return fmt.Errorf("failed to make adjudicator arguments: %w", err)
	}

	disputeResp, err := a.conn.PerunAgent.Dispute(disputeArgs)
	if err != nil || disputeResp != DisputeSuccess {
		return ErrFailDispute
	}

	// we fetch ALL events from the canister and check if any of them are disputed events
	evSub := chanconn.NewAdjudicatorSub(ctx, cid, a.conn)

	defer evSub.Close()
	err = a.waitForDisputed(ctx, evSub, cid, req.Tx.Version)
	if err != nil {
		return fmt.Errorf("failed to wait for dispute event: %w", err)
	}

	return nil
}

// checkRegister returns an `ErrAdjudicatorReqIncompatible` error if
// the passed request cannot be handled by the Adjudicator.
func (*Adjudicator) checkRegister(req pchannel.AdjudicatorReq, states []pchannel.SignedState) error {
	switch {
	case req.Tx.IsFinal:
		return errors.WithMessage(ErrAdjudicatorReqIncompatible, "cannot dispute a final state")
	case len(states) != 0:
		return errors.WithMessage(ErrAdjudicatorReqIncompatible, "sub-channels unsupported")
	default:
		return nil
	}
}

func fullySignedTx(tx pchannel.Transaction, parts []pwallet.Address) error {
	if len(tx.Sigs) != len(parts) {
		return errors.Errorf("wrong number of signatures")
	}

	for i, p := range parts {
		if ok, err := pchannel.Verify(p, tx.State, tx.Sigs[i]); err != nil {
			return errors.WithMessagef(err, "verifying signature[%d]", i)
		} else if !ok {
			return errors.Errorf("invalid signature[%d]", i)
		}
	}
	return nil
}

func (a *Adjudicator) GetAcc() *wallet.Account {
	return a.acc
}
