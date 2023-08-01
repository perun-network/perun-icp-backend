package icperun_test

import (
	"github.com/aviate-labs/agent-go"
	"perun.network/perun-icp-backend/channel/connector/icperun"

	"github.com/aviate-labs/agent-go/mock"
	"github.com/aviate-labs/agent-go/principal"
	"net/http/httptest"
	"net/url"
	//"testing"
	//"github.com/aviate-labs/agent-go/ic/perun"
)

// Test_TransactionNotification tests the "transaction_notification" method on the "perun" canister.
// func Test_TransactionNotification(t *testing.T) {
// 	a, err := newAgent([]mock.Method{
// 		{
// 			Name:      "transaction_notification",
// 			Arguments: []any{new(uint64)},
// 			Handler: func(request mock.Request) ([]any, error) {
// 				return []any{}, nil
// 			},
// 		},
// 	})
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	var a0 uint64
// 	var r0 int64
// 	r0, err = a.TransactionNotification(a0)
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// }

// newAgent creates a new agent with the given (mock) methods.
// Runs a mock replica in the background.
func newAgent(methods []mock.Method) (*icperun.Agent, error) {
	replica := mock.NewReplica()
	canisterId := principal.Principal{Raw: []byte("perun")}
	replica.AddCanister(canisterId, methods)
	s := httptest.NewServer(replica)
	u, _ := url.Parse(s.URL)
	a, err := icperun.NewAgent(canisterId, agent.Config{
		ClientConfig: &agent.ClientConfig{Host: u},
		FetchRootKey: true,
	})
	if err != nil {
		return nil, err
	}
	return a, nil
}
