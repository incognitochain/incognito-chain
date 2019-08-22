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

type CommitteePublicKey struct {
	IncPubKey    []byte
	MiningPubKey map[string][]byte
}

func (pubKey *CommitteePublicKey) CheckSanityData() bool {
	if (len(pubKey.IncPubKey) != common.PublicKeySize) ||
		(len(pubKey.MiningPubKey[common.BLS_CONSENSUS]) != common.BLSPublicKeySize) ||
		(len(pubKey.MiningPubKey[common.BRI_CONSENSUS]) != common.BriPublicKeySize) {
		return false
	}
	return true
}

func (pubKey *CommitteePublicKey) FromString(keyString string) error {
	keyBytes, ver, err := base58.Base58Check{}.Decode(keyString)
	if (ver != common.ZeroByte) || (err != nil) {
		return errors.New("Wrong input")
	}
	fmt.Println(keyBytes)
	return json.Unmarshal(keyBytes, pubKey)
}

func NewCommitteeKeyFromSeed(seed, incPubKey []byte) (CommitteePublicKey, error) {
	committeePubKey := new(CommitteePublicKey)
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

func (pubKey *CommitteePublicKey) FromBytes(keyBytes []byte) error {
	return json.Unmarshal(keyBytes, pubKey)
}

func (pubKey *CommitteePublicKey) Bytes() ([]byte, error) {

	return json.Marshal(pubKey)
}

func (pubKey *CommitteePublicKey) GetNormalKey() []byte {
	return pubKey.IncPubKey
}

func (pubKey *CommitteePublicKey) GetMiningKey(schemeName string) []byte {
	return pubKey.MiningPubKey[schemeName]
}
