// SPDX-License-Identifier: Apache-2.0

package channel

import (
	"crypto/sha512"
	"encoding/binary"
	"fmt"
	chanconn "github.com/perun-network/perun-icp-backend/channel/connector"
	"math/big"
	pchannel "perun.network/go-perun/channel"
	pwallet "perun.network/go-perun/wallet"
)

// backend implements the backend interface.
// The type is private since it only needs to be exposed as singleton by the
// `Backend` variable.
type backend struct{}

// Backend is the channel backend. Is a singleton since there is only one backend.
var Backend backend

// CalcID calculates the channelID.
func (b *backend) CalcID(params *pchannel.Params) (id pchannel.ID) {
	id, err := CalcID(params)
	if err != nil {
		return pchannel.ID{}
	}
	return id
}

// Sign signs a state with the passed account.
func (*backend) Sign(acc pwallet.Account, state *pchannel.State) (pwallet.Sig, error) {
	// Provide signature to the state such that the canister can verify it on-chain.
	stateCan, err := chanconn.StateForChain(state)
	if err != nil {
		return nil, err
	}

	var stateBytes []byte

	stateBytes = append(stateBytes, stateCan.Channel[:]...)
	stateBytes = append(stateBytes, Uint64ToBytes(stateCan.Version)...)
	for _, a := range stateCan.Balances {
		myBigInt := big.NewInt(0).SetUint64(a)
		stateBytes = append(stateBytes, BigToLittleEndianBytes(myBigInt)...)
	}

	stateBytes = append(stateBytes, BoolToBytes(stateCan.Final)...)

	return acc.SignData(stateBytes)
}

// Verify verifies a signature on a state.
func (*backend) Verify(addr pwallet.Address, state *pchannel.State, sig pwallet.Sig) (bool, error) {
	stateCan, err := chanconn.StateForChain(state)
	if err != nil {
		return false, err
	}

	var stateBytes []byte

	stateBytes = append(stateBytes, stateCan.Channel[:]...)
	stateBytes = append(stateBytes, Uint64ToBytes(stateCan.Version)...)
	for _, a := range stateCan.Balances {
		myBigInt := big.NewInt(0).SetUint64(a)
		stateBytes = append(stateBytes, BigToLittleEndianBytes(myBigInt)...)
	}

	stateBytes = append(stateBytes, BoolToBytes(stateCan.Final)...)

	return pwallet.VerifySignature(stateBytes, sig, addr)
}

func (*backend) NewAsset() pchannel.Asset {
	return Asset
}

func CalcID(params *pchannel.Params) (pchannel.ID, error) {
	paramsICP, err := chanconn.NewParams(params)
	if err != nil {
		return pchannel.ID{}, fmt.Errorf("cannot calculate channel ID: %v", err)
	}

	partsSlices := make([][]byte, len(paramsICP.Participants))
	for i, part := range paramsICP.Participants {
		partCopy := make([]byte, len(part))

		copy(partCopy, part[:])

		partsSlices[i] = partCopy
	}

	if err != nil {
		return pchannel.ID{}, fmt.Errorf("could not encode parameters: %v", err)
	}

	var paramsEnc []byte

	nonceBytes := paramsICP.Nonce[:]
	part1Bytes := partsSlices[0]
	part2Bytes := partsSlices[1]

	challDurBytes := make([]byte, 8)

	binary.LittleEndian.PutUint64(challDurBytes, paramsICP.ChallengeDuration)

	paramsEnc = append(paramsEnc, nonceBytes...)
	paramsEnc = append(paramsEnc, part1Bytes...)
	paramsEnc = append(paramsEnc, part2Bytes...)
	paramsEnc = append(paramsEnc, challDurBytes...)

	hasher := sha512.New()

	hasher.Write(paramsEnc)
	hashSum := hasher.Sum(nil)

	IDLen := chanconn.IDLen

	if len(hashSum) < IDLen {
		return pchannel.ID{}, fmt.Errorf("hash length is less than IDLen")
	}

	var id pchannel.ID
	copy(id[:], hashSum[:IDLen])

	return id, nil
}
