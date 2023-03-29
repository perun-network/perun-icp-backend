package channel

import (
	"fmt"
	utils "perun.network/perun-icp-backend/utils"
	"github.com/aviate-labs/agent-go/candid"
	"strings"
)

func (f *Funding) SerializeFundingCandid() ([]byte, error) {
	addr, err := f.L2Address.MarshalBinary()
	if err != nil {
		return nil, fmt.Errorf("failed to marshal L2Address to binary: %w", err)
	}

	fullArg := fmt.Sprintf("( record { channel = %s; participant = %s })", utils.FormatVec(f.ChannelId.ID), utils.FormatVec(addr))

	enc, err := candid.EncodeValue(fullArg)
	if err != nil {
		return nil, fmt.Errorf("failed to encode Funding as Candid value: %w", err)
	}

	_, err = candid.DecodeValue(enc)
	if err != nil {
		return nil, fmt.Errorf("failed to decode encoded Funding Candid value: %w", err)
	}

	return enc, nil
}


func (f *Funding) SerializeFundingStandard() ([]byte, error) {
	addrBytes, err := f.L2Address.MarshalBinary()

	if err != nil {
		return nil, err
	}

	fundingBytes := append(addrBytes, f.ChannelId.ID...)

	return fundingBytes, err
}

func (f *Funding) DeserializeFundingStandard(data []byte) error {
	addrBytes := data[:len(data)-len(f.ChannelId.ID)]
	err := f.L2Address.UnmarshalBinary(addrBytes)
	if err != nil {
		return err
	}

	f.ChannelId.ID = data[len(data)-len(f.ChannelId.ID):]

	return err

}

func DecodeArgs(args []byte) error {
	fmt.Println("args to decode: ", args)
	_, err := candid.DecodeValue(args)
    if err != nil {
        return fmt.Errorf("failed to decode args: %v", err)
    }
	return nil
}

func FormatTransferArgs(memo, amount, fee uint64, sendTo string) string {
	var builder strings.Builder

	builder.WriteString(fmt.Sprintf(
		"(record {memo = %d : nat64; amount = record { e8s=%s : nat64}; fee = record { e8s=%s : nat64}; from_subaccount = null; to = blob\"%s\"; created_at_time = null; })",
		memo,
		utils.FormatWithUnderscores(amount),
		utils.FormatWithUnderscores(fee),
		utils.FormatHex(sendTo),
	))

	return builder.String()
}

func FormatFundingMemoArgs(addr, chanId []byte, memo uint64) string {
	var builder strings.Builder

	builder.WriteString(fmt.Sprintf(
		"(record {channel = %s; participant = %s; memo = %d : nat64 })",
		utils.FormatVec(chanId),
		utils.FormatVec(addr),
		memo,
	))

	return builder.String()
}

func FormatQueryStateArgs(chanId []byte) string {
	return fmt.Sprintf("(%s)", utils.FormatVec(chanId))
}

func FormatFidArgs(addr, chanId []byte) string {
	return fmt.Sprintf("(record {channel = %s; participant = %s})", utils.FormatVec(chanId), utils.FormatVec(addr))
}

func FormatFundingArgs(addr, chanId []byte) string {
    return fmt.Sprintf("( record { channel = %s; participant = %s } )", utils.FormatVec(chanId), utils.FormatVec(addr))
}
