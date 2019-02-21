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

func (blockchain *BlockChain) StartSyncBlk() {
	blockchain.knownChainState.Beacon.Height = blockchain.BestState.Beacon.BeaconHeight
	if blockchain.syncStatus.Beacon {
		return
	}
	blockchain.syncStatus.Beacon = true

	for _, shardID := range blockchain.config.RelayShards {
		blockchain.SyncShard(shardID)
	}
	go func() {
		for {
			select {
			case <-blockchain.cQuitSync:
				return
			case <-time.Tick(defaultBroadcastStateTime):
				blockchain.InsertBlockFromPool()
				go blockchain.config.Server.BoardcastNodeState()
			}
		}
	}()
	for {
		select {
		case <-blockchain.cQuitSync:
			return
		case <-time.Tick(defaultProcessPeerStateTime):
			blockchain.InsertBlockFromPool()
			blockchain.syncStatus.Lock()
			blockchain.syncStatus.PeersStateLock.Lock()
			userRole, userShardID := blockchain.BestState.Beacon.GetPubkeyRole(blockchain.config.UserKeySet.GetPublicKeyB58(), blockchain.BestState.Beacon.BestBlock.Header.Round)
			userShardRole := blockchain.BestState.Shard[userShardID].GetPubkeyRole(blockchain.config.UserKeySet.GetPublicKeyB58(), blockchain.BestState.Shard[userShardID].BestBlock.Header.Round)
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
			for peerID, peerState := range blockchain.syncStatus.PeersState {
				if peerState.Beacon.Height >= blockchain.BestState.Beacon.BeaconHeight {
					if RCS.ClosestBeaconState.Height == 0 {
						RCS.ClosestBeaconState = *peerState.Beacon
					} else {
						if peerState.Beacon.Height < RCS.ClosestBeaconState.Height {
							RCS.ClosestBeaconState = *peerState.Beacon
						}
					}
					for shardID := range blockchain.syncStatus.Shards {
						if shardState, ok := peerState.Shard[shardID]; ok {
							if shardState.Height > blockchain.BestState.Shard[shardID].ShardHeight && shardState.Height >= blockchain.BestState.Beacon.BestShardHeight[shardID] {
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
					switch blockchain.config.NodeMode {
					case common.NODEMODE_AUTO:
						switch userRole {
						case common.PROPOSER_ROLE, common.VALIDATOR_ROLE:
							if peerState.ShardToBeaconPool != nil {
								for shardID, blksHash := range *peerState.ShardToBeaconPool {
									if _, ok := RCS.ShardToBeaconBlks[shardID]; !ok {
										RCS.ShardToBeaconBlks[shardID] = make(map[libp2p.ID][]common.Hash)
									}
									RCS.ShardToBeaconBlks[shardID][peerID] = blksHash
								}
							}
							for shardID := byte(0); shardID < common.MAX_SHARD_NUMBER; shardID++ {
								if shardState, ok := peerState.Shard[shardID]; ok {
									if shardState.Height > blockchain.BestState.Beacon.BestShardHeight[shardID] {
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
						case common.SHARD_ROLE:
							if userShardRole == common.PROPOSER_ROLE || userShardRole == common.VALIDATOR_ROLE {
								if pool, ok := peerState.CrossShardPool[userShardID]; ok {
									for shardID, blks := range *pool {
										if _, ok := RCS.CrossShardBlks[shardID]; !ok {
											RCS.CrossShardBlks[shardID] = make(map[libp2p.ID][]common.Hash)
										}
										RCS.CrossShardBlks[shardID][peerID] = blks
									}
								}
							}
						}
					case common.NODEMODE_BEACON:
						if userRole == common.PROPOSER_ROLE || userRole == common.VALIDATOR_ROLE {
							if peerState.ShardToBeaconPool != nil {
								for shardID, blksHash := range *peerState.ShardToBeaconPool {
									if _, ok := RCS.ShardToBeaconBlks[shardID]; !ok {
										RCS.ShardToBeaconBlks[shardID] = make(map[libp2p.ID][]common.Hash)
									}
									RCS.ShardToBeaconBlks[shardID][peerID] = blksHash
								}
							}
							for shardID := byte(0); shardID < common.MAX_SHARD_NUMBER; shardID++ {
								if shardState, ok := peerState.Shard[shardID]; ok {
									if shardState.Height > blockchain.BestState.Beacon.BestShardHeight[shardID] {
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
						}
					case common.NODEMODE_SHARD:
						if userShardRole == common.PROPOSER_ROLE || userShardRole == common.VALIDATOR_ROLE {
							if pool, ok := peerState.CrossShardPool[userShardID]; ok {
								for shardID, blks := range *pool {
									if _, ok := RCS.CrossShardBlks[shardID]; !ok {
										RCS.CrossShardBlks[shardID] = make(map[libp2p.ID][]common.Hash)
									}
									RCS.CrossShardBlks[shardID][peerID] = blks
								}
							}
						}
					}
				}
			}
			currentBcnReqHeight := blockchain.BestState.Beacon.BeaconHeight + 1
			for peerID := range blockchain.syncStatus.PeersState {
				if currentBcnReqHeight+defaultMaxBlkReqPerPeer-1 >= RCS.ClosestBeaconState.Height {
					blockchain.SyncBlkBeacon(false, false, nil, currentBcnReqHeight, RCS.ClosestBeaconState.Height, peerID)
				} else {
					blockchain.SyncBlkBeacon(false, false, nil, currentBcnReqHeight, currentBcnReqHeight+defaultMaxBlkReqPerPeer-1, peerID)
					currentBcnReqHeight += defaultMaxBlkReqPerPeer - 1
				}
			}

			switch blockchain.config.NodeMode {
			case common.NODEMODE_AUTO:
				switch userRole {
				case common.PROPOSER_ROLE, common.VALIDATOR_ROLE:
					for shardID, peer := range RCS.ShardToBeaconBlks {
						for peerID, blks := range peer {
							blockchain.SyncBlkShardToBeacon(shardID, true, true, blks, 0, 0, peerID)
						}
					}
					for shardID := byte(0); shardID < common.MAX_SHARD_NUMBER; shardID++ {
						if blockchain.BestState.Beacon.BestShardHeight[shardID] < RCS.ClosestShardsState[shardID].Height {
							currentShardReqHeight := blockchain.BestState.Beacon.BestShardHeight[shardID] + 1
							for peerID, peerState := range blockchain.syncStatus.PeersState {
								if _, ok := peerState.Shard[shardID]; ok {
									if currentShardReqHeight+defaultMaxBlkReqPerPeer-1 >= RCS.ClosestShardsState[shardID].Height {
										blockchain.SyncBlkShardToBeacon(shardID, false, false, nil, currentShardReqHeight, RCS.ClosestShardsState[shardID].Height, peerID)
									} else {
										blockchain.SyncBlkShardToBeacon(shardID, false, false, nil, currentShardReqHeight, currentShardReqHeight+defaultMaxBlkReqPerPeer-1, peerID)
										currentShardReqHeight += defaultMaxBlkReqPerPeer - 1
									}
								}
							}
						}
					}
				case common.SHARD_ROLE:
					if _, ok := blockchain.syncStatus.Shards[userShardID]; !ok {
						blockchain.syncStatus.Shards[userShardID] = struct{}{}
					}
					if userShardRole == common.PROPOSER_ROLE || userShardRole == common.VALIDATOR_ROLE {
						for shardID, peer := range RCS.CrossShardBlks {
							for peerID, blks := range peer {
								blockchain.SyncBlkCrossShard(true, blks, shardID, userShardID, peerID)
							}
						}
					}
				}
			case common.NODEMODE_BEACON:
				if userRole == common.PROPOSER_ROLE || userRole == common.VALIDATOR_ROLE {
					for shardID, peer := range RCS.ShardToBeaconBlks {
						for peerID, blks := range peer {
							blockchain.SyncBlkShardToBeacon(shardID, true, true, blks, 0, 0, peerID)
						}
					}
					for shardID := byte(0); shardID < common.MAX_SHARD_NUMBER; shardID++ {
						if blockchain.BestState.Beacon.BestShardHeight[shardID] < RCS.ClosestShardsState[shardID].Height {
							currentShardReqHeight := blockchain.BestState.Beacon.BestShardHeight[shardID] + 1
							for peerID, peerState := range blockchain.syncStatus.PeersState {
								if shardState, ok := peerState.Shard[shardID]; ok && shardState.Height > RCS.ClosestShardsState[shardID].Height {
									if currentShardReqHeight+defaultMaxBlkReqPerPeer-1 >= RCS.ClosestShardsState[shardID].Height {
										blockchain.SyncBlkShardToBeacon(shardID, false, false, nil, currentShardReqHeight, RCS.ClosestShardsState[shardID].Height, peerID)
									} else {
										blockchain.SyncBlkShardToBeacon(shardID, false, false, nil, currentShardReqHeight, currentShardReqHeight+defaultMaxBlkReqPerPeer-1, peerID)
										currentShardReqHeight += defaultMaxBlkReqPerPeer - 1
									}
								}
							}
						}
					}
				}
			case common.NODEMODE_SHARD:
				if _, ok := blockchain.syncStatus.Shards[userShardID]; !ok {
					blockchain.syncStatus.Shards[userShardID] = struct{}{}
				}
				if userShardRole == common.PROPOSER_ROLE || userShardRole == common.VALIDATOR_ROLE {
					for shardID, peer := range RCS.CrossShardBlks {
						for peerID, blks := range peer {
							blockchain.SyncBlkCrossShard(true, blks, shardID, userShardID, peerID)
						}
					}
				}
			}

			for shardID := range blockchain.syncStatus.Shards {
				currentShardReqHeight := blockchain.BestState.Shard[shardID].ShardHeight + 1
				for peerID := range blockchain.syncStatus.PeersState {
					if shardState, ok := blockchain.syncStatus.PeersState[peerID].Shard[shardID]; ok {
						if shardState.Height >= currentShardReqHeight {
							if currentShardReqHeight+defaultMaxBlkReqPerPeer-1 >= RCS.ClosestShardsState[shardID].Height {
								blockchain.SyncBlkShard(shardID, false, false, nil, currentShardReqHeight, RCS.ClosestShardsState[shardID].Height, peerID)
							} else {
								blockchain.SyncBlkShard(shardID, false, false, nil, currentShardReqHeight, currentShardReqHeight+defaultMaxBlkReqPerPeer-1, peerID)
								currentShardReqHeight += defaultMaxBlkReqPerPeer - 1
							}
						}
					}
				}
			}

			blockchain.syncStatus.PeersState = make(map[libp2p.ID]*peerState)
			blockchain.syncStatus.Unlock()
			blockchain.syncStatus.PeersStateLock.Unlock()
		}
	}
}

func (blockchain *BlockChain) SyncShard(shardID byte) error {
	blockchain.syncStatus.Lock()
	defer blockchain.syncStatus.Unlock()
	if _, ok := blockchain.syncStatus.Shards[shardID]; ok {
		return errors.New("Shard " + fmt.Sprintf("%d", shardID) + " synchronzation is already started")
	}
	blockchain.syncStatus.Shards[shardID] = struct{}{}
	return nil
}

func (blockchain *BlockChain) StopSyncUnnecessaryShard() {
	for shardID := byte(0); shardID < common.MAX_SHARD_NUMBER; shardID++ {
		blockchain.StopSyncShard(shardID)
	}
}

func (blockchain *BlockChain) StopSyncShard(shardID byte) error {
	blockchain.syncStatus.Lock()
	defer blockchain.syncStatus.Unlock()
	if blockchain.config.NodeMode == common.NODEMODE_AUTO || blockchain.config.NodeMode == common.NODEMODE_SHARD {
		userRole, userShardID := blockchain.BestState.Beacon.GetPubkeyRole(blockchain.config.UserKeySet.GetPublicKeyB58(), blockchain.BestState.Beacon.BestBlock.Header.Round)
		if userRole == "shard" && shardID == userShardID {
			return errors.New("Shard " + fmt.Sprintf("%d", shardID) + " synchronzation can't be stopped")
		}
	}
	if _, ok := blockchain.syncStatus.Shards[shardID]; ok {
		if common.IndexOfByte(shardID, blockchain.config.RelayShards) < 0 {
			delete(blockchain.syncStatus.Shards, shardID)
			fmt.Println("Shard " + fmt.Sprintf("%d", shardID) + " synchronzation stopped")
			return nil
		}
		return errors.New("Shard " + fmt.Sprintf("%d", shardID) + " synchronzation can't be stopped")
	}
	return errors.New("Shard " + fmt.Sprintf("%d", shardID) + " synchronzation is already stopped")
}

func (blockchain *BlockChain) GetCurrentSyncShards() []byte {
	blockchain.syncStatus.Lock()
	defer blockchain.syncStatus.Unlock()
	var currentSyncShards []byte
	for shardID := range blockchain.syncStatus.Shards {
		currentSyncShards = append(currentSyncShards, shardID)
	}
	return currentSyncShards
}

func (blockchain *BlockChain) StopSync() error {
	close(blockchain.cQuitSync)
	return nil
}

func (blockchain *BlockChain) ResetCurrentSyncRecord() {
	blockchain.syncStatus.Lock()
	defer blockchain.syncStatus.Unlock()
	blockchain.syncStatus.CurrentlySyncBeaconBlk = sync.Map{}
	blockchain.syncStatus.CurrentlySyncShardBlk = sync.Map{}
	blockchain.syncStatus.CurrentlySyncShardToBeaconBlk = sync.Map{}
	blockchain.syncStatus.CurrentlySyncCrossShardBlk = sync.Map{}
}

//SyncBlkBeacon Send a req to sync beacon block
/*
	- by Hash + blksHash: get by hash
	- from + to: get from main chain by height
	- GetFromPool: ignore mainchain, used only for hash
*/
func (blockchain *BlockChain) SyncBlkBeacon(byHash bool, getFromPool bool, blksHash []common.Hash, from uint64, to uint64, peerID libp2p.ID) {
	if byHash {
		//Sync block by hash
		tempInterface, init := blockchain.syncStatus.CurrentlySyncBeaconBlk.Load(SyncByHashKey)
		blksNeedToGet, blksSyncByHash := getBlkNeedToGetByHash(blksHash, tempInterface, init, peerID)
		if len(blksNeedToGet) > 0 {
			go blockchain.config.Server.PushMessageGetBlockBeaconByHash(blksNeedToGet, getFromPool, peerID)
		}
		blockchain.syncStatus.CurrentlySyncBeaconBlk.Store(SyncByHashKey, blksSyncByHash)
	} else {
		//Sync by height
		tempInterface, init := blockchain.syncStatus.CurrentlySyncBeaconBlk.Load(SyncByHeightKey)
		blkBatchsNeedToGet, blksSyncByHeight := getBlkNeedToGetByHeight(from, to, tempInterface, init, peerID)
		if len(blkBatchsNeedToGet) > 0 {
			for fromHeight, toHeight := range blkBatchsNeedToGet {
				go blockchain.config.Server.PushMessageGetBlockBeaconByHeight(fromHeight, toHeight, peerID)
			}
		}
		blockchain.syncStatus.CurrentlySyncBeaconBlk.Store(SyncByHeightKey, blksSyncByHeight)
	}
}

//SyncBlkShard Send a req to sync shard block
/*
	- by Hash + blksHash: get by hash
	- from + to: get from main chain by height
	- GetFromPool: ignore mainchain, used only for hash
*/
func (blockchain *BlockChain) SyncBlkShard(shardID byte, byHash bool, getFromPool bool, blksHash []common.Hash, from uint64, to uint64, peerID libp2p.ID) {
	if byHash {
		//Sync block by hash
		tempInterface, init := blockchain.syncStatus.CurrentlySyncShardBlk.Load(SyncByHashKey + fmt.Sprint(shardID))
		blksNeedToGet, blksSyncByHash := getBlkNeedToGetByHash(blksHash, tempInterface, init, peerID)
		if len(blksNeedToGet) > 0 {
			go blockchain.config.Server.PushMessageGetBlockShardByHash(shardID, blksNeedToGet, getFromPool, peerID)
		}
		blockchain.syncStatus.CurrentlySyncShardBlk.Store(SyncByHashKey+fmt.Sprint(shardID), blksSyncByHash)
	} else {
		//Sync by height
		tempInterface, init := blockchain.syncStatus.CurrentlySyncShardBlk.Load(SyncByHeightKey + fmt.Sprint(shardID))
		blkBatchsNeedToGet, blksSyncByHeight := getBlkNeedToGetByHeight(from, to, tempInterface, init, peerID)
		if len(blkBatchsNeedToGet) > 0 {
			for fromHeight, toHeight := range blkBatchsNeedToGet {
				go blockchain.config.Server.PushMessageGetBlockShardByHeight(shardID, fromHeight, toHeight, peerID)
			}
		}
		blockchain.syncStatus.CurrentlySyncShardBlk.Store(SyncByHeightKey+fmt.Sprint(shardID), blksSyncByHeight)
	}
}

//SyncBlkShardToBeacon Send a req to sync shardToBeacon block
/*
	- by Hash + blksHash: get by hash
	- from + to: get from main chain by height
	- GetFromPool: ignore mainchain, used only for hash
*/
func (blockchain *BlockChain) SyncBlkShardToBeacon(shardID byte, byHash bool, getFromPool bool, blksHash []common.Hash, from uint64, to uint64, peerID libp2p.ID) {
	fmt.Println()
	fmt.Println()
	fmt.Println("SyncShardToBeacon", shardID, from, to)
	fmt.Println()
	fmt.Println()
	if byHash {
		//Sync block by hash
		tempInterface, init := blockchain.syncStatus.CurrentlySyncShardToBeaconBlk.Load(SyncByHashKey + fmt.Sprint(shardID))
		blksNeedToGet, blksSyncByHash := getBlkNeedToGetByHash(blksHash, tempInterface, init, peerID)
		if len(blksNeedToGet) > 0 {
			go blockchain.config.Server.PushMessageGetBlockShardToBeaconByHash(shardID, blksNeedToGet, getFromPool, peerID)
		}
		blockchain.syncStatus.CurrentlySyncShardToBeaconBlk.Store(SyncByHashKey+fmt.Sprint(shardID), blksSyncByHash)
	} else {
		//Sync by height
		tempInterface, init := blockchain.syncStatus.CurrentlySyncShardToBeaconBlk.Load(SyncByHeightKey + fmt.Sprint(shardID))
		blkBatchsNeedToGet, blksSyncByHeight := getBlkNeedToGetByHeight(from, to, tempInterface, init, peerID)
		if len(blkBatchsNeedToGet) > 0 {
			for fromHeight, toHeight := range blkBatchsNeedToGet {
				go blockchain.config.Server.PushMessageGetBlockShardToBeaconByHeight(shardID, fromHeight, toHeight, peerID)
			}
		}
		blockchain.syncStatus.CurrentlySyncShardToBeaconBlk.Store(SyncByHeightKey+fmt.Sprint(shardID), blksSyncByHeight)
	}
}

//SyncBlkCrossShard Send a req to sync crossShard block
func (blockchain *BlockChain) SyncBlkCrossShard(getFromPool bool, blksHash []common.Hash, fromShard byte, toShard byte, peerID libp2p.ID) {
	tempInterface, init := blockchain.syncStatus.CurrentlySyncCrossShardBlk.Load(SyncByHashKey)
	blksNeedToGet, blksSyncByHash := getBlkNeedToGetByHash(blksHash, tempInterface, init, peerID)
	if len(blksNeedToGet) > 0 {
		go blockchain.config.Server.PushMessageGetBlockCrossShardByHash(fromShard, toShard, blksNeedToGet, getFromPool, peerID)
	}
	blockchain.syncStatus.CurrentlySyncCrossShardBlk.Store(SyncByHashKey, blksSyncByHash)
}

func (blockchain *BlockChain) InsertBlockFromPool() {
	blks, err := blockchain.config.NodeBeaconPool.GetBlocks(blockchain.BestState.Beacon.BeaconHeight + 1)
	if err != nil {
		Logger.log.Error(err)
	} else {
		for idx, newBlk := range blks {
			err = blockchain.InsertBeaconBlock(&newBlk, false)
			if err != nil {
				Logger.log.Error(err)
				for idx2 := idx; idx2 < len(blks); idx2++ {
					blockchain.config.NodeBeaconPool.PushBlock(blks[idx2])
				}
				break
			}
		}
	}

	for shardID := range blockchain.syncStatus.Shards {
		blks, err := blockchain.config.NodeShardPool.GetBlocks(shardID, blockchain.BestState.Shard[shardID].ShardHeight+1)
		if err != nil {
			Logger.log.Error(err)
			continue
		}
		for idx, newBlk := range blks {
			err = blockchain.InsertShardBlock(&newBlk)
			if err != nil {
				Logger.log.Error(err)
				if idx < len(blks)-1 {
					for idx2 := idx; idx2 < len(blks); idx2++ {
						blockchain.config.NodeShardPool.PushBlock(blks[idx2])
					}
					break
				}
			}
		}
	}
}
