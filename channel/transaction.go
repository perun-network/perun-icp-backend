package channel

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"regexp"
	"github.com/aviate-labs/agent-go/candid"
	"github.com/aviate-labs/agent-go/principal"
	"perun.network/perun-icp-backend/wallet"
	utils "perun.network/perun-icp-backend/utils"
)

// For the time being, we omit the subaccount and the timestamp
type TxArgs struct {
	Memo   uint64 //Memo for the transaction: In our case, the serialized funding struct -> identifies the transaction of the user aimed to fund the Perun channel
	Amount uint64 //Amount to be transferred to fund the channel
	Fee    uint64 //Fee for the transaction. Note that this is always zero if we use the minter as a sender
	//From_Subaccount []byte //Subaccount from which the funds are to be transferred
	To string //AccountIdentifier of the address as a string
	//CreatedAt       uint64 //Timestamp of the transaction
}

// Notifies the user of the block number of the transfer
type NotifyArgs struct {
	Blocknum uint64
}

// Queries the ledger for the funding properties of a channel
type QueryArgs struct {
	ChannelId []byte
	Participant wallet.Address 
}

// Defines the recipient of the fund transfer
type Recipient struct {
	ID principal.Principal
}

func transferDfxCLI(txArgs TxArgs, canID string, execPath string) (string, error) {
	formatedTransferArgs := FormatTransferArgs(txArgs.Memo, txArgs.Amount, txArgs.Fee, txArgs.To)
	path, err := exec.LookPath("dfx")
	if err != nil {
		return "", fmt.Errorf("dfx executable not found: %v", err)
	}

	txCmd := exec.Command(path, "canister", "call", canID, "transfer", formatedTransferArgs)
	txCmd.Dir = execPath
	output, err := txCmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("dfx transfer command failed: %v\nOutput: %s", err, output)
	}

	fmt.Printf("Response after transfer: %s\n", output)
	return string(output), nil
}



func (u *UserClient) TransferDfx(txArgs TxArgs, transferTo Recipient) error {

	formatedTransferArgs := FormatTransferArgs(txArgs.Memo, txArgs.Amount, txArgs.Fee, txArgs.To)
	encodedTxArgs, err := candid.EncodeValue(formatedTransferArgs)
	if err != nil {
		return fmt.Errorf("failed to encode transaction arguments: %w", err)
	}
	err = DecodeArgs(encodedTxArgs)

	fmt.Println("Encoded TxArgs: ", encodedTxArgs)
	if err != nil {
		return fmt.Errorf("failed to decode transaction arguments: %w", err)
	}
	respQS, err := u.Agent.Call(transferTo.ID, "transfer", encodedTxArgs)

	if err != nil {
		return fmt.Errorf("failed to call transfer method: %w", err)
	}

	fmt.Println("Sent transaction to ledger with response: ", respQS)
	return nil
}

func (u *UserClient) QueryFunding(fundingArgs DepositArgs, queryAt Recipient) error {
	// here we query the Perun Canister for the funding arguments which we send

	addr, err := fundingArgs.Participant.MarshalBinary()
	if err != nil {
		fmt.Println("Error: ", err)
	}

	formatedFundingArgs := FormatFundingArgs(addr, fundingArgs.ChannelId)
	encodedQueryFundingArgs, err := candid.EncodeValue(formatedFundingArgs)
	fmt.Println("Encoded QueryFunding Args: ", encodedQueryFundingArgs)

	if err != nil {
		return fmt.Errorf("failed to encode query funding arguments: %w", err)
	}

	respQuery, err := u.Agent.Call(queryAt.ID, "query_funding_only", encodedQueryFundingArgs)
	if err != nil {
		return fmt.Errorf("failed to call query state method: %w", err)
	}

	fmt.Println("Sent query for funding to Perun canister with response: ", respQuery)

	return nil
}

func (u *UserClient) QueryState(queryStateArgs DepositArgs, queryAt Recipient) error {
	formatedQueryStateArgs := FormatQueryStateArgs(queryStateArgs.ChannelId)

	encodedQueryStateArgs, err := candid.EncodeValue(formatedQueryStateArgs)
	fmt.Println("Encoded QueryStateArgs: ", encodedQueryStateArgs)

	if err != nil {
		return fmt.Errorf("failed to encode query state argument: %w", err)
	}

	respQuery, err := u.Agent.Call(queryAt.ID, "query_state", encodedQueryStateArgs)
	if err != nil {
		return fmt.Errorf("failed to call query state method: %w", err)
	}

	fmt.Println("Sent query for state to Perun canister with response: ", respQuery)
	return nil

}

func (u *UserClient) QueryMemo(memoArg uint64, queryAt Recipient) (string, error) {

	memoString := fmt.Sprintf("(%d : nat64)", memoArg)
	fmt.Println("memoString: ", memoString)
	encodedQueryMemoArgs, err := candid.EncodeValue(memoString)

	if err != nil {
		return "", fmt.Errorf("failed to encode query fid argument: %w", err)
	}

	respQuery, err := u.Agent.Call(queryAt.ID, "query_memo", encodedQueryMemoArgs)
	if err != nil {
		return "", fmt.Errorf("failed to call query memo method: %w", err)
	}

	fmt.Println("Sent query for memo to Perun canister with response: ", respQuery)
	return respQuery, nil

}

