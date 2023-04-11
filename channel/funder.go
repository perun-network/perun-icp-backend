// SPDX-License-Identifier: Apache-2.0

package channel

import (
	"crypto/sha512"
	"encoding/binary"

	"errors"
	"fmt"
	"github.com/aviate-labs/agent-go/candid"
	"math/rand"
	utils "perun.network/perun-icp-backend/utils"
	"perun.network/perun-icp-backend/wallet"
	"unsafe"
)

type Params struct {
	Nonce             []byte
	Parts             []wallet.Address
	ChallengeDuration uint64
}

type Funding struct {
	L2Address wallet.Address //Wallet account
	ChannelId ChannelID
}

type ChannelID struct {
	ID []byte
}

func NewFunding() *Funding {
	return &Funding{}
}

type DepositArgs struct {
	ChannelId   []byte
	Participant wallet.Address
	Memo        uint64
}

func (p *Params) SerializeParamsCandid() ([]byte, error) {
	if len(p.Parts) != 2 {
		return nil, fmt.Errorf("expected exactly two participants, got %d", len(p.Parts))
	}

	challengeDurationBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(challengeDurationBytes, p.ChallengeDuration)

	paramsMotokoArgs := "record { "
	paramsNonce := "nonce = " + utils.FormatVec(p.Nonce) + " ; "
	ParamsParts := "participants = vec { " + utils.FormatVec(p.Parts[0]) + " ; " + utils.FormatVec(p.Parts[1]) + "};"
	paramsChallDuration := "challenge_duration = " + utils.FormatVec(challengeDurationBytes) + " }"

	paramsMotokoArgs = paramsMotokoArgs + paramsNonce + ParamsParts + paramsChallDuration

	enc, err := candid.EncodeValue(paramsMotokoArgs)
	if err != nil {
		return nil, fmt.Errorf("failed to encode params as Candid: %w", err)
	}
	_, err = candid.DecodeValue(enc)
	if err != nil {
		return nil, fmt.Errorf("failed to decode encoded params Candid: %w", err)
	}
	return enc, nil
}

func (p *Params) SerializeParams() ([]byte, error) {
	paramsBytes := []byte{}
	for _, part := range p.Parts {
		partBytes, err := part.MarshalBinary()
		if err != nil {
			return nil, err
		}
		paramsBytes = append(paramsBytes, partBytes...)
	}

	paramsBytes = append(paramsBytes, p.Nonce...)
	challengeDurationBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(challengeDurationBytes, p.ChallengeDuration)
	paramsBytes = append(paramsBytes, challengeDurationBytes...)
	return paramsBytes, nil
}

func (p *Params) ParamsIDCandid() ([]byte, error) {
	hasher := sha512.New()
	msg, err := p.SerializeParamsCandid()
	if err != nil {
		return nil, fmt.Errorf("failed to serialize params: %w", err)
	}
	hasher.Write(msg)
	return hasher.Sum(nil), nil
}

func (p *Params) ParamsIDStandard() ([]byte, error) {
	hasher := sha512.New()
	msg, err := p.SerializeParams()
	if err != nil {
		return nil, fmt.Errorf("failed to serialize params: %w", err)
	}
	lenMsg := len(msg)
	fmt.Println("lenMsg: ", lenMsg)
	hasher.Write(msg)
	fullHash := hasher.Sum(nil)

	if lenMsg <= 64 {
		return fullHash[:lenMsg], nil
	}
	result := make([]byte, lenMsg)
	copy(result[:64], fullHash)
	// The remaining bytes of 'result' are already initialized to zero

	return result, nil
}

func (p *Params) DeserializeParamsStandard(data []byte) error {
	addressLength := int(unsafe.Sizeof(wallet.Address{}))
	challengeDurationBytes := int(unsafe.Sizeof(p.ChallengeDuration))
	nonceLength := len(p.Nonce)

	if len(data) < challengeDurationBytes+nonceLength {
		return errors.New("insufficient data length")
	}

	// Deserialize Parts
	p.Parts = []wallet.Address{}
	for len(data) > challengeDurationBytes+nonceLength {
		part := new(wallet.Address)
		err := part.UnmarshalBinary(data[:addressLength])
		if err != nil {
			return err
		}

		p.Parts = append(p.Parts, *part)
		data = data[addressLength:]
	}

	// Deserialize Nonce
	p.Nonce = data[:nonceLength]
	data = data[nonceLength:]

	// Deserialize ChallengeDuration
	p.ChallengeDuration = binary.LittleEndian.Uint64(data)

	return nil
}

func NonceHash(rng *rand.Rand) []byte {
	randomUint64 := rng.Uint64()
	bytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(bytes, randomUint64)
	hashArray := sha512.Sum512(bytes)
	hashSlice := hashArray[:]
	return hashSlice
}

func (f *Funding) Memo() (uint64, error) {
	serializedFunding, err := f.SerializeFundingCandid()
	if err != nil {
		return 0, fmt.Errorf("error in serializing funding: %w", err)
	}

	hasher := sha512.New()
	hasher.Write(serializedFunding)
	fullHash := hasher.Sum(nil)

	var arr [8]byte
	copy(arr[:], fullHash[:8])
	memo := binary.LittleEndian.Uint64(arr[:])

	return memo, nil
}
