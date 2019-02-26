package mempool

//
//import (
//	"errors"
//	"log"
//	"sort"
//	"sync"
//
//	"github.com/ninjadotorg/constant/blockchain"
//)
//
//var shardPoolLock sync.RWMutex
//
//var shardPool = map[byte]map[uint64][]blockchain.CrossShardBlock{}
//
//type CrossShardPool struct{}
//
//func (pool *CrossShardPool) GetBlock(bestStateInfos map[byte]uint64) map[byte][]blockchain.CrossShardBlock {
//	results := map[byte][]blockchain.CrossShardBlock{}
//
//	for ShardId, shardItems := range shardPool {
//		if shardItems == nil || len(shardItems) <= 0 {
//			continue
//		}
//		shardBestState, ok := bestStateInfos[ShardId]
//		if !ok || shardBestState < 0 {
//			continue
//		}
//
//		items := []blockchain.CrossShardBlock{}
//
//		for height, item := range shardItems {
//			if height > shardBestState {
//				continue
//			}
//
//			if len(item) <= 0 {
//				continue
//			}
//
//			items = append(items, item...)
//		}
//		results[ShardId] = items
//	}
//	return results
//}
//
//func (pool *CrossShardPool) RemoveBlock(blockItems map[byte]uint64) error {
//	if len(blockItems) <= 0 {
//		log.Println("Block items empty")
//		return nil
//	}
//
//	shardPoolLock.Lock()
//	for shardID, blockHeight := range blockItems {
//		shardItems, ok := shardPool[shardID]
//		if !ok || len(shardItems) <= 0 {
//			log.Println("Shard is not exist")
//			continue
//		}
//
//		items := map[uint64][]blockchain.CrossShardBlock{}
//		for i := blockHeight + 1; i < uint64(len(shardItems)); i++ {
//			items[i] = shardItems[i]
//		}
//
//		shardPool[shardID] = items
//	}
//	shardPoolLock.Unlock()
//	return nil
//}
//
//func (pool *CrossShardPool) AddCrossShardBlock(newBlock blockchain.CrossShardBlock) error {
//
//	blockHeader := newBlock.Header
//	ShardID := blockHeader.ShardID
//	Height := blockHeader.Height
//
//	if Height == 0 {
//		return errors.New("Invalid Block Heght")
//	}
//
//	shardPoolLock.Lock()
//
//	shardItems, ok := shardPool[ShardID]
//	if shardItems == nil || !ok {
//		shardItems = map[uint64][]blockchain.CrossShardBlock{}
//	}
//
//	items, ok := shardItems[Height]
//	if len(items) <= 0 || !ok {
//		items = []blockchain.CrossShardBlock{}
//	}
//	items = append(items, newBlock)
//
//	shardItems[Height] = items
//	shardPool[ShardID] = shardItems
//
//	shardPoolLock.Unlock()
//
//	return nil
//}
//
//func GetCrossShardPoolState() map[byte][]uint64 {
//	// map[byte]map[uint64][]blockchain.CrossShardBlock{}
//	result := map[byte][]uint64{}
//
//	poolState := map[byte]map[uint64]bool{}
//
//	for k, val := range shardPool {
//		items, ok := poolState[k]
//		if !ok || len(items) <= 0 {
//			items = map[uint64]bool{}
//		}
//
//		if len(val) <= 0 {
//			continue
//		}
//
//		for h, _ := range val {
//			items[h] = true
//		}
//		poolState[k] = items
//	}
//
//	for k, val := range poolState {
//		items := []uint64{}
//		for h, _ := range val {
//			items = append(items, h)
//		}
//		sort.Slice(items, func(i, j int) bool {
//			return items[i] < items[j]
//		})
//
//		result[k] = items
//	}
//	return result
//}
