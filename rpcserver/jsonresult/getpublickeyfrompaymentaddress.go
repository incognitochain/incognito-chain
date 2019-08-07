package jsonresult

import (
	"encoding/hex"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/common/base58"
)

type GetPublicKeyFromPaymentAddress struct {
	PublicKeyInBase58Check string
	PublicKeyInBytes       []int
	PublicKeyInHex         string
}

func (obj *GetPublicKeyFromPaymentAddress) Init(publicKeyInBytes []byte) {
	obj.PublicKeyInBase58Check = base58.Base58Check{}.Encode(publicKeyInBytes, common.ZeroByte)
	obj.PublicKeyInHex = hex.EncodeToString(publicKeyInBytes)
	obj.PublicKeyInBytes = make([]int, 0)
	for _, v := range publicKeyInBytes {
		obj.PublicKeyInBytes = append(obj.PublicKeyInBytes, int(v))
	}
}
