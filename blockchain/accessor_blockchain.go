package blockchain

import (
	"encoding/json"
	"fmt"

	"github.com/incognitochain/incognito-chain/blockchain/types"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/incdb"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/instruction"
	"github.com/incognitochain/incognito-chain/multiview"
	"github.com/pkg/errors"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/rawdbv2"
)

// get beacon block hash by height, with current view
func (blockchain *BlockChain) GetBeaconBlockHashByHeight(finalView, bestView multiview.View, height uint64) (*common.Hash, error) {

	blkheight := bestView.GetHeight()
	blkhash := *bestView.GetHash()

	if height > blkheight {
		return nil, fmt.Errorf("Beacon, Block Height %+v not found", height)
	}

	if height == blkheight {
		return &blkhash, nil
	}

	// => check if <= final block, using rawdb
	if height <= finalView.GetHeight() { //note there is chance that == final view, but block is not stored (store in progress)
		return blockchain.BeaconChain.BlockStorage.GetFinalizedBeaconBlock(height)
	}

	// => if > finalblock, we use best view to trace back the correct height
	blkhash = *bestView.GetPreviousHash()
	blkheight = blkheight - 1
	for height < blkheight {
		beaconBlock, _, err := blockchain.GetBeaconBlockByHash(blkhash)
		if err != nil {
			return nil, err
		}
		blkheight--
		blkhash = beaconBlock.GetPrevHash()
	}

	return &blkhash, nil
}

func (blockchain *BlockChain) GetBeaconBlockHashByView(view multiview.View, height uint64) (*common.Hash, error) {

	blkheight := view.GetHeight()
	blkhash := *view.GetHash()

	if height > blkheight {
		return nil, fmt.Errorf("Beacon, Block Height %+v not found", height)
	}

	if height == blkheight {
		return &blkhash, nil
	}

	// => check if <= final block, using rawdb
	return blockchain.BeaconChain.BlockStorage.GetFinalizedBeaconBlock(height)
}

func (blockchain *BlockChain) GetBeaconBlockByHeightV1(height uint64) (*types.BeaconBlock, error) {
	beaconBlocks, err := blockchain.GetBeaconBlockByHeight(height)
	if err != nil {
		return nil, err
	}
	if len(beaconBlocks) == 0 {
		return nil, fmt.Errorf("Beacon Block Height %+v NOT FOUND", height)
	}
	return beaconBlocks[0], nil
}

func (blockchain *BlockChain) GetBeaconBlockByHeight(height uint64) ([]*types.BeaconBlock, error) {

	if blockchain.IsTest {
		return []*types.BeaconBlock{}, nil
	}
	beaconBlocks := []*types.BeaconBlock{}

	blkhash, err := blockchain.GetBeaconBlockHashByHeight(blockchain.BeaconChain.GetFinalView(), blockchain.BeaconChain.GetBestView(), height)
	if err != nil {
		return nil, err
	}

	beaconBlock, _, err := blockchain.GetBeaconBlockByHash(*blkhash)
	if err != nil {
		return nil, err
	}
	beaconBlocks = append(beaconBlocks, beaconBlock)

	return beaconBlocks, nil
}

func (blockchain *BlockChain) GetBeaconBlockByView(view multiview.View, height uint64) (*types.BeaconBlock, error) {
	blkhash, err := blockchain.GetBeaconBlockHashByView(view, height)
	if err != nil {
		return nil, err
	}

	beaconBlock, _, err := blockchain.GetBeaconBlockByHash(*blkhash)
	if err != nil {
		return nil, err
	}

	return beaconBlock, nil
}

