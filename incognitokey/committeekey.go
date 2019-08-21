package incognitokey

import (
	"encoding/json"
	"errors"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/common/base58"
)

type CommitteePubKey struct {
	IncPubKey    []byte
	MiningPubKey map[string][]byte
}

// // func (priKey *CommitteePriKey) LoadKey(consensusPriKey map[string]string) error {
// // 	priKey.PriKey = map[string][]byte{}
// // 	for key, value := range consensusPriKey {
// // 		keyBytes, ver, err := base58.Base58Check{}.Decode(value)
// // 		if (ver != common.ZeroByte) || (err != nil) {
// // 			return err
// // 		}
// // 		priKey.PriKey[key] = keyBytes
// // 	}
// // 	return nil
// // }

// // func (pubKey *CommitteePubKey) LoadKey(consensusPubKey map[string]string) error {

// // }

// func (keyset *MiningKey) GetPublicKeyBase58() string {
// 	return base58.Base58Check{}.Encode(keyset.Publickey, common.ZeroByte)
// }

func (pubKey *CommitteePubKey) FromBytes(keyBytes []byte) error {
	return json.Unmarshal(keyBytes, pubKey)
}

func (pubKey *CommitteePubKey) Bytes() ([]byte, error) {
	return json.Marshal(pubKey)
}

func (pubKey *CommitteePubKey) GetNormalKey() []byte {
	return pubKey.IncPubKey
}

func (pubKey *CommitteePubKey) GetMiningKey(schemeName string) ([]byte, error) {
	result, ok := pubKey.MiningPubKey[schemeName]
	if !ok {
		return nil, errors.New("this schemeName doesn't exist")
	}
	return result, nil
}
func (pubKey *CommitteePubKey) GetMiningKeyBase58(schemeName string) string {
	return base58.Base58Check{}.Encode(pubKey.MiningPubKey[schemeName], common.Base58Version)
}
func (pubKey *CommitteePubKey) GetIncKeyBase58() string {

	return base58.Base58Check{}.Encode(pubKey.IncPubKey, common.Base58Version)
}

func (pubKey *CommitteePubKey) ToBase58() (string, error) {
	result, err := json.Marshal(pubKey)
	if err != nil {
		return "", err
	}
	return base58.Base58Check{}.Encode(result, common.Base58Version), nil
}

func (pubKey *CommitteePubKey) FromBase58(keyStr string) error {
	keyBytes, _, err := base58.Base58Check{}.Decode(keyStr)
	if err != nil {
		return err
	}
	err = json.Unmarshal(keyBytes, pubKey)
	if err != nil {
		return err
	}
	return nil
}
