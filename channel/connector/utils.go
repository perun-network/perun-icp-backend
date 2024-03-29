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
	"encoding/hex"
	"fmt"
	"github.com/aviate-labs/agent-go/principal"
	"path/filepath"
	pchannel "perun.network/go-perun/channel"
	"regexp"
	"runtime"
	"strconv"
	"strings"
)

func parseAdjEvents(input string, event AdjEvent,
	channelIDPattern, versionPattern, finalizedPattern, allocPattern, timeoutPattern, timestampPattern string) error {

	patterns := []string{channelIDPattern, versionPattern, finalizedPattern, allocPattern, timeoutPattern, timestampPattern}
	matches := make([][][]string, len(patterns))

	for i, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		matches[i] = re.FindAllStringSubmatch(input, -1)
		if matches[i] == nil || len(matches[i]) == 0 {
			return nil
		}
	}

	maxVersionIdx := 0
	if len(matches[0]) != 1 {
		maxVersionIdx = findMaxVersionIndex(matches[1])
	}

	cid, err := parseChannelID(matches[0][maxVersionIdx][1])
	if err != nil {
		return err
	}

	version, err := strconv.ParseUint(matches[1][maxVersionIdx][1], 10, 64)
	if err != nil {
		return err
	}

	finalized, err := strconv.ParseBool(matches[2][maxVersionIdx][1])
	if err != nil {
		return err
	}

	alloc1, alloc2, err := parseAllocations(matches[3][maxVersionIdx][1], matches[3][maxVersionIdx][2])
	if err != nil {
		return err
	}

	timeout, err := strconv.ParseUint(matches[4][maxVersionIdx][1], 10, 64)
	if err != nil {
		return err
	}

	timestamp, err := strconv.ParseUint(matches[5][maxVersionIdx][1], 10, 64)
	if err != nil {
		return err
	}

	return event.SetEventData(cid, version, finalized, [2]uint64{alloc1, alloc2}, timeout, timestamp) //timeout
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

func GetBasePath() string {
	_, filename, _, ok := runtime.Caller(1) // Call depth of 1 to get the caller of this function
	if !ok {
		panic("No caller information")
	}

	basePath := filepath.Dir(filename)
	return basePath
}