func (blockchain *BlockChain) GetBeaconBlockByHash(beaconBlockHash common.Hash) (*types.BeaconBlock, uint64, error) {
	if blockchain.IsTest {
		return &types.BeaconBlock{}, 2, nil
	}
	blk, size, err := blockchain.BeaconChain.BlockStorage.GetBlock(beaconBlockHash)
	if err != nil {
		return nil, 0, err
	}
	return blk.(*types.BeaconBlock), uint64(size), nil
}
func (blockchain *BlockChain) GetBeaconBlockByHashWithLatestValidationData(beaconBlockHash common.Hash) (*types.BeaconBlock, uint64, error) {
	if blockchain.IsTest {
		return &types.BeaconBlock{}, 2, nil
	}
	blk, size, err := blockchain.BeaconChain.BlockStorage.GetBlockWithLatestValidationData(beaconBlockHash)
	if err != nil {
		return nil, 0, err
	}
	return blk.(*types.BeaconBlock), uint64(size), nil
}

// SHARD
func (blockchain *BlockChain) GetShardBlockHashByHeight(finalView, bestView multiview.View, height uint64) (*common.Hash, error) {

	blkheight := bestView.GetHeight()
	blkhash := *bestView.GetHash()

	if height > blkheight {
		return nil, fmt.Errorf("Shard, Block Height %+v not found", height)
	}

	if height == blkheight {
		return &blkhash, nil
	}

	// => check if <= final block, using rawdb
	if height <= finalView.GetHeight() { //note there is chance that == final view, but block is not stored (store in progress)
		return blockchain.ShardChain[bestView.(*ShardBestState).ShardID].BlockStorage.GetFinalizedShardBlockHashByIndex(height)
	}

	// => if > finalblock, we use best view to trace back the correct height
	blkhash = *bestView.GetPreviousHash()
	blkheight = blkheight - 1
	for height < blkheight {
		shardBlock, err := blockchain.ShardChain[bestView.(*ShardBestState).ShardID].GetBlockByHash(blkhash)
		if err != nil {
			return nil, err
		}
		blkheight--
		blkhash = shardBlock.GetPrevHash()
	}

	return &blkhash, nil
}

func (blockchain *BlockChain) GetShardBlockHashByView(view multiview.View, height uint64) (*common.Hash, error) {

	blkheight := view.GetHeight()
	blkhash := *view.GetHash()

	if height > blkheight {
		return nil, fmt.Errorf("Shard, Block Height %+v not found", height)
	}

	if height == blkheight {
		return &blkhash, nil
	}

	return blockchain.ShardChain[view.(*ShardBestState).ShardID].BlockStorage.GetFinalizedShardBlockHashByIndex(height)
}

func (blockchain *BlockChain) GetShardBlockByHeight(height uint64, shardID byte) (map[common.Hash]*types.ShardBlock, error) {
	shardBlockMap := make(map[common.Hash]*types.ShardBlock)
	blkhash, err := blockchain.
		GetShardBlockHashByHeight(blockchain.ShardChain[shardID].
			GetFinalView(), blockchain.ShardChain[shardID].GetBestView(), height)
	if err != nil {
		return nil, err
	}
	blk, _, err := blockchain.ShardChain[shardID].BlockStorage.GetBlock(*blkhash)
	if err != nil {
		return nil, err
	}
	shardBlock := blk.(*types.ShardBlock)
	shardBlockMap[*shardBlock.Hash()] = shardBlock
	return shardBlockMap, err
}

func (blockchain *BlockChain) GetShardBlockByView(view multiview.View, height uint64, shardID byte) (*types.ShardBlock, error) {
	blkhash, err := blockchain.GetShardBlockHashByView(view, height)
	if err != nil {
		return nil, err
	}
	blk, _, err := blockchain.ShardChain[shardID].BlockStorage.GetBlock(*blkhash)
	if err != nil { //no transaction in this node
		return nil, err
	}
	return blk.(*types.ShardBlock), err
}

func (blockchain *BlockChain) GetShardBlockWithLatestValidation(height uint64, shardID byte) (*types.ShardBlock, error) {
	blkhash, err := blockchain.
		GetShardBlockHashByHeight(blockchain.ShardChain[shardID].
			GetFinalView(), blockchain.ShardChain[shardID].GetBestView(), height)
	if err != nil {
		return nil, fmt.Errorf("NOT FOUND Shard Block By ShardID %+v Height %+v", shardID, height)
	}
	blk, _, err := blockchain.ShardChain[shardID].BlockStorage.GetBlockWithLatestValidationData(*blkhash)
	if err != nil {
		return nil, fmt.Errorf("NOT FOUND Shard Block By ShardID %+v Height %+v", shardID, height)
	}
	shardBlock := blk.(*types.ShardBlock)
	return shardBlock, nil
}

