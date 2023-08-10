package connector

import (
	"fmt"
	pchannel "perun.network/go-perun/channel"
	"perun.network/go-perun/log"
	pwallet "perun.network/go-perun/wallet"
	"perun.network/perun-icp-backend/channel/connector/icperun"
	pkgsync "polycry.pt/poly-go/sync"
	"time"
)

// predecessor to the go-perun AdjEvent interface
type AdjEvent interface {
	SetData(cid pchannel.ID, version uint64, finalized bool, alloc [2]uint64, timeout, timestamp uint64) error
	ID() pchannel.ID
	Timeout() pchannel.Timeout
	Version() Version
	Tstamp() uint64
}

func ParseEventsConcluded(input string) ([]ConcludedEvent, error) {
	return parseEventsConcluded(input)
}

func parseEventsConcluded(input string) ([]ConcludedEvent, error) {
	event := ConcludedEvent{}
	err := parseAdjEvents(input, &event,
		`Conclude_state=ChannelIDStart([a-fA-F0-9]+)ChannelIDEnd`,
		`Conclude_state=VersionStart(\d+)VersionEnd`,
		`Conclude_timeout=FinalizedStart(\w+)FinalizedEnd`,
		`Conclude_alloc=AllocStart(\d+_\d+), (\d+_\d+)AllocEnd`,
		`Conclude_timeout=TimeoutStart(\d+)TimeoutEnd`,
		`Conclude_timestamp=TimestampStart(\d+)TimestampEnd`)
	if err != nil {
		return nil, err
	}

	zeroCID := [32]byte{}
	if event.IDV == zeroCID {
		return []ConcludedEvent{}, nil
	}

	return []ConcludedEvent{event}, nil
}

func parseEventsDisputed(input string) ([]DisputedEvent, error) {
	event := DisputedEvent{}
	err := parseAdjEvents(input, &event,
		`Dispute_state=ChannelIDStart([a-fA-F0-9]+)ChannelIDEnd`,
		`Dispute_state=VersionStart(\d+)VersionEnd`,
		`Dispute_timeout=FinalizedStart(\w+)FinalizedEnd`,
		`Dispute_alloc=AllocStart(\d+_\d+), (\d+_\d+)AllocEnd`,
		`Dispute_timeout=TimeoutStart(\d+)TimeoutEnd`,
		`Dispute_timestamp=TimestampStart(\d+)TimestampEnd`)
	if err != nil {
		return nil, err
	}

	zeroCID := [32]byte{}
	if event.IDV == zeroCID {
		return []DisputedEvent{}, nil
	}

	return []DisputedEvent{event}, nil
}

func parseEventsAll(input string) ([]AdjEvent, error) {

	var concEvents []ConcludedEvent
	var dispEvents []DisputedEvent
	var adjEvents []AdjEvent

	concEvents, err := parseEventsConcluded(input)
	if err != nil {
		return nil, err
	}

	dispEvents, err = parseEventsDisputed(input)
	if err != nil {
		return nil, err
	}

	for _, event := range concEvents {
		adjEvents = append(adjEvents, &event)
	}

	for _, event := range dispEvents {
		adjEvents = append(adjEvents, &event)
	}

	return adjEvents, nil

}

func NewAdjEventSub(addr pwallet.Address, chanID pchannel.ID, starttime uint64, req pchannel.AdjudicatorReq, conn *Connector) (*AdjEventSub, error) {

	queryArgs := icperun.ChannelTime{
		Channel:   chanID,
		Timestamp: starttime,
	}

	return &AdjEventSub{
		agent:     conn.PerunAgent,
		cid:       chanID,
		queryArgs: queryArgs,
		log:       log.MakeEmbedding(log.Default()),
		closer:    new(pkgsync.Closer),
	}, nil
}

func EvaluateConcludedEvents(events []ConcludedEvent) (bool, error) {

	if len(events) == 0 {
		return false, nil
	}

	if len(events) > 1 {
		return false, fmt.Errorf("Expected only one Concluded event, but got %d", len(events))
	}

	eventTime := events[0].Timestamp
	nowTime := uint64(time.Now().UnixNano())
	if eventTime > nowTime {
		return false, fmt.Errorf("Invalid timestamp: the channel conclusion is in the future")
	}

	return true, nil
}

func (e *ConcludedEvent) SetData(cid pchannel.ID, version Version, finalized bool, alloc [2]uint64, timeout, timestamp uint64) error {

	e.IDV = cid
	e.VersionV = version
	e.Finalized = finalized
	e.Alloc = alloc
	e.Tout = timeout
	e.Timestamp = timestamp
	e.concluded = true

	return nil

}

func (e *DisputedEvent) SetData(cid pchannel.ID, version Version, finalized bool, alloc [2]uint64, timeout, timestamp uint64) error {

	e.IDV = cid
	e.VersionV = version
	e.Finalized = finalized
	e.Alloc = alloc
	e.Tout = timeout
	e.Timestamp = timestamp
	e.disputed = true

	return nil
}
