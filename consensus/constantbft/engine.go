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
		currentPBFTBlkHeight := uint64(0)
		currentPBFTRound := 1
		prevRoundNodeRole := ""
		for {
			select {
			case <-engine.cQuit:
				return
			default:
				if engine.config.BlockChain.IsReady(false, 0) {
					fmt.Println("Blockchain is ready. Run consensus...")
					if prevRoundNodeRole == common.BEACON_ROLE {
						if currentPBFTBlkHeight <= engine.config.BlockChain.BestState.Beacon.BeaconHeight {
							// reset round
							currentPBFTBlkHeight = engine.config.BlockChain.BestState.Beacon.BeaconHeight + 1
							currentPBFTRound = 1
						}
					}
					userRole, shardID := engine.config.BlockChain.BestState.Beacon.GetPubkeyRole(engine.config.UserKeySet.GetPublicKeyB58(), currentPBFTRound)
					nodeRole := common.EmptyString
					if (engine.config.NodeMode == common.NODEMODE_BEACON || engine.config.NodeMode == common.NODEMODE_AUTO) && userRole != common.SHARD_ROLE && userRole != common.EmptyString {
						nodeRole = common.BEACON_ROLE
					}
					if (engine.config.NodeMode == common.NODEMODE_SHARD || engine.config.NodeMode == common.NODEMODE_AUTO) && userRole == common.SHARD_ROLE {
						nodeRole = common.SHARD_ROLE
					}

					engine.config.Server.UpdateConsensusState(nodeRole, engine.config.UserKeySet.GetPublicKeyB58(), nil, engine.config.BlockChain.BestState.Beacon.BeaconCommittee, engine.config.BlockChain.BestState.Beacon.ShardCommittee)
					fmt.Printf("%v \n", engine.config.BlockChain.BestState.Beacon.BeaconCommittee)
					fmt.Printf("%v \n", engine.config.BlockChain.BestState.Beacon.ShardCommittee)
					if currentPBFTRound > 3 && prevRoundNodeRole != "" {
						// os.Exit(1)
					}
					if userRole != common.EmptyString {

						bftProtocol := &BFTProtocol{
							cQuit:             engine.cQuit,
							cBFTMsg:           engine.cBFTMsg,
							BlockGen:          engine.config.BlockGen,
							UserKeySet:        engine.config.UserKeySet,
							BlockChain:        engine.config.BlockChain,
							Server:            engine.config.Server,
							ShardToBeaconPool: engine.config.ShardToBeaconPool,
							CrossShardPool:    engine.config.CrossShardPool,
						}

						if (engine.config.NodeMode == common.NODEMODE_BEACON || engine.config.NodeMode == common.NODEMODE_AUTO) && userRole != common.SHARD_ROLE {
							fmt.Printf("Node mode %+v, user role %+v, shardID %+v \n currentPBFTRound %+v, beacon height %+v, currentPBFTBlkHeight %+v, prevRoundNodeRole %+v \n ", engine.config.NodeMode, userRole, shardID, currentPBFTRound, engine.config.BlockChain.BestState.Beacon.BeaconHeight, currentPBFTBlkHeight, prevRoundNodeRole)
							bftProtocol.RoundData.ProposerOffset = (currentPBFTRound - 1) % len(engine.config.BlockChain.BestState.Beacon.BeaconCommittee)
							bftProtocol.RoundData.BestStateHash = engine.config.BlockChain.BestState.Beacon.Hash()
							bftProtocol.RoundData.Layer = common.BEACON_ROLE
							bftProtocol.RoundData.Committee = make([]string, len(engine.config.BlockChain.BestState.Beacon.BeaconCommittee))
							copy(bftProtocol.RoundData.Committee, engine.config.BlockChain.BestState.Beacon.BeaconCommittee)
							roundRole, _ := engine.config.BlockChain.BestState.Beacon.GetPubkeyRole(engine.config.UserKeySet.GetPublicKeyB58(), bftProtocol.RoundData.ProposerOffset)
							var (
								err    error
								resBlk interface{}
							)
							switch roundRole {
							case common.PROPOSER_ROLE:
								bftProtocol.RoundData.IsProposer = true
								currentPBFTBlkHeight = engine.config.BlockChain.BestState.Beacon.BeaconHeight + 1
								resBlk, err = bftProtocol.Start()
								if err != nil {
									currentPBFTRound++
									prevRoundNodeRole = nodeRole
								}
							case common.VALIDATOR_ROLE:
								bftProtocol.RoundData.IsProposer = false
								currentPBFTBlkHeight = engine.config.BlockChain.BestState.Beacon.BeaconHeight + 1
								resBlk, err = bftProtocol.Start()
								if err != nil {
									currentPBFTRound++
									prevRoundNodeRole = nodeRole
								}
							default:
								err = errors.New("Not your turn yet")
							}

							if err == nil {
								fmt.Println(resBlk.(*blockchain.BeaconBlock))
								err = engine.config.BlockChain.InsertBeaconBlock(resBlk.(*blockchain.BeaconBlock), false)
								if err != nil {
									Logger.log.Error("Insert beacon block error", err)
									continue
								}
								//PUSH BEACON TO ALL
								newBeaconBlock := resBlk.(*blockchain.BeaconBlock)
								newBeaconBlockMsg, err := MakeMsgBeaconBlock(newBeaconBlock)
								if err != nil {
									Logger.log.Error("Make new beacon block message error", err)
								} else {
									engine.config.Server.PushMessageToAll(newBeaconBlockMsg)
								}
								//reset round
								prevRoundNodeRole = ""
								currentPBFTRound = 1
							} else {
								Logger.log.Error(err)
							}
							continue
						}
						if (engine.config.NodeMode == common.NODEMODE_SHARD || engine.config.NodeMode == common.NODEMODE_AUTO) && userRole == common.SHARD_ROLE {
							if currentPBFTBlkHeight <= engine.config.BlockChain.BestState.Shard[shardID].ShardHeight {
								// reset
								currentPBFTBlkHeight = engine.config.BlockChain.BestState.Shard[shardID].ShardHeight + 1
								currentPBFTRound = 1
							}
							fmt.Printf("Node mode %+v, user role %+v, shardID %+v \n currentPBFTRound %+v, beacon height %+v, currentPBFTBlkHeight %+v, prevRoundNodeRole %+v \n ", engine.config.NodeMode, userRole, shardID, currentPBFTRound, engine.config.BlockChain.BestState.Shard[shardID].ShardCommittee, currentPBFTBlkHeight, prevRoundNodeRole)
							engine.config.BlockChain.SyncShard(shardID)
							engine.config.BlockChain.StopSyncUnnecessaryShard()
							bftProtocol.RoundData.ProposerOffset = (currentPBFTRound - 1) % len(engine.config.BlockChain.BestState.Shard[shardID].ShardCommittee)
							bftProtocol.RoundData.BestStateHash = engine.config.BlockChain.BestState.Shard[shardID].Hash()
							bftProtocol.RoundData.Layer = common.SHARD_ROLE
							bftProtocol.RoundData.ShardID = shardID
							bftProtocol.RoundData.Committee = make([]string, len(engine.config.BlockChain.BestState.Shard[shardID].ShardCommittee))
							copy(bftProtocol.RoundData.Committee, engine.config.BlockChain.BestState.Shard[shardID].ShardCommittee)
							var (
								err    error
								resBlk interface{}
							)
							if engine.config.BlockChain.IsReady(true, shardID) {
								roundRole := engine.config.BlockChain.BestState.Shard[shardID].GetPubkeyRole(engine.config.UserKeySet.GetPublicKeyB58(), bftProtocol.RoundData.ProposerOffset)
								fmt.Println("My shard role", roundRole)
								switch roundRole {
								case common.PROPOSER_ROLE:
									bftProtocol.RoundData.IsProposer = true
									currentPBFTBlkHeight = engine.config.BlockChain.BestState.Shard[shardID].ShardHeight + 1
									resBlk, err = bftProtocol.Start()
									if err != nil {
										currentPBFTRound++
										prevRoundNodeRole = nodeRole
									}
								case common.VALIDATOR_ROLE:
									bftProtocol.RoundData.IsProposer = false
									currentPBFTBlkHeight = engine.config.BlockChain.BestState.Shard[shardID].ShardHeight + 1

									resBlk, err = bftProtocol.Start()
									if err != nil {
										currentPBFTRound++
										prevRoundNodeRole = nodeRole
									}
								default:
									err = errors.New("Not your turn yet")
								}
								if err == nil {
									fmt.Println("========NEW SHARD BLOCK=======", resBlk.(*blockchain.ShardBlock).Header.Height)
									err = engine.config.BlockChain.InsertShardBlock(resBlk.(*blockchain.ShardBlock))
									if err != nil {
										Logger.log.Error("Insert shard block error", err)
										continue
									}
									fmt.Println(">>>>>>>>>>>>>+++++++++++++++<<<<<<<< 1")
									//PUSH SHARD TO BEACON
									shardBlk, ok := resBlk.(*blockchain.ShardBlock)
									if !ok {
										Logger.log.Debug("Got data of type %T but wanted blockchain.ShardBlock", resBlk)
										continue
									}
									fmt.Println(">>>>>>>>>>>>>+++++++++++++++<<<<<<<< 2")
									newShardToBeaconBlock := shardBlk.CreateShardToBeaconBlock(engine.config.BlockChain)
									fmt.Println(">>>>>>>>>>>>>+++++++++++++++<<<<<<<< 3")
									newShardToBeaconMsg, err := MakeMsgShardToBeaconBlock(newShardToBeaconBlock)
									//TODO: check lock later
									if err == nil {
										go engine.config.Server.PushMessageToBeacon(newShardToBeaconMsg)
									}
									fmt.Println(">>>>>>>>>>>>>+++++++++++++++<<<<<<<< 4")
									//PUSH CROSS-SHARD
									newCrossShardBlocks := shardBlk.CreateAllCrossShardBlock(engine.config.BlockChain.BestState.Beacon.ActiveShards)
									fmt.Println(newCrossShardBlocks)
									fmt.Println(">>>>>>>>>>>>>+++++++++++++++<<<<<<<< 5")
									for sID, newCrossShardBlock := range newCrossShardBlocks {
										newCrossShardMsg, err := MakeMsgCrossShardBlock(newCrossShardBlock)
										if err == nil {
											engine.config.Server.PushMessageToShard(newCrossShardMsg, sID)
										}
									}
									fmt.Println(">>>>>>>>>>>>>+++++++++++++++<<<<<<<< 6")
									if err != nil {
										Logger.log.Error("Make new block message error", err)
									}
									fmt.Println(">>>>>>>>>>>>>+++++++++++++++<<<<<<<< 7")

									//reset round
									prevRoundNodeRole = ""
									currentPBFTRound = 1
								} else {
									Logger.log.Error(err)
								}
							} else {
								//reset round
								prevRoundNodeRole = ""
								currentPBFTRound = 1
								// Logger.log.Error("Blockchain is not ready!")
							}
						}
					}
				} else {
					//reset round
					prevRoundNodeRole = ""
					currentPBFTRound = 1
					// Logger.log.Error("Blockchain is not ready!")
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
