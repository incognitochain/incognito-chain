package wallet

import (
	"github.com/incognitochain/incognito-chain/common/base58"
)

func addChecksumToBytes(data []byte) ([]byte, error) {
	checksum := base58.ChecksumFirst4Bytes(data)
	return append(data, checksum...), nil
}
