package blockchain

import (
	"encoding/json"
	"fmt"
	"os"
	"path"

	"github.com/incognitochain/incognito-chain/blockchain/types"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/flatfile"
	"github.com/incognitochain/incognito-chain/dataaccessobject/rawdbv2"
	"github.com/incognitochain/incognito-chain/incdb"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/incognitochain/incognito-chain/transaction"
	"github.com/pkg/errors"
)

type BlockStorage struct {
	rootDB         incdb.Database
	blockStorageDB incdb.Database
	flatfile       flatfile.FlatFile
	cid            int
	useFF          bool
}

func NewBlockStorage(db incdb.Database, ffPath string, cid int, useFF bool) *BlockStorage {
	var ff *flatfile.FlatFileManager
	var blockStorageDB incdb.Database

	ff, _ = flatfile.NewFlatFile(ffPath, 5000)
	blockStorageDB, _ = incdb.Open("leveldb", path.Join(ffPath, "blockKV"))

	return &BlockStorage{
		db, blockStorageDB, ff, cid, useFF,
	}
}

func (s *BlockStorage) ChangeMainDir(tmpDir, mainDir string) error {
	os.Rename(mainDir, mainDir+".bk")
	os.Rename(tmpDir, mainDir)
	blockStorageDB, _ := incdb.Open("leveldb", path.Join(mainDir, "blockstorage", "blockKV"))
	s.blockStorageDB = blockStorageDB
	s.flatfile, _ = flatfile.NewFlatFile(path.Join(mainDir, "blockstorage"), 5000)
	os.RemoveAll(mainDir + ".bk")

	return nil
}

func (s *BlockStorage) Truncate() error {
	return s.flatfile.Truncate(s.flatfile.Size() - 500)
}
func (s *BlockStorage) ReplaceBlock(blk types.BlockInterface) error {
	if s.useFF {
		if !s.IsExisted(*blk.Hash()) {
			return s.storeBlockUsingFF(blk)
		} else {
			return rawdbv2.StoreValidationDataByBlockHash(s.blockStorageDB, *blk.Hash(), []byte(blk.GetValidationField()))
		}
	} else {
		return s.storeBlockUsingDB(blk)
	}
}

func (s *BlockStorage) storeStakingTx(tx metadata.Transaction) error {
	if s.useFF {
		b, err := json.Marshal(tx)
		if err != nil {
			panic(err)
		}
		return rawdbv2.StoreStakingTx(s.blockStorageDB, byte(s.cid), *tx.Hash(), b)
	} else {
		return nil
	}
}

func (s *BlockStorage) GetStakingTx(hash common.Hash) (metadata.Transaction, error) {

	if s.useFF { //get from blockKV
		b, err := rawdbv2.GetStakingTx(s.blockStorageDB, byte(s.cid), hash)
		if err != nil {
			return nil, err
		}
		// fmt.Println(string(b))
		txChoice, parseErr := transaction.DeserializeTransactionJSON(b)
		if parseErr != nil {
			return nil, fmt.Errorf("unmarshall Json Shard Block Is Failed. Error %v", parseErr)
		}
		tx := txChoice.ToTx()
		if tx == nil {
			return nil, fmt.Errorf("unmarshall Json Shard Block Is Failed. Corrupted TX")
		}
		return tx, nil
	} else { //legacy flow, get from block
		blockHash, index, err := s.GetTXIndex(hash)
		if err != nil {
			return nil, err
		}
		blk, _, err := s.GetBlock(blockHash)
		if err != nil || blk == nil {
			Logger.log.Error("ERROR", err, "NO Transaction in block with hash", blockHash, "and index", index, "contains", index)
			return nil, err
		}
		shardBlock := blk.(*types.ShardBlock)
		tx := shardBlock.Body.Transactions[index]
		return tx, nil
	}
}

func (s *BlockStorage) StoreBeaconConfirmShardBlockByHeight(shardID byte, height uint64, hash common.Hash) error {
	if s.useFF {
		return rawdbv2.StoreBeaconConfirmInstantFinalityShardBlock(s.blockStorageDB, shardID, height, hash)
	} else {
		return rawdbv2.StoreBeaconConfirmInstantFinalityShardBlock(s.rootDB, shardID, height, hash)
	}
}

func (s *BlockStorage) GetBeaconConfirmShardBlockByHeight(shardID byte, height uint64) (*common.Hash, error) {
	if s.useFF {
		return rawdbv2.GetBeaconConfirmInstantFinalityShardBlock(s.blockStorageDB, shardID, height)
	} else {
		return rawdbv2.GetBeaconConfirmInstantFinalityShardBlock(s.rootDB, shardID, height)
	}
}

