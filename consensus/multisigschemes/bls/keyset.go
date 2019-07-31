package bls

import (
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/common/base58"
)

type KeySet struct {
	Publickey  []byte
	PrivateKey []byte
}

func (keyset *KeySet) GetPubkeyB58() string {
	return base58.Base58Check{}.Encode(keyset.Publickey, common.ZeroByte)
}
