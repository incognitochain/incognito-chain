package blockchain

import (
	"errors"
	"fmt"
	"sync"

	"github.com/constant-money/constant-chain/common"
	libp2p "github.com/libp2p/go-libp2p-peer"
	"github.com/patrickmn/go-cache"
)

type synker struct {
	Status struct {
		sync.Mutex
		Beacon            bool
		Shards            map[byte]struct{}
		CurrentlySyncBlks *cache.Cache
		IsReady           struct {
			sync.Mutex
			Beacon bool
			Shards map[byte]bool
		}
	}
	States struct {
		PeersState        map[libp2p.ID]*peerState
		ClosestPoolsState struct {
			ShardToBeaconPool map[byte]uint64
			CrossShardPool    map[byte]map[byte]uint64
			ShardsPool        map[byte]uint64
		}
		ClosestBeaconState ChainState
		ClosestShardsState map[byte]ChainState
		sync.Mutex
	}
	BestState *BestState
	config    Config
}

func (synker *synker) Start() {
	if synker.Status.Beacon {
		return
	}
	synker.Status.Beacon = true
	synker.Status.CurrentlySyncBlks = cache.New(defaultMaxBlockSyncTime, defaultCacheCleanupTime)

	synker.Status.Lock()
	synker.startSyncRelayShards()
	synker.Status.Unlock()

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
	for _, shardID := range synker.config.RelayShards {
		if shardID > byte(synker.BestState.Beacon.ActiveShards-1) {
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
	if synker.config.NodeMode == common.NODEMODE_AUTO || synker.config.NodeMode == common.NODEMODE_SHARD {
		userRole, userShardID := synker.BestState.Beacon.GetPubkeyRole(synker.config.UserKeySet.GetPublicKeyB58(), synker.BestState.Beacon.BestBlock.Header.Round)
		if userRole == common.SHARD_ROLE && shardID == userShardID {
			return errors.New("Shard " + fmt.Sprintf("%d", shardID) + " synchronzation can't be stopped")
		}
	}
	if _, ok := synker.Status.Shards[shardID]; ok {
		if common.IndexOfByte(shardID, synker.config.RelayShards) < 0 {
			delete(synker.Status.Shards, shardID)
			return nil
		}
		return errors.New("Shard " + fmt.Sprintf("%d", shardID) + " synchronzation can't be stopped")
	}
	return errors.New("Shard " + fmt.Sprintf("%d", shardID) + " synchronzation is already stopped")
}