func (blockchain *BlockChain) GetShardBlockByHeightV1(height uint64, shardID byte) (*types.ShardBlock, error) {
	blkhash, err := blockchain.
		GetShardBlockHashByHeight(blockchain.ShardChain[shardID].
			GetFinalView(), blockchain.ShardChain[shardID].GetBestView(), height)
	if err != nil {
		return nil, fmt.Errorf("NOT FOUND Shard Block By ShardID %+v Height %+v", shardID, height)
	}
	blk, _, err := blockchain.ShardChain[shardID].BlockStorage.GetBlock(*blkhash)
	if err != nil {
		return nil, fmt.Errorf("NOT FOUND Shard Block By ShardID %+v Height %+v", shardID, height)
	}
	shardBlock := blk.(*types.ShardBlock)
	return shardBlock, nil
}

func (blockchain *BlockChain) GetShardBlockByHashWithShardID(hash common.Hash, shardID byte) (*types.ShardBlock, uint64, error) {
	blk, size, err := blockchain.ShardChain[shardID].BlockStorage.GetBlock(hash)
	if err != nil {
		return nil, 0, err
	}
	return blk.(*types.ShardBlock), uint64(size), nil
}

func (blockchain *BlockChain) GetShardBlockByHash(hash common.Hash) (*types.ShardBlock, uint64, error) {
	for _, i := range blockchain.GetShardIDs() {
		shardID := byte(i)
		blk, size, _ := blockchain.ShardChain[shardID].BlockStorage.GetBlock(hash)
		if size != 0 {
			return blk.(*types.ShardBlock), uint64(size), nil
		}
	}
	return nil, 0, NewBlockChainError(GetShardBlockByHashError, fmt.Errorf("Not found shard block by hash %+v", hash))
}

// traverse finalview back to certain block height from beacon chain
func (blockchain *BlockChain) GetShardBlockForBridge(from uint64, to common.Hash, newBeaconBlock *types.BeaconBlock, blockShardStates map[byte][]types.ShardState) (map[byte][]*types.ShardBlock, map[uint64]map[byte][]*types.ShardBlock, error) {
	beaconBlk, _, _ := blockchain.GetBeaconBlockByHash(to)
	shardBlksForBridge := map[byte][]*types.ShardBlock{}
	shardBlksForBridgeAgg := map[uint64]map[byte][]*types.ShardBlock{}
	for {
		if beaconBlk == nil {
			return nil, nil, NewBlockChainError(rawdbv2.GetBeaconBlockByHashError, fmt.Errorf("Cannot find beacon block %v", to))
		}
		beaconHeight := beaconBlk.GetHeight()
		if beaconHeight < from {
			break
		}
		Logger.log.Infof("[Bridge Debug] Checking bridge for beacon block %v", beaconHeight)

		shardBlksForBridgeAgg[beaconHeight] = map[byte][]*types.ShardBlock{}
		for sid, shardStates := range beaconBlk.Body.ShardState {
			shardBlocks := []*types.ShardBlock{}
			for _, sState := range shardStates {
				shardBlk, _ := blockchain.ShardChain[sid].GetBlockByHash(sState.Hash)
				if shardBlk == nil {
					return nil, nil, NewBlockChainError(rawdbv2.GetShardBlockByHashError, fmt.Errorf("Cannot find shard block %v", sState.Hash))
				}
				shardBlocks = append(shardBlocks, shardBlk.(*types.ShardBlock))
			}
			shardBlksForBridge[sid] = append(shardBlocks, shardBlksForBridge[sid]...)
			shardBlksForBridgeAgg[beaconHeight][sid] = shardBlocks
		}
		beaconBlk, _, _ = blockchain.GetBeaconBlockByHash(beaconBlk.GetPrevHash())
	}

	if newBeaconBlock.Header.Version >= types.INSTANT_FINALITY_VERSION_V2 {
		if shardBlksForBridgeAgg[newBeaconBlock.GetHeight()] == nil {
			shardBlksForBridgeAgg[newBeaconBlock.GetHeight()] = map[byte][]*types.ShardBlock{}
		}
		for sid, shardStates := range blockShardStates {
			shardBlocks := []*types.ShardBlock{}
			for _, sState := range shardStates {
				shardBlk, _ := blockchain.ShardChain[sid].GetBlockByHash(sState.Hash)
				if shardBlk == nil {
					return nil, nil, NewBlockChainError(rawdbv2.GetShardBlockByHashError, fmt.Errorf("Cannot find shard block %v", sState.Hash))
				}
				shardBlocks = append(shardBlocks, shardBlk.(*types.ShardBlock))
			}
			shardBlksForBridge[sid] = append(shardBlksForBridge[sid], shardBlocks...)
			shardBlksForBridgeAgg[newBeaconBlock.GetHeight()][sid] = shardBlocks
		}
	}
	return shardBlksForBridge, shardBlksForBridgeAgg, nil
}

