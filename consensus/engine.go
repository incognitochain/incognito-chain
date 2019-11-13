package consensus

import (
	"errors"
	"fmt"
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
	userCurrentState     struct {
		UserLayer  string
		UserRole   string
		ShardID    byte
		Keys       *incognitokey.CommitteePublicKey
		KeysBase58 map[string]string
	}
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
	allcommittee := engine.config.Blockchain.Chains[common.BeaconChainKey].(BeaconInterface).GetAllCommittees()

	for consensusType, publickey := range engine.userMiningPublicKeys {
		if committees, ok := allcommittee[consensusType]; ok {
			for chainName, committee := range committees {
				keys, _ := incognitokey.ExtractPublickeysFromCommitteeKeyList(committee, consensusType)
				if common.IndexOfStr(publickey.GetMiningKeyBase58(consensusType), keys) != -1 {
					engine.CurrentMiningChain = chainName
					var userRole, userLayer string
					var shardID byte
					if chainName != common.BeaconChainKey {
						userLayer = common.ShardRole
						userRole = common.CommitteeRole
						shardID = getShardFromChainName(chainName)
					} else {
						userLayer = common.BeaconRole
						userRole = common.CommitteeRole
					}
					engine.updateUserState(&publickey, userLayer, userRole, shardID)
					break
				}
			}
		}
	}

	if engine.CurrentMiningChain == "" {

		shardsPendingLists := engine.config.Blockchain.Chains[common.BeaconChainKey].(BeaconInterface).GetShardsPendingList()

		for consensusType, publickey := range engine.userMiningPublicKeys {
			beaconPendingList, _ := incognitokey.ExtractPublickeysFromCommitteeKeyList(engine.config.Blockchain.Chains[common.BeaconChainKey].(BeaconInterface).GetBeaconPendingList(), consensusType)
			beaconWaitingList, _ := incognitokey.ExtractPublickeysFromCommitteeKeyList(engine.config.Blockchain.Chains[common.BeaconChainKey].(BeaconInterface).GetBeaconWaitingList(), consensusType)
			shardsWaitingList, _ := incognitokey.ExtractPublickeysFromCommitteeKeyList(engine.config.Blockchain.Chains[common.BeaconChainKey].(BeaconInterface).GetShardsWaitingList(), consensusType)

			var shardsPendingList map[string][]string
			shardsPendingList = make(map[string][]string)

			for chainName, committee := range shardsPendingLists[consensusType] {
				shardsPendingList[chainName], _ = incognitokey.ExtractPublickeysFromCommitteeKeyList(committee, consensusType)
			}

			if common.IndexOfStr(publickey.GetMiningKeyBase58(consensusType), beaconPendingList) != -1 {
				engine.CurrentMiningChain = common.BeaconChainKey
				engine.updateUserState(&publickey, common.BeaconRole, common.PendingRole, 0)
				break
			}
			if common.IndexOfStr(publickey.GetMiningKeyBase58(consensusType), beaconWaitingList) != -1 {
				engine.CurrentMiningChain = common.BeaconChainKey
				engine.updateUserState(&publickey, common.BeaconRole, common.WaitingRole, 0)
				break
			}
			if common.IndexOfStr(publickey.GetMiningKeyBase58(consensusType), shardsWaitingList) != -1 {
				engine.CurrentMiningChain = common.BeaconChainKey
				engine.updateUserState(&publickey, common.ShardRole, common.WaitingRole, 0)
				break
			}
			for chainName, committee := range shardsPendingList {
				if common.IndexOfStr(publickey.GetMiningKeyBase58(consensusType), committee) != -1 {
					engine.CurrentMiningChain = chainName
					shardID := getShardFromChainName(chainName)
					if engine.config.Blockchain.BestState.Shard[shardID].GetShardHeight() > engine.config.Blockchain.BestState.Beacon.GetBestHeightOfShard(shardID) {
						role, shardID := engine.config.Blockchain.Chains[chainName].GetPubkeyRole(publickey.GetMiningKeyBase58(consensusType), 0)
						if role == common.ProposerRole || role == common.ValidatorRole {
							engine.updateUserState(&publickey, common.ShardRole, common.CommitteeRole, shardID)
						} else {
							if role == common.PendingRole {
								engine.updateUserState(&publickey, common.ShardRole, common.PendingRole, getShardFromChainName(chainName))
							}
						}
						break
					}
					engine.updateUserState(&publickey, common.ShardRole, common.PendingRole, getShardFromChainName(chainName))
				}
			}
			if engine.CurrentMiningChain != "" {
				break
			}
		}

	}

	for chainName, chain := range engine.config.Blockchain.Chains {
		if _, ok := AvailableConsensus[chain.GetConsensusType()]; ok {
			engine.ChainConsensusList[chainName] = AvailableConsensus[chain.GetConsensusType()].NewInstance(chain, chainName, engine.config.Node, Logger.log)
		}
	}

	if engine.CurrentMiningChain == common.BeaconChainKey {
		go engine.NotifyBeaconRole(true)
		go engine.NotifyShardRole(-1)
	}
	if engine.CurrentMiningChain != common.BeaconChainKey && engine.CurrentMiningChain != "" {
		go engine.NotifyBeaconRole(false)
		go engine.NotifyShardRole(int(getShardFromChainName(engine.CurrentMiningChain)))
	}

	for {
		select {
		case <-engine.cQuit:
		case chainName := <-engine.chainCommitteeChange:
			Logger.log.Critical("chain committee change", chainName)
			consensusType := engine.config.Blockchain.Chains[chainName].GetConsensusType()
			userCurrentPublicKey, ok := engine.userCurrentState.KeysBase58[consensusType]
			var userMiningKey incognitokey.CommitteePublicKey
			if !ok {
				userMiningKey, ok = engine.userMiningPublicKeys[consensusType]
				if !ok {
					continue
				}
				userCurrentPublicKey = userMiningKey.GetMiningKeyBase58(consensusType)
			} else {
				userMiningKey = engine.userMiningPublicKeys[consensusType]
			}

			if chainName == common.BeaconChainKey || engine.userCurrentState.UserRole == common.WaitingRole {
				allcommittee := engine.config.Blockchain.Chains[common.BeaconChainKey].(BeaconInterface).GetAllCommittees()
				isSkip := false
				if committees, ok := allcommittee[consensusType]; ok {
					for chainname, committee := range committees {
						keys, _ := incognitokey.ExtractPublickeysFromCommitteeKeyList(committee, consensusType)
						if common.IndexOfStr(userCurrentPublicKey, keys) != -1 {
							engine.CurrentMiningChain = chainname
							var userRole, userLayer string
							var shardID byte
							if chainname != common.BeaconChainKey {
								shardID = getShardFromChainName(chainname)
								userLayer = common.ShardRole
								//member still in shard committee on beacon beststate but not on shard beststate
								if engine.config.Blockchain.BestState.Shard[shardID].GetShardHeight() > engine.config.Blockchain.BestState.Beacon.GetBestHeightOfShard(shardID) {
									role, _ := engine.config.Blockchain.Chains[chainname].GetPubkeyRole(userCurrentPublicKey, 0)
									if role == common.EmptyString {
										isSkip = true
										engine.CurrentMiningChain = common.EmptyString
										engine.updateUserState(&userMiningKey, common.EmptyString, common.EmptyString, 0)
										break
									}
								}
							} else {
								isSkip = true
								userLayer = common.BeaconRole
							}

							userRole = common.CommitteeRole
							engine.updateUserState(&userMiningKey, userLayer, userRole, shardID)
							break
						} else {
							if chainname == engine.CurrentMiningChain && chainname != common.BeaconChainKey {
								shardID := getShardFromChainName(chainname)
								if engine.config.Blockchain.BestState.Shard[shardID].GetShardHeight() > engine.config.Blockchain.BestState.Beacon.GetBestHeightOfShard(shardID) {
									role, _ := engine.config.Blockchain.Chains[chainname].GetPubkeyRole(userCurrentPublicKey, 0)
									if role == common.ValidatorRole || role == common.ProposerRole {
										isSkip = true
										engine.updateUserState(&userMiningKey, common.ShardRole, common.CommitteeRole, shardID)
										break
									}
								} else {
									engine.CurrentMiningChain = common.EmptyString
									engine.updateUserState(&userMiningKey, common.EmptyString, common.EmptyString, 0)
									break
								}
							}
						}
					}
				}

				if isSkip {
					continue
				}

				if engine.CurrentMiningChain == common.EmptyString || engine.userCurrentState.UserRole == common.WaitingRole {
					shardsPendingLists := engine.config.Blockchain.Chains[common.BeaconChainKey].(BeaconInterface).GetShardsPendingList()
					beaconPendingList, _ := incognitokey.ExtractPublickeysFromCommitteeKeyList(engine.config.Blockchain.Chains[common.BeaconChainKey].(BeaconInterface).GetBeaconPendingList(), consensusType)
					shardsWaitingList, _ := incognitokey.ExtractPublickeysFromCommitteeKeyList(engine.config.Blockchain.Chains[common.BeaconChainKey].(BeaconInterface).GetShardsWaitingList(), consensusType)

					var shardsPendingList map[string][]string
					shardsPendingList = make(map[string][]string)

					for chainName, committee := range shardsPendingLists[consensusType] {
						shardsPendingList[chainName], _ = incognitokey.ExtractPublickeysFromCommitteeKeyList(committee, consensusType)
					}

					if common.IndexOfStr(userCurrentPublicKey, beaconPendingList) != -1 {
						engine.CurrentMiningChain = common.BeaconChainKey
						engine.updateUserState(&userMiningKey, common.BeaconRole, common.PendingRole, 0)
						continue
					}
					if common.IndexOfStr(userCurrentPublicKey, shardsWaitingList) != -1 {
						engine.CurrentMiningChain = common.BeaconChainKey
						engine.updateUserState(&userMiningKey, common.ShardRole, common.WaitingRole, 0)
						continue
					}
					for chainName, committee := range shardsPendingList {
						if common.IndexOfStr(userCurrentPublicKey, committee) != -1 {
							engine.CurrentMiningChain = chainName
							engine.updateUserState(&userMiningKey, common.ShardRole, common.PendingRole, getShardFromChainName(chainName))
							break
						}
					}
					if engine.CurrentMiningChain != "" {
						continue
					}
				}

				if engine.userCurrentState.UserLayer == common.BeaconRole {
					engine.CurrentMiningChain = common.EmptyString
					engine.updateUserState(&userMiningKey, common.EmptyString, common.EmptyString, 0)
				}
			} else {
				role, shardID := engine.config.Blockchain.Chains[chainName].GetPubkeyRole(userCurrentPublicKey, 0)
				if role != common.EmptyString {
					if role == common.ValidatorRole || role == common.ProposerRole {
						role = common.CommitteeRole
						engine.updateUserState(&userMiningKey, common.ShardRole, role, shardID)
					}
				} else {
					if engine.CurrentMiningChain == chainName {
						shardID := getShardFromChainName(chainName)
						if engine.config.Blockchain.BestState.Shard[shardID].GetShardHeight() > engine.config.Blockchain.BestState.Beacon.GetBestHeightOfShard(shardID) {
							engine.CurrentMiningChain = common.EmptyString
							engine.updateUserState(&userMiningKey, common.EmptyString, common.EmptyString, 0)
						}
					}
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
		chainStatus := map[string]bool{}
		for {
			select {
			case <-engine.cQuit:
				return
			default:
				time.Sleep(time.Millisecond * 1000)
				for chainName, consensus := range engine.ChainConsensusList {
					if chainName == engine.CurrentMiningChain && engine.userCurrentState.UserRole == common.CommitteeRole {
						if _, ok := chainStatus[chainName]; !ok {
							Logger.log.Critical("BFT: starting bft engine ", chainName)
						}
						consensus.Start()
						if _, ok := chainStatus[chainName]; !ok {
							Logger.log.Critical("BFT: started bft engine ", chainName)
							chainStatus[chainName] = true
						}
					} else {
						if _, ok := chainStatus[chainName]; ok {
							Logger.log.Critical("BFT: stopping bft engine ", chainName)
						}
						consensus.Stop()
						if _, ok := chainStatus[chainName]; ok {
							Logger.log.Critical("BFT: stopped bft engine ", chainName)
							delete(chainStatus, chainName)
						}
					}
				}
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

func (engine *Engine) GetUserLayer() (string, int) {
	if engine.CurrentMiningChain != "" {
		if engine.userCurrentState.UserLayer == common.BeaconChainKey {
			return engine.userCurrentState.UserLayer, -1
		}
		return engine.userCurrentState.UserLayer, int(engine.userCurrentState.ShardID)
	}
	return "", -2
}

func (engine *Engine) GetUserRole() (string, string, int) {
	//layer,role,shardID
	if engine.CurrentMiningChain != "" {
		userLayer := engine.userCurrentState.UserLayer
		userRole := engine.userCurrentState.UserRole
		shardID := -1
		if userRole == common.CommitteeRole || userRole == common.PendingRole {
			if userLayer == common.ShardRole {
				shardID = int(engine.userCurrentState.ShardID)
			}
		}
		return engine.userCurrentState.UserLayer, engine.userCurrentState.UserRole, shardID
	}
	return "", "", -2
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
	} else if config.Node.GetMiningKeys() != "" {
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

func (engine *Engine) updateConsensusState() {
	userLayer := ""
	if engine.CurrentMiningChain == common.BeaconChainKey {
		userLayer = common.BeaconRole
		go engine.NotifyBeaconRole(true)
		go engine.NotifyShardRole(-1)
	}
	if engine.CurrentMiningChain != common.BeaconChainKey && engine.CurrentMiningChain != "" {
		userLayer = common.ShardRole
		go engine.NotifyBeaconRole(false)
		go engine.NotifyShardRole(int(getShardFromChainName(engine.CurrentMiningChain)))
	}
	publicKey, err := engine.GetMiningPublicKeyByConsensus(engine.config.Blockchain.BestState.Beacon.ConsensusAlgorithm)
	if err != nil {
		Logger.log.Error(err)
		return
	}
	//ExtractMiningPublickeysFromCommitteeKeyList
	allcommittee := engine.config.Blockchain.Chains[common.BeaconChainKey].(BeaconInterface).GetAllCommittees()
	beaconCommittee := []string{}
	shardCommittee := map[byte][]string{}
	shardCommittee = make(map[byte][]string)

	for keyType, committees := range allcommittee {
		for chain, committee := range committees {
			if chain == common.BeaconChainKey {
				keyList, _ := incognitokey.ExtractMiningPublickeysFromCommitteeKeyList(committee, keyType)
				beaconCommittee = append(beaconCommittee, keyList...)
			} else {
				keyList, _ := incognitokey.ExtractMiningPublickeysFromCommitteeKeyList(committee, keyType)
				shardCommittee[getShardFromChainName(chain)] = keyList
			}
		}
	}
	if userLayer == common.ShardRole {
		committee := engine.config.Blockchain.Chains[engine.CurrentMiningChain].GetCommittee()
		keyList, _ := incognitokey.ExtractMiningPublickeysFromCommitteeKeyList(committee, engine.config.Blockchain.Chains[engine.CurrentMiningChain].GetConsensusType())
		shardCommittee[getShardFromChainName(engine.CurrentMiningChain)] = keyList
	}

	fmt.Printf("UpdateConsensusState %v %v\n", userLayer, publicKey)
	if userLayer == common.ShardRole {
		shardID := getShardFromChainName(engine.CurrentMiningChain)
		engine.config.Node.UpdateConsensusState(userLayer, publicKey, &shardID, beaconCommittee, shardCommittee)
	} else {
		engine.config.Node.UpdateConsensusState(userLayer, publicKey, nil, beaconCommittee, shardCommittee)
	}
}

func (engine *Engine) updateUserState(keySet *incognitokey.CommitteePublicKey, layer string, role string, shardID byte) {
	isChange := false

	if engine.userCurrentState.ShardID != shardID {
		isChange = true
	}
	if engine.userCurrentState.UserLayer != layer {
		isChange = true
	}
	if engine.userCurrentState.Keys != nil {
		incKey, ok := engine.userCurrentState.KeysBase58[common.IncKeyType]
		if ok {
			if incKey != keySet.GetIncKeyBase58() {
				isChange = true
			}
		}
	}

	if role == "" {
		engine.userCurrentState.UserLayer = ""
		engine.userCurrentState.UserRole = ""
		engine.userCurrentState.ShardID = 0
		engine.userCurrentState.Keys = nil
		engine.userCurrentState.KeysBase58 = make(map[string]string)
	} else {
		engine.userCurrentState.ShardID = shardID
		engine.userCurrentState.UserLayer = layer
		engine.userCurrentState.UserRole = role
		engine.userCurrentState.Keys = keySet
		engine.userCurrentState.KeysBase58 = make(map[string]string)
		engine.userCurrentState.KeysBase58[common.IncKeyType] = keySet.GetIncKeyBase58()
		for keyType := range keySet.MiningPubKey {
			engine.userCurrentState.KeysBase58[keyType] = keySet.GetMiningKeyBase58(keyType)
		}
	}

	engine.updateConsensusState()

	if isChange {
		engine.config.Node.DropAllConnections()
	}
}

func (engine *Engine) GetMiningPublicKeys() map[string]incognitokey.CommitteePublicKey {
	return engine.userMiningPublicKeys
}
