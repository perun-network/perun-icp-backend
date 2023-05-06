// SPDX-License-Identifier: Apache-2.0

package channel

import (
	"fmt"
	"github.com/aviate-labs/agent-go/ledger"
	"github.com/aviate-labs/agent-go/principal"
	"os/exec"
	"time"
)

func MakeTransferArgs(memo uint64, amount uint64, fee uint64, recipient string) ledger.TransferArgs {
	p, _ := principal.Decode(recipient)
	subAccount := ledger.SubAccount(principal.DefaultSubAccount)

	return ledger.TransferArgs{
		Memo: memo,
		Amount: ledger.Tokens{
			E8S: amount,
		},
		Fee: ledger.Tokens{
			E8S: fee,
		},
		FromSubAccount: &subAccount,
		To:             p.AccountIdentifier(principal.DefaultSubAccount),
		CreatedAtTime: &ledger.TimeStamp{
			TimestampNanos: uint64(time.Now().UnixNano()),
		},
	}
}

func QueryCandidCLI(queryStateArgs string, canID string, execPath string) error {
	// Query the state of the Perun canister

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
