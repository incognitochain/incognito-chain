package mempool

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"sync"

	"github.com/incognitochain/incognito-chain/database"

	"github.com/incognitochain/incognito-chain/blockchain"
	"github.com/incognitochain/incognito-chain/common"
)

/*
	Cross Shard pool served as pool for cross shard block from other shard
	Each node has 256 cross shard pool, each cross shard pool has an id (shardID) (each pool responsible for one particular shard corresponding with its shardID)
	Cross Shard pool only receive cross shard block when
	- Block contain valid signature along with some condition below (AddCrossShardBlock)
	Cross Shard Pool Contains:
	- Valid pool: ordered cross shard block for each shard
	- Pending pool: un-ordered cross shard block for each shard
	- CrossShardState: highest cross shard block height confirmed by beacon committee
	Whenever new beacon best state is updated, we should validate pending pool (check order)
	Whenever new cross shard block receive, validate it, and also validate pending pool (check order)

	Ex: Cross Shard Pool with Shard ID 0
	Valid Pool:
	- contain cross shard block sent to shard 0 (any shard except shard 0 can send cross shard block to shard 0)
	- valid cross shard block map contains 256 ordered list, each list is available to be processed
	Pending Pool:
	- contain cross shard block sent to shard 0 (any shard except shard 0 can send cross shard block to shard 0)
	- pending cross shard block map contain 256 un-ordered list, each list still not available yet to be processed
	- pending cross shard block will enter valid pool until
	 + Beacon state confirm the next valid cross shard block height
*/
type CrossShardPool struct {
	shardID         byte                                   // pool shard ID
	validPool       map[byte][]*blockchain.CrossShardBlock // cross shard block from all other shard to this shard
	pendingPool     map[byte][]*blockchain.CrossShardBlock // cross shard block from all other shard to this shard
	crossShardState map[byte]uint64                        // cross shard state (marked the current state of cross shard block from all shard)
	mtx             *sync.RWMutex
	db              database.DatabaseInterface
	// When beacon chain confirm new cross shard block, it will store these block height in database
	// Cross Shard Pool using database to detect either is valid or pending
}

var crossShardPoolMap = make(map[byte]*CrossShardPool)

func InitCrossShardPool(pool map[byte]blockchain.CrossShardPool, db database.DatabaseInterface) {
	for i := 0; i < 255; i++ {
		crossShardPoolMap[byte(i)] = GetCrossShardPool(byte(i))
		pool[byte(i)] = crossShardPoolMap[byte(i)]
		crossShardPoolMap[byte(i)].db = db
	}
}

func GetCrossShardPool(shardID byte) *CrossShardPool {
	p, ok := crossShardPoolMap[shardID]
	if ok == false {
		p = new(CrossShardPool)
		p.shardID = shardID
		p.validPool = make(map[byte][]*blockchain.CrossShardBlock)
		p.pendingPool = make(map[byte][]*blockchain.CrossShardBlock)
		p.mtx = new(sync.RWMutex)
		crossShardPoolMap[shardID] = p
	}
	return p
}

// Validate pending pool again, to move pending block to valid block
// When receive new cross shard block or new beacon state arrive
func (crossShardPool *CrossShardPool) UpdatePool() map[byte]uint64 {
	crossShardPool.mtx.Lock()
	defer crossShardPool.mtx.Unlock()
	expectedHeight := crossShardPool.updatePool()
	return expectedHeight
}

