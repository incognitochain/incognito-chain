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

func (pubKey *CommitteePubKey) CheckSanityData() bool {
	if (len(pubKey.IncPubKey) != common.PublicKeySize) ||
		(len(pubKey.MiningPubKey[common.BLS_CONSENSUS]) != common.BLSPublicKeySize) ||
		(len(pubKey.MiningPubKey[common.BRI_CONSENSUS]) != common.BriPublicKeySize) {
		return false
	}
	return true
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

func (pubKey *CommitteePubKey) GetMiningKey(schemeName string) ([]byte, error) {
	result, ok := pubKey.MiningPubKey[schemeName]
	if !ok {
		return nil, errors.New("this schemeName doesn't exist")
	}
	return result, nil
}

func (pubKey *CommitteePubKey) GetMiningKeyBase58(schemeName string) string {
	keyBytes, ok := pubKey.MiningPubKey[schemeName]
	if !ok {
		return ""
	}
	return base58.Base58Check{}.Encode(keyBytes, common.Base58Version)
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

func (pubKey *CommitteePubKey) FromBase58(keyString string) error {
	keyBytes, ver, err := base58.Base58Check{}.Decode(keyString)
	if (ver != common.ZeroByte) || (err != nil) {
		return errors.New("Wrong input")
	}
	return json.Unmarshal(keyBytes, pubKey)
}
