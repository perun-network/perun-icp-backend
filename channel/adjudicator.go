// SPDX-License-Identifier: Apache-2.0
package channel

//package connector

import (
	"context"
	"fmt"
	"github.com/aviate-labs/agent-go/candid"
	//"github.com/aviate-labs/agent-go/principal"

	"github.com/pkg/errors"
	"os/exec"
	pchannel "perun.network/go-perun/channel"
	"perun.network/go-perun/log"
	chanconn "perun.network/perun-icp-backend/channel/connector"
	"perun.network/perun-icp-backend/utils"

	"perun.network/perun-icp-backend/wallet"

	pwallet "perun.network/go-perun/wallet"
)

// Adjudicator implements the Perun Adjudicator interface.

type Adjudicator struct {
	acc  *wallet.Account
	Log  log.Embedding
	conn *chanconn.Connector
}

var (
	// ErrConcludedDifferentVersion a channel was concluded with a different version.
	ErrConcludedDifferentVersion = errors.New("channel was concluded with a different version")
	// ErrAdjudicatorReqIncompatible the adjudicator request was not compatible.
	ErrAdjudicatorReqIncompatible = errors.New("adjudicator request was not compatible")
	// ErrAdjudicatorReqIncompatible the adjudicator request was not compatible.
	ErrReqVersionTooLow = errors.New("request version too low")
)

// NewAdjudicator returns a new Adjudicator.
func NewAdjudicator(acc wallet.Account, c *chanconn.Connector) *Adjudicator {
	return &Adjudicator{
		acc:  &acc,
		conn: c,
		Log:  log.MakeEmbedding(log.Default()),
	}
}

func (a *Adjudicator) disputeRaw(ctx context.Context, req *pchannel.AdjudicatorReq) error {
	return nil
}

func (a *Adjudicator) Subscribe(ctx context.Context, cid pchannel.ID) (pchannel.AdjudicatorSubscription, error) {
	c := a.conn
	return chanconn.NewAdjudicatorSub(cid, c)
}

// func (a *Adjudicator) Subscribe(ctx context.Context, cid pchannel.ID) (pchannel.AdjudicatorSubscription, error) {
// 	return NewAdjudicatorSub(cid, a.pallet, a.storage, a.pastBlocks)
// }

func (a *Adjudicator) Dispute(nonce chanconn.Nonce, parts []pwallet.Address, chDur uint64, chanId chanconn.ChannelID, vers chanconn.Version, alloc *pchannel.Allocation, finalized bool, sigs []pwallet.Sig) (string, error) {

	addrs := make([][]byte, len(parts))
	for i, part := range parts {
		partMb, err := part.MarshalBinary()
		if err != nil {
			return "", fmt.Errorf("failed to marshal address: %w", err)
		}
		addrs[i] = partMb
	}
	allocInts := make([]int, len(alloc.Balances[0]))
	for i, balance := range alloc.Balances[0] {
		allocInts[i] = int(balance.Int64())
	}

	formatedRequestConcludeArgs := utils.FormatConcludeCLIArgs(nonce[:], addrs, chDur, chanId[:], vers, allocInts, true, sigs[:]) //finalized
	path, err := exec.LookPath("dfx")
	if err != nil {
		return "", fmt.Errorf("failed to find 'dfx' executable: %w", err)
	}

	canID := a.conn.PerunID
	canIDString := canID.String()
	execPath := a.conn.ExecPath
	output, err := chanconn.ExecCanisterCommand(path, canIDString, "dispute", formatedRequestConcludeArgs, execPath)

	if err != nil {
		return "", fmt.Errorf("failed a dispute call: %w", err)
	}

	return output, nil
}

func (a *Adjudicator) ConcludeDfxCLI(nonce chanconn.Nonce, parts []pwallet.Address, chDur uint64, chanId chanconn.ChannelID, vers chanconn.Version, alloc *pchannel.Allocation, finalized bool, sigs []pwallet.Sig) (string, error) {

	addrs := make([][]byte, len(parts))
	for i, part := range parts {
		partMb, err := part.MarshalBinary()
		if err != nil {
			return "", fmt.Errorf("failed to marshal address: %w", err)
		}
		addrs[i] = partMb
	}
	allocInts := make([]int, len(alloc.Balances[0]))
	for i, balance := range alloc.Balances[0] {
		allocInts[i] = int(balance.Int64())
	}

	formatedRequestConcludeArgs := utils.FormatConcludeCLIArgs(nonce[:], addrs, chDur, chanId[:], vers, allocInts, true, sigs[:]) //finalized
	//formatedRequestConcludeAGArgs := utils.FormatConcludeAGArgs(nonce[:], addrs, chDur, chanId[:], vers, allocInts, true, sigs[:]) //finalized

	path, err := exec.LookPath("dfx")
	if err != nil {
		return "", fmt.Errorf("failed to find 'dfx' executable: %w", err)
	}

	canID := a.conn.PerunID
	canIDString := canID.String()
	execPath := a.conn.ExecPath
	output, err := chanconn.ExecCanisterCommand(path, canIDString, "conclude", formatedRequestConcludeArgs, execPath)

	if err != nil {
		return "", fmt.Errorf("failed conclude the channel: %w", err)
	}

	return output, nil
}

