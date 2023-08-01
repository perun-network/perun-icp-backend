// SPDX-License-Identifier: Apache-2.0
package connector

import (
	pchannel "perun.network/go-perun/channel"
	"time"
)

const ChanBuffSize = 1024
const MaxNumIters = 4

type DepositedEvent struct {
	Funded bool
}

type (

	// EventPredicate can be used to filter events.
	EventPredicate func(PerunEvent) bool
	// PerunEvent is a Perun event.
	PerunEvent interface {
		ID() pchannel.ID
		Timeout() pchannel.Timeout
		Version() uint64
	}

	DisputedEvent struct {
		State     State
		Finalized bool
		Alloc     [2]uint64
		Tout      uint64
		Timestamp uint64
		disputed  bool
		VersionV  uint64
		IDV       pchannel.ID
	}

	ConcludedEvent struct {
		Finalized bool
		Alloc     [2]uint64
		Tout      uint64
		Timestamp uint64
		State     State
		concluded bool
		VersionV  uint64
		IDV       pchannel.ID
	}

	FundedEvent struct {
		Cid     FundingID
		Balance Balance
		Vs      uint64
	}
)

func (e *AdjEventSub) QueryEvents() (string, error) {
	return e.agent.QueryEvents(e.queryArgs)
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

func (c DisputedEvent) Timeout() pchannel.Timeout {
	when := time.Now().Add(10 * time.Second)
	pollInterval := 1 * time.Second
	return NewTimeout(when, pollInterval)
}
func (c DisputedEvent) ID() pchannel.ID {
	return c.IDV
}
func (c DisputedEvent) Version() uint64 {
	return c.VersionV
}
func (c DisputedEvent) Tstamp() uint64 {
	return c.Timestamp
}

func (c ConcludedEvent) Timeout() pchannel.Timeout {
	when := time.Now().Add(10 * time.Second)
	pollInterval := 1 * time.Second
	return NewTimeout(when, pollInterval)
}
func (c ConcludedEvent) ID() pchannel.ID {
	return c.IDV
}
func (c ConcludedEvent) Version() uint64 {
	return c.VersionV
}
func (c ConcludedEvent) Tstamp() uint64 {
	return c.Timestamp
}

// EventIsDisputed checks whether an event is a DisputedEvent for a
// specific channel.
func EventIsDisputed(cid ChannelID) func(PerunEvent) bool {
	return func(e PerunEvent) bool {
		event, ok := e.(*DisputedEvent)
		return ok && event.IDV == cid
	}
}

// EventIsConcluded checks whether an event is a ConcludedEvent for a
// specific channel.
func EventIsConcluded(cid ChannelID) func(PerunEvent) bool {
	return func(e PerunEvent) bool {
		event, ok := e.(*ConcludedEvent)
		return ok && event.IDV == cid
	}
}

func (s *AdjEventSub) Events() <-chan AdjEvent {
	return s.events
}

// Err returns the error channel. Will be closed when the EventSource is closed.
func (s *AdjEventSub) Err() error {
	return s.err
}

func (s *AdjEventSub) PanicErr() <-chan error {
	return s.panicErr
}
