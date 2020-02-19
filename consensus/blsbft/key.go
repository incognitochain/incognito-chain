package blsbft

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/incognitokey"
)

func (engine *Engine) LoadMiningKeys(keysString string) error {
	if len(keysString) > 0 {
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

				f := &BLSBFT{}
				if engine.currentMiningProcess != nil {
					f = engine.currentMiningProcess.(*BLSBFT)
				}
				err := f.LoadUserKey(keyConsensus)
				if err != nil {
					return errors.New("Key for this consensus can not load - " + keyConsensus)
				}
				engine.userMiningPublicKeys[availableConsensus] = f.GetUserPublicKey()
			}
		}
	}
	return nil
}

func (engine *Engine) GetCurrentMiningPublicKey() (publickey string, keyType string) {
	if engine != nil {
		name := engine.consensusName
		pubkey := engine.userMiningPublicKeys[name].GetMiningKeyBase58(name)
		return pubkey, name
	}
	return "", ""
}

func (engine *Engine) GetMiningPublicKeyByConsensus(consensusName string) (publickey string, err error) {
	keyBytes := map[string][]byte{}
	if engine != nil && engine.currentMiningProcess != nil {
		keytype := engine.currentMiningProcess.GetConsensusName()
		lightweightKey, exist := engine.userMiningPublicKeys[keytype].MiningPubKey[common.BridgeConsensus]
		if !exist {
			return "", NewConsensusError(LoadKeyError, errors.New("Lightweight key not found"))
		}
		keyBytes[keytype], exist = engine.userMiningPublicKeys[keytype].MiningPubKey[keytype]
		if !exist {
			return "", NewConsensusError(LoadKeyError, errors.New("Key not found"))
		}
		keyBytes[common.BridgeConsensus] = lightweightKey
		// pubkey := engine.userMiningPublicKeys[keytype]
		// return pubkey.GetMiningKeyBase58(keytype), keytype

	}
	res, err := json.Marshal(keyBytes)
	if err != nil {
		return "", NewConsensusError(UnExpectedError, err)
	}
	return string(res), nil
}

func (engine *Engine) GetAllMiningPublicKeys() []string {
	var keys []string
	for keyType, key := range engine.userMiningPublicKeys {
		keys = append(keys, fmt.Sprintf("%v:%v", keyType, key.GetMiningKeyBase58(keyType)))
	}
	return keys
}

func (engine *Engine) SignDataWithCurrentMiningKey(
	data []byte,
) (
	publicKeyStr string,
	publicKeyType string,
	signature string,
	err error,
) {
	publicKeyStr = ""
	publicKeyType = ""
	signature = ""
	if engine != nil && engine.currentMiningProcess != nil {
		publicKeyType = engine.config.Blockchain.BestState.Beacon.ConsensusAlgorithm
		publicKeyStr, err = engine.GetMiningPublicKeyByConsensus(publicKeyType)
		if err != nil {
			return
		}
		signature, err = engine.currentMiningProcess.SignData(data)
	}
	return
}

func (engine *Engine) VerifyData(data []byte, sig string, publicKey string, consensusType string) error {
	mapPublicKey := map[string][]byte{}
	err := json.Unmarshal([]byte(publicKey), &mapPublicKey)
	if err != nil {
		return NewConsensusError(LoadKeyError, err)
	}
	return engine.currentMiningProcess.ValidateData(data, sig, string(mapPublicKey[common.BridgeConsensus]))
}

func (engine *Engine) ValidateProducerSig(block common.BlockInterface, consensusType string) error {
	return engine.currentMiningProcess.ValidateProducerSig(block)
}

func (engine *Engine) ValidateBlockCommitteSig(block common.BlockInterface, committee []incognitokey.CommitteePublicKey, consensusType string) error {
	return engine.currentMiningProcess.ValidateCommitteeSig(block, committee)
}

func (engine *Engine) GenMiningKeyFromPrivateKey(privateKey string) (string, error) {
	var keyList string
	var key string
	key = "bls"
	consensusKey, err := LoadUserKeyFromIncPrivateKey(privateKey)
	if err != nil {
		return "", err
	}
	key += ":" + consensusKey
	if len(keyList) > 0 {
		key += "|"
	}
	keyList += key
	return keyList, nil
}

func (engine *Engine) ExtractBridgeValidationData(block common.BlockInterface) ([][]byte, []int, error) {
	if engine.currentMiningProcess != nil {
		return engine.currentMiningProcess.ExtractBridgeValidationData(block)
	}
	return nil, nil, NewConsensusError(ConsensusTypeNotExistError, errors.New(block.GetConsensusType()))
}
