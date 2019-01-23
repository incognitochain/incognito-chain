package mempool

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/ninjadotorg/constant/blockchain"
	"github.com/ninjadotorg/constant/database"
)

type ShardToBeaconPoolConfig struct {
	MaxPending uint
	MaxQueue   uint
	LifeTime   time.Duration
}
type ShardToBeaconPool struct {
	config     ShardToBeaconPoolConfig
	pending    map[byte][]blockchain.ShardToBeaconBlock
	queue      map[byte][]blockchain.ShardToBeaconBlock
	mu         sync.RWMutex
	shardState map[byte]uint64
	db         database.DatabaseInterface
}

var DefaultShardToBeaconPoolConfig = ShardToBeaconPoolConfig{
	MaxPending: 1000,
	MaxQueue:   1000,
	LifeTime:   10 * time.Hour,
}

func NewShardToBeaconPool(shardToBeaconPoolConfig ShardToBeaconPoolConfig, db database.DatabaseInterface) *ShardToBeaconPool {
	temp, err := db.FetchBeaconBestState()
	if err != nil {
		Logger.log.Error(DatabaseError, err)
		panic("Fail to get state from db")
	}
	beaconBestState := blockchain.BestStateBeacon{}
	if err := json.Unmarshal(temp, &beaconBestState); err != nil {
		Logger.log.Error(DatabaseError, err)
		panic("Can't Unmarshal beacon beststate")
	}
	pool := &ShardToBeaconPool{
		config:     shardToBeaconPoolConfig,
		pending:    make(map[byte][]blockchain.ShardToBeaconBlock),
		queue:      make(map[byte][]blockchain.ShardToBeaconBlock),
		shardState: beaconBestState.BestShardHeight,
		db:         db,
	}
	return pool
}
func (pool *ShardToBeaconPool) GetFinalBlock() map[byte][]blockchain.ShardToBeaconBlock {
	results := map[byte][]blockchain.ShardToBeaconBlock{}
	for shardID, shardItems := range pool.pending {
		results[shardID] = shardItems
	}
	return results

}

func (pool *ShardToBeaconPool) RemovePendingBlock(blockItems map[byte]uint64) error {
	if len(blockItems) <= 0 {
		Logger.log.Infof("ShardToBeaconPool: Remove Block items but got empty")
		return nil
	}
	pool.mu.Lock()
	for shardID, blockHeight := range blockItems {
		shardItems, ok := pool.pending[shardID]
		if !ok || len(shardItems) <= 0 {
			Logger.log.Infof("ShardToBeaconPool: No pending block in shard %+v exsit", shardID)
			continue
		}
		for index, block := range pool.pending[shardID] {
			if block.Header.Height <= blockHeight {
				continue
			} else {
				pool.pending[shardID] = pool.pending[shardID][index:]
				break
			}
		}
		pool.shardState[shardID] = blockHeight
	}
	pool.mu.Unlock()
	return nil
}

func (pool *ShardToBeaconPool) AddShardBeaconBlock(newBlock blockchain.ShardToBeaconBlock) error {
	Logger.log.Debugf("Current pending shard to beacon block %+v \n", pool.pending)
	blockHeader := newBlock.Header
	ShardID := blockHeader.ShardID
	Height := blockHeader.Height

	if Height == 0 {
		err := MempoolTxError{}
		err.Init(ShardToBeaconBoolError, fmt.Errorf("Add Invalid Block Heght to pool, height :%+v", 0))
		return err
	}
	pool.mu.Lock()
	// TODO validate block pool item
	beaconPoolShardItem, ok := pool.pending[ShardID]
	if beaconPoolShardItem == nil || !ok {
		beaconPoolShardItem = []blockchain.ShardToBeaconBlock{}
	}
	beaconPoolShardItem = append(beaconPoolShardItem, newBlock)
	pool.pending[ShardID] = beaconPoolShardItem
	Logger.log.Info("Update previous block items with same height")

	pool.mu.Unlock()

	return nil
}

// func UpdateBeaconPool(shardID byte, blockHeight uint64, preBlockHash common.Hash) error {
// 	if blockHeight == 0 {
// 		return errors.New("Invalid Block Heght")
// 	}
// 	if len(preBlockHash) <= 0 {
// 		return errors.New("Invalid Previous Block Hash")
// 	}
// 	shardItems, ok := beaconPool[shardID]
// 	if !ok || len(shardItems) <= 0 {
// 		log.Println("pool shard items not exists")
// 		return nil
// 	}
// 	prevBlockHeight := blockHeight - 1
// 	if prevBlockHeight < 0 {
// 		return nil
// 	}
// 	blocks, ok := shardItems[prevBlockHeight]
// 	if !ok || len(blocks) <= 0 {
// 		return nil
// 	}

// 	for _, block := range blocks {
// 		header := block.Header
// 		hash := header.Hash()
// 		if hash == preBlockHash {
// 			shardItems[prevBlockHeight] = []blockchain.ShardToBeaconBlock{block}
// 			beaconPool[shardID] = shardItems
// 			break
// 		}
// 	}

// 	return nil
// }

// func (pool *ShardToBeaconPool) GetDistinctBlockMap() map[byte]map[uint64][]common.Hash {
// 	var poolBlksMap map[byte]map[uint64][]common.Hash
// 	poolBlksMap = make(map[byte]map[uint64][]common.Hash)
// 	beaconPoolLock.Lock()
// 	defer beaconPoolLock.Unlock()
// 	for ShardId, shardItems := range beaconPool {
// 		if shardItems == nil || len(shardItems) <= 0 {
// 			continue
// 		}
// 		items := map[uint64][]common.Hash{}
// 		items = make(map[uint64][]common.Hash)
// 		for height, blks := range shardItems {
// 			for _, blk := range blks {
// 				items[height] = append(items[height], *blk.Hash())
// 			}

// 		}
// 		poolBlksMap[ShardId] = items
// 	}
// 	return poolBlksMap
// }

// func GetBeaconBlock(ShardId byte, BlockHeight uint64) (blockchain.ShardToBeaconBlock, error) {
// 	result := blockchain.ShardToBeaconBlock{}
// 	if ShardId < 0 || BlockHeight < 0 {
// 		return blockchain.ShardToBeaconBlock{}, errors.New("Invalid Shard ID or Block Heght")
// 	}
// 	shardItems, ok := beaconPool[ShardId]
// 	if shardItems == nil || !ok {
// 		return blockchain.ShardToBeaconBlock{}, errors.New("Shard not exist")
// 	}
// 	blocks, ok := shardItems[BlockHeight]
// 	if blocks == nil || len(blocks) <= 0 || !ok {
// 		return blockchain.ShardToBeaconBlock{}, errors.New("Block not exist")
// 	}

// 	result = blocks[0]
// 	return result, nil
// }
