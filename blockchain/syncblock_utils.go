package blockchain

import (
	"fmt"

	"github.com/incognitochain/incognito-chain/common"
	libp2p "github.com/libp2p/go-libp2p-peer"
	"github.com/patrickmn/go-cache"
)

func getBlkNeedToGetByHash(prefix string, blksHash []common.Hash, cachedItems map[string]cache.Item, peerID libp2p.ID) []common.Hash {
	var blocksNeedToGet []common.Hash
	for _, blkHash := range blksHash {
		if _, ok := cachedItems[blkHash.String()]; !ok {
			blocksNeedToGet = append(blocksNeedToGet, blkHash)
		}
	}
	return blocksNeedToGet
}

func getBlkNeedToGetByHeight(prefix string, fromHeight uint64, toHeight uint64, cachedItems map[string]cache.Item, poolItems []uint64, peerID libp2p.ID) map[uint64]uint64 {
	blocksNeedToGet := make(map[uint64]uint64)

	latestBatchBegin := uint64(0)
	for blkHeight := fromHeight; blkHeight <= toHeight; blkHeight++ {
		if exist, _ := common.SliceExists(poolItems, blkHeight); !exist {
			if _, ok := cachedItems[fmt.Sprint(blkHeight)]; !ok {
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

func getBlkNeedToGetBySpecificHeight(prefix string, blksHeight []uint64, cachedItems map[string]cache.Item, poolItems []uint64, peerID libp2p.ID) []uint64 {
	var blocksNeedToGet []uint64
	for _, blkHeight := range blksHeight {
		if _, ok := cachedItems[fmt.Sprint(blkHeight)]; !ok {
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
