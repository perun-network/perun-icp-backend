// SPDX-License-Identifier: Apache-2.0
package connector

import (
	"errors"
	"github.com/perun-network/perun-icp-backend/wallet"
	"math/big"
	pchannel "perun.network/go-perun/channel"
	pwallet "perun.network/go-perun/wallet"
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

// MakeChallengeDuration creates a new ChallengeDuration from the argument.
func MakeChallengeDuration(challengeDuration uint64) ChallengeDuration {
	return challengeDuration
}

// MakeNonce creates a new Nonce or an error if the argument was out of range.
func MakeNonce(nonce *big.Int) (Nonce, error) {
	var ret Nonce

	if nonce.Sign() < 0 {
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
func MakeBalance(bal *big.Int) (Balance, error) {
	if bal.Sign() < 0 {
		return 0, errors.New("invalid balance: negative value")
	}

	maxBal := new(big.Int).SetUint64(MaxBalance)
	if bal.Cmp(maxBal) > 0 {
		return 0, errors.New("invalid balance: exceeds max balance")
	}

	return Balance(bal.Uint64()), nil
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

func NewState(s *pchannel.State) (*State, error) {
	return StateForChain(s)
}
