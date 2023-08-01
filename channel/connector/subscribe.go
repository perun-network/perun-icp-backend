package connector

import (
	"context"
	"fmt"
	"github.com/pkg/errors"

	pchannel "perun.network/go-perun/channel"
	"perun.network/perun-icp-backend/channel/connector/icperun"
	pkgsync "polycry.pt/poly-go/sync"

	"perun.network/go-perun/log"
	"time"
)

const (
	DefaultBufferSize                  = 3
	DefaultSubscriptionPollingInterval = time.Duration(4) * time.Second
)

// AdjudicatorSub implements the AdjudicatorSubscription interface.
type AdjEventSub struct {
	agent        *icperun.Agent
	queryArgs    icperun.ChannelTime
	log          log.Embedding
	cid          ChannelID
	events       chan AdjEvent
	Ev           []AdjEvent
	err          error
	panicErr     chan error
	cancel       context.CancelFunc
	closer       *pkgsync.Closer
	pollInterval time.Duration
}

type EventSource struct {
	sink chan AdjEvent
}

func (e *EventSource) Events() <-chan AdjEvent {
	return e.sink
}

func (e *AdjEventSub) GetEvents() <-chan AdjEvent {
	return e.events
}

func NewAdjudicatorSub(ctx context.Context, cid pchannel.ID, conn *Connector) *AdjEventSub {

	queryArgs := icperun.ChannelTime{
		Channel:   cid,
		Timestamp: uint64(time.Now().Unix())}

	sub := &AdjEventSub{
		queryArgs:    queryArgs,
		agent:        conn.PerunAgent,
		events:       make(chan AdjEvent, DefaultBufferSize),
		Ev:           make([]AdjEvent, 0),
		panicErr:     make(chan error, 1),
		pollInterval: DefaultSubscriptionPollingInterval,
		closer:       new(pkgsync.Closer),
		log:          log.MakeEmbedding(log.Default()),
	}

	ctx, sub.cancel = context.WithCancel(ctx)
	go sub.run(ctx)
	return sub

}

func (s *AdjEventSub) run(ctx context.Context) {
	fmt.Println("inside adjeventsub run")
	s.log.Log().Info("EventSource Listening started from start time")
	finish := func(err error) {
		s.err = err
		close(s.events)

	}
polling:
	for {
		s.log.Log().Debug("Inside EventSource.run loop")
		select {
		case <-ctx.Done():
			finish(nil)
			return
		case <-s.events:
			finish(nil)
			return
		case <-time.After(s.pollInterval):
			eventStr, err := s.QueryEvents()
			fmt.Println("eventStr in run(ctx)", eventStr)
			if err != nil {
				// QueryEvents should be executable correctly, so we abort the conclusion subscription
				s.panicErr <- err
			}

			// Check if eventStr is empty
			if eventStr == "" {
				s.log.Log().Debug("No events yet, continuing polling...")
				continue polling

				// here TODO implement elseif for a funded event

			} else {
				s.log.Log().Debug("Event detected, evaluating events...")

				// Parse the events

				adjEvents, err := parseEventsAll(eventStr) //Concluded(eventStr)
				if err != nil {
					s.panicErr <- errors.Wrap(err, "failed to parse events during polling")
				}

				if len(adjEvents) == 0 {
					s.log.Log().Warn("No events detected but, continuing polling...")
					continue polling
				}

				s.log.Log().Debugf("Parsed events: %v", adjEvents)

				for _, event := range adjEvents {
					fmt.Println("adjevent in run(ctx)", event)
					s.events <- event
					//s.Ev = append(s.Ev, event)
				}

				s.log.Log().Infof("Found new event/s")
				return
			}
		}
	}
}

// Next implements the AdjudicatorSub.Next function.
func (s *AdjEventSub) Next() pchannel.AdjudicatorEvent {
	if s.closer.IsClosed() {
		return nil
	}

	if s.Events() == nil {
		return nil
	}
	// Wait for event or closed.
	select {
	case event := <-s.Events():
		if event == nil {
			return nil
		}

		fmt.Println("event in Next()", event)
		timestamp := event.Tstamp()

		switch e := event.(type) {
		case *DisputedEvent:

			dispEvent := pchannel.AdjudicatorEventBase{
				VersionV: event.Version(),
				IDV:      event.ID(),
				TimeoutV: MakeTimeout(timestamp),
			}

			ddn := &pchannel.RegisteredEvent{AdjudicatorEventBase: dispEvent,
				State: nil,
				Sigs:  nil}
			s.closer.Close()
			return ddn
		case *ConcludedEvent:
			fmt.Println("Got a ConcludedEvent: ", e)
			conclEvent := pchannel.AdjudicatorEventBase{
				VersionV: event.Version(),
				IDV:      event.ID(),
				TimeoutV: MakeTimeout(timestamp),
			}
			ccn := &pchannel.ConcludedEvent{
				AdjudicatorEventBase: conclEvent,
			}
			s.closer.Close()
			return ccn
		default:
			// Handle other cases here or do nothing
			fmt.Println("Got an unknown event type")
			return nil
		}

	case <-s.closer.Closed():
		return nil
	}
}

func (s *AdjEventSub) Close() error {
	s.cancel()
	return nil
}