func (a *Adjudicator) ConcludeAgentGo(nonce chanconn.Nonce, parts []pwallet.Address, chDur uint64, chanId chanconn.ChannelID, vers chanconn.Version, alloc *pchannel.Allocation, finalized bool, sigs []pwallet.Sig) (string, error) {

	addrs := make([][]byte, len(parts))
	for i, part := range parts {
		partMb, err := part.MarshalBinary()
		if err != nil {
			return "", fmt.Errorf("failed to marshal address: %w", err)
		}
		addrs[i] = partMb
	}
	allocInts := make([]int, len(alloc.Balances[0]))
	for i, balance := range alloc.Balances[0] {
		allocInts[i] = int(balance.Int64())
	}

	formatedRequestConcludeArgs := utils.FormatConcludeAGArgs(nonce[:], addrs, chDur, chanId[:], vers, allocInts, true, sigs[:]) //finalized

	fmt.Println("formatedRequestConcludeArgs", formatedRequestConcludeArgs)

	encodedConcludeArgs, err := candid.EncodeValueString(formatedRequestConcludeArgs)
	if err != nil {
		return "", fmt.Errorf("failed to encode conclude args: %w", err)
	}

	_, err = candid.DecodeValueString(encodedConcludeArgs)
	if err != nil {
		return "", fmt.Errorf("failed to decode conclude args: %w", err)
	}

	canID := a.conn.PerunID

	respNote, err := a.conn.Agent.CallString(*canID, "conclude", encodedConcludeArgs)
	if err != nil {
		return "", fmt.Errorf("failed to call notify method: %w", err)
	}

	if err != nil {
		return "", fmt.Errorf("failed conclude the channel: %w", err)
	}

	return respNote, nil
}

// Register registers and disputes a channel.
func (a *Adjudicator) Register(ctx context.Context, req pchannel.AdjudicatorReq, states []pchannel.SignedState) error {
	defer a.Log.Log().Trace("register done")
	// Input validation.
	if err := a.checkRegister(req, states); err != nil {
		return err
	}
	// Execute dispute.
	return a.dispute(ctx, req)
}

// Progress returns an error because app channels are currently not supported.
func (a *Adjudicator) Progress(ctx context.Context, req pchannel.ProgressReq) error {

	return nil
}

func (a *Adjudicator) Withdraw(ctx context.Context, req pchannel.AdjudicatorReq, smap pchannel.StateMap) error {
	pid := req.Params.ID()
	pidSlice := pid[:]
	addr := a.acc.Address()
	addrSlice, err := addr.MarshalBinary()
	if err != nil {
		return fmt.Errorf("failed to marshal address: %w", err)
	}

	sig := req.Tx.Sigs[0]
	formatedRequestWithdrawalArgs := utils.FormatWithdrawalArgs(addrSlice, pidSlice, sig)

	canIDString := a.conn.PerunID.Encode()

	path, err := exec.LookPath("dfx")
	if err != nil {
		return fmt.Errorf("failed to find 'dfx' executable: %w", err)
	}

	if err != nil {
		return fmt.Errorf("failed to find 'dfx' executable: %w", err)
	}

	execPath := a.conn.ExecPath

	output, err := chanconn.ExecCanisterCommand(path, canIDString, "withdraw", formatedRequestWithdrawalArgs, execPath)

	fmt.Println("output", output)

	if err != nil {
		return fmt.Errorf("failed to withdraw funds: %w", err)
	}

	return nil
}

func (a *Adjudicator) dispute(ctx context.Context, req pchannel.AdjudicatorReq) error {
	//defer a.Log.Log().Trace("Dispute done")

	// Setup the subscription for Disputed events.
	sub, err := a.conn.Subscribe(chanconn.EventIsDisputed(req.Params.ID()))
	if err != nil {
		return err
	}
	defer sub.Close()
	// Build Dispute Tx.
	disputeArgs, err := a.conn.BuildDispute(a.acc, req.Params, req.Tx.State, req.Tx.Sigs)
	if err != nil {
		return err
	}

	path, err := exec.LookPath("dfx")
	if err != nil {
		return fmt.Errorf("failed to find 'dfx' executable: %w", err)
	}

	canID := a.conn.PerunID
	canIDString := canID.String()
	execPath := a.conn.ExecPath
	output, err := chanconn.ExecCanisterCommand(path, canIDString, "dispute", disputeArgs, execPath)

	if err != nil {
		return fmt.Errorf("failed a dispute call: %w", err)
	}

	fmt.Println("output", output)

	// Wait for disputed event.
	return nil //a.waitForDispute(ctx, sub, req.Tx.Version)
}

// checkRegister returns an `ErrAdjudicatorReqIncompatible` error if
// the passed request cannot be handled by the Adjudicator.
func (*Adjudicator) checkRegister(req pchannel.AdjudicatorReq, states []pchannel.SignedState) error {
	switch {
	case req.Secondary:
		return errors.WithMessage(ErrAdjudicatorReqIncompatible, "secondary is not supported")
	case req.Tx.IsFinal:
		return errors.WithMessage(ErrAdjudicatorReqIncompatible, "cannot dispute a final state")
	case len(states) != 0:
		return errors.WithMessage(ErrAdjudicatorReqIncompatible, "sub-channels unsupported")
	default:
		return nil
	}
}