func (blockchain *BlockChain) GetShardBlockForBeaconProducer(bestShardHeights map[byte]uint64) map[byte][]*types.ShardBlock {
	allShardBlocks := make(map[byte][]*types.ShardBlock)
	for shardID, bestShardHeight := range bestShardHeights {
		expectedFinalView := blockchain.ShardChain[shardID].multiView.GetExpectedFinalView()
		finalizedShardHeight := expectedFinalView.GetHeight()
		// limit maximum number of shard blocks for beacon producer
		if finalizedShardHeight > bestShardHeight && finalizedShardHeight-bestShardHeight > MAX_S2B_BLOCK {
			finalizedShardHeight = bestShardHeight + MAX_S2B_BLOCK
		}

		listShardBlock := map[uint64]*types.ShardBlock{}
		pointerView := expectedFinalView
		tempShardBlock := pointerView.GetBlock().(*types.ShardBlock)
		for {
			if tempShardBlock.GetHeight() <= bestShardHeight {
				break
			}
			listShardBlock[tempShardBlock.GetHeight()] = tempShardBlock

			pointerView = blockchain.ShardChain[shardID].multiView.GetViewByHash(tempShardBlock.GetPrevHash())
			if pointerView == nil {
				shardHeight := tempShardBlock.GetHeight() - 1
				var err error
				tempShardBlock, err = blockchain.GetShardBlockByHeightV1(shardHeight, shardID)
				if err != nil {
					Logger.log.Errorf("Cannot get block at height %v", shardHeight)
					return nil
				}
			} else {
				tempShardBlock = pointerView.GetBlock().(*types.ShardBlock)
			}
		}

		shardBlocks := []*types.ShardBlock{}
		lastEpoch := uint64(0)
		limitTxs := map[int]int{}
		for shardHeight := bestShardHeight + 1; shardHeight <= finalizedShardHeight; shardHeight++ {
			tempShardBlock, err := blockchain.GetShardBlockWithLatestValidation(shardHeight, shardID)
			if err != nil {
				Logger.log.Errorf("GetShardBlockByHeightV1 shard %+v, error  %+v", shardID, err)
				break
			}
			//only get shard block within epoch
			if lastEpoch == 0 {
				lastEpoch = tempShardBlock.GetCurrentEpoch() //update epoch of first block
			} else {
				if lastEpoch != tempShardBlock.GetCurrentEpoch() { //if next block have different epoch than break
					break
				}
			}
			if ok := checkLimitTxAction(true, limitTxs, tempShardBlock); !ok {
				Logger.log.Infof("Maximum tx action, return %v/%v block for shard %v", len(shardBlocks), finalizedShardHeight-bestShardHeight, shardID)
				break
			}
			shardBlocks = append(shardBlocks, tempShardBlock)

			containSwap := func(inst [][]string) bool {
				for _, inst := range inst {
					if inst[0] == instruction.SWAP_ACTION {
						return true
					}
				}
				return false
			}
			if containSwap(tempShardBlock.Body.Instructions) {
				break
			}
		}
		allShardBlocks[shardID] = shardBlocks
	}
	return allShardBlocks
}

