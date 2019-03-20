package mempool

//
//import (
//	"bytes"
//	"encoding/json"
//	"errors"
//	"fmt"
//	"sort"
//	"sync"
//	"time"
//
//	"github.com/constant-money/constant-chain/blockchain"
//	"github.com/constant-money/constant-chain/cashec"
//	"github.com/constant-money/constant-chain/common"
//	"github.com/constant-money/constant-chain/database"
//)
//
//type ShardToBeaconPoolConfig struct {
//	MaxPending uint
//	MaxQueue   uint
//	LifeTime   time.Duration
//}
//
///*
//	ShardState: will be init when running node by fetching bestShardHeight in beaconBestState
//	Pending Shard To Beacon Block:
//		- Valid block
//		- Begin with current shard state in pool
//		- Can be validate with current shard committee stored in beacon best state
//	Queue Shard To Beacon Block:
//		- Valid block
//		- Greater than current shard state in pool but
//			+ Not consecutive with current shard state in pool
//			+ Not consecutive any block in pending
//		- Can't be validate with current shard committee stored in becon best state
//
//*/
//type ShardToBeaconPool struct {
//	config     ShardToBeaconPoolConfig
//	pending    map[byte][]blockchain.ShardToBeaconBlock          //executable
//	queue      map[byte]map[uint64]blockchain.ShardToBeaconBlock //non-executable
//	mu         sync.RWMutex
//	shardState map[byte]uint64
//	db         database.DatabaseInterface
//}
//
////var (
////	ReachMaxPendingError = "Pending reach limit"
////	ReachMaxQueueError   = "Queue reach limit"
////)
////var DefaultShardToBeaconPoolConfig = ShardToBeaconPoolConfig{
////	MaxPending: 1000,
////	MaxQueue:   1000,
////	LifeTime:   10 * time.Hour,
////}
//
//var shardToBeaconPool = &ShardToBeaconPool{}
//
//func InitShardToBeaconPool(shardToBeaconPoolConfig ShardToBeaconPoolConfig) *ShardToBeaconPool {
//	if shardToBeaconPool != nil {
//		return shardToBeaconPool
//	}
//	// return &ShardToBeaconPool{
//	// 	config:  shardToBeaconPoolConfig,
//	// 	pending: make(map[byte][]blockchain.ShardToBeaconBlock),
//	// 	queue:   make(map[byte]map[uint64]blockchain.ShardToBeaconBlock),
//	// }
//	shardToBeaconPool = &ShardToBeaconPool{
//		config:  shardToBeaconPoolConfig,
//		pending: make(map[byte][]blockchain.ShardToBeaconBlock),
//		queue:   make(map[byte]map[uint64]blockchain.ShardToBeaconBlock),
//	}
//	return shardToBeaconPool
//}
//func (pool *ShardToBeaconPool) SetDatabase(db database.DatabaseInterface) {
//	beaconBestState := blockchain.BestStateBeacon{}
//	temp, err := db.FetchBeaconBestState()
//	if err != nil {
//		Logger.log.Error(DatabaseError, err)
//		//TODO db is empty when there not db folder?
//		panic("Fail to get state from db")
//	} else {
//		if err := json.Unmarshal(temp, &beaconBestState); err != nil {
//			Logger.log.Error(DatabaseError, err)
//			panic("Can't Unmarshal beacon beststate")
//		}
//	}
//	if len(beaconBestState.BestShardHeight) == 0 {
//		pool.shardState = make(map[byte]uint64)
//	} else {
//		pool.shardState = beaconBestState.BestShardHeight
//	}
//	pool.db = db
//}
//func (pool *ShardToBeaconPool) GetFinalBlock() map[byte][]blockchain.ShardToBeaconBlock {
//	results := map[byte][]blockchain.ShardToBeaconBlock{}
//	for shardID, shardItems := range pool.pending {
//		results[shardID] = shardItems
//	}
//	return results
//}
//
//func (pool *ShardToBeaconPool) RemovePendingBlock(blockItems map[byte]uint64) error {
//	if len(blockItems) <= 0 {
//		Logger.log.Infof("ShardToBeaconPool: Remove Block items but got empty")
//		return nil
//	}
//	pool.mu.Lock()
//	for shardID, blockHeight := range blockItems {
//		shardItems, ok := pool.pending[shardID]
//		if !ok || len(shardItems) <= 0 {
//			Logger.log.Infof("ShardToBeaconPool: No pending block in shard %+v exsit", shardID)
//			continue
//		}
//		for index, block := range pool.pending[shardID] {
//			if block.Header.Height <= blockHeight {
//				continue
//			} else {
//				pool.pending[shardID] = pool.pending[shardID][index:]
//				break
//			}
//		}
//		pool.shardState[shardID] = blockHeight
//	}
//	pool.mu.Unlock()
//	return nil
//}
//
///*
//	Validate ShardToBeaconBlock
//	- Producer Signature
//	- Version
//	- Height > shardState of pool
//	- Height not overlap with any block in pool
//	- If prev Block Height found then compare PrevHash
//	- If prev Block Height found then compare timestamp
//	- Compare Beacon Height and Epoch
//*/
//
//func (pool *ShardToBeaconPool) ValidateShardToBeaconBlock(block blockchain.ShardToBeaconBlock) error {
//	blkHash := block.Header.Hash()
//	err := cashec.ValidateDataB58(block.Header.Producer, block.ProducerSig, blkHash.GetBytes())
//	if err != nil {
//		err := MempoolTxError{}
//		err.Init(ShardToBeaconBoolError, fmt.Errorf("Invalid Producer Signature of block height %+v in Shard %+v", block.Header.Height, block.Header.ShardID))
//		return err
//	}
//	if block.Header.Version != blockchain.VERSION {
//		err := MempoolTxError{}
//		err.Init(ShardToBeaconBoolError, fmt.Errorf("Invalid Verion of block height %+v in Shard %+v", block.Header.Height, block.Header.ShardID))
//		return err
//	}
//	//TODO: @merman review logic
//	// if block.Header.Epoch != uint64(block.Header.BeaconHeight/blockchain.EPOCH) {
//	// 	err := MempoolTxError{}
//	// 	err.Init(ShardToBeaconBoolError, fmt.Errorf("Invalid Epoch of block height %+v in Shard %+v", block.Header.Height, block.Header.ShardID))
//	// 	return err
//	// }
//	if block.Header.Height <= pool.shardState[block.Header.ShardID] {
//		err := MempoolTxError{}
//		err.Init(ShardToBeaconBoolError, fmt.Errorf("Block Height is too low, block height %+v in Shard %+v", block.Header.Height, block.Header.ShardID))
//		return err
//	}
//	if _, isOk := contains1(pool.pending[block.Header.ShardID], block.Header.Height); isOk {
//		err := MempoolTxError{}
//		err.Init(ShardToBeaconBoolError, fmt.Errorf("Block Height already in pending list, block height %+v in Shard %+v", block.Header.Height, block.Header.ShardID))
//		return err
//	}
//	if _, isOk := contains2(pool.queue[block.Header.ShardID], block.Header.Height); isOk {
//		err := MempoolTxError{}
//		err.Init(ShardToBeaconBoolError, fmt.Errorf("Block Height already in queue list, block height %+v in Shard %+v", block.Header.Height, block.Header.ShardID))
//		return err
//	}
//	if prevBlock, isOk := contains1(pool.pending[block.Header.ShardID], block.Header.Height-1); isOk {
//		if bytes.Compare(prevBlock.Hash().GetBytes(), block.Header.PrevBlockHash.GetBytes()) != 0 {
//			err := MempoolTxError{}
//			err.Init(ShardToBeaconBoolError, fmt.Errorf("Block PreviousHash not compatible with pending list, block height %+v in Shard %+v", block.Header.Height, block.Header.ShardID))
//			return err
//		}
//	}
//	if prevBlock, isOk := contains2(pool.queue[block.Header.ShardID], block.Header.Height-1); isOk {
//		if bytes.Compare(prevBlock.Hash().GetBytes(), block.Header.PrevBlockHash.GetBytes()) != 0 {
//			err := MempoolTxError{}
//			err.Init(ShardToBeaconBoolError, fmt.Errorf("Block PreviousHash not compatible with pending list, block height %+v in Shard %+v", block.Header.Height, block.Header.ShardID))
//			return err
//		}
//	}
//	return nil
//}
//func contains1(s []blockchain.ShardToBeaconBlock, e uint64) (blockchain.ShardToBeaconBlock, bool) {
//	for _, a := range s {
//		if a.Header.Height == e {
//			return a, true
//		}
//	}
//	return blockchain.ShardToBeaconBlock{}, false
//}
//func contains2(s map[uint64]blockchain.ShardToBeaconBlock, e uint64) (blockchain.ShardToBeaconBlock, bool) {
//	for _, a := range s {
//		if a.Header.Height == e {
//			return a, true
//		}
//	}
//	return blockchain.ShardToBeaconBlock{}, false
//}
//func (pool *ShardToBeaconPool) PromoteExecutable(block blockchain.ShardToBeaconBlock, committees []string) error {
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
//func (pool *ShardToBeaconPool) AddShardBeaconBlock(block blockchain.ShardToBeaconBlock, committees []string) error {
//	Logger.log.Debugf("Current pending shard to beacon block %+v \n", pool.pending)
//	blockHeight := block.Header.Height
//	shardID := block.Header.ShardID
//
//	pool.mu.Lock()
//	defer pool.mu.Unlock()
//	//Get pending shard block
//	pendingShardBlocks, ok := pool.pending[shardID]
//	if pendingShardBlocks == nil || !ok {
//		pendingShardBlocks = []blockchain.ShardToBeaconBlock{}
//
//	}
//	// if new block is expected block to be added in pending
//	// 1. Next block after shard state
//	// 2. Next block height after last block
//
//	// Pending block should be able to verified multisignature
//	err := blockchain.ValidateAggSignature(block.ValidatorsIdx, committees, block.AggregatedSig, block.R, block.Hash())
//	if err == nil {
//		if blockHeight-pool.shardState[shardID] == 1 {
//			pendingShardBlocks = append(pendingShardBlocks, block)
//			pool.pending[shardID] = pendingShardBlocks
//			pool.PromoteExecutable(block, committees)
//		} else {
//			if len(pool.pending[shardID]) > 0 {
//				if pool.pending[shardID][len(pool.pending[shardID])-1].Header.Height == blockHeight {
//					if err := pool.IsEnough(true, shardID); err == nil {
//						pendingShardBlocks = append(pendingShardBlocks, block)
//						pool.pending[shardID] = pendingShardBlocks
//						pool.PromoteExecutable(block, committees)
//					} else {
//						Logger.log.Error(ShardToBeaconBoolError, err)
//						err := MempoolTxError{}
//						err.Init(ShardToBeaconBoolError, fmt.Errorf("Error %+v in block height %+v of shard %+v", err, block.Header.Height, block.Header.ShardID))
//						return err
//					}
//				}
//			} else {
//				pool.pending[shardID] = append(pool.pending[shardID], block)
//				pool.PromoteExecutable(block, committees)
//			}
//		}
//	} else {
//		queueShardBlocks, ok := pool.queue[shardID]
//		if queueShardBlocks == nil || !ok {
//			queueShardBlocks = make(map[uint64]blockchain.ShardToBeaconBlock)
//		}
//		if err := pool.IsEnough(false, shardID); err == nil {
//			queueShardBlocks[block.Header.Height] = block
//			pool.queue[shardID] = queueShardBlocks
//		} else {
//			Logger.log.Error(ShardToBeaconBoolError, err)
//			err := MempoolTxError{}
//			err.Init(ShardToBeaconBoolError, fmt.Errorf("Error %+v in block height %+v of shard %+v", err, block.Header.Height, block.Header.ShardID))
//			return err
//		}
//	}
//	return nil
//}
//func (pool *ShardToBeaconPool) GetPendingBlockHashes() map[byte][]common.Hash {
//	pendingBlockHashes := make(map[byte][]common.Hash)
//	pool.mu.Lock()
//	defer pool.mu.Unlock()
//	for shardID, shardItems := range pool.pending {
//		if shardItems == nil || len(shardItems) <= 0 {
//			continue
//		}
//		items := []common.Hash{}
//		for _, block := range shardItems {
//			items = append(items, block.Header.Hash())
//		}
//		pendingBlockHashes[shardID] = items
//	}
//	return pendingBlockHashes
//}
//
//func (pool *ShardToBeaconPool) IsEnough(isPending bool, shardID byte) error {
//	if isPending {
//		if uint(len(pool.pending[shardID])) < pool.config.MaxPending {
//			return nil
//		} else {
//			return errors.New(ReachMaxPendingError)
//		}
//	} else {
//		if uint(len(pool.queue[shardID])) < pool.config.MaxQueue {
//			return nil
//		} else {
//			return errors.New(ReachMaxQueueError)
//		}
//	}
//}
//
//func GetShardToBeaconPoolInstance() *ShardToBeaconPool {
//	return shardToBeaconPool
//}
//func (pool *ShardToBeaconPool) GetShardToBeaconPoolState() map[byte][]uint64 {
//	result := map[byte][]uint64{}
//	pending := pool.pending
//	queue := pool.queue
//
//	poolState := map[byte]map[uint64]bool{}
//
//	for k, val := range queue {
//		if len(val) <= 0 {
//			continue
//		}
//		items := map[uint64]bool{}
//
//		for h, _ := range val {
//			items[h] = true
//		}
//		poolState[k] = items
//	}
//
//	for k, val := range pending {
//		items, ok := poolState[k]
//		if !ok || len(items) <= 0 {
//			items = map[uint64]bool{}
//		}
//
//		if len(val) <= 0 {
//			continue
//		}
//
//		for _, block := range val {
//			items[block.Header.Height] = true
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
//
//// func UpdateBeaconPool(shardID byte, blockHeight uint64, preBlockHash common.Hash) error {
//// 	if blockHeight == 0 {
//// 		return errors.New("Invalid Block Heght")
//// 	}
//// 	if len(preBlockHash) <= 0 {
//// 		return errors.New("Invalid Previous Block Hash")
//// 	}
//// 	shardItems, ok := beaconPool[shardID]
//// 	if !ok || len(shardItems) <= 0 {
//// 		log.Println("pool shard items not exists")
//// 		return nil
//// 	}
//// 	prevBlockHeight := blockHeight - 1
//// 	if prevBlockHeight < 0 {
//// 		return nil
//// 	}
//// 	blocks, ok := shardItems[prevBlockHeight]
//// 	if !ok || len(blocks) <= 0 {
//// 		return nil
//// 	}
//
//// 	for _, block := range blocks {
//// 		header := block.Header
//// 		hash := header.Hash()
//// 		if hash == preBlockHash {
//// 			shardItems[prevBlockHeight] = []blockchain.ShardToBeaconBlock{block}
//// 			beaconPool[shardID] = shardItems
//// 			break
//// 		}
//// 	}
//
//// 	return nil
//// }
//
//// func GetBeaconBlock(ShardId byte, BlockHeight uint64) (blockchain.ShardToBeaconBlock, error) {
//// 	result := blockchain.ShardToBeaconBlock{}
//// 	if ShardId < 0 || BlockHeight < 0 {
//// 		return blockchain.ShardToBeaconBlock{}, errors.New("Invalid Shard ID or Block Heght")
//// 	}
//// 	shardItems, ok := beaconPool[ShardId]
//// 	if shardItems == nil || !ok {
//// 		return blockchain.ShardToBeaconBlock{}, errors.New("Shard not exist")
//// 	}
//// 	blocks, ok := shardItems[BlockHeight]
//// 	if blocks == nil || len(blocks) <= 0 || !ok {
//// 		return blockchain.ShardToBeaconBlock{}, errors.New("Block not exist")
//// 	}
//
//// 	result = blocks[0]
//// 	return result, nil
//// }
