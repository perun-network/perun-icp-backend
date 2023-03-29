package channel

import (
	"crypto/rand"
	"fmt"
	"github.com/aviate-labs/agent-go"
	"github.com/aviate-labs/agent-go/identity"
	"net/url"
	"os"
	"perun.network/perun-icp-backend/wallet"
)

type DfxConfig struct {
	Host        string
	Port        int
	ExecPath    string
	AccountPath string // use local path to a minter .pem file
}

type UserClient struct {
	Agent     *agent.Agent
	L2Account wallet.Account
}

func (u UserClient) NewL2Account() (wallet.Account, error) {
	wlt := wallet.NewRAMWallet(rand.Reader)
	acc := wlt.NewAccount()

	return acc, nil
}

func NewUserClient(config DfxConfig) (*UserClient, error) {
	agent, err := NewAgent(config)
	if err != nil {
		return nil, err
	}
	userClient := &UserClient{
		Agent: agent,
	}
	userClient.L2Account, err = userClient.NewL2Account()
	if err != nil {
		return nil, err
	}

	return userClient, nil
}

func NewAgent(config DfxConfig) (*agent.Agent, error) {
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

	agent := agent.New(agent.AgentConfig{
		Identity: &agentID,
		ClientConfig: &agent.ClientConfig{
			Host: ic0,
		}})
	return &agent, nil
}