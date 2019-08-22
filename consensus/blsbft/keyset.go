package blsbft

import (
	"encoding/json"
	"errors"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/common/base58"
	"github.com/incognitochain/incognito-chain/consensus/blsmultisig"
	"github.com/incognitochain/incognito-chain/consensus/bridgesig"
	"github.com/incognitochain/incognito-chain/incognitokey"
)

// type blsKeySet struct {
// 	Publickey  []byte
// 	PrivateKey []byte
// }

type MiningKey struct {
	PriKey map[string][]byte
	PubKey map[string][]byte
}

func (miningKey *MiningKey) GetPublicKey() incognitokey.CommitteePubKey {
	key := incognitokey.CommitteePubKey{}
	key.MiningPubKey = make(map[string][]byte)
	key.MiningPubKey[common.BLS_CONSENSUS] = miningKey.PubKey[BLS]
	key.MiningPubKey[common.BRI_CONSENSUS] = miningKey.PubKey[BRI]
	return key
}

func (miningKey *MiningKey) GetPublicKeyBase58() string {
	key := miningKey.GetPublicKey()
	keyBytes, err := json.Marshal(key)
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
	[]byte,
	error,
) {
	sigBytes, err := blsmultisig.Sign(data, miningKey.PriKey[BLS], selfIdx, committee)
	if err != nil {
		return nil, err
	}
	return sigBytes, nil
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

func validateSingleBLSSig(
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

func validateSingleBriSig(
	dataHash *common.Hash,
	blsSig string,
) error {
	return nil
}

func validateBLSSig(
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

func (e *BLSBFT) LoadUserKey(privateSeed string) error {
	var miningKey MiningKey
	privateSeedBytes, _, err := base58.Base58Check{}.Decode(privateSeed)
	if err != nil {
		return err
	}

	blsPriKey, blsPubKey := blsmultisig.KeyGen(privateSeedBytes)

	// privateKey := blsmultisig.B2I(privateKeyBytes)
	// publicKeyBytes := blsmultisig.PKBytes(blsmultisig.PKGen(privateKey))
	miningKey.PriKey = map[string][]byte{}
	miningKey.PubKey = map[string][]byte{}
	miningKey.PriKey[BLS] = blsmultisig.SKBytes(blsPriKey)
	miningKey.PubKey[BLS] = blsmultisig.PKBytes(blsPubKey)
	bridgePriKey, bridgePubKey := bridgesig.KeyGen(privateSeedBytes)
	miningKey.PriKey[BRI] = bridgesig.SKBytes(&bridgePriKey)
	miningKey.PubKey[BRI] = bridgesig.PKBytes(&bridgePubKey)
	e.UserKeySet = &miningKey
	return nil
}
func (e BLSBFT) GetUserPublicKey() *incognitokey.CommitteePubKey {
	if e.UserKeySet != nil {
		key := e.UserKeySet.GetPublicKey()
		return &key
	}
	return nil
}

func (e BLSBFT) GetUserPrivateKey() string {
	if e.UserKeySet != nil {
		return e.UserKeySet.GetPrivateKeyBase58()
	}
	return ""
}

func combineSigs(sigs [][]byte) ([]byte, error) {
	return blsmultisig.Combine(sigs)
}
