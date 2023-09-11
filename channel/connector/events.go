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
	pchannel "perun.network/go-perun/channel"
	"time"
)

const ChanBuffSize = 1024
const MaxNumIters = 4

type (

	// EventPredicate can be used to filter events.
	EventPredicate func(PerunEvent) bool
	// PerunEvent is a Perun event.
	PerunEvent interface {
		ID() pchannel.ID
		Timeout() pchannel.Timeout
		Version() Version
	}

	DisputedEvent struct {
		Finalized bool
		Alloc     [2]uint64
		Tout      uint64
		Timestamp uint64
		disputed  bool
		VersionV  Version
		IDV       pchannel.ID
	}

	ConcludedEvent struct {
		Finalized bool
		Alloc     [2]uint64
		Tout      uint64
		Timestamp uint64
		concluded bool
		VersionV  Version
		IDV       pchannel.ID
	}

	FundedEvent struct {
		Cid     FundingID
		Balance Balance
		Vs      Version
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
func (c *FundedEvent) Version() Version {
	return c.Vs
}

func (d *DisputedEvent) Timeout() pchannel.Timeout {
	when := time.Now().Add(10 * time.Second)
	pollInterval := 1 * time.Second
	return NewTimeout(when, pollInterval)
}
func (d *DisputedEvent) ID() pchannel.ID {
	return d.IDV
}
func (d *DisputedEvent) Version() Version {
	return d.VersionV
}
func (d *DisputedEvent) Tstamp() Version {
	return d.Timestamp
}

func (c *ConcludedEvent) Timeout() pchannel.Timeout {
	when := time.Now().Add(10 * time.Second)
	pollInterval := 1 * time.Second
	return NewTimeout(when, pollInterval)
}
func (c *ConcludedEvent) ID() pchannel.ID {
	return c.IDV
}
func (c *ConcludedEvent) Version() Version {
	return c.VersionV
}
func (c *ConcludedEvent) Tstamp() uint64 {
	return c.Timestamp
}

func (s *AdjEventSub) Events() <-chan AdjEvent {
	return s.events
}

func (s *AdjEventSub) Err() error {
	return s.err
}

func (s *AdjEventSub) PanicErr() <-chan error {
	return s.panicErr
}
