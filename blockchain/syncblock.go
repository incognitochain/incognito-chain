package blockchain

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/constant-money/constant-chain/common"
	libp2p "github.com/libp2p/go-libp2p-peer"
	"github.com/patrickmn/go-cache"
)

type peerState struct {
	Shard             map[byte]*ChainState
	Beacon            *ChainState
	ShardToBeaconPool *map[byte][]uint64
	CrossShardPool    map[byte]*map[byte][]uint64
	Peer              libp2p.ID
}

type ChainState struct {
	Height        uint64
	BlockHash     common.Hash
	BestStateHash common.Hash
}
type reportedChainState struct {
	ClosestBeaconState ChainState
	ClosestShardsState map[byte]ChainState
	ShardToBeaconBlks  map[byte]map[libp2p.ID][]uint64
	CrossShardBlks     map[byte]map[libp2p.ID][]uint64
}

func (blockchain *BlockChain) StartSyncBlk() {
	// blockchain.knownChainState.Beacon.Height = blockchain.BestState.Beacon.BeaconHeight
	if blockchain.syncStatus.Beacon {
		return
	}
	blockchain.syncStatus.Beacon = true

	blockchain.syncStatus.CurrentlySyncBeaconBlkByHash = cache.New(defaultMaxBlockSyncTime, defaultCacheCleanupTime)
	blockchain.syncStatus.CurrentlySyncBeaconBlkByHeight = cache.New(defaultMaxBlockSyncTime, defaultCacheCleanupTime)
	blockchain.syncStatus.CurrentlySyncShardBlkByHash = make(map[byte]*cache.Cache)
	blockchain.syncStatus.CurrentlySyncShardBlkByHeight = make(map[byte]*cache.Cache)
	blockchain.syncStatus.CurrentlySyncShardToBeaconBlkByHash = make(map[byte]*cache.Cache)
	blockchain.syncStatus.CurrentlySyncShardToBeaconBlkByHeight = make(map[byte]*cache.Cache)
	blockchain.syncStatus.CurrentlySyncCrossShardBlkByHash = make(map[byte]*cache.Cache)
	blockchain.syncStatus.CurrentlySyncCrossShardBlkByHeight = make(map[byte]*cache.Cache)
	blockchain.syncStatus.Lock()
	blockchain.startSyncRelayShards()
	blockchain.syncStatus.Unlock()

	broadcastTicker := time.NewTicker(defaultBroadcastStateTime)
	insertPoolTicker := time.NewTicker(time.Millisecond * 100)
	peersProcessTicker := time.NewTicker(defaultProcessPeerStateTime)

	defer func() {
		broadcastTicker.Stop()
		insertPoolTicker.Stop()
		peersProcessTicker.Stop()
	}()
	go func() {
		for {
			select {
			case <-blockchain.cQuitSync:
				return
			case <-broadcastTicker.C:
				blockchain.config.Server.BoardcastNodeState()
			}
		}
	}()

	go func() {
		for {
			select {
			case <-blockchain.cQuitSync:
				return
			case <-insertPoolTicker.C:
				blockchain.InsertBlockFromPool()
			}
		}
	}()

	for {
		select {
		case <-blockchain.cQuitSync:
			return
		case <-peersProcessTicker.C:
			blockchain.syncStatus.Lock()
			blockchain.syncStatus.PeersStateLock.Lock()

			var (
				userRole      string
				userShardID   byte
				userShardRole string
				userPK        string
			)
			if blockchain.config.UserKeySet != nil {
				userPK = blockchain.config.UserKeySet.GetPublicKeyB58()
				userRole, userShardID = blockchain.BestState.Beacon.GetPubkeyRole(userPK, blockchain.BestState.Beacon.BestBlock.Header.Round)
				blockchain.syncShard(userShardID)
				blockchain.stopSyncUnnecessaryShard()
				// blockchain.startSyncRelayShards()
				userShardRole = blockchain.BestState.Shard[userShardID].GetPubkeyRole(userPK, blockchain.BestState.Shard[userShardID].BestBlock.Header.Round)
			}

			RCS := reportedChainState{
				ClosestBeaconState: ChainState{
					Height: blockchain.BestState.Beacon.BeaconHeight,
				},
				ClosestShardsState: make(map[byte]ChainState),
				ShardToBeaconBlks:  make(map[byte]map[libp2p.ID][]uint64),
				CrossShardBlks:     make(map[byte]map[libp2p.ID][]uint64),
			}
			for shardID := range blockchain.syncStatus.Shards {
				RCS.ClosestShardsState[shardID] = ChainState{
					Height: blockchain.BestState.Shard[shardID].ShardHeight,
				}
			}
			for peerID, peerState := range blockchain.syncStatus.PeersState {
				for shardID := range blockchain.syncStatus.Shards {
					if shardState, ok := peerState.Shard[shardID]; ok {
						if shardState.Height >= GetBestStateBeacon().GetBestHeightOfShard(shardID) && shardState.Height > GetBestStateShard(shardID).ShardHeight {
							if RCS.ClosestShardsState[shardID].Height == blockchain.BestState.Shard[shardID].ShardHeight {
								RCS.ClosestShardsState[shardID] = *shardState
							} else {
								if shardState.Height < RCS.ClosestShardsState[shardID].Height {
									RCS.ClosestShardsState[shardID] = *shardState
								}
							}
						}
					}
				}
				if peerState.Beacon.Height >= blockchain.BestState.Beacon.BeaconHeight {
					if peerState.Beacon.Height > blockchain.BestState.Beacon.BeaconHeight {
						if RCS.ClosestBeaconState.Height == blockchain.BestState.Beacon.BeaconHeight {
							RCS.ClosestBeaconState = *peerState.Beacon
						}
						if peerState.Beacon.Height < RCS.ClosestBeaconState.Height {
							RCS.ClosestBeaconState = *peerState.Beacon
						}
					}

					// record pool state
					switch userRole {
					case common.PROPOSER_ROLE, common.VALIDATOR_ROLE:
						if blockchain.config.NodeMode == common.NODEMODE_AUTO || blockchain.config.NodeMode == common.NODEMODE_BEACON {
							if peerState.ShardToBeaconPool != nil {
								for shardID, blkHeights := range *peerState.ShardToBeaconPool {
									if _, ok := RCS.ShardToBeaconBlks[shardID]; !ok {
										RCS.ShardToBeaconBlks[shardID] = make(map[libp2p.ID][]uint64)
									}
									RCS.ShardToBeaconBlks[shardID][peerID] = blkHeights
								}
							}
							for shardID := byte(0); shardID < common.MAX_SHARD_NUMBER; shardID++ {
								if shardState, ok := peerState.Shard[shardID]; ok {
									if shardState.Height > GetBestStateBeacon().GetBestHeightOfShard(shardID) {
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
					case common.SHARD_ROLE:
						if (blockchain.config.NodeMode == common.NODEMODE_AUTO || blockchain.config.NodeMode == common.NODEMODE_SHARD) && (userShardRole == common.PROPOSER_ROLE || userShardRole == common.VALIDATOR_ROLE) {
							if pool, ok := peerState.CrossShardPool[userShardID]; ok {
								for shardID, blks := range *pool {
									if _, ok := RCS.CrossShardBlks[shardID]; !ok {
										RCS.CrossShardBlks[shardID] = make(map[libp2p.ID][]uint64)
									}
									RCS.CrossShardBlks[shardID][peerID] = blks
								}
							}
						}
					}
				}
			}
			if len(blockchain.syncStatus.PeersState) > 0 {
				if userRole != common.SHARD_ROLE && RCS.ClosestBeaconState.Height == blockchain.BestState.Beacon.BeaconHeight {
					blockchain.SetReadyState(false, 0, true)
				} else {
					blockchain.SetReadyState(false, 0, false)
				}

				if userRole == common.SHARD_ROLE && RCS.ClosestBeaconState.Height-1 <= blockchain.BestState.Beacon.BeaconHeight {
					if RCS.ClosestShardsState[userShardID].Height == GetBestStateShard(userShardID).ShardHeight && RCS.ClosestShardsState[userShardID].Height >= GetBestStateBeacon().BestShardHeight[userShardID] {
						blockchain.SetReadyState(false, 0, true)
						blockchain.SetReadyState(true, userShardID, true)
					} else {
						blockchain.SetReadyState(false, 0, false)
						blockchain.SetReadyState(true, userShardID, false)
					}
				}
			}

			currentBcnReqHeight := blockchain.BestState.Beacon.BeaconHeight + 1
			if RCS.ClosestBeaconState.Height-blockchain.BestState.Beacon.BeaconHeight > defaultMaxBlkReqPerTime {
				RCS.ClosestBeaconState.Height = blockchain.BestState.Beacon.BeaconHeight + defaultMaxBlkReqPerTime
			}
			for peerID := range blockchain.syncStatus.PeersState {
				if currentBcnReqHeight+defaultMaxBlkReqPerPeer-1 >= RCS.ClosestBeaconState.Height {
					//fmt.Println("SyncBlk1:", currentBcnReqHeight, RCS.ClosestBeaconState.Height)
					blockchain.SyncBlkBeacon(false, false, nil, currentBcnReqHeight, RCS.ClosestBeaconState.Height, peerID)
					break
				} else {
					//fmt.Println("SyncBlk2:", currentBcnReqHeight, currentBcnReqHeight+defaultMaxBlkReqPerPeer-1)
					blockchain.SyncBlkBeacon(false, false, nil, currentBcnReqHeight, currentBcnReqHeight+defaultMaxBlkReqPerPeer-1, peerID)
					currentBcnReqHeight += defaultMaxBlkReqPerPeer - 1
				}
			}
			// sync pool & shard
			if blockchain.IsReady(false, 0) {
				switch userRole {
				case common.PROPOSER_ROLE, common.VALIDATOR_ROLE:
					if blockchain.config.NodeMode == common.NODEMODE_AUTO || blockchain.config.NodeMode == common.NODEMODE_BEACON {
						for shardID, peer := range RCS.ShardToBeaconBlks {
							for peerID, blks := range peer {
								blockchain.SyncBlkShardToBeacon(shardID, false, true, true, nil, blks, 0, 0, peerID)
							}
						}
						for shardID := byte(0); shardID < common.MAX_SHARD_NUMBER; shardID++ {
							if GetBestStateBeacon().GetBestHeightOfShard(shardID) < RCS.ClosestShardsState[shardID].Height {
								currentShardReqHeight := GetBestStateBeacon().GetBestHeightOfShard(shardID) + 1
								for peerID, peerState := range blockchain.syncStatus.PeersState {
									if _, ok := peerState.Shard[shardID]; ok {
										if currentShardReqHeight+defaultMaxBlkReqPerPeer-1 >= RCS.ClosestShardsState[shardID].Height {
											blockchain.SyncBlkShardToBeacon(shardID, false, false, false, nil, nil, currentShardReqHeight, RCS.ClosestShardsState[shardID].Height, peerID)
										} else {
											blockchain.SyncBlkShardToBeacon(shardID, false, false, false, nil, nil, currentShardReqHeight, currentShardReqHeight+defaultMaxBlkReqPerPeer-1, peerID)
											currentShardReqHeight += defaultMaxBlkReqPerPeer - 1
										}
									}
								}
							}
						}
					}
				case common.SHARD_ROLE:
					if (blockchain.config.NodeMode == common.NODEMODE_AUTO || blockchain.config.NodeMode == common.NODEMODE_SHARD) && (userShardRole == common.PROPOSER_ROLE || userShardRole == common.VALIDATOR_ROLE) {
						if blockchain.IsReady(true, userShardID) {
							for shardID, peer := range RCS.CrossShardBlks {
								for peerID, blks := range peer {
									blockchain.SyncBlkCrossShard(true, false, nil, blks, shardID, userShardID, peerID)
								}
							}
						}
					}
				}
			}

			for shardID := range blockchain.syncStatus.Shards {
				currentShardReqHeight := blockchain.BestState.Shard[shardID].ShardHeight + 1
				if RCS.ClosestShardsState[shardID].Height-blockchain.BestState.Shard[shardID].ShardHeight > defaultMaxBlkReqPerTime {
					RCS.ClosestShardsState[shardID] = ChainState{
						Height: blockchain.BestState.Shard[shardID].ShardHeight + defaultMaxBlkReqPerTime,
					}
				}

				for peerID := range blockchain.syncStatus.PeersState {
					if shardState, ok := blockchain.syncStatus.PeersState[peerID].Shard[shardID]; ok {
						//fmt.Println("SyncShard 123 ", shardState.Height, shardID)
						if shardState.Height >= currentShardReqHeight {
							if currentShardReqHeight+defaultMaxBlkReqPerPeer-1 >= RCS.ClosestShardsState[shardID].Height {
								// fmt.Println("SyncShard 1234 ")
								blockchain.SyncBlkShard(shardID, false, false, nil, currentShardReqHeight, RCS.ClosestShardsState[shardID].Height, peerID)
								break
							} else {
								// fmt.Println("SyncShard 12345")
								blockchain.SyncBlkShard(shardID, false, false, nil, currentShardReqHeight, currentShardReqHeight+defaultMaxBlkReqPerPeer-1, peerID)
								currentShardReqHeight += defaultMaxBlkReqPerPeer - 1
							}
						}
					}
				}
			}
			userLayer := userRole
			switch userRole {
			case common.VALIDATOR_ROLE, common.PROPOSER_ROLE:
				userLayer = common.BEACON_ROLE
			}
			blockchain.config.Server.UpdateConsensusState(userLayer, userPK, nil, blockchain.BestState.Beacon.BeaconCommittee, blockchain.BestState.Beacon.ShardCommittee)

			blockchain.syncStatus.PeersState = make(map[libp2p.ID]*peerState)
			blockchain.syncStatus.Unlock()
			blockchain.syncStatus.PeersStateLock.Unlock()
		}
	}
}

func (blockchain *BlockChain) SyncShard(shardID byte) error {
	blockchain.syncStatus.Lock()
	defer blockchain.syncStatus.Unlock()
	return blockchain.syncShard(shardID)
}

func (blockchain *BlockChain) syncShard(shardID byte) error {
	if _, ok := blockchain.syncStatus.Shards[shardID]; ok {
		return errors.New("Shard " + fmt.Sprintf("%d", shardID) + " synchronzation is already started")
	}
	blockchain.syncStatus.Shards[shardID] = struct{}{}
	blockchain.syncStatus.CurrentlySyncShardBlkByHash[shardID] = cache.New(defaultMaxBlockSyncTime, defaultCacheCleanupTime)
	blockchain.syncStatus.CurrentlySyncShardBlkByHeight[shardID] = cache.New(defaultMaxBlockSyncTime, defaultCacheCleanupTime)

	return nil
}

func (blockchain *BlockChain) startSyncRelayShards() {
	for _, shardID := range blockchain.config.RelayShards {
		if shardID > byte(blockchain.BestState.Beacon.ActiveShards-1) {
			break
		}
		blockchain.syncShard(shardID)
	}
}
func (blockchain *BlockChain) StopSyncUnnecessaryShard() {
	blockchain.syncStatus.Lock()
	defer blockchain.syncStatus.Unlock()
	blockchain.stopSyncUnnecessaryShard()
}

func (blockchain *BlockChain) stopSyncUnnecessaryShard() {
	for shardID := byte(0); shardID < common.MAX_SHARD_NUMBER; shardID++ {
		if err := blockchain.stopSyncShard(shardID); err != nil {
			//Logger.log.Error(err)
		}
	}
}

func (blockchain *BlockChain) stopSyncShard(shardID byte) error {
	if blockchain.config.NodeMode == common.NODEMODE_AUTO || blockchain.config.NodeMode == common.NODEMODE_SHARD {
		userRole, userShardID := blockchain.BestState.Beacon.GetPubkeyRole(blockchain.config.UserKeySet.GetPublicKeyB58(), blockchain.BestState.Beacon.BestBlock.Header.Round)
		if userRole == common.SHARD_ROLE && shardID == userShardID {
			return errors.New("Shard " + fmt.Sprintf("%d", shardID) + " synchronzation can't be stopped")
		}
	}
	if _, ok := blockchain.syncStatus.Shards[shardID]; ok {
		if common.IndexOfByte(shardID, blockchain.config.RelayShards) < 0 {
			delete(blockchain.syncStatus.Shards, shardID)
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

//SyncBlkBeacon Send a req to sync beacon block
/*
	- by Hash + blksHash: get by hash
	- from + to: get from main chain by height
	- GetFromPool: ignore mainchain, used only for hash
*/
func (blockchain *BlockChain) SyncBlkBeacon(byHash bool, getFromPool bool, blksHash []common.Hash, from uint64, to uint64, peerID libp2p.ID) {
	if byHash {
		//Sync block by hash
		blockchain.syncStatus.CurrentlySyncBeaconBlkByHash.DeleteExpired()
		cacheItems := blockchain.syncStatus.CurrentlySyncBeaconBlkByHash.Items()
		blksNeedToGet := getBlkNeedToGetByHash(blksHash, cacheItems, peerID)
		if len(blksNeedToGet) > 0 {
			go blockchain.config.Server.PushMessageGetBlockBeaconByHash(blksNeedToGet, getFromPool, peerID)
		}
		for _, blkHash := range blksNeedToGet {
			blockchain.syncStatus.CurrentlySyncBeaconBlkByHash.Add(blkHash.String(), time.Now().Unix(), defaultMaxBlockSyncTime)
		}
	} else {
		//Sync by height
		blockchain.syncStatus.CurrentlySyncBeaconBlkByHeight.DeleteExpired()
		cacheItems := blockchain.syncStatus.CurrentlySyncBeaconBlkByHeight.Items()
		blkBatchsNeedToGet := getBlkNeedToGetByHeight(from, to, cacheItems, peerID)
		if len(blkBatchsNeedToGet) > 0 {
			for fromHeight, toHeight := range blkBatchsNeedToGet {
				go blockchain.config.Server.PushMessageGetBlockBeaconByHeight(fromHeight, toHeight, peerID)
			}
		}
		for fromHeight, toHeight := range blkBatchsNeedToGet {
			for height := fromHeight; height <= toHeight; height++ {
				blockchain.syncStatus.CurrentlySyncBeaconBlkByHeight.Add(fmt.Sprint(height), time.Now().Unix(), defaultMaxBlockSyncTime)
			}
		}
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
		if _, ok := blockchain.syncStatus.CurrentlySyncShardBlkByHash[shardID]; !ok {
			blockchain.syncStatus.CurrentlySyncShardBlkByHash[shardID] = cache.New(defaultMaxBlockSyncTime, defaultCacheCleanupTime)
		}
		blockchain.syncStatus.CurrentlySyncShardBlkByHash[shardID].DeleteExpired()
		cacheItems := blockchain.syncStatus.CurrentlySyncShardBlkByHash[shardID].Items()
		blksNeedToGet := getBlkNeedToGetByHash(blksHash, cacheItems, peerID)
		if len(blksNeedToGet) > 0 {
			go blockchain.config.Server.PushMessageGetBlockShardByHash(shardID, blksNeedToGet, getFromPool, peerID)
		}
		for _, blkHash := range blksNeedToGet {
			blockchain.syncStatus.CurrentlySyncShardBlkByHash[shardID].Add(blkHash.String(), time.Now().Unix(), defaultMaxBlockSyncTime)
		}
	} else {
		//Sync by height
		if _, ok := blockchain.syncStatus.CurrentlySyncShardBlkByHeight[shardID]; !ok {
			blockchain.syncStatus.CurrentlySyncShardBlkByHeight[shardID] = cache.New(defaultMaxBlockSyncTime, defaultCacheCleanupTime)
		}
		blockchain.syncStatus.CurrentlySyncShardBlkByHeight[shardID].DeleteExpired()
		cacheItems := blockchain.syncStatus.CurrentlySyncShardBlkByHeight[shardID].Items()
		blkBatchsNeedToGet := getBlkNeedToGetByHeight(from, to, cacheItems, peerID)
		//fmt.Println("SyncBlkShard", from, to, blkBatchsNeedToGet)
		if len(blkBatchsNeedToGet) > 0 {
			for fromHeight, toHeight := range blkBatchsNeedToGet {
				//fmt.Println("SyncBlkShard", shardID, fromHeight, toHeight, peerID)
				go blockchain.config.Server.PushMessageGetBlockShardByHeight(shardID, fromHeight, toHeight, peerID)
			}
		}
		for fromHeight, toHeight := range blkBatchsNeedToGet {
			for height := fromHeight; height <= toHeight; height++ {
				blockchain.syncStatus.CurrentlySyncShardBlkByHeight[shardID].Add(fmt.Sprint(height), time.Now().Unix(), defaultMaxBlockSyncTime)
			}
		}
	}
}

//SyncBlkShardToBeacon Send a req to sync shardToBeacon block
/*
	- by Hash + blksHash: get by hash
	- from + to: get from main chain by height
	- GetFromPool: ignore mainchain, used only for hash
*/
func (blockchain *BlockChain) SyncBlkShardToBeacon(shardID byte, byHash bool, bySpecificHeights bool, getFromPool bool, blksHash []common.Hash, blkHeights []uint64, from uint64, to uint64, peerID libp2p.ID) {
	if byHash {
		//Sync block by hash
		if _, ok := blockchain.syncStatus.CurrentlySyncShardToBeaconBlkByHash[shardID]; !ok {
			blockchain.syncStatus.CurrentlySyncShardToBeaconBlkByHash[shardID] = cache.New(defaultMaxBlockSyncTime, defaultCacheCleanupTime)
		}
		blockchain.syncStatus.CurrentlySyncShardToBeaconBlkByHash[shardID].DeleteExpired()
		cacheItems := blockchain.syncStatus.CurrentlySyncShardToBeaconBlkByHash[shardID].Items()
		blksNeedToGet := getBlkNeedToGetByHash(blksHash, cacheItems, peerID)
		if len(blksNeedToGet) > 0 {
			go blockchain.config.Server.PushMessageGetBlockShardToBeaconByHash(shardID, blksNeedToGet, getFromPool, peerID)
		}
		for _, blkHash := range blksNeedToGet {
			blockchain.syncStatus.CurrentlySyncShardToBeaconBlkByHash[shardID].Add(blkHash.String(), time.Now().Unix(), defaultMaxBlockSyncTime)
		}
	} else {
		//Sync by height
		if _, ok := blockchain.syncStatus.CurrentlySyncShardToBeaconBlkByHeight[shardID]; !ok {
			blockchain.syncStatus.CurrentlySyncShardToBeaconBlkByHeight[shardID] = cache.New(defaultMaxBlockSyncTime, defaultCacheCleanupTime)
		}
		blockchain.syncStatus.CurrentlySyncShardToBeaconBlkByHeight[shardID].DeleteExpired()
		cacheItems := blockchain.syncStatus.CurrentlySyncShardToBeaconBlkByHeight[shardID].Items()
		if bySpecificHeights {
			blksNeedToGet := getBlkNeedToGetBySpecificHeight(blkHeights, cacheItems, peerID)
			if len(blksNeedToGet) > 0 {
				go blockchain.config.Server.PushMessageGetBlockShardToBeaconBySpecificHeight(shardID, blksNeedToGet, getFromPool, peerID)
			}
			for _, blkHeight := range blksNeedToGet {
				blockchain.syncStatus.CurrentlySyncShardToBeaconBlkByHeight[shardID].Add(fmt.Sprint(blkHeight), time.Now().Unix(), defaultMaxBlockSyncTime)
			}
		} else {
			blkBatchsNeedToGet := getBlkNeedToGetByHeight(from, to, cacheItems, peerID)
			if len(blkBatchsNeedToGet) > 0 {
				for fromHeight, toHeight := range blkBatchsNeedToGet {
					go blockchain.config.Server.PushMessageGetBlockShardToBeaconByHeight(shardID, fromHeight, toHeight, peerID)
				}
			}
			for fromHeight, toHeight := range blkBatchsNeedToGet {
				for height := fromHeight; height <= toHeight; height++ {
					blockchain.syncStatus.CurrentlySyncShardToBeaconBlkByHeight[shardID].Add(fmt.Sprint(height), time.Now().Unix(), defaultMaxBlockSyncTime)
				}
			}
		}
	}
}

//SyncBlkCrossShard Send a req to sync crossShard block
/*
	From Shard: shard creates cross shard block
	To  Shard: shard receive cross shard block
*/
func (blockchain *BlockChain) SyncBlkCrossShard(getFromPool bool, byHash bool, blksHash []common.Hash, blksHeight []uint64, fromShard byte, toShard byte, peerID libp2p.ID) {
	Logger.log.Criticalf("Shard %+v request CrossShardBlock with Height %+v from shard %+v \n", fromShard, blksHeight, toShard)
	if byHash {
		if _, ok := blockchain.syncStatus.CurrentlySyncCrossShardBlkByHash[fromShard]; !ok {
			blockchain.syncStatus.CurrentlySyncCrossShardBlkByHash[fromShard] = cache.New(defaultMaxBlockSyncTime, defaultCacheCleanupTime)
		}
		blockchain.syncStatus.CurrentlySyncCrossShardBlkByHash[fromShard].DeleteExpired()
		cacheItems := blockchain.syncStatus.CurrentlySyncCrossShardBlkByHash[fromShard].Items()
		blksNeedToGet := getBlkNeedToGetByHash(blksHash, cacheItems, peerID)
		if len(blksNeedToGet) > 0 {
			go blockchain.config.Server.PushMessageGetBlockCrossShardByHash(fromShard, toShard, blksNeedToGet, getFromPool, peerID)
		}
		for _, blkHash := range blksNeedToGet {
			blockchain.syncStatus.CurrentlySyncCrossShardBlkByHash[fromShard].Add(blkHash.String(), time.Now().Unix(), defaultMaxBlockSyncTime)
		}
	} else {
		//Sync by specific heights
		var tempBlksHeight []uint64
		for _, value := range blksHeight {
			if value != 0 {
				tempBlksHeight = append(tempBlksHeight, value)
			}
		}
		blksHeight = tempBlksHeight
		if len(blksHeight) == 0 {
			return
		}
		if _, ok := blockchain.syncStatus.CurrentlySyncCrossShardBlkByHeight[fromShard]; !ok {
			blockchain.syncStatus.CurrentlySyncCrossShardBlkByHeight[fromShard] = cache.New(defaultMaxBlockSyncTime, defaultCacheCleanupTime)
		}
		blockchain.syncStatus.CurrentlySyncCrossShardBlkByHeight[fromShard].DeleteExpired()
		cacheItems := blockchain.syncStatus.CurrentlySyncCrossShardBlkByHeight[fromShard].Items()
		blksNeedToGet := getBlkNeedToGetBySpecificHeight(blksHeight, cacheItems, peerID)
		if len(blksNeedToGet) > 0 {
			go blockchain.config.Server.PushMessageGetBlockCrossShardBySpecificHeight(fromShard, toShard, blksNeedToGet, getFromPool, peerID)
		}
		for _, blkHeight := range blksNeedToGet {
			blockchain.syncStatus.CurrentlySyncCrossShardBlkByHeight[fromShard].Add(fmt.Sprint(blkHeight), time.Now().Unix(), defaultMaxBlockSyncTime)
		}
	}
}

var lasttime = time.Now()
var currentInsert = struct {
	Beacon sync.Mutex
	Shards map[byte]*sync.Mutex
}{
	Shards: make(map[byte]*sync.Mutex),
}

func (blockchain *BlockChain) InsertBlockFromPool() {
	// fmt.Println("InsertBlockFromPool")

	if time.Since(lasttime) >= 30*time.Millisecond {
		lasttime = time.Now()
	} else {
		return
	}

	go blockchain.InsertBeaconBlockFromPool()

	blockchain.syncStatus.Lock()
	for shardID := range blockchain.syncStatus.Shards {
		if _, ok := currentInsert.Shards[shardID]; !ok {
			currentInsert.Shards[shardID] = &sync.Mutex{}
		}
		go func(shardID byte) {
			blockchain.InsertShardBlockFromPool(shardID)
		}(shardID)

	}
	blockchain.syncStatus.Unlock()
}

func (blockchain *BlockChain) InsertBeaconBlockFromPool() {
	currentInsert.Beacon.Lock()
	defer currentInsert.Beacon.Unlock()
	blks := blockchain.config.BeaconPool.GetValidBlock()
	for _, newBlk := range blks {
		err := blockchain.InsertBeaconBlock(newBlk, false)
		if err != nil {
			Logger.log.Error(err)
			break
		}
	}
}

func (blockchain *BlockChain) InsertShardBlockFromPool(shardID byte) {

	currentInsert.Shards[shardID].Lock()
	defer currentInsert.Shards[shardID].Unlock()
	blks := blockchain.config.ShardPool[shardID].GetValidBlock()
	for _, newBlk := range blks {
		err := blockchain.InsertShardBlock(newBlk, false)
		if err != nil {
			Logger.log.Error(err)
			break
		}
	}
}
