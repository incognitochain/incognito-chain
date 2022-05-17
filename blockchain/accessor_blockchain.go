package blockchain

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/incognitochain/incognito-chain/blockchain/types"
	"github.com/incognitochain/incognito-chain/config"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/incdb"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/instruction"
	"github.com/incognitochain/incognito-chain/multiview"
	"github.com/incognitochain/incognito-chain/proto"
	"github.com/pkg/errors"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/rawdbv2"
)

//get beacon block hash by height, with current view
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
		return rawdbv2.GetFinalizedBeaconBlockHashByIndex(blockchain.GetBeaconChainDatabase(), height)
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
	if blkhash == nil {
		blkhash, err = rawdbv2.GetFinalizedBeaconBlockHashByIndex(blockchain.GetBeaconChainDatabase(), height)
		if err != nil {
			return nil, err
		}
	}

	beaconBlock, _, err := blockchain.GetBeaconBlockByHash(*blkhash)
	if err != nil {
		return nil, err
	}
	beaconBlocks = append(beaconBlocks, beaconBlock)

	return beaconBlocks, nil
}

func (blockchain *BlockChain) GetBeaconBlockByView(view multiview.View, height uint64) (*types.BeaconBlock, error) {
	blkhash, err := blockchain.GetBeaconBlockHashByHeight(blockchain.BeaconChain.GetFinalView(), view, height)
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
	blkCached, beaconHeaderBytes, beaconBodyBytes, err := blockchain.BeaconChain.blkManager.GetBlockByHash(&beaconBlockHash, proto.BlkType_BlkBc)
	if err != nil {
		return nil, 0, err
	}
	beaconBlock := types.NewBeaconBlock()
	if blkCached != nil {
		var ok bool
		if beaconBlock, ok = blkCached.(*types.BeaconBlock); ok {
			return beaconBlock, 0, nil
		} else {
			return beaconBlock, 0, errors.Errorf("Cached wrong data for beacon block %v", beaconBlockHash.String())
		}
	}
	if config.Config().EnableFFStorage {
		err = beaconBlock.SetHeaderFromBytes(beaconHeaderBytes)
		if err != nil {
			return nil, 0, err
		}
		if len(beaconBodyBytes) > 0 {
			err = beaconBlock.SetBodyFromBytes(beaconBodyBytes)
			if err != nil {
				return nil, 0, err
			}
		}
		if beaconBlock.GetHeight() > 1 {
			valData, err := blockchain.BeaconChain.blkManager.GetBlockValidation(beaconBlockHash)
			if err != nil {
				return nil, 0, err
			}
			beaconBlock.AddValidationField(valData)
		}
	} else {
		err = json.Unmarshal(beaconHeaderBytes, beaconBlock)
		if err != nil {
			return nil, 0, err
		}
	}
	return beaconBlock, uint64(len(beaconHeaderBytes)), nil
}

//SHARD
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
		return rawdbv2.GetFinalizedShardBlockHashByIndex(blockchain.GetShardChainDatabase(bestView.(*ShardBestState).ShardID), bestView.(*ShardBestState).ShardID, height)
	}

	// => if > finalblock, we use best view to trace back the correct height
	blkhash = *bestView.GetPreviousHash()
	blkheight = blkheight - 1
	for height < blkheight {
		shardBlock, _, err := blockchain.GetShardBlockByHash(blkhash)
		if err != nil {
			return nil, err
		}
		blkheight--
		blkhash = shardBlock.GetPrevHash()
	}

	return &blkhash, nil
}

func (blockchain *BlockChain) GetShardBlockByHeight(height uint64, shardID byte) (map[common.Hash]*types.ShardBlock, error) {
	shardBlockMap := make(map[common.Hash]*types.ShardBlock)
	sChain := blockchain.ShardChain[shardID]
	blkhash, err := blockchain.GetShardBlockHashByHeight(sChain.GetFinalView(), sChain.GetBestView(), height)
	if err != nil {
		return nil, err
	}

	shardBlock, _, err := blockchain.GetShardBlockByHashWithShardID(*blkhash, shardID)
	if err != nil {
		return nil, err
	}
	shardBlockMap[*shardBlock.Hash()] = shardBlock
	return shardBlockMap, err
}

