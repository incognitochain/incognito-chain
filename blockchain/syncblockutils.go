package blockchain

import (
	"fmt"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/database"
	libp2p "github.com/libp2p/go-libp2p-peer"
	"github.com/patrickmn/go-cache"
)

func getBlkNeedToGetByHash(prefix string, blksHash []common.Hash, cachedItems map[string]cache.Item, peerID libp2p.ID) []common.Hash {
	var blocksNeedToGet []common.Hash
	for _, blkHash := range blksHash {
		if _, ok := cachedItems[fmt.Sprintf("%v%v", prefix, blkHash.String())]; !ok {
			blocksNeedToGet = append(blocksNeedToGet, blkHash)
		}
	}
	return blocksNeedToGet
}

func getBlkNeedToGetByHeight(prefix string, fromHeight uint64, toHeight uint64, cachedItems map[string]cache.Item, poolItems []uint64) map[uint64]uint64 {
	blocksNeedToGet := make(map[uint64]uint64)

	latestBatchBegin := uint64(0)
	for blkHeight := fromHeight; blkHeight <= toHeight; blkHeight++ {
		if exist, _ := common.SliceExists(poolItems, blkHeight); !exist {
			if _, ok := cachedItems[fmt.Sprintf("%v%v", prefix, blkHeight)]; !ok {
				if latestBatchEnd, ok := blocksNeedToGet[latestBatchBegin]; !ok {
					blocksNeedToGet[blkHeight] = blkHeight
					latestBatchBegin = blkHeight
				} else {
					if latestBatchEnd+1 == blkHeight {
						blocksNeedToGet[latestBatchBegin] = blkHeight
					} else {
						latestBatchBegin = blkHeight
					}
				}
			} else {
				latestBatchBegin = blkHeight
			}
		} else {
			latestBatchBegin = blkHeight
		}
	}
	return blocksNeedToGet
}

func getBlkNeedToGetBySpecificHeight(prefix string, blksHeight []uint64, cachedItems map[string]cache.Item, poolItems []uint64) []uint64 {
	var blocksNeedToGet []uint64
	for _, blkHeight := range blksHeight {
		if _, ok := cachedItems[fmt.Sprintf("%v%v", prefix, blkHeight)]; !ok {
			if exist, _ := common.SliceExists(poolItems, blkHeight); !exist {
				blocksNeedToGet = append(blocksNeedToGet, blkHeight)
			}
		}
	}
	return blocksNeedToGet
}

type blockType int

const (
	BeaconBlk blockType = iota
	ShardBlk
	ShardToBeaconBlk
	CrossShardBlk
)

func getBlkPrefixSyncKey(isByHash bool, blkType blockType, shardID byte, fromShard byte) string {
	key := "height"
	if isByHash {
		key = "hash"
	}
	switch blkType {
	case BeaconBlk:
		return fmt.Sprintf("%v-%v-", key, "beablk")
	case ShardBlk:
		return fmt.Sprintf("%v-%v-%v-", key, "shardblk", shardID)
	case ShardToBeaconBlk:
		return fmt.Sprintf("%v-%v-", key, "shtobblk")
	case CrossShardBlk:
		return fmt.Sprintf("%v-%v-%v-%v-", key, "crossblk", shardID, fromShard)
	default:
		return ""
	}
}

func arrayCommonElements(a, b []uint64) (c []uint64) {
	m := make(map[uint64]bool)

	for _, item := range a {
		m[item] = true
	}

	for _, item := range b {
		if _, ok := m[item]; ok {
			c = append(c, item)
		}
	}
	return
}

func GetMissingBlockInPool(
	latestHeight uint64,
	listPendingBlks []uint64,
) []uint64 {
	listBlkToSync := []uint64{}
	start := latestHeight
	for _, blkHeight := range listPendingBlks {
		for blkNeedToSync := start; blkNeedToSync < blkHeight; blkNeedToSync++ {
			listBlkToSync = append(listBlkToSync, blkNeedToSync)
			if len(listBlkToSync) >= DefaultMaxBlkReqPerPeer {
				return listBlkToSync
			}
		}
		start = blkHeight + 1
	}
	if len(listBlkToSync) == 0 {
		return nil
	}
	return listBlkToSync
}

