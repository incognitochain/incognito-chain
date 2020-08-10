package mempool

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/incognitochain/incognito-chain/blockchain"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/rawdbv2"
	"github.com/incognitochain/incognito-chain/incdb"
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
	// When beacon chain confirm new cross shard block, it will store these block height in database
	// Cross Shard Pool using database to detect either is valid or pending
	db     incdb.Database
	bc     *blockchain.BlockChain
	isTest bool
	// Store shard blocks from a particular shard confirmed in a beacon block
	//(mapping from (shardID, beaconHeight) ==> confirmed shard height via shard state/beacon body)
	confirmedHeight map[heightPair][]uint64
}

type heightPair struct {
	Height  uint64
	ShardID byte
}

var crossShardPoolMap = make(map[byte]*CrossShardPool)

func InitCrossShardPool(pool map[byte]blockchain.CrossShardPool, db incdb.Database, bc *blockchain.BlockChain) {
	for i := 0; i < bc.BestState.Beacon.ActiveShards; i++ {
		crossShardPoolMap[byte(i)] = GetCrossShardPool(byte(i))
		pool[byte(i)] = crossShardPoolMap[byte(i)]
		crossShardPoolMap[byte(i)].db = db
		crossShardPoolMap[byte(i)].bc = bc
		go func(pool blockchain.CrossShardPool) {
			for {
				time.Sleep(time.Minute)
				pool.UpdatePool()
			}
		}(pool[byte(i)])
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
		p.isTest = false
		p.confirmedHeight = make(map[heightPair][]uint64)
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
	nextHeight, err := rawdbv2.GetCrossShardNextHeight(crossShardPool.db, fromShard, toShard, startHeight)
	if err != nil {
		return 0
	}
	return nextHeight

}
func (crossShardPool *CrossShardPool) RevertCrossShardPool(latestValidHeight uint64) {
	crossShardPool.mtx.Lock()
	defer crossShardPool.mtx.Unlock()
	Logger.log.Infof("Begin Revert CrossShardPool of Shard %+v with latest valid height %+v", crossShardPool.shardID, latestValidHeight)
	crossShardBlocks := []*blockchain.CrossShardBlock{}
	if _, ok := crossShardPool.validPool[crossShardPool.shardID]; ok {
		for _, crossShardBlock := range crossShardPool.validPool[crossShardPool.shardID] {
			crossShardBlocks = append(crossShardBlocks, crossShardBlock)
		}
		crossShardPool.validPool[crossShardPool.shardID] = []*blockchain.CrossShardBlock{}
		for _, crossShardBlock := range crossShardBlocks {
			_, _, err := crossShardPool.addCrossShardBlock(crossShardBlock)
			if err == nil {
				continue
			} else {
				return
			}
		}
	} else {
		return
	}
}

/*
	Validate Cross Shard Block Before Signature Validation

*/

func (crossShardPool *CrossShardPool) addCrossShardBlock(crossShardBlock *blockchain.CrossShardBlock) (map[byte]uint64, byte, error) {
	shardID := crossShardBlock.Header.ShardID
	blockHeight := crossShardBlock.Header.Height

	Logger.log.Infof("Receiver Block %+v from shard %+v at Cross Shard Pool \n", blockHeight, shardID)
	// validate cross shard block
	if err := crossShardPool.validateCrossShardBlockBeforeSignatureValidation(crossShardBlock); err != nil {
		return nil, crossShardPool.shardID, err
	}
	if len(crossShardPool.pendingPool[shardID]) > maxPendingCrossShardInPool {
		if crossShardPool.pendingPool[shardID][len(crossShardPool.pendingPool[shardID])-1].Header.Height > crossShardBlock.Header.Height {
			crossShardPool.pendingPool[shardID] = crossShardPool.pendingPool[shardID][:len(crossShardPool.pendingPool[shardID])-1]
		} else {
			return nil, crossShardPool.shardID, NewBlockPoolError(MaxPoolSizeError, fmt.Errorf("Reach max pending cross shard block %+v", maxPendingCrossShardInPool))
		}
	}
	// check pool size
	crossShardPool.pendingPool[shardID] = append(crossShardPool.pendingPool[shardID], crossShardBlock)
	sort.Slice(crossShardPool.pendingPool[shardID], func(i, j int) bool {
		return crossShardPool.pendingPool[shardID][i].Header.Height < crossShardPool.pendingPool[shardID][j].Header.Height
	})
	Logger.log.Infof("Finish Verify and Add To Pending Pool Cross Shard Block %+v from shard %+v \n", blockHeight, shardID)
	// update pool states
	expectedHeight := crossShardPool.updatePool()
	return expectedHeight, crossShardPool.shardID, nil
}
func (crossShardPool *CrossShardPool) AddCrossShardBlock(crossShardBlock *blockchain.CrossShardBlock) (map[byte]uint64, byte, error) {
	crossShardPool.mtx.Lock()
	defer crossShardPool.mtx.Unlock()
	return crossShardPool.addCrossShardBlock(crossShardBlock)
}

/*
	validateCrossShardBlock with these following conditions:
	1. Destination Shard of Cross Shard Block match Cross Shard Pool ID
	2. Check Cross Shard BLock Height With Current Cross Shard Pool State
	3. Check Shard Block existence in Pool or Blockchain by Cross Shard Block Height
*/
func (crossShardPool *CrossShardPool) validateCrossShardBlockBeforeSignatureValidation(crossShardBlock *blockchain.CrossShardBlock) error {
	if crossShardBlock.ToShardID != crossShardPool.shardID {
		return NewBlockPoolError(WrongShardIDError, errors.New("This crossShardPool cannot receive this cross shard block, this block for another shard "+strconv.Itoa(int(crossShardBlock.Header.Height))))
	}
	//If receive old block, it will ignore
	startHeight := crossShardPool.crossShardState[crossShardBlock.Header.ShardID]
	if crossShardBlock.Header.Height <= startHeight {
		return NewBlockPoolError(OldBlockError, errors.New("receive old block"))
	}
	//If block already in crossShardPool, it will ignore
	for _, blkItem := range crossShardPool.validPool[crossShardBlock.Header.ShardID] {
		if blkItem.Header.Height == crossShardBlock.Header.Height {
			return NewBlockPoolError(DuplicateBlockError, errors.New("receive duplicate block"))
		}
	}
	for _, blkItem := range crossShardPool.pendingPool[crossShardBlock.Header.ShardID] {
		if blkItem.Header.Height == crossShardBlock.Header.Height {
			return NewBlockPoolError(DuplicateBlockError, errors.New("receive duplicate block"))
		}
	}
	return nil
}

// Validate Agg Signature of Cross Shard Block
func (crossShardPool *CrossShardPool) validateCrossShardBlock(crossShardBlock *blockchain.CrossShardBlock) error {
	if crossShardPool.isTest {
		return nil
	}
	Logger.log.Infof("Begin Verify Cross Shard Block from shard %+v, height %+v", crossShardBlock.Header.ShardID, crossShardBlock.Header.Height)
	beaconHeight, err := crossShardPool.findBeaconHeightForCrossShardBlock(crossShardBlock.Header.BeaconHeight, crossShardBlock.Header.ShardID, crossShardBlock.Header.Height)
	if err != nil {
		return NewBlockPoolError(ValidateAggSignatureForCrossShardBlockError, err)
	}
	beaconHeight = beaconHeight - 1
	consensusStateRootHash, err := crossShardPool.bc.GetBeaconConsensusRootHash(crossShardPool.bc.GetDatabase(), beaconHeight)
	if err != nil {
		return NewBlockPoolError(ValidateAggSignatureForCrossShardBlockError, fmt.Errorf("Can't found ConsensusStateRootHash of beacon height %+v, error %+v", beaconHeight, err))
	}
	consensusStateDB, err := statedb.NewWithPrefixTrie(consensusStateRootHash, statedb.NewDatabaseAccessWarper(crossShardPool.bc.GetDatabase()))
	if err != nil {
		return NewBlockPoolError(ValidateAggSignatureForCrossShardBlockError, err)
	}
	shardCommittee := statedb.GetOneShardCommittee(consensusStateDB, crossShardBlock.Header.ShardID)
	if err := crossShardBlock.VerifyCrossShardBlock(crossShardPool.bc, shardCommittee); err != nil {
		return NewBlockPoolError(ValidateAggSignatureForCrossShardBlockError, err)
	}
	Logger.log.Infof("Finish Verify Successfully Cross Shard Block from shard %+v, height %+v", crossShardBlock.Header.ShardID, crossShardBlock.Header.Height)
	return nil
}

/*
	updatePool: promote cross shard block from pending pool to valid pool:
	Promoted cross shard block must pass these following condition:
	- Validate Agg Signature
	- Beacon Confirm this Cross Shard Block and Store in Database
*/
func (crossShardPool *CrossShardPool) updatePool() map[byte]uint64 {
	Logger.log.Debugf("Update Cross Shard Pool %+v State", crossShardPool.shardID)
	crossShardPool.crossShardState = blockchain.GetBestStateShard(crossShardPool.shardID).BestCrossShard
	crossShardPool.removeBlockByHeight(crossShardPool.crossShardState)
	expectedHeight := make(map[byte]uint64)
	for blkShardID, crossShardBlocks := range crossShardPool.pendingPool {
		startHeight := crossShardPool.crossShardState[blkShardID]
		if len(crossShardPool.validPool[blkShardID]) > 0 {
			startHeight = crossShardPool.validPool[blkShardID][len(crossShardPool.validPool[blkShardID])-1].Header.Height
		}
		index := 0
		removeIndex := 0
		for _, crossShardBlock := range crossShardBlocks {
			// verify cross shard block signature
			if err := crossShardPool.validateCrossShardBlock(crossShardBlock); err != nil {
				Logger.log.Error(err)
				break
			}
			//only when beacon confirm (save next cross shard height), we make cross shard block valid
			//if waitHeight > blockHeight, remove that block
			waitHeight := crossShardPool.GetNextCrossShardHeight(blkShardID, crossShardPool.shardID, startHeight)
			if waitHeight > crossShardBlock.Header.Height {
				removeIndex++
				index++
			} else if waitHeight == crossShardBlock.Header.Height {
				index++
				startHeight = waitHeight
				continue
			} else {
				Logger.log.Infof("Get Cross Shard Block %+v from ShardID %+v But next expected is height", crossShardBlock.Header.Height, crossShardBlock.Header.ShardID, waitHeight)
				expectedHeight[blkShardID] = waitHeight
				break
			}
		}
		if index > 0 || removeIndex > 0 {
			var validCrossShardBlocks []*blockchain.CrossShardBlock
			validCrossShardBlocks, crossShardPool.pendingPool[blkShardID] = crossShardPool.pendingPool[blkShardID][removeIndex:index], crossShardPool.pendingPool[blkShardID][index:]
			if len(validCrossShardBlocks) > 0 {
				crossShardPool.validPool[blkShardID] = append(crossShardPool.validPool[blkShardID], validCrossShardBlocks...)
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

func (crossShardPool *CrossShardPool) findBeaconHeightForCrossShardBlock(beaconHeight uint64, shardID byte, crossShardBlockHeight uint64) (uint64, error) {
	for {
		p := heightPair{
			Height:  beaconHeight,
			ShardID: shardID,
		}
		if _, ok := crossShardPool.confirmedHeight[p]; !ok {
			beaconBlockHashes, err := rawdbv2.GetBeaconBlockHashByIndex(crossShardPool.db, beaconHeight)
			if err != nil {
				return 0, NewBlockPoolError(FindBeaconHeightForCrossShardBlockError, err)
			}
			beaconBlockHash := beaconBlockHashes[0]
			beaconBlockBytes, err := rawdbv2.GetBeaconBlockByHash(crossShardPool.db, beaconBlockHash)
			if err != nil {
				return 0, NewBlockPoolError(FetchBeaconBlockFromDatabaseError, err)
			}
			beaconBlock := blockchain.NewBeaconBlock()
			err = json.Unmarshal(beaconBlockBytes, beaconBlock)
			if err != nil {
				return 0, NewBlockPoolError(FindBeaconHeightForCrossShardBlockError, err)
			}
			heights := []uint64{}
			if shardStates, ok := beaconBlock.Body.ShardState[shardID]; ok {
				for _, shardState := range shardStates {
					heights = append(heights, shardState.Height)
				}
			}
			crossShardPool.confirmedHeight[p] = heights
		}
		// Cache confirmed heights
		heights := crossShardPool.confirmedHeight[p]
		for _, h := range heights {
			if h == crossShardBlockHeight {
				return beaconHeight, nil
			}
		}
		beaconHeight += 1
	}
}

func (crossShardPool *CrossShardPool) FindBeaconHeightForCrossShardBlock(beaconHeight uint64, fromShardID byte, crossShardBlockHeight uint64) (uint64, error) {
	Logger.log.Infof("FindBeaconHeightForCrossShardBlock %d lock", crossShardPool.shardID)
	crossShardPool.mtx.Lock()
	defer crossShardPool.mtx.Unlock()
	Logger.log.Infof("FindBeaconHeightForCrossShardBlock %d locked", crossShardPool.shardID)
	defer func() {
		Logger.log.Infof("FindBeaconHeightForCrossShardBlock %d unlocked", crossShardPool.shardID)
	}()
	return crossShardPool.findBeaconHeightForCrossShardBlock(beaconHeight, fromShardID, crossShardBlockHeight)
}
