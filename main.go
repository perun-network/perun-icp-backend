package main

import (
	"fmt"
	"perun.network/go-perun/wire"
	"perun.network/perun-icp-backend/client"
	"perun.network/perun-icp-backend/wallet"

	"sync"
)

const (
	Host            = "http://127.0.0.1"
	Port            = 4943
	ledgerHost      = "http://127.0.0.1:4943"
	perunPrincipal  = "r7inp-6aaaa-aaaaa-aaabq-cai"
	ledgerPrincipal = "rrkah-fqaaa-aaaaa-aaaaq-cai"
	userAId         = "97520b79b03e38d3f6b38ce5026a813ccc9d1a3e830edb6df5970e6ca6ad84be"
	userBId         = "40fd2dc85bc7d264b31f1fa24081d7733d303b49b7df84e3d372338f460aa678"
	userAbalance    = 100000
	userBbalance    = 200000
)

func main() {

	replica := client.NewReplica()

	err := replica.StartDeployDfx()
	if err != nil {
		panic(err)
	}

	perunWltA := wallet.NewWallet()
	perunWltB := wallet.NewWallet()

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

	mtx := &sync.Mutex{}
	fmt.Printf("%v\n", mtx)

	// perun := chanconn.NewConnector(perunPrincipal, ledgerPrincipal, "./test/testdata/identities/usera_identity.pem", "./", Host, Port)
	// perun.Mutex = mtx
	alice, err := client.SetupPaymentClient(bus, perunWltA, mtx, perunPrincipal, ledgerPrincipal, Host, Port, "./test/testdata/identities/usera_identity.pem", "./")
	if err != nil {
		panic(err)
	}

	bob, err := client.SetupPaymentClient(bus, perunWltB, mtx, perunPrincipal, ledgerPrincipal, Host, Port, "./test/testdata/identities/userb_identity.pem", "./")
	if err != nil {
		panic(err)
	}
	achan := alice.OpenChannel(bob.WireAddress(), 10)
	fmt.Println(userA, bob, alice, userB)

	fmt.Println("alicechan: ", achan.GetChannelParams().ID(), "State: ", achan.GetChannelState())
	err = replica.StopDFX()
	if err != nil {
		panic(err)
	}
}
