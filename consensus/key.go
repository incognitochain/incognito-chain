package consensus

import (
	"errors"
	"fmt"
	"strings"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/incognitokey"
)

func (engine *Engine) LoadMiningKeys(keysString string) error {
	engine.userMiningPublicKeys = make(map[string]incognitokey.CommitteePublicKey)
	keys := strings.Split(keysString, "|")
	for _, key := range keys {
		keyParts := strings.Split(key, ":")
		if len(keyParts) == 2 {
			if _, ok := AvailableConsensus[keyParts[0]]; ok {
				err := AvailableConsensus[keyParts[0]].LoadUserKey(keyParts[1])
				if err != nil {
					panic(err)
				}
				engine.userMiningPublicKeys[keyParts[0]] = *AvailableConsensus[keyParts[0]].GetUserPublicKey()
			} else {
				return errors.New("Consensus type for this key isn't exist " + keyParts[0])
			}
		}
	}
	return nil
}
func (engine *Engine) GetCurrentMiningPublicKey() (publickey string, keyType string) {
	if engine != nil && engine.CurrentMiningChain != "" {
		if _, ok := engine.ChainConsensusList[engine.CurrentMiningChain]; ok {
			keytype := engine.ChainConsensusList[engine.CurrentMiningChain].GetConsensusName()
			pubkey := engine.userMiningPublicKeys[keytype]
			return pubkey.GetMiningKeyBase58(keytype), keytype
		}
	}
	return "", ""
}

func (engine *Engine) GetAllMiningPublicKeys() []string {
	var keys []string
	for keyType, key := range engine.userMiningPublicKeys {
		keys = append(keys, fmt.Sprintf("%v:%v", keyType, key.GetMiningKeyBase58(keyType)))
	}
	return keys
}

func (engine *Engine) SignDataWithCurrentMiningKey(data []byte) (string, error) {
	if engine != nil && engine.CurrentMiningChain != "" {
		if _, ok := engine.ChainConsensusList[engine.CurrentMiningChain]; ok {
			keytype := engine.ChainConsensusList[engine.CurrentMiningChain].GetConsensusName()
			return AvailableConsensus[keytype].SignData(data)
		}
	}
	return "", errors.New("oops")
}

func (engine *Engine) VerifyData(data []byte, sig string, publicKey string, consensusType string) error {
	if _, ok := AvailableConsensus[consensusType]; !ok {
		return NewConsensusError(ConsensusTypeNotExistError, errors.New(consensusType))
	}
	return AvailableConsensus[consensusType].ValidateData(data, sig, publicKey)
}

func (engine *Engine) ValidateProducerSig(block common.BlockInterface, consensusType string) error {
	if _, ok := AvailableConsensus[consensusType]; !ok {
		return NewConsensusError(ConsensusTypeNotExistError, errors.New(consensusType))
	}
	return AvailableConsensus[consensusType].ValidateProducerSig(block)
}

func (engine *Engine) ValidateBlockCommitteSig(block common.BlockInterface, committee []incognitokey.CommitteePublicKey, consensusType string) error {
	if _, ok := AvailableConsensus[consensusType]; !ok {
		return NewConsensusError(ConsensusTypeNotExistError, errors.New(consensusType))
	}
	return AvailableConsensus[consensusType].ValidateCommitteeSig(block, committee)
}

func (engine *Engine) GenMiningKeyFromPrivateKey(privateKey string) (string, error) {
	var keyList string
	for consensusType, consensus := range AvailableConsensus {
		var key string
		key = consensusType
		consensusKey, err := consensus.LoadUserKeyFromIncPrivateKey(privateKey)
		if err != nil {
			return "", err
		}
		key += ":" + consensusKey
		if len(keyList) > 0 {
			key += "|"
		}
		keyList += key
	}
	return keyList, nil
}
