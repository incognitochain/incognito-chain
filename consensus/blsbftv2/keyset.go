package blsbftv2

import (
	"encoding/json"
	"sort"

	"github.com/incognitochain/incognito-chain/consensus/signatureschemes/blsmultisig"
	"github.com/incognitochain/incognito-chain/consensus/signatureschemes/bridgesig"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/common/base58"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/wallet"
)

type MiningKey struct {
	PriKey map[string][]byte
	PubKey map[string][]byte
}

func (miningKey *MiningKey) GetPublicKey() incognitokey.CommitteePublicKey {
	key := incognitokey.CommitteePublicKey{}
	key.MiningPubKey = make(map[string][]byte)
	key.MiningPubKey[common.BlsConsensus] = miningKey.PubKey[common.BlsConsensus]
	key.MiningPubKey[common.BridgeConsensus] = miningKey.PubKey[common.BridgeConsensus]
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

func (miningKey *MiningKey) BLSSignData(
	data []byte,
	selfIdx int,
	committee []blsmultisig.PublicKey,
) (
	[]byte,
	error,
) {
	sigBytes, err := blsmultisig.Sign(data, miningKey.PriKey[common.BlsConsensus], selfIdx, committee)
	if err != nil {
		return nil, NewConsensusError(SignDataError, err)
	}
	return sigBytes, nil
}

func (miningKey *MiningKey) BriSignData(
	data []byte,
) (
	[]byte,
	error,
) {
	sig, err := bridgesig.Sign(miningKey.PriKey[common.BridgeConsensus], data)
	if err != nil {
		return nil, NewConsensusError(SignDataError, err)
	}
	return sig, nil
}

func (e *BLSBFT_V2) LoadUserKey(privateSeed string) error {
	var miningKey MiningKey
	privateSeedBytes, _, err := base58.Base58Check{}.Decode(privateSeed)
	if err != nil {
		return NewConsensusError(LoadKeyError, err)
	}

	blsPriKey, blsPubKey := blsmultisig.KeyGen(privateSeedBytes)

	// privateKey := blsmultisig.B2I(privateKeyBytes)
	// publicKeyBytes := blsmultisig.PKBytes(blsmultisig.PKGen(privateKey))
	miningKey.PriKey = map[string][]byte{}
	miningKey.PubKey = map[string][]byte{}
	miningKey.PriKey[common.BlsConsensus] = blsmultisig.SKBytes(blsPriKey)
	miningKey.PubKey[common.BlsConsensus] = blsmultisig.PKBytes(blsPubKey)
	bridgePriKey, bridgePubKey := bridgesig.KeyGen(privateSeedBytes)
	miningKey.PriKey[common.BridgeConsensus] = bridgesig.SKBytes(&bridgePriKey)
	miningKey.PubKey[common.BridgeConsensus] = bridgesig.PKBytes(&bridgePubKey)
	e.UserKeySet = &miningKey
	return nil
}

func (e *BLSBFT_V2) LoadUserKeyFromIncPrivateKey(privateKey string) (string, error) {
	wl, err := wallet.Base58CheckDeserialize(privateKey)
	if err != nil {
		return "", NewConsensusError(LoadKeyError, err)
	}
	privateSeedBytes := common.HashB(common.HashB(wl.KeySet.PrivateKey))
	if err != nil {
		return "", NewConsensusError(LoadKeyError, err)
	}
	privateSeed := base58.Base58Check{}.Encode(privateSeedBytes, common.Base58Version)
	return privateSeed, nil
}

func (e *BLSBFT_V2) GetUserPublicKey() *incognitokey.CommitteePublicKey {
	if e.UserKeySet != nil {
		key := e.UserKeySet.GetPublicKey()
		return &key
	}
	return nil
}

func (e BLSBFT_V2) SignData(data []byte) (string, error) {
	result, err := e.UserKeySet.BriSignData(data) //, 0, []blsmultisig.PublicKey{e.UserKeySet.PubKey[common.BlsConsensus]})
	if err != nil {
		return "", NewConsensusError(SignDataError, err)
	}

	return base58.Base58Check{}.Encode(result, common.Base58Version), nil
}

func combineVotes(votes map[string]BFTVote, committee []string) (aggSig []byte, brigSigs [][]byte, validatorIdx []int, err error) {
	var blsSigList [][]byte
	for validator, _ := range votes {
		if index := common.IndexOfStr(validator, committee); index != -1 {
			validatorIdx = append(validatorIdx, index)
		}
	}
	sort.Ints(validatorIdx)
	for _, idx := range validatorIdx {
		blsSigList = append(blsSigList, votes[committee[idx]].BLS)
		brigSigs = append(brigSigs, votes[committee[idx]].BRI)
	}
	aggSig, err = blsmultisig.Combine(blsSigList)
	if err != nil {
		return nil, nil, nil, NewConsensusError(CombineSignatureError, err)
	}
	return
}

func GetMiningKeyFromPrivateSeed(privateSeed string) (*MiningKey, error) {
	var miningKey MiningKey
	privateSeedBytes, _, err := base58.Base58Check{}.Decode(privateSeed)
	if err != nil {
		return nil, NewConsensusError(LoadKeyError, err)
	}

	blsPriKey, blsPubKey := blsmultisig.KeyGen(privateSeedBytes)

	miningKey.PriKey = map[string][]byte{}
	miningKey.PubKey = map[string][]byte{}
	miningKey.PriKey[common.BlsConsensus] = blsmultisig.SKBytes(blsPriKey)
	miningKey.PubKey[common.BlsConsensus] = blsmultisig.PKBytes(blsPubKey)
	bridgePriKey, bridgePubKey := bridgesig.KeyGen(privateSeedBytes)
	miningKey.PriKey[common.BridgeConsensus] = bridgesig.SKBytes(&bridgePriKey)
	miningKey.PubKey[common.BridgeConsensus] = bridgesig.PKBytes(&bridgePubKey)

	return &miningKey, nil
}

func LoadUserKeyFromIncPrivateKey(privateKey string) (string, error) {
	wl, err := wallet.Base58CheckDeserialize(privateKey)
	if err != nil {
		return "", NewConsensusError(LoadKeyError, err)
	}
	privateSeedBytes := common.HashB(common.HashB(wl.KeySet.PrivateKey))
	if err != nil {
		return "", NewConsensusError(LoadKeyError, err)
	}
	privateSeed := base58.Base58Check{}.Encode(privateSeedBytes, common.Base58Version)
	return privateSeed, nil
}
