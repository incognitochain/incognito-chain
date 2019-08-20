package blsbft

import (
	"encoding/json"

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
	ConsensusPriKey map[string][]byte
	ConsensusPubKey map[string][]byte
}

func (miningKey *MiningKey) GetPublicKeyBase58() string {
	keyBytes, err := json.Marshal(miningKey.ConsensusPubKey)
	if err != nil {
		return ""
	}
	return base58.Base58Check{}.Encode(keyBytes, common.ZeroByte)
}

func (miningKey *MiningKey) GetPrivateKeyBase58() string {
	keyBytes, err := json.Marshal(miningKey.ConsensusPriKey)
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

func (miningKey *MiningKey) BLSSignData(data []byte, bridgeSig bool, listCommittee []pu) (string, error) {
	sigs := map[string]string{}

	return nil, nil
}

func (miningKey *MiningKey) validateSingleBLSSig(
	dataHash *common.Hash,
	blsSig string,
) error {
	return nil
}

func (miningKey *MiningKey) validateSingleBriSig(
	dataHash *common.Hash,
	blsSig string,
) error {
	return nil
}

func (miningKey *MiningKey) validateAggregatedSig(dataHash *common.Hash, aggSig string, validatorPubkeyList []string) error {
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
	miningKey.ConsensusPriKey = map[string][]byte{}
	miningKey.ConsensusPubKey = map[string][]byte{}
	miningKey.ConsensusPriKey[BLS] = privateKeyBytes
	miningKey.ConsensusPubKey[BLS] = publicKeyBytes
	bridgePriKey, bridgePubKey := bridgesig.KeyGen(privateKeyBytes)
	miningKey.ConsensusPriKey[BRI] = bridgesig.PriKeyToBytes(&bridgePriKey)
	miningKey.ConsensusPubKey[BRI] = bridgesig.PubKeyToBytes(&bridgePubKey)
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
