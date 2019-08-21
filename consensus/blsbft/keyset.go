package blsbft

import (
	"encoding/json"
	"errors"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/common/base58"
	"github.com/incognitochain/incognito-chain/consensus/blsmultisig"
	"github.com/incognitochain/incognito-chain/consensus/bridgesig"
)

// type blsKeySet struct {
// 	Publickey  []byte
// 	PrivateKey []byte
// }

type MiningKey struct {
	PriKey map[string][]byte
	PubKey map[string][]byte
}

func (miningKey *MiningKey) GetPublicKeyBase58() string {
	keyBytes, err := json.Marshal(miningKey.PubKey)
	if err != nil {
		return ""
	}
	return base58.Base58Check{}.Encode(keyBytes, common.ZeroByte)
}

func (miningKey *MiningKey) GetPrivateKeyBase58() string {
	keyBytes, err := json.Marshal(miningKey.PriKey)
	if err != nil {
		return ""
	}
	return base58.Base58Check{}.Encode(keyBytes, common.ZeroByte)
}

// func (keyset *blsKeySet) GetPublicKeyBase58() string {
// 	return base58.Base58Check{}.Encode(keyset.Publickey, common.ZeroByte)
// }
// func (keyset *blsKeySet) GetPrivateKeyBase58() string {
// 	return base58.Base58Check{}.Encode(keyset.PrivateKey, common.ZeroByte)
// }

func (miningKey *MiningKey) BLSSignData(
	data []byte,
	selfIdx int,
	committee []blsmultisig.PublicKey,
) (
	string,
	error,
) {
	sigBytes, err := blsmultisig.Sign(data, miningKey.PriKey[BLS], selfIdx, committee)
	if err != nil {
		return "", err
	}
	sig := base58.Base58Check{}.Encode(sigBytes, common.ZeroByte)
	return sig, nil
}

func (miningKey *MiningKey) BriSignData(
	data []byte,
) (
	string,
	error,
) {
	sig, err := bridgesig.Sign(data, miningKey.PriKey[BRI])
	if err != nil {
		return "", err
	}
	return sig, nil
}

func (miningKey *MiningKey) validateSingleBLSSig(
	dataHash *common.Hash,
	blsSig string,
	selfIdx int,
	committee []blsmultisig.PublicKey,
) error {
	sigBytes, ver, err := base58.Base58Check{}.Decode(blsSig)
	if err != nil {
		return err
	}
	if ver != common.ZeroByte {
		return errors.New("Decode failed")
	}
	result, err := blsmultisig.Verify(sigBytes, dataHash.GetBytes(), []int{selfIdx}, committee)
	if err != nil {
		return err
	}
	if !result {
		return errors.New("Invalid Signature!")
	}
	return nil
}

func (miningKey *MiningKey) validateSingleBriSig(
	dataHash *common.Hash,
	blsSig string,
) error {

	return nil
}

func (miningKey *MiningKey) validateBLSSig(
	dataHash *common.Hash,
	aggSig string,
	validatorIdx []int,
	committee []blsmultisig.PublicKey,
) error {
	sigBytes, ver, err := base58.Base58Check{}.Decode(aggSig)
	if err != nil {
		return err
	}
	if ver != common.ZeroByte {
		return errors.New("Decode failed")
	}
	result, err := blsmultisig.Verify(sigBytes, dataHash.GetBytes(), validatorIdx, committee)
	if err != nil {
		return err
	}
	if !result {
		return errors.New("Invalid Signature!")
	}
	return nil
}

// func (keyset *blsKeySet) SignData(data []byte) (string, error) {
// 	return "", nil
// }
// func (keyset *blsKeySet) validateAggregatedSig(dataHash *common.Hash, aggSig string, validatorPubkeyList []string) error {
// 	return nil
// }
// func (keyset *blsKeySet) validateSingleSig(dataHash *common.Hash, sig string, pubkey string) error {
// 	return nil
// }

func (e *BLSBFT) LoadUserKey(privateKeyStr string) error {
	var miningKey MiningKey
	privateKeyBytes, _, err := base58.Base58Check{}.Decode(privateKeyStr)
	if err != nil {
		return err
	}
	privateKey := blsmultisig.B2I(privateKeyBytes)
	publicKeyBytes := blsmultisig.PKBytes(blsmultisig.PKGen(privateKey))
	miningKey.PriKey = map[string][]byte{}
	miningKey.PubKey = map[string][]byte{}
	miningKey.PriKey[BLS] = privateKeyBytes
	miningKey.PubKey[BLS] = publicKeyBytes
	bridgePriKey, bridgePubKey := bridgesig.KeyGen(privateKeyBytes)
	miningKey.PriKey[BRI] = bridgesig.SKBytes(&bridgePriKey)
	miningKey.PubKey[BRI] = bridgesig.PKBytes(&bridgePubKey)
	e.UserKeySet = &miningKey
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
