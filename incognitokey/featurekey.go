package incognitokey

import (
	"encoding/json"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/common/base58"
	"github.com/incognitochain/incognito-chain/privacy/operation"
	"math/big"
)

// OTDepositKey represents a pair of one-time depositing key for shielding.
type OTDepositKey struct {
	PrivateKey []byte
	PublicKey  []byte
	Index      uint64
}

func (k OTDepositKey) MarshalJSON() ([]byte, error) {
	type holder struct {
		PrivateKey string
		PublicKey  string
		Index      uint64
	}

	privateKeyStr := base58.Base58Check{}.NewEncode(k.PrivateKey, 0)
	pubKeyStr := base58.Base58Check{}.NewEncode(k.PublicKey, 0)
	h := holder{
		PrivateKey: privateKeyStr,
		PublicKey:  pubKeyStr,
		Index:      k.Index,
	}

	return json.Marshal(h)
}

func (k *OTDepositKey) UnmarshalJSON(data []byte) error {
	type holder struct {
		PrivateKey string
		PublicKey  string
		Index      uint64
	}
	var tmpH holder
	err := json.Unmarshal(data, &tmpH)
	if err != nil {
		return err
	}

	privateKey, _, err := base58.Base58Check{}.Decode(tmpH.PrivateKey)
	if err != nil {
		return err
	}
	pubKey, _, err := base58.Base58Check{}.Decode(tmpH.PublicKey)
	if err != nil {
		return nil
	}

	k.Index = tmpH.Index
	k.PublicKey = pubKey
	k.PrivateKey = privateKey

	return nil
}

// GenerateOTDepositKey generates a new OTDepositKey from the keySet with the given tokenID and index.
func (keySet *KeySet) GenerateOTDepositKey(tokenIDStr string, index uint64) (*OTDepositKey, error) {
	tokenID, err := new(common.Hash).NewHashFromStr(tokenIDStr)
	if err != nil {
		return nil, err
	}

	tmp := append([]byte(common.PortalV4DepositKeyGenSeed), tokenID[:]...)
	masterDepositSeed := common.SHA256(append(keySet.PrivateKey[:], tmp...))
	indexBig := new(big.Int).SetUint64(index)

	privateKey := operation.HashToScalar(append(masterDepositSeed, indexBig.Bytes()...))
	pubKey := new(operation.Point).ScalarMult(operation.PedCom.G[operation.PedersenPrivateKeyIndex], privateKey)

	return &OTDepositKey{
		PrivateKey: privateKey.ToBytesS(),
		PublicKey:  pubKey.ToBytesS(),
		Index:      index,
	}, nil
}

// GenerateOTDepositKeyFromPrivateKey generates a new OTDepositKey from the given privateKey, tokenID and index.
func GenerateOTDepositKeyFromPrivateKey(incPrivateKey []byte, tokenIDStr string, index uint64) (*OTDepositKey, error) {
	tokenID, err := new(common.Hash).NewHashFromStr(tokenIDStr)
	if err != nil {
		return nil, err
	}

	tmp := append([]byte(common.PortalV4DepositKeyGenSeed), tokenID[:]...)
	masterDepositSeed := common.SHA256(append(incPrivateKey[:], tmp...))
	indexBig := new(big.Int).SetUint64(index)

	privateKey := operation.HashToScalar(append(masterDepositSeed, indexBig.Bytes()...))
	pubKey := new(operation.Point).ScalarMult(operation.PedCom.G[operation.PedersenPrivateKeyIndex], privateKey)

	return &OTDepositKey{
		PrivateKey: privateKey.ToBytesS(),
		PublicKey:  pubKey.ToBytesS(),
		Index:      index,
	}, nil
}

// GenerateOTDepositKeyFromMasterDepositSeed generates a new OTDepositKey from the given masterDepositSeed, tokenID and index.
func GenerateOTDepositKeyFromMasterDepositSeed(masterDepositSeed []byte, index uint64) (*OTDepositKey, error) {
	indexBig := new(big.Int).SetUint64(index)

	privateKey := operation.HashToScalar(append(masterDepositSeed, indexBig.Bytes()...))
	pubKey := new(operation.Point).ScalarMult(operation.PedCom.G[operation.PedersenPrivateKeyIndex], privateKey)

	return &OTDepositKey{
		PrivateKey: privateKey.ToBytesS(),
		PublicKey:  pubKey.ToBytesS(),
		Index:      index,
	}, nil
}
