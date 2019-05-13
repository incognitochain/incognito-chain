package constantbft

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/constant-money/constant-chain/blockchain"
	"github.com/constant-money/constant-chain/cashec"
	"github.com/constant-money/constant-chain/common"
	"github.com/constant-money/constant-chain/wire"
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
	prevRoundUserLayer  string
	userLayer           string
}

type EngineConfig struct {
	BlockChain        *blockchain.BlockChain
	ChainParams       *blockchain.Params
	BlockGen          *blockchain.BlkTmplGenerator
	UserKeySet        *cashec.KeySet
	NodeMode          string
	Server            serverInterface
	ShardToBeaconPool blockchain.ShardToBeaconPool
	CrossShardPool    map[byte]blockchain.CrossShardPool
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
	engine.cBFTMsg = make(chan wire.Message)
	engine.started = true
	Logger.log.Info("Start consensus with key", engine.config.UserKeySet.GetPublicKeyB58())
	fmt.Println(engine.config.BlockChain.BestState.Beacon.BeaconCommittee)

	time.AfterFunc(DelayTime*time.Millisecond, func() {
		engine.currentBFTRound = 1
		for {
			select {
			case <-engine.cQuit:
				return
			default:
				if !engine.config.BlockChain.IsReady(false, 0) {
					time.Sleep(time.Millisecond * 100)
				} else {
					userRole, shardID := engine.config.BlockChain.BestState.Beacon.GetPubkeyRole(engine.config.UserKeySet.GetPublicKeyB58(), engine.currentBFTRound)
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
					engine.config.Server.UpdateConsensusState(engine.userLayer, engine.config.UserKeySet.GetPublicKeyB58(), nil, engine.config.BlockChain.BestState.Beacon.BeaconCommittee, engine.config.BlockChain.BestState.Beacon.ShardCommittee)
					switch engine.userLayer {
					case common.BEACON_ROLE:
						if engine.config.NodeMode == common.NODEMODE_BEACON || engine.config.NodeMode == common.NODEMODE_AUTO {
							engine.config.BlockChain.ConsensusOngoing = true
							engine.execBeaconRole()
							engine.config.BlockChain.ConsensusOngoing = false
						}
					case common.SHARD_ROLE:
						if engine.config.NodeMode == common.NODEMODE_SHARD || engine.config.NodeMode == common.NODEMODE_AUTO {
							if !engine.config.BlockChain.IsReady(true, shardID) {
								time.Sleep(time.Millisecond * 100)
							} else {
								engine.config.BlockChain.ConsensusOngoing = true
								engine.execShardRole(shardID)
								engine.config.BlockChain.ConsensusOngoing = false
							}
						}
					case common.EmptyString:
						time.Sleep(time.Second * 1)
					}
				}
			}
		}
	})
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
	roundRole, _ := engine.config.BlockChain.BestState.Beacon.GetPubkeyRole(engine.config.UserKeySet.GetPublicKeyB58(), bftProtocol.RoundData.Round)
	var (
		err    error
		resBlk interface{}
	)
	switch roundRole {
	case common.PROPOSER_ROLE:
		bftProtocol.RoundData.IsProposer = true
		engine.currentBFTBlkHeight = engine.config.BlockChain.BestState.Beacon.BeaconHeight + 1
		//fmt.Println("[db] bftProtocol.Start() beacon proposer_role")
		resBlk, err = bftProtocol.Start()
		if err != nil {
			engine.currentBFTRound++
			engine.prevRoundUserLayer = engine.userLayer
		}
	case common.VALIDATOR_ROLE:
		bftProtocol.RoundData.IsProposer = false
		engine.currentBFTBlkHeight = engine.config.BlockChain.BestState.Beacon.BeaconHeight + 1
		//fmt.Println("[db] bftProtocol.Start() beacon validator_role")
		resBlk, err = bftProtocol.Start()
		if err != nil {
			engine.currentBFTRound++
			engine.prevRoundUserLayer = engine.userLayer
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
	} else {
		Logger.log.Error(err)
	}
	return
}

func (engine *Engine) execShardRole(shardID byte) {
	if engine.currentBFTBlkHeight <= engine.config.BlockChain.BestState.Shard[shardID].ShardHeight {
		// reset
		engine.currentBFTBlkHeight = engine.config.BlockChain.BestState.Shard[shardID].ShardHeight + 1
		engine.currentBFTRound = 1
	}
	engine.config.BlockChain.SyncShard(shardID)
	bftProtocol := &BFTProtocol{
		cBFTMsg:   engine.cBFTMsg,
		EngineCfg: &engine.config,
	}
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
	roundRole := engine.config.BlockChain.BestState.Shard[shardID].GetPubkeyRole(engine.config.UserKeySet.GetPublicKeyB58(), bftProtocol.RoundData.Round)
	fmt.Println("My shard role", roundRole)
	switch roundRole {
	case common.PROPOSER_ROLE:
		bftProtocol.RoundData.IsProposer = true
		engine.currentBFTBlkHeight = engine.config.BlockChain.BestState.Shard[shardID].ShardHeight + 1
		resBlk, err = bftProtocol.Start()
		if err != nil {
			engine.currentBFTRound++
			engine.prevRoundUserLayer = engine.userLayer
		}
	case common.VALIDATOR_ROLE:
		bftProtocol.RoundData.IsProposer = false
		engine.currentBFTBlkHeight = engine.config.BlockChain.BestState.Shard[shardID].ShardHeight + 1
		resBlk, err = bftProtocol.Start()
		if err != nil {
			engine.currentBFTRound++
			engine.prevRoundUserLayer = engine.userLayer
		}
	default:
		err = errors.New("Not your turn yet")
		time.Sleep(time.Millisecond * 300)
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
	} else {
		Logger.log.Error(err)
	}
}
