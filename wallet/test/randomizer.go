// SPDX-License-Identifier: Apache-2.0
package test

import (
	cr "crypto/rand"
	"math/rand"
	pwallet "perun.network/go-perun/wallet"
	ptest "perun.network/go-perun/wallet/test"

	"github.com/perun-network/perun-icp-backend/wallet"
)

// Randomizer implements the wallet/test.Randomizer interface.
type Randomizer struct {
	wallet *wallet.FsWallet
}

// NewRandomizer returns a new Randomizer.
func NewRandomizer() *Randomizer {
	return &Randomizer{NewWallet()}
}

func NewWallet() *wallet.FsWallet {
	w, err := wallet.NewRAMWallet(cr.Reader)
	if err != nil {
		panic("NewWallet: failed to create wallet: " + err.Error())
	}
	return w
}

func (r *Randomizer) NewWallet() ptest.Wallet {
	return NewWallet()
}

func (r *Randomizer) RandomWallet() ptest.Wallet {
	return r.wallet
}

// NewRandomAccount creates a new random account using the wallet package.
func (r *Randomizer) NewRandomAccount(rng *rand.Rand) pwallet.Account {
	return r.wallet.NewRandomAccount(rng)
}

// NewRandomAddress creates a new random address using the wallet package.
func (r *Randomizer) NewRandomAddress(rng *rand.Rand) pwallet.Address {
	account := r.NewRandomAccount(rng)
	return account.Address()
}
