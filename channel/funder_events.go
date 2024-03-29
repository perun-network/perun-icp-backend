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

package channel

import (
	"fmt"
	"perun.network/perun-icp-backend/wallet"
	"regexp"
	"strconv"
	"strings"
)

type Event struct {
	EventType string
	Address   []uint8
	Timestamp int64
	Total     int
}

type FundedEvent struct {
	Address   wallet.Address
	Total     uint64
	Timestamp uint64
}

func parseEvents(input string) ([]FundedEvent, error) {
	rpkeys := regexp.MustCompile(`Funded event: Funded_who=PublicKey\(CompressedEdwardsY: \[(.*?)\]\)`)
	rptotal := regexp.MustCompile(`Funded_total=TotalStart(\d+(?:_\d+)*?)TotalEnd`)
	rptimestamp := regexp.MustCompile(`Funded_timestamp=TimestampStart(\d+)TimestampEnd`)

	matchesPkeys := rpkeys.FindAllStringSubmatch(input, -1)
	matchesTotal := rptotal.FindAllStringSubmatch(input, -1)
	matchesTimestamp := rptimestamp.FindAllStringSubmatch(input, -1)

	if len(matchesPkeys) != len(matchesTotal) || len(matchesPkeys) != len(matchesTimestamp) {
		return nil, fmt.Errorf("number of matches for public key, total, and timestamp don't match")
	}

	var events []FundedEvent

	for i, match := range matchesPkeys {
		byteString := extractBytesFromString(match[1])
		totalStr := matchesTotal[i][1]
		timestamp, err := strconv.ParseUint(matchesTimestamp[i][1], 10, 64)
		if err != nil {
			return nil, err
		}

		totalStr = strings.ReplaceAll(totalStr, "_", "")
		total, err := strconv.ParseUint(totalStr, 10, 64)
		if err != nil {
			return nil, err
		}

		event := FundedEvent{
			Address:   byteString,
			Total:     total,
			Timestamp: timestamp,
		}
		events = append(events, event)
	}

	return events, nil
}

func extractBytesFromString(input string) []byte {
	strBytes := regexp.MustCompile(`\d+`).FindAllString(input, -1)
	bytes := make([]byte, len(strBytes))

	for i, strByte := range strBytes {
		byteVal, _ := strconv.Atoi(strByte)
		bytes[i] = byte(byteVal)
	}

	return bytes
}
func SortEvents(eventString string) ([]FundedEvent, error) { //
	// Sort the events by their timestamp
	parsedEvents, err := parseEvents(eventString)
	if err != nil {
		return nil, err
	}
	return parsedEvents, nil
}

func EvaluateFundedEvents(events []FundedEvent, funderAddr wallet.Address, freqAmount, fundedTotal uint64) (bool, error) {

	for _, ev := range events {
		if ev.Address.Equal(&funderAddr) {
			fundedTotal += ev.Total
		}
	}

	if fundedTotal >= freqAmount {

		return true, nil
	}

	return false, nil
}
