package mempool

import (
	"github.com/ninjadotorg/constant/blockchain"
	"sync"
)

/*
	This pool will contains blocks from other nodes
	but not yet insert into blockchain
	- blocks from all shard
	- blocks from beacon chain
*/
var nodeBeaconPoolLock sync.RWMutex
var nodeShardPoolLock sync.RWMutex

var nodeShardPool = map[byte]map[uint64][]blockchain.ShardBlock{}
var nodeBeaconPool = map[uint64][]blockchain.BeaconBlock{}

/*
	ShardState:
		- Height of newest shard block in blockchain
	 	- will be init when running node by fetching bestShardHeight in beaconbestState and shardbeststate
	Pending Shard Blocks List
		- Valid block
		- Begin with current shard state in pool
		- Can be validate with current shard committee stored in beacon best state
	Queue Shard Blocks Map
		- Valid block
		- Greater than current shard state in pool but
			+ Not consecutive with current shard state in pool
			+ Not consecutive any block in pending
		- Can't be validate with current shard committee stored in becon best state
*/
//type NodeShardPool struct {
//	config     NodeShardPoolConfig
//	pending    map[byte][]blockchain.ShardBlock
//	queue      map[byte]map[uint64]blockchain.ShardBlock
//	shardState map[byte]uint64
//	db         database.DatabaseInterface
//	mu         sync.RWMutex
//}
//
////Other information will be config
//type NodeShardPoolConfig struct {
//	MaxPending uint
//	MaxQueue   uint
//	LifeTime   time.Duration
//}
//
//var DefaultNodeShardPoolConfig = NodeShardPoolConfig{
//	MaxPending: 1000,
//	MaxQueue:   1000,
//	LifeTime:   10 * time.Hour,
//}
//
//func (pool *NodeShardPool) AddShardBlock(block blockchain.ShardBlock, committees []string) error {
//	blockHeader := block.Header
//	shardID := blockHeader.ShardID
//	blockHeight := blockHeader.Height
//	if blockHeight == 0 {
//		return errors.New("Invalid Block Heght")
//	}
//
//	pool.mu.Lock()
//	defer pool.mu.Unlock()
//
//	//Get pending shard block
//	pendingShardBlocks, ok := pool.pending[shardID]
//	if pendingShardBlocks == nil || !ok {
//		pendingShardBlocks = []blockchain.ShardBlock{}
//	}
//
//	//TODO:
//	// 1. Validate shard block
//	// 2. Validate agg sig for current shard committees
//	err := blockchain.ValidateAggSignature(block.ValidatorsIdx, committees, block.AggregatedSig, block.R, block.Hash())
//	if err != nil {
//		return err
//	} else {
//		// in this circumstance, list pending is empty
//		if blockHeight-pool.shardState[shardID] == 1 {
//			pendingShardBlocks = append(pendingShardBlocks, block)
//			pool.pending[shardID] = pendingShardBlocks
//			pool.PromoteExecutable(block, committees)
//			// in this circumstance, list pending is not empty
//		} else if pool.pending[shardID][len(pool.pending[shardID])-1].Header.Height == blockHeight {
//			if err := pool.IsEnough(true, shardID); err == nil {
//				pendingShardBlocks = append(pendingShardBlocks, block)
//				pool.pending[shardID] = pendingShardBlocks
//				pool.PromoteExecutable(block, committees)
//			} else {
//				Logger.log.Error(ShardToBeaconBoolError, err)
//				err := MempoolTxError{}
//				err.Init(ShardToBeaconBoolError, fmt.Errorf("Error %+v in block height %+v of shard %+v", err, block.Header.Height, block.Header.ShardID))
//				return err
//			}
//		}
//	}
//
//	return nil
//}
//func (pool *NodeShardPool) PromoteExecutable(block blockchain.ShardBlock, committees []string) error {
//	shardID := block.Header.ShardID
//	lastBlockHeight := pool.pending[shardID][len(pool.pending[shardID])-1].Header.Height
//	for {
//		newBlock, isHas := pool.queue[shardID][lastBlockHeight+1]
//		if isHas {
//			err := blockchain.ValidateAggSignature(newBlock.ValidatorsIdx, committees, newBlock.AggregatedSig, newBlock.R, newBlock.Hash())
//			if err == nil {
//				if err := pool.IsEnough(true, shardID); err == nil {
//					pool.pending[shardID] = append(pool.pending[shardID], newBlock)
//					delete(pool.queue[shardID], lastBlockHeight+1)
//				} else {
//					break
//				}
//			} else {
//				break
//			}
//		} else {
//			break
//		}
//		lastBlockHeight++
//	}
//	return nil
//}
//func (pool *NodeShardPool) IsEnough(isPending bool, shardID byte) error {
//	if isPending {
//		if uint(len(pool.pending[shardID])) <= pool.config.MaxPending {
//			return nil
//		} else {
//			return errors.New("")
//		}
//	} else {
//		if uint(len(pool.queue[shardID])) <= pool.config.MaxQueue {
//			return nil
//		} else {
//			return errors.New("")
//		}
//	}
//}
//
//func (pool *NodeShardPool) GetShardBlocks(shardID byte) ([]blockchain.ShardBlock, error) {
//	pool.mu.Lock()
//	defer pool.mu.Unlock()
//	shardBlocks := []blockchain.ShardBlock{}
//	for _, block := range pool.pending[shardID] {
//		shardBlocks = append(shardBlocks, block)
//	}
//	return shardBlocks, nil
//}
//func (pool *NodeShardPool) RemoveShardBlocks(shardID byte, blockHeight uint64) error {
//	pool.mu.Lock()
//	defer pool.mu.Unlock()
//
//	if blockHeight <= pool.shardState[shardID] {
//		Logger.log.Infof("NodeShardPool %+v: Fail to remove in pool beacause blocks is too old ", shardID)
//		return nil
//	}
//	for index, block := range pool.pending[shardID] {
//		if block.Header.Height <= blockHeight {
//			continue
//		} else {
//			pool.pending[shardID] = pool.pending[shardID][index:]
//			break
//		}
//	}
//	pool.shardState[shardID] = blockHeight
//	return nil
//}
//
//func (pool *NodeShardPool) PushBlock(block blockchain.ShardBlock) error {
//	blockHeader := block.Header
//	shardID := blockHeader.ShardID
//	height := blockHeader.Height
//	if height == 0 {
//		return errors.New("Invalid Block Heght")
//	}
//
//	nodeShardPoolLock.Lock()
//	shardItems := nodeShardPool[shardID]
//	if shardItems == nil {
//		shardItems = make(map[uint64][]blockchain.ShardBlock)
//	}
//	shardItems[height] = append(shardItems[height], block)
//	nodeShardPool[shardID] = shardItems
//	nodeShardPoolLock.Unlock()
//
//	return nil
//}
//
//func (pool *NodeShardPool) GetBlocks(shardID byte, blockHeight uint64) ([]blockchain.ShardBlock, error) {
//
//	if blockHeight == 0 {
//		return []blockchain.ShardBlock{}, errors.New("Invalid ShardId or block Height")
//	}
//	shardItems := nodeShardPool[shardID]
//
//	return shardItems[blockHeight], nil
//}
//
//func (pool *NodeShardPool) RemoveBlocks(shardID byte, blockHeight uint64) error {
//	if shardID <= 0 || blockHeight == 0 {
//		return errors.New("Invalid ShardId or block Height")
//	}
//
//	nodeShardPoolLock.Lock()
//	shardItems := nodeShardPool[shardID]
//	delete(shardItems, blockHeight)
//	nodeShardPool[shardID] = shardItems
//	nodeShardPoolLock.Unlock()
//
//	return nil
//}
//
//type NodeBeaconPool struct{}

//func (pool *NodeBeaconPool) PushBlock(block blockchain.BeaconBlock) error {
//
//	blockHeader := block.Header
//	height := blockHeader.Height
//	if height == 0 {
//		return errors.New("Invalid Block Heght")
//	}
//
//	nodeBeaconPoolLock.Lock()
//	defer nodeBeaconPoolLock.Unlock()
//	if _, ok := nodeBeaconPool[height]; ok {
//		isNew := true
//		for _, poolblk := range nodeBeaconPool[height] {
//			if poolblk.Hash() == block.Hash() {
//				isNew = false
//				return nil
//			}
//		}
//		if isNew {
//			nodeBeaconPool[height] = append(nodeBeaconPool[height], block)
//		}
//	} else {
//		nodeBeaconPool[height] = append(nodeBeaconPool[height], block)
//	}
//	return nil
//}
//
//func (pool *NodeBeaconPool) GetBlocks(blockHeight uint64) ([]blockchain.BeaconBlock, error) {
//	return nodeBeaconPool[blockHeight], nil
//}
//
//func (pool *NodeBeaconPool) RemoveBlocks(blockHeight uint64) error {
//	delete(nodeBeaconPool, blockHeight)
//	return nil
//}
