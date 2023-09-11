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
	"crypto"

	ed "github.com/oasisprotocol/curve25519-voi/primitives/ed25519"
	"perun.network/go-perun/wallet"
)

// Account is an ed25519 signing key. It signs messages for a perun off-chain
// identity in internet computer channels.
type Account ed.PrivateKey

var _ wallet.Account = (*Account)(nil)

func (a Account) Address() wallet.Address {
	addr := Address(ed.PrivateKey(a).Public().(ed.PublicKey))
	return &addr
}

func (a Account) L2Address() Address {
	addr := Address(ed.PrivateKey(a).Public().(ed.PublicKey))
	return addr
}

func (a Account) SignData(data []byte) ([]byte, error) {
	return ed.PrivateKey(a).Sign(nil, data, crypto.Hash(0))
}

func (a Account) clear() {
	for i := range a[:] {
		a[i] = 0
	}
}
