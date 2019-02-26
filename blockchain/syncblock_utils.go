package blockchain

import (
	"fmt"

	libp2p "github.com/libp2p/go-libp2p-peer"
	"github.com/ninjadotorg/constant/common"
	cache "github.com/patrickmn/go-cache"
)

func getBlkNeedToGetByHash(blksHash []common.Hash, cachedItems map[string]cache.Item, peerID libp2p.ID) []common.Hash {
	var blksNeedToGet []common.Hash
	for _, blkHash := range blksHash {
		if _, ok := cachedItems[blkHash.String()]; !ok {
			blksNeedToGet = append(blksNeedToGet, blkHash)
		}
	}
	return blksNeedToGet
}

func getBlkNeedToGetByHeight(fromHeight uint64, toHeight uint64, cachedItems map[string]cache.Item, peerID libp2p.ID) map[uint64]uint64 {
	blkBatchsNeedToGet := make(map[uint64]uint64)

	latestBatchBegin := uint64(0)
	for blkHeight := fromHeight; blkHeight <= toHeight; blkHeight++ {
		if _, ok := cachedItems[fmt.Sprint(blkHeight)]; !ok {
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
		} else {
			blkBatchsNeedToGet[blkHeight] = blkHeight
			latestBatchBegin = blkHeight
		}
	}
	return blkBatchsNeedToGet
}

// //GetDiffHashesOf Get unique hashes of 1st slice compare to 2nd slice
// func GetDiffHashesOf(slice1 []common.Hash, slice2 []common.Hash) []common.Hash {
// 	var diff []common.Hash

// 	for _, s1 := range slice1 {
// 		found := false
// 		for _, s2 := range slice2 {
// 			if s1 == s2 {
// 				found = true
// 				break
// 			}
// 		}
// 		if !found {
// 			diff = append(diff, s1)
// 		}
// 	}

// 	return diff
// }
