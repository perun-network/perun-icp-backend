//

package wallet

import (
	"crypto"

	ed "github.com/oasisprotocol/curve25519-voi/primitives/ed25519"
	"perun.network/go-perun/wallet"
)

type Account ed.PrivateKey

var _ wallet.Account = (*Account)(nil)

func (a Account) Address() wallet.Address {
	addr := Address(ed.PrivateKey(a).Public().(ed.PublicKey))
	return &addr
}

func (a Account) SignData(data []byte) ([]byte, error) {
	return ed.PrivateKey(a).Sign(nil, data, crypto.Hash(0))
}

func (a Account) clear() {
	for i := range a[:] {
		a[i] = 0
	}
}
