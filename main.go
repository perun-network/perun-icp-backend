package main

import (
	"log"
	"os"
	vc "perun.network/perun-demo-tui/client"
	"perun.network/perun-demo-tui/view"
	"perun.network/perun-icp-backend/client"
	//utils "perun.network/perun-icp-backend/utils"
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

func SetLogFile(path string) {
	logFile, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	log.SetOutput(logFile)
}

func main() {
	SetLogFile("demo.log")
	perunWltA := wallet.NewWallet()
	perunWltB := wallet.NewWallet()

	sharedComm := client.InitSharedComm()

	alice, err := client.SetupPaymentClient("alice", perunWltA, sharedComm, perunPrincipal, ledgerPrincipal, Host, Port, userAPemPath)
	if err != nil {
		panic(err)
	}

	bob, err := client.SetupPaymentClient("bob", perunWltB, sharedComm, perunPrincipal, ledgerPrincipal, Host, Port, userBPemPath)
	if err != nil {
		panic(err)
	}

	// alice.OpenChannel(bob.WireAddress(), channelCollateral)
	// achan := alice.Channel
	// bob.AcceptedChannel()
	// bchan := bob.Channel

	// balanceA := alice.GetOwnBalance()
	// balanceB := bob.GetOwnBalance()

	// log.Println("alice balance: ", balanceA)
	// log.Println("bob balance: ", balanceB)

	// aliceBal := alice.GetChannelBalance()
	// bobBal := bob.GetChannelBalance()

	// log.Println("Perun Canister total balance: ", aliceBal, bobBal)

	// // sending payment/s

	// log.Println("Sending payments...")
	// achan.SendPayment(1000)
	// bchan.SendPayment(2000)

	// log.Println("Settling channel")
	// bchan.Settle()

	// achan.Settle()

	// perunBalAfterSettle := alice.GetChannelBalance()
	// perunBalBfterSettle := bob.GetChannelBalance()

	// log.Println("Perun Canister total balance after Settlement: ", perunBalAfterSettle, perunBalBfterSettle)

	// alice.Shutdown()
	// bob.Shutdown()

	// recipPerunID, err := utils.DecodePrincipal(perunPrincipal)
	// if err != nil {
	// 	panic(err)
	// }
	// perunBal := alice.GetExtBalance(*recipPerunID)
	// balanceAZ := alice.GetOwnBalance()
	// balanceBZ := bob.GetOwnBalance()

	// log.Println("alice balance after settlement: ", balanceAZ)
	// log.Println("bob balance after settlement: ", balanceBZ)

	// log.Println("Perun Canister total balance after Settle: ", perunBal)
	clients := []vc.DemoClient{alice, bob}
	_ = view.RunDemo("Internet Computer Payment Channel Demo", clients)

}
