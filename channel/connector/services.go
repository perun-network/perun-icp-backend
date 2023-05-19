package connector

import (
	"fmt"
	"github.com/aviate-labs/agent-go/candid"
	"github.com/aviate-labs/agent-go/principal"

	"os/exec"
	pchannel "perun.network/go-perun/channel"
	pwallet "perun.network/go-perun/wallet"
	"perun.network/perun-icp-backend/utils"
	"perun.network/perun-icp-backend/wallet"

	"regexp"
	"strconv"
	"strings"
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

	formatedRequestWithdrawalArgs := utils.FormatConcludeCLIArgs(nonce[:], addrs, chDur, chanId[:], vers, allocInts, true, sigs[:])    //finalized
	formatedRequestWithdrawalArgsTEST := utils.FormatConcludeAGArgs(nonce[:], addrs, chDur, chanId[:], vers, allocInts, true, sigs[:]) //finalized

	path, err := exec.LookPath("dfx")
	if err != nil {
		return "", fmt.Errorf("failed to find 'dfx' executable: %w", err)
	}

	fmt.Println("formatedRequestWithdrawalArgs: ", formatedRequestWithdrawalArgs)
	fmt.Println("formatedRequestWithdrawalArgsTEST: ", formatedRequestWithdrawalArgsTEST)

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
		return "", fmt.Errorf("failed to deposit amount identified by a memo: %w", err)
	}

	return string(output), nil
}

// func (c *Connector) subscribe() {
// 	return nil
// }

func (c *Connector) DepositToPerunChannel(addr wallet.Address, chanID ChannelID, memoFunding Memo, perunID principal.Principal, execPath ExecPath) (uint64, error) {
	depositArgs := DepositArgs{
		ChannelId:   chanID,
		Participant: addr,
		Memo:        memoFunding,
	}

	_, err := queryFundingCLI(depositArgs, perunID, execPath)
	if err != nil {
		return 0, fmt.Errorf("error querying funding: %v", err)
	}

	if err := depositFundMemPerunCLI(depositArgs, perunID, execPath); err != nil {
		return 0, fmt.Errorf("error depositing: %v", err)
	}

	holdingsOutput, err := queryHoldingsCLI(depositArgs, perunID, execPath)
	if err != nil {
		return 0, fmt.Errorf("error querying holdings: %v", err)
	}

	channelAlloc, err := utils.ExtractHoldingsNat(holdingsOutput)
	if err != nil {
		return 0, fmt.Errorf("error querying holdings: %v", err)
	}

	return channelAlloc, nil
}

func queryHoldingsCLI(queryArgs DepositArgs, canID principal.Principal, execPath ExecPath) (string, error) {
	addr, err := queryArgs.Participant.MarshalBinary()
	if err != nil {
		return "", fmt.Errorf("failed to marshal participant address: %w", err)
	}
	channelIdSlice := []byte(queryArgs.ChannelId[:])
	formatedQueryArgs := utils.FormatFundingArgs(addr, channelIdSlice)

	path, err := exec.LookPath("dfx")
	if err != nil {
		return "", fmt.Errorf("unable to find 'dfx' executable in the system PATH: %w", err)
	}

	canIDString := canID.Encode()

	output, err := execCanisterCommand(path, canIDString, "query_holdings", formatedQueryArgs, execPath)
	if err != nil {
		return "", fmt.Errorf("failed to query holdings: %w", err)
	}

	fmt.Printf("User holdings in the channel after the deposit: %s\n", output)
	return output, nil
}

func queryFundingMemoCLI(depositArgs DepositArgs, canID string, execPath string) (string, error) {
	// Query the state of the Perun canister

	addr, err := depositArgs.Participant.MarshalBinary()
	if err != nil {
		return "", fmt.Errorf("failed to marshal participant address: %w", err)
	}
	channelIdSlice := []byte(depositArgs.ChannelId[:])

	formatedQueryFundingMemoArgs := utils.FormatFundingMemoArgs(addr, channelIdSlice, depositArgs.Memo) //addr []byte, chanId []byte, memo uint64

	path, err := exec.LookPath("dfx")
	if err != nil {
		return "", fmt.Errorf("unable to find 'dfx' executable in the system PATH: %w", err)
	}

	txCmd := exec.Command(path, "canister", "call", canID, "query_funding_memo", formatedQueryFundingMemoArgs)
	txCmd.Dir = execPath
	output, err := txCmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to query canister funding memo: %w\nOutput: %s", err, output)
	}

	return string(output), nil
}

