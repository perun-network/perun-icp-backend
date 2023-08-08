package client

import (
	//"context"
	"log"
	"math/big"
	"strconv"
	"time"
)

func FormatBalance(bal *big.Int) string {
	log.Printf("balance: %s", bal.String())
	balIC := bigIntToFloat64(bal)
	return strconv.FormatFloat(balIC, 'f', 6, 64) + " IC Token"
}

func bigIntToFloat64(bi *big.Int) float64 {
	bf := new(big.Float).SetInt(bi)
	f64, _ := bf.Float64()
	return f64
}

// ShannonToCKByte converts a given amount in Shannon to CKByte.
func ShannonToCKByte(shannonAmount *big.Int) (adaAmount *big.Float) {
	shannonPerCKByte := new(big.Int).Exp(big.NewInt(10), big.NewInt(9), nil)
	shannonPerCKByteFloat := new(big.Float).SetInt(shannonPerCKByte)
	shannonAmountFloat := new(big.Float).SetInt(shannonAmount)
	return new(big.Float).Quo(shannonAmountFloat, shannonPerCKByteFloat)
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
			p.NotifyAllBalance(bal) // TODO: Update demo tui to allow for big.Int balances
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
