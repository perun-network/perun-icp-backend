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
	"encoding/binary"
	"math/big"
)

func Uint64ToBytes(i uint64) []byte {
	var buf [8]byte
	binary.LittleEndian.PutUint64(buf[:], i)
	return buf[:]
}

func BoolToBytes(b bool) []byte {
	if b {
		return []byte{1}
	} else {
		return []byte{0}
	}
}

func BigToLittleEndianBytes(big *big.Int) []byte {
	bigEndianBytes := big.Bytes()
	for i, j := 0, len(bigEndianBytes)-1; i < j; i, j = i+1, j-1 {
		bigEndianBytes[i], bigEndianBytes[j] = bigEndianBytes[j], bigEndianBytes[i]
	}
	return bigEndianBytes
}
