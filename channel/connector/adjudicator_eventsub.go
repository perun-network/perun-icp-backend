package connector

import (
	"encoding/hex"
	"fmt"

	"github.com/pkg/errors"

	"perun.network/go-perun/log"
	pwallet "perun.network/go-perun/wallet"
	"perun.network/perun-icp-backend/channel/connector/icperun"
	pkgsync "polycry.pt/poly-go/sync"

	"regexp"
	"strconv"
	"strings"
	"time"

	pchannel "perun.network/go-perun/channel"
)

// make a predecessor to the real go-perun AdjEvent interface
type AdjEvent interface {
	SetData(cid pchannel.ID, version uint64, finalized bool, alloc [2]uint64, timeout, timestamp uint64) error
	ID() pchannel.ID
	Timeout() pchannel.Timeout
	Version() uint64
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

	//zero := ConcludedEvent{}
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

	fmt.Println("Input in parseEvents: ", input)

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

		fmt.Println("Concluded Event in parsing: ", event)

		adjEvents = append(adjEvents, &event)
	}
	for _, event := range dispEvents {
		fmt.Println("Disputed Event in parsing: ", event)

		adjEvents = append(adjEvents, &event)
	}

	return adjEvents, nil

}

func parseAdjEvents(input string, event AdjEvent,
	channelIDPattern, versionPattern, finalizedPattern, allocPattern, timeoutPattern, timestampPattern string) error {

	rptChannelID := regexp.MustCompile(channelIDPattern)
	rptVersion := regexp.MustCompile(versionPattern)
	rptFinalized := regexp.MustCompile(finalizedPattern)
	rptAlloc := regexp.MustCompile(allocPattern)
	rptTimeout := regexp.MustCompile(timeoutPattern)
	rptTimestamp := regexp.MustCompile(timestampPattern)

	matchesChannelID := rptChannelID.FindAllStringSubmatch(input, -1)
	matchesVersion := rptVersion.FindStringSubmatch(input)
	matchesFinalized := rptFinalized.FindStringSubmatch(input)
	matchesAlloc := rptAlloc.FindStringSubmatch(input)
	matchesTimeout := rptTimeout.FindStringSubmatch(input)
	matchesTimestamp := rptTimestamp.FindStringSubmatch(input)

	if matchesChannelID == nil || matchesVersion == nil || matchesFinalized == nil || matchesAlloc == nil || matchesTimeout == nil || matchesTimestamp == nil {
		fmt.Println("No match found")
		return nil
	}

	if len(matchesChannelID) != 1 {
		return errors.New("multiple matches found for channelID")
	}

	var cid pchannel.ID

	for _, match := range matchesChannelID {

		fmt.Println("match in parseFunbction: ", match)
		byteString, err := hex.DecodeString(match[1])
		if err != nil {
			return err
		}
		fmt.Println("byteString in parseFunbction: ", byteString)
		copy(cid[:], byteString)
	}

	version, err := strconv.ParseUint(matchesVersion[1], 10, 64)
	if err != nil {
		return err
	}

	finalized, err := strconv.ParseBool(matchesFinalized[1])
	if err != nil {
		return err
	}

	// Remove the underscore from the alloc matches and convert them to uint64
	allocStr1 := strings.Replace(matchesAlloc[1], "_", "", -1)
	allocStr2 := strings.Replace(matchesAlloc[2], "_", "", -1)

	alloc1, err := strconv.ParseUint(allocStr1, 10, 64)
	if err != nil {
		return err
	}

	alloc2, err := strconv.ParseUint(allocStr2, 10, 64)
	if err != nil {
		return err
	}

	timeout, err := strconv.ParseUint(matchesTimeout[1], 10, 64)
	if err != nil {
		return err
	}

	timestamp, err := strconv.ParseUint(matchesTimestamp[1], 10, 64)
	if err != nil {
		return err
	}

	err = event.SetData(cid, version, finalized, [2]uint64{alloc1, alloc2}, timeout, timestamp) //timeout
	if err != nil {
		return err
	}
	fmt.Println("Inside parseAdjEvents, event: ", event)
	return nil
}

func NewAdjEventSub(addr pwallet.Address, chanID [32]byte, starttime uint64, req pchannel.AdjudicatorReq, conn *Connector) (*AdjEventSub, error) {
	//userIdx := req.Idx
	a := conn.PerunAgent

	queryArgs := icperun.ChannelTime{
		Channel:   chanID,
		Timestamp: starttime,
	}

	fmt.Println("Inside NewConcludeEventSub, queryArgs: ", queryArgs)

	return &AdjEventSub{
		agent:     a,
		cid:       chanID,
		queryArgs: queryArgs,
		log:       log.MakeEmbedding(log.Default()),
		closer:    new(pkgsync.Closer),
	}, nil
}

func EvaluateConcludedEvents(events []ConcludedEvent) (bool, error) {
	// Assert that the length of events is 1

	if len(events) == 0 {
		fmt.Println("No Conclude events found")
		return false, nil
	}

	if len(events) > 1 {
		return false, fmt.Errorf("Expected only one Concluded event, but got %d", len(events))
	}

	// Check if the event's timestamp is in the past
	eventTime := events[0].Timestamp
	nowTime := uint64(time.Now().UnixNano())
	fmt.Println("Event time: ", eventTime, "Now: ", nowTime)
	if eventTime > nowTime {
		return false, fmt.Errorf("Invalid timestamp: the channel conclusion is in the future")
	}

	fmt.Printf("Timestamp is valid, channel has been already concluded at %d\n", eventTime)
	return true, nil
}

func (e *ConcludedEvent) SetData(cid pchannel.ID, version uint64, finalized bool, alloc [2]uint64, timeout, timestamp uint64) error {

	// here we check everything and make the event for output
	e.IDV = cid
	e.VersionV = version
	e.Finalized = finalized
	e.Alloc = alloc
	e.Tout = timeout
	e.Timestamp = timestamp
	e.concluded = true

	fmt.Printf("cid: %v\n", e.IDV)
	fmt.Printf("Vs: %v\n", e.VersionV)
	fmt.Printf("Finalized: %v\n", e.Finalized)
	fmt.Printf("Alloc: %v\n", e.Alloc)
	fmt.Printf("Tout: %v\n", e.Tout)
	fmt.Printf("Timestamp: %v\n", e.Timestamp)
	fmt.Printf("concluded: %v\n", e.concluded)

	return nil

}

func (e *DisputedEvent) SetData(cid pchannel.ID, version uint64, finalized bool, alloc [2]uint64, timeout, timestamp uint64) error {

	e.IDV = cid
	e.VersionV = version
	e.Finalized = finalized
	e.Alloc = alloc
	e.Tout = timeout
	e.Timestamp = timestamp
	e.disputed = true

	fmt.Printf("cid: %v\n", e.IDV)
	fmt.Printf("Vs: %v\n", e.VersionV)
	fmt.Printf("Finalized: %v\n", e.Finalized)
	fmt.Printf("Alloc: %v\n", e.Alloc)
	fmt.Printf("Tout: %v\n", e.Tout)
	fmt.Printf("Timestamp: %v\n", e.Timestamp)
	fmt.Printf("disputed: %v\n", e.disputed)

	return nil
}