func (blockchain *BlockChain) GetShardBlocksForBeaconValidator(allRequiredShardBlockHeight map[byte][]uint64) (map[byte][]*types.ShardBlock, error) {
	allRequireShardBlocks := make(map[byte][]*types.ShardBlock)
	bestShardHeight := blockchain.GetBeaconBestState().BestShardHeight
	for shardID, requiredShardBlockHeight := range allRequiredShardBlockHeight {
		expectedFinalView := blockchain.ShardChain[shardID].multiView.GetExpectedFinalView()
		listShardBlock := map[uint64]*types.ShardBlock{}
		pointerView := expectedFinalView
		tempShardBlock := pointerView.GetBlock().(*types.ShardBlock)

		for {
			if tempShardBlock.GetHeight() <= bestShardHeight[shardID] {
				break
			}
			listShardBlock[tempShardBlock.GetHeight()] = tempShardBlock

			pointerView = blockchain.ShardChain[shardID].multiView.GetViewByHash(tempShardBlock.GetPrevHash())
			if pointerView == nil {
				shardHeight := tempShardBlock.GetHeight() - 1
				var err error
				tempShardBlock, err = blockchain.GetShardBlockByHeightV1(shardHeight, shardID)
				if err != nil {
					return nil, fmt.Errorf("Cannot get block at height %v", shardHeight)
				}
			} else {
				tempShardBlock = pointerView.GetBlock().(*types.ShardBlock)
			}
		}

		limitTxs := map[int]int{}
		shardBlocks := []*types.ShardBlock{}
		lastEpoch := uint64(0)
		for _, height := range requiredShardBlockHeight {
			shardBlock, err := blockchain.GetShardBlockWithLatestValidation(height, shardID)
			if err != nil {
				return nil, err
			}
			if ok := checkLimitTxAction(true, limitTxs, shardBlock); !ok {
				return nil, errors.Errorf("Total txs of range ShardBlocks [%v..%v] is lager than limit", requiredShardBlockHeight[0], requiredShardBlockHeight[len(requiredShardBlockHeight)-1])
			}
			//only get shard block within epoch
			if lastEpoch == 0 {
				lastEpoch = shardBlock.GetCurrentEpoch() //update epoch of first block
			} else {
				if lastEpoch != shardBlock.GetCurrentEpoch() { //if next block have different epoch than break
					return nil, fmt.Errorf("Contain block in different epoch")
				}
			}

			shardBlocks = append(shardBlocks, shardBlock)

			containSwap := func(inst [][]string) bool {
				for _, inst := range inst {
					if inst[0] == instruction.SWAP_ACTION {
						return true
					}
				}
				return false
			}
			if containSwap(shardBlock.Body.Instructions) {
				break
			}

		}
		allRequireShardBlocks[shardID] = shardBlocks
	}
	return allRequireShardBlocks, nil
}

func (blockchain *BlockChain) GetBestStateShardRewardStateDB(shardID byte) *statedb.StateDB {
	return blockchain.GetBestStateShard(shardID).GetShardRewardStateDB()
}

func (blockchain *BlockChain) GetBestStateTransactionStateDB(shardID byte) *statedb.StateDB {
	return blockchain.GetBestStateShard(shardID).GetCopiedTransactionStateDB()
}

func (blockchain *BlockChain) GetShardFeatureStateDB(shardID byte) *statedb.StateDB {
	return blockchain.GetBestStateShard(shardID).GetCopiedFeatureStateDB()
}

func (blockchain *BlockChain) GetBestStateBeaconFeatureStateDB() *statedb.StateDB {
	return blockchain.GetBeaconBestState().GetBeaconFeatureStateDB()
}

