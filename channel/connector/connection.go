// SPDX-License-Identifier: Apache-2.0

package connector

import (
	"crypto/sha512"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/aviate-labs/agent-go"
	"github.com/aviate-labs/agent-go/identity"
	"github.com/aviate-labs/agent-go/principal"
	"net/url"
	"os"

	"math/big"

	"perun.network/go-perun/log"
	utils "perun.network/perun-icp-backend/utils"
)

type Connector struct {
	Log      log.Embedding
	Agent    *agent.Agent
	Source   *EventSource
	PerunID  *principal.Principal
	LedgerID *principal.Principal
	ExecPath ExecPath
}

func NewConnector(perunID, ledgerID, accountPath, execPath, host string, port int) *Connector {

	newAgent, err := NewAgent(accountPath, host, port)
	if err != nil {
		panic(err)
	}

	recipPerunID, err := utils.DecodePrincipal(perunID)
	if err != nil {
		panic(err)
	}

	recipLedgerID, err := utils.DecodePrincipal(ledgerID)
	if err != nil {
		panic(err)
	}

	chanConn := &Connector{
		Agent:    newAgent,
		Log:      log.MakeEmbedding(log.Default()),
		Source:   NewEventSource(),
		PerunID:  recipPerunID,
		LedgerID: recipLedgerID,
		ExecPath: ExecPath(execPath),
	}

	return chanConn
}

func NewAgent(accountPath, host string, port int) (*agent.Agent, error) {
	data, err := os.ReadFile(accountPath)
	if err != nil {
		return nil, err
	}
	var agentID identity.Identity
	agentID, err = identity.NewSecp256k1IdentityFromPEM(data)
	if err != nil {
		return nil, err
	}
	ic0, err := url.Parse(fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		return nil, err
	}

	agent := agent.New(agent.Config{
		Identity: agentID,
		ClientConfig: &agent.ClientConfig{
			Host: ic0,
		}})

	return &agent, nil
}

func NewExecPath(s string) ExecPath {
	return ExecPath(s)
}

func (f *Funding) Memo() (uint64, error) {
	// The memo is the unique channel ID, which is the first 8 bytes of the hash of the serialized funding candid
	serializedFunding, err := f.SerializeFundingCandid()
	if err != nil {
		return 0, fmt.Errorf("error in serializing funding: %w", err)
	}

	hasher := sha512.New()
	hasher.Write(serializedFunding)
	fullHash := hasher.Sum(nil)

	var arr [8]byte
	copy(arr[:], fullHash[:8])
	memo := binary.LittleEndian.Uint64(arr[:])

	return memo, nil
}

func MakeBalance(bal *big.Int) (Balance, error) {
	if bal.Sign() < 0 {
		return 0, errors.New("invalid balance: negative value")
	}

	maxBal := new(big.Int).SetUint64(MaxBalance)
	if bal.Cmp(maxBal) > 0 {
		return 0, errors.New("invalid balance: exceeds max balance")
	}

	return Balance(bal.Uint64()), nil
}
