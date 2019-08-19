package blsbft

import (
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/common/base58"
	"github.com/incognitochain/incognito-chain/consensus/blsmultisig"
)

type blsKeySet struct {
	Publickey  []byte
	PrivateKey []byte
}

func (keyset *blsKeySet) GetPublicKeyBase58() string {
	return base58.Base58Check{}.Encode(keyset.Publickey, common.ZeroByte)
}
func (keyset *blsKeySet) GetPrivateKeyBase58() string {
	return base58.Base58Check{}.Encode(keyset.PrivateKey, common.ZeroByte)
}

func (keyset *blsKeySet) SignData(data *common.Hash) (string, error) {
	return "", nil
}
func (keyset *blsKeySet) validateAggregatedSig(dataHash *common.Hash, aggSig string, validatorPubkeyList []string) error {
	return nil
}
func (keyset *blsKeySet) validateSingleSig(dataHash *common.Hash, sig string, pubkey string) error {
	return nil
}

func (e *BLSBFT) LoadUserKey(privateKeyStr string) error {
	privateKeyBytes, _, err := base58.Base58Check{}.Decode(privateKeyStr)
	if err != nil {
		return err
	}
	privateKey := blsmultisig.B2I(privateKeyBytes)
	publicKeyBytes := blsmultisig.PKBytes(blsmultisig.PKGen(privateKey))
	var keyset blsKeySet
	keyset.Publickey = publicKeyBytes
	keyset.PrivateKey = privateKeyBytes

	e.UserKeySet = &keyset
	return nil
}
func (e BLSBFT) GetUserPublicKey() string {
	if e.UserKeySet != nil {
		return e.UserKeySet.GetPublicKeyBase58()
	}
	return ""
}
func (e BLSBFT) GetUserPrivateKey() string {
	if e.UserKeySet != nil {
		return e.UserKeySet.GetPrivateKeyBase58()
	}
	return ""
}
