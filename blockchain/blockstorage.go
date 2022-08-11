package blockchain

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"github.com/incognitochain/incognito-chain/blockchain/types"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/flatfile"
	"github.com/incognitochain/incognito-chain/dataaccessobject/rawdbv2"
	"github.com/incognitochain/incognito-chain/incdb"
	"io/ioutil"
	"path"
)

type BlockStorage struct {
	rootDB         incdb.Database
	blockStorageDB incdb.Database
	flatfile       flatfile.FlatFile
	cid            int
	useFF          bool
	useProtobuf    bool
}

func NewBlockStorage(db incdb.Database, ffPath string, cid int, useFF, useProtoBuf bool) *BlockStorage {
	var ff *flatfile.FlatFileManager
	var blockStorageDB incdb.Database

	ff, _ = flatfile.NewFlatFile(ffPath, 5000)
	blockStorageDB, _ = incdb.Open("leveldb", path.Join(ffPath, "blockKV"))

	return &BlockStorage{
		db, blockStorageDB, ff, cid, useFF, useProtoBuf,
	}
}

func (s *BlockStorage) StoreBlock(blk types.BlockInterface) error {
	if s.useFF {
		return s.storeBlockUsingFF(blk, s.useProtobuf)
	} else {
		return s.storeBlockUsingDB(blk)
	}
}

func (s *BlockStorage) StoreTXIndex(blk *types.ShardBlock) error {
	for index, tx := range blk.Body.Transactions {
		Logger.log.Infof("Process storing tx %v, index %x, shard %v, height %v, blockHash %v\n",
			tx.Hash().String(), index, blk.GetShardID(), blk.GetHeight(), blk.Hash().String())
		if s.useFF {
			if err := rawdbv2.StoreTransactionIndex(s.blockStorageDB, *tx.Hash(), blk.Header.Hash(), index); err != nil {
				panic(err)
			}
		} else {
			if err := rawdbv2.StoreTransactionIndex(s.rootDB, *tx.Hash(), blk.Header.Hash(), index); err != nil {
				panic(err)
			}
		}
	}
	return nil
}

func (s *BlockStorage) GetTXIndex(tx common.Hash) (common.Hash, int, error) {
	if s.useFF {
		return rawdbv2.GetTransactionByHash(s.blockStorageDB, tx)
	} else {
		return rawdbv2.GetTransactionByHash(s.rootDB, tx)
	}
}

func (s *BlockStorage) StoreFinalizedShardBlock(sid byte, index uint64, hash common.Hash) error {
	if s.useFF {
		if err := rawdbv2.StoreFinalizedShardBlockHashByIndex(s.blockStorageDB, sid, index, hash); err != nil {
			panic(err)
		}
	} else {
		if err := rawdbv2.StoreFinalizedShardBlockHashByIndex(s.rootDB, sid, index, hash); err != nil {
			panic(err)
		}
	}
	return nil
}

func (s *BlockStorage) GetFinalizedShardBlock(sid byte, index uint64) (*common.Hash, error) {
	if s.useFF {
		return rawdbv2.GetFinalizedShardBlockHashByIndex(s.blockStorageDB, sid, index)
	} else {
		return rawdbv2.GetFinalizedShardBlockHashByIndex(s.rootDB, sid, index)
	}
}

func (s *BlockStorage) IsExisted(blkHash common.Hash) bool {
	if s.useFF {
		if _, err := rawdbv2.GetFlatFileIndexByBlockHash(s.rootDB, blkHash); err != nil {
			return false
		}
		return true
	} else {
		return s.checkBlockExistUsingDB(blkHash)
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
	//zip
	var bb bytes.Buffer
	gz := gzip.NewWriter(&bb)
	if _, err := gz.Write(b); err != nil {
		panic(err)
	}
	if err := gz.Close(); err != nil {
		panic(err)
	}

	return bb.Bytes()
}

func (s *BlockStorage) decode(data []byte) (types.BlockInterface, error) {
	//unzip
	reader := bytes.NewReader([]byte(data))
	gzreader, e1 := gzip.NewReader(reader)
	if e1 != nil {
		panic(e1)
	}
	rawData, e2 := ioutil.ReadAll(gzreader)
	if e2 != nil {
		panic(e2)
	}

	switch s.cid {
	case -1:
		beaconBlock := types.NewBeaconBlock()
		err := json.Unmarshal(rawData, beaconBlock)
		if err != nil {
			return nil, err
		}
		return beaconBlock, nil
	default:
		shardBlock := types.NewShardBlock()
		err := json.Unmarshal(rawData, shardBlock)
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
	if err := rawdbv2.StoreFlatFileIndexByBlockHash(s.blockStorageDB, *blk.Hash(), ffIndex); err != nil {
		return err
	}
	return nil
}
func (s *BlockStorage) getBlockUsingFF(blkHash common.Hash, useProtobuf bool) (types.BlockInterface, int, error) {
	if ffIndex, err := rawdbv2.GetFlatFileIndexByBlockHash(s.blockStorageDB, blkHash); err != nil {
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
		if err := rawdbv2.StoreBeaconBlockByHash(s.rootDB, *blk.Hash(), blk); err != nil {
			return NewBlockChainError(StoreBeaconBlockError, err)
		}
	default:
		if err := rawdbv2.StoreShardBlock(s.rootDB, *blk.Hash(), blk); err != nil {
			return NewBlockChainError(StoreBeaconBlockError, err)
		}
	}
	return nil
}

func (s *BlockStorage) getBlockUsingDB(blkHash common.Hash) (types.BlockInterface, int, error) {
	switch s.cid {
	case -1:
		beaconBlockBytes, err := rawdbv2.GetBeaconBlockByHash(s.rootDB, blkHash)
		if err != nil {
			return nil, 0, err
		}
		beaconBlock, err := s.decode(beaconBlockBytes)
		if err != nil {
			return nil, 0, err
		}
		return beaconBlock, len(beaconBlockBytes), err
	default:
		shardBlockBytes, err := rawdbv2.GetShardBlockByHash(s.rootDB, blkHash)
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

func (s *BlockStorage) checkBlockExistUsingDB(blkHash common.Hash) bool {
	switch s.cid {
	case -1:
		exist, _ := rawdbv2.HasBeaconBlock(s.rootDB, blkHash)
		return exist
	default:
		exist, _ := rawdbv2.HasShardBlock(s.rootDB, blkHash)
		return exist
	}
}
