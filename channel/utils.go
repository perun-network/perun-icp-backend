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