func (crossShardPool *CrossShardPool) GetNextCrossShardHeight(fromShard, toShard byte, startHeight uint64) uint64 {
	nextHeight, err := crossShardPool.db.FetchCrossShardNextHeight(fromShard, toShard, startHeight)
	if err != nil {
		return 0
	}
	return nextHeight

}
func (crossShardPool *CrossShardPool) updatePool() map[byte]uint64 {
	crossShardPool.crossShardState = blockchain.GetBestStateShard(crossShardPool.shardID).BestCrossShard
	crossShardPool.removeBlockByHeight(crossShardPool.crossShardState)
	expectedHeight := make(map[byte]uint64)
	for blkShardID, blks := range crossShardPool.pendingPool {
		startHeight := crossShardPool.crossShardState[blkShardID]
		if len(crossShardPool.validPool[blkShardID]) > 0 {
			startHeight = crossShardPool.validPool[blkShardID][len(crossShardPool.validPool[blkShardID])-1].Header.Height
		}
		index := 0
		removeIndex := 0
		for _, blk := range blks {
			//only when beacon confirm (save next cross shard height), we make cross shard block valid
			//if waitHeight > blockHeight, remove that block
			waitHeight := crossShardPool.GetNextCrossShardHeight(blkShardID, crossShardPool.shardID, startHeight)
			if waitHeight > blk.Header.Height {
				removeIndex++
				index++
			} else if waitHeight == blk.Header.Height {
				index++
				startHeight = waitHeight
				continue
			} else {
				Logger.log.Info("crossshard next expectedHeight", waitHeight)
				expectedHeight[blkShardID] = waitHeight
				break
			}
		}
		if index > 0 || removeIndex > 0 {
			var valid []*blockchain.CrossShardBlock
			valid, crossShardPool.pendingPool[blkShardID] = crossShardPool.pendingPool[blkShardID][removeIndex:index], crossShardPool.pendingPool[blkShardID][index:]
			if len(valid) > 0 {
				crossShardPool.validPool[blkShardID] = append(crossShardPool.validPool[blkShardID], valid...)
			}
		}
	}

	//===============For log
	validPoolHeight := make(map[byte][]uint64)
	pendingPoolHeight := make(map[byte][]uint64)
	for shardID, blocks := range crossShardPool.validPool {
		for _, block := range blocks {
			validPoolHeight[shardID] = append(validPoolHeight[shardID], block.Header.Height)
		}
	}
	for shardID, blocks := range crossShardPool.pendingPool {
		for _, block := range blocks {
			pendingPoolHeight[shardID] = append(pendingPoolHeight[shardID], block.Header.Height)
		}
	}
	return expectedHeight
}

/*
	Validate Condition:
	1. Block come into exact destination shardID
	2. Greater than current pool cross shard state
	3. Duplicate block in pending or valid
	4. Signature
*/
func (crossShardPool *CrossShardPool) AddCrossShardBlock(crossShardBlock *blockchain.CrossShardBlock) (map[byte]uint64, byte, error) {
	crossShardPool.mtx.Lock()
	defer crossShardPool.mtx.Unlock()

	shardID := crossShardBlock.Header.ShardID
	blockHeight := crossShardBlock.Header.Height

	Logger.log.Criticalf("Receiver Block %+v from shard %+v at Cross Shard Pool \n", blockHeight, shardID)
	if crossShardBlock.ToShardID != crossShardPool.shardID {
		return nil, crossShardPool.shardID, NewBlockPoolError(WrongShardIDError, errors.New("This crossShardPool cannot receive this cross shard block, this block for another shard "+strconv.Itoa(int(crossShardBlock.Header.Height))))
	}

	//If receive old block, it will ignore
	startHeight := crossShardPool.crossShardState[shardID]
	if blockHeight <= startHeight {
		return nil, crossShardPool.shardID, NewBlockPoolError(OldBlockError, errors.New("receive old block"))
	}

	//If block already in crossShardPool, it will ignore
	for _, blkItem := range crossShardPool.validPool[shardID] {
		if blkItem.Header.Height == blockHeight {
			return nil, crossShardPool.shardID, NewBlockPoolError(DuplicateBlockError, errors.New("receive duplicate block"))
		}
	}
	for _, blkItem := range crossShardPool.pendingPool[shardID] {
		if blkItem.Header.Height == blockHeight {
			return nil, crossShardPool.shardID, NewBlockPoolError(DuplicateBlockError, errors.New("receive duplicate block"))
		}
	}
	beaconHeight, err := crossShardPool.FindBeaconHeightForCrossShardBlock(crossShardBlock.Header.BeaconHeight, shardID, blockHeight)
	if err != nil {
		return nil, crossShardPool.shardID, NewBlockPoolError(FindBeaconHeightForCrossShardBlockError, fmt.Errorf("No Beacon Block For Cross Shard Block %+v from Shard %+v", blockHeight, shardID))
	}
	shardCommitteeByte, err := crossShardPool.db.FetchShardCommitteeByHeight(beaconHeight)
	if err != nil {
		return nil, crossShardPool.shardID, NewBlockPoolError(DatabaseError, fmt.Errorf("No Committee For Cross Shard Block %+v from ShardID %+v", blockHeight, shardID))
	}
	shardCommittee := make(map[byte][]string)
	if err := json.Unmarshal(shardCommitteeByte, &shardCommittee); err != nil {
		return nil, crossShardPool.shardID, NewBlockPoolError(UnmarshalError, errors.New("Fail to unmarshal shard committee"))
	}
	if err := blockchain.ValidateAggSignature(crossShardBlock.ValidatorsIndex, shardCommittee[shardID], crossShardBlock.AggregatedSig, crossShardBlock.R, crossShardBlock.Hash()); err != nil {
		return nil, crossShardPool.shardID, err
	}

	if len(crossShardPool.pendingPool[shardID]) > maxPendingCrossShardInPool {
		if crossShardPool.pendingPool[shardID][len(crossShardPool.pendingPool[shardID])-1].Header.Height > crossShardBlock.Header.Height {
			crossShardPool.pendingPool[shardID] = crossShardPool.pendingPool[shardID][:len(crossShardPool.pendingPool[shardID])-1]
		} else {
			return nil, crossShardPool.shardID, NewBlockPoolError(MaxPoolSizeError, errors.New("Reach max pending cross shard block"))
		}
	}

	crossShardPool.pendingPool[shardID] = append(crossShardPool.pendingPool[shardID], crossShardBlock)
	sort.Slice(crossShardPool.pendingPool[shardID], func(i, j int) bool {
		return crossShardPool.pendingPool[shardID][i].Header.Height < crossShardPool.pendingPool[shardID][j].Header.Height
	})
	Logger.log.Infof("Finish Verify Cross Shard Block %+v from shard %+v \n", blockHeight, shardID)
	expectedHeight := crossShardPool.updatePool()
	return expectedHeight, crossShardPool.shardID, nil
}