func (blockchain *BlockChain) GetShardBlockByHeightV1(height uint64, shardID byte) (*types.ShardBlock, error) {
	shardBlocks, err := blockchain.GetShardBlockByHeight(height, shardID)
	if err != nil {
		return nil, err
	}
	if len(shardBlocks) == 0 {
		return nil, fmt.Errorf("NOT FOUND Shard Block By ShardID %+v Height %+v", shardID, height)
	}
	nShardBlocks, err := blockchain.GetShardBlockByHeight(height+1, shardID)
	if err == nil {
		for _, blk := range nShardBlocks {
			if sBlk, ok := shardBlocks[blk.GetPrevHash()]; ok {
				return sBlk, nil
			}
		}
	}
	for _, v := range shardBlocks {
		return v, nil
	}
	return nil, fmt.Errorf("NOT FOUND Shard Block By ShardID %+v Height %+v", shardID, height)
}

func (blockchain *BlockChain) GetShardBlockByHashWithShardID(hash common.Hash, shardID byte) (*types.ShardBlock, uint64, error) {
	blkCached, shardBlockBytes, shardBodyBytes, err := blockchain.ShardChain[shardID].blkManager.GetBlockByHash(&hash, proto.BlkType_BlkShard)
	if err != nil {
		return nil, 0, NewBlockChainError(GetShardBlockByHashError, errors.Errorf("Can not get block %v, error %v ", hash.String(), err))
	}
	shardBlock := types.NewShardBlock()
	if blkCached != nil {
		var ok bool
		if shardBlock, ok = blkCached.(*types.ShardBlock); ok {
			return shardBlock, 0, nil
		} else {
			return shardBlock, 0, errors.Errorf("Cached wrong data for shard block %v cid %v", hash.String(), shardID)
		}
	}
	if config.Config().EnableFFStorage {
		err = shardBlock.SetHeaderFromBytes(shardBlockBytes)
		if err != nil {
			return nil, 0, err
		}
		err = shardBlock.SetBodyFromBytes(shardBodyBytes)
		if err != nil {
			return nil, 0, err
		}
		if shardBlock.GetHeight() > 1 {
			valData, err := blockchain.ShardChain[shardID].blkManager.GetBlockValidation(hash)
			if err != nil {
				return nil, 0, err
			}
			shardBlock.AddValidationField(valData)
		}
	} else {
		err = json.Unmarshal(shardBlockBytes, shardBlock)
		if err != nil {
			return nil, 0, err
		}
	}
	return shardBlock, shardBlock.Header.Height, nil
}

func (blockchain *BlockChain) HasShardBlockByHash(hash common.Hash) (bool, error) {
	for _, i := range blockchain.GetShardIDs() {
		shardID := byte(i)
		has, err := rawdbv2.HasShardBlock(blockchain.GetShardChainDatabase(shardID), hash)
		if err != nil {
			return false, NewBlockChainError(GetShardBlockByHashError, err)
		}
		if has {
			return true, nil
		}
	}
	return false, NewBlockChainError(GetShardBlockByHashError, fmt.Errorf("Not found shard block by hash %+v", hash))
}

func (blockchain *BlockChain) GetShardBlockByHash(hash common.Hash) (*types.ShardBlock, uint64, error) {
	for _, i := range blockchain.GetShardIDs() {
		shardID := byte(i)
		shardBlk, height, err := blockchain.GetShardBlockByHashWithShardID(hash, shardID)
		if err == nil {
			return shardBlk, height, nil
		}
	}
	return nil, 0, NewBlockChainError(GetShardBlockByHashError, fmt.Errorf("Not found shard block by hash %+v", hash))
}

