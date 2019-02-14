package blockchain

import (
	"errors"
	"fmt"
	"sync"
	"time"

	libp2p "github.com/libp2p/go-libp2p-peer"
	"github.com/ninjadotorg/constant/common"
)

type peerState struct {
	Shard             map[byte]*ChainState
	Beacon            *ChainState
	ShardToBeaconPool *map[byte][]common.Hash
	CrossShardPool    *map[byte][]common.Hash
	Peer              libp2p.ID
}

type peerSyncTimestamp struct {
	Time   int64
	PeerID libp2p.ID
}

type ChainState struct {
	Height        uint64
	BlockHash     common.Hash
	BestStateHash common.Hash
}

func (self *BlockChain) StartSyncBlk() {
	self.knownChainState.Beacon.Height = self.BestState.Beacon.BeaconHeight
	self.syncStatus.Beacon = true
	go func() {
		for {
			select {
			case <-self.cQuitSync:
				return
			case <-time.Tick(defaultBroadcastStateTime):
				self.InsertBlockFromPool()
				go self.config.Server.BoardcastNodeState()
			}
		}
	}()
	for {
		select {
		case <-self.cQuitSync:
			return
		case <-time.Tick(defaultProcessPeerStateTime):
			self.InsertBlockFromPool()
			self.syncStatus.PeersStateLock.Lock()

			userRole, shardID := self.BestState.Beacon.GetPubkeyRole(self.config.UserKeySet.GetPublicKeyB58())
			type reportedChainState struct {
				ClosestBeaconState ChainState
				ClosestShardsState map[byte]ChainState
				ShardToBeaconBlks  map[byte]map[common.Hash][]libp2p.ID
				CrossShardBlks     map[common.Hash][]libp2p.ID
			}
			var RCS reportedChainState
			for _, peerState := range self.syncStatus.PeersState {

			}
			currentBcnReqHeight := self.BestState.Beacon.BeaconHeight + 1
			for peerID := range self.syncStatus.PeersState {
				if currentBcnReqHeight+defaultMaxBlkReqPerPeer-1 >= RCS.ClosestBeaconState.Height {
					self.SyncBlkBeacon(false, false, nil, currentBcnReqHeight, RCS.ClosestBeaconState.Height, peerID)
				} else {
					self.SyncBlkBeacon(false, false, nil, currentBcnReqHeight, currentBcnReqHeight+defaultMaxBlkReqPerPeer-1, peerID)
					currentBcnReqHeight += defaultMaxBlkReqPerPeer - 1
				}
			}

			switch self.config.NodeMode {
			case "auto":
				switch userRole {
				case "beacon-proposer", "beacon-validator":
					for shardID := 0; shardID < common.SHARD_NUMBER; shardID++ {
						self.SyncBlkShardToBeacon(shardID, true, true)
					}
				case "shard":
					if _, ok := self.syncStatus.Shards[shardID]; !ok {
						self.SyncShard(shardID)
					}
					userShardRole := self.BestState.Shard[shardID].GetPubkeyRole(self.config.UserKeySet.GetPublicKeyB58())
					if userShardRole == "shard-proposer" || userShardRole == "shard-validator" {

					}
				}
			case "shard":
				if _, ok := self.syncStatus.Shards[shardID]; !ok {
					self.SyncShard(shardID)
				}
				userShardRole := self.BestState.Shard[shardID].GetPubkeyRole(self.config.UserKeySet.GetPublicKeyB58())
				if userShardRole == "shard-proposer" || userShardRole == "shard-validator" {

				}
			}

			for shardID := range self.syncStatus.Shards {
				currentShardReqHeight := self.BestState.Shard[shardID].ShardHeight + 1
				for peerID := range self.syncStatus.PeersState {
					if currentBcnReqHeight+defaultMaxBlkReqPerPeer-1 >= RCS.ClosestShardsState[shardID].Height {
						self.SyncBlkShard(shardID, false, false, nil, currentShardReqHeight, RCS.ClosestShardsState[shardID].Height, peerID)
					} else {
						self.SyncBlkShard(shardID, false, false, nil, currentShardReqHeight, currentShardReqHeight+defaultMaxBlkReqPerPeer-1, peerID)
						currentBcnReqHeight += defaultMaxBlkReqPerPeer - 1
					}
				}
			}

			self.syncStatus.PeersState = make(map[libp2p.ID]*peerState)
			self.syncStatus.PeersStateLock.Unlock()
		}
	}
}

