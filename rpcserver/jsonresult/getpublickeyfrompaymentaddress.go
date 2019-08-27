package jsonresult

import (
	"encoding/hex"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/common/base58"
)

type GetPublicKeyFromPaymentAddressResult struct {
	PublicKeyInBase58Check string
	PublicKeyInBytes       []int
	PublicKeyInHex         string
}

func NewGetPublicKeyFromPaymentAddressResult(publicKeyInBytes []byte) *GetPublicKeyFromPaymentAddressResult {
	obj := &GetPublicKeyFromPaymentAddressResult{}
	obj.PublicKeyInBase58Check = base58.Base58Check{}.Encode(publicKeyInBytes, common.ZeroByte)
	obj.PublicKeyInHex = hex.EncodeToString(publicKeyInBytes)
	obj.PublicKeyInBytes = make([]int, 0)
	for _, v := range publicKeyInBytes {
		obj.PublicKeyInBytes = append(obj.PublicKeyInBytes, int(v))
	}
	return obj
}
