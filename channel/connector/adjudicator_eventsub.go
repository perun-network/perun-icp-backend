package connector

import (
	"encoding/hex"
	"fmt"

	//"github.com/pkg/errors"

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

func parseAdjEvents(input string, event AdjEvent,
	channelIDPattern, versionPattern, finalizedPattern, allocPattern, timeoutPattern, timestampPattern string) error {

	rptChannelID := regexp.MustCompile(channelIDPattern)
	rptVersion := regexp.MustCompile(versionPattern)
	rptFinalized := regexp.MustCompile(finalizedPattern)
	rptAlloc := regexp.MustCompile(allocPattern)
	rptTimeout := regexp.MustCompile(timeoutPattern)
	rptTimestamp := regexp.MustCompile(timestampPattern)

	matchesChannelID := rptChannelID.FindAllStringSubmatch(input, -1)
	matchesVersion := rptVersion.FindAllStringSubmatch(input, -1)
	matchesFinalized := rptFinalized.FindAllStringSubmatch(input, -1)
	matchesAlloc := rptAlloc.FindAllStringSubmatch(input, -1)
	matchesTimeout := rptTimeout.FindAllStringSubmatch(input, -1)
	matchesTimestamp := rptTimestamp.FindAllStringSubmatch(input, -1)

	if matchesChannelID == nil || matchesVersion == nil || matchesFinalized == nil || matchesAlloc == nil || matchesTimeout == nil || matchesTimestamp == nil {
		return nil
	}

	var maxVersionIdx int

	if len(matchesChannelID) == 0 {
		return nil
	}

	if len(matchesChannelID) == 1 {
		maxVersionIdx = 0
	}

	if len(matchesChannelID) != 1 {
		highestVersion := uint64(0)
		maxVersionIdx = -1

		for i := 0; i < len(matchesChannelID); i++ {
			vers, err := strconv.ParseUint(matchesVersion[i][1], 10, 64) // assuming matchesVersion[i] is the correct way to access the version
			if err != nil {
				return err
			}

			if vers > highestVersion {
				highestVersion = vers
				maxVersionIdx = i
			}
		}
	}

	var cid pchannel.ID

	byteString, err := hex.DecodeString(matchesChannelID[maxVersionIdx][1])
	if err != nil {
		return err
	}
	copy(cid[:], byteString)
	//}

	version, err := strconv.ParseUint(matchesVersion[maxVersionIdx][1], 10, 64)
	if err != nil {
		return err
	}

	finalized, err := strconv.ParseBool(matchesFinalized[maxVersionIdx][1])
	if err != nil {
		return err
	}

	// Remove the underscore from the alloc matches and convert them to uint64
	allocStr1 := strings.Replace(matchesAlloc[maxVersionIdx][1], "_", "", -1)
	allocStr2 := strings.Replace(matchesAlloc[maxVersionIdx][2], "_", "", -1)

	alloc1, err := strconv.ParseUint(allocStr1, 10, 64)
	if err != nil {
		return err
	}

	alloc2, err := strconv.ParseUint(allocStr2, 10, 64)
	if err != nil {
		return err
	}

	timeout, err := strconv.ParseUint(matchesTimeout[maxVersionIdx][1], 10, 64)
	if err != nil {
		return err
	}

	timestamp, err := strconv.ParseUint(matchesTimestamp[maxVersionIdx][1], 10, 64)
	if err != nil {
		return err
	}

	err = event.SetData(cid, version, finalized, [2]uint64{alloc1, alloc2}, timeout, timestamp) //timeout
	if err != nil {
		return err
	}
	return nil
}

func NewAdjEventSub(addr pwallet.Address, chanID [32]byte, starttime uint64, req pchannel.AdjudicatorReq, conn *Connector) (*AdjEventSub, error) {
	//userIdx := req.Idx
	a := conn.PerunAgent

	queryArgs := icperun.ChannelTime{
		Channel:   chanID,
		Timestamp: starttime,
	}

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
		return false, nil
	}

	if len(events) > 1 {
		return false, fmt.Errorf("Expected only one Concluded event, but got %d", len(events))
	}

	// Check if the event's timestamp is in the past
	eventTime := events[0].Timestamp
	nowTime := uint64(time.Now().UnixNano())
	if eventTime > nowTime {
		return false, fmt.Errorf("Invalid timestamp: the channel conclusion is in the future")
	}

	return true, nil
}

func (e *ConcludedEvent) SetData(cid ChannelID, version Version, finalized bool, alloc [2]uint64, timeout, timestamp uint64) error {

	// here we check everything and make the event for output
	e.IDV = cid
	e.VersionV = version
	e.Finalized = finalized
	e.Alloc = alloc
	e.Tout = timeout
	e.Timestamp = timestamp
	e.concluded = true

	return nil

}

func (e *DisputedEvent) SetData(cid ChannelID, version Version, finalized bool, alloc [2]uint64, timeout, timestamp uint64) error {

	e.IDV = cid
	e.VersionV = version
	e.Finalized = finalized
	e.Alloc = alloc
	e.Tout = timeout
	e.Timestamp = timestamp
	e.disputed = true

	return nil
}
