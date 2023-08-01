package connector

import (
	"fmt"
	"github.com/aviate-labs/agent-go/candid"

	"github.com/aviate-labs/agent-go/ic/icpledger"

	"github.com/aviate-labs/agent-go/principal"
	"math/big"
	"os/exec"
	pchannel "perun.network/go-perun/channel"
	pwallet "perun.network/go-perun/wallet"
	"perun.network/perun-icp-backend/channel/connector/icperun"
	"perun.network/perun-icp-backend/utils"
	"perun.network/perun-icp-backend/wallet"
	"time"
)

func Conclude(nonce Nonce, parts []pwallet.Address, chDur uint64, chanId ChannelID, vers Version, alloc *pchannel.Allocation, finalized bool, sigs []pwallet.Sig, canID principal.Principal, execPath ExecPath) (string, error) {

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

	formatedRequestWithdrawalArgs := utils.FormatConcludeCLIArgs(nonce[:], addrs, chDur, chanId[:], vers, allocInts, true, sigs[:]) //finalized

	path, err := exec.LookPath("dfx")
	if err != nil {
		return "", fmt.Errorf("failed to find 'dfx' executable: %w", err)
	}

	canIDString := canID.Encode()

	output, err := execCanisterCommand(path, canIDString, "conclude", formatedRequestWithdrawalArgs, execPath)

	if err != nil {
		return "", fmt.Errorf("failed conclude the channel: %w", err)
	}

	return string(output), nil
}

func Withdraw(funding Funding, signature pwallet.Sig, canID principal.Principal, execPath ExecPath) (string, error) {

	formatedRequestWithdrawalArgs := utils.FormatWithdrawalArgs(funding.Part[:], funding.Channel[:], signature[:])
	path, err := exec.LookPath("dfx")
	if err != nil {
		return "", fmt.Errorf("failed to find 'dfx' executable: %w", err)
	}

	canIDString := canID.Encode()

	output, err := execCanisterCommand(path, canIDString, "withdraw", formatedRequestWithdrawalArgs, execPath)

	if err != nil {
		return "", fmt.Errorf("failed to withdraw funds: %w", err)
	}

	return string(output), nil
}

