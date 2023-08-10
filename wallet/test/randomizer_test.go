// SPDX-License-Identifier: Apache-2.0
package test_test

import (
	"testing"

	test "github.com/perun-network/perun-icp-backend/wallet/test"
	"github.com/stretchr/testify/require"
	pkgtest "polycry.pt/poly-go/test"
)

func TestRandomizer_RandomAddress(t *testing.T) {
	rng := pkgtest.Prng(t)
	r := test.NewRandomizer()
	addr := r.NewRandomAddress(rng)

	for i := 0; i < 1000; i++ {
		addr2 := r.NewRandomAddress(rng)
		require.False(t, addr.Equal(addr2))
	}
}