func (s *BlockStorage) GetBlockWithLatestValidationData(hash common.Hash) (types.BlockInterface, int, error) {
	if s.useFF {
		blk, size, err := s.getBlockUsingFF(hash)
		if err != nil {
			return nil, 0, errors.Wrap(err, "Cannot get block by hash (ffstorage)")
		}
		valString, err := rawdbv2.GetValidationDataByBlockHash(s.blockStorageDB, *blk.Hash())
		if err != nil {
			return blk, size, nil
		}
		blk.SetValidationField(string(valString))
		return blk, size, err
	} else {
		return s.getBlockUsingDB(hash)
	}
}

func (s *BlockStorage) StoreBlock(blk types.BlockInterface) error {
	if s.useFF {
		return s.storeBlockUsingFF(blk)
	} else {
		return s.storeBlockUsingDB(blk)
	}
}

func (s *BlockStorage) StoreTXIndex(blk *types.ShardBlock) error {
	for index, tx := range blk.Body.Transactions {
		//Logger.log.Infof("Process storing tx %v, index %x, shard %v, height %v, blockHash %v\n",
		//	tx.Hash().String(), index, blk.GetShardID(), blk.GetHeight(), blk.Hash().String())

		_, ok := tx.GetMetadata().(*metadata.StakingMetadata)
		if ok {
			if err := s.storeStakingTx(tx); err != nil {
				panic(err)
			}
		}

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

func (s *BlockStorage) StoreFinalizedBeaconBlock(index uint64, hash common.Hash) error {
	if s.useFF {
		if err := rawdbv2.StoreFinalizedBeaconBlockHashByIndex(s.blockStorageDB, index, hash); err != nil {
			panic(err)
		}
	} else {
		if err := rawdbv2.StoreFinalizedBeaconBlockHashByIndex(s.rootDB, index, hash); err != nil {
			panic(err)
		}
	}
	return nil
}

func (s *BlockStorage) GetFinalizedBeaconBlock(index uint64) (*common.Hash, error) {
	if s.useFF {
		return rawdbv2.GetFinalizedBeaconBlockHashByIndex(s.blockStorageDB, index)
	} else {
		return rawdbv2.GetFinalizedBeaconBlockHashByIndex(s.rootDB, index)
	}
}

func (s *BlockStorage) StoreFinalizedShardBlock(index uint64, hash common.Hash) error {
	if s.useFF {
		if err := rawdbv2.StoreFinalizedShardBlockHashByIndex(s.blockStorageDB, byte(s.cid), index, hash); err != nil {
			panic(err)
		}
	} else {
		if err := rawdbv2.StoreFinalizedShardBlockHashByIndex(s.rootDB, byte(s.cid), index, hash); err != nil {
			panic(err)
		}
	}
	return nil
}

func (s *BlockStorage) GetFinalizedShardBlockHashByIndex(index uint64) (*common.Hash, error) {
	if s.useFF {
		return rawdbv2.GetFinalizedShardBlockHashByIndex(s.blockStorageDB, byte(s.cid), index)
	} else {
		return rawdbv2.GetFinalizedShardBlockHashByIndex(s.rootDB, byte(s.cid), index)
	}
}

func (s *BlockStorage) IsExisted(blkHash common.Hash) bool {
	if s.useFF {
		if _, err := rawdbv2.GetFlatFileIndexByBlockHash(s.blockStorageDB, blkHash); err != nil {
			return false
		}
		return true
	} else {
		return s.checkBlockExistUsingDB(blkHash)
	}
}

func (s *BlockStorage) GetBlock(blkHash common.Hash) (types.BlockInterface, int, error) {
	if s.useFF {
		return s.getBlockUsingFF(blkHash)
	} else {
		return s.getBlockUsingDB(blkHash)
	}
}
func (s *BlockStorage) encode(blk types.BlockInterface) []byte {
	b, _ := json.Marshal(blk)
	var bb = b

	if s.useFF {
		//zip
		var err error
		bb, err = common.GZipFromBytes(b)
		if err != nil {
			panic(err)
		}
	}

	return bb
}

func (s *BlockStorage) decode(data []byte) (types.BlockInterface, error) {
	//unzip if using ffstorage
	var rawData = data
	if s.useFF {
		var err error
		rawData, err = common.GZipToBytes(data)
		if err != nil {
			panic(err)
		}
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

func (s *BlockStorage) storeBlockUsingFF(blk types.BlockInterface) error {
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

func (s *BlockStorage) getBlockUsingFF(blkHash common.Hash) (types.BlockInterface, int, error) {
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

func (s *BlockStorage) getFFFileByHeight(blkHeight uint64) (int, error) {
	return int(blkHeight / s.flatfile.FileSize()), nil
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
