// Copyright 2023 - See NOTICE file for copyright holders.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package connector

import (
	"crypto/sha512"
	"encoding/binary"
	"fmt"

	"github.com/aviate-labs/agent-go"
	"github.com/aviate-labs/agent-go/ic/icpledger"

	"github.com/aviate-labs/agent-go/identity"
	"github.com/aviate-labs/agent-go/principal"
	"net/url"
	"os"
	"perun.network/go-perun/log"
	"perun.network/perun-icp-backend/channel/connector/icperun"
	"sync"
)

// Connects Perun users with the Internet Computer
type Connector struct {
	Log         log.Embedding
	ICAgent     *agent.Agent
	Mutex       *sync.Mutex
	PerunID     *principal.Principal
	LedgerID    *principal.Principal
	L1Account   *principal.Principal
	LedgerAgent *icpledger.Agent
	PerunAgent  *icperun.Agent
}

func NewICConnector(perunID, ledgerID, pemAccountPath, host string, port int) *Connector {

	ICAgent, err := NewICAgent(pemAccountPath, host, port)
	if err != nil {
		panic(err)
	}

	ICAccount := ICAgent.Sender()

	recipPerunID, err := DecodePrincipal(perunID)
	if err != nil {
		panic(err)
	}

	recipLedgerID, err := DecodePrincipal(ledgerID)
	if err != nil {
		panic(err)
	}

	LedgerAgent, err := NewLedgerAgent(*recipLedgerID, pemAccountPath, host, port)
	if err != nil {
		panic(err)
	}

	PerunAgent, err := NewPerunAgent(*recipPerunID, pemAccountPath, host, port)
	if err != nil {
		panic(err)
	}
	chanConn := &Connector{
		ICAgent:     ICAgent,
		Log:         log.MakeEmbedding(log.Default()),
		PerunID:     recipPerunID,
		LedgerID:    recipLedgerID,
		L1Account:   &ICAccount,
		LedgerAgent: LedgerAgent,
		PerunAgent:  PerunAgent,
	}

	return chanConn
}

func NewIdentity(accountPath string) (*identity.Identity, error) {
	data, err := os.ReadFile(accountPath)
	if err != nil {
		return nil, err
	}
	var agentID identity.Identity
	agentID, err = identity.NewSecp256k1IdentityFromPEM(data)
	if err != nil {
		return nil, err
	}

	return &agentID, nil
}

func NewICAgent(accountPath string, host string, port int) (*agent.Agent, error) {
	agentID, err := NewIdentity(accountPath)
	if err != nil {
		return nil, err
	}
	ic0, err := url.Parse(fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		return nil, err
	}

	agent, err := agent.New(agent.Config{
		Identity: *agentID,
		ClientConfig: &agent.ClientConfig{
			Host: ic0,
		},
		FetchRootKey: true})

	if err != nil {
		return nil, err
	}

	return agent, nil
}

func NewPerunAgent(canID principal.Principal, accountPath, host string, port int) (*icperun.Agent, error) {
	agentID, err := NewIdentity(accountPath)
	if err != nil {
		return nil, err
	}
	ic0, err := url.Parse(fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		return nil, err
	}

	agent, err := icperun.NewAgent(canID, agent.Config{
		Identity: *agentID,
		ClientConfig: &agent.ClientConfig{
			Host: ic0,
		},
		FetchRootKey: true,
	})
	if err != nil {
		return nil, err
	}

	return agent, nil
}

func NewLedgerAgent(canID principal.Principal, accountPath, host string, port int) (*icpledger.Agent, error) {
	agentID, err := NewIdentity(accountPath)
	if err != nil {
		return nil, err
	}
	ic0, err := url.Parse(fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		return nil, err
	}

	agent, err := icpledger.NewAgent(canID, agent.Config{
		Identity: *agentID,
		ClientConfig: &agent.ClientConfig{
			Host: ic0,
		},
		FetchRootKey: true,
	})
	if err != nil {
		return nil, err
	}

	return agent, nil
}

func (f *Funding) Memo() (Memo, error) {

	hasher := sha512.New()

	chanBytes := f.Channel[:]
	addrBytes := f.Part[:]

	channelAddr := append(chanBytes, addrBytes...)

	hasher.Write(channelAddr)

	fullHash := hasher.Sum(nil)

	var arr [8]byte
	copy(arr[:], fullHash[:8])
	memo := binary.LittleEndian.Uint64(arr[:])

	return memo, nil
}
