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

// func extractEventData(input string) ([]Event, error) {
// 	lines := strings.Split(input, "\n")
// 	var events []Event
// 	var currentEvent *Event

// 	for _, line := range lines {
// 		line = strings.TrimSpace(line)

// 		eventTypes := []string{"Funded", "Disputed", "Concluded"}
// 		for _, eventType := range eventTypes {
// 			if strings.Contains(line, eventType) {
// 				if currentEvent != nil {
// 					events = append(events, *currentEvent)
// 				}
// 				currentEvent = &Event{EventType: eventType}
// 				break
// 			}
// 		}

// 		startIdx := strings.Index(line, "blob \"")
// 		endIdx := strings.Index(line, "\";")
// 		if startIdx != -1 && endIdx != -1 && startIdx < endIdx {
// 			who := line[startIdx+len("blob \"") : endIdx]
// 			if currentEvent != nil {
// 				currentEvent.Who = who
// 			}
// 		}

// 		totalStartIdx := strings.Index(line, "total =")
// 		if totalStartIdx != -1 {
// 			totalEndIdx := strings.Index(line, " : nat;")
// 			cleanedNumber := strings.ReplaceAll(line[totalStartIdx+len("total ="):totalEndIdx], "_", "")
// 			cleanedNumber = strings.TrimSpace(cleanedNumber)
// 			total, err := strconv.ParseUint(cleanedNumber, 10, 64)
// 			if err != nil {
// 				return nil, fmt.Errorf("failed to parse total: %v", err)
// 			}
// 			if currentEvent != nil {
// 				currentEvent.Total = total
// 			}
// 		}
// 	}

// 	if currentEvent != nil {
// 		events = append(events, *currentEvent)
// 	}

// 	return events, nil
// }

// func StringIntoEvents(input string) ([]Event, error) {
// 	events, err := extractEventData(input)
// 	if err != nil {
// 		fmt.Printf("Error: %v\n", err)
// 		return []Event{}, err
// 	}

// 	for _, event := range events {
// 		fmt.Printf("What: %s, Who: %s, Total: %d\n", event.EventType, event.Who, event.Total)
// 	}
// 	return events, nil
// }

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
