// SPDX-License-Identifier: Apache-2.0

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

// // pollTime returns the current time of the blockchain.
// func (t *Timeout) pollTime() (time.Time, error) {
// 	key, err := t.storage.BuildKey("Timestamp", "Now")
// 	if err != nil {
// 		return time.Unix(0, 0), err
// 	}
// 	_now, err := t.storage.QueryOne(0, key)
// 	if err != nil {
// 		return time.Unix(0, 0), err
// 	}

// 	var now TimePoint
// 	if err := types.DecodeFromBytes(_now.StorageData, &now); err != nil {
// 		return time.Unix(0, 0), err
// 	}
// 	unixNow := time.Unix(int64(now/1000), 0)
// 	t.Log().Tracef("Polled time: %v", unixNow.UTC())
// 	return unixNow, nil
// }
