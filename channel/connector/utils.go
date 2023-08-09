// SPDX-License-Identifier: Apache-2.0
package connector

import (
	"crypto/sha512"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"github.com/aviate-labs/agent-go/principal"
	"math/rand"
	"os/exec"
	pchannel "perun.network/go-perun/channel"
	"strconv"
	"strings"
)

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

func findMaxVersionIndex(matchesVersion [][]string) int {
	highestVersion := uint64(0)
	maxVersionIdx := -1

	for i, match := range matchesVersion {
		vers, err := strconv.ParseUint(match[1], 10, 64)
		if err != nil {
			return -1
		}
		if vers > highestVersion {
			highestVersion = vers
			maxVersionIdx = i
		}
	}

	return maxVersionIdx
}

func parseAllocations(allocStr1, allocStr2 string) (uint64, uint64, error) {
	allocStr1 = strings.Replace(allocStr1, "_", "", -1)
	allocStr2 = strings.Replace(allocStr2, "_", "", -1)

	alloc1, err := strconv.ParseUint(allocStr1, 10, 64)
	if err != nil {
		return 0, 0, err
	}

	alloc2, err := strconv.ParseUint(allocStr2, 10, 64)
	if err != nil {
		return 0, 0, err
	}

	return alloc1, alloc2, nil
}

func parseChannelID(hexString string) (pchannel.ID, error) {
	var cid pchannel.ID
	byteString, err := hex.DecodeString(hexString)
	if err != nil {
		return cid, err
	}
	copy(cid[:], byteString)
	return cid, nil
}

func DecodePrincipal(principalString string) (*principal.Principal, error) {
	decPrincipal, err := principal.Decode(principalString)
	if err != nil {
		return &principal.Principal{}, fmt.Errorf("error decoding Principal String: %w", err)
	}
	return &decPrincipal, nil
}
