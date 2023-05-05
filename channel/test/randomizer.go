// SPDX-License-Identifier: Apache-2.0

package test

import (
	"math/big"
	"math/rand"

	pchannel "perun.network/go-perun/channel"
	pchtest "perun.network/go-perun/channel/test"
	"perun.network/perun-icp-backend/channel"
)

// randomizer implements the channel/test.Randomizer interface.
type randomizer struct{}

const (
	// MaxBalance that will be used for deposit testing.
	MaxBalance = uint64(1) << 30
	// MinBalance is the minimal amount that will to be deposited.
	MinBalance = uint64(1) << 20
)

// NewRandomAsset returns the only asset that is available.
func (randomizer) NewRandomAsset(rng *rand.Rand) pchannel.Asset {
	return channel.Asset
}

// WithBalancesRange specifies the Balances range.
func WithBalancesRange() pchtest.RandomOpt {
	return pchtest.WithBalancesInRange(
		new(big.Int).SetUint64(MinBalance),
		new(big.Int).SetUint64(MaxBalance))
}

// DefaultRandomOpts returns the default options for tests value random generation.
func DefaultRandomOpts() pchtest.RandomOpt {
	return WithBalancesRange().
		Append(pchtest.WithoutApp()).
		Append(pchtest.WithNumLocked(0)).
		Append(pchtest.WithAssets(channel.Asset)).
		Append(pchtest.WithNumParts(2))
}
