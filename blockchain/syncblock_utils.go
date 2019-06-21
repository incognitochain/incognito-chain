package blockchain

import (
	"fmt"

	"github.com/incognitochain/incognito-chain/common"
	libp2p "github.com/libp2p/go-libp2p-peer"
	cache "github.com/patrickmn/go-cache"
)

func getBlkNeedToGetByHash(prefix string, blksHash []common.Hash, cachedItems map[string]cache.Item, peerID libp2p.ID) []common.Hash {
	var blksNeedToGet []common.Hash
	for _, blkHash := range blksHash {
		if _, ok := cachedItems[blkHash.String()]; !ok {
			blksNeedToGet = append(blksNeedToGet, blkHash)
		}
	}
	return blksNeedToGet
}

func getBlkNeedToGetByHeight(prefix string, fromHeight uint64, toHeight uint64, cachedItems map[string]cache.Item, poolItems []uint64, peerID libp2p.ID) map[uint64]uint64 {
	blkBatchsNeedToGet := make(map[uint64]uint64)

	latestBatchBegin := uint64(0)
	for blkHeight := fromHeight; blkHeight <= toHeight; blkHeight++ {
		if exist, _ := common.SliceExists(poolItems, blkHeight); !exist {
			if _, ok := cachedItems[fmt.Sprint(blkHeight)]; !ok {
				if latestBatchEnd, ok := blkBatchsNeedToGet[latestBatchBegin]; !ok {
					blkBatchsNeedToGet[blkHeight] = blkHeight
					latestBatchBegin = blkHeight
				} else {
					if latestBatchEnd+1 == blkHeight {
						blkBatchsNeedToGet[latestBatchBegin] = blkHeight
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
	return blkBatchsNeedToGet
}

func getBlkNeedToGetBySpecificHeight(prefix string, blksHeight []uint64, cachedItems map[string]cache.Item, poolItems []uint64, peerID libp2p.ID) []uint64 {
	var blksNeedToGet []uint64
	for _, blkHeight := range blksHeight {
		if _, ok := cachedItems[fmt.Sprint(blkHeight)]; !ok {
			if exist, _ := common.SliceExists(poolItems, blkHeight); !exist {
				blksNeedToGet = append(blksNeedToGet, blkHeight)
			}
		}
	}
	return blksNeedToGet
}

type blockType int

const (
	beaconBlk blockType = iota
	shardBlk
	shardToBeaconBlk
	crossShardBlk
)

func getBlkPrefixSyncKey(isByHash bool, blkType blockType, shardID byte, fromShard byte) string {
	key := "height"
	if isByHash {
		key = "hash"
	}
	switch blkType {
	case beaconBlk:
		return fmt.Sprintf("%v-%v-", key, "beablk")
	case shardBlk:
		return fmt.Sprintf("%v-%v-%v-", key, "shardblk", shardID)
	case shardToBeaconBlk:
		return fmt.Sprintf("%v-%v-", key, "shtobblk")
	case crossShardBlk:
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
