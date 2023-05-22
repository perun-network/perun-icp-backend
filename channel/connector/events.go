// SPDX-License-Identifier: Apache-2.0
package connector

import (
	"context"
	"fmt"
	"os/exec"
	pkgsync "polycry.pt/poly-go/sync"
	"time"

	"github.com/aviate-labs/agent-go/principal"
	pchannel "perun.network/go-perun/channel"
	"perun.network/go-perun/log"
)

const ChanBuffSize = 1024

func EventIsDeposited(e PerunEvent) bool {
	_, ok := e.(*FundedEvent)
	return ok
}

type DepositedEvent struct {
	Funded bool
}

type (
	// EventSub listens on events and can filter them with an EventPredicate.
	EventSub struct {
		*pkgsync.Closer
		log.Embedding

		source  *EventSource
		p       EventPredicate
		sink    chan PerunEvent
		errChan chan error
	}

	// EventPredicate can be used to filter events.
	EventPredicate func(PerunEvent) bool
	// PerunEvent is a Perun event.
	PerunEvent interface {
		ID() pchannel.ID
		Timeout() pchannel.Timeout
		Version() uint64
	}
	// DisputedEvent is emitted when a dispute was opened or updated.
	DisputedEvent struct {
		Cid   ChannelID
		State State
		Vs    uint64
	}

	FundedEvent struct {
		Cid     FundingID
		Balance Balance
		Vs      uint64
	}

	// ConcludedEvent is emitted when a channel is concluded.
	ConcludedEvent struct {
		Cid ChannelID
		Vs  uint64
	}
)

type EventSource struct {
	Embedding log.Embedding
	EventChan chan Event
	ErrorChan chan error
}

func (c *FundedEvent) Timeout() pchannel.Timeout {
	when := time.Now().Add(10 * time.Second)
	pollInterval := 1 * time.Second
	return NewTimeout(when, pollInterval)
}
func (c *FundedEvent) ID() pchannel.ID {
	return c.Cid
}
func (c *FundedEvent) Version() uint64 {
	return c.Vs
}

func (c *DisputedEvent) Timeout() pchannel.Timeout {
	when := time.Now().Add(10 * time.Second)
	pollInterval := 1 * time.Second
	return NewTimeout(when, pollInterval)
}
func (c *DisputedEvent) ID() pchannel.ID {
	return c.Cid
}
func (c *DisputedEvent) Version() uint64 {
	return c.Vs
}

func (c *ConcludedEvent) Timeout() pchannel.Timeout {
	when := time.Now().Add(10 * time.Second)
	pollInterval := 1 * time.Second
	return NewTimeout(when, pollInterval)
}
func (c *ConcludedEvent) ID() pchannel.ID {
	return c.Cid
}
func (c *ConcludedEvent) Version() uint64 {
	return c.Vs
}

var _ pchannel.AdjudicatorEvent = &FundedEvent{}
var _ pchannel.AdjudicatorEvent = &ConcludedEvent{}
var _ pchannel.AdjudicatorEvent = &DisputedEvent{}

// EventIsDisputed checks whether an event is a DisputedEvent for a
// specific channel.
func EventIsDisputed(cid ChannelID) func(PerunEvent) bool {
	return func(e PerunEvent) bool {
		event, ok := e.(*DisputedEvent)
		return ok && event.Cid == cid
	}
}

// EventIsConcluded checks whether an event is a ConcludedEvent for a
// specific channel.
func EventIsConcluded(cid ChannelID) func(PerunEvent) bool {
	return func(e PerunEvent) bool {
		event, ok := e.(*ConcludedEvent)
		return ok && event.Cid == cid
	}
}

func NewEventSub() *EventSub {
	return &EventSub{
		Closer:    new(pkgsync.Closer),
		Embedding: log.MakeEmbedding(log.Default()),
		source:    NewEventSource(),
		sink:      make(chan PerunEvent, ChanBuffSize),
		errChan:   make(chan error, 1),
	}
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

// Events returns the channel that contains all Perun events.
// Will never be closed.
func (p *EventSub) Events() <-chan PerunEvent {
	return p.sink
}

// Err returns the error channel. Will be closed when the subscription is closed.
func (p *EventSub) Err() <-chan error {
	return p.errChan
}
