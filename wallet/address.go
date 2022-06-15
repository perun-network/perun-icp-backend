//

package wallet

import (
	"bytes"
	"encoding/base64"
	"fmt"

	ed "github.com/oasisprotocol/curve25519-voi/primitives/ed25519"
	"perun.network/go-perun/wallet"
)

type Address ed.PublicKey

var _ wallet.Address = (*Address)(nil)

var addrEncoding = base64.NewEncoding(
	"ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789_$",
).WithPadding(base64.NoPadding)

func (a Address) MarshalBinary() ([]byte, error) {
	return a[:], nil
}

func (a *Address) UnmarshalBinary(data []byte) error {
	if len(data) != ed.PublicKeySize {
		return fmt.Errorf("invalid PK length: %d/%d", len(data), ed.PublicKeySize)
	}

	*a = make(Address, ed.PublicKeySize)
	copy(*a, data)
	return nil
}

func (a Address) String() string {
	return addrEncoding.EncodeToString(a[:])
}

func (a Address) Equal(b wallet.Address) bool {
	return bytes.Equal(a[:], (*b.(*Address))[:])
}

func (a Address) Cmp(b wallet.Address) int {
	return bytes.Compare(a[:], (*b.(*Address))[:])
}
