// SPDX-License-Identifier: Apache-2.0
package connector

import (
	"crypto/sha256"
	"fmt"
	"github.com/aviate-labs/agent-go/candid"
	"math/big"
	pchannel "perun.network/go-perun/channel"
	pwallet "perun.network/go-perun/wallet"

	"perun.network/perun-icp-backend/utils"
	"perun.network/perun-icp-backend/wallet"
)

// NewFunding returns a new Funding.
func NewFunding(id ChannelID, part OffIdentity) *Funding {
	return &Funding{id, part}
}

func MakeOffIdent(part pwallet.Address) (OffIdentity, error) {
	var ret OffIdentity
	data, err := part.MarshalBinary()
	if err != nil {
		return ret, err
	}

	if len(data) != IDLen {
		return ret, ErrIdentLenMismatch
	}
	copy(ret[:], data)

	return ret, nil
}

func MakeOnIdent(addr wallet.Address) (OnIdentity, error) {
	var ret OnIdentity
	data, err := addr.MarshalBinary()
	if err != nil {
		return ret, err
	}

	if len(data) != OnIdentityLen {
		return ret, ErrIdentLenMismatch
	}
	copy(ret[:], data)

	return ret, nil
}

// ID calculates the funding ID by encoding and hashing the Funding.
func (f Funding) ID() (FundingID, error) {
	var fid FundingID
	addr := f.Part[:]
	fullArg := fmt.Sprintf("( record { channel = %s; participant = %s })", utils.FormatVec(f.Channel[:8]), utils.FormatVec(addr))
	data, err := candid.EncodeValueString(fullArg)
	if err != nil {
		return fid, fmt.Errorf("calculating funding ID: %w", err)
	}
	hashSum := sha256.Sum256(data)

	// Copy the first 32 bytes of the hashSum to the fid variable
	copy(fid[:], hashSum[:FIDLen])

	return fid, nil
}

func (f *Funding) SerializeFundingCandidFull() ([]byte, error) {
	// Encodes the funding struct
	addr := f.Part[:]

	fullArg := fmt.Sprintf("( record { channel = %s; participant = %s })", utils.FormatVec(f.Channel[:8]), utils.FormatVec(addr))

	enc, err := candid.EncodeValueString(fullArg)
	if err != nil {
		return nil, fmt.Errorf("failed to encode Funding as Candid value: %w", err)
	}

	_, err = candid.DecodeValueString(enc)
	if err != nil {
		return nil, fmt.Errorf("failed to decode encoded Funding Candid value: %w", err)
	}

	return enc, nil
}

func (f *Funding) SerializeFundingCandid() ([]byte, error) {
	// Encodes the funding struct
	addr := f.Part[:]

	fullArg := fmt.Sprintf("( record { channel = %s; participant = %s })", utils.FormatVec(f.Channel[:8]), utils.FormatVec(addr))

	enc, err := candid.EncodeValueString(fullArg)
	if err != nil {
		return nil, fmt.Errorf("failed to encode Funding as Candid value: %w", err)
	}

	_, err = candid.DecodeValueString(enc)
	if err != nil {
		return nil, fmt.Errorf("failed to decode encoded Funding Candid value: %w", err)
	}

	return enc, nil
}

// MakeChallengeDuration creates a new ChallengeDuration from the argument.
func MakeChallengeDuration(challengeDuration uint64) ChallengeDuration {
	return challengeDuration
}

// MakeNonce creates a new Nonce or an error if the argument was out of range.
func MakeNonce(nonce *big.Int) (Nonce, error) {
	var ret Nonce

	if nonce.Sign() < 0 { // negative?
		return ret, ErrNonceOutOfRange
	}
	if nonce.BitLen() > (8*NonceLen) || nonce.BitLen() == 0 { // too long/short?
		return ret, ErrNonceOutOfRange
	}
	copy(ret[:], nonce.Bytes())

	return ret, nil
}

// MakeOffIdents creates a new []OffIdentity.
func MakeOffIdents(parts []pwallet.Address) ([]OffIdentity, error) {
	var err error
	ret := make([]OffIdentity, len(parts))

	for i, part := range parts {
		if ret[i], err = MakeOffIdent(part); err != nil {
			break
		}
	}

	return ret, err
}

func MakeAlloc(a *pchannel.Allocation) ([]Balance, error) {
	var err error
	ret := make([]Balance, len(a.Balances[0]))

	if len(a.Assets) != 1 || len(a.Balances) != 1 || len(a.Locked) != 0 {
		return ret, ErrAllocIncompatible
	}
	for i, bal := range a.Balances[0] {
		if ret[i], err = MakeBalance(bal); err != nil {
			break
		}
	}

	return ret, err
}

func StateForChain(s *pchannel.State) (*State, error) {
	if err := s.Valid(); err != nil {
		return nil, ErrStateIncompatible
	}

	bals, err := MakeAlloc(&s.Allocation)
	if err != nil {
		return nil, err
	}

	data, err := s.Data.MarshalBinary()
	if err != nil {
		return nil, err
	}
	return &State{
		Channel:  s.ID,
		Version:  s.Version,
		Balances: bals,
		Final:    s.IsFinal,
		Data:     data,
	}, err
}

// NewParams creates backend-specific parameters from generic Perun parameters.
func NewParams(p *pchannel.Params) (*Params, error) {
	nonce, err := MakeNonce(p.Nonce)
	if err != nil {
		return nil, err
	}
	parts, err := MakeOffIdents(p.Parts)
	if err != nil {
		return nil, err
	}

	var appID OffIdentity
	if !pchannel.IsNoApp(p.App) {
		appID, err = MakeOffIdent(p.App.Def())
		if err != nil {
			return nil, err
		}
	}

	return &Params{
		Nonce:             nonce,
		Participants:      parts,
		ChallengeDuration: MakeChallengeDuration(p.ChallengeDuration),
		App:               appID,
	}, err
}
