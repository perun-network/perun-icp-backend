// Copyright 2023 - See NOTICE file for copyright holders.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"log"
	"os"
	vc "perun.network/perun-demo-tui/client"
	"perun.network/perun-demo-tui/view"
	"perun.network/perun-icp-backend/client"
	"perun.network/perun-icp-backend/wallet"
)

const (
	Host         = "http://127.0.0.1"
	Port         = 4943
	perunID      = "be2us-64aaa-aaaaa-qaabq-cai"
	ledgerID     = "bkyz2-fmaaa-aaaaa-qaaaq-cai"
	userAPemPath = "./userdata/identities/usera_identity.pem"
	userBPemPath = "./userdata/identities/userb_identity.pem"
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

	alice, err := client.SetupPaymentClient("alice", perunWltA, sharedComm, perunID, ledgerID, Host, Port, userAPemPath)
	if err != nil {
		panic(err)
	}

	bob, err := client.SetupPaymentClient("bob", perunWltB, sharedComm, perunID, ledgerID, Host, Port, userBPemPath)
	if err != nil {
		panic(err)
	}

	clients := []vc.DemoClient{alice, bob}
	_ = view.RunDemo("Perun Payment Channel on the Internet Computer ", clients)

}
