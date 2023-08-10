// SPDX-License-Identifier: Apache-2.0

package wire

import (
	"math/rand"
	"perun.network/perun-icp-backend/wallet"

	"perun.network/perun-icp-backend/wallet/test"

	"perun.network/go-perun/wire"
)

// Address is a wrapper for wallet.Address.
type Address struct {
	*wallet.Address
}

// NewAddress returns a new address.
func NewAddress() *Address {
	return &Address{}
}

// Equal returns whether the two addresses are equal.
func (a Address) Equal(b wire.Address) bool {
	bTyped, ok := b.(*Address)
	if !ok {
		panic("wrong type")
	}
	return a.Address.Equal(bTyped.Address)
}

// Cmp compares the byte representation of two addresses. For `a.Cmp(b)`
// returns -1 if a < b, 0 if a == b, 1 if a > b.
func (a Address) Cmp(b wire.Address) int {
	bTyped, ok := b.(*Address)
	if !ok {
		panic("wrong type")
	}
	return a.Address.Cmp(bTyped.Address)
}

// NewRandomAddress returns a new random peer address.
func NewRandomAddress(rng *rand.Rand) *Address {
	r := test.NewRandomizer()
	addr := r.NewRandomAddress(rng)
	walletAddr, ok := addr.(*wallet.Address)
	if !ok {
		return nil
	}
	return &Address{walletAddr}
}
