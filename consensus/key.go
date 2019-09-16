package consensus

import (
	"errors"
	"fmt"
	"strings"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/incognitokey"
)

func (engine *Engine) LoadMiningKeys(keysString string) error {
	if len(keysString) > 0 {
		engine.userMiningPublicKeys = make(map[string]incognitokey.CommitteePublicKey)
		keys := strings.Split(keysString, "|")
		if len(keys) > 0 {
			for _, key := range keys {
				keyParts := strings.Split(key, ":")
				availableConsensus := common.BlsConsensus
				keyConsensus := keyParts[0]
				if len(keyParts) == 2 {
					availableConsensus = keyParts[0]
					keyConsensus = keyParts[1]
				}

				if _, ok := AvailableConsensus[availableConsensus]; ok {
					err := AvailableConsensus[availableConsensus].LoadUserKey(keyConsensus)
					if err != nil {
						return errors.New("Key for this consensus can not load - " + keyConsensus)
					}
					engine.userMiningPublicKeys[availableConsensus] = *AvailableConsensus[availableConsensus].GetUserPublicKey()
				} else {
					return errors.New("Consensus type for this key isn't exist " + availableConsensus)
				}
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
