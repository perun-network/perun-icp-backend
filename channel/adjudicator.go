// SPDX-License-Identifier: Apache-2.0
package channel

//package connector

import (
	"fmt"
	"github.com/aviate-labs/agent-go/candid"

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
	log  log.Embedding
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
func NewAdjudicator(acc *wallet.Account, c *chanconn.Connector) *Adjudicator {
	return &Adjudicator{
		acc:  acc,
		conn: c,
		log:  log.MakeEmbedding(log.Default()),
	}
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
		allocInts[i] = int(balance.Int64()) // Convert *big.Int to int64 and then to int
	}

	formatedRequestConcludeArgs := utils.FormatConcludeArgs(nonce[:], addrs, chDur, chanId[:], vers, allocInts, true, sigs[:]) //finalized
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

	formatedRequestConcludeArgs := utils.FormatConcludeArgs(nonce[:], addrs, chDur, chanId[:], vers, allocInts, true, sigs[:]) //finalized

	encodedConcludeArgs, err := candid.EncodeValueString(formatedRequestConcludeArgs)
	if err != nil {
		return "", fmt.Errorf("failed to encode conclude args: %w", err)
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