func (self *BlockChain) SyncShard(shardID byte) error {
	self.syncStatus.Lock()
	defer self.syncStatus.Unlock()
	if _, ok := self.syncStatus.Shards[shardID]; ok {
		return errors.New("Shard " + fmt.Sprintf("%d", shardID) + " synchronzation is already started")
	}
	self.syncStatus.Shards[shardID] = struct{}{}
	return nil
}

func (self *BlockChain) StopSyncUnnecessaryShard() {
	for shardID := byte(0); shardID < common.SHARD_NUMBER; shardID++ {
		self.StopSyncShard(shardID)
	}
}

func (self *BlockChain) StopSyncShard(shardID byte) error {
	self.syncStatus.Lock()
	defer self.syncStatus.Unlock()
	if self.config.NodeMode == "auto" || self.config.NodeMode == "shard" {
		userRole, userShardID := self.BestState.Beacon.GetPubkeyRole(self.config.UserKeySet.GetPublicKeyB58())
		if userRole == "shard" && shardID == userShardID {
			return errors.New("Shard " + fmt.Sprintf("%d", shardID) + " synchronzation can't be stopped")
		}
	}
	if _, ok := self.syncStatus.Shards[shardID]; ok {
		if common.IndexOfByte(shardID, self.config.RelayShards) < 0 {
			delete(self.syncStatus.Shards, shardID)
			return nil
		}
		return errors.New("Shard " + fmt.Sprintf("%d", shardID) + " synchronzation can't be stopped")
	}
	return errors.New("Shard " + fmt.Sprintf("%d", shardID) + " synchronzation is already stopped")
}

func (self *BlockChain) GetCurrentSyncShards() []byte {
	self.syncStatus.Lock()
	defer self.syncStatus.Unlock()
	var currentSyncShards []byte
	for shardID, _ := range self.syncStatus.Shards {
		currentSyncShards = append(currentSyncShards, shardID)
	}
	return currentSyncShards
}

func (self *BlockChain) StopSync() error {
	close(self.cQuitSync)
	return nil
}

func (self *BlockChain) ResetCurrentSyncRecord() {
	self.syncStatus.Lock()
	defer self.syncStatus.Unlock()
	self.syncStatus.CurrentlySyncBeaconBlk = sync.Map{}
	self.syncStatus.CurrentlySyncShardBlk = sync.Map{}
	self.syncStatus.CurrentlySyncShardToBeaconBlk = sync.Map{}
	self.syncStatus.CurrentlySyncCrossShardBlk = sync.Map{}

}

func (self *BlockChain) SyncBlkBeacon(byHash bool, getFromPool bool, blksHash []common.Hash, from uint64, to uint64, peerID libp2p.ID) {
	if byHash {
		//Sync block by hash
		tempInterface, init := self.syncStatus.CurrentlySyncBeaconBlk.Load("byhash")
		blksNeedToGet := getBlkNeedToGetByHash(blksHash, tempInterface, init, peerID)
		if len(blksNeedToGet) > 0 {
			go self.config.Server.PushMessageGetBlockBeaconByHash(blksNeedToGet, getFromPool, peerID)
		}
	} else {
		//Sync by height
		tempInterface, init := self.syncStatus.CurrentlySyncBeaconBlk.Load("byheight")
		blkBatchsNeedToGet := getBlkNeedToGetByHeight(from, to, tempInterface, init, peerID)
		if len(blkBatchsNeedToGet) > 0 {
			for fromHeight, toHeight := range blkBatchsNeedToGet {
				go self.config.Server.PushMessageGetBlockBeaconByHeight(fromHeight, toHeight, peerID)
			}
		}
	}
}

