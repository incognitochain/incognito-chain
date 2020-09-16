package blsbftv2

import (
	signatureschemes2 "github.com/incognitochain/incognito-chain/consensus_multi/signatureschemes"
	"sort"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/common/base58"
	"github.com/incognitochain/incognito-chain/consensus_multi/signatureschemes/blsmultisig"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/wallet"
)

func (e *BLSBFT_V2) LoadUserKeys(miningKey []signatureschemes2.MiningKey) error {
	e.UserKeySet = &miningKey[0]
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
		return key
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
		validatorIdx = append(validatorIdx, common.IndexOfStr(validator, committee))
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
