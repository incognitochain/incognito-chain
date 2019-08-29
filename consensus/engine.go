package consensus

import (
	"errors"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/incognitochain/incognito-chain/blockchain"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/pubsub"
	"github.com/incognitochain/incognito-chain/wire"
)

var AvailableConsensus map[string]ConsensusInterface

type Engine struct {
	sync.Mutex
	cQuit                chan struct{}
	started              bool
	ChainConsensusList   map[string]ConsensusInterface
	CurrentMiningChain   string
	userMiningPublicKeys map[string]incognitokey.CommitteePublicKey
	chainCommitteeChange chan string
	config               *EngineConfig
}

type EngineConfig struct {
	Node          NodeInterface
	Blockchain    *blockchain.BlockChain
	BlockGen      *blockchain.BlockGenerator
	PubSubManager *pubsub.PubSubManager
}

func New() *Engine {
	return &Engine{}
}

func (engine *Engine) CommitteeChange(chainName string) {
	engine.chainCommitteeChange <- chainName
}

//watchConsensusState will watch MiningKey Role as well as chain consensus type
func (engine *Engine) watchConsensusCommittee() {
	Logger.log.Info("start watching consensus committee...")
	allcommittee := engine.config.Blockchain.Chains[common.BEACON_CHAINKEY].(BeaconInterface).GetAllCommittees()

	for consensusType, publickey := range engine.userMiningPublicKeys {
		if engine.CurrentMiningChain != "" {
			break
		}
		if committees, ok := allcommittee[consensusType]; ok {
			for chainName, committee := range committees {
				keys, _ := incognitokey.ExtractPublickeysFromCommitteeKeyList(committee, consensusType)
				if common.IndexOfStr(publickey.GetMiningKeyBase58(consensusType), keys) != -1 {
					engine.CurrentMiningChain = chainName
				}
			}
		}
	}
	for chainName, chain := range engine.config.Blockchain.Chains {
		if _, ok := AvailableConsensus[chain.GetConsensusType()]; ok {
			engine.ChainConsensusList[chainName] = AvailableConsensus[chain.GetConsensusType()].NewInstance(chain, chainName, engine.config.Node, Logger.log)
		}
	}

	for {
		select {
		case <-engine.cQuit:
		case chainName := <-engine.chainCommitteeChange:
			Logger.log.Critical("chain committee change", chainName)
			consensusType := engine.config.Blockchain.Chains[chainName].GetConsensusType()
			userPublicKey, ok := engine.userMiningPublicKeys[consensusType]
			if !ok {
				continue
			}
			if chainName != common.BEACON_CHAINKEY {
				role, shardID := engine.config.Blockchain.Chains[common.BEACON_CHAINKEY].GetPubkeyRole(userPublicKey.GetMiningKeyBase58(consensusType), 0)
				if role == common.SHARD_ROLE && chainName == common.GetShardChainKey(shardID) {
					engine.CurrentMiningChain = chainName
				}
			}
			if engine.config.Blockchain.Chains[chainName].GetPubKeyCommitteeIndex(userPublicKey.GetMiningKeyBase58(consensusType)) != -1 {
				if engine.CurrentMiningChain != chainName {
					engine.CurrentMiningChain = chainName
					panic("Yoesssss")
				}
			}
		}
	}
}

func (engine *Engine) Start() error {
	engine.Lock()
	defer engine.Unlock()
	if engine.started {
		return errors.New("Consensus engine is already started")
	}
	Logger.log.Info("starting consensus...")
	go engine.config.BlockGen.Start(engine.cQuit)
	go func() {
		go engine.watchConsensusCommittee()
		for {
			select {
			case <-engine.cQuit:
				return
			default:
				time.Sleep(time.Millisecond * 1000)

				for chainName, consensus := range engine.ChainConsensusList {
					if chainName == engine.CurrentMiningChain {
						Logger.log.Critical("current mining chain", chainName)
						consensus.Start()
					} else {
						consensus.Stop()
					}
				}
				userLayer := ""
				if engine.CurrentMiningChain == common.BEACON_CHAINKEY {
					userLayer = common.BEACON_ROLE
					go engine.NotifyBeaconRole(true)
					go engine.NotifyShardRole(-1)
				}
				if engine.CurrentMiningChain != common.BEACON_CHAINKEY && engine.CurrentMiningChain != "" {
					userLayer = common.SHARD_ROLE
					go engine.NotifyBeaconRole(false)
					go engine.NotifyShardRole(int(getShardFromChainName(engine.CurrentMiningChain)))
				}
				publicKey, _ := engine.GetCurrentMiningPublicKey()

				allcommittee := engine.config.Blockchain.Chains[common.BEACON_CHAINKEY].(BeaconInterface).GetAllCommittees()
				beaconCommittee := []string{}
				shardCommittee := map[byte][]string{}
				shardCommittee = make(map[byte][]string)
				for keyType, committees := range allcommittee {
					for chain, committee := range committees {
						if chain == common.BEACON_CHAINKEY {
							keyList, _ := incognitokey.ExtractPublickeysFromCommitteeKeyList(committee, keyType)
							beaconCommittee = append(beaconCommittee, keyList...)
						} else {
							keyList, _ := incognitokey.ExtractPublickeysFromCommitteeKeyList(committee, keyType)
							shardCommittee[getShardFromChainName(chain)] = keyList
						}
					}
				}
				engine.config.Node.UpdateConsensusState(userLayer, publicKey, nil, beaconCommittee, shardCommittee)
			}
		}
	}()
	return nil
}

