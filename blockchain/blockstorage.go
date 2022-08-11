package blockchain

import (
	"encoding/json"
	"github.com/incognitochain/incognito-chain/blockchain/types"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/config"
	"github.com/incognitochain/incognito-chain/dataaccessobject/flatfile"
	"github.com/incognitochain/incognito-chain/dataaccessobject/rawdbv2"
	"github.com/incognitochain/incognito-chain/incdb"
	"path"
)

type BlockStorage struct {
	keyValueDB  incdb.Database
	flatfile    flatfile.FlatFile
	cid         int
	useFF       bool
	useProtobuf bool
}

func NewBlockStorage(db incdb.Database, cid int, useFF, useProtoBuf bool) *BlockStorage {
	cfg := config.Config()
	p := path.Join(cfg.DataDir, cfg.DatabaseDir, "blockstorage")
	ff, _ := flatfile.NewFlatFile(p, 5000)

	return &BlockStorage{
		db, ff, cid, useFF, useProtoBuf,
	}
}

func (s *BlockStorage) StoreBlock(blk types.BlockInterface) error {
	if s.useFF {
		return s.storeBlockUsingFF(blk, s.useProtobuf)
	} else {
		return s.storeBlockUsingDB(blk)
	}
}

func (s *BlockStorage) GetBlock(blkHash common.Hash) (types.BlockInterface, int, error) {
	if s.useFF {
		return s.getBlockUsingFF(blkHash, s.useProtobuf)
	} else {
		return s.getBlockUsingDB(blkHash)
	}

}
func (s *BlockStorage) encode(blk types.BlockInterface) []byte {
	b, _ := json.Marshal(blk)
	return b
}

func (s *BlockStorage) decode(data []byte) (types.BlockInterface, error) {
	switch s.cid {
	case -1:
		beaconBlock := types.NewBeaconBlock()
		err := json.Unmarshal(data, beaconBlock)
		if err != nil {
			return nil, err
		}
		return beaconBlock, nil
	default:
		shardBlock := types.NewShardBlock()
		err := json.Unmarshal(data, shardBlock)
		if err != nil {
			return nil, err
		}
		return shardBlock, nil
	}
}

func (s *BlockStorage) storeBlockUsingFF(blk types.BlockInterface, useProtobuf bool) error {
	dataByte := s.encode(blk)
	ffIndex, err := s.flatfile.Append(dataByte)
	if err != nil {
		return err
	}
	if err := rawdbv2.StoreFlatFileIndexByBlockHash(s.keyValueDB, *blk.Hash(), ffIndex); err != nil {
		return err
	}
	return nil
}
func (s *BlockStorage) getBlockUsingFF(blkHash common.Hash, useProtobuf bool) (types.BlockInterface, int, error) {
	if ffIndex, err := rawdbv2.GetFlatFileIndexByBlockHash(s.keyValueDB, blkHash); err != nil {
		return nil, 0, err
	} else {
		data, err := s.flatfile.Read(ffIndex)
		if err != nil {
			return nil, 0, err
		}
		blk, err := s.decode(data)
		if err != nil {
			return nil, 0, err
		}
		return blk, len(data), nil
	}

}

func (s *BlockStorage) storeBlockUsingDB(blk types.BlockInterface) error {
	switch s.cid {
	case -1:
		if err := rawdbv2.StoreBeaconBlockByHash(s.keyValueDB, *blk.Hash(), blk); err != nil {
			return NewBlockChainError(StoreBeaconBlockError, err)
		}
	default:
		if err := rawdbv2.StoreShardBlock(s.keyValueDB, *blk.Hash(), blk); err != nil {
			return NewBlockChainError(StoreBeaconBlockError, err)
		}
	}
	return nil
}

func (s *BlockStorage) getBlockUsingDB(blkHash common.Hash) (types.BlockInterface, int, error) {
	switch s.cid {
	case -1:
		beaconBlockBytes, err := rawdbv2.GetBeaconBlockByHash(s.keyValueDB, blkHash)
		if err != nil {
			return nil, 0, err
		}
		beaconBlock, err := s.decode(beaconBlockBytes)
		if err != nil {
			return nil, 0, err
		}
		return beaconBlock, len(beaconBlockBytes), err
	default:
		shardBlockBytes, err := rawdbv2.GetShardBlockByHash(s.keyValueDB, blkHash)
		if err != nil {
			return nil, 0, err
		}
		shardBlock, err := s.decode(shardBlockBytes)
		if err != nil {
			return nil, 0, err
		}
		return shardBlock, len(shardBlockBytes), err
	}
}
