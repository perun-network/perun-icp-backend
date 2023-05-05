// SPDX-License-Identifier: Apache-2.0

package channel

import (
	"crypto/sha256"

	"fmt"
	"github.com/aviate-labs/agent-go/candid"
	"log"
	pchannel "perun.network/go-perun/channel"
	pwallet "perun.network/go-perun/wallet"
	chanconn "perun.network/perun-icp-backend/channel/connector"
	"perun.network/perun-icp-backend/utils"
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
		// Log the error
		log.Printf("Error calculating channel ID: %v", err)
		return pchannel.ID{}
	}
	return id
}

// Sign signs a state with the passed account.
func (*backend) Sign(acc pwallet.Account, state *pchannel.State) (pwallet.Sig, error) {
	stateCan, err := chanconn.NewState(state)
	if err != nil {
		return nil, err
	}

	stateArgs := utils.FormatStateArgs(stateCan.Channel[:], stateCan.Version, stateCan.Balances, stateCan.Final)

	// Here we encode the state the way it is going to be used to transmit it to the canister: encoding a string into candid format.

	data, err := candid.EncodeValueString(stateArgs)
	if err != nil {
		return nil, err
	}
	return acc.SignData(data)
}

// Verify verifies a signature on a state.
func (*backend) Verify(addr pwallet.Address, state *pchannel.State, sig pwallet.Sig) (bool, error) {
	stateCan, err := chanconn.NewState(state)
	if err != nil {
		return false, err
	}
	stateArgs := utils.FormatStateArgs(stateCan.Channel[:], stateCan.Version, stateCan.Balances, stateCan.Final)

	data, err := candid.EncodeValueString(stateArgs)
	if err != nil {
		return false, err
	}
	return pwallet.VerifySignature(data, sig, addr)
}

// NewAsset returns a variable of type Asset, which can be used
// for unmarshalling an asset from its binary representation.
func (*backend) NewAsset() pchannel.Asset {
	return Asset
}

func CalcID(params *pchannel.Params) (pchannel.ID, error) {
	paramsICP, err := chanconn.NewParams(params)
	if err != nil {
		return pchannel.ID{}, fmt.Errorf("cannot calculate channel ID: %v", err)
	}
	nonceSlice := paramsICP.Nonce[:]

	partsSlices := make([][]byte, len(paramsICP.Participants))
	for i, part := range paramsICP.Participants {
		partsSlices[i] = part[:]
	}
	valueStr := utils.FormatParamsArgs(nonceSlice, partsSlices, params.ChallengeDuration)

	enc, err := candid.EncodeValueString(valueStr)
	if err != nil {
		return pchannel.ID{}, fmt.Errorf("could not encode parameters: %v", err)
	}

	_, err = candid.DecodeValueString(enc)

	if err != nil {
		return pchannel.ID{}, fmt.Errorf("could not decode parameters: %v", err)
	}

	hasher := sha256.New()

	hasher.Write(enc)
	hashSum := hasher.Sum(nil)

	IDLen := chanconn.IDLen

	if len(hashSum) < IDLen {
		return pchannel.ID{}, fmt.Errorf("hash length is less than IDLen")
	}

	var id pchannel.ID
	copy(id[:], hashSum) //
	return id, nil
}
