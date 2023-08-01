// SPDX-License-Identifier: Apache-2.0
package connector_test

import (
	"math/rand"
	"time"
)

func mockQueryEventsCLI(queryEventsArgs, canID, execPath string) (string, error) {
	// Generate a random sleep duration between 10ms and 200ms
	sleepDuration := time.Duration(rand.Intn(190)+10) * time.Millisecond
	time.Sleep(sleepDuration)
	return "event1;event2;event3", nil
}

// func mockStringIntoEvents(input string) ([]chanconn.Event, error) {
// 	events := make([]chanconn.Event, 0)

// 	for i, s := range strings.Split(input, ";") {
// 		if s == "error" {
// 			return nil, errors.New("conversion error")
// 		}

// 		identifier := fmt.Sprintf("User%d", rand.Intn(100))
// 		amount := uint64(rand.Intn(100) + 1)

// 		events = append(events, chanconn.Event{
// 			EventType: fmt.Sprintf("event%d", i+1),
// 			Who:       identifier,
// 			Total:     amount,
// 		})
// 	}

// 	return events, nil
// }
