package constantpos

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/ninjadotorg/constant/blockchain"
	"github.com/ninjadotorg/constant/cashec"
	"github.com/ninjadotorg/constant/connmanager"
	"github.com/ninjadotorg/constant/wire"
)

type Engine struct {
	sync.Mutex
	started bool

	// channel
	cQuit   chan struct{}
	cBFTMsg chan wire.Message

	config EngineConfig
	// Layers struct {
	// 	Beacon *Layerbeacon
	// 	Shard  *Layershard
	// }
	CurrentRole role
}

type role struct {
	Role    string
	ShardID byte
}

type EngineConfig struct {
	BlockChain  *blockchain.BlockChain
	ConnManager *connmanager.ConnManager
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
	self.started = true
	Logger.log.Info("Start consensus with key", self.config.UserKeySet.GetPublicKeyB58())
	fmt.Println(self.config.BlockChain.BestState.Beacon.BeaconCommittee)

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

					fmt.Println(self.config.NodeMode, role)
					if role != "" {
						self.cBFTMsg = make(chan wire.Message)
						bftProtocol := &BFTProtocol{
							cQuit:      self.cQuit,
							cBFTMsg:    self.cBFTMsg,
							BlockGen:   self.config.BlockGen,
							UserKeySet: self.config.UserKeySet,
							Chain:      self.config.BlockChain,
							Server:     self.config.Server,
						}
						if (self.config.NodeMode == "beacon" || self.config.NodeMode == "auto") && role != "shard" {
							bftProtocol.Committee = make([]string, len(self.config.BlockChain.BestState.Beacon.BeaconCommittee))
							copy(bftProtocol.Committee, self.config.BlockChain.BestState.Beacon.BeaconCommittee)
							switch role {
							case "beacon-proposer":
								// prevBlock :=	self.config.BlockChain.GetMayBeAcceptBlockBeacon()
								prevBlock := &blockchain.BeaconBlock{}
								err := bftProtocol.Start(true, "beacon", 0, prevBlock.AggregatedSig, prevBlock.ValidatorsIdx)
								if err != nil {
									Logger.log.Error("PBFT fatal error", err)
									continue
								}
								//TODO Insert block to chain
							case "beacon-validator":
								err := bftProtocol.Start(false, "beacon", 0, "", []int{})
								if err != nil {
									Logger.log.Error("PBFT fatal error", err)
									continue
								}
								//TODO Insert block to chain
							// case "beacon-pending":
							default:
							}

							continue
						}
						if (self.config.NodeMode == "shard" || self.config.NodeMode == "auto") && role == "shard" {
							bftProtocol.Committee = make([]string, len(self.config.BlockChain.BestState.Shard[shardID].ShardCommittee))
							copy(bftProtocol.Committee, self.config.BlockChain.BestState.Shard[shardID].ShardCommittee)
							if self.config.BlockChain.IsReady(true, shardID) {
								shardRole := self.config.BlockChain.BestState.Shard[shardID].GetPubkeyRole(self.config.UserKeySet.GetPublicKeyB58())
								switch shardRole {
								case "shard-proposer":
									prevBlock := &blockchain.ShardBlock{}
									err := bftProtocol.Start(true, "shard", 0, prevBlock.AggregatedSig, prevBlock.ValidatorsIdx)
									if err != nil {
										Logger.log.Error("PBFT fatal error", err)
										continue
									}
									//TODO Insert block to chain
								case "shard-validator":
									err := bftProtocol.Start(false, "shard", 0, "", []int{})
									if err != nil {
										Logger.log.Error("PBFT fatal error", err)
										continue
									}
									//TODO Insert block to chain
								default:
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