func (u *UserClient) notifyDfx(notifyArgs NotifyArgs, notifyTo Recipient) (string, error) {
	// Notification of token transfer to the Perun canister

	formatedNotifyArgs := utils.FormatNotifyArgs(notifyArgs.Blocknum)
	encodedNotifyArgs, err := candid.EncodeValue(formatedNotifyArgs)

	if err != nil {
		return "", fmt.Errorf("failed to encode notification arguments: %w", err)
	}

	respNote, err := u.Agent.Call(notifyTo.ID, "transaction_notification", encodedNotifyArgs)
	if err != nil {
		return "", fmt.Errorf("failed to call notify method: %w", err)
	}

	fmt.Println("Sent notification to the Perun canister with response: ", respNote)

	return string(respNote), nil

}
func QueryFidCLI(queryFidArgs DepositArgs, canID string, execPath string) (fid uint64, err error) {
    // Query the state of the Perun canister

	addr, err := queryFidArgs.Participant.MarshalBinary()
	if err != nil {
		fmt.Println("Error: ", err)
	}

    formatedQueryFidArgs := FormatFidArgs(addr, queryFidArgs.ChannelId)

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

func queryFundingCLI(queryFundingArgs DepositArgs, canID string, execPath string) (string, error) {
	// Query the state of the Perun canister

	addr, err := queryFundingArgs.Participant.MarshalBinary()
	if err != nil {
		fmt.Println("Error: ", err)
	}

	formatedQueryFundingArgs := FormatFundingArgs(addr, queryFundingArgs.ChannelId)

	path, err := exec.LookPath("dfx")
	if err != nil {
		return "", fmt.Errorf("unable to find 'dfx' executable in the system PATH: %w", err)
	}

	txCmd := exec.Command(path, "canister", "call", canID, "query_funding_only", formatedQueryFundingArgs)
	txCmd.Dir = execPath
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

func QueryCandidCLI(queryStateArgs string, canID string, execPath string) error {
	// Query the state of the Perun canister
	//formatedQueryStateArgs := FormatQueryStateArgs(queryStateArgs.ChannelId)

	path, err := exec.LookPath("dfx")
	if err != nil {
		return fmt.Errorf("unable to find 'dfx' executable in the system PATH: %w", err)
	}

	txCmd := exec.Command(path, "canister", "call", canID, "__get_candid_interface_tmp_hack", queryStateArgs)
	txCmd.Dir = execPath
	output, err := txCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to query canister state: %w\nOutput: %s", err, output)
	}

	fmt.Printf("Queried candid after attempted deposit: %s", output)
	return nil
}

func QueryStateCLI(queryStateArgs string, canID string, execPath string) error {
	// Query the state of the Perun canister
	//formatedQueryStateArgs := FormatQueryStateArgs(queryStateArgs.ChannelId)

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

	fmt.Printf("Queried state after attempted deposit: %s", output)
	return nil
}


func QueryEventsCLI(queryStateArgs string, canID string, execPath string) error {
	// Query the state of the Perun canister
	//formatedQueryStateArgs := FormatQueryStateArgs(queryStateArgs.ChannelId)

	path, err := exec.LookPath("dfx")
	if err != nil {
		return fmt.Errorf("unable to find 'dfx' executable in the system PATH: %w", err)
	}

	txCmd := exec.Command(path, "canister", "call", canID, "query_events", queryStateArgs)
	txCmd.Dir = execPath
	output, err := txCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to query canister events: %w\nOutput: %s", err, output)
	}

	fmt.Printf("Queried events after attempted deposit: %s", output)
	return nil
}

func queryHoldingsCLI(queryArgs DepositArgs, canID string, execPath string) (string, error) {
	addr, err := queryArgs.Participant.MarshalBinary()
	if err != nil {
		return "", fmt.Errorf("failed to marshal participant address: %w", err)
	}

	formatedQueryArgs := FormatFundingArgs(addr, queryArgs.ChannelId)

	path, err := exec.LookPath("dfx")
	if err != nil {
		return "", fmt.Errorf("unable to find 'dfx' executable in the system PATH: %w", err)
	}

	txCmd := exec.Command(path, "canister", "call", canID, "query_holdings", formatedQueryArgs)
	txCmd.Dir = execPath
	output, err := txCmd.CombinedOutput()

	if err != nil {
		return "", fmt.Errorf("failed to query holdings: %w\nOutput: %s", err, output)
	}

	fmt.Printf("User holdings in the channel after the deposit: %s\n", output)
	return string(output), nil
}

func queryFundingMemoCLI(depositArgs DepositArgs, canID string, execPath string) (string, error) {
	// Query the state of the Perun canister

	addr, err := depositArgs.Participant.MarshalBinary()
	if err != nil {
		return "", fmt.Errorf("failed to marshal participant address: %w", err)
	}

	formatedQueryFundingMemoArgs := FormatFundingMemoArgs(addr, depositArgs.ChannelId, depositArgs.Memo) //addr []byte, chanId []byte, memo uint64

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

	fmt.Printf("Funding and Memo define the Deposit identifier (Memo) and unique Channel ID with the querying participant: %s", output)
	return string(output), nil
}

func depositFundMemPerunCLI(depositArgs DepositArgs, canID string, execPath string) error {

	addr, err := depositArgs.Participant.MarshalBinary()
	if err != nil {
		return fmt.Errorf("failed to marshal participant address: %w", err)
	}

	formatedQueryFundingMemoArgs := FormatFundingMemoArgs(addr, depositArgs.ChannelId, depositArgs.Memo)

	path, err := exec.LookPath("dfx")
	if err != nil {
		return fmt.Errorf("failed to find 'dfx' executable: %w", err)
	}
	txCmd := exec.Command(path, "canister", "call", canID, "deposit_memo", formatedQueryFundingMemoArgs)
	txCmd.Dir = execPath
	output, err := txCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to execute deposit command: %w\nOutput: %s", err, string(output))
	}
	fmt.Println("deposit_memo output: ", string(output))
	return nil
}