func GetReportChainState(
	peersState map[string]*PeerState,
	report *ReportedChainState,
	shards map[byte]struct{},
	beaconBestState *BeaconBestState,
	ShBestState map[byte]ShardBestState,
) ReportedChainState {
	// Get common report, for all of node, committee, fullnode,...
	for _, peerState := range peersState {
		for shardID := range shards {
			shardBestState := ShBestState[shardID]
			if shardState, ok := peerState.Shard[shardID]; ok {
				if shardState.Height >= beaconBestState.GetBestHeightOfShard(shardID) && shardState.Height > shardBestState.ShardHeight {
					// report.ClosestShardsState[shardID].Height == shardBestState.ShardHeight
					// => this is default value => set report.ClosestShardsState[shardID].Height = *shardState
					if report.ClosestShardsState[shardID].Height == shardBestState.ShardHeight {
						report.ClosestShardsState[shardID] = *shardState
					} else {
						if shardState.Height < report.ClosestShardsState[shardID].Height {
							report.ClosestShardsState[shardID] = *shardState
						}
					}
				}
			}
		}
		if peerState.Beacon.Height > beaconBestState.BeaconHeight {
			if report.ClosestBeaconState.Height == beaconBestState.BeaconHeight {
				report.ClosestBeaconState = *peerState.Beacon
			} else {
				if peerState.Beacon.Height < report.ClosestBeaconState.Height {
					report.ClosestBeaconState = *peerState.Beacon
				}
			}
		}
	}
	return *report
}

func GetMissingBlockHashesFromPeersState(
	peersState map[string]*PeerState,
	shards map[byte]struct{},
	beaconBestStateGetter func() *BeaconBestState,
	shardBestStateGetter func(shardID byte) *ShardBestState,
) map[int][]common.Hash {
	BEACON_ID := -1
	res := map[int][]common.Hash{}
	for _, peerState := range peersState {
		if peerState.Beacon.Height == beaconBestStateGetter().BeaconHeight && !peerState.Beacon.BlockHash.IsEqual(&beaconBestStateGetter().BestBlockHash) {
			res[BEACON_ID] = append(res[BEACON_ID], peerState.Beacon.BlockHash)
		}
		for shardID := range shards {
			if shardState, ok := peerState.Shard[shardID]; ok {
				if shardState.Height == shardBestStateGetter(shardID).ShardHeight && !shardState.BlockHash.IsEqual(&shardBestStateGetter(shardID).BestBlockHash) {
					res[int(shardID)] = append(res[int(shardID)], shardState.BlockHash)
				}
			}
		}
	}
	return res
}

func GetMissingCrossShardBlock(
	db database.DatabaseInterface,
	bestCrossShardState map[byte]map[byte]uint64,
	latestValidHeight map[byte]uint64,
	userShardID byte,
) map[byte][]uint64 {
	res := map[byte][]uint64{}
	for fromShardID, start := range latestValidHeight {
		missingBlock := []uint64{}
		curHeight := start
		for {
			nextBlk, err := db.FetchCrossShardNextHeight(fromShardID, userShardID, curHeight)
			if err != nil {
				Logger.log.Errorf("Can not fetch Cross Shard Next Height formshard %v toShard %v currentHeight %v", fromShardID, userShardID, curHeight)
				break
			}
			if nextBlk <= bestCrossShardState[fromShardID][userShardID] && nextBlk > 1 {
				missingBlock = append(missingBlock, nextBlk)
				curHeight = nextBlk
			} else {
				break
			}
		}
		res[fromShardID] = missingBlock
	}
	return res
}
