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

func (blockchain *BlockChain) StoreShardBestState(shardID byte) error {
	return rawdbv2.StoreShardBestState(blockchain.GetShardChainDatabase(shardID), shardID, blockchain.GetBestStateShard(shardID))
}

func (blockchain *BlockChain) FinalizedShardBlock(shardBlock *ShardBlock) error {
	return rawdbv2.FinalizedShardBlock(blockchain.GetShardChainDatabase(shardBlock.Header.ShardID), shardBlock.Header.ShardID, shardBlock.Header.Hash())
}

func (blockchain *BlockChain) GetFinalizedShardBlock(shardID byte) (*ShardBlock, uint64, error) {
	hash, err := rawdbv2.GetFinalizedShardBlock(blockchain.GetShardChainDatabase(shardID), shardID)
	if err != nil {
		return nil, 0, err
	}
	shardBlock, height, err := blockchain.GetShardBlockByHash(hash)
	if err != nil {
		return nil, 0, err
	}
	return shardBlock, height, nil
}

func (blockchain *BlockChain) DeleteShardBlockByView(shardID byte, view common.Hash) error {
	return rawdbv2.DeleteShardBlockByView(blockchain.GetShardChainDatabase(shardID), view)
}

func (blockchain *BlockChain) GetShardBlockByHeightAndView(shardID byte, height uint64, view common.Hash) (*ShardBlock, error) {
	finalShardBlock, finalHeight, err := blockchain.GetFinalizedShardBlock(shardID)
	if err != nil {
		return nil, err
	}
	if height == finalHeight {
		return finalShardBlock, nil
	}
	if height < finalHeight {
		shardBlocks, err := blockchain.GetShardBlockByHeight(height, shardID)
		if err != nil {
			return nil, err
		}
		shardBlock := NewShardBlock()
		for _, v := range shardBlocks {
			shardBlock = v
			break
		}
		return shardBlock, nil
	}
	if height > finalHeight {
		shardBlockIndexes, err := rawdbv2.GetShardBlockByView(blockchain.GetShardChainDatabase(shardID), view)
		if err != nil {
			return nil, err
		}
		if blockHash, ok := shardBlockIndexes[height]; !ok {
			return nil, fmt.Errorf("Shard %+v, Block Height %+v, View %+v, not found", shardID, height, view)
		} else {
			shardBlock, shardHeight, err := blockchain.GetShardBlockByHash(blockHash)
			if err != nil {
				return nil, err
			}
			if shardHeight != height {
				return nil, fmt.Errorf("Shard %+v, Block Height %+v, View %+v, not found", shardID, height, view)
			}
			return shardBlock, nil
		}
	}
	return nil, fmt.Errorf("Shard %+v, Block Height %+v, View %+v, not found", shardID, height, view)
}

func (blockchain *BlockChain) StoreBeaconBestState() error {
	beaconBestStateBytes, err := json.Marshal(blockchain.GetBeaconBestState())
	if err != nil {
		return err
	}
	return rawdbv2.StoreBeaconBestState(blockchain.GetBeaconChainDatabase(), beaconBestStateBytes)
}

func (blockchain *BlockChain) FinalizedBeaconBlock(beaconBlock *BeaconBlock) error {
	return rawdbv2.FinalizedBeaconBlock(blockchain.GetBeaconChainDatabase(), beaconBlock.Header.Hash())
}

func (blockchain *BlockChain) GetFinalizedBeaconBlock() (*BeaconBlock, uint64, error) {
	hash, err := rawdbv2.GetFinalizedBeaconBlock(blockchain.GetBeaconChainDatabase())
	if err != nil {
		return nil, 0, err
	}
	beaconBlock, height, err := blockchain.GetBeaconBlockByHash(hash)
	if err != nil {
		return nil, 0, err
	}
	return beaconBlock, height, nil
}

func (blockchain *BlockChain) DeleteBeaconBlockByView(view common.Hash) error {
	return rawdbv2.DeleteBeaconBlockByView(blockchain.GetBeaconChainDatabase(), view)
}

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

func (blockchain *BlockChain) GetShardBlockHeightByHash(hash common.Hash) (uint64, byte, error) {
	for _, v := range blockchain.GetShardIDs() {
		shardID := byte(v)
		height, index, err := rawdbv2.GetIndexOfBlock(blockchain.GetShardChainDatabase(shardID), hash)
		if err == nil {
			return height, index, nil
		}
	}
	return 0, 0, NewBlockChainError(GetShardBlockHeightByHashError, fmt.Errorf("Not found shard block height by hash %+v ", hash))
}

func (blockchain *BlockChain) GetBeaconBlockHashByHeight(height uint64) ([]common.Hash, error) {
	return rawdbv2.GetBeaconBlockHashByIndex(blockchain.GetBeaconChainDatabase(), height)
}

func (blockchain *BlockChain) GetBeaconBlockByHeight(height uint64) ([]*BeaconBlock, error) {
	if blockchain.IsTest {
		return []*BeaconBlock{}, nil
	}
	beaconBlocks := []*BeaconBlock{}
	beaconBlockHashes, err := rawdbv2.GetBeaconBlockHashByIndex(blockchain.GetBeaconChainDatabase(), height)
	if err != nil {
		return nil, err
	}
	for _, beaconBlockHash := range beaconBlockHashes {
		beaconBlock, _, err := blockchain.GetBeaconBlockByHash(beaconBlockHash)
		if err != nil {
			return nil, err
		}
		beaconBlocks = append(beaconBlocks, beaconBlock)
	}

	return beaconBlocks, nil
}

func (blockchain *BlockChain) GetBeaconBlockByHeightV1(height uint64) (*BeaconBlock, error) {
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

func (blockchain *BlockChain) GetShardBlockHashByHeight(height uint64, shardID byte) ([]common.Hash, error) {
	hashes := []common.Hash{}
	m, err := rawdbv2.GetShardBlockByIndex(blockchain.GetShardChainDatabase(shardID), shardID, height)
	if err != nil {
		return hashes, err
	}
	for k, _ := range m {
		hashes = append(hashes, k)
	}
	return hashes, nil
}

func (blockchain *BlockChain) GetShardBlockByHeight(height uint64, shardID byte) (map[common.Hash]*ShardBlock, error) {
	shardBlockMap := make(map[common.Hash]*ShardBlock)
	m, err := rawdbv2.GetShardBlockByIndex(blockchain.GetShardChainDatabase(shardID), shardID, height)
	if err != nil {
		return nil, err
	}
	for k, v := range m {
		shardBlock := NewShardBlock()
		err := json.Unmarshal(v, shardBlock)
		if err != nil {
			return nil, err
		}
		shardBlockMap[k] = shardBlock
	}
	return shardBlockMap, err
}

func (blockchain *BlockChain) GetShardBlockByHeightV1(height uint64, shardID byte) (*ShardBlock, error) {
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
	rootHash, err := blockchain.GetBeaconFeatureRootHash(blockchain.GetBeaconChainDatabase(), height)
	if err != nil {
		return nil, fmt.Errorf("Beacon Feature State DB not found, height %+v, error %+v", height, err)
	}
	return statedb.NewWithPrefixTrie(rootHash, statedb.NewDatabaseAccessWarper(db))
}

func (blockchain *BlockChain) GetBestBeaconSlashStateDB() *statedb.StateDB {
	return blockchain.GetBeaconBestState().slashStateDB
}

func (blockchain *BlockChain) GetBestBeaconRewardStateDB() *statedb.StateDB {
	return blockchain.GetBeaconBestState().rewardStateDB
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
