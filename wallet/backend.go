package wallet

import (
	"fmt"
	"io"

	ed "github.com/oasisprotocol/curve25519-voi/primitives/ed25519"

	"perun.network/go-perun/wallet"
)

type Backend struct{}

var _ wallet.Backend = Backend{}

func init() {
	wallet.SetBackend(Backend{})
}

func (Backend) NewAddress() wallet.Address {
	a := make(Address, 0)
	return &a
}

func (Backend) DecodeSig(r io.Reader) (wallet.Sig, error) {
	sig := make([]byte, ed.SignatureSize)
	if _, err := io.ReadFull(r, sig); err != nil {
		return nil, err
	}
	return wallet.Sig(sig), nil
}

func (Backend) VerifySignature(
	msg []byte,
	sign wallet.Sig,
	a wallet.Address,
) (ok bool, err error) {
	defer func() {
		if e := recover(); e != nil {
			var ok bool
			if err, ok = e.(error); !ok {
				err = fmt.Errorf("%v", e)
			}
		}
	}()
	ok = ed.Verify(ed.PublicKey(*a.(*Address)), msg, sign[:])
	return
}
