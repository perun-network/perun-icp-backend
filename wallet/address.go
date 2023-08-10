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

package wallet

import (
	"bytes"
	"encoding/base64"
	"fmt"
	ed "github.com/oasisprotocol/curve25519-voi/primitives/ed25519"
	"perun.network/go-perun/wallet"
	_ "perun.network/go-perun/wire"
)

// Address is an ed25519 public key and represents a perun off-chain
// identity in internet computer channels.
type Address ed.PublicKey

var _ wallet.Address = (*Address)(nil)

var addrEncoding = base64.NewEncoding(
	"ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789_$",
).WithPadding(base64.NoPadding)

func (a *Address) MarshalBinary() ([]byte, error) {
	return (*a)[:], nil
}

func (a *Address) UnmarshalBinary(data []byte) error {
	if len(data) != ed.PublicKeySize {
		return fmt.Errorf("invalid PK length: %d/%d", len(data), ed.PublicKeySize)
	}

	*a = make(Address, ed.PublicKeySize)
	copy(*a, data)
	return nil
}

func (a *Address) String() string {
	return addrEncoding.EncodeToString((*a)[:])
}

func (a *Address) Equal(b wallet.Address) bool {
	b_, ok := b.(*Address)
	if !ok {
		return false
	}
	return bytes.Equal((*a)[:], (*b_)[:])
}

func (a *Address) Cmp(b wallet.Address) int {
	return bytes.Compare((*a)[:], (*b.(*Address))[:])
}

func AsAddr(acc wallet.Address) *Address {
	return acc.(*Address)
}
