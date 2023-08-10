// SPDX-License-Identifier: Apache-2.0
package test

import (
	wtest "github.com/perun-network/perun-icp-backend/wallet/test"
	pchtest "perun.network/go-perun/channel/test"
	pwtest "perun.network/go-perun/wallet/test"
)

func init() {
	pchtest.SetRandomizer(new(randomizer))
	walletRdz := wtest.NewRandomizer()
	pwtest.SetRandomizer(walletRdz)

}
