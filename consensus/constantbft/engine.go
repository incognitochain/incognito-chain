package constantbft

import (
	"errors"
	"fmt"
	"github.com/incognitochain/incognito-chain/pubsub"
	"sync"
	"time"

	"github.com/incognitochain/incognito-chain/blockchain"
	"github.com/incognitochain/incognito-chain/cashec"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/wire"
)

type Engine struct {
	sync.Mutex
	started bool
	// channel
	cQuit   chan struct{}
	cBFTMsg chan wire.Message
	config EngineConfig
	currentBFTBlkHeight uint64
	currentBFTRound     int
	// prevRoundUserLayer  string
	userLayer string
	retries   int
	userPk    string
}

type EngineConfig struct {
	BlockChain                  *blockchain.BlockChain
	ChainParams                 *blockchain.Params
	BlockGen                    *blockchain.BlkTmplGenerator
	UserKeySet                  *cashec.KeySet
	NodeMode                    string
	Server                      serverInterface
	ShardToBeaconPool           blockchain.ShardToBeaconPool
	CrossShardPool              map[byte]blockchain.CrossShardPool
	PubsubManager               *pubsub.PubsubManager
	CRoleInCommitteesMempool    chan int
	CRoleInCommitteesNetSync    chan int
	CRoleInCommitteesBeaconPool chan bool
	CRoleInCommitteesShardPool  []chan int
}

//Init apply configuration to consensus engine
func (engine Engine) Init(cfg *EngineConfig) (*Engine, error) {
	return &Engine{
		config: *cfg,
	}, nil
}

func (engine *Engine) Start() error {
	engine.Lock()
	defer engine.Unlock()
	if engine.started {
		return errors.New("Consensus engine is already started")
	}
	engine.cQuit = make(chan struct{})
	//Start block generator
	go engine.config.BlockGen.Start(engine.cQuit)
	engine.cBFTMsg = make(chan wire.Message)
	engine.started = true
	engine.userPk = engine.config.UserKeySet.GetPublicKeyB58()
	Logger.log.Info("Start consensus with key", engine.userPk)
	fmt.Println(engine.config.BlockChain.BestState.Beacon.BeaconCommittee)

	go func() {
		engine.currentBFTRound = 1
		for {
			select {
			case <-engine.cQuit:
				return
			default:
				if !engine.config.BlockChain.Synker.IsLatest(false, 0) {
					userRole, shardID := engine.config.BlockChain.BestState.Beacon.GetPubkeyRole(engine.userPk, 0)
					if userRole == common.SHARD_ROLE {
						go engine.NotifyShardRole(int(shardID))
						go engine.NotifyBeaconRole(false)
					} else {
						if userRole == common.PROPOSER_ROLE || userRole == common.VALIDATOR_ROLE {
							go engine.NotifyBeaconRole(true)
							go engine.NotifyShardRole(-1)
						}
					}
					time.Sleep(time.Millisecond * 100)
				} else {
					userRole, shardID := engine.config.BlockChain.BestState.Beacon.GetPubkeyRole(engine.userPk, engine.currentBFTRound)
					if engine.config.NodeMode == common.NODEMODE_BEACON && userRole == common.SHARD_ROLE {
						userRole = common.EmptyString
					}
					if engine.config.NodeMode == common.NODEMODE_SHARD && userRole != common.SHARD_ROLE {
						userRole = common.EmptyString
					}
					engine.userLayer = userRole
					switch userRole {
					case common.VALIDATOR_ROLE, common.PROPOSER_ROLE:
						engine.userLayer = common.BEACON_ROLE
					}
					engine.config.Server.UpdateConsensusState(engine.userLayer, engine.userPk, nil, engine.config.BlockChain.BestState.Beacon.BeaconCommittee, engine.config.BlockChain.BestState.Beacon.GetShardCommittee())
					switch engine.userLayer {
					case common.BEACON_ROLE:
						if engine.config.NodeMode == common.NODEMODE_BEACON || engine.config.NodeMode == common.NODEMODE_AUTO {
							engine.config.BlockChain.ConsensusOngoing = true
							engine.execBeaconRole()
							engine.config.BlockChain.ConsensusOngoing = false
						}
					case common.SHARD_ROLE:
						if engine.config.NodeMode == common.NODEMODE_SHARD || engine.config.NodeMode == common.NODEMODE_AUTO {
							if !engine.config.BlockChain.Synker.IsLatest(true, shardID) {
								time.Sleep(time.Millisecond * 100)
							} else {
								engine.config.BlockChain.ConsensusOngoing = true
								engine.execShardRole(shardID)
								fmt.Println("BFT: exit")
								engine.config.BlockChain.ConsensusOngoing = false
							}
						}
					case common.EmptyString:
						time.Sleep(time.Second * 1)
					}
				}
			}
		}
	}()
	return nil
}

