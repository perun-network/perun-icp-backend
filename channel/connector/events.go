// SPDX-License-Identifier: Apache-2.0
package connector

import (
	"context"
	"fmt"
	"os/exec"
	"time"

	"github.com/aviate-labs/agent-go/principal"
	"perun.network/go-perun/log"
)

const ChanBuffSize = 1024

type FundedEvent struct {
	Fid     FundingID
	Balance Balance // total deposit of the Fid
}

func EventIsDeposited(e PerunEvent) bool {
	_, ok := e.(*FundedEvent)
	return ok
}

// PerunEvent is a Perun event.
type PerunEvent interface{}

type DepositedEvent struct {
	Funded bool
}

type EventSource struct {
	Embedding log.Embedding
	EventChan chan Event
	ErrorChan chan error
}

func NewEventSource() *EventSource {
	return &EventSource{
		EventChan: make(chan Event, ChanBuffSize),
		ErrorChan: make(chan error, 1),
	}
}

func (s *EventSource) Events() <-chan Event {
	return s.EventChan
}

// Err returns the error channel. Will be closed when the EventSource is closed.
func (s *EventSource) Err() <-chan error {
	return s.ErrorChan
}

func queryEventsCLI(queryEventsArgs string, canID principal.Principal, execPath ExecPath) (string, error) {
	// Query the state of the Perun canister

	path, err := exec.LookPath("dfx")
	if err != nil {
		return "", fmt.Errorf("unable to find 'dfx' executable in the system PATH: %w", err)
	}

	canIDStr := canID.Encode()

	output, err := execCanisterCommand(path, canIDStr, "query_events", queryEventsArgs, execPath)
	if err != nil {
		return "", fmt.Errorf("failed to query events: %w", err)
	}

	return string(output), nil
}

func (c *Connector) QueryEventsCLI(queryEventsArgs string, canID principal.Principal, execPath ExecPath) (string, error) {
	return queryEventsCLI(queryEventsArgs, canID, execPath)
}

func ListenEvents(ctx context.Context, queryEventsFunc func(string, string, string) (string, error), stringIntoEventsFunc func(string) ([]Event, error), queryEventsArgs, canID, execPath string, queryFrequency time.Duration, eventsChan chan<- Event) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				newEventsString, err := queryEventsFunc(queryEventsArgs, canID, execPath)
				if err != nil {
					fmt.Printf("Error querying events: %v\n", err)
					time.Sleep(queryFrequency)
					continue
				}
				newEvents, err := stringIntoEventsFunc(newEventsString)
				if err != nil {
					fmt.Printf("Error converting string to events: %v\n", err)
					time.Sleep(queryFrequency)
					continue
				}

				for _, event := range newEvents {
					fmt.Println("New Event in ListenEvents: ", event)
					eventsChan <- event
				}
				time.Sleep(queryFrequency) //  interval between querying for events
			}
		}
	}()
}