func BuildDeposit(addr wallet.Address, cid ChannelID) icperun.Funding {

	addrbytes, err := addr.MarshalBinary()
	if err != nil {
		panic(err) //fmt.Errorf("failed to marshal participant address: %w", err)
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

	depoOutput, err := c.PerunAgent.DepositMemo(depositArgs)
	if err != nil {
		return fmt.Errorf("error depositing: %v", err)
	}

	c.Mutex.Unlock()
	fmt.Println("depoOutput: ", depoOutput)

	return nil
}

func MakeTransferArgs(memo uint64, amount uint64, fee uint64, recipient string) icpledger.TransferArgs {
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

func (c *Connector) TransferDfxAG(txArgs icpledger.TransferArgs) (*icpledger.TransferResult, error) {

	a := c.LedgerAgent
	transferResult, err := a.Transfer(txArgs)
	if err != nil {
		return nil, ErrFundTransfer
	}

	if transferResult.Err != nil {
		switch {
		case transferResult.Err.BadFee != nil:
			fmt.Printf("Transfer failed due to bad fee: expected fee: %v\n", transferResult.Err.BadFee.ExpectedFee)
		case transferResult.Err.InsufficientFunds != nil:
			fmt.Printf("Transfer failed due to insufficient funds: current balance: %v\n", transferResult.Err.InsufficientFunds.Balance)
		case transferResult.Err.TxTooOld != nil:
			fmt.Printf("Transfer failed because the transaction is too old. Allowed window (in nanos): %v\n", transferResult.Err.TxTooOld.AllowedWindowNanos)
		case transferResult.Err.TxCreatedInFuture != nil:
			fmt.Println("Transfer failed because the transaction was created in the future.")
		case transferResult.Err.TxDuplicate != nil:
			fmt.Printf("Transfer failed because it's a duplicate of transaction at block index: %v\n", transferResult.Err.TxDuplicate.DuplicateOf)
		default:
			fmt.Println("Transfer failed due to unknown reasons.")
		}
		return nil, fmt.Errorf("transfer failed with error: %v", transferResult.Err)
	}

	blnm := transferResult.Ok
	if blnm == nil {
		return nil, fmt.Errorf("Blocknumber is nil")
	}

	return transferResult, nil
}

func (c *Connector) NotifyTransferToPerun(blockNum BlockNum, recipientPerun principal.Principal) (uint64, error) {
	c.Mutex.Lock()
	defer c.Mutex.Unlock()
	receiptAmount, err := c.notifyDfx(blockNum, recipientPerun)
	fmt.Println("receiptAmount: ", receiptAmount)
	if err != nil {
		return 0, fmt.Errorf("error notifying transfer to Perun: %v", err)
	}

	fundedValue, err := utils.ExtractTxAmount(receiptAmount)
	if err != nil {
		return 0, fmt.Errorf("error extracting transaction amount: %v", err)
	}

	_, err = c.notifyDfx(blockNum, recipientPerun)
	if err != nil {
		return 0, fmt.Errorf("error for the (optional) second notification on the funding transfer: %v", err)
	}

	return uint64(fundedValue), nil
}

func (c *Connector) notifyDfx(blockNum BlockNum, notifyTo principal.Principal) (string, error) {
	// Notification of token transfer to the Perun canister
	formatedNotifyArgs := utils.FormatNotifyArgs(uint64(blockNum))
	encodedNotifyArgs, err := candid.EncodeValueString(formatedNotifyArgs)
	if err != nil {
		return "", fmt.Errorf("failed to encode notification arguments: %w", err)
	}
	respNote, err := c.DfxAgent.CallString(notifyTo, "transaction_notification", encodedNotifyArgs)
	if err != nil {
		return "", fmt.Errorf("failed to call notify method: %w", err)
	}

	return respNote, nil

}

func QueryStateCLI(queryStateArgs string, canID string, execPath string) error {
	// Query the state of the Perun canister

	path, err := exec.LookPath("dfx")
	if err != nil {
		return fmt.Errorf("unable to find 'dfx' executable in the system PATH: %w", err)
	}

	txCmd := exec.Command(path, "canister", "call", canID, "query_state", queryStateArgs)
	txCmd.Dir = execPath
	output, err := txCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to query the canister state: %w\nOutput: %s", err, output)
	}

	return nil
}

func (c *Connector) BuildDispute(acc pwallet.Account, params *pchannel.Params, state *pchannel.State, sigs []pwallet.Sig) (string, error) {

	parts := params.Parts

	addrs := make([][]byte, len(parts))
	for i, part := range parts {
		partMb, err := part.MarshalBinary()
		if err != nil {
			return "", fmt.Errorf("failed to marshal address: %w", err)
		}
		addrs[i] = partMb
	}

	alloc := state.Allocation

	allocInts := make([]int, len(alloc.Balances[0]))
	for i, balance := range alloc.Balances[0] {
		allocInts[i] = int(balance.Int64())
	}

	pid := params.ID()
	nonceHash, err := MakeNonce(params.Nonce)
	if err != nil {
		return "", fmt.Errorf("failed to make nonce hash: %w", err)
	}

	formatedRequestWithdrawalArgs := utils.FormatConcludeCLIArgs(nonceHash[:], addrs, params.ChallengeDuration, pid[:], state.Version, allocInts, true, sigs[:]) //finalized

	return formatedRequestWithdrawalArgs, nil
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

func (c *Connector) QueryMemo(memoArg uint64, queryAt principal.Principal) (string, error) {

	memoString := fmt.Sprintf("(%d : nat64)", memoArg)
	encodedQueryMemoArgs, err := candid.EncodeValueString(memoString)

	if err != nil {
		return "", fmt.Errorf("failed to encode query fid argument: %w", err)
	}

	respQuery, err := c.DfxAgent.CallString(queryAt, "query_memo", encodedQueryMemoArgs)
	if err != nil {
		return "", fmt.Errorf("failed to call query memo method: %w", err)
	}

	return respQuery, nil

}

// func depositFundMemPerunCLI(depositArgs DepositArgs, cid [64]byte, canID principal.Principal, execPath ExecPath) error {

// 	fmt.Println("depositing amount identified by a memo inside depositFundMemPerunCLI")

// 	addr, err := depositArgs.Participant.MarshalBinary()
// 	if err != nil {
// 		return fmt.Errorf("failed to marshal participant address: %w", err)
// 	}
// 	//channelIdSlice := []byte(depositArgs.ChannelId[:])
// 	channelIdSlice := []byte(cid[:])
// 	formatedQueryFundingMemoArgs := utils.FormatFundingMemoArgs(channelIdSlice, addr, depositArgs.Memo)

// 	fmt.Println("formatedQueryFundingMemoArgs: ", formatedQueryFundingMemoArgs)

// 	path, err := exec.LookPath("dfx")

// 	if err != nil {
// 		return fmt.Errorf("failed to find 'dfx' executable: %w", err)
// 	}

// 	canIDString := canID.Encode()

// 	_, err = execCanisterCommand(path, canIDString, "deposit_memo", formatedQueryFundingMemoArgs, execPath)
// 	if err != nil {
// 		return fmt.Errorf("failed to deposit amount identified by a memo: %w", err)
// 	}

// 	return nil
// }

// func (c *Connector) DepositFundMemPerun(depositArgs DepositArgs, canID principal.Principal) error {

// 	addr, err := depositArgs.Participant.MarshalBinary()
// 	if err != nil {
// 		return fmt.Errorf("failed to marshal participant address: %w", err)
// 	}
// 	channelIdSlice := []byte(depositArgs.ChannelId[:])

// 	formatedQueryFundingMemoArgs := utils.FormatFundingMemoArgs(channelIdSlice, addr, depositArgs.Memo)

// 	fmt.Println("formatedQueryFundingMemoArgs: ", formatedQueryFundingMemoArgs)

// 	if err != nil {
// 		return fmt.Errorf("failed to find 'dfx' executable: %w", err)
// 	}

// 	//canIDString := canID.Encode()
// 	args, err := candid.EncodeValueString(formatedQueryFundingMemoArgs)
// 	if err != nil {
// 		return fmt.Errorf("failed to encode query fid argument: %w", err)
// 	}

// 	dec, err := candid.DecodeValueString(args)
// 	if err != nil {
// 		return fmt.Errorf("failed to decode deposit_memo argument: %w", err)
// 	}
// 	fmt.Println("dec: ", dec)

// 	respQuery, err := c.Agent.CallString(canID, "deposit_memo", args)
// 	if err != nil {
// 		return fmt.Errorf("failed to call deposit_memo method: %w", err)
// 	}
// 	fmt.Println(respQuery)

// 	return nil
// }
