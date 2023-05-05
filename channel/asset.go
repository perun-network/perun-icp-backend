// SPDX-License-Identifier: Apache-2.0

package channel

import (
	"perun.network/go-perun/channel"
	pchannel "perun.network/go-perun/channel"
)

// Implements the Perun Asset interface.
// Does not contain any fields since there is only one asset per chain.
type asset struct{}

// Asset is the unique asset that is supported by the chain.
var Asset = &asset{}

func (asset) Index() pchannel.Index {
	return 0
}

// MarshalBinary does nothing and returns nil since the backend has only one asset.
func (asset) MarshalBinary() (data []byte, err error) {
	return
}

// UnmarshalBinary does nothing and returns nil since the backend has only one asset.
func (*asset) UnmarshalBinary(data []byte) error {
	return nil
}

// Equal returns true if the assets are the same.
func (asset) Equal(b channel.Asset) bool {
	_, ok := b.(*asset)
	return ok
}
