package blockchain

import (
	"encoding/json"
	"fmt"

	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/incdb"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/rawdbv2"
	btcrelaying "github.com/incognitochain/incognito-chain/relaying/btc"
)

func (blockchain *BlockChain) GetBeaconBlockByHeightAndView(height uint64, viewHash common.Hash) (*BeaconBlock, error) {
	finalView := blockchain.BeaconChain.GetFinalView()

	if height == finalView.GetHeight() {
		return finalView.GetBlock().(*BeaconBlock), nil
	}
	if height < finalView.GetHeight() {
		beaconBlocks, err := blockchain.GetBeaconBlockByHeight(height)
		if err != nil {
			return nil, err
		}
		return beaconBlocks[0], nil
	}
	if height > finalView.GetHeight() {
		view := blockchain.BeaconChain.GetViewByHash(viewHash)
		if view != nil {
			return view.GetBlock().(*BeaconBlock), nil
		}
	}
	return nil, fmt.Errorf("Beacon, Block Height %+v, View %+v, not found", height, viewHash)
}

func (blockchain *BlockChain) GetBeaconBlockByHeight(height uint64) ([]*BeaconBlock, error) {
	if blockchain.IsTest {
		return []*BeaconBlock{}, nil
	}
	beaconBlocks := []*BeaconBlock{}
	beaconBlockHash, err := statedb.GetBeaconBlockHashByIndex(blockchain.GetBeaconBestState().GetBeaconConsensusStateDB(), height)
	if err != nil {
		return nil, err
	}

	beaconBlock, _, err := blockchain.GetBeaconBlockByHash(beaconBlockHash)
	if err != nil {
		return nil, err
	}
	beaconBlocks = append(beaconBlocks, beaconBlock)

	return beaconBlocks, nil
}

func (blockchain *BlockChain) GetFinalizedBeaconBlockByHeight(height uint64) (*BeaconBlock, error) {
	beaconBlocks, err := blockchain.GetBeaconBlockByHeight(height)
	if err != nil {
		return nil, err
	}
	if len(beaconBlocks) == 0 {
		return nil, fmt.Errorf("Beacon Block Height %+v NOT FOUND", height)
	}
	return beaconBlocks[0], nil
}

func (blockchain *BlockChain) GetBeaconBlockByHash(beaconBlockHash common.Hash) (*BeaconBlock, uint64, error) {
	if blockchain.IsTest {
		return &BeaconBlock{}, 2, nil
	}
	beaconBlockBytes, err := rawdbv2.GetBeaconBlockByHash(blockchain.GetBeaconChainDatabase(), beaconBlockHash)
	if err != nil {
		return nil, 0, err
	}
	beaconBlock := NewBeaconBlock()
	err = json.Unmarshal(beaconBlockBytes, beaconBlock)
	if err != nil {
		return nil, 0, err
	}
	return beaconBlock, uint64(len(beaconBlockBytes)), nil
}

func (blockchain *BlockChain) GetShardBlockHashByHeight(height uint64, shardID byte) (common.Hash, error) {
	hash, err := statedb.GetShardBlockHashByIndex(blockchain.GetBestStateShard(shardID).consensusStateDB.Copy(), shardID, height)
	if err != nil {
		return hash, err
	}

	return hash, nil
}

func (blockchain *BlockChain) GetShardBlockByHeight(height uint64, shardID byte) (map[common.Hash]*ShardBlock, error) {
	shardBlockMap := make(map[common.Hash]*ShardBlock)
	hash, err := statedb.GetShardBlockHashByIndex(blockchain.GetBestStateShard(shardID).consensusStateDB.Copy(), shardID, height)
	if err != nil {
		return nil, err
	}
	data, err := rawdbv2.GetShardBlockByHash(blockchain.GetShardChainDatabase(shardID), hash)
	if err != nil {
		return nil, err
	}
	shardBlock := NewShardBlock()
	err = json.Unmarshal(data, shardBlock)
	if err != nil {
		return nil, err
	}
	shardBlockMap[*shardBlock.Hash()] = shardBlock
	return shardBlockMap, err
}

func (blockchain *BlockChain) GetFinalizedShardBlockByHeight(height uint64, shardID byte) (*ShardBlock, error) {
	res, err := blockchain.GetShardBlockByHeight(height, shardID)
	if err != nil {
		return nil, err
	}
	for _, v := range res {
		return v, nil
	}
	return nil, fmt.Errorf("NOT FOUND Shard Block By ShardID %+v Height %+v", shardID, height)
}

