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
	return rawdbv2.StoreShardBestState(blockchain.GetDatabase(), shardID, blockchain.BestState.Shard[shardID])
}

func (blockchain *BlockChain) StoreBeaconBestState() error {
	beaconBestStateBytes, err := json.Marshal(blockchain.BestState.Beacon)
	if err != nil {
		return err
	}
	return rawdbv2.StoreBeaconBestState(blockchain.config.DataBase, beaconBestStateBytes)
}

func (blockchain *BlockChain) GetBlockHeightByBlockHash(hash common.Hash) (uint64, byte, error) {
	return rawdbv2.GetIndexOfBlock(blockchain.GetDatabase(), hash)
}

func (blockchain *BlockChain) GetBeaconBlockHashByHeight(height uint64) ([]common.Hash, error) {
	return rawdbv2.GetBeaconBlockHashByIndex(blockchain.GetDatabase(), height)
}

func (blockchain *BlockChain) GetBeaconBlockByHeight(height uint64) ([]*BeaconBlock, error) {
	if blockchain.IsTest {
		return []*BeaconBlock{}, nil
	}
	beaconBlocks := []*BeaconBlock{}
	beaconBlockHashes, err := rawdbv2.GetBeaconBlockHashByIndex(blockchain.GetDatabase(), height)
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
	beaconBlockBytes, err := rawdbv2.GetBeaconBlockByHash(blockchain.GetDatabase(), beaconBlockHash)
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

func (blockchain *BlockChain) GetShardBlockHeightByHash(hash common.Hash) (uint64, byte, error) {
	return rawdbv2.GetIndexOfBlock(blockchain.GetDatabase(), hash)
}

func (blockchain *BlockChain) GetShardBlockHashByHeight(height uint64, shardID byte) ([]common.Hash, error) {
	hashes := []common.Hash{}
	m, err := rawdbv2.GetShardBlockByIndex(blockchain.GetDatabase(), shardID, height)
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
	m, err := rawdbv2.GetShardBlockByIndex(blockchain.GetDatabase(), shardID, height)
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

func (blockchain *BlockChain) GetShardBlockByHash(hash common.Hash) (*ShardBlock, uint64, error) {
	if blockchain.IsTest {
		return &ShardBlock{}, 2, nil
	}
	shardBlockBytes, err := rawdbv2.GetShardBlockByHash(blockchain.config.DataBase, hash)
	if err != nil {
		return nil, 0, err
	}
	shardBlock := NewShardBlock()
	err = json.Unmarshal(shardBlockBytes, shardBlock)
	if err != nil {
		return nil, 0, err
	}
	return shardBlock, shardBlock.Header.Height, nil
}

func (blockchain *BlockChain) GetShardRewardStateDB(shardID byte) *statedb.StateDB {
	return blockchain.BestState.Shard[shardID].GetCopiedRewardStateDB()
}

func (blockchain *BlockChain) GetTransactionStateDB(shardID byte) *statedb.StateDB {
	return blockchain.BestState.Shard[shardID].GetCopiedTransactionStateDB()
}

func (blockchain *BlockChain) GetShardFeatureStateDB(shardID byte) *statedb.StateDB {
	return blockchain.BestState.Shard[shardID].GetCopiedFeatureStateDB()
}

func (blockchain *BlockChain) GetBeaconFeatureStateDB() *statedb.StateDB {
	return blockchain.BestState.Beacon.GetCopiedFeatureStateDB()
}

func (blockchain *BlockChain) GetBeaconFeatureStateDBByHeight(height uint64, db incdb.Database) (*statedb.StateDB, error) {
	rootHash, err := blockchain.GetBeaconFeatureRootHash(blockchain.GetDatabase(), height)
	if err != nil {
		return nil, fmt.Errorf("Beacon Feature State DB not found, height %+v, error %+v", height, err)
	}
	return statedb.NewWithPrefixTrie(rootHash, statedb.NewDatabaseAccessWarper(db))
}

func (blockchain *BlockChain) GetBeaconSlashStateDB() *statedb.StateDB {
	return blockchain.BestState.Beacon.slashStateDB
}

func (blockchain *BlockChain) GetBeaconRewardStateDB() *statedb.StateDB {
	return blockchain.BestState.Beacon.rewardStateDB
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