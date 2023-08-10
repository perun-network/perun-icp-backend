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

package test_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	test "perun.network/perun-icp-backend/wallet/test"
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
