package main

import (
	"fmt"

	//"github.com/aviate-labs/agent-go/principal"
	"log"
	utils "perun.network/perun-icp-backend/utils"

	"perun.network/perun-icp-backend/client"
	"perun.network/perun-icp-backend/wallet"
)

const (
	Host              = "http://127.0.0.1"
	Port              = 4943
	perunPrincipal    = "be2us-64aaa-aaaaa-qaabq-cai"
	ledgerPrincipal   = "bkyz2-fmaaa-aaaaa-qaaaq-cai"
	userAId           = "97520b79b03e38d3f6b38ce5026a813ccc9d1a3e830edb6df5970e6ca6ad84be"
	userBId           = "40fd2dc85bc7d264b31f1fa24081d7733d303b49b7df84e3d372338f460aa678"
	userAPemPath      = "./userdata/identities/usera_identity.pem"
	userBPemPath      = "./userdata/identities/userb_identity.pem"
	channelCollateral = 50000
)

func main() {

	perunWltA := wallet.NewWallet()
	perunWltB := wallet.NewWallet()

	sharedComm := client.InitSharedComm()

	alice, err := client.SetupPaymentClient(perunWltA, sharedComm, perunPrincipal, ledgerPrincipal, Host, Port, userAPemPath)
	if err != nil {
		panic(err)
	}

	bob, err := client.SetupPaymentClient(perunWltB, sharedComm, perunPrincipal, ledgerPrincipal, Host, Port, userBPemPath)
	if err != nil {
		panic(err)
	}
	alice.OpenChannel(bob.WireAddress(), channelCollateral)
	achan := alice.Channel
	bob.AcceptedChannel()
	bchan := bob.Channel

	balanceA := alice.GetOwnBalance()
	balanceB := bob.GetOwnBalance()

	fmt.Println("balance alice: ", balanceA)
	fmt.Println("balance bob: ", balanceB)

	fmt.Println("Seechan: ", achan, &achan)

	aliceBal := alice.GetChannelBalance()
	bobBal := bob.GetChannelBalance()

	fmt.Println("Perun Canister total balance: ", aliceBal, bobBal)

	// sending payment/s

	log.Println("Sending payments...")
	achan.SendPayment(1000)
	bchan.SendPayment(2000)

	fmt.Println("achan: ", achan.GetChannel().State(), "bchan: ", bchan.GetChannel().State())

	log.Println("Settling channel")
	bchan.Settle()
	fmt.Println("still blocking??")

	perunBalAfter1 := alice.GetChannelBalance()
	perunBalBfter1 := bob.GetChannelBalance()

	fmt.Println("Perun Canister total balance after 1st Settle: ", perunBalAfter1, perunBalBfter1)

	achan.Settle()
	fmt.Println("Did i settle? bob")
	perunBalAfter2 := alice.GetChannelBalance()
	perunBalBfter2 := bob.GetChannelBalance()

	fmt.Println("Perun Canister total balance after 2nd Settle: ", perunBalAfter2, perunBalBfter2)

	alice.Shutdown()
	bob.Shutdown()

	recipPerunID, err := utils.DecodePrincipal(perunPrincipal)
	if err != nil {
		panic(err)
	}
	perunBal := alice.GetExtBalance(*recipPerunID)
	balanceAZ := alice.GetOwnBalance()
	balanceBZ := bob.GetOwnBalance()

	fmt.Println("balance alice: ", balanceAZ)
	fmt.Println("balance bob: ", balanceBZ)

	fmt.Println("Perun Canister total balance after Settle: ", perunBal)
}
