package wallet

import (
	"io"
	"testing"

	ed "github.com/oasisprotocol/curve25519-voi/primitives/ed25519"
	"perun.network/go-perun/wallet/test"
	ptest "polycry.pt/poly-go/test"
)

func setup(rng io.Reader) *test.Setup {
	w := NewRAMWallet(rng)
	marshalledAddr, err := NewRAMWallet(rng).NewAccount().Address().MarshalBinary()
	if err != nil {
		panic(err)
	}
	zero := make(Address, ed.PublicKeySize)
	return &test.Setup{
		Backend:           Backend{},
		Wallet:            w,
		AddressInWallet:   w.NewAccount().Address(),
		ZeroAddress:       &zero,
		AddressMarshalled: marshalledAddr,
	}
}

func TestAddress(t *testing.T) {
	test.TestAddress(t, setup(ptest.Prng(t)))
}
func TestGenericSignatureSize(t *testing.T) {
	test.GenericSignatureSizeTest(t, setup(ptest.Prng(t)))
}
func TestAccountWithWalletAndBackend(t *testing.T) {
	test.TestAccountWithWalletAndBackend(t, setup(ptest.Prng(t)))
}
