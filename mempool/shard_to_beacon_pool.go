package mempool

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/ninjadotorg/constant/blockchain"
	"github.com/ninjadotorg/constant/cashec"
	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/database"
)

type ShardToBeaconPoolConfig struct {
	MaxPending uint
	MaxQueue   uint
	LifeTime   time.Duration
}

/*
	ShardState: will be init when running node by fetching bestShardHeight in beaconBestState
	Pending Shard To Beacon Block:
		- Valid block
		- Begin with current shard state in pool
		- Can be validate with current shard committee stored in beacon best state
	Queue Shard To Beacon Block:
		- Valid block
		- Greater than current shard state in pool but
			+ Not consecutive with current shard state in pool
			+ Not consecutive any block in pending
		- Can't be validate with current shard committee stored in becon best state

*/
type ShardToBeaconPool struct {
	config     ShardToBeaconPoolConfig
	pending    map[byte][]blockchain.ShardToBeaconBlock //executable
	queue      map[byte][]blockchain.ShardToBeaconBlock //non-executable
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

/*
	Validate ShardToBeaconBlock
	- Producer Signature
	- Version
	- Height > shardState of pool
	- Height not overlap with any block in pool
	- If prev Block Height found then compare PrevHash
	- If prev Block Height found then compare timestamp
	- Compare Beacon Height and Epoch
*/

func (pool *ShardToBeaconPool) ValidateShardToBeaconBlock(block blockchain.ShardToBeaconBlock) error {
	blkHash := block.Header.Hash()
	err := cashec.ValidateDataB58(block.Header.Producer, block.ProducerSig, blkHash.GetBytes())
	if err != nil {
		err := MempoolTxError{}
		err.Init(ShardToBeaconBoolError, fmt.Errorf("Invalid Producer Signature of block height %+v in Shard %+v", block.Header.Height, block.Header.ShardID))
		return err
	}
	if block.Header.Version != blockchain.VERSION {
		err := MempoolTxError{}
		err.Init(ShardToBeaconBoolError, fmt.Errorf("Invalid Verion of block height %+v in Shard %+v", block.Header.Height, block.Header.ShardID))
		return err
	}
	if block.Header.Epoch != uint64(block.Header.BeaconHeight/blockchain.EPOCH) {
		err := MempoolTxError{}
		err.Init(ShardToBeaconBoolError, fmt.Errorf("Invalid Epoch of block height %+v in Shard %+v", block.Header.Height, block.Header.ShardID))
		return err
	}
	if block.Header.Height <= pool.shardState[block.Header.ShardID] {
		err := MempoolTxError{}
		err.Init(ShardToBeaconBoolError, fmt.Errorf("Block Height is too low, block height %+v in Shard %+v", block.Header.Height, block.Header.ShardID))
		return err
	}
	if _, isOk := contains(pool.pending[block.Header.ShardID], block.Header.Height); isOk {
		err := MempoolTxError{}
		err.Init(ShardToBeaconBoolError, fmt.Errorf("Block Height already in pending list, block height %+v in Shard %+v", block.Header.Height, block.Header.ShardID))
		return err
	}
	if _, isOk := contains(pool.queue[block.Header.ShardID], block.Header.Height); isOk {
		err := MempoolTxError{}
		err.Init(ShardToBeaconBoolError, fmt.Errorf("Block Height already in queue list, block height %+v in Shard %+v", block.Header.Height, block.Header.ShardID))
		return err
	}
	if prevBlock, isOk := contains(pool.pending[block.Header.ShardID], block.Header.Height-1); isOk {
		if bytes.Compare(prevBlock.Hash().GetBytes(), block.Header.PrevBlockHash.GetBytes()) != 0 {
			err := MempoolTxError{}
			err.Init(ShardToBeaconBoolError, fmt.Errorf("Block PreviousHash not compatible with pending list, block height %+v in Shard %+v", block.Header.Height, block.Header.ShardID))
			return err
		}
	}
	if prevBlock, isOk := contains(pool.queue[block.Header.ShardID], block.Header.Height-1); isOk {
		if bytes.Compare(prevBlock.Hash().GetBytes(), block.Header.PrevBlockHash.GetBytes()) != 0 {
			err := MempoolTxError{}
			err.Init(ShardToBeaconBoolError, fmt.Errorf("Block PreviousHash not compatible with pending list, block height %+v in Shard %+v", block.Header.Height, block.Header.ShardID))
			return err
		}
	}
	return nil
}
func contains(s []blockchain.ShardToBeaconBlock, e uint64) (blockchain.ShardToBeaconBlock, bool) {
	for _, a := range s {
		if a.Header.Height == e {
			return a, true
		}
	}
	return blockchain.ShardToBeaconBlock{}, false
}

func (pool *ShardToBeaconPool) AddShardBeaconBlock(block blockchain.ShardToBeaconBlock) error {
	Logger.log.Debugf("Current pending shard to beacon block %+v \n", pool.pending)
	blockHeight := block.Header.Height
	shardID := block.Header.ShardID

	pool.mu.Lock()
	defer pool.mu.Unlock()
	pendingShardBlocks, ok := pool.pending[shardID]
	if pendingShardBlocks == nil || !ok {
		pendingShardBlocks = []blockchain.ShardToBeaconBlock{}

	}
	if blockHeight-pool.shardState[shardID] == 1 {
		pendingShardBlocks = append(pendingShardBlocks, block)
		pool.pending[shardID] = pendingShardBlocks
		// TODO: Promote executbale block from queue to pending
	} else if _, isOk := contains(pool.pending[shardID], blockHeight-1); isOk {
		pendingShardBlocks = append(pendingShardBlocks, block)
		pool.pending[shardID] = pendingShardBlocks
		// TODO: Promote executbale block from queue to pending
	} else {
		queueShardBlocks, ok := pool.pending[shardID]
		if queueShardBlocks == nil || !ok {
			queueShardBlocks = []blockchain.ShardToBeaconBlock{}
		}
		queueShardBlocks = append(queueShardBlocks, block)
		pool.queue[shardID] = queueShardBlocks
		// TODO: find out the structure of queue
	}
	return nil
}
func (pool *ShardToBeaconPool) GetPendingBlockHashes() map[byte][]common.Hash {
	pendingBlockHashes := make(map[byte][]common.Hash)
	pool.mu.Lock()
	defer pool.mu.Unlock()
	for shardID, shardItems := range pool.pending {
		if shardItems == nil || len(shardItems) <= 0 {
			continue
		}
		items := []common.Hash{}
		for _, block := range shardItems {
			items = append(items, block.Header.Hash())
		}
		pendingBlockHashes[shardID] = items
	}
	return pendingBlockHashes
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