func (crossShardPool *CrossShardPool) RemoveBlockByHeight(removeSinceBlkHeight map[byte]uint64) {
	crossShardPool.mtx.Lock()
	defer crossShardPool.mtx.Unlock()
	crossShardPool.removeBlockByHeight(removeSinceBlkHeight)
}

func (crossShardPool *CrossShardPool) removeBlockByHeight(removeSinceBlkHeight map[byte]uint64) {
	for shardID, blks := range crossShardPool.validPool {
		removeIndex := 0
		for _, blk := range blks {
			if blk.Header.Height <= removeSinceBlkHeight[shardID] {
				removeIndex++
				continue
			} else {
				break
			}
		}
		crossShardPool.validPool[shardID] = crossShardPool.validPool[shardID][removeIndex:]
	}

	for shardID, blks := range crossShardPool.pendingPool {
		removeIndex := 0
		for _, blk := range blks {
			if blk.Header.Height <= removeSinceBlkHeight[shardID] {
				removeIndex++
				continue
			} else {
				break
			}
		}
		crossShardPool.pendingPool[shardID] = crossShardPool.pendingPool[shardID][removeIndex:]
	}
}

func (crossShardPool *CrossShardPool) GetValidBlock(limit map[byte]uint64) map[byte][]*blockchain.CrossShardBlock {
	crossShardPool.mtx.RLock()
	defer crossShardPool.mtx.RUnlock()
	finalBlocks := make(map[byte][]*blockchain.CrossShardBlock)
	for shardID, blks := range crossShardPool.validPool {
		for _, blk := range blks {
			if limit != nil && limit[shardID] != 0 && limit[shardID] < blk.Header.Height {
				break
			}
			finalBlocks[shardID] = append(finalBlocks[shardID], blk)
		}

	}
	return finalBlocks
}

func (crossShardPool *CrossShardPool) GetValidBlockHash() map[byte][]common.Hash {
	crossShardPool.mtx.RLock()
	defer crossShardPool.mtx.RUnlock()
	finalBlockHash := make(map[byte][]common.Hash)
	for shardID, blkItems := range crossShardPool.validPool {
		for _, blk := range blkItems {
			finalBlockHash[shardID] = append(finalBlockHash[shardID], *blk.Hash())
		}
	}
	return finalBlockHash
}

