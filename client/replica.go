package client

import (
	"perun.network/perun-icp-backend/setup"
)

func NewReplica() *setup.DfxSetup {

	demoConfig := setup.DfxConfig{
		Host:        "http://127.0.0.1",
		Port:        8000,
		ExecPath:    "./test/testdata/",
		AccountPath: "./test/testdata/identities/minter_identity.pem",
	}

	// perunID := "r7inp-6aaaa-aaaaa-aaabq-cai"
	// ledgerID := "rrkah-fqaaa-aaaaa-aaaaq-cai"

	dfx := setup.NewDfxSetup(demoConfig)
	// acc1, err := NewRandL2Account()
	// if err != nil {
	// 	t.Fatal("Error generating random account 1:", err)
	// }
	// acc2, err := NewRandL2Account()
	// if err != nil {
	// 	t.Fatal("Error generating random account 2:", err)
	// }
	// accs := []wallet.Account{acc1, acc2}
	// conn1 := chanconn.NewConnector(perunID, ledgerID, testConfig.AccountPath, testConfig.ExecPath, testConfig.Host, testConfig.Port)
	// conn2 := chanconn.NewConnector(perunID, ledgerID, testConfig.AccountPath, testConfig.ExecPath, testConfig.Host, testConfig.Port)
	// conns := []*chanconn.Connector{conn1, conn2}

	//return &Setup{t, pkgtest.Prng(t), accs, accs[0], accs[1], dfx, conns} //, chanConn}
	return dfx
}