func (engine *Engine) Stop(name string) error {
	engine.Lock()
	defer engine.Unlock()
	if !engine.started {
		return errors.New("Consensus engine is already stopped")
	}
	engine.started = false
	close(engine.cQuit)
	return nil
}

func RegisterConsensus(name string, consensus ConsensusInterface) error {
	if len(AvailableConsensus) == 0 {
		AvailableConsensus = make(map[string]ConsensusInterface)
	}
	if consensus == nil {
		return NewConsensusError(UnExpectedError, errors.New("consensus can't be nil"))
	}
	AvailableConsensus[name] = consensus
	return nil
}

func (engine *Engine) IsOngoing(chainName string) bool {
	consensusModule, ok := engine.ChainConsensusList[chainName]
	if ok {
		return consensusModule.IsOngoing()
	}
	return false
}

func (engine *Engine) OnBFTMsg(msg *wire.MessageBFT) {
	if engine.CurrentMiningChain == msg.ChainKey {
		engine.ChainConsensusList[msg.ChainKey].ProcessBFTMsg(msg)
	}
}

func (engine *Engine) GetUserRole() (string, int) {
	if engine.CurrentMiningChain != "" {
		publicKey, _ := engine.GetCurrentMiningPublicKey()
		userRole, _ := engine.config.Blockchain.Chains[engine.CurrentMiningChain].GetPubkeyRole(publicKey, 0)
		if engine.CurrentMiningChain == common.BEACON_CHAINKEY {
			return userRole, -1
		}
		return userRole, engine.config.Blockchain.Chains[engine.CurrentMiningChain].GetShardID()
	}
	return "", 0
}

func getShardFromChainName(chainName string) byte {
	s := strings.Split(chainName, "-")[1]
	s1, _ := strconv.Atoi(s)
	return byte(s1)
}

func (engine *Engine) NotifyBeaconRole(beaconRole bool) {
	engine.config.PubSubManager.PublishMessage(pubsub.NewMessage(pubsub.BeaconRoleTopic, beaconRole))
}
func (engine *Engine) NotifyShardRole(shardRole int) {
	engine.config.PubSubManager.PublishMessage(pubsub.NewMessage(pubsub.ShardRoleTopic, shardRole))
}

func (engine *Engine) Init(config *EngineConfig) error {
	if config.BlockGen == nil {
		return NewConsensusError(UnExpectedError, errors.New("BlockGen can't be nil"))
	}
	if config.Blockchain == nil {
		return NewConsensusError(UnExpectedError, errors.New("Blockchain can't be nil"))
	}
	if config.Node == nil {
		return NewConsensusError(UnExpectedError, errors.New("Node can't be nil"))
	}
	if config.PubSubManager == nil {
		return NewConsensusError(UnExpectedError, errors.New("PubSubManager can't be nil"))
	}
	engine.config = config
	engine.cQuit = make(chan struct{})
	engine.chainCommitteeChange = make(chan string)
	engine.ChainConsensusList = make(map[string]ConsensusInterface)
	if config.Node.GetPrivateKey() != "" {
		keyList, err := engine.GenMiningKeyFromPrivateKey(config.Node.GetPrivateKey())
		if err != nil {
			panic(err)
		}
		err = engine.LoadMiningKeys(keyList)
		if err != nil {
			panic(err)
		}
	} else {
		err := engine.LoadMiningKeys(config.Node.GetMiningKeys())
		if err != nil {
			panic(err)
		}
	}
	return nil
}

func (engine *Engine) ExtractBridgeValidationData(block common.BlockInterface) ([][]byte, []int, error) {
	if _, ok := AvailableConsensus[block.GetConsensusType()]; ok {
		return AvailableConsensus[block.GetConsensusType()].ExtractBridgeValidationData(block)
	}
	return nil, nil, NewConsensusError(ConsensusTypeNotExistError, errors.New(block.GetConsensusType()))
}

// func (engine *Engine) SwitchConsensus(chainkey string, consensus string) error {
// 	if engine.ChainConsensusList[common.BEACON_CHAINKEY].GetConsensusName() != engine.config.Blockchain.BestState.Beacon.ConsensusAlgorithm {
// 		consensus, ok := AvailableConsensus[engine.ChainConsensusList[common.BEACON_CHAINKEY].GetConsensusName()]
// 		if ok {
// 			engine.ChainConsensusList[common.BEACON_CHAINKEY] = consensus.NewInstance(engine.config.Blockchain.Chains[common.BEACON_CHAINKEY], chainkey, engine.config.Node, Logger.log)
// 		} else {
// 			panic("Update code please")
// 		}
// 	}
// 	for idx := 0; idx < engine.config.Blockchain.BestState.Beacon.ActiveShards; idx++ {
// 		shard, ok := engine.config.Blockchain.BestState.Shard[byte(idx)]
// 		if ok {
// 			chainKey := common.GetShardChainKey(byte(idx))
// 			if shard.ConsensusAlgorithm != engine.ChainConsensusList[chainKey].GetConsensusName() {
// 				consensus, ok := AvailableConsensus[engine.ChainConsensusList[chainKey].GetConsensusName()]
// 				if ok {
// 					engine.ChainConsensusList[chainKey] = consensus.NewInstance(engine.config.Blockchain.Chains[chainKey], chainkey, engine.config.Node, Logger.log)
// 				} else {
// 					panic("Update code please")
// 				}
// 			}
// 		} else {
// 			panic("Oops... Maybe a bug cause this, please update code")
// 		}
// 	}
// 	return nil
// }
