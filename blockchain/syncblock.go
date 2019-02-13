package blockchain

import (
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
				blks, err := self.config.NodeBeaconPool.GetBlocks(self.BestState.Beacon.BeaconHeight + 1)
				if err != nil {
					Logger.log.Error(err)
					continue
				}
				for _, newBlk := range blks {
					err = self.InsertBeaconBlock(&newBlk)
					if err != nil {
						Logger.log.Error(err)
						continue
					}
				}
				go self.config.Server.BoardcastNodeState()
			}
		}
	}()
	for {
		select {
		case <-self.cQuitSync:
			return
		case <-time.Tick(defaultProcessPeerStateTime):
			self.syncStatus.PeersStateLock.Lock()
			for _, peerState := range self.syncStatus.PeersState {
				_ = peerState
			}
			self.syncStatus.PeersState = make(map[libp2p.ID]*peerState)
			self.syncStatus.PeersStateLock.Unlock()

			// userRole, shardID := self.BestState.Beacon.GetPubkeyRole(self.config.UserKeySet.GetPublicKeyB58())
			// if self.BestState.Beacon.BeaconHeight < pState.Beacon.Height {
			// 	if self.knownChainState.Beacon.Height < pState.Beacon.Height {
			// 		self.knownChainState.Beacon = *pState.Beacon
			// 	}
			// } else {
			// 	if self.BestState.Beacon.BeaconHeight == pState.Beacon.Height {
			// 		// check shardToBeacon pool state
			// 		if userRole == "beacon-proposer" || userRole == "beacon-validator" {
			// 			if len(beaconState.State.ShardsPoolState) > 0 {
			// 				myPoolPending := self.config.ShardToBeaconPool.GetValidPendingBlockHash()
			// 				for shardID, peerPoolBlks := range beaconState.State.ShardsPoolState {
			// 					myShardPoolBlks, ok := myPoolPending[shardID]
			// 					if ok {
			// 						blksNeedToSync := GetDiffHashesOf(peerPoolBlks, myShardPoolBlks)
			// 						self.SyncBlkShardToBeacon(true, true, blks9NeedToSync, 0, 0, beaconState.Peer)
			// 					} else {
			// 						// sync all blks of this shard
			// 						self.SyncBlkShardToBeacon(true, true, peerPoolBlks, 0, 0, beaconState.Peer)
			// 					}
			// 				}
			// 			}
			// 		}
			// 	}
			// }
			// if self.config.NodeMode == "auto" || self.config.NodeMode == "shard" {
			// 	if state, ok := pState.Shard[shardID]; ok{
			// 		if self.BestState.Shard[shardID].ShardHeight < p
			// 	}
			// }
		}
	}
}