func (engine *Engine) Stop() error {
	engine.Lock()
	defer engine.Unlock()
	if !engine.started {
		return errors.New("Consensus engine is already stopped")
	}
	engine.started = false
	close(engine.cQuit)
	return nil
}

func (engine *Engine) execBeaconRole() {
	if engine.currentBFTBlkHeight <= engine.config.BlockChain.BestState.Beacon.BeaconHeight {
		// reset round
		engine.currentBFTBlkHeight = engine.config.BlockChain.BestState.Beacon.BeaconHeight + 1
		engine.currentBFTRound = 1
		engine.retries = 0
	}
	if engine.retries >= MaxNormalRetryTime {
		timeSinceLastBlk := time.Since(time.Unix(engine.config.BlockChain.BestState.Beacon.BestBlock.Header.Timestamp, 0))
		engine.currentBFTRound = int(timeSinceLastBlk.Round(common.MinBeaconBlkInterval).Seconds()) / int(common.MinBeaconBlkInterval.Seconds())
	}

	bftProtocol := &BFTProtocol{
		cBFTMsg:   engine.cBFTMsg,
		EngineCfg: &engine.config,
	}
	bftProtocol.RoundData.Round = engine.currentBFTRound
	bftProtocol.RoundData.BestStateHash = engine.config.BlockChain.BestState.Beacon.Hash()
	bftProtocol.RoundData.Layer = common.BEACON_ROLE
	bftProtocol.RoundData.Committee = make([]string, len(engine.config.BlockChain.BestState.Beacon.BeaconCommittee))
	copy(bftProtocol.RoundData.Committee, engine.config.BlockChain.BestState.Beacon.BeaconCommittee)
	roundRole, _ := engine.config.BlockChain.BestState.Beacon.GetPubkeyRole(engine.userPk, bftProtocol.RoundData.Round)
	var (
		err    error
		resBlk interface{}
	)
	go engine.NotifyBeaconRole(true)
	go engine.NotifyShardRole(-1)
	switch roundRole {
	case common.PROPOSER_ROLE:
		bftProtocol.RoundData.IsProposer = true
		engine.currentBFTBlkHeight = engine.config.BlockChain.BestState.Beacon.BeaconHeight + 1
		//fmt.Println("[db] bftProtocol.Start() beacon proposer_role")
		resBlk, err = bftProtocol.Start()
		if err != nil {
			engine.currentBFTRound++
			engine.retries++
			// engine.prevRoundUserLayer = engine.userLayer
		}
	case common.VALIDATOR_ROLE:
		bftProtocol.RoundData.IsProposer = false
		engine.currentBFTBlkHeight = engine.config.BlockChain.BestState.Beacon.BeaconHeight + 1
		//fmt.Println("[db] bftProtocol.Start() beacon validator_role")
		resBlk, err = bftProtocol.Start()
		if err != nil {
			engine.currentBFTRound++
			engine.retries++
			// engine.prevRoundUserLayer = engine.userLayer
		}
	default:
		err = errors.New("Not your turn yet")
	}

	if err == nil {
		fmt.Println(resBlk.(*blockchain.BeaconBlock))
		err = engine.config.BlockChain.InsertBeaconBlock(resBlk.(*blockchain.BeaconBlock), true)
		if err != nil {
			Logger.log.Error("Insert beacon block error", err)
			return
		}
		//PUSH BEACON TO ALL
		newBeaconBlock := resBlk.(*blockchain.BeaconBlock)
		newBeaconBlockMsg, err := MakeMsgBeaconBlock(newBeaconBlock)
		if err != nil {
			Logger.log.Error("Make new beacon block message error", err)
		} else {
			engine.config.Server.PushMessageToAll(newBeaconBlockMsg)
		}
		engine.retries = 0
	} else {
		Logger.log.Error(err)
	}
}

