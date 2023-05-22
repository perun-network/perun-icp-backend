package connector

import (
	"fmt"
	pchannel "perun.network/go-perun/channel"
	"perun.network/go-perun/log"
	pkgsync "polycry.pt/poly-go/sync"
	"time"
)

// AdjudicatorSub implements the AdjudicatorSubscription interface.
type AdjudicatorSub struct {
	*pkgsync.Closer
	log.Embedding

	cid  ChannelID
	sub  *EventSub
	conn *Connector
	//storage substrate.StorageQueryer
	err chan error
}

// NewAdjudicatorSub returns a new AdjudicatorSub. Will return all events from
// the `pastBlocks` past blocks and all events from future blocks.
func NewAdjudicatorSub(cid ChannelID, conn *Connector) (*AdjudicatorSub, error) {
	sub, err := conn.Subscribe(isAdjEvent(cid))
	if err != nil {
		return nil, err
	}

	ret := &AdjudicatorSub{new(pkgsync.Closer), log.MakeEmbedding(log.Default()), cid, sub, conn, make(chan error, 1)}
	// ret.OnCloseAlways(func() {
	// 	if err := ret.sub.Close(); err != nil {
	// 		ret.Log().WithError(err).Error("Could not close Closer.")
	// 	}
	// 	close(ret.err)
	// })
	return ret, nil
}

// isAdjEvent returns a predicate that decides whether or not an event is
// relevant for the adjudicator and concerns a specific channel.
func isAdjEvent(cid ChannelID) EventPredicate {
	return func(e PerunEvent) bool {
		return EventIsDisputed(cid)(e) || EventIsConcluded(cid)(e)
	}
}

// Next implements the AdjudicatorSub.Next function.
func (s *AdjudicatorSub) Next() pchannel.AdjudicatorEvent {
	if s.IsClosed() {
		return nil
	}
	// Wait for event or closed.

	select {
	case event := <-s.sub.Events():
		if EventIsDisputed(s.cid)(event) || EventIsConcluded(s.cid)(event) { //EventIsProgressed(s.cid)(event) ||
			return event
		}
	case <-s.Closed():
		return nil
	}
	return nil
}

// Err implements the AdjudicatorSub.Err function.
func (s *AdjudicatorSub) Err() error {
	return <-s.err
}

// Subscribe returns an EventSub that listens on all events of the pallet.
func (c *Connector) Subscribe(f EventPredicate) (*EventSub, error) {

	return NewEventSub(), nil
}

// makePerunEvent creates a Perun event from a generic event.
func (s *AdjudicatorSub) makePerunEvent(event PerunEvent) (pchannel.AdjudicatorEvent, error) {
	switch event := event.(type) {
	case *DisputedEvent:

		duration := 10 * time.Second
		seconds := uint64(duration.Seconds())
		return &pchannel.RegisteredEvent{
			AdjudicatorEventBase: pchannel.AdjudicatorEventBase{
				IDV:      event.Cid,
				VersionV: event.State.Version,
				//TimeoutV: MakeTimeout(event.Timeout),
				TimeoutV: MakeTimeout(seconds),
			},
			State: nil, // only needed for virtual channel support
			Sigs:  nil, // only needed for virtual channel support
		}, nil

	case *ConcludedEvent:

		duration := 10 * time.Second
		seconds := uint64(duration.Seconds())
		return &pchannel.ConcludedEvent{
			AdjudicatorEventBase: pchannel.AdjudicatorEventBase{
				IDV:      event.Cid,
				VersionV: 0,
				TimeoutV: MakeTimeout(seconds),
			},
		}, nil
	default:
		panic(fmt.Sprintf("unknown event: %#v", event))
	}
}
