// SPDX-License-Identifier: Apache-2.0
package connector

import (
	//"github.com/aviate-labs/agent-go/principal"
	"github.com/pkg/errors"
	pchannel "perun.network/go-perun/channel"
	"perun.network/perun-icp-backend/wallet"
)

// Unique channel identifier

const (
	// OffIdentityLen is the length of an OffIdentity in byte.
	OffIdentityLen = 32
	// OnIdentityLen is the length of an OnIdentity in byte.
	OnIdentityLen = 32
	// NonceLen is the length of a Nonce in byte.
	NonceLen = 32
	// SigLen is the length of a Sig in byte.
	SigLen = 64
	// FIDLen is the length of a FundingId in byte.
	FIDLen = 32
)

type DisputePhase = uint8

const (
	RegisterPhase DisputePhase = iota
	ProgressPhase
	ConcludePhase
)

type (
	// Nonce makes a channels ID unique by providing randomness to the params.
	Nonce = [NonceLen]byte
	// ChannelID the ID of a channel as defined by go-perun.
	ChannelID = pchannel.ID
	// FundingID used to a the funding of a participant in a channel.
	FundingID = [FIDLen]byte
	// OffIdentity is an off-chain identity.
	OffIdentity = [OffIdentityLen]byte
	// OnIdentity is an on-chain identity.
	OnIdentity = [OnIdentityLen]byte
	// Version of a state as defined by go-perun.
	Version = uint64
	// ChallengeDuration the duration of a challenge as defined by go-perun.
	ChallengeDuration = uint64
	// Balance is the balance of an on- or off-chain Address.
	Balance = uint64
	// Sig is an off-chain signature.
	Sig = [SigLen]byte
	// AppID is the identifier of a channel application.
	AppID = OffIdentity

	// Params holds the fixed parameters of a channel and uniquely identifies it.
	Params struct {
		// Nonce is the unique nonce of a channel.
		Nonce Nonce
		// Participants are the off-chain participants of a channel.
		Participants []OffIdentity
		// ChallengeDuration is the duration that disputes can be refuted in.
		ChallengeDuration ChallengeDuration
		// App is the identifier of the channel application.
		App AppID
	}

	// State is the state of a channel.
	State struct {
		// Channel is the unique ID of the channel that this state belongs to.
		Channel ChannelID
		// Version is the version of the state.
		Version Version
		// Balances are the balances of the participants.
		Balances []Balance
		// Final whether or not this state is the final one.
		Final bool
		// Data is the channel's application data.
		Data []byte
	}

	// Withdrawal is used by a participant to withdraw his on-chain funds.
	Withdrawal struct {
		// Channel is the channel from which to withdraw.
		Channel ChannelID
		// Part is the participant who wants to withdraw.
		Part OffIdentity
		// Receiver is the receiver of the withdrawal.
		Receiver OnIdentity
	}

	// Funding is used to calculate a FundingId.
	Funding struct {
		// Channel is the channel to fund.
		Channel ChannelID
		// Part is the participant who wants to fund.
		Part OffIdentity
	}

	// RegisteredState is a channel state that was registered on-chain.
	RegisteredState struct {
		// Phase is the phase of the dispute.
		Phase DisputePhase
		// State is the state of the channel.
		State State
		// Timeout is the duration that the dispute can be refuted in.
		Timeout ChallengeDuration
	}
)

const IDLen = 32
const DfxTransferFee = 10000
const MaxBalance = uint64(1) << 30

type ChannelIdx = pchannel.ID
type Memo = uint64
type BlockNum uint64
type ExecPath string

type DepositArgs struct {
	ChannelId   ChannelID
	Participant wallet.Address
	Memo        Memo
}

// type TxArgs struct {
// 	Memo   Memo   //Memo for the transaction: In our case, the serialized funding struct -> identifies the transaction of the user aimed to fund the Perun channel
// 	Amount uint64 //Amount to be transferred to fund the channel
// 	Fee    uint64 //Fee for the transaction. Note that this is always zero if we use the minter as a sender
// 	From   wallet.Address
// 	//From_Subaccount []byte //Subaccount from which the funds are to be transferred
// 	To       principal.AccountIdentifier //AccountIdentifier of the address as a string
// 	ExecPath ExecPath
// 	//CreatedAt       uint64 //Timestamp of the transaction
// }

var ErrFundingReqIncompatible = errors.New("incompatible funding request")

// Notifies the user of the block number of the transfer
type NotifyArgs struct {
	Blocknum uint64
}

// Queries the ledger for the funding properties of a channel
type QueryArgs struct {
	ChannelId   []byte
	Participant wallet.Address
}

type Event struct {
	EventType string
	Who       string
	Total     uint64
}
