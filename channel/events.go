package channel

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

type EventSub struct {
	//types.EventSub
}

type EventArchive struct {
	Events []Event
}

type EventRecords struct {
	//types.EventRecords

	PerunModule_Funded []FundedEvent // nolint: stylecheck
	//PerunModule_Concluded []ConcludedEvent // nolint: stylecheck
	//PerunModule_Disputed  []DisputedEvent  // nolint: stylecheck
}

// PerunEvent is a Perun event.
type PerunEvent interface{}

type FundedEvent struct {
	Funded bool
}

type Event struct {
	EventType string
	Who       string
	Total     uint64
}

type EventHandler struct {
	EventChans chan []Event
}

func subscribeToFundingEvents() error {

	return fmt.Errorf("not implemented")
}

func extractEventData(input string) ([]Event, error) {
	lines := strings.Split(input, "\n")
	var events []Event
	var currentEvent *Event

	for _, line := range lines {
		line = strings.TrimSpace(line)

		eventTypes := []string{"Funded", "Disputed", "Concluded"}
		for _, eventType := range eventTypes {
			if strings.Contains(line, eventType) {
				if currentEvent != nil {
					events = append(events, *currentEvent)
				}
				currentEvent = &Event{EventType: eventType}
				break
			}
		}

		startIdx := strings.Index(line, "blob \"")
		endIdx := strings.Index(line, "\";")
		if startIdx != -1 && endIdx != -1 && startIdx < endIdx {
			who := line[startIdx+len("blob \"") : endIdx]
			if currentEvent != nil {
				currentEvent.Who = who
			}
		}

		totalStartIdx := strings.Index(line, "total =")
		if totalStartIdx != -1 {
			totalEndIdx := strings.Index(line, " : nat;")
			cleanedNumber := strings.ReplaceAll(line[totalStartIdx+len("total ="):totalEndIdx], "_", "")
			cleanedNumber = strings.TrimSpace(cleanedNumber)
			total, err := strconv.ParseUint(cleanedNumber, 10, 64)
			if err != nil {
				return nil, fmt.Errorf("failed to parse total: %v", err)
			}
			if currentEvent != nil {
				currentEvent.Total = total
			}
		}
	}

	if currentEvent != nil {
		events = append(events, *currentEvent)
	}

	return events, nil
}

func StringIntoEvents(input string) ([]Event, error) {
	events, err := extractEventData(input)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return []Event{}, err
	}

	for _, event := range events {
		fmt.Printf("What: %s, Who: %s, Total: %d\n", event.EventType, event.Who, event.Total)
	}
	return events, nil
}

func queryEventsCLI(queryEventsArgs string, canID string, execPath string) (string, error) {
	// Query the state of the Perun canister

	path, err := exec.LookPath("dfx")
	if err != nil {
		return "", fmt.Errorf("unable to find 'dfx' executable in the system PATH: %w", err)
	}

	txCmd := exec.Command(path, "canister", "call", canID, "query_events", queryEventsArgs)
	txCmd.Dir = execPath
	output, err := txCmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to query canister events: %w\nOutput: %s", err, output)
	}

	return string(output), nil
}

func QueryEventsCLI(queryEventsArgs string, canID string, execPath string) (string, error) {
	return queryEventsCLI(queryEventsArgs, canID, execPath)
}

// func ListenEvents(queryEventsArgs, canID, execPath string, queryFrequency time.Duration, eventsChan chan<- Event) {
// 	go func() {
// 		for {
// 			newEventsString, err := queryEventsCLI(queryEventsArgs, canID, execPath)
// 			if err != nil {
// 				fmt.Printf("Error querying events: %v\n", err)
// 				time.Sleep(queryFrequency)
// 				continue
// 			}

// 			newEvents, err := StringIntoEvents(newEventsString)
// 			if err != nil {
// 				fmt.Printf("Error converting string to events: %v\n", err)
// 				time.Sleep(queryFrequency)
// 				continue
// 			}

// 			for _, event := range newEvents {
// 				eventsChan <- event
// 			}
// 			time.Sleep(queryFrequency) // Use the input parameter for the interval between querying for events
// 		}
// 	}()
// }

func ListenEvents(queryEventsFunc func(string, string, string) (string, error), stringIntoEventsFunc func(string) ([]Event, error), queryEventsArgs, canID, execPath string, queryFrequency time.Duration, eventsChan chan<- Event) {
	go func() {
		for {
			newEventsString, err := queryEventsFunc(queryEventsArgs, canID, execPath)
			// ...
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
				eventsChan <- event
			}
			time.Sleep(queryFrequency) // Use the input parameter for the interval between querying for events
		}
	}()
}
