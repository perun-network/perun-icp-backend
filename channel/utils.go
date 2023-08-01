package channel

import (
	"crypto/sha512"
	"encoding/binary"
	"math/big"
)

func HashTo256(inp []byte) [32]byte {
	newhash := sha512.New()
	newhash.Write([]byte(inp))
	hashSum := newhash.Sum(nil)

	var hash [32]byte
	copy(hash[:], hashSum)

	return hash

}

func SerializeState(cid [32]byte, version uint64, alloc []*big.Int, finalized bool) []byte {
	var stateBytes []byte

	stateBytes = append(stateBytes, cid[:]...)
	stateBytes = append(stateBytes, Uint64ToBytes(version)...)

	for _, a := range alloc {
		stateBytes = append(stateBytes, a.Bytes()...)
	}

	stateBytes = append(stateBytes, BoolToBytes(finalized)...)

	return stateBytes
}

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