func (self *BlockChain) SyncShard(shardID byte) error {
	self.syncStatus.Lock()
	defer self.syncStatus.Unlock()

	// if _, ok := self.syncStatus.Shard[shardID]; ok {
	// 	return errors.New("Shard " + fmt.Sprintf("%d", shardID) + " synchronzation is already started")
	// }
	// var cSyncShardQuit chan struct{}
	// cSyncShardQuit = make(chan struct{})
	// self.syncStatus.Shard[shardID] = cSyncShardQuit
	// self.knownChainState.Shards[shardID] = ShardChainState{
	// 	Height:  self.BestState.Shard[shardID].ShardHeight,
	// 	ShardID: shardID,
	// }
	// var shardStateCh chan *PeerShardChainState
	// shardStateCh = make(chan *PeerShardChainState)

	// self.ShardStateCh[shardID] = shardStateCh
	// self.newShardBlkCh[shardID] = &newShardBlkCh
	// go func(shardID byte) {
	// 	//used for fancy block retriever but not implement that now :p
	// 	var peerChainState map[libp2p.ID]PeerShardChainState
	// 	peerChainState = make(map[libp2p.ID]PeerShardChainState)
	// 	_ = peerChainState
	// 	getStateWaitTime := time.Duration(defaultGetStateWaitTime)
	// 	for {
	// 		select {
	// 		case <-self.cQuitSync:
	// 			return
	// 		case <-cSyncShardQuit:
	// 			close(shardStateCh)
	// 			close(newShardBlkCh)
	// 			delete(self.newShardBlkCh, shardID)
	// 			delete(self.ShardStateCh, shardID)
	// 			delete(self.syncStatus.Shard, shardID)
	// 			return
	// 		case shardState := <-shardStateCh:
	// 			if self.BestState.Shard[shardID].ShardHeight < shardState.State.Height {
	// 				if self.knownChainState.Shards[shardID].Height < shardState.State.Height {
	// 					self.knownChainState.Shards[shardID] = *shardState.State
	// 					if getStateWaitTime == defaultGetStateWaitTime*2 {
	// 						getStateWaitTime -= defaultGetStateWaitTime
	// 					}
	// 					// go self.config.Server.PushMessageGetBlockShard(shardID, self.BestState.Shard[shardID].ShardHeight+1, shardState.State.Height, shardState.Peer)
	// 				} else {
	// 					if getStateWaitTime == defaultGetStateWaitTime {
	// 						getStateWaitTime += defaultGetStateWaitTime
	// 					}
	// 				}
	// 			} else {
	// 				if getStateWaitTime == defaultGetStateWaitTime {
	// 					getStateWaitTime += defaultGetStateWaitTime
	// 				}
	// 			}

	// 			// check shardToBeacon pool state
	// 			if len(shardState.State.CrossShardsPoolState) > 0 {
	// 				// myPoolPending := self.config.CrossShardPool.GetPendingBlockHashes()
	// 				// for shardID, peerPoolBlks := range shardState.State.CrossShardsPoolState {
	// 				// 	myShardPoolBlks, ok := myPoolPending[shardID]
	// 				// 	if ok {
	// 				// 		blksNeedToSync := GetDiffHashesOf(peerPoolBlks, myShardPoolBlks)
	// 				// 		for _, blkHash := range blksNeedToSync {
	// 				// 			go self.config.Server.PushMessageGetShardToBeacon(shardID, blkHash)
	// 				// 		}
	// 				// 	} else {
	// 				// 		// sync all blks of this shard
	// 				// 		for _, blkHash := range peerPoolBlks {
	// 				// 			go self.config.Server.PushMessageGetShardToBeacon(shardID, blkHash)
	// 				// 		}
	// 				// 	}
	// 				// }
	// 			}
	// 		case newBlk := <-newShardBlkCh:
	// 			fmt.Println("Shard block received")
	// 			if self.BestState.Shard[shardID].ShardHeight < newBlk.Header.Height {
	// 				blkHash := newBlk.Header.Hash()
	// 				err := cashec.ValidateDataB58(newBlk.Header.Producer, newBlk.ProducerSig, blkHash.GetBytes())
	// 				if err != nil {
	// 					Logger.log.Error(err)
	// 					continue
	// 				} else {
	// 					if self.BestState.Shard[shardID].ShardHeight == newBlk.Header.Height-1 {
	// 						err = self.InsertShardBlock(newBlk)
	// 						if err != nil {
	// 							Logger.log.Error(err)
	// 							continue
	// 						}
	// 					} else {
	// 						self.config.NodeShardPool.PushBlock(*newBlk)
	// 					}
	// 				}
	// 			}
	// 		default:
	// 			time.Sleep(getStateWaitTime * time.Second)
	// 			self.config.Server.PushMessageGetShardState(shardID)
	// 			if self.knownChainState.Shards[shardID].Height > self.BestState.Shard[shardID].ShardHeight {
	// 				needToSync := self.knownChainState.Beacon.Height - self.BestState.Beacon.BeaconHeight
	// 				for offset := uint64(0); offset <= needToSync; offset++ {
	// 					blks, err := self.config.NodeShardPool.GetBlocks(shardID, self.BestState.Shard[shardID].ShardHeight+1)
	// 					if err != nil {
	// 						Logger.log.Error(err)
	// 						continue
	// 					}
	// 					for _, newBlk := range blks {
	// 						err = self.InsertShardBlock(&newBlk)
	// 						if err != nil {
	// 							Logger.log.Error(err)
	// 							continue
	// 						}
	// 					}
	// 				}
	// 			}
	// 		}
	// 	}
	// }(shardID)

	return nil
}

func (self *BlockChain) StopSyncShard(shardID byte) {
	self.syncStatus.Lock()
	defer self.syncStatus.Unlock()
	if _, ok := self.syncStatus.Shard[shardID]; ok {
		delete(self.syncStatus.Shard, shardID)
	}
}

func (self *BlockChain) GetCurrentSyncShards() []byte {
	self.syncStatus.Lock()
	defer self.syncStatus.Unlock()
	var currentSyncShards []byte
	for shardID, _ := range self.syncStatus.Shard {
		currentSyncShards = append(currentSyncShards, shardID)
	}
	return currentSyncShards
}

func (self *BlockChain) StopSync() error {
	close(self.cQuitSync)
	return nil
}

//GetDiffHashesOf Get unique hashes of 1st slice compare to 2nd slice
func GetDiffHashesOf(slice1 []common.Hash, slice2 []common.Hash) []common.Hash {
	var diff []common.Hash

	for _, s1 := range slice1 {
		found := false
		for _, s2 := range slice2 {
			if s1 == s2 {
				found = true
				break
			}
		}
		if !found {
			diff = append(diff, s1)
		}
	}

	return diff
}