func (crossShardPool *CrossShardPool) GetValidBlockHeight() map[byte][]uint64 {
	crossShardPool.mtx.RLock()
	defer crossShardPool.mtx.RUnlock()
	finalBlockHeight := make(map[byte][]uint64)
	for shardID, blkItems := range crossShardPool.validPool {
		for _, blk := range blkItems {
			finalBlockHeight[shardID] = append(finalBlockHeight[shardID], blk.Header.Height)
		}
	}
	return finalBlockHeight
}

func (crossShardPool *CrossShardPool) GetPendingBlockHeight() map[byte][]uint64 {
	crossShardPool.mtx.RLock()
	defer crossShardPool.mtx.RUnlock()
	finalBlockHeight := make(map[byte][]uint64)
	for shardID, blkItems := range crossShardPool.pendingPool {
		for _, blk := range blkItems {
			finalBlockHeight[shardID] = append(finalBlockHeight[shardID], blk.Header.Height)
		}
	}
	return finalBlockHeight
}

func (crossShardPool *CrossShardPool) GetAllBlockHeight() map[byte][]uint64 {
	crossShardPool.mtx.RLock()
	defer crossShardPool.mtx.RUnlock()
	finalBlockHeight := make(map[byte][]uint64)

	for shardID, blkItems := range crossShardPool.validPool {
		for _, blk := range blkItems {
			finalBlockHeight[shardID] = append(finalBlockHeight[shardID], blk.Header.Height)
		}
	}

	for shardID, blkItems := range crossShardPool.pendingPool {
		for _, blk := range blkItems {
			finalBlockHeight[shardID] = append(finalBlockHeight[shardID], blk.Header.Height)
		}
	}
	return finalBlockHeight
}

func (crossShardPool *CrossShardPool) GetLatestValidBlockHeight() map[byte]uint64 {
	crossShardPool.mtx.RLock()
	defer crossShardPool.mtx.RUnlock()
	finalBlockHeight := make(map[byte]uint64)
	for shardID, blkItems := range crossShardPool.validPool {
		if len(blkItems) > 0 {
			finalBlockHeight[shardID] = blkItems[len(blkItems)-1].Header.Height
		} else {
			finalBlockHeight[shardID] = 0
		}

	}
	return finalBlockHeight
}

func (crossShardPool *CrossShardPool) GetBlockByHeight(_shardID byte, height uint64) *blockchain.CrossShardBlock {
	crossShardPool.mtx.RLock()
	defer crossShardPool.mtx.RUnlock()
	for shardID, blkItems := range crossShardPool.validPool {
		if shardID != _shardID {
			continue
		}
		for _, blk := range blkItems {
			if blk.Header.Height == height {
				return blk
			}
		}
	}

	for shardID, blkItems := range crossShardPool.pendingPool {
		if shardID != _shardID {
			continue
		}
		for _, blk := range blkItems {
			if blk.Header.Height == height {
				return blk
			}
		}
	}
	return nil
}

func (crossShardPool *CrossShardPool) FindBeaconHeightForCrossShardBlock(beaconHeight uint64, fromShardID byte, crossShardBlockHeight uint64) (uint64, error) {
	for {
		beaconBlockHash, err := crossShardPool.db.GetBeaconBlockHashByIndex(beaconHeight)
		if err != nil {
			return 0, NewBlockPoolError(GetBeaconBlockHashFromDatabaseError, err)
		}
		beaconBlockBytes, err := crossShardPool.db.FetchBeaconBlock(beaconBlockHash)
		if err != nil {
			return 0, NewBlockPoolError(FetchBeaconBlockFromDatabaseError, err)
		}
		beaconBlock := blockchain.NewBeaconBlock()
		err = json.Unmarshal(beaconBlockBytes, beaconBlock)
		if err != nil {
			return 0, NewBlockPoolError(UnmarshalBeaconBlockError, err)
		}
		if shardStates, ok := beaconBlock.Body.ShardState[fromShardID]; ok {
			for _, shardState := range shardStates {
				if shardState.Height == crossShardBlockHeight {
					return beaconBlock.Header.Height, nil
				}
			}
		}
		beaconHeight += 1
	}
}
