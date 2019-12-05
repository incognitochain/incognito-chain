package blockchain

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/pubsub"

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
	Timestamp     int64
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

type Synker struct {
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
			ClosestShardsState sync.Map
			ShardToBeaconPool  sync.Map
			CrossShardPool     sync.Map
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
	Event struct {
		requestSyncShardBlockByHashEvent    pubsub.EventChannel
		requestSyncShardBlockByHeightEvent  pubsub.EventChannel
		requestSyncBeaconBlockByHashEvent   pubsub.EventChannel
		requestSyncBeaconBlockByHeightEvent pubsub.EventChannel
	}
	blockchain    *BlockChain
	pubSubManager *pubsub.PubSubManager
	cQuit         chan struct{}
}

var currentInsert = struct {
	Beacon sync.Mutex
	Shards map[byte]*sync.Mutex
}{
	Shards: make(map[byte]*sync.Mutex),
}

func newSyncker(cQuit chan struct{}, blockchain *BlockChain, pubSubManager *pubsub.PubSubManager) Synker {
	s := Synker{
		blockchain:    blockchain,
		cQuit:         cQuit,
		pubSubManager: pubSubManager,
	}
	_, s.Event.requestSyncShardBlockByHashEvent, _ = pubSubManager.RegisterNewSubscriber(pubsub.RequestShardBlockByHashTopic)
	_, s.Event.requestSyncShardBlockByHeightEvent, _ = pubSubManager.RegisterNewSubscriber(pubsub.RequestShardBlockByHeightTopic)
	_, s.Event.requestSyncBeaconBlockByHashEvent, _ = pubSubManager.RegisterNewSubscriber(pubsub.RequestBeaconBlockByHashTopic)
	_, s.Event.requestSyncBeaconBlockByHeightEvent, _ = pubSubManager.RegisterNewSubscriber(pubsub.RequestBeaconBlockByHeightTopic)
	return s
}
func (synker *Synker) Start() {
	if synker.Status.Beacon {
		return
	}
	synker.Status.Beacon = true
	synker.Status.CurrentlySyncBlks = cache.New(DefaultMaxBlockSyncTime, DefaultCacheCleanupTime)
	synker.Status.Shards = make(map[byte]struct{})
	synker.Status.IsLatest.Shards = make(map[byte]bool)
	synker.States.PeersState = make(map[libp2p.ID]*peerState)
	synker.States.ClosestState.ClosestShardsState = sync.Map{}
	synker.States.ClosestState.ShardToBeaconPool = sync.Map{}
	synker.States.ClosestState.CrossShardPool = sync.Map{}
	synker.States.PoolsState.ShardToBeaconPool = make(map[byte][]uint64)
	synker.States.PoolsState.CrossShardPool = make(map[byte][]uint64)
	synker.States.PoolsState.ShardsPool = make(map[byte][]uint64)
	for shardID := 0; shardID < common.MaxShardNumber; shardID++ {
		currentInsert.Shards[byte(shardID)] = &sync.Mutex{}
	}
	synker.Status.Lock()
	synker.startSyncRelayShards()
	synker.Status.Unlock()

	broadcastTicker := time.NewTicker(DefaultBroadcastStateTime)
	insertPoolTicker := time.NewTicker(1 * time.Second)
	updateStatesTicker := time.NewTicker(DefaultStateUpdateTime)
	defer func() {
		broadcastTicker.Stop()
		insertPoolTicker.Stop()
		updateStatesTicker.Stop()
	}()

	for {
		select {
		case <-synker.cQuit:
			return
		case <-insertPoolTicker.C:
			synker.InsertBlockFromPool()
		case <-broadcastTicker.C:
			synker.blockchain.config.Server.BoardcastNodeState()
		case <-updateStatesTicker.C:
			synker.UpdateState()
		case msg := <-synker.Event.requestSyncShardBlockByHashEvent:
			// Message Value: "[shardID],[BlockHash]"
			str, ok := msg.Value.(string)
			if !ok {
				continue
			}
			strs := strings.Split(str, ",")
			shardID, err := strconv.Atoi(strs[0])
			if err != nil {
				continue
			}
			hash, err := common.Hash{}.NewHashFromStr(strs[1])
			if err != nil {
				continue
			}
			synker.SyncBlkShard(byte(shardID), true, false, true, []common.Hash{*hash}, []uint64{}, 0, 0, "")
		case msg := <-synker.Event.requestSyncShardBlockByHeightEvent:
			// Message Value: "[shardID],[blockheight]"
			str, ok := msg.Value.(string)
			if !ok {
				continue
			}
			strs := strings.Split(str, ",")
			shardID, err := strconv.Atoi(strs[0])
			if err != nil {
				continue
			}
			height, err := strconv.Atoi(strs[1])
			if err != nil {
				continue
			}
			synker.SyncBlkShard(byte(shardID), false, true, true, []common.Hash{}, []uint64{uint64(height)}, uint64(height), uint64(height), "")
		case msg := <-synker.Event.requestSyncBeaconBlockByHashEvent:
			// Message Value: [BlockHash]
			hash, ok := msg.Value.(common.Hash)
			if !ok {
				continue
			}
			synker.SyncBlkBeacon(true, false, true, []common.Hash{hash}, []uint64{}, 0, 0, "")
		case msg := <-synker.Event.requestSyncBeaconBlockByHeightEvent:
			// Message Value: [blockheight]
			height, ok := msg.Value.(uint64)
			if !ok {
				continue
			}
			synker.SyncBlkBeacon(false, true, true, []common.Hash{}, []uint64{uint64(height)}, uint64(height), uint64(height), "")
		}
	}
}

