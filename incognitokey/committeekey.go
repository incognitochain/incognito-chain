package incognitokey

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/common/base58"
	"github.com/incognitochain/incognito-chain/consensus/blsmultisig"
	"github.com/incognitochain/incognito-chain/consensus/bridgesig"
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

func (pubKey *CommitteePubKey) CheckSanityData() bool {
	if (len(pubKey.IncPubKey) != common.PublicKeySize) ||
		(len(pubKey.MiningPubKey[common.BLS_CONSENSUS]) != common.BLSPublicKeySize) ||
		(len(pubKey.MiningPubKey[common.BRI_CONSENSUS]) != common.BriPublicKeySize) {
		return false
	}
	return true
}

func (pubKey *CommitteePubKey) FromString(keyString string) error {
	keyBytes, ver, err := base58.Base58Check{}.Decode(keyString)
	if (ver != common.ZeroByte) || (err != nil) {
		return errors.New("Wrong input")
	}
	fmt.Println(keyBytes)
	return json.Unmarshal(keyBytes, pubKey)
}

func NewCommitteeKeyFromSeed(seed, incPubKey []byte) (CommitteePubKey, error) {
	committeePubKey := new(CommitteePubKey)
	committeePubKey.IncPubKey = incPubKey
	committeePubKey.MiningPubKey = map[string][]byte{}
	_, blsPubKey := blsmultisig.KeyGen(seed)
	blsPubKeyBytes := blsmultisig.PKBytes(blsPubKey)
	committeePubKey.MiningPubKey[common.BLS_CONSENSUS] = blsPubKeyBytes
	_, briPubKey := bridgesig.KeyGen(seed)
	briPubKeyBytes := bridgesig.PKBytes(&briPubKey)
	committeePubKey.MiningPubKey[common.BRI_CONSENSUS] = briPubKeyBytes
	return *committeePubKey, nil
}

func (pubKey *CommitteePubKey) FromBytes(keyBytes []byte) error {
	return json.Unmarshal(keyBytes, pubKey)
}

func (pubKey *CommitteePubKey) Bytes() ([]byte, error) {

	return json.Marshal(pubKey)
}

func (pubKey *CommitteePubKey) GetNormalKey() []byte {
	return pubKey.IncPubKey
}

func (pubKey *CommitteePubKey) GetMiningKey(schemeName string) []byte {
	return pubKey.MiningPubKey[schemeName]
}
