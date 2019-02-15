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
	CrossShardPool    map[byte]*map[byte][]common.Hash
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

			userRole, userShardID := self.BestState.Beacon.GetPubkeyRole(self.config.UserKeySet.GetPublicKeyB58())
			userShardRole := self.BestState.Shard[userShardID].GetPubkeyRole(self.config.UserKeySet.GetPublicKeyB58())
			type reportedChainState struct {
				ClosestBeaconState ChainState
				ClosestShardsState map[byte]ChainState
				ShardToBeaconBlks  map[byte]map[libp2p.ID][]common.Hash
				CrossShardBlks     map[byte]map[libp2p.ID][]common.Hash
			}
			RCS := reportedChainState{
				ClosestShardsState: make(map[byte]ChainState),
				ShardToBeaconBlks:  make(map[byte]map[libp2p.ID][]common.Hash),
				CrossShardBlks:     make(map[byte]map[libp2p.ID][]common.Hash),
			}
			for peerID, peerState := range self.syncStatus.PeersState {
				if peerState.Beacon.Height > self.BestState.Beacon.BeaconHeight {
					if RCS.ClosestBeaconState.Height == 0 {
						RCS.ClosestBeaconState = *peerState.Beacon
					} else {
						if peerState.Beacon.Height < RCS.ClosestBeaconState.Height {
							RCS.ClosestBeaconState = *peerState.Beacon
						}
					}
					for shardID := range self.syncStatus.Shards {
						if shardState, ok := peerState.Shard[shardID]; ok {
							if shardState.Height > self.BestState.Shard[shardID].ShardHeight {
								if RCS.ClosestShardsState[shardID].Height == 0 {
									RCS.ClosestShardsState[shardID] = *shardState
								} else {
									if shardState.Height < RCS.ClosestShardsState[shardID].Height {
										RCS.ClosestShardsState[shardID] = *shardState
									}
								}
							}
						}
					}
					// record pool state
					switch self.config.NodeMode {
					case "auto":
						switch userRole {
						case "beacon-proposer", "beacon-validator":
							if peerState.ShardToBeaconPool != nil {
								for shardID, blksHash := range *peerState.ShardToBeaconPool {
									RCS.ShardToBeaconBlks[shardID][peerID] = blksHash
								}
							}
						case "shard":
							if userShardRole == "shard-proposer" || userShardRole == "shard-validator" {
								if pool, ok := peerState.CrossShardPool[userShardID]; ok {
									for shardID, blks := range *pool {
										RCS.CrossShardBlks[shardID][peerID] = blks
									}
								}
							}
						}
					case "beacon":
						if userRole == "beacon-proposer" || userRole == "beacon-validator" {
							if peerState.ShardToBeaconPool != nil {
								for shardID, blksHash := range *peerState.ShardToBeaconPool {
									RCS.ShardToBeaconBlks[shardID][peerID] = blksHash
								}
							}
						}
					case "shard":
						if userShardRole == "shard-proposer" || userShardRole == "shard-validator" {
							if pool, ok := peerState.CrossShardPool[userShardID]; ok {
								for shardID, blks := range *pool {
									RCS.CrossShardBlks[shardID][peerID] = blks
								}
							}
						}
					}
				}
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
					for shardID, peer := range RCS.ShardToBeaconBlks {
						for peerID, blks := range peer {
							self.SyncBlkShardToBeacon(shardID, true, true, blks, 0, 0, peerID)
						}
					}
				case "shard":
					if userShardRole == "shard-pending" || userShardRole == "shard-proposer" || userShardRole == "shard-validator" {
						if _, ok := self.syncStatus.Shards[userShardID]; !ok {
							self.SyncShard(userShardID)
						}
						if userShardRole == "shard-proposer" || userShardRole == "shard-validator" {

						}
					}
				}
			case "beacon":
				if userRole == "beacon-proposer" || userRole == "beacon-validator" {
					for shardID, peer := range RCS.ShardToBeaconBlks {
						for peerID, blks := range peer {
							self.SyncBlkShardToBeacon(shardID, true, true, blks, 0, 0, peerID)
						}
					}
				}
			case "shard":
				if userShardRole == "shard-pending" || userShardRole == "shard-proposer" || userShardRole == "shard-validator" {
					if _, ok := self.syncStatus.Shards[userShardID]; !ok {
						self.SyncShard(userShardID)
					}
					if userShardRole == "shard-proposer" || userShardRole == "shard-validator" {
						for shardID, peer := range RCS.CrossShardBlks {
							for peerID, blks := range peer {
								self.SyncBlkCrossShard(true, blks, shardID, userShardID, peerID)
							}
						}
					}
				}
			}

			for shardID := range self.syncStatus.Shards {
				currentShardReqHeight := self.BestState.Shard[shardID].ShardHeight + 1
				for peerID := range self.syncStatus.PeersState {
					if _, ok := self.syncStatus.PeersState[peerID].Shard[shardID]; ok {
						if currentBcnReqHeight+defaultMaxBlkReqPerPeer-1 >= RCS.ClosestShardsState[shardID].Height {
							self.SyncBlkShard(shardID, false, false, nil, currentShardReqHeight, RCS.ClosestShardsState[shardID].Height, peerID)
						} else {
							self.SyncBlkShard(shardID, false, false, nil, currentShardReqHeight, currentShardReqHeight+defaultMaxBlkReqPerPeer-1, peerID)
							currentBcnReqHeight += defaultMaxBlkReqPerPeer - 1
						}
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
