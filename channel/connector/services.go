// Copyright 2023 - See NOTICE file for copyright holders.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package connector

import (
	"fmt"
	"github.com/aviate-labs/agent-go/ic/icpledger"
	"github.com/aviate-labs/agent-go/principal"
	"math/big"
	"perun.network/perun-icp-backend/channel/connector/icperun"
	"perun.network/perun-icp-backend/wallet"
	"time"
)

func BuildDeposit(addr wallet.Address, cid ChannelID) icperun.Funding {

	addrbytes, err := addr.MarshalBinary()
	if err != nil {
		panic(err)
	}

	depositArgs := icperun.Funding{
		Channel:     cid,
		Participant: addrbytes,
	}

	return depositArgs
}

func (c *Connector) DepositToPerunChannel(addr wallet.Address, cid ChannelID) error {

	depositArgs := BuildDeposit(addr, cid)
	c.Mutex.Lock()
	defer c.Mutex.Unlock()

	_, err := c.PerunAgent.Deposit(depositArgs)
	if err != nil {
		return fmt.Errorf("error depositing: %v", err)
	}

	return nil
}

func MakeTransferArgs(memo Memo, amount uint64, fee uint64, recipient string) icpledger.TransferArgs {
	p, _ := principal.Decode(recipient)
	subAccount := icpledger.SubAccount(principal.DefaultSubAccount[:])

	return icpledger.TransferArgs{
		Memo: memo,
		Amount: icpledger.Tokens{
			E8s: amount,
		},
		Fee: icpledger.Tokens{
			E8s: fee,
		},
		FromSubaccount: &subAccount,
		To:             p.AccountIdentifier(principal.DefaultSubAccount).Bytes(),
		CreatedAtTime: &icpledger.TimeStamp{
			TimestampNanos: uint64(time.Now().UnixNano()),
		},
	}
}

func (c *Connector) TransferIC(txArgs icpledger.TransferArgs) (*icpledger.TransferResult, error) {
	transferResult, err := c.LedgerAgent.Transfer(txArgs)
	if err != nil {
		return nil, ErrFundTransfer
	}

	if transferResult.Err != nil {
		return nil, HandleTransferError(transferResult.Err)
	}

	if transferResult.Ok == nil {
		return nil, fmt.Errorf("Blocknumber is nil")
	}

	return transferResult, nil
}

func HandleTransferError(err *icpledger.TransferError) error {
	switch {
	case err.BadFee != nil:
		return fmt.Errorf("Transfer failed due to bad fee: expected fee: %v", err.BadFee.ExpectedFee)
	case err.InsufficientFunds != nil:
		return fmt.Errorf("Transfer failed due to insufficient funds: current balance: %v", err.InsufficientFunds.Balance)
	case err.TxTooOld != nil:
		return fmt.Errorf("Transfer failed because the transaction is too old. Allowed window (in nanos): %v", err.TxTooOld.AllowedWindowNanos)
	case err.TxCreatedInFuture != nil:
		return fmt.Errorf("Transfer failed because the transaction was created in the future.")
	case err.TxDuplicate != nil:
		return fmt.Errorf("Transfer failed because it's a duplicate of transaction at block index: %v", err.TxDuplicate.DuplicateOf)
	default:
		return fmt.Errorf("Transfer failed due to unknown reasons.")
	}
}

func (c *Connector) NotifyTransferToPerun(blockNum BlockNum, recipientPerun principal.Principal) (uint64, error) {
	c.Mutex.Lock()
	defer c.Mutex.Unlock()
	receiptAmount, err := c.PerunAgent.TransactionNotification(blockNum)
	if err != nil {
		return 0, fmt.Errorf("error notifying transfer to Perun: %v", err)
	}
	bnn := **receiptAmount

	bn := bnn.BigInt()
	bnint := bn.Uint64()

	if err != nil {
		return 0, fmt.Errorf("error extracting transaction amount: %v", err)
	}

	return bnint, nil
}

func (c *Connector) BuildTransfer(transactor principal.Principal, _amount, _fee *big.Int, funding Funding, receiver principal.Principal) (icpledger.TransferArgs, error) {

	amount, err := MakeBalance(_amount)
	if err != nil {

		return icpledger.TransferArgs{}, err
	}
	fee, err := MakeBalance(_fee)
	if err != nil {
		return icpledger.TransferArgs{}, err
	}

	memo, err := funding.Memo()

	if err != nil {

		return icpledger.TransferArgs{}, err
	}

	return icpledger.TransferArgs{
		Memo: memo,
		Amount: struct {
			E8s uint64 "ic:\"e8s\""
		}{E8s: amount + fee},
		Fee: struct {
			E8s uint64 "ic:\"e8s\""
		}{E8s: fee},
		//FromSubaccount: &accIDBytes,
		To: receiver.AccountIdentifier(principal.DefaultSubAccount).Bytes(),
	}, nil
}