func (blockchain *BlockChain) GetShardBlockForBeaconProducer(bestShardHeights map[byte]uint64) map[byte][]*types.ShardBlock {
	allShardBlocks := make(map[byte][]*types.ShardBlock)
	for shardID, bestShardHeight := range bestShardHeights {
		finalizedShardHeight := blockchain.ShardChain[shardID].multiView.GetFinalView().GetHeight()
		shardBlocks := []*types.ShardBlock{}
		// limit maximum number of shard blocks for beacon producer
		if finalizedShardHeight > bestShardHeight && finalizedShardHeight-bestShardHeight > MAX_S2B_BLOCK {
			finalizedShardHeight = bestShardHeight + MAX_S2B_BLOCK
		}
		lastEpoch := uint64(0)
		limitTxs := map[int]int{}
		for shardHeight := bestShardHeight + 1; shardHeight <= finalizedShardHeight; shardHeight++ {
			tempShardBlock, err := blockchain.GetShardBlockByHeightV1(shardHeight, shardID)
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
	for shardID, requiredShardBlockHeight := range allRequiredShardBlockHeight {
		limitTxs := map[int]int{}
		shardBlocks := []*types.ShardBlock{}
		lastEpoch := uint64(0)
		for _, height := range requiredShardBlockHeight {
			shardBlock, err := blockchain.GetShardBlockByHeightV1(height, shardID)
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

func (blockchain *BlockChain) GetShardRootsHashFromBlockHeight(shardID byte, height uint64) (*ShardRootHashv2, error) {
	h, err := blockchain.GetShardBlockHashByHeight(blockchain.ShardChain[shardID].GetFinalView(), blockchain.ShardChain[shardID].GetBestView(), height)
	if err != nil {
		return nil, err
	}
	data, err := rawdbv2.GetShardRootsHash(blockchain.GetShardChainDatabase(shardID), shardID, *h)
	if err != nil {
		return nil, err
	}
	sRH := &ShardRootHashv2{}
	err = json.Unmarshal(data, sRH)
	return sRH, err
}

func GetShardRootsHashByBlockHash(db incdb.Database, shardID byte, hash common.Hash) (*ShardRootHashv2, error) {
	data, e := rawdbv2.GetShardRootsHash(db, shardID, hash)
	if e != nil {
		return nil, e
	}
	bRH := &ShardRootHashv2{}
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

	if (config.Config().SyncMode == common.STATEDB_LITE_MODE) && (height != 1) {
		bcBlks, err := s.GetConfig().Syncker.ReSyncBeaconBlockByHeight(height, height)
		if err != nil {
			return nil, err
		}
		return &(bcBlks[0]), nil
	} else {
		blkhash, err := rawdbv2.GetFinalizedBeaconBlockHashByIndex(s.GetBeaconChainDatabase(), height)
		if err != nil {
			return nil, err
		}
		beaconBlock, _, err := s.GetBeaconBlockByHash(*blkhash)
		if err != nil {
			return nil, err
		}
		return beaconBlock, nil
	}
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
	sRH := &ShardRootHashv2{}
	err1 := json.Unmarshal(data, sRH)
	if err1 != nil {
		return nil, err1
	}
	stateDB, err := statedb.NewWithPrefixTrie(sRH.ConsensusStateDBRootHash.GetRootHash(), statedb.NewDatabaseAccessWarper(db))
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

//GetStakerInfo : Return staker info from statedb
func (beaconBestState *BeaconBestState) GetStakerInfo(stakerPubkey string) (*statedb.StakerInfo, bool, error) {
	return statedb.GetStakerInfo(beaconBestState.consensusStateDB.Copy(), stakerPubkey)
}

type CommitteeCheckPoint struct {
	Height   uint64      `json:"h"`
	RootHash common.Hash `json:"rh"`
	FromBC   bool
}

type CommitteeChangeCheckpoint struct {
	Data   map[uint64]common.Hash
	Epochs []uint64
}

func (bc *BlockChain) initCommitChangeCheckpoint() error {
	bc.committeeChangeCheckpoint = struct {
		data   map[byte]CommitteeChangeCheckpoint
		locker *sync.RWMutex
	}{
		data:   map[byte]CommitteeChangeCheckpoint{},
		locker: &sync.RWMutex{},
	}
	for sID := byte(0); sID < byte(config.Param().ActiveShards); sID++ {
		bc.committeeChangeCheckpoint.data[sID] = CommitteeChangeCheckpoint{
			Data:   map[uint64]common.Hash{},
			Epochs: []uint64{},
		}
	}
	bc.committeeChangeCheckpoint.data[byte(common.BeaconChainSyncID)] = CommitteeChangeCheckpoint{
		Data:   map[uint64]common.Hash{},
		Epochs: []uint64{},
	}
	return nil
}

func (bc *BlockChain) updateCommitteeChangeCheckpointByBC(sID byte, epoch uint64, rootHash common.Hash) {
	bc.committeeChangeCheckpoint.locker.Lock()
	defer bc.committeeChangeCheckpoint.locker.Unlock()
	Logger.log.Debugf("[debugcachecommittee] beacon updateCommitteeChangeCheckpoint for shard %v, epoch %v", sID, epoch)
	sCommitteeChange := bc.committeeChangeCheckpoint.data[sID]
	if _, ok := sCommitteeChange.Data[epoch]; !ok {
		sCommitteeChange.Epochs = append(sCommitteeChange.Epochs, epoch)
	}
	sCommitteeChange.Data[epoch] = rootHash
	bc.committeeChangeCheckpoint.data[sID] = sCommitteeChange
	err := bc.backupCheckpoint()
	if err != nil {
		Logger.log.Error(err)
	}
}

func (bc *BlockChain) backupCheckpoint() error {
	data, err := json.Marshal(bc.committeeChangeCheckpoint.data)
	if err != nil {
		return err
	}
	db := bc.GetBeaconChainDatabase()
	err = rawdbv2.StoreCommitteeChangeCheckpoint(db, data)
	if err != nil {
		return err
	}
	return nil
}

func (bc *BlockChain) restoreCheckpoint() error {
	db := bc.GetBeaconChainDatabase()
	data, err := rawdbv2.GetCommitteeChangeCheckpoint(db)
	if err != nil {
		return err
	}
	committeeCheckpoint := map[byte]CommitteeChangeCheckpoint{}
	err = json.Unmarshal(data, &committeeCheckpoint)
	if err != nil {
		return err
	}
	bc.committeeChangeCheckpoint = struct {
		data   map[byte]CommitteeChangeCheckpoint
		locker *sync.RWMutex
	}{
		data:   committeeCheckpoint,
		locker: &sync.RWMutex{},
	}
	return nil
}

func (bc *BlockChain) initCheckpoint(initBeaconBestState *BeaconBestState) error {
	epochs := []uint64{}
	epochForCache := initBeaconBestState.BestBlock.Header.Epoch
	epochs = append(epochs, epochForCache)
	bc.committeeChangeCheckpoint.data[byte(common.BeaconChainSyncID)] = CommitteeChangeCheckpoint{
		Data: map[uint64]common.Hash{
			epochForCache: initBeaconBestState.ConsensusStateDBRootHash,
		},
		Epochs: epochs,
	}
	key := getCommitteeCacheKeyByEpoch(epochForCache, common.BeaconChainSyncID)
	bc.committeeByEpochCache.Add(key, initBeaconBestState.GetBeaconCommittee())
	for sID := 0; sID < initBeaconBestState.ActiveShards; sID++ {
		epochsForShard := append([]uint64{}, epochForCache)
		bc.committeeChangeCheckpoint.data[byte(sID)] = CommitteeChangeCheckpoint{
			Data: map[uint64]common.Hash{
				epochForCache: initBeaconBestState.ConsensusStateDBRootHash,
			},
			Epochs: epochsForShard,
		}
		key := getCommitteeCacheKeyByEpoch(epochForCache, byte(sID))
		bc.committeeByEpochCache.Add(key, initBeaconBestState.GetShardCommittee()[byte(sID)])
	}
	return nil
}

func (bc *BlockChain) GetCheckpointChangeCommitteeByEpoch(sID byte, epoch uint64) (
	epochForCache uint64,
	chkPnt common.Hash,
	err error,
) {
	bc.committeeChangeCheckpoint.locker.RLock()
	defer bc.committeeChangeCheckpoint.locker.RUnlock()
	sCommitteeChange := bc.committeeChangeCheckpoint.data[sID]
	epochs := sCommitteeChange.Epochs
	if len(epochs) == 0 {
		return 0, common.Hash{}, errors.Errorf("[CmtChkPnt] Committee change for epoch %v cID %v not found, list checkpoint is empty", epoch, sID)
	}
	idx, existed := SearchUint64(epochs, epoch)
	if existed {
		return epochs[idx], sCommitteeChange.Data[epoch], nil
	}
	if idx > len(epochs) {
		return epochs[len(epochs)-1], sCommitteeChange.Data[epochs[len(epochs)-1]], nil
	}
	if idx > 0 {
		return epochs[idx-1], sCommitteeChange.Data[epochs[idx-1]], nil
	}
	return 0, common.Hash{}, errors.Errorf("[CmtChkPnt] Committee change for epoch %v cID %v not found", epoch, sID)
}