func (c *Connector) ExecuteDFXTransfer(txArgs TxArgs, ledgerID principal.Principal, execPath ExecPath, transferFn TransferFunction) (BlockNum, error) {
	transferFirstOutput, err := transferFn(txArgs, ledgerID, execPath)

	if err != nil {
		return 0, fmt.Errorf("error for first DFX transfer: %v", err)
	}

	blockNum, err := utils.ExtractBlock(transferFirstOutput)

	if err != nil {
		return 0, fmt.Errorf("error querying blocks: %v", err)
	}

	return BlockNum(blockNum), nil
}

func (c *Connector) QueryFunding(fundingArgs DepositArgs, queryAt principal.Principal) error {
	// here we query the Perun Canister for the funding arguments which we send

	addr, err := fundingArgs.Participant.MarshalBinary()
	if err != nil {
		fmt.Println("Error: ", err)
	}
	channelIdSlice := []byte(fundingArgs.ChannelId[:])

	formatedFundingArgs := utils.FormatFundingArgs(addr, channelIdSlice)
	encodedQueryFundingArgs, err := candid.EncodeValueString(formatedFundingArgs)

	if err != nil {
		return fmt.Errorf("failed to encode query funding arguments: %w", err)
	}

	respQuery, err := c.Agent.CallString(queryAt, "query_funding_only", encodedQueryFundingArgs)
	if err != nil {
		return fmt.Errorf("failed to call query state method: %w", err)
	}

	fmt.Println("Sent query for funding to Perun canister with response: ", respQuery)

	return nil
}

func queryFundingCLI(queryFundingArgs DepositArgs, canID principal.Principal, execPath ExecPath) (string, error) {
	// Query the state of the Perun canister

	addr, err := queryFundingArgs.Participant.MarshalBinary()
	if err != nil {
		fmt.Println("Error: ", err)
	}

	formatedQueryFundingArgs := utils.FormatFundingArgs(addr, queryFundingArgs.ChannelId[:])

	path, err := exec.LookPath("dfx")
	if err != nil {
		return "", fmt.Errorf("unable to find 'dfx' executable in the system PATH: %w", err)
	}
	canIDString := canID.Encode()

	txCmd := exec.Command(path, "canister", "call", canIDString, "query_funding_only", formatedQueryFundingArgs)
	txCmd.Dir = string(execPath)
	rawOutput, err := txCmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to query canister state: %w\nOutput: %s", err, rawOutput)
	}

	output := string(rawOutput)
	startIndex := strings.Index(output, "record {")
	endIndex := strings.Index(output, "},") + 1

	if startIndex == -1 || endIndex == -1 {
		return "", fmt.Errorf("unexpected output format: %s", output)
	}

	formattedOutput := output[startIndex:endIndex]

	return formattedOutput, nil
}

func QueryFidCLI(queryFidArgs DepositArgs, canID string, execPath string) (fid uint64, err error) {
	// Query the state of the Perun canister

	addr, err := queryFidArgs.Participant.MarshalBinary()
	if err != nil {
		fmt.Println("Error: ", err)
	}
	channelIdSlice := []byte(queryFidArgs.ChannelId[:])

	formatedQueryFidArgs := utils.FormatFidArgs(addr, channelIdSlice)

	path, err := exec.LookPath("dfx")
	if err != nil {
		return 0, fmt.Errorf("unable to find 'dfx' executable in the system PATH: %w", err)
	}

	txCmd := exec.Command(path, "canister", "call", canID, "query_fid", formatedQueryFidArgs)
	txCmd.Dir = execPath
	output, err := txCmd.CombinedOutput()

	fmt.Println("Fid Output: ", string(output))

	if err != nil {
		return 0, fmt.Errorf("failed to query canister state: %w\nOutput: %s", err, output)
	}

	// Use regular expression to extract the nat64 value
	re := regexp.MustCompile(`\((\d{1,3}(?:_\d{3})*\d*) : nat64\)`)
	matches := re.FindStringSubmatch(string(output))

	if len(matches) < 2 {
		return 0, fmt.Errorf("failed to extract nat64 value from the output")
	}

	// Remove underscores from the matched string
	withoutUnderscores := strings.ReplaceAll(matches[1], "_", "")

	// Parse the nat64 value as uint64
	memo, err := strconv.ParseUint(withoutUnderscores, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse the memo value as uint64: %w", err)
	}

	return memo, nil
}