func (blockchain *BlockChain) GetBestStateBeaconFeatureStateDBByHeight(height uint64, db incdb.Database) (*statedb.StateDB, error) {
	rootHash, err := blockchain.GetBeaconFeatureRootHash(blockchain.GetBeaconBestState(), height)
	if err != nil {
		return nil, fmt.Errorf("Beacon Feature State DB not found, height %+v, error %+v", height, err)
	}
	return statedb.NewWithPrefixTrie(rootHash, statedb.NewDatabaseAccessWarper(db))
}

func (blockchain *BlockChain) GetBeaconConsensusRootHash(beaconbestState *BeaconBestState, height uint64) (common.Hash, error) {
	bRH, e := blockchain.GetBeaconRootsHashFromBlockHeight(height)
	if e != nil {
		return common.Hash{}, e
	}
	return bRH.ConsensusStateDBRootHash, nil
}

func (blockchain *BlockChain) GetBeaconFeatureRootHash(beaconbestState *BeaconBestState, height uint64) (common.Hash, error) {
	bRH, e := blockchain.GetBeaconRootsHashFromBlockHeight(height)
	if e != nil {
		return common.Hash{}, e
	}
	return bRH.FeatureStateDBRootHash, nil
}

func (blockchain *BlockChain) GetBeaconRootsHashFromBlockHeight(height uint64) (*BeaconRootHash, error) {
	h, e := blockchain.GetBeaconBlockHashByHeight(blockchain.BeaconChain.GetFinalView(), blockchain.BeaconChain.GetBestView(), height)
	if e != nil {
		return nil, e
	}
	return GetBeaconRootsHashByBlockHash(blockchain.GetBeaconChainDatabase(), *h)
}

func GetBeaconRootsHashByBlockHash(db incdb.Database, hash common.Hash) (*BeaconRootHash, error) {
	data, e := rawdbv2.GetBeaconRootsHash(db, hash)
	if e != nil {
		return nil, e
	}
	bRH := &BeaconRootHash{}
	err := json.Unmarshal(data, bRH)
	return bRH, err
}

func (blockchain *BlockChain) GetShardRootsHashFromBlockHeight(shardID byte, height uint64) (*ShardRootHash, error) {
	h, err := blockchain.GetShardBlockHashByHeight(blockchain.ShardChain[shardID].GetFinalView(), blockchain.ShardChain[shardID].GetBestView(), height)
	if err != nil {
		return nil, err
	}
	data, err := rawdbv2.GetShardRootsHash(blockchain.GetShardChainDatabase(shardID), shardID, *h)
	if err != nil {
		return nil, err
	}
	sRH := &ShardRootHash{}
	err = json.Unmarshal(data, sRH)
	return sRH, err
}

func GetShardRootsHashByBlockHash(db incdb.Database, shardID byte, hash common.Hash) (*ShardRootHash, error) {
	data, e := rawdbv2.GetShardRootsHash(db, shardID, hash)
	if e != nil {
		return nil, e
	}
	bRH := &ShardRootHash{}
	err := json.Unmarshal(data, bRH)
	return bRH, err
}

func (s *BlockChain) FetchNextCrossShard(fromSID, toSID int, currentHeight uint64) *NextCrossShardInfo {
	b, err := rawdbv2.GetCrossShardNextHeight(s.GetBeaconChainDatabase(), byte(fromSID), byte(toSID), uint64(currentHeight))
	if err != nil {
		//Logger.log.Error(fmt.Sprintf("Cannot FetchCrossShardNextHeight fromSID %d toSID %d with currentHeight %d", fromSID, toSID, currentHeight))
		return nil
	}
	var res = new(NextCrossShardInfo)
	err = json.Unmarshal(b, res)
	if err != nil {
		return nil
	}
	return res
}

func (s *BlockChain) FetchConfirmBeaconBlockByHeight(height uint64) (*types.BeaconBlock, error) {
	blkhash, err := s.BeaconChain.BlockStorage.GetFinalizedBeaconBlock(height)
	if err != nil {
		return nil, err
	}
	beaconBlock, _, err := s.GetBeaconBlockByHash(*blkhash)
	if err != nil {
		return nil, err
	}
	return beaconBlock, nil
}

