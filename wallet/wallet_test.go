// SPDX-License-Identifier: Apache-2.0

package wallet

import (
	"fmt"
	"io"
	"testing"

	ed "github.com/oasisprotocol/curve25519-voi/primitives/ed25519"
	"github.com/stretchr/testify/require"

	"perun.network/go-perun/wallet/test"
	ptest "polycry.pt/poly-go/test"
)

func setup(rng io.Reader) (*test.Setup, error) {
	w, err := NewRAMWallet(rng)
	if err != nil {
		return nil, fmt.Errorf("error creating wallet: %v", err)
	}

	newWallet, err := NewRAMWallet(rng)
	if err != nil {
		return nil, fmt.Errorf("error creating newWallet: %v", err)
	}
	marshalledAddr, err := newWallet.NewAccount().Address().MarshalBinary()
	if err != nil {
		return nil, fmt.Errorf("error marshalling address: %v", err)
	}

	zero := make(Address, ed.PublicKeySize)
	return &test.Setup{
		Backend:           Backend{},
		Wallet:            w,
		AddressInWallet:   w.NewAccount().Address(),
		ZeroAddress:       &zero,
		AddressMarshalled: marshalledAddr,
	}, nil
}

func TestAddress(t *testing.T) {
	setupResult, err := setup(ptest.Prng(t))
	if err != nil {
		t.Fatalf("Error in setup: %v", err)
	}
	test.TestAddress(t, setupResult)
}

func TestGenericSignatureSize(t *testing.T) {
	setupResult, err := setup(ptest.Prng(t))
	if err != nil {
		t.Fatalf("Error in setup: %v", err)
	}
	test.GenericSignatureSizeTest(t, setupResult)
}

func TestAccountWithWalletAndBackend(t *testing.T) {
	setupResult, err := setup(ptest.Prng(t))
	if err != nil {
		t.Fatalf("Error in setup: %v", err)
	}
	test.TestAccountWithWalletAndBackend(t, setupResult)
}

func TestFsWallet(t *testing.T) {
	path := "/tmp/.perun_icp_packend_test_wallet"

	w, err := CreateOrLoadFsWallet(path, ptest.Prng(t))
	require.NoError(t, err, "creating wallet")

	acc := w.NewAccount()

	load, err := CreateOrLoadFsWallet(path, nil)
	require.NoError(t, err, "loading wallet")

	_, err = load.Unlock(acc.Address())
	require.Error(t, err, "expected unlocking to fail")

	w.IncrementUsage(acc.Address())
	load, err = CreateOrLoadFsWallet(path, nil)
	require.NoError(t, err, "loading wallet")

	acc2, err := load.Unlock(acc.Address())
	require.NoError(t, err, "unlocking account")
	require.Equal(t, acc, acc2, "loaded account must be the generated account")
}