func (c *Connector) TransferDfxAG(txArgs TxArgs, canID principal.Principal, execPath ExecPath) (string, error) {
	ToString := txArgs.To.String()
	formatedTransferArgs := utils.FormatTransferArgs(txArgs.Memo, txArgs.Amount, txArgs.Fee, ToString)

	encodedTransferArgs, err := candid.EncodeValueString(formatedTransferArgs)

	if err != nil {
		return "", fmt.Errorf("failed to encode transfer arguments: %w", err)
	}
	fmt.Println("encodedTransferArgs for funding: ", encodedTransferArgs)
	respNote, err := c.Agent.CallString(canID, "transfer", encodedTransferArgs)
	if err != nil {
		return "", fmt.Errorf("failed to call notify method: %w", err)
	}

	fmt.Println("formatedTransferArgs for funding: ", formatedTransferArgs)

	fmt.Println("output from transfer: ", string(respNote))

	return string(respNote), nil
}

func (c *Connector) TransferDfx(txArgs TxArgs, canID principal.Principal, execPath ExecPath) (string, error) {
	ToString := txArgs.To.String()
	formatedTransferArgs := utils.FormatTransferArgs(txArgs.Memo, txArgs.Amount, txArgs.Fee, ToString)
	fmt.Println("formatedTransferArgs for funding: ", formatedTransferArgs)
	path, err := exec.LookPath("dfx")
	canIDString := canID.Encode()
	if err != nil {
		return "", fmt.Errorf("dfx executable not found: %v", err)
	}

	txCmd := exec.Command(path, "canister", "call", canIDString, "transfer", formatedTransferArgs)
	txCmd.Dir = string(execPath)
	output, err := txCmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("dfx transfer command failed: %v\nOutput: %s", err, output)
	}
	fmt.Println("output from transfer: ", string(output))

	return string(output), nil
}

type TransferFunction func(TxArgs, principal.Principal, ExecPath) (string, error)

func (c *Connector) NotifyTransferToPerun(blockNum BlockNum, recipientPerun principal.Principal, execPath ExecPath) (uint64, error) {
	receiptAmount, err := c.notifyDfx(blockNum, recipientPerun, execPath)
	if err != nil {
		return 0, fmt.Errorf("error notifying transfer to Perun: %v", err)
	}

	fundedValue, err := utils.ExtractTxAmount(receiptAmount)
	if err != nil {
		return 0, fmt.Errorf("error extracting transaction amount: %v", err)
	}

	_, err = c.notifyDfx(blockNum, recipientPerun, execPath)
	if err != nil {
		return 0, fmt.Errorf("error for the (optional) second notification on the transfer to fund the Perun channel: %v", err)
	}

	return uint64(fundedValue), nil
}

func (c *Connector) notifyDfx(blockNum BlockNum, notifyTo principal.Principal, execPath ExecPath) (string, error) {
	// Notification of token transfer to the Perun canister
	formatedNotifyArgs := utils.FormatNotifyArgs(uint64(blockNum))
	encodedNotifyArgs, err := candid.EncodeValueString(formatedNotifyArgs)
	if err != nil {
		return "", fmt.Errorf("failed to encode notification arguments: %w", err)
	}
	respNote, err := c.Agent.CallString(notifyTo, "transaction_notification", encodedNotifyArgs)
	if err != nil {
		return "", fmt.Errorf("failed to call notify method: %w", err)
	}

	return respNote, nil

}

func (c *Connector) QueryState(queryStateArgs DepositArgs, queryAt principal.Principal) error {
	formatedQueryStateArgs := FormatQueryStateArgs(queryStateArgs.ChannelId)

	encodedQueryStateArgs, err := candid.EncodeValueString(formatedQueryStateArgs)

	if err != nil {
		return fmt.Errorf("failed to encode query state argument: %w", err)
	}

	respQuery, err := c.Agent.CallString(queryAt, "query_state", encodedQueryStateArgs)
	if err != nil {
		return fmt.Errorf("failed to call query state method: %w", err)
	}

	fmt.Println("Sent query for state to Perun canister with response: ", respQuery)
	return nil

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
		return fmt.Errorf("failed to query canister state: %w\nOutput: %s", err, output)
	}

	return nil
}

func (c *Connector) VerifySig(nonce Nonce, parts []pwallet.Address, chDur uint64, chanId ChannelID, vers Version, alloc *pchannel.Allocation, finalized bool, sigs []pwallet.Sig, canID principal.Principal, execPath ExecPath) (string, error) {

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

	formatedRequestConcludeArgs := utils.FormatConcludeCLIArgs(nonce[:], addrs, chDur, chanId[:], vers, allocInts, true, sigs[:]) //finalized
	path, err := exec.LookPath("dfx")
	if err != nil {
		return "", fmt.Errorf("failed to find 'dfx' executable: %w", err)
	}

	canIDString := canID.String()
	output, err := ExecCanisterCommand(path, canIDString, "verify_sig", formatedRequestConcludeArgs, execPath)

	if err != nil {
		return "", fmt.Errorf("failed conclude the channel: %w", err)
	}

	return output, nil
}