func getOneShardCommitteeFromShardDB(db incdb.Database, shardID byte, blockHash common.Hash) ([]incognitokey.CommitteePublicKey, error) {
	consensusStateDB, err := getShardConsensusStateDB(db, shardID, blockHash)
	if err != nil {
		return []incognitokey.CommitteePublicKey{}, err
	}
	committees := statedb.GetOneShardCommittee(consensusStateDB, shardID)
	return committees, nil
}

func getShardConsensusStateDB(db incdb.Database, shardID byte, blockHash common.Hash) (*statedb.StateDB, error) {
	data, err := rawdbv2.GetShardRootsHash(db, shardID, blockHash)
	if err != nil {
		return nil, err
	}
	sRH := &ShardRootHash{}
	err1 := json.Unmarshal(data, sRH)
	if err1 != nil {
		return nil, err1
	}
	stateDB, err := statedb.NewWithPrefixTrie(sRH.ConsensusStateDBRootHash, statedb.NewDatabaseAccessWarper(db))
	if err != nil {
		return nil, err
	}
	return stateDB, nil
}

func getOneShardCommitteeFromBeaconDB(db incdb.Database, shardID byte, beaconHashForCommittee common.Hash) ([]incognitokey.CommitteePublicKey, error) {
	consensusStateDB, err := getBeaconConsensusStateDB(db, beaconHashForCommittee)
	if err != nil {
		return []incognitokey.CommitteePublicKey{}, err
	}
	committees := statedb.GetOneShardCommittee(consensusStateDB, shardID)
	return committees, nil
}

func getBeaconConsensusStateDB(db incdb.Database, hash common.Hash) (*statedb.StateDB, error) {
	data, err := rawdbv2.GetBeaconRootsHash(db, hash)
	if err != nil {
		return nil, err
	}
	bRH := &BeaconRootHash{}
	err1 := json.Unmarshal(data, bRH)
	if err1 != nil {
		return nil, err1
	}
	stateDB, err := statedb.NewWithPrefixTrie(bRH.ConsensusStateDBRootHash, statedb.NewDatabaseAccessWarper(db))
	if err != nil {
		return nil, err
	}
	return stateDB, nil
}

func (blockchain *BlockChain) GetBeaconRootsHash(height uint64) (*BeaconRootHash, error) {
	h, e := blockchain.GetBeaconBlockHashByHeight(blockchain.BeaconChain.GetFinalView(), blockchain.BeaconChain.GetBestView(), height)
	if e != nil {
		return nil, e
	}
	data, e := rawdbv2.GetBeaconRootsHash(blockchain.GetBeaconChainDatabase(), *h)
	if e != nil {
		return nil, e
	}
	bRH := &BeaconRootHash{}
	err := json.Unmarshal(data, bRH)
	return bRH, err
}

func (blockchain *BlockChain) GetBeaconCommitteeOfEpoch(epoch uint64) ([]incognitokey.CommitteePublicKey, error) {
	lastHeightOfEpoch := blockchain.GetLastBeaconHeightInEpoch(epoch)
	blkHeightForGetCmt := lastHeightOfEpoch - 1
	blkHashForGetCmt, err := blockchain.GetBeaconBlockHashByHeight(blockchain.BeaconChain.GetFinalView(), blockchain.BeaconChain.GetBestView(), blkHeightForGetCmt)
	if err != nil {
		return nil, err
	}
	return blockchain.BeaconChain.GetCommitteeByHash(*blkHashForGetCmt, blkHeightForGetCmt)
}

// GetStakerInfo : Return staker info from statedb
func (beaconBestState *BeaconBestState) GetStakerInfo(stakerPubkey string) (*statedb.StakerInfo, bool, error) {
	return statedb.GetStakerInfo(beaconBestState.consensusStateDB.Copy(), stakerPubkey)
}

func (beaconBestState *BeaconBestState) GetBeaconStakerInfo(stakerPubkey string) (*statedb.BeaconStakerInfo, bool, error) {
	return statedb.GetBeaconStakerInfo(beaconBestState.consensusStateDB.Copy(), stakerPubkey)
}
