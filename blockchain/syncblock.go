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
			self.syncStatus.Lock()
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
				if peerState.Beacon.Height >= self.BestState.Beacon.BeaconHeight {
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
					case common.NODEMODE_AUTO:
						switch userRole {
						case common.BEACON_PROPOSER_ROLE, common.BEACON_VALIDATOR_ROLE:
							if peerState.ShardToBeaconPool != nil {
								for shardID, blksHash := range *peerState.ShardToBeaconPool {
									RCS.ShardToBeaconBlks[shardID][peerID] = blksHash
								}
							}
						case common.SHARD_ROLE:
							if userShardRole == common.SHARD_PROPOSER_ROLE || userShardRole == common.SHARD_VALIDATOR_ROLE {
								if pool, ok := peerState.CrossShardPool[userShardID]; ok {
									for shardID, blks := range *pool {
										RCS.CrossShardBlks[shardID][peerID] = blks
									}
								}
							}
						}
					case common.NODEMODE_BEACON:
						if userRole == common.BEACON_PROPOSER_ROLE || userRole == common.BEACON_VALIDATOR_ROLE {
							if peerState.ShardToBeaconPool != nil {
								for shardID, blksHash := range *peerState.ShardToBeaconPool {
									RCS.ShardToBeaconBlks[shardID][peerID] = blksHash
								}
							}
						}
					case common.NODEMODE_SHARD:
						if userShardRole == common.SHARD_PROPOSER_ROLE || userShardRole == common.SHARD_VALIDATOR_ROLE {
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
			case common.NODEMODE_AUTO:
				switch userRole {
				case common.BEACON_PROPOSER_ROLE, common.BEACON_VALIDATOR_ROLE:
					for shardID, peer := range RCS.ShardToBeaconBlks {
						for peerID, blks := range peer {
							self.SyncBlkShardToBeacon(shardID, true, true, blks, 0, 0, peerID)
						}
					}
					for shardID := byte(0); shardID < common.SHARD_NUMBER; shardID++ {
						if self.BestState.Beacon.BestShardHeight[shardID] < RCS.ClosestShardsState[shardID].Height {
							currentShardReqHeight := self.BestState.Beacon.BestShardHeight[shardID] + 1
							for peerID, peerState := range self.syncStatus.PeersState {
								if _, ok := peerState.Shard[shardID]; ok {
									if currentShardReqHeight+defaultMaxBlkReqPerPeer-1 >= RCS.ClosestShardsState[shardID].Height {
										self.SyncBlkShardToBeacon(shardID, false, false, nil, currentShardReqHeight, RCS.ClosestShardsState[shardID].Height, peerID)
									} else {
										self.SyncBlkShardToBeacon(shardID, false, false, nil, currentShardReqHeight, currentShardReqHeight+defaultMaxBlkReqPerPeer-1, peerID)
										currentShardReqHeight += defaultMaxBlkReqPerPeer - 1
									}
								}
							}
						}
					}
				case common.SHARD_ROLE:
					if userShardRole == common.SHARD_PENDING_ROLE || userShardRole == common.SHARD_PROPOSER_ROLE || userShardRole == common.SHARD_VALIDATOR_ROLE {
						if _, ok := self.syncStatus.Shards[userShardID]; !ok {
							self.SyncShard(userShardID)
						}
						if userShardRole == common.SHARD_PROPOSER_ROLE || userShardRole == common.SHARD_VALIDATOR_ROLE {
							for shardID, peer := range RCS.CrossShardBlks {
								for peerID, blks := range peer {
									self.SyncBlkCrossShard(true, blks, shardID, userShardID, peerID)
								}
							}
						}
					}
				}
			case common.NODEMODE_BEACON:
				if userRole == common.BEACON_PROPOSER_ROLE || userRole == common.BEACON_VALIDATOR_ROLE {
					for shardID, peer := range RCS.ShardToBeaconBlks {
						for peerID, blks := range peer {
							self.SyncBlkShardToBeacon(shardID, true, true, blks, 0, 0, peerID)
						}
					}
					for shardID := byte(0); shardID < common.SHARD_NUMBER; shardID++ {
						if self.BestState.Beacon.BestShardHeight[shardID] < RCS.ClosestShardsState[shardID].Height {
							currentShardReqHeight := self.BestState.Beacon.BestShardHeight[shardID] + 1
							for peerID, peerState := range self.syncStatus.PeersState {
								if _, ok := peerState.Shard[shardID]; ok {
									if currentShardReqHeight+defaultMaxBlkReqPerPeer-1 >= RCS.ClosestShardsState[shardID].Height {
										self.SyncBlkShardToBeacon(shardID, false, false, nil, currentShardReqHeight, RCS.ClosestShardsState[shardID].Height, peerID)
									} else {
										self.SyncBlkShardToBeacon(shardID, false, false, nil, currentShardReqHeight, currentShardReqHeight+defaultMaxBlkReqPerPeer-1, peerID)
										currentShardReqHeight += defaultMaxBlkReqPerPeer - 1
									}
								}
							}
						}
					}
				}
			case common.NODEMODE_SHARD:
				if userShardRole == common.SHARD_PENDING_ROLE || userShardRole == common.SHARD_PROPOSER_ROLE || userShardRole == common.SHARD_VALIDATOR_ROLE {
					if _, ok := self.syncStatus.Shards[userShardID]; !ok {
						self.SyncShard(userShardID)
					}
					if userShardRole == common.SHARD_PROPOSER_ROLE || userShardRole == common.SHARD_VALIDATOR_ROLE {
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
						if currentShardReqHeight+defaultMaxBlkReqPerPeer-1 >= RCS.ClosestShardsState[shardID].Height {
							self.SyncBlkShard(shardID, false, false, nil, currentShardReqHeight, RCS.ClosestShardsState[shardID].Height, peerID)
						} else {
							self.SyncBlkShard(shardID, false, false, nil, currentShardReqHeight, currentShardReqHeight+defaultMaxBlkReqPerPeer-1, peerID)
							currentShardReqHeight += defaultMaxBlkReqPerPeer - 1
						}
					}
				}
			}

			self.syncStatus.PeersState = make(map[libp2p.ID]*peerState)
			self.syncStatus.Unlock()
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
	if self.config.NodeMode == common.NODEMODE_AUTO || self.config.NodeMode == common.NODEMODE_SHARD {
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

//SyncBlkBeacon Send a req to sync beacon block
func (self *BlockChain) SyncBlkBeacon(byHash bool, getFromPool bool, blksHash []common.Hash, from uint64, to uint64, peerID libp2p.ID) {
	if byHash {
		//Sync block by hash
		tempInterface, init := self.syncStatus.CurrentlySyncBeaconBlk.Load(SyncByHashKey)
		blksNeedToGet, blksSyncByHash := getBlkNeedToGetByHash(blksHash, tempInterface, init, peerID)
		if len(blksNeedToGet) > 0 {
			go self.config.Server.PushMessageGetBlockBeaconByHash(blksNeedToGet, getFromPool, peerID)
		}
		self.syncStatus.CurrentlySyncBeaconBlk.Store(SyncByHashKey, blksSyncByHash)
	} else {
		//Sync by height
		tempInterface, init := self.syncStatus.CurrentlySyncBeaconBlk.Load(SyncByHeightKey)
		blkBatchsNeedToGet, blksSyncByHeight := getBlkNeedToGetByHeight(from, to, tempInterface, init, peerID)
		if len(blkBatchsNeedToGet) > 0 {
			for fromHeight, toHeight := range blkBatchsNeedToGet {
				go self.config.Server.PushMessageGetBlockBeaconByHeight(fromHeight, toHeight, peerID)
			}
		}
		self.syncStatus.CurrentlySyncBeaconBlk.Store(SyncByHeightKey, blksSyncByHeight)
	}
}

//SyncBlkShard Send a req to sync shard block
func (self *BlockChain) SyncBlkShard(shardID byte, byHash bool, getFromPool bool, blksHash []common.Hash, from uint64, to uint64, peerID libp2p.ID) {
	if byHash {
		//Sync block by hash
		tempInterface, init := self.syncStatus.CurrentlySyncShardBlk.Load(SyncByHashKey + fmt.Sprint(shardID))
		blksNeedToGet, blksSyncByHash := getBlkNeedToGetByHash(blksHash, tempInterface, init, peerID)
		if len(blksNeedToGet) > 0 {
			go self.config.Server.PushMessageGetBlockShardByHash(shardID, blksNeedToGet, getFromPool, peerID)
		}
		self.syncStatus.CurrentlySyncShardBlk.Store(SyncByHashKey+fmt.Sprint(shardID), blksSyncByHash)
	} else {
		//Sync by height
		tempInterface, init := self.syncStatus.CurrentlySyncShardBlk.Load(SyncByHeightKey + fmt.Sprint(shardID))
		blkBatchsNeedToGet, blksSyncByHeight := getBlkNeedToGetByHeight(from, to, tempInterface, init, peerID)
		if len(blkBatchsNeedToGet) > 0 {
			for fromHeight, toHeight := range blkBatchsNeedToGet {
				go self.config.Server.PushMessageGetBlockShardByHeight(shardID, fromHeight, toHeight, peerID)
			}
		}
		self.syncStatus.CurrentlySyncShardBlk.Store(SyncByHeightKey+fmt.Sprint(shardID), blksSyncByHeight)
	}
}

//SyncBlkShardToBeacon Send a req to sync shardToBeacon block
func (self *BlockChain) SyncBlkShardToBeacon(shardID byte, byHash bool, getFromPool bool, blksHash []common.Hash, from uint64, to uint64, peerID libp2p.ID) {
	if byHash {
		//Sync block by hash
		tempInterface, init := self.syncStatus.CurrentlySyncShardToBeaconBlk.Load(SyncByHashKey + fmt.Sprint(shardID))
		blksNeedToGet, blksSyncByHash := getBlkNeedToGetByHash(blksHash, tempInterface, init, peerID)
		if len(blksNeedToGet) > 0 {
			go self.config.Server.PushMessageGetBlockShardToBeaconByHash(shardID, blksNeedToGet, getFromPool, peerID)
		}
		self.syncStatus.CurrentlySyncShardToBeaconBlk.Store(SyncByHashKey+fmt.Sprint(shardID), blksSyncByHash)
	} else {
		//Sync by height
		tempInterface, init := self.syncStatus.CurrentlySyncShardToBeaconBlk.Load(SyncByHeightKey + fmt.Sprint(shardID))
		blkBatchsNeedToGet, blksSyncByHeight := getBlkNeedToGetByHeight(from, to, tempInterface, init, peerID)
		if len(blkBatchsNeedToGet) > 0 {
			for fromHeight, toHeight := range blkBatchsNeedToGet {
				go self.config.Server.PushMessageGetBlockShardToBeaconByHeight(shardID, fromHeight, toHeight, peerID)
			}
		}
		self.syncStatus.CurrentlySyncShardToBeaconBlk.Store(SyncByHeightKey+fmt.Sprint(shardID), blksSyncByHeight)
	}
}

//SyncBlkCrossShard Send a req to sync crossShard block
func (self *BlockChain) SyncBlkCrossShard(getFromPool bool, blksHash []common.Hash, fromShard byte, toShard byte, peerID libp2p.ID) {
	tempInterface, init := self.syncStatus.CurrentlySyncCrossShardBlk.Load(SyncByHashKey)
	blksNeedToGet, blksSyncByHash := getBlkNeedToGetByHash(blksHash, tempInterface, init, peerID)
	if len(blksNeedToGet) > 0 {
		go self.config.Server.PushMessageGetBlockCrossShardByHash(fromShard, toShard, blksNeedToGet, getFromPool, peerID)
	}
	self.syncStatus.CurrentlySyncCrossShardBlk.Store(SyncByHashKey, blksSyncByHash)
}

func (self *BlockChain) InsertBlockFromPool() {
	blks, err := self.config.NodeBeaconPool.GetBlocks(self.BestState.Beacon.BeaconHeight + 1)
	if err != nil {
		Logger.log.Error(err)
	} else {
		for idx, newBlk := range blks {
			err = self.InsertBeaconBlock(&newBlk, false)
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
