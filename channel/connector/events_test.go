// SPDX-License-Identifier: Apache-2.0
package connector_test

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"testing"
	"time"

	chanconn "perun.network/perun-icp-backend/channel/connector"
)

func mockQueryEventsCLI(queryEventsArgs, canID, execPath string) (string, error) {
	// Generate a random sleep duration between 10ms and 200ms
	sleepDuration := time.Duration(rand.Intn(190)+10) * time.Millisecond
	time.Sleep(sleepDuration)
	return "event1;event2;event3", nil
}

func mockStringIntoEvents(input string) ([]chanconn.Event, error) {
	events := make([]chanconn.Event, 0)

	for i, s := range strings.Split(input, ";") {
		if s == "error" {
			return nil, errors.New("conversion error")
		}

		identifier := fmt.Sprintf("User%d", rand.Intn(100))
		amount := uint64(rand.Intn(100) + 1)

		events = append(events, chanconn.Event{
			EventType: fmt.Sprintf("event%d", i+1),
			Who:       identifier,
			Total:     amount,
		})
	}

	return events, nil
}

func TestListenEvents(t *testing.T) {
	queryEventsArgs := "mock_query_events_args"
	canID := "mock_canister_id"
	execPath := "mock_executable_path"
	queryFrequency := time.Duration(rand.Intn(200)+100) * time.Millisecond // Randomize query frequency between 100ms and 300ms

	eventsChan := make(chan chanconn.Event)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	chanconn.ListenEvents(ctx, mockQueryEventsCLI, mockStringIntoEvents, queryEventsArgs, canID, execPath, queryFrequency, eventsChan)

	timeout := time.After(4 * time.Second)
	eventCount := 0

	for {
		select {
		case event := <-eventsChan:
			t.Logf("Received event: %v", event)
			eventCount++
		case <-timeout:
			cancel() // Stop the ListenEvents function
			if eventCount < 1 {
				t.Errorf("Expected to receive at least 1 event, but received %d", eventCount)
				t.Logf("Test failed: Received %d events", eventCount)
			} else {
				t.Logf("Test passed: Received %d events", eventCount)
			}
			return
		}
	}
}
