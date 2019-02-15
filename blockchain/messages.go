package blockchain

import (
	"fmt"

	libp2p "github.com/libp2p/go-libp2p-peer"
	"github.com/ninjadotorg/constant/cashec"
	"github.com/ninjadotorg/constant/common"
)

func (self *BlockChain) OnPeerStateReceived(beacon *ChainState, shard *map[byte]ChainState, shardToBeaconPool *map[byte][]common.Hash, crossShardPool *map[byte]map[byte][]common.Hash, peerID libp2p.ID) {
	if beacon.Height >= self.BestState.Beacon.BeaconHeight {
		pState := &peerState{
			Shard:  make(map[byte]*ChainState),
			Beacon: beacon,
			Peer:   peerID,
		}
		userRole, userShardID := self.BestState.Beacon.GetPubkeyRole(self.config.UserKeySet.GetPublicKeyB58())
		nodeMode := self.config.NodeMode
		if userRole == "beacon-proposer" || userRole == "beacon-validator" {
			pState.ShardToBeaconPool = shardToBeaconPool
		}
		if userRole == "shard" && (nodeMode == "auto" || nodeMode == "shard") {
			userRole = self.BestState.Shard[userShardID].GetPubkeyRole(self.config.UserKeySet.GetPublicKeyB58())
			if userRole == "shard-proposer" || userRole == "shard-validator" {
				if shardState, ok := (*shard)[userShardID]; ok && shardState.Height >= self.BestState.Shard[userShardID].ShardHeight {
					pState.Shard[userShardID] = &shardState
					if pool, ok := (*crossShardPool)[userShardID]; ok {
						pState.CrossShardPool = make(map[byte]*map[byte][]common.Hash)
						pState.CrossShardPool[userShardID] = &pool
					}
				}
			}
		}
		for shardID := range self.syncStatus.Shards {
			if shardState, ok := (*shard)[shardID]; ok {
				if shardState.Height > self.BestState.Shard[shardID].ShardHeight {
					pState.Shard[shardID] = &shardState
				}
			}
		}
		self.syncStatus.PeersStateLock.Lock()
		self.syncStatus.PeersState[pState.Peer] = pState
		self.syncStatus.PeersStateLock.Unlock()
	}
}

func (self *BlockChain) OnBlockShardReceived(newBlk *ShardBlock) {
	if _, ok := self.syncStatus.Shards[newBlk.Header.ShardID]; ok {
		fmt.Println("Shard block received")
		if self.BestState.Shard[newBlk.Header.ShardID].ShardHeight < newBlk.Header.Height {
			blkHash := newBlk.Header.Hash()
			err := cashec.ValidateDataB58(newBlk.Header.Producer, newBlk.ProducerSig, blkHash.GetBytes())
			if err != nil {
				Logger.log.Error(err)
				return
			} else {
				if self.BestState.Beacon.BeaconHeight == newBlk.Header.Height-1 {
					err = self.InsertShardBlock(newBlk)
					if err != nil {
						Logger.log.Error(err)
						return
					}
				} else {
					self.config.NodeShardPool.PushBlock(*newBlk)
				}
			}
		}
	}
}

func (self *BlockChain) OnBlockBeaconReceived(newBlk *BeaconBlock) {
	if self.syncStatus.Beacon {
		fmt.Println("Beacon block received")
		if self.BestState.Beacon.BeaconHeight < newBlk.Header.Height {
			blkHash := newBlk.Header.Hash()
			err := cashec.ValidateDataB58(newBlk.Header.Producer, newBlk.ProducerSig, blkHash.GetBytes())
			if err != nil {
				Logger.log.Error(err)
				return
			} else {
				if self.BestState.Beacon.BeaconHeight == newBlk.Header.Height-1 {
					err = self.InsertBeaconBlock(newBlk, false)
					if err != nil {
						Logger.log.Error(err)
						return
					}
				} else {
					self.config.NodeBeaconPool.PushBlock(*newBlk)
				}
			}
		}
	}
}

func (self *BlockChain) OnShardToBeaconBlockReceived(block ShardToBeaconBlock) {
	//TODO: check node mode -> node mode & role before add block to pool
	fmt.Println("Blockchain Message/OnShardToBeaconBlockReceived: Block Height", block.Header.Height)
	blkHash := block.Header.Hash()
	err := cashec.ValidateDataB58(block.Header.Producer, block.ProducerSig, blkHash.GetBytes())

	if err != nil {
		Logger.log.Debugf("Invalid Producer Signature of block height %+v in Shard %+v", block.Header.Height, block.Header.ShardID)
		return
	}
	if block.Header.Version != VERSION {
		Logger.log.Debugf("Invalid Verion of block height %+v in Shard %+v", block.Header.Height, block.Header.ShardID)
		return
	}

	//TODO: what if shard to beacon from old committee
	if err = ValidateAggSignature(block.ValidatorsIdx, self.BestState.Beacon.ShardCommittee[block.Header.ShardID], block.AggregatedSig, block.R, block.Hash()); err != nil {
		Logger.log.Error(err)
		return
	}

	from, to, err := self.config.ShardToBeaconPool.AddShardToBeaconBlock(block)
	if err != nil {
		Logger.log.Error(err)
		return
	}
	if from != 0 || to != 0 {
		fmt.Printf("Message/SyncBlkShardToBeacon, from %+v to %+v \n", from, to)
		self.SyncBlkShardToBeacon(block.Header.ShardID, false, false, []common.Hash{}, from, to, "")
	}
}

func (self *BlockChain) OnCrossShardBlockReceived(block CrossShardBlock) {
	//TODO: check node mode -> node role before add block to pool
	err := self.config.CrossShardPool.AddCrossShardBlock(block)
	if err != nil {
		Logger.log.Error(err)
	}
}
