// SPDX-License-Identifier: Apache-2.0

package channel

import (
	"fmt"
	utils "perun.network/perun-icp-backend/utils"
	"strings"
)

func ExecuteDFXTransfer(txArgs TxArgs, ledgerID, execPath string) (uint64, error) {
	transferFirstOutput, err := transferDfxCLI(txArgs, ledgerID, execPath)
	if err != nil {
		return 0, fmt.Errorf("error for first DFX transfer: %v", err)
	}

	blockNum, err := utils.ExtractBlock(transferFirstOutput)
	if err != nil {
		return 0, fmt.Errorf("error querying blocks: %v", err)
	}

	return blockNum, nil
}

func NotifyTransferToPerun(user *UserClient, blockNum uint64, recipientPerun Recipient) (uint64, error) {
	notifyArgs := NotifyArgs{Blocknum: blockNum}

	receiptAmount, err := user.notifyDfx(notifyArgs, recipientPerun)
	if err != nil {
		return 0, fmt.Errorf("error notifying transfer to Perun: %v", err)
	}

	fundedValue, err := utils.ExtractTxAmount(receiptAmount)
	if err != nil {
		return 0, fmt.Errorf("error extracting transaction amount: %v", err)
	}
	_, err = user.notifyDfx(notifyArgs, recipientPerun)
	if err != nil {
		return 0, fmt.Errorf("error for the (optional) second notification on the transfer to fund the Perun channel: %v", err)
	}

	return uint64(fundedValue), nil
}

type depositResult struct {
	FundingOutput  string
	OutputFundMemo string
	ChannelAlloc   int
}

func DepositToPerunChannel(user *UserClient, funding *Funding, memoFunding uint64, perunID, execPath string) (*depositResult, error) {
	depositArgs := DepositArgs{
		ChannelId:   funding.ChannelId.ID,
		Participant: user.L2Account.ICPAddress(),
		Memo:        memoFunding,
	}

	fundingOutput, err := queryFundingCLI(depositArgs, perunID, execPath)
	if err != nil {
		return nil, fmt.Errorf("error querying funding using the dfx CLI: %v", err)
	}

	fundMemoOutputVerbose, err := queryFundingMemoCLI(depositArgs, perunID, execPath)
	if err != nil {
		return nil, fmt.Errorf("error querying funding memo with CLI that I have provided: %v", err)
	}

	fundMemoOutput := strings.Replace(fundMemoOutputVerbose, "opt record {", "record {", 1)

	err = depositFundMemPerunCLI(depositArgs, perunID, execPath)
	if err != nil {
		return nil, fmt.Errorf("error depositing: %v", err)
	}

	holdingsOutput, err := queryHoldingsCLI(depositArgs, perunID, execPath)
	if err != nil {
		return nil, fmt.Errorf("error querying holdings: %v", err)
	}

	channelAlloc, err := utils.ExtractHoldingsNat(holdingsOutput)
	if err != nil {
		return nil, fmt.Errorf("error querying holdings: %v", err)
	}

	result := &depositResult{
		FundingOutput:  fundingOutput,
		OutputFundMemo: fundMemoOutput,
		ChannelAlloc:   channelAlloc,
	}

	return result, nil
}
