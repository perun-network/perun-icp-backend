// SPDX-License-Identifier: Apache-2.0

package connector

import (
	"crypto/sha512"
	"encoding/binary"
	"errors"
	"fmt"

	"github.com/aviate-labs/agent-go"
	"github.com/aviate-labs/agent-go/ic/icpledger"

	"math/big"
	"net/url"
	"os"
	"sync"

	"github.com/aviate-labs/agent-go/identity"
	"github.com/aviate-labs/agent-go/principal"

	"perun.network/go-perun/log"
	"perun.network/perun-icp-backend/channel/connector/icperun"
	utils "perun.network/perun-icp-backend/utils"
)

type Connector struct {
	Log         log.Embedding
	DfxAgent    *agent.Agent
	Mutex       *sync.Mutex
	PerunID     *principal.Principal
	LedgerID    *principal.Principal
	L1Account   *principal.Principal
	LedgerAgent *icpledger.Agent
	PerunAgent  *icperun.Agent
}

// func NewDfxConnector(pemAccountPath string, ledgerAddr, perunAddr string, host string, port int) (*DfxConnector, error) {

// 	pemAccountFullPath := filepath.Join(utils.SetHomeDir(), ".config", "dfx", "identity", pemAccountPath, "identity.pem")

// 	ledgerPrincipal, err := principal.Decode(ledgerAddr)
// 	if err != nil {
// 		return nil, err
// 	}

// 	perunPrincipal, err := principal.Decode(perunAddr)
// 	if err != nil {
// 		return nil, err
// 	}

// 	dfxAgent, err := NewDfxAgent(pemAccountFullPath, host, port)
// 	if err != nil {
// 		return nil, err
// 	}

// 	dfxConnector := &DfxConnector{
// 		dfxAgent: dfxAgent,
// 	}

// 	dfxConnector.ledgerAgent, err = MakeLedgerAgent(pemAccountFullPath, host, port, ledgerPrincipal)
// 	if err != nil {
// 		return nil, err
// 	}

// 	dfxConnector.perunAgent, err = MakePerunAgent(pemAccountFullPath, host, port, perunPrincipal)
// 	if err != nil {
// 		return nil, err
// 	}

// 	// perunUser.l2Account, err = perunUser.NewL2Account()
// 	// if err != nil {
// 	// 	return nil, err
// 	// }

// 	return dfxConnector, nil
// }

// func NewDfxConnector(pemAccountPath string, ledgerAddr, perunAddr string, host string, port int) (*DfxConnector, error) {
func NewDfxConnector(perunID, ledgerID, pemAccountPath, host string, port int) *Connector {

	dfxAgent, err := NewDfxAgent(pemAccountPath, host, port)
	if err != nil {
		panic(err)
	}

	dfxAccount := dfxAgent.Sender()

	recipPerunID, err := utils.DecodePrincipal(perunID)
	if err != nil {
		panic(err)
	}

	recipLedgerID, err := utils.DecodePrincipal(ledgerID)
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
		DfxAgent:    dfxAgent,
		Log:         log.MakeEmbedding(log.Default()),
		PerunID:     recipPerunID,
		LedgerID:    recipLedgerID,
		L1Account:   &dfxAccount,
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

func NewDfxAgent(accountPath string, host string, port int) (*agent.Agent, error) {
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

	agent, err := agent.New(agent.Config{
		Identity: agentID,
		ClientConfig: &agent.ClientConfig{
			Host: ic0,
		},
		FetchRootKey: true})

	if err != nil {
		return nil, err
	}

	return agent, nil
}

// func NewAgent(accountPath, host string, port int) (*agent.Agent, error) {
// 	data, err := os.ReadFile(accountPath)
// 	if err != nil {
// 		return nil, err
// 	}
// 	var agentID identity.Identity
// 	agentID, err = identity.NewSecp256k1IdentityFromPEM(data)
// 	if err != nil {
// 		return nil, err
// 	}
// 	ic0, err := url.Parse(fmt.Sprintf("%s:%d", host, port))
// 	if err != nil {
// 		return nil, err
// 	}

// 	fmt.Println("ic0: ", ic0)

// 	agent, err := agent.New(agent.Config{
// 		Identity: agentID,
// 		ClientConfig: &agent.ClientConfig{
// 			Host: ic0,
// 		},
// 		FetchRootKey: true,
// 	})
// 	if err != nil {
// 		return nil, err
// 	}

// 	return agent, nil
// }

func NewPerunAgent(canID principal.Principal, accountPath, host string, port int) (*icperun.Agent, error) {
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

	agent, err := icperun.NewAgent(canID, agent.Config{
		Identity: agentID,
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

	agent, err := icpledger.NewAgent(canID, agent.Config{
		Identity: agentID,
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

// func NewExecPath(s string) ExecPath {
// 	return ExecPath(s)
// }

func (f *Funding) Memo() (uint64, error) {

	hasher := sha512.New() // Assuming Hash::digest uses SHA-512.

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
