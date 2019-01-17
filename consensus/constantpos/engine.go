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
func (self Engine) Init(cfg *EngineConfig) (*Engine, error) {
	return &Engine{
		config: *cfg,
	}, nil
}

func (self *Engine) Start() error {
	self.Lock()
	defer self.Unlock()
	if self.started {
		return errors.New("Consensus engine is already started")
	}
	self.cQuit = make(chan struct{})
	self.cBFTMsg = make(chan wire.Message)
	self.started = true
	Logger.log.Info("Start consensus with key", self.config.UserKeySet.GetPublicKeyB58())
	fmt.Println(self.config.BlockChain.BestState.Beacon.BeaconCommittee)

	//Note: why goroutine in this function
	go func() {
		for {
			select {
			case <-self.cQuit:
				return
			default:
				if self.config.BlockChain.IsReady(false, 0) {
					role, shardID := self.config.BlockChain.BestState.Beacon.GetPubkeyRole(self.config.UserKeySet.GetPublicKeyB58())
					nodeRole := ""
					if (self.config.NodeMode == "beacon" || self.config.NodeMode == "auto") && role != "shard" {
						nodeRole = "beacon"
					}
					if (self.config.NodeMode == "shard" || self.config.NodeMode == "auto") && role == "shard" {
						nodeRole = "shard"
					}
					go self.config.Server.UpdateConsensusState(nodeRole, self.config.UserKeySet.GetPublicKeyB58(), nil, self.config.BlockChain.BestState.Beacon.BeaconCommittee, self.config.BlockChain.BestState.Beacon.ShardCommittee)
					time.Sleep(2 * time.Second)

					fmt.Println(self.config.NodeMode, role, shardID)
					if role != "" {
						bftProtocol := &BFTProtocol{
							cQuit:      self.cQuit,
							cBFTMsg:    self.cBFTMsg,
							BlockGen:   self.config.BlockGen,
							UserKeySet: self.config.UserKeySet,
							Chain:      self.config.BlockChain,
							Server:     self.config.Server,
						}
						if (self.config.NodeMode == "beacon" || self.config.NodeMode == "auto") && role != "shard" {
							bftProtocol.RoleData.Committee = make([]string, len(self.config.BlockChain.BestState.Beacon.BeaconCommittee))
							copy(bftProtocol.RoleData.Committee, self.config.BlockChain.BestState.Beacon.BeaconCommittee)
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
								msgReady, _ := MakeMsgBFTReady()
								self.config.Server.PushMessageToBeacon(msgReady)
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
								err = self.config.BlockChain.InsertBeaconBlock(resBlk.(*blockchain.BeaconBlock))
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
									self.config.Server.PushMessageToAll(newBeaconBlockMsg)
								}

							} else {
								Logger.log.Error(err)
							}
							continue
						}
						if (self.config.NodeMode == "shard" || self.config.NodeMode == "auto") && role == "shard" {
							self.config.BlockChain.SyncShard(shardID)
							bftProtocol.RoleData.Committee = make([]string, len(self.config.BlockChain.BestState.Shard[shardID].ShardCommittee))
							copy(bftProtocol.RoleData.Committee, self.config.BlockChain.BestState.Shard[shardID].ShardCommittee)
							var (
								err    error
								resBlk interface{}
							)
							if self.config.BlockChain.IsReady(true, shardID) {
								shardRole := self.config.BlockChain.BestState.Shard[shardID].GetPubkeyRole(self.config.UserKeySet.GetPublicKeyB58())
								fmt.Println("My shard role", shardRole)
								switch shardRole {
								case "shard-proposer":
									resBlk, err = bftProtocol.Start(true, "shard", shardID)
									if err != nil {
										Logger.log.Error("PBFT fatal error", err)
										continue
									}
								case "shard-validator":
									msgReady, _ := MakeMsgBFTReady()
									self.config.Server.PushMessageToShard(msgReady, shardID)
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
									err = self.config.BlockChain.InsertShardBlock(resBlk.(*blockchain.ShardBlock))
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
										self.config.Server.PushMessageToBeacon(newShardToBeaconMsg)
									}
									//PUSH CROSS-SHARD
									newCrossShardBlocks := shardBlk.CreateAllCrossShardBlock()
									for sID, newCrossShardBlock := range newCrossShardBlocks {
										newCrossShardMsg, err := MakeMsgCrossShardBlock(newCrossShardBlock)
										if err == nil {
											self.config.Server.PushMessageToShard(newCrossShardMsg, sID)
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

func (self *Engine) Stop() error {
	self.Lock()
	defer self.Unlock()
	if !self.started {
		return errors.New("Consensus engine is already stopped")
	}

	self.started = false
	close(self.cQuit)
	return nil
}

// func (self *Engine) UpdateShardChain(block *blockchain.BlockV2) {
// 	err := self.config.BlockChain.ConnectBlock(block)
// 	if err != nil {
// 		Logger.log.Error(err)
// 		return
// 	}

// 	// update tx pool
// 	for _, tx := range block.Body.(*blockchain.BlockBodyShard).Transactions {
// 		self.config.MemPool.RemoveTx(tx)
// 	}

// 	// update candidate list
// 	// err = self.config.BlockChain.BestState[block.Header.shardID].Update(block)
// 	// if err != nil {
// 	// 	Logger.log.Errorf("Can not update merkle tree for block: %+v", err)
// 	// 	return
// 	// }
// 	// self.config.BlockChain.StoreBestState(block.Header.shardID)

// 	// self.knownChainsHeight.Lock()
// 	// if self.knownChainsHeight.Heights[block.Header.shardID] < int(block.Header.Height) {
// 	// 	self.knownChainsHeight.Heights[block.Header.shardID] = int(block.Header.Height)
// 	// 	self.sendBlockMsg(block)
// 	// }
// 	// self.knownChainsHeight.Unlock()
// 	// self.validatedChainsHeight.Lock()
// 	// self.validatedChainsHeight.Heights[block.Header.shardID] = int(block.Header.Height)
// 	// self.validatedChainsHeight.Unlock()

// 	// self.Committee().UpdateCommitteePoint(block.BlockProducer, block.Header.BlockCommitteeSigs)
// }
