package incognitokey

import (
	"encoding/json"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/common/base58"
	"github.com/incognitochain/incognito-chain/consensus/signatureschemes/blsmultisig"
	"github.com/incognitochain/incognito-chain/consensus/signatureschemes/bridgesig"
	"github.com/pkg/errors"
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
	return json.Unmarshal(keyBytes, pubKey)
}

func NewCommitteeKeyFromSeed(seed, incPubKey []byte) (CommitteePublicKey, error) {
	CommitteePublicKey := new(CommitteePublicKey)
	CommitteePublicKey.IncPubKey = incPubKey
	CommitteePublicKey.MiningPubKey = map[string][]byte{}
	_, blsPubKey := blsmultisig.KeyGen(seed)
	blsPubKeyBytes := blsmultisig.PKBytes(blsPubKey)
	CommitteePublicKey.MiningPubKey[common.BLS_CONSENSUS] = blsPubKeyBytes
	_, briPubKey := bridgesig.KeyGen(seed)
	briPubKeyBytes := bridgesig.PKBytes(&briPubKey)
	CommitteePublicKey.MiningPubKey[common.BRI_CONSENSUS] = briPubKeyBytes
	return *CommitteePublicKey, nil
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

func (pubKey *CommitteePublicKey) GetMiningKey(schemeName string) ([]byte, error) {
	result, ok := pubKey.MiningPubKey[schemeName]
	if !ok {
		return nil, errors.New("this schemeName doesn't exist")
	}
	return result, nil
}

func (pubKey *CommitteePublicKey) GetMiningKeyBase58(schemeName string) string {
	keyBytes, ok := pubKey.MiningPubKey[schemeName]
	if !ok {
		return ""
	}
	return base58.Base58Check{}.Encode(keyBytes, common.Base58Version)
}

func (pubKey *CommitteePublicKey) GetIncKeyBase58() string {
	return base58.Base58Check{}.Encode(pubKey.IncPubKey, common.Base58Version)
}

func (pubKey *CommitteePublicKey) ToBase58() (string, error) {
	result, err := json.Marshal(pubKey)
	if err != nil {
		return "", err
	}
	return base58.Base58Check{}.Encode(result, common.Base58Version), nil
}

func (pubKey *CommitteePublicKey) FromBase58(keyString string) error {
	keyBytes, ver, err := base58.Base58Check{}.Decode(keyString)
	if (ver != common.ZeroByte) || (err != nil) {
		return errors.New("Wrong input")
	}
	return json.Unmarshal(keyBytes, pubKey)
}