func (self *BlockChain) SyncBlkShard(shardID byte, byHash bool, getFromPool bool, blksHash []common.Hash, from uint64, to uint64, peerID libp2p.ID) {
	if byHash {
		//Sync block by hash
		tempInterface, init := self.syncStatus.CurrentlySyncShardBlk.Load(SyncByHashKey + fmt.Sprint(shardID))
		blksNeedToGet := getBlkNeedToGetByHash(blksHash, tempInterface, init, peerID)
		if len(blksNeedToGet) > 0 {
			go self.config.Server.PushMessageGetBlockBeaconByHash(blksNeedToGet, getFromPool, peerID)
		}
	} else {
		//Sync by height
		tempInterface, init := self.syncStatus.CurrentlySyncShardBlk.Load(SyncByHeightKey + fmt.Sprint(shardID))
		blkBatchsNeedToGet := getBlkNeedToGetByHeight(from, to, tempInterface, init, peerID)
		if len(blkBatchsNeedToGet) > 0 {
			for fromHeight, toHeight := range blkBatchsNeedToGet {
				go self.config.Server.PushMessageGetBlockBeaconByHeight(fromHeight, toHeight, peerID)
			}
		}
	}
}

func (self *BlockChain) SyncBlkShardToBeacon(shardID byte, byHash bool, getFromPool bool, blksHash []common.Hash, from uint64, to uint64, peerID libp2p.ID) {
	if byHash {
		//Sync block by hash
		tempInterface, init := self.syncStatus.CurrentlySyncShardToBeaconBlk.Load(SyncByHashKey + fmt.Sprint(shardID))
		blksNeedToGet := getBlkNeedToGetByHash(blksHash, tempInterface, init, peerID)
		if len(blksNeedToGet) > 0 {
			go self.config.Server.PushMessageGetBlockShardToBeaconByHash(shardID, blksNeedToGet, getFromPool, peerID)
		}
	} else {
		//Sync by height
		tempInterface, init := self.syncStatus.CurrentlySyncShardToBeaconBlk.Load(SyncByHeightKey + fmt.Sprint(shardID))
		blkBatchsNeedToGet := getBlkNeedToGetByHeight(from, to, tempInterface, init, peerID)
		if len(blkBatchsNeedToGet) > 0 {
			for fromHeight, toHeight := range blkBatchsNeedToGet {
				go self.config.Server.PushMessageGetBlockShardToBeaconByHeight(shardID, fromHeight, toHeight, peerID)
			}
		}
	}
}

func (self *BlockChain) SyncBlkCrossShard(getFromPool bool, blksHash []common.Hash, fromShard byte, toShard byte, peerID libp2p.ID) {
	tempInterface, init := self.syncStatus.CurrentlySyncCrossShardBlk.Load(SyncByHashKey)
	blksNeedToGet := getBlkNeedToGetByHash(blksHash, tempInterface, init, peerID)
	if len(blksNeedToGet) > 0 {
		go self.config.Server.PushMessageGetBlockCrossShardByHash(fromShard, toShard, blksNeedToGet, getFromPool, peerID)
	}
}

func (self *BlockChain) InsertBlockFromPool() {
	blks, err := self.config.NodeBeaconPool.GetBlocks(self.BestState.Beacon.BeaconHeight + 1)
	if err != nil {
		Logger.log.Error(err)
	} else {
		for idx, newBlk := range blks {
			err = self.InsertBeaconBlock(&newBlk)
			if err != nil {
				Logger.log.Error(err)
				for idx2 := idx; idx2 < len(blks); idx2++ {
					self.config.NodeBeaconPool.PushBlock(blks[idx2])
				}
				break
			}
		}
	}

	for shardID, _ := range self.syncStatus.Shards {
		blks, err := self.config.NodeShardPool.GetBlocks(shardID, self.BestState.Shard[shardID].ShardHeight+1)
		if err != nil {
			Logger.log.Error(err)
			continue
		}
		for idx, newBlk := range blks {
			err = self.InsertShardBlock(&newBlk)
			if err != nil {
				Logger.log.Error(err)
				if idx < len(blks)-1 {
					for idx2 := idx; idx2 < len(blks); idx2++ {
						self.config.NodeShardPool.PushBlock(blks[idx2])
					}
					break
				}
			}
		}
	}
}
