// SPDX-License-Identifier: Apache-2.0

package channel

import (
	"fmt"
	"os/exec"
	pchannel "perun.network/go-perun/channel"
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

type FundEvent struct {
	Amount         pchannel.Bal
	Receiver       wallet.Address
	ChannelID      [64]byte
	Timestamp      uint64
	ParticipantIdx pchannel.Index
}

func parseEvents(input string) ([]FundedEvent, error) {
	rpkeys := regexp.MustCompile(`Funded event: Funded_who=PublicKey\(CompressedEdwardsY: \[(.*?)\]\)`)
	rptotal := regexp.MustCompile(`Funded_total=TotalStart(\d+(?:_\d+)*?)TotalEnd`)
	rptimestamp := regexp.MustCompile(`Funded_timestamp=TimestampStart(\d+)TimestampEnd`)

	matchesPkeys := rpkeys.FindAllStringSubmatch(input, -1)
	matchesTotal := rptotal.FindAllStringSubmatch(input, -1)
	matchesTimestamp := rptimestamp.FindAllStringSubmatch(input, -1)
	fmt.Printf("Length of matchesPkeys: %d\n", len(matchesPkeys))
	fmt.Printf("Length of matchesTotal: %d\n", len(matchesTotal))
	fmt.Printf("Length of matchesTimestamp: %d\n", len(matchesTimestamp))
	fmt.Println("Matches Total: ", matchesTotal)

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
	fmt.Println("Event string: ", parsedEvents)
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
