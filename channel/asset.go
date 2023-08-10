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
