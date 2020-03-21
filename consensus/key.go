package consensus

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/incognitochain/incognito-chain/blockchain"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/consensus/blsbft"
	"github.com/incognitochain/incognito-chain/consensus/blsbftv2"
	"github.com/incognitochain/incognito-chain/incognitokey"
)

func (engine *Engine) LoadMiningKeys(keysString string) error {
	fmt.Println("Start LoadMiningKey keysString", keysString)
	defer fmt.Println("End LoadMiningKey")
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

				var f ConsensusInterface
				if engine.version == 1 {
					f = &blsbft.BLSBFT{}
					if engine.currentMiningProcess != nil {
						f = engine.currentMiningProcess.(*blsbft.BLSBFT)
					}
				} else {
					f = &blsbftv2.BLSBFT_V2{}
					if engine.currentMiningProcess != nil {
						f = engine.currentMiningProcess.(*blsbftv2.BLSBFT_V2)
					}
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
	if engine != nil && engine.userMiningPublicKeys[engine.consensusName] != nil {
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
			return "", blsbft.NewConsensusError(blsbft.LoadKeyError, errors.New("Lightweight key not found"))
		}
		keyBytes[keytype], exist = engine.userMiningPublicKeys[keytype].MiningPubKey[keytype]
		if !exist {
			return "", blsbft.NewConsensusError(blsbft.LoadKeyError, errors.New("Key not found"))
		}
		keyBytes[common.BridgeConsensus] = lightweightKey
		// pubkey := engine.userMiningPublicKeys[keytype]
		// return pubkey.GetMiningKeyBase58(keytype), keytype

	}
	res, err := json.Marshal(keyBytes)
	if err != nil {
		return "", blsbft.NewConsensusError(blsbft.UnExpectedError, err)
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
		publicKeyType = engine.config.Blockchain.GetBeaconBestState().ConsensusAlgorithm
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
		return blsbft.NewConsensusError(blsbft.LoadKeyError, err)
	}
	return engine.currentMiningProcess.ValidateData(data, sig, string(mapPublicKey[common.BridgeConsensus]))
}

func (engine *Engine) ValidateProducerPosition(blk common.BlockInterface, committee []incognitokey.CommitteePublicKey) error {

	//check producer,proposer,agg sig with this version
	if blk.GetVersion() == 1 {
		producerPosition := 0
		//validate producer
		producer := blk.GetProducer()
		tmpProducer, _ := committee[producerPosition].ToBase58()
		if tmpProducer != producer {
			return fmt.Errorf("Producer should be %+v but get %v index %v", tmpProducer, producer, producerPosition)
		}
	} else {

		//validate producer
		producer := blk.GetProducer()
		produceTime := blk.GetProduceTime()
		tempProducer := blockchain.GetProposerByTimeSlot(common.CalculateTimeSlot(produceTime), committee)
		b58Str, _ := tempProducer.ToBase58()
		if strings.Compare(b58Str, producer) != 0 {
			return fmt.Errorf("Expect Producer Public Key to be equal but get %+v From Index, %+v From Header", b58Str, producer)
		}

		//validate proposer
		proposer := blk.GetProposer()
		proposeTime := blk.GetProposeTime()
		tempProducer = blockchain.GetProposerByTimeSlot(common.CalculateTimeSlot(proposeTime), committee)
		b58Str, _ = tempProducer.ToBase58()
		if strings.Compare(b58Str, proposer) != 0 {
			return fmt.Errorf("Expect Proposer Public Key to be equal but get %+v From Index, %+v From Header", b58Str, proposer)
		}
	}

	return nil
}

func (engine *Engine) ValidateProducerSig(block common.BlockInterface, consensusType string) error {
	if block.GetVersion() == 1 {
		return blsbft.ValidateProducerSig(block)
	} else if block.GetVersion() == 2 {
		return blsbftv2.ValidateProducerSig(block)
	}
	return fmt.Errorf("Wrong block version: %v", block.GetVersion())
}

func (engine *Engine) ValidateBlockCommitteSig(block common.BlockInterface, committee []incognitokey.CommitteePublicKey) error {
	if block.GetVersion() == 1 {
		return blsbft.ValidateCommitteeSig(block, committee)
	} else if block.GetVersion() == 2 {
		return blsbftv2.ValidateCommitteeSig(block, committee)
	}
	return fmt.Errorf("Wrong block version: %v", block.GetVersion())
}

func (engine *Engine) GenMiningKeyFromPrivateKey(privateKey string) (string, error) {
	var keyList string
	var key string
	key = "bls"
	consensusKey, err := blsbft.LoadUserKeyFromIncPrivateKey(privateKey)
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
	return nil, nil, blsbft.NewConsensusError(blsbft.ConsensusTypeNotExistError, errors.New(block.GetConsensusType()))
}

func (engine *Engine) GetCurrentConsensusVersion() int {
	return engine.version
}
