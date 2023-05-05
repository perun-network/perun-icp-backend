// SPDX-License-Identifier: Apache-2.0

package channel

import (
	"crypto/rand"
	"fmt"
	"net/url"
	"os"

	"github.com/aviate-labs/agent-go"
	"github.com/aviate-labs/agent-go/identity"
	"github.com/aviate-labs/agent-go/ledger"
	"github.com/aviate-labs/agent-go/principal"
	"perun.network/perun-icp-backend/setup"
	"perun.network/perun-icp-backend/wallet"
)

// type UserClient struct {
// 	Agent     *agent.Agent
// 	L2Account wallet.Account
// 	Ledger    *ledger.Agent
// }

type PerunUser struct {
	L2Account wallet.Account
	Agent     *agent.Agent
	Ledger    *ledger.Agent
}

// func (u UserClient) NewLedger() (ledger.Agent, error) {
// 	ledgerAgent, err := ledger.NewAgent(u.Agent, u.L2Account)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return ledgerAgent, nil
// }

func (u *PerunUser) NewL2Account() (wallet.Account, error) {
	wlt, err := wallet.NewRAMWallet(rand.Reader)
	if err != nil {
		return nil, err
	}
	acc := wlt.NewAccount()

	return acc, nil
}

func MakeLedger(config setup.DfxConfig, canisterId principal.Principal) (*ledger.Agent, error) {
	data, err := os.ReadFile(config.AccountPath)
	if err != nil {
		return nil, err
	}

	var agentID identity.Identity
	agentID, err = identity.NewSecp256k1IdentityFromPEM(data)
	if err != nil {
		return nil, err
	}

	host, err := url.Parse(config.Host)
	if err != nil {
		return nil, fmt.Errorf("error parsing host URL: %v", err)
	}

	a := ledger.NewWithIdentity(canisterId, host, agentID)

	return &a, nil
}

func NewPerunUser(config setup.DfxConfig, prLedger principal.Principal) (*PerunUser, error) {
	agent, err := NewAgent(config)
	if err != nil {
		return nil, err
	}
	perunUser := &PerunUser{
		Agent: agent,
	}
	//agentLedger, err := MakeLedger(config, prLedger)

	perunUser.Ledger, err = MakeLedger(config, prLedger)
	if err != nil {
		return nil, err
	}

	perunUser.L2Account, err = perunUser.NewL2Account()
	if err != nil {
		return nil, err
	}

	return perunUser, nil
}

func NewAgent(config setup.DfxConfig) (*agent.Agent, error) {
	data, err := os.ReadFile(config.AccountPath)
	if err != nil {
		return nil, err
	}
	var agentID identity.Identity
	agentID, err = identity.NewSecp256k1IdentityFromPEM(data)
	if err != nil {
		return nil, err
	}
	ic0, err := url.Parse(fmt.Sprintf("%s:%d", config.Host, config.Port))
	if err != nil {
		return nil, err
	}

	// agent := agent.New(agent.AgentConfig{
	// 	Identity: &agentID,
	// 	ClientConfig: &agent.ClientConfig{
	// 		Host: ic0,
	// 	}})

	agent := agent.New(agent.Config{
		Identity: agentID,
		ClientConfig: &agent.ClientConfig{
			Host: ic0,
		}})

	return &agent, nil
}
