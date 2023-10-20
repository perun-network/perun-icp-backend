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

package client

import (
	"log"
	"math/big"
	"strconv"
	"time"
)

func FormatBalance(bal *big.Int) string {
	log.Printf("balance: %s", bal.String())
	balIC := bigIntToFloat64(bal)
	return strconv.FormatFloat(balIC, 'f', 6, 64) + " Token"
}

func bigIntToFloat64(bi *big.Int) float64 {
	bf := new(big.Float).SetInt(bi)
	f64, _ := bf.Float64()
	return f64
}

func (p *PaymentClient) PollBalances() {
	defer log.Println("PollBalances: stopped")
	pollingInterval := time.Second

	log.Println("PollBalances")
	updateBalance := func() {

		balance := p.GetOwnBalance()

		p.balanceMutex.Lock()
		if balance.Cmp(p.balance) != 0 {
			p.balance = balance
			bal := p.balance.Int64()
			p.balanceMutex.Unlock()
			p.NotifyAllBalance(bal)
		} else {
			p.balanceMutex.Unlock()
		}
	}
	// Poll the balance every 5 seconds.
	for {
		updateBalance()
		time.Sleep(pollingInterval)
	}
}