func (blockchain *BlockChain) GetShardBlockByHashWithShardID(hash common.Hash, shardID byte) (*ShardBlock, uint64, error) {
	shardBlockBytes, err := rawdbv2.GetShardBlockByHash(blockchain.GetShardChainDatabase(shardID), hash)
	if err != nil {
		return nil, 0, err
	}
	shardBlock := NewShardBlock()
	err = json.Unmarshal(shardBlockBytes, shardBlock)
	if err != nil {
		return nil, 0, NewBlockChainError(GetShardBlockByHashError, err)
	}
	return shardBlock, shardBlock.Header.Height, nil
}

func (blockchain *BlockChain) GetShardBlockByHash(hash common.Hash) (*ShardBlock, uint64, error) {
	if blockchain.IsTest {
		return &ShardBlock{}, 2, nil
	}
	for _, i := range blockchain.GetShardIDs() {
		shardID := byte(i)
		shardBlockBytes, err := rawdbv2.GetShardBlockByHash(blockchain.GetShardChainDatabase(shardID), hash)
		if err == nil {
			shardBlock := NewShardBlock()
			err = json.Unmarshal(shardBlockBytes, shardBlock)
			if err != nil {
				return nil, 0, NewBlockChainError(GetShardBlockByHashError, err)
			}
			return shardBlock, shardBlock.Header.Height, nil
		}
	}
	return nil, 0, NewBlockChainError(GetShardBlockByHashError, fmt.Errorf("Not found shard block by hash %+v", hash))
}

func (blockchain *BlockChain) GetShardBlockForBeaconProducer(bestShardHeights map[byte]uint64) map[byte][]*ShardBlock {
	allShardBlocks := make(map[byte][]*ShardBlock)
	for shardID, bestShardHeight := range bestShardHeights {
		finalizedShardHeight := blockchain.GetBestStateShard(shardID).ShardHeight
		//_, finalizedShardHeight, err := blockchain.GetLatestFinalizedShardBlock(shardID)
		//if err != nil {
		//	Logger.log.Errorf("GetLatestFinalizedShardBlock shard %+v, error  %+v", shardID, err)
		//	continue
		//}
		shardBlocks := []*ShardBlock{}
		// limit maximum number of shard blocks for beacon producer
		if finalizedShardHeight-bestShardHeight > MAX_S2B_BLOCK {
			finalizedShardHeight = bestShardHeight + MAX_S2B_BLOCK
		}
		for shardHeight := bestShardHeight + 1; shardHeight <= finalizedShardHeight; shardHeight++ {
			tempShardBlock, err := blockchain.GetFinalizedShardBlockByHeight(shardHeight, shardID)
			if err != nil {
				Logger.log.Errorf("GetFinalizedShardBlockByHeight shard %+v, error  %+v", shardID, err)
				break
			}
			shardBlocks = append(shardBlocks, tempShardBlock)
		}
		allShardBlocks[shardID] = shardBlocks
	}
	return allShardBlocks
}

func (blockchain *BlockChain) GetShardBlocksForBeaconValidator(allRequiredShardBlockHeight map[byte][]uint64) (map[byte][]*ShardBlock, error) {
	allRequireShardBlocks := make(map[byte][]*ShardBlock)
	for shardID, requiredShardBlockHeight := range allRequiredShardBlockHeight {
		shardBlocks := []*ShardBlock{}
		for _, height := range requiredShardBlockHeight {
			shardBlock, err := blockchain.GetFinalizedShardBlockByHeight(height, shardID)
			if err != nil {
				return allRequireShardBlocks, err
			}
			shardBlocks = append(shardBlocks, shardBlock)
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

func (blockchain *BlockChain) GetBNBChainID() string {
	return blockchain.GetConfig().ChainParams.BNBRelayingHeaderChainID
}

func (blockchain *BlockChain) GetBTCChainID() string {
	return blockchain.GetConfig().ChainParams.BTCRelayingHeaderChainID
}

func (blockchain *BlockChain) GetBTCHeaderChain() *btcrelaying.BlockChain {
	return blockchain.GetConfig().BTCChain
}

func (blockchain *BlockChain) GetPortalFeederAddress() string {
	return blockchain.GetConfig().ChainParams.PortalFeederAddress
}