func (self *BlockChain) SyncBlkBeacon(byHash bool, getFromPool bool, blksHash []common.Hash, from uint64, to uint64, peerID libp2p.ID) {
	if byHash {
		//Sync block by hash
		var blksNeedToGet []common.Hash
		var blksSyncByHash *map[string]peerSyncTimestamp
		tempInterface, ok := self.syncStatus.CurrentlySyncBeaconBlk.Load("byhash")
		if ok {
			blksSyncByHash = tempInterface.(*map[string]peerSyncTimestamp)
			for _, blkHash := range blksHash {
				if timeStamp, ok := (*blksSyncByHash)[blkHash.String()]; ok {
					if time.Since(time.Unix(timeStamp.Time, 0)) > defaultMaxBlockSyncTime {
						(*blksSyncByHash)[blkHash.String()] = peerSyncTimestamp{
							Time:   time.Now().Unix(),
							PeerID: peerID,
						}
						blksNeedToGet = append(blksNeedToGet, blkHash)
					}
				} else {
					(*blksSyncByHash)[blkHash.String()] = peerSyncTimestamp{
						Time:   time.Now().Unix(),
						PeerID: peerID,
					}
					blksNeedToGet = append(blksNeedToGet, blkHash)
				}
			}
		} else {
			blksSyncByHash = &map[string]peerSyncTimestamp{}
			for _, blkHash := range blksHash {
				(*blksSyncByHash)[blkHash.String()] = peerSyncTimestamp{
					Time:   time.Now().Unix(),
					PeerID: peerID,
				}
			}
			blksNeedToGet = append(blksNeedToGet, blksHash...)
		}

		if len(blksNeedToGet) > 0 {
			go self.config.Server.PushMessageGetBlockBeaconByHash(blksNeedToGet, getFromPool, peerID)
		}
	} else {
		//Sync by height
		var blkBatchsNeedToGet map[uint64]uint64
		blkBatchsNeedToGet = make(map[uint64]uint64)
		var blksSyncByHeight *map[uint64]peerSyncTimestamp
		tempInterface, ok := self.syncStatus.CurrentlySyncBeaconBlk.Load("byheight")
		if ok {
			blksSyncByHeight = tempInterface.(*map[uint64]peerSyncTimestamp)
			latestBatchBegin := uint64(0)
			for blkHeight := from; blkHeight <= to; blkHeight++ {
				if timeStamp, ok := (*blksSyncByHeight)[blkHeight]; ok {
					if time.Since(time.Unix(timeStamp.Time, 0)) > defaultMaxBlockSyncTime {
						(*blksSyncByHeight)[blkHeight] = peerSyncTimestamp{
							Time:   time.Now().Unix(),
							PeerID: peerID,
						}
						if latestBatchEnd, ok := blkBatchsNeedToGet[latestBatchBegin]; !ok {
							blkBatchsNeedToGet[blkHeight] = blkHeight
							latestBatchBegin = blkHeight
						} else {
							if latestBatchEnd+1 == blkHeight {
								blkBatchsNeedToGet[latestBatchBegin] = blkHeight
							} else {
								blkBatchsNeedToGet[blkHeight] = blkHeight
								latestBatchBegin = blkHeight
							}
						}
					}
				} else {
					(*blksSyncByHeight)[blkHeight] = peerSyncTimestamp{
						Time:   time.Now().Unix(),
						PeerID: peerID,
					}
					if latestBatchEnd, ok := blkBatchsNeedToGet[latestBatchBegin]; !ok {
						blkBatchsNeedToGet[blkHeight] = blkHeight
						latestBatchBegin = blkHeight
					} else {
						if latestBatchEnd+1 == blkHeight {
							blkBatchsNeedToGet[latestBatchBegin] = blkHeight
						} else {
							blkBatchsNeedToGet[blkHeight] = blkHeight
							latestBatchBegin = blkHeight
						}
					}
				}
			}
		} else {
			blksSyncByHeight = &map[uint64]peerSyncTimestamp{}
			for blkHeight := from; blkHeight <= to; blkHeight++ {
				(*blksSyncByHeight)[blkHeight] = peerSyncTimestamp{
					Time:   time.Now().Unix(),
					PeerID: peerID,
				}
			}
			blkBatchsNeedToGet[from] = to
		}
		if len(blkBatchsNeedToGet) > 0 {
			for fromHeight, toHeight := range blkBatchsNeedToGet {
				go self.config.Server.PushMessageGetBlockBeaconByHeight(fromHeight, toHeight, peerID)
			}
		}
	}
}

func (self *BlockChain) SyncBlkShard(byHash bool, getFromPool bool, shardID byte, blkHash []common.Hash, from uint64, to uint64, peerID libp2p.ID) {
	// if byHash {
	// 	var blkSyncByHash *map[string]int64
	// 	// blkSyncByHash = make(map[string]int64)
	// 	tempInterface, ok := self.syncStatus.CurrentlySyncShardBlk.Load("0")
	// 	blkSyncByHash = tempInterface.(*map[string]int64)
	// 	if ok {
	// 		if time, ok := (*blkSyncByHash)[blkHash.String()]; ok {
	// 			_ = time
	// 		} else {
	// 			go self.config.Server.PushMessageGetBlockShardByHash(shardID, blkHash, getFromPool, peerID)
	// 		}
	// 	} else {

	// 	}
	// } else {

	// }
}

func (self *BlockChain) SyncBlkShardToBeacon(byHash bool, getFromPool bool, blkHash []common.Hash, from uint64, to uint64, peerID libp2p.ID) {

}

func (self *BlockChain) SyncBlkCrossShard(getFromPool bool, blkHash common.Hash, fromShard uint64, toShard uint64, peerID libp2p.ID) {

}

func getBlkNeedToGetByHash() {

}

func getBlkNeedToGetByHeight() {

}