func (synker *Synker) SyncShard(shardID byte) error {
	synker.Status.Lock()
	defer synker.Status.Unlock()
	return synker.syncShard(shardID)
}

func (synker *Synker) syncShard(shardID byte) error {
	if _, ok := synker.Status.Shards[shardID]; ok {
		return errors.New("Shard " + fmt.Sprintf("%d", shardID) + " synchronzation is already started")
	}
	Logger.log.Infof("*** Start syncing shard %+v ***", shardID)
	synker.Status.Shards[shardID] = struct{}{}
	return nil
}

func (synker *Synker) startSyncRelayShards() {
	for _, shardID := range synker.blockchain.config.RelayShards {
		if shardID > byte(synker.blockchain.BestState.Beacon.ActiveShards-1) {
			break
		}
		synker.syncShard(shardID)
	}
}

func (synker *Synker) StopSyncUnnecessaryShard() {
	synker.Status.Lock()
	defer synker.Status.Unlock()
	synker.stopSyncUnnecessaryShard()
}

func (synker *Synker) stopSyncUnnecessaryShard() {
	for shardID := byte(0); shardID < common.MaxShardNumber; shardID++ {
		synker.stopSyncShard(shardID)
	}
}

func (synker *Synker) stopSyncShard(shardID byte) error {
	if synker.blockchain.config.NodeMode == common.NodeModeAuto || synker.blockchain.config.NodeMode == common.NodeModeShard {
		_, _, userShardIDInt := synker.blockchain.config.ConsensusEngine.GetUserRole()
		if userShardIDInt >= 0 {
			if byte(userShardIDInt) == shardID {
				return errors.New("Shard " + fmt.Sprintf("%d", shardID) + " synchronzation can't be stopped")
			}
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

func (synker *Synker) UpdateState() {
	synker.Status.Lock()
	synker.States.Lock()
	synker.GetPoolsState()
	synker.Status.CurrentlySyncBlks.DeleteExpired()
	var shardsStateClone map[byte]ShardBestState
	shardsStateClone = make(map[byte]ShardBestState)
	beaconStateCloneBytes, err := synker.blockchain.BestState.Beacon.MarshalJSON()
	if err != nil {
		synker.Status.Unlock()
		synker.States.Unlock()
		panic(err)
	}
	var beaconStateClone BeaconBestState
	err = json.Unmarshal(beaconStateCloneBytes, &beaconStateClone)
	if err != nil {
		synker.Status.Unlock()
		synker.States.Unlock()
		panic(err)
	}
	var (
		userRole       string
		userLayer      string
		userShardRole  string
		userShardIDInt int
	)

	userLayer, userRole, userShardIDInt = synker.blockchain.config.ConsensusEngine.GetUserRole()

	Logger.log.Infof("Chain: %v, Role: %v, ShardID: %v", userLayer, userRole, userShardIDInt)

	if userLayer == common.ShardRole && userRole != common.WaitingRole {
		// userShardID = byte(userShardIDInt)
		synker.syncShard(byte(userShardIDInt))
		userKeyForCheckRole, _ := synker.blockchain.config.ConsensusEngine.GetCurrentMiningPublicKey()
		userShardRole = synker.blockchain.BestState.Shard[byte(userShardIDInt)].GetPubkeyRole(userKeyForCheckRole, synker.blockchain.BestState.Shard[byte(userShardIDInt)].BestBlock.Header.Round)
	}

	synker.stopSyncUnnecessaryShard()

	synker.States.ClosestState.ClosestBeaconState = beaconStateClone.BeaconHeight
	for shardID, beststate := range synker.blockchain.BestState.Shard {
		synker.States.ClosestState.ClosestShardsState.Store(shardID, beststate.ShardHeight)
	}

	for k, v := range synker.blockchain.config.ShardToBeaconPool.GetLatestValidPendingBlockHeight() {
		synker.States.ClosestState.ShardToBeaconPool.Store(k, v)
	}
	if userShardIDInt >= 0 {
		for k, v := range synker.blockchain.config.CrossShardPool[byte(userShardIDInt)].GetLatestValidBlockHeight() {
			synker.States.ClosestState.CrossShardPool.Store(k, v)
		}
	}

	RCS := reportedChainState{
		ClosestBeaconState: ChainState{
			Height: beaconStateClone.BeaconHeight,
		},
		ClosestShardsState: make(map[byte]ChainState),
		ShardToBeaconBlks:  make(map[byte]map[libp2p.ID][]uint64),
		CrossShardBlks:     make(map[byte]map[libp2p.ID][]uint64),
	}

	bestShardsHeight := beaconStateClone.GetBestShardHeight()
	for shardID := byte(0); shardID < common.MaxShardNumber; shardID++ {
		RCS.ClosestShardsState[shardID] = ChainState{
			Height: bestShardsHeight[shardID],
		}
	}
	for shardID := range synker.Status.Shards {
		cloneState := ShardBestState{}
		shardStateCloneBytes, err := synker.blockchain.BestState.Shard[shardID].MarshalJSON()
		if err != nil {
			panic(err)
		}
		err = json.Unmarshal(shardStateCloneBytes, &cloneState)
		if err != nil {
			panic(err)
		}
		shardsStateClone[shardID] = cloneState
		RCS.ClosestShardsState[shardID] = ChainState{
			Height: shardsStateClone[shardID].ShardHeight,
		}
	}
	for peerID, peerState := range synker.States.PeersState {
		for shardID := range synker.Status.Shards {
			if shardState, ok := peerState.Shard[shardID]; ok {
				if shardState.Height >= GetBeaconBestState().GetBestHeightOfShard(shardID) && shardState.Height > GetBestStateShard(shardID).ShardHeight {
					if RCS.ClosestShardsState[shardID].Height == shardsStateClone[shardID].ShardHeight {
						RCS.ClosestShardsState[shardID] = *shardState
					} else {
						if shardState.Height < RCS.ClosestShardsState[shardID].Height {
							RCS.ClosestShardsState[shardID] = *shardState
						}
					}
				}
			}
		}

		if peerState.Beacon.Height > beaconStateClone.BeaconHeight {
			if peerState.Beacon.Height < RCS.ClosestBeaconState.Height {
				RCS.ClosestBeaconState = *peerState.Beacon
			}

			if RCS.ClosestBeaconState.Height == beaconStateClone.BeaconHeight {
				RCS.ClosestBeaconState = *peerState.Beacon
			}
		}

		// record pool state
		switch userLayer {
		case common.BeaconRole:
			if (synker.blockchain.config.NodeMode == common.NodeModeAuto || synker.blockchain.config.NodeMode == common.NodeModeBeacon) && userRole == common.CommitteeRole {
				//fmt.Println("SYN: s2b", peerState.ShardToBeaconPool)
				if peerState.ShardToBeaconPool != nil {
					for shardID, blkHeights := range *peerState.ShardToBeaconPool {
						if len(synker.States.PoolsState.ShardToBeaconPool[shardID]) > 0 {
							if _, ok := RCS.ShardToBeaconBlks[shardID]; !ok {
								RCS.ShardToBeaconBlks[shardID] = make(map[libp2p.ID][]uint64)
							}
							RCS.ShardToBeaconBlks[shardID][peerID] = blkHeights
							if len(blkHeights) > 0 && len(blkHeights) <= len(synker.States.PoolsState.ShardToBeaconPool[shardID]) {
								commonHeights := arrayCommonElements(blkHeights, synker.States.PoolsState.ShardToBeaconPool[shardID])
								if len(commonHeights) > 0 {
									sort.Slice(commonHeights, func(i, j int) bool { return commonHeights[i] < commonHeights[j] })
									height, _ := synker.States.ClosestState.ShardToBeaconPool.Load(shardID)
									for idx := len(commonHeights) - 1; idx == 0; idx-- {
										if height.(uint64) > commonHeights[idx] {
											synker.States.ClosestState.ShardToBeaconPool.Store(shardID, commonHeights[idx])
											break
										}
									}
								}
							}
						}
					}
				}
				for shardID := byte(0); shardID < common.MaxShardNumber; shardID++ {
					//fmt.Println("SYN: set ClosestShardsState", peerState.Shard, RCS.ClosestShardsState)
					if shardState, ok := peerState.Shard[shardID]; ok {
						if shardState.Height >= GetBeaconBestState().GetBestHeightOfShard(shardID) {
							if RCS.ClosestShardsState[shardID].Height == GetBeaconBestState().GetBestHeightOfShard(shardID) {
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
		case common.ShardRole:
			if (synker.blockchain.config.NodeMode == common.NodeModeAuto || synker.blockchain.config.NodeMode == common.NodeModeShard) && (userShardRole == common.ProposerRole || userShardRole == common.ValidatorRole) {
				if pool, ok := peerState.CrossShardPool[byte(userShardIDInt)]; ok {
					for shardID, blkHeights := range *pool {
						if _, ok := RCS.CrossShardBlks[shardID]; !ok {
							RCS.CrossShardBlks[shardID] = make(map[libp2p.ID][]uint64)
						}
						RCS.CrossShardBlks[shardID][peerID] = blkHeights

						if len(blkHeights) > 0 && len(blkHeights) <= len(synker.States.PoolsState.CrossShardPool[shardID]) {
							commonHeights := arrayCommonElements(blkHeights, synker.States.PoolsState.CrossShardPool[shardID])
							if len(commonHeights) > 0 {
								sort.Slice(commonHeights, func(i, j int) bool { return commonHeights[i] < commonHeights[j] })
								for idx := len(commonHeights) - 1; idx < 0; idx-- {
									height, _ := synker.States.ClosestState.CrossShardPool.Load(shardID)
									if height.(uint64) > commonHeights[idx] {
										synker.States.ClosestState.CrossShardPool.Store(shardID, commonHeights[idx])
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
		synker.States.ClosestState.ClosestShardsState.Store(shardID, state.Height)
	}

	if len(synker.States.PeersState) > 0 {
		if userLayer != common.ShardRole {
			if RCS.ClosestBeaconState.Height == beaconStateClone.BeaconHeight {
				synker.SetChainState(false, 0, true)
			} else {
				fmt.Printf("beacon not ready %v %v\n\n\n\n", userRole, RCS.ClosestBeaconState.Height)
				synker.SetChainState(false, 0, false)
			}
		}

		if userLayer == common.ShardRole && RCS.ClosestBeaconState.Height-1 <= beaconStateClone.BeaconHeight {
			if RCS.ClosestShardsState[byte(userShardIDInt)].Height == GetBestStateShard(byte(userShardIDInt)).ShardHeight && RCS.ClosestShardsState[byte(userShardIDInt)].Height >= GetBeaconBestState().GetBestHeightOfShard(byte(userShardIDInt)) {
				synker.SetChainState(false, 0, true)
				synker.SetChainState(true, byte(userShardIDInt), true)
			} else {
				fmt.Println("shard not ready", RCS.ClosestShardsState[byte(userShardIDInt)].Height)
				synker.SetChainState(false, 0, false)
				synker.SetChainState(true, byte(userShardIDInt), false)
			}
		}
	}

	// sync ShardToBeacon & CrossShard pool
	if synker.IsLatest(false, 0) {
		switch userLayer {
		case common.BeaconRole:
			if (synker.blockchain.config.NodeMode == common.NodeModeAuto || synker.blockchain.config.NodeMode == common.NodeModeBeacon) && userRole == common.CommitteeRole {
				for shardID, peer := range RCS.ShardToBeaconBlks {
					for peerID, blks := range peer {
						synker.SyncBlkShardToBeacon(shardID, false, true, true, nil, blks, 0, 0, peerID)
					}
				}
				for shardID := byte(0); shardID < common.MaxShardNumber; shardID++ {
					if GetBeaconBestState().GetBestHeightOfShard(shardID) < RCS.ClosestShardsState[shardID].Height {
						currentShardReqHeight := GetBeaconBestState().GetBestHeightOfShard(shardID) + 1
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
		case common.ShardRole:
			if (synker.blockchain.config.NodeMode == common.NodeModeAuto || synker.blockchain.config.NodeMode == common.NodeModeShard) && (userShardRole == common.ProposerRole || userShardRole == common.ValidatorRole) {
				if synker.IsLatest(true, byte(userShardIDInt)) {
					for shardID, peer := range RCS.CrossShardBlks {
						for peerID, blks := range peer {
							synker.SyncBlkCrossShard(true, false, nil, blks, shardID, byte(userShardIDInt), peerID)
						}
					}
				}
			}
		}
	}

	// sync beacon
	currentBcnReqHeight := beaconStateClone.BeaconHeight + 1
	if RCS.ClosestBeaconState.Height-beaconStateClone.BeaconHeight > DefaultMaxBlkReqPerTime {
		RCS.ClosestBeaconState.Height = beaconStateClone.BeaconHeight + DefaultMaxBlkReqPerTime
	}
	for peerID, peerState := range synker.States.PeersState {
		if peerState.Beacon.Height == GetBeaconBestState().BeaconHeight && !peerState.Beacon.BlockHash.IsEqual(&GetBeaconBestState().BestBlockHash) {
			synker.SyncBlkBeacon(true, false, false, []common.Hash{peerState.Beacon.BlockHash}, nil, 0, 0, peerID)
		}

		if peerState.Beacon.Height >= currentBcnReqHeight {
			if currentBcnReqHeight+DefaultMaxBlkReqPerPeer-1 >= RCS.ClosestBeaconState.Height {
				//fmt.Println("SyncBlk1:", currentBcnReqHeight, RCS.ClosestBeaconState.Height)
				synker.SyncBlkBeacon(false, false, false, nil, nil, currentBcnReqHeight, RCS.ClosestBeaconState.Height+10, peerID)
				break
			} else {
				//fmt.Println("SyncBlk2:", currentBcnReqHeight, currentBcnReqHeight+DefaultMaxBlkReqPerPeer-1)
				synker.SyncBlkBeacon(false, false, false, nil, nil, currentBcnReqHeight, currentBcnReqHeight+DefaultMaxBlkReqPerPeer-1, peerID)
				currentBcnReqHeight += DefaultMaxBlkReqPerPeer - 1
			}
		}
	}

	// sync shard
	for shardID := range synker.Status.Shards {
		currentShardReqHeight := shardsStateClone[shardID].ShardHeight + 1
		if RCS.ClosestShardsState[shardID].Height-shardsStateClone[shardID].ShardHeight > DefaultMaxBlkReqPerTime {
			RCS.ClosestShardsState[shardID] = ChainState{
				Height: shardsStateClone[shardID].ShardHeight + DefaultMaxBlkReqPerTime,
			}
		}

		for peerID := range synker.States.PeersState {
			if shardState, ok := synker.States.PeersState[peerID].Shard[shardID]; ok {
				//fmt.Println("SyncShard state from other shard", shardID, currentShardReqHeight, shardState.Height, RCS.ClosestShardsState[shardID].Height)
				if shardState.Height == GetBestStateShard(shardID).ShardHeight && !shardState.BlockHash.IsEqual(&GetBestStateShard(shardID).BestBlockHash) {
					synker.SyncBlkShard(shardID, true, false, false, []common.Hash{shardState.BlockHash}, nil, 0, 0, peerID)
				}

				if shardState.Height >= currentShardReqHeight {
					if currentShardReqHeight+DefaultMaxBlkReqPerPeer-1 >= RCS.ClosestShardsState[shardID].Height {
						fmt.Println("SyncShard sent to a peer ", currentShardReqHeight, RCS.ClosestShardsState[shardID].Height+1)
						synker.SyncBlkShard(shardID, false, false, false, nil, nil, currentShardReqHeight, RCS.ClosestShardsState[shardID].Height+10, peerID)
						break
					} else {
						//fmt.Println("SyncShard sent to a peer ", currentShardReqHeight, currentShardReqHeight+DefaultMaxBlkReqPerPeer-1)
						synker.SyncBlkShard(shardID, false, false, false, nil, nil, currentShardReqHeight, currentShardReqHeight+DefaultMaxBlkReqPerPeer-1, peerID)
						currentShardReqHeight += DefaultMaxBlkReqPerPeer - 1
					}
				}
			}
		}
	}
	//TODO hy Get Committee LightWeightPublicKey ExtractMiningPublickeysFromCommitteeKeyList
	// beaconCommittee, _ := incognitokey.ExtractPublickeysFromCommitteeKeyList(beaconStateClone.BeaconCommittee, beaconStateClone.ConsensusAlgorithm)
	// shardCommittee := make(map[byte][]string)
	// for shardID, committee := range beaconStateClone.GetShardCommittee() {
	// 	shardCommittee[shardID], _ = incognitokey.ExtractPublickeysFromCommitteeKeyList(committee, beaconStateClone.ShardConsensusAlgorithm[shardID])
	// }

	beaconCommittee, _ := incognitokey.ExtractMiningPublickeysFromCommitteeKeyList(beaconStateClone.BeaconCommittee, beaconStateClone.ConsensusAlgorithm)
	shardCommittee := make(map[byte][]string)
	for shardID, committee := range beaconStateClone.GetShardCommittee() {
		shardCommittee[shardID], _ = incognitokey.ExtractMiningPublickeysFromCommitteeKeyList(committee, beaconStateClone.ShardConsensusAlgorithm[shardID])
	}
	userMiningKey, err := synker.blockchain.config.ConsensusEngine.GetMiningPublicKeyByConsensus(synker.blockchain.BestState.Beacon.ConsensusAlgorithm)
	if err != nil {
		synker.Status.Unlock()
		synker.States.Unlock()
		panic(err)
	}
	if userLayer == common.ShardRole {
		shardID := byte(userShardIDInt)
		shardState := shardsStateClone[shardID]
		keyList, _ := incognitokey.ExtractMiningPublickeysFromCommitteeKeyList(shardState.ShardCommittee, shardState.ConsensusAlgorithm)
		shardCommittee[shardID] = keyList
		synker.blockchain.config.Server.UpdateConsensusState(userLayer, userMiningKey, &shardID, beaconCommittee, shardCommittee)
	} else {
		synker.blockchain.config.Server.UpdateConsensusState(userLayer, userMiningKey, nil, beaconCommittee, shardCommittee)
	}

	if userLayer == common.ShardRole && (userShardRole == common.ProposerRole || userShardRole == common.ValidatorRole) {
		for shardID, shard := range synker.blockchain.BestState.Beacon.LastCrossShardState {
			height, ok := shard[byte(userShardIDInt)]
			if !ok {
				continue
			}
			if height > synker.blockchain.BestState.Shard[byte(userShardIDInt)].BestCrossShard[shardID] {
				for peerID := range synker.States.PeersState {
					if shardState, ok := synker.States.PeersState[peerID].Shard[shardID]; ok {
						if shardState.Height >= height {
							synker.SyncBlkCrossShard(false, false, nil, []uint64{height}, shardID, byte(userShardIDInt), peerID)
							break
						}
					}
				}
			}
		}

	}
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
func (synker *Synker) SyncBlkBeacon(byHash bool, bySpecificHeights bool, getFromPool bool, blksHash []common.Hash, blkHeights []uint64, from uint64, to uint64, peerID libp2p.ID) {
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
					fmt.Println("SYNC beacon:", fromHeight, toHeight, peerID.String())
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
func (synker *Synker) SyncBlkShard(shardID byte, byHash bool, bySpecificHeights bool, getFromPool bool, blksHash []common.Hash, blkHeights []uint64, from uint64, to uint64, peerID libp2p.ID) {
	cacheItems := synker.Status.CurrentlySyncBlks.Items()
	if byHash {
		//Sync block by hash
		prefix := getBlkPrefixSyncKey(true, ShardBlk, shardID, 0)
		blksNeedToGet := getBlkNeedToGetByHash(prefix, blksHash, cacheItems, peerID)
		if len(blksNeedToGet) > 0 {
			go synker.blockchain.config.Server.PushMessageGetBlockShardByHash(shardID, blksNeedToGet, getFromPool, peerID)
		}
		for _, blkHash := range blksNeedToGet {
			Logger.log.Criticalf("Block need to get hash=%+v", blkHash.String())
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
func (synker *Synker) SyncBlkShardToBeacon(shardID byte, byHash bool, bySpecificHeights bool, getFromPool bool, blksHash []common.Hash, blkHeights []uint64, from uint64, to uint64, peerID libp2p.ID) {
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
		Logger.log.Info("REQUEST SYNC S2B", blkHeights, shardID)
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
func (synker *Synker) SyncBlkCrossShard(getFromPool bool, byHash bool, blksHash []common.Hash, blksHeight []uint64, fromShard byte, toShard byte, peerID libp2p.ID) {
	Logger.log.Criticalf("Receiver Shard %+v request CrossShardBlock with Height %+v from sender shard %+v \n", toShard, blksHeight, fromShard)
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

func (synker *Synker) SetChainState(shard bool, shardID byte, ready bool) {
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
		// } else {
		// 	fmt.Println("Beacon is not ready")
		// }
	}
}

func (synker *Synker) IsLatest(shard bool, shardID byte) bool {
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

func (synker *Synker) GetPoolsState() {

	var (
		userRole      string
		userShardID   byte
		userShardRole string
		userPK        string
	)
	userPK, _ = synker.blockchain.config.ConsensusEngine.GetCurrentMiningPublicKey()

	if userPK != "" {
		userRole, userShardID = synker.blockchain.BestState.Beacon.GetPubkeyRole(userPK, synker.blockchain.BestState.Beacon.BestBlock.Header.Round)
		userShardRole = synker.blockchain.BestState.Shard[userShardID].GetPubkeyRole(userPK, synker.blockchain.BestState.Shard[userShardID].BestBlock.Header.Round)
	}

	synker.States.PoolsState.Lock()
	defer synker.States.PoolsState.Unlock()

	synker.States.PoolsState.BeaconPool = synker.blockchain.config.BeaconPool.GetAllBlockHeight()

	for shardID := range synker.Status.Shards {
		synker.States.PoolsState.ShardsPool[shardID] = synker.blockchain.config.ShardPool[shardID].GetAllBlockHeight()
	}

	if userRole == common.ProposerRole || userRole == common.ValidatorRole {
		synker.States.PoolsState.ShardToBeaconPool = synker.blockchain.config.ShardToBeaconPool.GetAllBlockHeight()
	}

	if userShardRole == common.ProposerRole || userShardRole == common.ValidatorRole {
		synker.States.PoolsState.CrossShardPool = synker.blockchain.config.CrossShardPool[userShardID].GetAllBlockHeight()
	}
}

func (synker *Synker) GetBeaconPoolStateByHeight() []uint64 {
	synker.States.PoolsState.Lock()
	defer synker.States.PoolsState.Unlock()
	result := make([]uint64, len(synker.States.PoolsState.BeaconPool))
	copy(result, synker.States.PoolsState.BeaconPool)
	return result
}

func (synker *Synker) GetShardPoolStateByHeight(shardID byte) []uint64 {
	synker.States.PoolsState.Lock()
	defer synker.States.PoolsState.Unlock()
	result := make([]uint64, len(synker.States.PoolsState.ShardsPool[shardID]))
	copy(result, synker.States.PoolsState.ShardsPool[shardID])
	return result
}

func (synker *Synker) GetShardToBeaconPoolStateByHeight(shardID byte) []uint64 {
	synker.States.PoolsState.Lock()
	defer synker.States.PoolsState.Unlock()
	if blks, ok := synker.States.PoolsState.ShardToBeaconPool[shardID]; ok {
		result := make([]uint64, len(blks))
		copy(result, blks)
		return result
	}
	return nil
}

func (synker *Synker) GetCrossShardPoolStateByHeight(fromShard byte) []uint64 {
	synker.States.PoolsState.Lock()
	defer synker.States.PoolsState.Unlock()
	if blks, ok := synker.States.PoolsState.CrossShardPool[fromShard]; ok {
		result := make([]uint64, len(blks))
		copy(result, blks)
		return result
	}
	return nil
}

func (synker *Synker) GetCurrentSyncShards() []byte {
	synker.Status.Lock()
	defer synker.Status.Unlock()
	var currentSyncShards []byte
	for shardID := range synker.Status.Shards {
		currentSyncShards = append(currentSyncShards, shardID)
	}
	return currentSyncShards
}

func (synker *Synker) InsertBlockFromPool() {
	go func() {
		if !synker.blockchain.config.ConsensusEngine.IsOngoing(common.BeaconChainKey) {
			synker.InsertBeaconBlockFromPool()
		}
	}()

	synker.Status.Lock()
	for shardID := range synker.Status.Shards {
		if !synker.blockchain.config.ConsensusEngine.IsOngoing(common.GetShardChainKey(shardID)) {
			go func(shardID byte) {
				synker.InsertShardBlockFromPool(shardID)
			}(shardID)
		}
	}
	synker.Status.Unlock()
}

func (synker *Synker) InsertBeaconBlockFromPool() {
	currentInsert.Beacon.Lock()
	defer currentInsert.Beacon.Unlock()
	blocks := synker.blockchain.config.BeaconPool.GetValidBlock()
	if len(blocks) > 0 {
		fmt.Println("InsertBeaconBlockFromPool", len(blocks))
	}

	chain := synker.blockchain.Chains[common.BeaconChainKey]

	curEpoch := GetBeaconBestState().Epoch
	sameCommitteeBlock := blocks
	for i, v := range blocks {
		if v.GetCurrentEpoch() == curEpoch+1 {
			sameCommitteeBlock = blocks[:i+1]
			break
		}
	}

	for i, blk := range sameCommitteeBlock {
		if i == len(sameCommitteeBlock)-1 {
			break
		}
		if blk.Header.Height != sameCommitteeBlock[i+1].Header.Height-1 {
			//fmt.Println(sameCommitteeBlock[0].Header.Height, blk.Header.Height, sameCommitteeBlock[i+1].Header.Height)
			sameCommitteeBlock = blocks[:i+1]
			break
		}
	}

	for i := len(sameCommitteeBlock) - 1; i >= 0; i-- {
		if err := chain.ValidateBlockSignatures(sameCommitteeBlock[i], beaconBestState.BeaconCommittee); err != nil {
			sameCommitteeBlock = sameCommitteeBlock[:i]
			//TODO: remove invalid block
		} else {
			break
		}
	}

	if len(sameCommitteeBlock) > 0 {
		if sameCommitteeBlock[0].Header.Height-1 != GetBeaconBestState().BeaconHeight {
			//fmt.Println("DEBUG: beacon", sameCommitteeBlock[0].Header.Height-1, GetBeaconBestState().BeaconHeight)
			return
		}
	}

	for _, v := range sameCommitteeBlock {
		err := chain.InsertBlk(v)
		if err != nil {
			Logger.log.Error(err)
			break
		}
	}
}

func (synker *Synker) InsertShardBlockFromPool(shardID byte) {
	fmt.Println("InsertShardBlockFromPool start")
	currentInsert.Shards[shardID].Lock()
	defer currentInsert.Shards[shardID].Unlock()

	blocks := synker.blockchain.config.ShardPool[shardID].GetValidBlock()
	if len(blocks) > 0 {
		fmt.Println("InsertShardBlockFromPool", len(blocks))
	}

	chain := synker.blockchain.Chains[common.GetShardChainKey(shardID)]
	curEpoch := GetBestStateShard(shardID).Epoch
	sameCommitteeBlock := blocks

	for i, v := range blocks {
		if v.GetCurrentEpoch() == curEpoch+1 {
			sameCommitteeBlock = blocks[:i+1]
			break
		}
	}

	for i, blk := range sameCommitteeBlock {
		if i == len(sameCommitteeBlock)-1 {
			break
		}
		if blk.Header.Height != sameCommitteeBlock[i+1].Header.Height-1 {
			//fmt.Println(sameCommitteeBlock[0].Header.Height, blk.Header.Height, sameCommitteeBlock[i+1].Header.Height)
			sameCommitteeBlock = blocks[:i+1]
			break
		}
	}

	for i := len(sameCommitteeBlock) - 1; i >= 0; i-- {
		if err := chain.ValidateBlockSignatures(sameCommitteeBlock[i], chain.GetCommittee()); err != nil {
			sameCommitteeBlock = sameCommitteeBlock[:i]
			//TODO: remove invalid block
		} else {
			break
		}
	}

	if len(sameCommitteeBlock) > 0 {
		if sameCommitteeBlock[0].Header.Height-1 != GetBestStateShard(shardID).ShardHeight {
			//fmt.Println("DEBUG: shard", sameCommitteeBlock[0].Header.Height-1, GetBestStateShard(shardID).ShardHeight)
			return
		}
	}

	for _, v := range sameCommitteeBlock {
		//time1 := time.Now()
		//fmt.Println("DEBUG: shard", v.Header.Height)
		err := chain.InsertBlk(v)
		//fmt.Println("DEBUG: shard ", time.Since(time1).Seconds())
		if err != nil {
			Logger.log.Error(err)
			break
		}
	}

}

func (synker *Synker) GetClosestShardToBeaconPoolState() map[byte]uint64 {
	result := make(map[byte]uint64)
	synker.States.ClosestState.ShardToBeaconPool.Range(func(k interface{}, v interface{}) bool {
		shardID := k.(byte)
		height := v.(uint64)
		result[shardID] = height
		return true
	})
	return result
}

func (synker *Synker) GetClosestCrossShardPoolState() map[byte]uint64 {
	result := make(map[byte]uint64)
	synker.States.ClosestState.CrossShardPool.Range(func(k interface{}, v interface{}) bool {
		shardID := k.(byte)
		height := v.(uint64)
		result[shardID] = height
		return true
	})
	return result
}
