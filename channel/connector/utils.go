// SPDX-License-Identifier: Apache-2.0
package connector

import (
	"crypto/sha512"
	"encoding/binary"
	"fmt"
	"math/rand"
	"os/exec"
	"perun.network/perun-icp-backend/utils"
)

func FormatQueryStateArgs(chanId ChannelID) string {
	return fmt.Sprintf("(%s)", utils.FormatVec(chanId[:8]))
}

func execCanisterCommand(path, canID, method, args string, execPath ExecPath) (string, error) {
	txCmd := exec.Command(path, "canister", "call", canID, method, args)
	txCmd.Dir = string(execPath)
	output, err := txCmd.CombinedOutput()

	if err != nil {
		return "", fmt.Errorf("failed to execute canister command: %w\nOutput: %s", err, output)
	}

	return string(output), nil
}

// ExecCanisterCommand is a wrapper around the unexported execCanisterCommand function
// to make it accessible outside the utils package.
func ExecCanisterCommand(path, canID, method, args string, execPath ExecPath) (string, error) {
	return execCanisterCommand(path, canID, method, args, execPath)
}

func NonceHash(rng *rand.Rand) []byte {
	randomUint64 := rng.Uint64()
	bytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(bytes, randomUint64)
	hashArray := sha512.Sum512(bytes)
	hashSlice := hashArray[:]
	return hashSlice
}
