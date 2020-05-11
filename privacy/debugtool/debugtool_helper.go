package debugtool

import "github.com/incognitochain/incognito-chain/common/base58"

func EncodeBase58Check(data []byte) string {
	b := base58.Base58Check{}.Encode(data, 0)
	return b
}

func DecodeBase58Check(s string) ([]byte, error) {
	b, _, err := base58.Base58Check{}.Decode(s)
	return b, err
}