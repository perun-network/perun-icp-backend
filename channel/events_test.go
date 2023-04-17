package channel_test

import (
	"errors"
	"testing"
	"time"

	"perun.network/perun-icp-backend/channel"
)

func mockQueryEventsCLI(queryEventsArgs, canID, execPath string) (string, error) {
	return "event1;event2;event3", nil
}

func mockStringIntoEvents(input string) ([]channel.Event, error) {
	if input == "error" {
		return nil, errors.New("conversion error")
	}

	return []channel.Event{
		{EventType: "event1", Who: "Alice", Total: 10},
		{EventType: "event2", Who: "Bob", Total: 20},
		{EventType: "event3", Who: "Carol", Total: 30},
	}, nil
}

func TestListenEvents(t *testing.T) {
	queryEventsArgs := "mock_query_events_args"
	canID := "mock_canister_id"
	execPath := "mock_executable_path"
	queryFrequency := 100 * time.Millisecond

	eventsChan := make(chan channel.Event)

	channel.ListenEvents(mockQueryEventsCLI, mockStringIntoEvents, queryEventsArgs, canID, execPath, queryFrequency, eventsChan)

	timeout := time.After(1 * time.Second)
	eventCount := 0

	for {
		select {
		case event := <-eventsChan:
			t.Logf("Received event: %v", event)
			eventCount++
		case <-timeout:
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
