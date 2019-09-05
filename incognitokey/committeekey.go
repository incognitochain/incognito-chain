package incognitokey

import (
	"encoding/json"
	"errors"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/common/base58"
	"github.com/incognitochain/incognito-chain/consensus/signatureschemes/blsmultisig"
	"github.com/incognitochain/incognito-chain/consensus/signatureschemes/bridgesig"
	"github.com/incognitochain/incognito-chain/privacy"
)

type CommitteePublicKey struct {
	IncPubKey    privacy.PublicKey
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
		return NewCashecError(B58DecodePubKeyErr, errors.New(ErrCodeMessage[B58DecodePubKeyErr].Message))
	}
	err = json.Unmarshal(keyBytes, pubKey)
	if err != nil {
		return NewCashecError(JSONError, errors.New(ErrCodeMessage[JSONError].Message))
	}
	return nil
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
	err := json.Unmarshal(keyBytes, pubKey)
	if err != nil {
		return NewCashecError(JSONError, err)
	}
	return nil
}

func (pubKey *CommitteePublicKey) Bytes() ([]byte, error) {
	res, err := json.Marshal(pubKey)
	if err != nil {
		return []byte{0}, NewCashecError(JSONError, err)
	}
	return res, nil
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
		return errors.New("wrong input")
	}
	return json.Unmarshal(keyBytes, pubKey)
}
