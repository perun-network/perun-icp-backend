package main

import (
	"fmt"

	"perun.network/go-perun/wire"
	"perun.network/perun-icp-backend/client"
	"perun.network/perun-icp-backend/wallet"
)

const (
	Host            = "http://127.0.0.1"
	Port            = 8000
	perunPrincipal  = "r7inp-6aaaa-aaaaa-aaabq-cai"
	ledgerPrincipal = "rrkah-fqaaa-aaaaa-aaaaq-cai"
	userA           = "97520b79b03e38d3f6b38ce5026a813ccc9d1a3e830edb6df5970e6ca6ad84be"
	userB           = "40fd2dc85bc7d264b31f1fa24081d7733d303b49b7df84e3d372338f460aa678"
	userAbalance    = 100000
	userBbalance    = 200000
)

func main() {
	fmt.Println("Hello World")

	perunWlt := wallet.NewWallet()
	acc := perunWlt.NewAccount()

	fmt.Println("perunWlt: ", acc)

	clientAConfig, err := client.NewUserConfig(userAbalance, "usera", Host, Port)
	if err != nil {
		panic(err)
	}

	_, err = client.NewPerunUser(clientAConfig, ledgerPrincipal)
	if err != nil {
		panic(err)
	}

	_ = wire.NewLocalBus()

}
