package blockchain

import (
	"errors"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/incognitochain/incognito-chain/common"
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

type synker struct {
	Status struct {
		sync.Mutex
		Beacon            bool
		Shards            map[byte]struct{}
		CurrentlySyncBlks *cache.Cache
		IsLatest          struct {
			Beacon bool
			Shards map[byte]bool
			sync.Mutex
		}
	}
	States struct {
		PeersState   map[libp2p.ID]*peerState
		ClosestState struct {
			ClosestBeaconState uint64
			ClosestShardsState map[byte]uint64
			ShardToBeaconPool  map[byte]uint64
			CrossShardPool     map[byte]uint64
		}
		PoolsState struct {
			BeaconPool        []uint64
			ShardToBeaconPool map[byte][]uint64
			CrossShardPool    map[byte][]uint64
			ShardsPool        map[byte][]uint64
			sync.Mutex
		}
		sync.Mutex
	}
	blockchain *BlockChain
	cQuit      chan struct{}
}

func (synker *synker) Start() {
	if synker.Status.Beacon {
		return
	}
	synker.Status.Beacon = true
	synker.Status.CurrentlySyncBlks = cache.New(DefaultMaxBlockSyncTime, DefaultCacheCleanupTime)
	synker.Status.Shards = make(map[byte]struct{})
	synker.Status.IsLatest.Shards = make(map[byte]bool)
	synker.States.PeersState = make(map[libp2p.ID]*peerState)
	synker.States.ClosestState.ClosestShardsState = make(map[byte]uint64)
	synker.States.ClosestState.ShardToBeaconPool = make(map[byte]uint64)
	synker.States.ClosestState.CrossShardPool = make(map[byte]uint64)
	synker.States.PoolsState.ShardToBeaconPool = make(map[byte][]uint64)
	synker.States.PoolsState.CrossShardPool = make(map[byte][]uint64)
	synker.States.PoolsState.ShardsPool = make(map[byte][]uint64)
	synker.Status.Lock()
	synker.startSyncRelayShards()
	synker.Status.Unlock()

	broadcastTicker := time.NewTicker(DefaultBroadcastStateTime)
	insertPoolTicker := time.NewTicker(time.Millisecond * 100)
	updateStatesTicker := time.NewTicker(DefaultStateUpdateTime)
	defer func() {
		broadcastTicker.Stop()
		insertPoolTicker.Stop()
		updateStatesTicker.Stop()
	}()
	go func() {
		for {
			select {
			case <-synker.cQuit:
				return
			case <-broadcastTicker.C:
				synker.blockchain.config.Server.BoardcastNodeState()
			}
		}
	}()

	go func() {
		for {
			select {
			case <-synker.cQuit:
				return
			case <-insertPoolTicker.C:
				synker.InsertBlockFromPool()
			}
		}
	}()

	for {
		select {
		case <-synker.cQuit:
			return
		case <-updateStatesTicker.C:
			synker.UpdateState()
		}
	}
}

func (synker *synker) SyncShard(shardID byte) error {
	synker.Status.Lock()
	defer synker.Status.Unlock()
	return synker.syncShard(shardID)
}

func (synker *synker) syncShard(shardID byte) error {
	if _, ok := synker.Status.Shards[shardID]; ok {
		return errors.New("Shard " + fmt.Sprintf("%d", shardID) + " synchronzation is already started")
	}
	synker.Status.Shards[shardID] = struct{}{}
	return nil
}

func (synker *synker) startSyncRelayShards() {
	for _, shardID := range synker.blockchain.config.RelayShards {
		if shardID > byte(synker.blockchain.BestState.Beacon.ActiveShards-1) {
			break
		}
		synker.syncShard(shardID)
	}
}

func (synker *synker) StopSyncUnnecessaryShard() {
	synker.Status.Lock()
	defer synker.Status.Unlock()
	synker.stopSyncUnnecessaryShard()
}

func (synker *synker) stopSyncUnnecessaryShard() {
	for shardID := byte(0); shardID < common.MAX_SHARD_NUMBER; shardID++ {
		synker.stopSyncShard(shardID)
	}
}

func (synker *synker) stopSyncShard(shardID byte) error {
	if synker.blockchain.config.NodeMode == common.NODEMODE_AUTO || synker.blockchain.config.NodeMode == common.NODEMODE_SHARD {
		userRole, userShardID := synker.blockchain.BestState.Beacon.GetPubkeyRole(synker.blockchain.config.UserKeySet.GetPublicKeyB58(), synker.blockchain.BestState.Beacon.BestBlock.Header.Round)
		if userRole == common.SHARD_ROLE && shardID == userShardID {
			return errors.New("Shard " + fmt.Sprintf("%d", shardID) + " synchronzation can't be stopped")
		}
	}
	if _, ok := synker.Status.Shards[shardID]; ok {
		if common.IndexOfByte(shardID, synker.blockchain.config.RelayShards) < 0 {
			delete(synker.Status.Shards, shardID)
			return nil
		}
		return errors.New("Shard " + fmt.Sprintf("%d", shardID) + " synchronzation can't be stopped")
	}
	return errors.New("Shard " + fmt.Sprintf("%d", shardID) + " synchronzation is already stopped")
}

func (synker *synker) UpdateState() {
	synker.Status.Lock()
	synker.States.Lock()
	synker.GetPoolsState()
	synker.Status.CurrentlySyncBlks.DeleteExpired()

	var (
		userRole      string
		userShardID   byte
		userShardRole string
		userPK        string
	)
	if synker.blockchain.config.UserKeySet != nil {
		userPK = synker.blockchain.config.UserKeySet.GetPublicKeyB58()
		userRole, userShardID = synker.blockchain.BestState.Beacon.GetPubkeyRole(userPK, synker.blockchain.BestState.Beacon.BestBlock.Header.Round)
		synker.syncShard(userShardID)
		synker.stopSyncUnnecessaryShard()
		userShardRole = synker.blockchain.BestState.Shard[userShardID].GetPubkeyRole(userPK, synker.blockchain.BestState.Shard[userShardID].BestBlock.Header.Round)
	}

	synker.States.ClosestState.ClosestBeaconState = synker.blockchain.BestState.Beacon.BeaconHeight
	for shardID, beststate := range synker.blockchain.BestState.Shard {
		synker.States.ClosestState.ClosestShardsState[shardID] = beststate.ShardHeight
	}
	synker.States.ClosestState.ShardToBeaconPool = synker.blockchain.config.ShardToBeaconPool.GetLatestValidPendingBlockHeight()
	synker.States.ClosestState.CrossShardPool = synker.blockchain.config.CrossShardPool[userShardID].GetLatestValidBlockHeight()

	RCS := reportedChainState{
		ClosestBeaconState: ChainState{
			Height: synker.blockchain.BestState.Beacon.BeaconHeight,
		},
		ClosestShardsState: make(map[byte]ChainState),
		ShardToBeaconBlks:  make(map[byte]map[libp2p.ID][]uint64),
		CrossShardBlks:     make(map[byte]map[libp2p.ID][]uint64),
	}
	for shardID := range synker.Status.Shards {
		RCS.ClosestShardsState[shardID] = ChainState{
			Height: synker.blockchain.BestState.Shard[shardID].ShardHeight,
		}
	}

	for peerID, peerState := range synker.States.PeersState {
		for shardID := range synker.Status.Shards {
			if shardState, ok := peerState.Shard[shardID]; ok {
				if shardState.Height >= GetBestStateBeacon().GetBestHeightOfShard(shardID) && shardState.Height > GetBestStateShard(shardID).ShardHeight {
					if RCS.ClosestShardsState[shardID].Height == synker.blockchain.BestState.Shard[shardID].ShardHeight {
						RCS.ClosestShardsState[shardID] = *shardState
					} else {
						if shardState.Height < RCS.ClosestShardsState[shardID].Height {
							RCS.ClosestShardsState[shardID] = *shardState
						}
					}
				}
			}
		}
		if peerState.Beacon.Height >= synker.blockchain.BestState.Beacon.BeaconHeight {
			if peerState.Beacon.Height > synker.blockchain.BestState.Beacon.BeaconHeight {
				if RCS.ClosestBeaconState.Height == synker.blockchain.BestState.Beacon.BeaconHeight {
					RCS.ClosestBeaconState = *peerState.Beacon
				}
				if peerState.Beacon.Height < RCS.ClosestBeaconState.Height {
					RCS.ClosestBeaconState = *peerState.Beacon
				}
			}

			// record pool state
			switch userRole {
			case common.PROPOSER_ROLE, common.VALIDATOR_ROLE:
				if synker.blockchain.config.NodeMode == common.NODEMODE_AUTO || synker.blockchain.config.NodeMode == common.NODEMODE_BEACON {
					if peerState.ShardToBeaconPool != nil {
						for shardID, blkHeights := range *peerState.ShardToBeaconPool {
							if len(synker.States.PoolsState.ShardToBeaconPool[shardID]) > 0 {
								if _, ok := RCS.ShardToBeaconBlks[shardID]; !ok {
									RCS.ShardToBeaconBlks[shardID] = make(map[libp2p.ID][]uint64)
								}
								RCS.ShardToBeaconBlks[shardID][peerID] = blkHeights

								if len(blkHeights) > 0 && len(blkHeights) <= len(synker.States.PoolsState.ShardToBeaconPool[shardID]) {
									commonHeights := arrayCommonElements(blkHeights, synker.States.PoolsState.ShardToBeaconPool[shardID])
									sort.Slice(commonHeights, func(i, j int) bool { return blkHeights[i] < blkHeights[j] })
									if len(commonHeights) > 0 {
										for idx := len(commonHeights) - 1; idx == 0; idx-- {
											if idx == 0 {
												synker.States.ClosestState.ShardToBeaconPool[shardID] = synker.blockchain.BestState.Beacon.GetBestHeightOfShard(shardID)
											}
											if synker.States.ClosestState.ShardToBeaconPool[shardID] > commonHeights[idx] {
												synker.States.ClosestState.ShardToBeaconPool[shardID] = commonHeights[idx]
												break
											}
										}
									}
								}
							}
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
				if (synker.blockchain.config.NodeMode == common.NODEMODE_AUTO || synker.blockchain.config.NodeMode == common.NODEMODE_SHARD) && (userShardRole == common.PROPOSER_ROLE || userShardRole == common.VALIDATOR_ROLE) {
					if pool, ok := peerState.CrossShardPool[userShardID]; ok {
						for shardID, blkHeights := range *pool {
							if _, ok := RCS.CrossShardBlks[shardID]; !ok {
								RCS.CrossShardBlks[shardID] = make(map[libp2p.ID][]uint64)
							}
							RCS.CrossShardBlks[shardID][peerID] = blkHeights

							if len(blkHeights) > 0 && len(blkHeights) <= len(synker.States.PoolsState.CrossShardPool[shardID]) {
								commonHeights := arrayCommonElements(blkHeights, synker.States.PoolsState.CrossShardPool[shardID])
								sort.Slice(commonHeights, func(i, j int) bool { return blkHeights[i] < blkHeights[j] })
								if len(commonHeights) > 0 {
									for idx := len(commonHeights) - 1; idx < 0; idx-- {
										if synker.States.ClosestState.CrossShardPool[shardID] > commonHeights[idx] {
											synker.States.ClosestState.CrossShardPool[shardID] = commonHeights[idx]
										}
									}
								}
							}
						}
					}
				}
			}
		}
	}
	synker.States.ClosestState.ClosestBeaconState = RCS.ClosestBeaconState.Height
	for shardID, state := range RCS.ClosestShardsState {
		synker.States.ClosestState.ClosestShardsState[shardID] = state.Height
	}

	if len(synker.States.PeersState) > 0 {
		if userRole != common.SHARD_ROLE && RCS.ClosestBeaconState.Height == synker.blockchain.BestState.Beacon.BeaconHeight {
			synker.SetChainState(false, 0, true)
		} else {
			synker.SetChainState(false, 0, false)
		}

		if userRole == common.SHARD_ROLE && RCS.ClosestBeaconState.Height-1 <= synker.blockchain.BestState.Beacon.BeaconHeight {
			if RCS.ClosestShardsState[userShardID].Height == GetBestStateShard(userShardID).ShardHeight && RCS.ClosestShardsState[userShardID].Height >= GetBestStateBeacon().GetBestHeightOfShard(userShardID) {
				synker.SetChainState(false, 0, true)
				synker.SetChainState(true, userShardID, true)
			} else {
				synker.SetChainState(false, 0, false)
				synker.SetChainState(true, userShardID, false)
			}
		}
	}

	// sync ShardToBeacon & CrossShard pool
	if synker.IsLatest(false, 0) {
		switch userRole {
		case common.PROPOSER_ROLE, common.VALIDATOR_ROLE:
			if synker.blockchain.config.NodeMode == common.NODEMODE_AUTO || synker.blockchain.config.NodeMode == common.NODEMODE_BEACON {
				for shardID, peer := range RCS.ShardToBeaconBlks {
					for peerID, blks := range peer {
						synker.SyncBlkShardToBeacon(shardID, false, true, true, nil, blks, 0, 0, peerID)
					}
				}
				for shardID := byte(0); shardID < common.MAX_SHARD_NUMBER; shardID++ {
					if GetBestStateBeacon().GetBestHeightOfShard(shardID) < RCS.ClosestShardsState[shardID].Height {
						currentShardReqHeight := GetBestStateBeacon().GetBestHeightOfShard(shardID) + 1
						for peerID, peerState := range synker.States.PeersState {
							if _, ok := peerState.Shard[shardID]; ok {
								if currentShardReqHeight+DefaultMaxBlkReqPerPeer-1 >= RCS.ClosestShardsState[shardID].Height {
									synker.SyncBlkShardToBeacon(shardID, false, false, false, nil, nil, currentShardReqHeight, RCS.ClosestShardsState[shardID].Height, peerID)
								} else {
									synker.SyncBlkShardToBeacon(shardID, false, false, false, nil, nil, currentShardReqHeight, currentShardReqHeight+DefaultMaxBlkReqPerPeer-1, peerID)
									currentShardReqHeight += DefaultMaxBlkReqPerPeer - 1
								}
							}
						}
					}
				}
			}
		case common.SHARD_ROLE:
			if (synker.blockchain.config.NodeMode == common.NODEMODE_AUTO || synker.blockchain.config.NodeMode == common.NODEMODE_SHARD) && (userShardRole == common.PROPOSER_ROLE || userShardRole == common.VALIDATOR_ROLE) {
				if synker.IsLatest(true, userShardID) {
					for shardID, peer := range RCS.CrossShardBlks {
						for peerID, blks := range peer {
							synker.SyncBlkCrossShard(true, false, nil, blks, shardID, userShardID, peerID)
						}
					}
				}
			}
		}
	}

	// sync beacon
	currentBcnReqHeight := synker.blockchain.BestState.Beacon.BeaconHeight + 1
	if RCS.ClosestBeaconState.Height-synker.blockchain.BestState.Beacon.BeaconHeight > DefaultMaxBlkReqPerTime {
		RCS.ClosestBeaconState.Height = synker.blockchain.BestState.Beacon.BeaconHeight + DefaultMaxBlkReqPerTime
	}
	for peerID := range synker.States.PeersState {
		if currentBcnReqHeight+DefaultMaxBlkReqPerPeer-1 >= RCS.ClosestBeaconState.Height {
			//fmt.Println("SyncBlk1:", currentBcnReqHeight, RCS.ClosestBeaconState.Height)
			synker.SyncBlkBeacon(false, false, false, nil, nil, currentBcnReqHeight, RCS.ClosestBeaconState.Height, peerID)
			break
		} else {
			//fmt.Println("SyncBlk2:", currentBcnReqHeight, currentBcnReqHeight+DefaultMaxBlkReqPerPeer-1)
			synker.SyncBlkBeacon(false, false, false, nil, nil, currentBcnReqHeight, currentBcnReqHeight+DefaultMaxBlkReqPerPeer-1, peerID)
			currentBcnReqHeight += DefaultMaxBlkReqPerPeer - 1
		}
	}

	// sync shard
	for shardID := range synker.Status.Shards {
		currentShardReqHeight := synker.blockchain.BestState.Shard[shardID].ShardHeight + 1
		if RCS.ClosestShardsState[shardID].Height-synker.blockchain.BestState.Shard[shardID].ShardHeight > DefaultMaxBlkReqPerTime {
			RCS.ClosestShardsState[shardID] = ChainState{
				Height: synker.blockchain.BestState.Shard[shardID].ShardHeight + DefaultMaxBlkReqPerTime,
			}
		}

		for peerID := range synker.States.PeersState {
			if shardState, ok := synker.States.PeersState[peerID].Shard[shardID]; ok {
				fmt.Println("SyncShard state from other shard", shardID, shardState.Height)
				if shardState.Height >= currentShardReqHeight {
					if currentShardReqHeight+DefaultMaxBlkReqPerPeer-1 >= RCS.ClosestShardsState[shardID].Height {
						fmt.Println("SyncShard 1234 ", currentShardReqHeight, RCS.ClosestShardsState[shardID].Height)
						synker.SyncBlkShard(shardID, false, false, false, nil, nil, currentShardReqHeight, RCS.ClosestShardsState[shardID].Height, peerID)
						break
					} else {
						fmt.Println("SyncShard 12345")
						synker.SyncBlkShard(shardID, false, false, false, nil, nil, currentShardReqHeight, currentShardReqHeight+DefaultMaxBlkReqPerPeer-1, peerID)
						currentShardReqHeight += DefaultMaxBlkReqPerPeer - 1
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
	synker.blockchain.config.Server.UpdateConsensusState(userLayer, userPK, nil, synker.blockchain.BestState.Beacon.BeaconCommittee, synker.blockchain.BestState.Beacon.GetShardCommittee())

	synker.States.PeersState = make(map[libp2p.ID]*peerState)
	synker.Status.Unlock()
	synker.States.Unlock()
}

//SyncBlkBeacon Send a req to sync beacon block
/*
	- by Hash + blksHash: get by hash
	- from + to: get from main chain by height
	- GetFromPool: ignore mainchain, used only for hash
*/
func (synker *synker) SyncBlkBeacon(byHash bool, bySpecificHeights bool, getFromPool bool, blksHash []common.Hash, blkHeights []uint64, from uint64, to uint64, peerID libp2p.ID) {
	cacheItems := synker.Status.CurrentlySyncBlks.Items()
	if byHash {
		//Sync block by hash
		prefix := getBlkPrefixSyncKey(true, BeaconBlk, 0, 0)
		blksNeedToGet := getBlkNeedToGetByHash(prefix, blksHash, cacheItems, peerID)
		if len(blksNeedToGet) > 0 {
			go synker.blockchain.config.Server.PushMessageGetBlockBeaconByHash(blksNeedToGet, getFromPool, peerID)
		}
		for _, blkHash := range blksNeedToGet {
			synker.Status.CurrentlySyncBlks.Add(fmt.Sprintf("%v%v", prefix, blkHash.String()), time.Now().Unix(), DefaultMaxBlockSyncTime)
		}
	} else {
		//Sync by height
		prefix := getBlkPrefixSyncKey(false, BeaconBlk, 0, 0)
		if bySpecificHeights {
		} else {
			blkBatchsNeedToGet := getBlkNeedToGetByHeight(prefix, from, to, cacheItems, synker.GetBeaconPoolStateByHeight(), peerID)
			if len(blkBatchsNeedToGet) > 0 {
				for fromHeight, toHeight := range blkBatchsNeedToGet {
					go synker.blockchain.config.Server.PushMessageGetBlockBeaconByHeight(fromHeight, toHeight, peerID)
					for height := fromHeight; height <= toHeight; height++ {
						synker.Status.CurrentlySyncBlks.Add(fmt.Sprintf("%v%v", prefix, height), time.Now().Unix(), DefaultMaxBlockSyncTime)
					}
				}
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
func (synker *synker) SyncBlkShard(shardID byte, byHash bool, bySpecificHeights bool, getFromPool bool, blksHash []common.Hash, blkHeights []uint64, from uint64, to uint64, peerID libp2p.ID) {
	cacheItems := synker.Status.CurrentlySyncBlks.Items()
	if byHash {
		//Sync block by hash
		prefix := getBlkPrefixSyncKey(true, ShardBlk, shardID, 0)
		blksNeedToGet := getBlkNeedToGetByHash(prefix, blksHash, cacheItems, peerID)
		if len(blksNeedToGet) > 0 {
			go synker.blockchain.config.Server.PushMessageGetBlockShardByHash(shardID, blksNeedToGet, getFromPool, peerID)
		}
		for _, blkHash := range blksNeedToGet {
			synker.Status.CurrentlySyncBlks.Add(fmt.Sprintf("%v%v", prefix, blkHash.String()), time.Now().Unix(), DefaultMaxBlockSyncTime)
		}
	} else {
		//Sync by height
		prefix := getBlkPrefixSyncKey(false, ShardBlk, shardID, 0)
		if bySpecificHeights {
			blksNeedToGet := getBlkNeedToGetBySpecificHeight(prefix, blkHeights, cacheItems, synker.GetShardPoolStateByHeight(shardID), peerID)
			if len(blksNeedToGet) > 0 {
				go synker.blockchain.config.Server.PushMessageGetBlockShardBySpecificHeight(shardID, blksNeedToGet, getFromPool, peerID)
				for _, blkHeight := range blksNeedToGet {
					synker.Status.CurrentlySyncBlks.Add(fmt.Sprintf("%v%v", prefix, blkHeight), time.Now().Unix(), DefaultMaxBlockSyncTime)
				}
			}
		} else {
			blkBatchsNeedToGet := getBlkNeedToGetByHeight(prefix, from, to, cacheItems, synker.GetShardPoolStateByHeight(shardID), peerID)
			fmt.Println("SyncBlkShard", from, to, blkBatchsNeedToGet)
			if len(blkBatchsNeedToGet) > 0 {
				for fromHeight, toHeight := range blkBatchsNeedToGet {
					fmt.Println("SyncBlkShard", shardID, fromHeight, toHeight, peerID)
					go synker.blockchain.config.Server.PushMessageGetBlockShardByHeight(shardID, fromHeight, toHeight, peerID)
					for height := fromHeight; height <= toHeight; height++ {
						synker.Status.CurrentlySyncBlks.Add(fmt.Sprintf("%v%v", prefix, height), time.Now().Unix(), DefaultMaxBlockSyncTime)
					}
				}
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
func (synker *synker) SyncBlkShardToBeacon(shardID byte, byHash bool, bySpecificHeights bool, getFromPool bool, blksHash []common.Hash, blkHeights []uint64, from uint64, to uint64, peerID libp2p.ID) {
	cacheItems := synker.Status.CurrentlySyncBlks.Items()
	if byHash {
		//Sync block by hash
		prefix := getBlkPrefixSyncKey(true, ShardToBeaconBlk, shardID, 0)
		blksNeedToGet := getBlkNeedToGetByHash(prefix, blksHash, cacheItems, peerID)
		if len(blksNeedToGet) > 0 {
			go synker.blockchain.config.Server.PushMessageGetBlockShardToBeaconByHash(shardID, blksNeedToGet, getFromPool, peerID)
		}
		for _, blkHash := range blksNeedToGet {
			synker.Status.CurrentlySyncBlks.Add(fmt.Sprintf("%v%v", prefix, blkHash.String()), time.Now().Unix(), DefaultMaxBlockSyncTime)
		}
	} else {
		//Sync by height
		prefix := getBlkPrefixSyncKey(false, ShardToBeaconBlk, shardID, 0)
		if bySpecificHeights {
			blksNeedToGet := getBlkNeedToGetBySpecificHeight(prefix, blkHeights, cacheItems, synker.GetShardToBeaconPoolStateByHeight(shardID), peerID)
			if len(blksNeedToGet) > 0 {
				go synker.blockchain.config.Server.PushMessageGetBlockShardToBeaconBySpecificHeight(shardID, blksNeedToGet, getFromPool, peerID)
				for _, blkHeight := range blksNeedToGet {
					synker.Status.CurrentlySyncBlks.Add(fmt.Sprintf("%v%v", prefix, blkHeight), time.Now().Unix(), DefaultMaxBlockSyncTime)
				}
			}
		} else {
			blkBatchsNeedToGet := getBlkNeedToGetByHeight(prefix, from, to, cacheItems, synker.GetShardToBeaconPoolStateByHeight(shardID), peerID)
			if len(blkBatchsNeedToGet) > 0 {
				for fromHeight, toHeight := range blkBatchsNeedToGet {
					go synker.blockchain.config.Server.PushMessageGetBlockShardToBeaconByHeight(shardID, fromHeight, toHeight, peerID)
					for height := fromHeight; height <= toHeight; height++ {
						synker.Status.CurrentlySyncBlks.Add(fmt.Sprintf("%v%v", prefix, height), time.Now().Unix(), DefaultMaxBlockSyncTime)
					}
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
func (synker *synker) SyncBlkCrossShard(getFromPool bool, byHash bool, blksHash []common.Hash, blksHeight []uint64, fromShard byte, toShard byte, peerID libp2p.ID) {
	Logger.log.Criticalf("Shard %+v request CrossShardBlock with Height %+v from shard %+v \n", fromShard, blksHeight, toShard)
	cacheItems := synker.Status.CurrentlySyncBlks.Items()
	if byHash {
		prefix := getBlkPrefixSyncKey(true, CrossShardBlk, toShard, fromShard)
		blksNeedToGet := getBlkNeedToGetByHash(prefix, blksHash, cacheItems, peerID)
		if len(blksNeedToGet) > 0 {
			go synker.blockchain.config.Server.PushMessageGetBlockCrossShardByHash(fromShard, toShard, blksNeedToGet, getFromPool, peerID)
		}
		for _, blkHash := range blksNeedToGet {
			synker.Status.CurrentlySyncBlks.Add(fmt.Sprintf("%v%v", prefix, blkHash.String()), time.Now().Unix(), DefaultMaxBlockSyncTime)
		}
	} else {
		//Sync by specific heights
		prefix := getBlkPrefixSyncKey(false, CrossShardBlk, toShard, fromShard)
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
		blksNeedToGet := getBlkNeedToGetBySpecificHeight(prefix, blksHeight, cacheItems, synker.GetCrossShardPoolStateByHeight(fromShard), peerID)
		if len(blksNeedToGet) > 0 {
			go synker.blockchain.config.Server.PushMessageGetBlockCrossShardBySpecificHeight(fromShard, toShard, blksNeedToGet, getFromPool, peerID)
			for _, blkHeight := range blksNeedToGet {
				synker.Status.CurrentlySyncBlks.Add(fmt.Sprintf("%v%v", prefix, blkHeight), time.Now().Unix(), DefaultMaxBlockSyncTime)
			}
		}

	}
}

func (synker *synker) SetChainState(shard bool, shardID byte, ready bool) {
	synker.Status.IsLatest.Lock()
	defer synker.Status.IsLatest.Unlock()
	if shard {
		synker.Status.IsLatest.Shards[shardID] = ready
		// if ready {
		// 	fmt.Println("Shard is ready", shardID)
		// }
	} else {
		synker.Status.IsLatest.Beacon = ready
		// if ready {
		// 	fmt.Println("Beacon is ready")
		// }
	}
}

func (synker *synker) IsLatest(shard bool, shardID byte) bool {
	synker.Status.IsLatest.Lock()
	defer synker.Status.IsLatest.Unlock()
	if shard {
		if _, ok := synker.Status.IsLatest.Shards[shardID]; !ok {
			return false
		}
		return synker.Status.IsLatest.Shards[shardID]
	}
	return synker.Status.IsLatest.Beacon
}

func (synker *synker) GetPoolsState() {

	var (
		userRole      string
		userShardID   byte
		userShardRole string
		userPK        string
	)

	if synker.blockchain.config.UserKeySet != nil {
		userPK = synker.blockchain.config.UserKeySet.GetPublicKeyB58()
		userRole, userShardID = synker.blockchain.BestState.Beacon.GetPubkeyRole(userPK, synker.blockchain.BestState.Beacon.BestBlock.Header.Round)
		userShardRole = synker.blockchain.BestState.Shard[userShardID].GetPubkeyRole(userPK, synker.blockchain.BestState.Shard[userShardID].BestBlock.Header.Round)
	}

	synker.States.PoolsState.Lock()
	defer synker.States.PoolsState.Unlock()

	synker.States.PoolsState.BeaconPool = synker.blockchain.config.BeaconPool.GetAllBlockHeight()

	for shardID := range synker.Status.Shards {
		synker.States.PoolsState.ShardsPool[shardID] = synker.blockchain.config.ShardPool[shardID].GetAllBlockHeight()
	}

	if userRole == common.PROPOSER_ROLE || userRole == common.VALIDATOR_ROLE {
		synker.States.PoolsState.ShardToBeaconPool = synker.blockchain.config.ShardToBeaconPool.GetAllBlockHeight()
	}

	if userShardRole == common.PROPOSER_ROLE || userShardRole == common.VALIDATOR_ROLE {
		synker.States.PoolsState.CrossShardPool = synker.blockchain.config.CrossShardPool[userShardID].GetAllBlockHeight()
	}
}

func (synker *synker) GetBeaconPoolStateByHeight() []uint64 {
	synker.States.PoolsState.Lock()
	defer synker.States.PoolsState.Unlock()
	result := make([]uint64, len(synker.States.PoolsState.BeaconPool))
	copy(result, synker.States.PoolsState.BeaconPool)
	return result
}

func (synker *synker) GetShardPoolStateByHeight(shardID byte) []uint64 {
	synker.States.PoolsState.Lock()
	defer synker.States.PoolsState.Unlock()
	result := make([]uint64, len(synker.States.PoolsState.ShardsPool[shardID]))
	copy(result, synker.States.PoolsState.ShardsPool[shardID])
	return result
}

func (synker *synker) GetShardToBeaconPoolStateByHeight(shardID byte) []uint64 {
	synker.States.PoolsState.Lock()
	defer synker.States.PoolsState.Unlock()
	if blks, ok := synker.States.PoolsState.ShardToBeaconPool[shardID]; ok {
		result := make([]uint64, len(blks))
		copy(result, blks)
		return result
	}
	return nil
}

func (synker *synker) GetCrossShardPoolStateByHeight(fromShard byte) []uint64 {
	synker.States.PoolsState.Lock()
	defer synker.States.PoolsState.Unlock()
	if blks, ok := synker.States.PoolsState.CrossShardPool[fromShard]; ok {
		result := make([]uint64, len(blks))
		copy(result, blks)
		return result
	}
	return nil
}

func (synker *synker) GetCurrentSyncShards() []byte {
	synker.Status.Lock()
	defer synker.Status.Unlock()
	var currentSyncShards []byte
	for shardID := range synker.Status.Shards {
		currentSyncShards = append(currentSyncShards, shardID)
	}
	return currentSyncShards
}

var currentInsert = struct {
	Beacon sync.Mutex
	Shards map[byte]*sync.Mutex
}{
	Shards: make(map[byte]*sync.Mutex),
}

func (synker *synker) InsertBlockFromPool() {

	go synker.InsertBeaconBlockFromPool()

	synker.Status.Lock()
	for shardID := range synker.Status.Shards {
		if _, ok := currentInsert.Shards[shardID]; !ok {
			currentInsert.Shards[shardID] = &sync.Mutex{}
		}
		go func(shardID byte) {
			synker.InsertShardBlockFromPool(shardID)
		}(shardID)
	}
	synker.Status.Unlock()
}

func (synker *synker) InsertBeaconBlockFromPool() {
	currentInsert.Beacon.Lock()
	defer currentInsert.Beacon.Unlock()
	blks := synker.blockchain.config.BeaconPool.GetValidBlock()
	for _, newBlk := range blks {
		err := synker.blockchain.InsertBeaconBlock(newBlk, false)
		if err != nil {
			Logger.log.Error(err)
			break
		}
	}
}

func (synker *synker) InsertShardBlockFromPool(shardID byte) {
	currentInsert.Shards[shardID].Lock()
	blks := synker.blockchain.config.ShardPool[shardID].GetValidBlock()
	for _, newBlk := range blks {
		err := synker.blockchain.InsertShardBlock(newBlk, false)
		if err != nil {
			//@Notice: remove or keep invalid block
			Logger.log.Error(err)
			break
		}
	}
	currentInsert.Shards[shardID].Unlock()
}

func (synker *synker) GetClosestShardToBeaconPoolState() map[byte]uint64 {
	synker.States.Lock()
	result := make(map[byte]uint64)
	fmt.Println("ClosestShardToBeaconPoolState", synker.States.ClosestState.ShardToBeaconPool)
	for shardID, height := range synker.States.ClosestState.ShardToBeaconPool {
		result[shardID] = height
	}
	synker.States.Unlock()
	return result
}

func (synker *synker) GetClosestCrossShardPoolState() map[byte]uint64 {
	synker.States.Lock()
	result := make(map[byte]uint64)
	for shardID, height := range synker.States.ClosestState.CrossShardPool {
		result[shardID] = height
	}
	synker.States.Unlock()
	return result
}
