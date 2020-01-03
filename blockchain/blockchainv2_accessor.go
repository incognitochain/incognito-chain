package blockchain

import (
	"encoding/json"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/rawdbv2"
)

// Database Accessor, package, which import blockchain package, should invoke database via these accessor
func (blockchain *BlockChain) GetBlockHeightByBlockHashV2(hash common.Hash) (uint64, byte, error) {
	return rawdbv2.GetIndexOfBlock(blockchain.GetDatabase(), hash)
}

func (blockchain *BlockChain) GetBeaconBlockHashByHeightV2(height uint64) ([]common.Hash, error) {
	return rawdbv2.GetBeaconBlockHashByIndex(blockchain.GetDatabase(), height)
}

func (blockchain *BlockChain) GetBeaconBlockByHeightV2(height uint64) ([]*BeaconBlock, error) {
	if blockchain.IsTest {
		return []*BeaconBlock{}, nil
	}
	beaconBlocks := []*BeaconBlock{}
	beaconBlockHashes, err := rawdbv2.GetBeaconBlockHashByIndex(blockchain.GetDatabase(), height)
	if err != nil {
		return nil, err
	}
	for _, beaconBlockHash := range beaconBlockHashes {
		beaconBlock, _, err := blockchain.GetBeaconBlockByHashV2(beaconBlockHash)
		if err != nil {
			return nil, err
		}
		beaconBlocks = append(beaconBlocks, beaconBlock)
	}

	return beaconBlocks, nil
}

func (blockchain *BlockChain) GetBeaconBlockByHashV2(beaconBlockHash common.Hash) (*BeaconBlock, uint64, error) {
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

func (blockchain *BlockChain) GetShardBlockHeightByHashV2(hash common.Hash) (uint64, byte, error) {
	return rawdbv2.GetIndexOfBlock(blockchain.GetDatabase(), hash)
}

func (blockchain *BlockChain) GetShardBlockHashByHeightV2(height uint64, shardID byte) ([]common.Hash, error) {
	hashes := []common.Hash{}
	m, err := rawdbv2.GetBlockByIndex(blockchain.GetDatabase(), shardID, height)
	if err != nil {
		return hashes, err
	}
	for k, _ := range m {
		hashes = append(hashes, k)
	}
	return hashes, nil
}

func (blockchain *BlockChain) GetShardBlockByHeightV2(height uint64, shardID byte) (map[common.Hash]*ShardBlock, error) {
	shardBlockMap := make(map[common.Hash]*ShardBlock)
	m, err := rawdbv2.GetBlockByIndex(blockchain.GetDatabase(), shardID, height)
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

func (blockchain *BlockChain) GetShardBlockByHashV2(hash common.Hash) (*ShardBlock, uint64, error) {
	if blockchain.IsTest {
		return &ShardBlock{}, 2, nil
	}
	shardBlockBytes, err := rawdbv2.FetchBlock(blockchain.config.DataBase, hash)
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