// req.Account, req.Balance, req.FundingID
//
//	type ConcludeRequest = record {
//		nonce : Nonce;
//		participants : vec L2Account;
//		challenge_duration : Duration;
//		channel : ChannelId;
//		version : nat64;
//		allocation : vec Amount;
//		finalized : bool;
//		sigs : vec Signature;
//	};
//

// func (c *Connector) BuildDispute(acc wallet.Account, _amount, _fee pchannel.Bal, funding Funding) (TxArgs, error) {
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
		allocInts[i] = int(balance.Int64()) // Convert *big.Int to int64 and then to int
	}

	pid := params.ID()
	nonceHash, err := MakeNonce(params.Nonce)
	if err != nil {
		return "", fmt.Errorf("failed to make nonce hash: %w", err)
	}

	formatedRequestWithdrawalArgs := utils.FormatConcludeCLIArgs(nonceHash[:], addrs, params.ChallengeDuration, pid[:], state.Version, allocInts, true, sigs[:]) //finalized
	// path, err := exec.LookPath("dfx")
	// if err != nil {
	// 	return "", fmt.Errorf("failed to find 'dfx' executable: %w", err)
	// }

	// canIDString := canID.Encode()

	// output, err := execCanisterCommand(path, canIDString, "dispute", formatedRequestWithdrawalArgs, execPath)

	// if err != nil {
	// 	return "", fmt.Errorf("failed conclude the channel: %w", err)
	// }

	return formatedRequestWithdrawalArgs, nil
}

// req.Account, req.Balance, req.FundingID
func (c *Connector) BuildDeposit(acc wallet.Account, _amount, _fee pchannel.Bal, funding Funding) (TxArgs, error) {

	amount, err := MakeBalance(_amount)
	if err != nil {
		fmt.Println("Error m akebalanceamount: ", err)

		return TxArgs{}, err
	}
	fee, err := MakeBalance(_fee)
	if err != nil {
		fmt.Println("Error m akebalancefee: ", err)
		return TxArgs{}, err
	}

	memo, err := funding.Memo()

	if err != nil {
		fmt.Println("Error m memo: ", err)

		return TxArgs{}, err
	}

	perunaccountID := c.PerunID.AccountIdentifier(principal.DefaultSubAccount)

	return TxArgs{
		Memo:   memo,
		Amount: amount,
		Fee:    fee,
		From:   acc.ICPAddress(), // ICPAddress is the Layer2 address on the Perun canister on the replica, it is NOT the same as the ICP address on the ledger canister on the replica
		To:     perunaccountID,   //c.LedgerID.AccountIdentifier(principal.DefaultSubAccount),
	}, nil
}

func (c *Connector) QueryMemo(memoArg uint64, queryAt principal.Principal) (string, error) {

	memoString := fmt.Sprintf("(%d : nat64)", memoArg)
	fmt.Println("memoString: ", memoString)
	encodedQueryMemoArgs, err := candid.EncodeValueString(memoString)

	if err != nil {
		return "", fmt.Errorf("failed to encode query fid argument: %w", err)
	}

	respQuery, err := c.Agent.CallString(queryAt, "query_memo", encodedQueryMemoArgs)
	if err != nil {
		return "", fmt.Errorf("failed to call query memo method: %w", err)
	}

	fmt.Println("Sent query for memo to Perun canister with response: ", respQuery)
	return respQuery, nil

}

func depositFundMemPerunCLI(depositArgs DepositArgs, canID principal.Principal, execPath ExecPath) error {

	addr, err := depositArgs.Participant.MarshalBinary()
	if err != nil {
		return fmt.Errorf("failed to marshal participant address: %w", err)
	}
	channelIdSlice := []byte(depositArgs.ChannelId[:])

	formatedQueryFundingMemoArgs := utils.FormatFundingMemoArgs(addr, channelIdSlice, depositArgs.Memo)

	path, err := exec.LookPath("dfx")
	if err != nil {
		return fmt.Errorf("failed to find 'dfx' executable: %w", err)
	}

	canIDString := canID.Encode()

	output, err := execCanisterCommand(path, canIDString, "deposit_memo", formatedQueryFundingMemoArgs, execPath)
	if err != nil {
		return fmt.Errorf("failed to deposit amount identified by a memo: %w", err)
	}

	fmt.Println("Deposited amount: ", string(output))

	return nil
}
