package constantpos

import (
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/ninjadotorg/constant/blockchain"
	"github.com/ninjadotorg/constant/cashec"
	"github.com/ninjadotorg/constant/wire"
)

type Engine struct {
	sync.Mutex
	started bool

	// channel
	cQuit   chan struct{}
	cBFTMsg chan wire.Message

	config      EngineConfig
	CurrentRole role
}

type role struct {
	Role    string
	ShardID byte
}

type EngineConfig struct {
	BlockChain  *blockchain.BlockChain
	ChainParams *blockchain.Params
	BlockGen    *blockchain.BlkTmplGenerator
	UserKeySet  *cashec.KeySet
	NodeMode    string
	Server      serverInterface
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

	//Note: why goroutine in this function
	go func() {
		for {
			select {
			case <-engine.cQuit:
				return
			default:
				if engine.config.BlockChain.IsReady(false, 0) {
					role, shardID := engine.config.BlockChain.BestState.Beacon.GetPubkeyRole(engine.config.UserKeySet.GetPublicKeyB58())
					nodeRole := ""
					if (engine.config.NodeMode == "beacon" || engine.config.NodeMode == "auto") && role != "shard" {
						nodeRole = "beacon"
					}
					if (engine.config.NodeMode == "shard" || engine.config.NodeMode == "auto") && role == "shard" {
						nodeRole = "shard"
					}
					go engine.config.Server.UpdateConsensusState(nodeRole, engine.config.UserKeySet.GetPublicKeyB58(), nil, engine.config.BlockChain.BestState.Beacon.BeaconCommittee, engine.config.BlockChain.BestState.Beacon.ShardCommittee)
					time.Sleep(2 * time.Second)

					fmt.Println(engine.config.NodeMode, role, shardID)
					if role != "" {
						bftProtocol := &BFTProtocol{
							cQuit:      engine.cQuit,
							cBFTMsg:    engine.cBFTMsg,
							BlockGen:   engine.config.BlockGen,
							UserKeySet: engine.config.UserKeySet,
							Chain:      engine.config.BlockChain,
							Server:     engine.config.Server,
						}
						if (engine.config.NodeMode == "beacon" || engine.config.NodeMode == "auto") && role != "shard" {
							bftProtocol.RoleData.Committee = make([]string, len(engine.config.BlockChain.BestState.Beacon.BeaconCommittee))
							copy(bftProtocol.RoleData.Committee, engine.config.BlockChain.BestState.Beacon.BeaconCommittee)
							var (
								err    error
								resBlk interface{}
							)
							switch role {
							case "beacon-proposer":
								resBlk, err = bftProtocol.Start(true, "beacon", 0)
								if err != nil {
									Logger.log.Error("PBFT fatal error", err)
									continue
								}
							case "beacon-validator":
								msgReady, _ := MakeMsgBFTReady(engine.config.BlockChain.BestState.Beacon.Hash())
								engine.config.Server.PushMessageToBeacon(msgReady)
								resBlk, err = bftProtocol.Start(false, "beacon", 0)
								if err != nil {
									Logger.log.Error("PBFT fatal error", err)
									continue
								}
							default:
								err = errors.New("Not your turn yet")
							}

							if err == nil {
								fmt.Println(resBlk.(*blockchain.BeaconBlock))
								err = engine.config.BlockChain.InsertBeaconBlock(resBlk.(*blockchain.BeaconBlock))
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

							} else {
								Logger.log.Error(err)
							}
							continue
						}
						if (engine.config.NodeMode == "shard" || engine.config.NodeMode == "auto") && role == "shard" {
							engine.config.BlockChain.SyncShard(shardID)
							bftProtocol.RoleData.Committee = make([]string, len(engine.config.BlockChain.BestState.Shard[shardID].ShardCommittee))
							copy(bftProtocol.RoleData.Committee, engine.config.BlockChain.BestState.Shard[shardID].ShardCommittee)
							var (
								err    error
								resBlk interface{}
							)
							if engine.config.BlockChain.IsReady(true, shardID) {
								shardRole := engine.config.BlockChain.BestState.Shard[shardID].GetPubkeyRole(engine.config.UserKeySet.GetPublicKeyB58())
								fmt.Println("My shard role", shardRole)
								switch shardRole {
								case "shard-proposer":
									resBlk, err = bftProtocol.Start(true, "shard", shardID)
									if err != nil {
										Logger.log.Error("PBFT fatal error", err)
										continue
									}
								case "shard-validator":
									msgReady, _ := MakeMsgBFTReady(engine.config.BlockChain.BestState.Shard[shardID].Hash())
									engine.config.Server.PushMessageToShard(msgReady, shardID)
									resBlk, err = bftProtocol.Start(false, "shard", shardID)
									if err != nil {
										Logger.log.Error("PBFT fatal error", err)
										continue
									}
								default:
									err = errors.New("Not your turn yet")
								}
								if err == nil {
									fmt.Println(resBlk.(*blockchain.ShardBlock))
									err = engine.config.BlockChain.InsertShardBlock(resBlk.(*blockchain.ShardBlock))
									if err != nil {
										Logger.log.Error("Insert beacon block error", err)
										continue
									}
									//PUSH SHARD TO BEACON
									shardBlk, ok := resBlk.(*blockchain.ShardBlock)
									if !ok {
										log.Printf("Got data of type %T but wanted blockchain.ShardBlock", resBlk)
										continue
									}

									newShardToBeaconBlock := shardBlk.CreateShardToBeaconBlock()
									newShardToBeaconMsg, err := MakeMsgShardToBeaconBlock(&newShardToBeaconBlock)
									if err == nil {
										engine.config.Server.PushMessageToBeacon(newShardToBeaconMsg)
									}
									//PUSH CROSS-SHARD
									newCrossShardBlocks := shardBlk.CreateAllCrossShardBlock()
									for sID, newCrossShardBlock := range newCrossShardBlocks {
										newCrossShardMsg, err := MakeMsgCrossShardBlock(newCrossShardBlock)
										if err == nil {
											engine.config.Server.PushMessageToShard(newCrossShardMsg, sID)
										}
									}

									if err != nil {
										Logger.log.Error("Make new block message error", err)
									}

								} else {
									Logger.log.Error(err)
								}
							} else {
								Logger.log.Warn("Blockchain is not ready!")
							}
						}
					} else {
						time.Sleep(5 * time.Second)
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
