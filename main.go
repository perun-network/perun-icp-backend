package main

import (
	"fmt"

	"perun.network/go-perun/wire"
	"perun.network/perun-icp-backend/client"
	"perun.network/perun-icp-backend/wallet"
)

const (
	Host            = "http://127.0.0.1"
	Port            = 4943
	ledgerHost      = "http://127.0.0.1:4943"
	perunPrincipal  = "r7inp-6aaaa-aaaaa-aaabq-cai"
	ledgerPrincipal = "rrkah-fqaaa-aaaaa-aaaaq-cai"
	userA           = "97520b79b03e38d3f6b38ce5026a813ccc9d1a3e830edb6df5970e6ca6ad84be"
	userB           = "40fd2dc85bc7d264b31f1fa24081d7733d303b49b7df84e3d372338f460aa678"
	userAbalance    = 100000
	userBbalance    = 200000
)

func main() {

	replica := client.NewReplica()

	err := replica.StartDeployDfx()
	if err != nil {
		panic(err)
	}

	perunWlt := wallet.NewWallet()
	_ = perunWlt.NewAccount()

	clientAConfig, err := client.NewUserConfig(userAbalance, "usera", Host, Port)
	if err != nil {
		panic(err)
	}
	clientBConfig, err := client.NewUserConfig(userAbalance, "userb", Host, Port)
	if err != nil {
		panic(err)
	}
	userA, err := client.NewPerunUser(clientAConfig, ledgerPrincipal)
	if err != nil {
		panic(err)
	}

	userB, err := client.NewPerunUser(clientBConfig, ledgerPrincipal)
	if err != nil {
		panic(err)
	}

	bus := wire.NewLocalBus()

	alice, err := client.SetupPaymentClient(bus, perunWlt, perunPrincipal, ledgerPrincipal, Host, Port, "./test/testdata/identities/usera_identity.pem", "./")
	if err != nil {
		panic(err)
	}

	bob, err := client.SetupPaymentClient(bus, perunWlt, perunPrincipal, ledgerPrincipal, Host, Port, "./test/testdata/identities/userb_identity.pem", "./")
	if err != nil {
		panic(err)
	}

	alice.OpenChannel(bob.WireAddress(), 10)
	fmt.Println(userA, bob, alice, userB)
	err = replica.StopDFX()
	if err != nil {
		panic(err)
	}
}