func (engine *Engine) execShardRole(shardID byte) {
	if engine.currentBFTBlkHeight <= engine.config.BlockChain.BestState.Shard[shardID].ShardHeight {
		// reset
		engine.currentBFTBlkHeight = engine.config.BlockChain.BestState.Shard[shardID].ShardHeight + 1
		engine.currentBFTRound = 1
		engine.retries = 0
	}
	if engine.retries >= MaxNormalRetryTime {
		timeSinceLastBlk := time.Since(time.Unix(engine.config.BlockChain.BestState.Shard[shardID].BestBlock.Header.Timestamp, 0))
		engine.currentBFTRound = int(timeSinceLastBlk.Round(common.MinShardBlkInterval).Seconds()) / int(common.MinShardBlkInterval.Seconds())
	}
	engine.config.BlockChain.Synker.SyncShard(shardID)
	bftProtocol := &BFTProtocol{
		cBFTMsg:   engine.cBFTMsg,
		EngineCfg: &engine.config,
	}
	bftProtocol.RoundData.MinBeaconHeight = engine.config.BlockChain.BestState.Beacon.BeaconHeight
	bftProtocol.RoundData.Round = engine.currentBFTRound
	bftProtocol.RoundData.BestStateHash = engine.config.BlockChain.BestState.Shard[shardID].Hash()
	bftProtocol.RoundData.Layer = common.SHARD_ROLE
	bftProtocol.RoundData.ShardID = shardID
	bftProtocol.RoundData.Committee = make([]string, len(engine.config.BlockChain.BestState.Shard[shardID].ShardCommittee))
	copy(bftProtocol.RoundData.Committee, engine.config.BlockChain.BestState.Shard[shardID].ShardCommittee)
	var (
		err    error
		resBlk interface{}
	)
	roundRole := engine.config.BlockChain.BestState.Shard[shardID].GetPubkeyRole(engine.userPk, bftProtocol.RoundData.Round)
	Logger.log.Infof("My shard role %+v, ShardID %+v \n", roundRole, shardID)
	go engine.NotifyBeaconRole(false)
	go engine.NotifyShardRole(int(shardID))
	switch roundRole {
	case common.PROPOSER_ROLE:
		bftProtocol.RoundData.IsProposer = true
		engine.currentBFTBlkHeight = engine.config.BlockChain.BestState.Shard[shardID].ShardHeight + 1
		resBlk, err = bftProtocol.Start()
		if err != nil {
			engine.currentBFTRound++
			engine.retries++
			// engine.prevRoundUserLayer = engine.userLayer
		}
	case common.VALIDATOR_ROLE:
		bftProtocol.RoundData.IsProposer = false
		engine.currentBFTBlkHeight = engine.config.BlockChain.BestState.Shard[shardID].ShardHeight + 1
		resBlk, err = bftProtocol.Start()
		if err != nil {
			engine.currentBFTRound++
			engine.retries++
			// engine.prevRoundUserLayer = engine.userLayer
		}
	default:
		err = errors.New("Not your turn yet")
	}

	if err == nil {
		shardBlk := resBlk.(*blockchain.ShardBlock)
		Logger.log.Critical("===============NEW SHARD BLOCK==============")
		Logger.log.Critical("Shard Block Height", shardBlk.Header.Height)

		err = engine.config.BlockChain.InsertShardBlock(shardBlk, true)
		if err != nil {
			Logger.log.Error("Insert shard block error", err)
			return
		}
		go func() {
			//PUSH SHARD TO BEACON
			//fmt.Println("Create And Push Shard To Beacon Block")
			newShardToBeaconBlock := shardBlk.CreateShardToBeaconBlock(engine.config.BlockChain)
			newShardToBeaconMsg, err := MakeMsgShardToBeaconBlock(newShardToBeaconBlock)
			if err == nil {
				go engine.config.Server.PushMessageToBeacon(newShardToBeaconMsg)
			}
			//fmt.Println("Create and Push all Cross Shard Block")
			//PUSH CROSS-SHARD
			newCrossShardBlocks := shardBlk.CreateAllCrossShardBlock(engine.config.BlockChain.BestState.Beacon.ActiveShards)
			//fmt.Println("New Cross Shard Blocks ", newCrossShardBlocks, shardBlk.Header.Height, shardBlk.Header.CrossShards)

			for sID, newCrossShardBlock := range newCrossShardBlocks {
				newCrossShardMsg, err := MakeMsgCrossShardBlock(newCrossShardBlock)
				if err == nil {
					engine.config.Server.PushMessageToShard(newCrossShardMsg, sID)
				}
			}
		}()
		//PUSH TO ALL
		newShardBlock := resBlk.(*blockchain.ShardBlock)
		newShardBlockMsg, err := MakeMsgShardBlock(newShardBlock)
		if err != nil {
			Logger.log.Error("Make new shard block message error", err)
		} else {
			engine.config.Server.PushMessageToAll(newShardBlockMsg)
		}
		engine.retries = 0
	} else {
		Logger.log.Error(err)
	}
}

func (engine *Engine) NotifyBeaconRole(beaconRole bool) {
	engine.config.PubsubManager.PublishMessage(pubsub.NewMessage(pubsub.BeaconRoleTopic, beaconRole))
}
func (engine *Engine) NotifyShardRole(shardRole int) {
	engine.config.PubsubManager.PublishMessage(pubsub.NewMessage(pubsub.ShardRoleTopic, shardRole))
}
