// SPDX-License-Identifier: Apache-2.0

package channel

import (
	"fmt"

	"os/exec"
)

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
	fmt.Println("Query Perun canister methods: ", string(output))

	return nil
}
