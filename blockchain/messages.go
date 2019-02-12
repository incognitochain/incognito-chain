package blockchain

import (
	"fmt"

	libp2p "github.com/libp2p/go-libp2p-peer"
	"github.com/ninjadotorg/constant/cashec"
	"github.com/ninjadotorg/constant/common"
)

func (self *BlockChain) OnBlockShardReceived(block *ShardBlock) {
	if _, ok := self.syncStatus.Shard[block.Header.ShardID]; ok {

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
					err = self.InsertBeaconBlock(newBlk)
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

func (self *BlockChain) OnPeerStateReceived(beacon *ChainState, shard *map[byte]ChainState, shardToBeaconPool *map[byte][]common.Hash, crossShardPool *map[byte]map[byte][]common.Hash, peerID libp2p.ID) {
	var pState *peerState
	pState.Beacon = beacon
	userRole, shardID := self.BestState.Beacon.GetPubkeyRole(self.config.UserKeySet.GetPublicKeyB58())
	nodeMode := self.config.NodeMode
	if userRole == "beacon-proposer" || userRole == "beacon-validator" {
		pState.ShardToBeaconPool = shardToBeaconPool
	}
	if userRole == "shard" && (nodeMode == "auto" || nodeMode == "shard") {
		userRole = self.BestState.Shard[shardID].GetPubkeyRole(self.config.UserKeySet.GetPublicKeyB58())
		if userRole == "shard-proposer" || userRole == "shard-validator" {
			if state, ok := (*shard)[shardID]; ok {
				pState.Shard[shardID] = &state
				if pool, ok := (*crossShardPool)[shardID]; ok {
					pState.CrossShardPool = &pool
				}
			}
		}
	}
	for _, shardID := range self.config.RelayShards {
		if state, ok := (*shard)[shardID]; ok {
			pState.Shard[shardID] = &state
		}
	}
	self.syncStatus.PeersStateLock.Lock()
	self.syncStatus.PeersState[pState.Peer] = pState
	self.syncStatus.PeersStateLock.Unlock()

}

func (self *BlockChain) OnShardToBeaconBlockReceived(block ShardToBeaconBlock) {
	//TODO: check node mode -> node mode & role before add block to pool

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

	if err = self.config.ShardToBeaconPool.AddShardToBeaconBlock(block); err != nil {
		Logger.log.Error(err)
		return
	}

	//TODO review: synblock already find?
	//if self.BestState.Beacon.BestShardHeight[block.Header.ShardID] < block.Header.Height-1 {
	//	self.config.Server.PushMessageGetShardToBeacons(block.Header.ShardID, self.BestState.Beacon.BestShardHeight[block.Header.ShardID]+1, block.Header.Height)
	//}
}

func (self *BlockChain) OnCrossShardBlockReceived(block CrossShardBlock) {
	//TODO: check node mode -> node role before add block to pool
	err := self.config.CrossShardPool.AddCrossShardBlock(block)
	if err != nil {
		Logger.log.Error(err)
	}
}
