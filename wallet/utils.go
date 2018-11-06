package wallet

import (
	"encoding/binary"
	"github.com/ninjadotorg/constant/common/base58"
)

func addChecksumToBytes(data []byte) ([]byte, error) {
	checksum := base58.ChecksumFirst4Bytes(data)
	return append(data, checksum...), nil
}

/*
Numerical
*/
func uint32Bytes(i uint32) []byte {
	bytes := make([]byte, 4)
	binary.BigEndian.PutUint32(bytes, i)
	return bytes
}
