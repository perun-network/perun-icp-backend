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
	"context"
	"time"

	pchannel "perun.network/go-perun/channel"
	"perun.network/go-perun/log"
)

type (
	// ExpiredTimeout is always expired.
	// Implements the Perun Timeout interface.
	ExpiredTimeout struct{}

	// Timeout can be used to wait until a specific timepoint is reached by
	// the blockchain. Implements the Perun Timeout interface.
	Timeout struct {
		log.Embedding

		when         time.Time
		pollInterval time.Duration
	}

	// TimePoint as defined by pallet Timestamp.
	TimePoint uint64
)

// DefaultTimeoutPollInterval default value for the PollInterval of a Timeout.
const DefaultTimeoutPollInterval = 1 * time.Second

// NewExpiredTimeout returns a new ExpiredTimeout.
func NewExpiredTimeout() *ExpiredTimeout {
	return &ExpiredTimeout{}
}

func (*ExpiredTimeout) IsElapsed(context.Context) bool {
	return true
}

// Wait returns nil.
func (*ExpiredTimeout) Wait(context.Context) error {
	return nil
}

// NewTimeout returns a new Timeout which expires at the given time.
func NewTimeout(when time.Time, pollInterval time.Duration) *Timeout {
	return &Timeout{log.MakeEmbedding(log.Default()), when, pollInterval}
}

// IsElapsed returns whether the timeout is elapsed.
func (t *Timeout) IsElapsed(ctx context.Context) bool {
	now := time.Now()

	elapsed := t.when.Before(now) || t.when.Equal(now)

	delta := now.Sub(t.when)
	if elapsed {
		t.Log().Printf("Timeout elapsed since %v", delta)
	} else {
		t.Log().Printf("Timeout target in %v", delta)
	}

	return elapsed
}

// Wait waits for the timeout or until the context is cancelled.
func (t *Timeout) Wait(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(t.pollInterval):
			if t.IsElapsed(ctx) {
				return nil
			}
		}
	}
}

// MakeTimeout creates a new timeout.
func MakeTimeout(sec ChallengeDuration) pchannel.Timeout {
	return NewTimeout(MakeTime(sec), DefaultTimeoutPollInterval)
}

// MakeTime creates a new time from the argument.
func MakeTime(sec ChallengeDuration) time.Time {
	return time.Unix(int64(sec), 0)
}
