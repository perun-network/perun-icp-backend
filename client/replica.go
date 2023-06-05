package client

import (
	"perun.network/perun-icp-backend/setup"
)

const transferFee = 10000

func NewReplica() *setup.DfxSetup {

	demoConfig := setup.DfxConfig{
		Host:        "http://127.0.0.1",
		Port:        4943,
		ExecPath:    "./test/testdata/",
		AccountPath: "./test/testdata/identities/minter_identity.pem",
	}

	dfx := setup.NewDfxSetup(demoConfig)

	return dfx
}
