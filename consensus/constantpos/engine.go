package constantpos

import (
	"errors"
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
	Layers struct {
		Beacon *Layerbeacon
		Shard  *Layershard
	}
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
	NodeRole    string
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
	self.started = true
	Logger.log.Info("Start consensus with key", self.config.UserKeySet.GetPublicKeyB58())

	go func() {
		for {
			select {
			case <-self.cQuit:
				return
			default:
				if self.config.BlockChain.IsReady() {
					role, shardID := self.config.BlockChain.BestState.Beacon.GetPubkeyRole(self.config.UserKeySet.GetPublicKeyB58())
					if role != "" {
						bftProtocol := &BFTProtocol{
							cBFTMsg:    self.cBFTMsg,
							BlockGen:   self.config.BlockGen,
							UserKeySet: self.config.UserKeySet,
							Chain:      self.config.BlockChain,
							Server:     self.config.Server,
							// Committee:  self.config.BlockChain.BestState.Beacon.BeaconCommittee,
						}
						if (self.config.NodeRole == "beacon" || self.config.NodeRole == "auto") && role != "shard" {
							bftProtocol.Committee = self.config.BlockChain.BestState.Beacon.BeaconCommittee
							switch role {
							case "beacon-proposer":
								// prevBlock :=	self.config.BlockChain.GetMayBeAcceptBlockBeacon()
								prevBlock := &blockchain.BeaconBlock{}
								bftProtocol.Start(true, "beacon", 0, prevBlock.AggregatedSig, prevBlock.ValidatorsIdx)
							case "beacon-validator":
								bftProtocol.Start(false, "beacon", 0, "", []int{})
								// case "beacon-pending":
							}
							continue
						}
						if (self.config.NodeRole == "shard" || self.config.NodeRole == "auto") && role == "shard" {
							bftProtocol.Committee = self.config.BlockChain.BestState.Shard[shardID].ShardCommittee
							shardRole := self.config.BlockChain.BestState.Shard[shardID].GetPubkeyRole(self.config.UserKeySet.GetPublicKeyB58())
							switch shardRole {
							case "shard-proposer":
								prevBlock := &blockchain.ShardBlock{}
								bftProtocol.Start(true, "shard", 0, prevBlock.AggregatedSig, prevBlock.ValidatorsIdx)
							case "shard-validator":
								bftProtocol.Start(false, "shard", 0, "", []int{})
							default:
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
	if self.Layers.Beacon != nil {
		self.Layers.Beacon.Stop()
	}
	if self.Layers.Beacon != nil {
		self.Layers.Beacon.Stop()
	}
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

func (self *Engine) GetShardCommittee(shardID byte) CommitteeStruct {
	return CommitteeStruct{}
}

func (self *Engine) GetBeaconCommittee() CommitteeStruct {
	return CommitteeStruct{}
}
