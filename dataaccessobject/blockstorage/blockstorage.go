package blockstorage

import (
	"fmt"
	"path"

	"github.com/incognitochain/incognito-chain/blockchain/types"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/config"
	"github.com/incognitochain/incognito-chain/consensus_v2/consensustypes"
	"github.com/incognitochain/incognito-chain/dataaccessobject/flatfile"
	"github.com/incognitochain/incognito-chain/dataaccessobject/rawdbv2"
	"github.com/incognitochain/incognito-chain/incdb"
	"github.com/incognitochain/incognito-chain/proto"
	"github.com/pkg/errors"
)

type BlockService interface {
	GetBlockByHash(
		hash *common.Hash,
		blkType proto.BlkType,
	) (
		types.BlockInterface,
		[]byte,
		[]byte,
		error,
	)
	GetBlockValidation(
		blkHash common.Hash,
	) (string, error)
	StoreBlockValidation(
		blkHash common.Hash,
		valData string,
	) error
	CheckBlockByHash(
		hash *common.Hash,
	) (
		bool,
		error,
	)
	StoreBlock(
		blkType proto.BlkType,
		blkData types.BlockInterface,
	) error
}

type BlockInfor struct {
	Height   uint64
	Hash     common.Hash
	PrevHash common.Hash
}

type BlockManager struct {
	chainID int
	rDB     incdb.Database    //rawDB, store block with mode ff=false
	sRDB    incdb.Database    //mini raw db, store information of ff system (blk index, validation data,...)
	fDB     flatfile.FlatFile // store block header
	rmDB    flatfile.FlatFile // removable ff system (store block body)
	cacher  common.Cacher
}

func NewBlockService(
	rawDB incdb.Database,
	flatfileManager flatfile.FlatFile,
	removableDB flatfile.FlatFile,
	chainID int,
) (
	BlockService,
	error,
) {
	mCache, err := common.NewRistrettoMemCache(int64(config.Param().FlatFileParam.MaxCacheSize / uint64(config.Param().ActiveShards+1)))
	if err != nil {
		return nil, err
	}
	dbPath := path.Join(rawDB.GetPath(), "ffdata")
	subDB, err := incdb.OpenDBWithPath("leveldb", dbPath)
	if err != nil {
		return nil, err
	}
	res := &BlockManager{
		chainID: chainID,
		rDB:     rawDB,
		sRDB:    subDB,
		fDB:     flatfileManager,
		rmDB:    removableDB,
		cacher:  mCache,
	}
	return res, nil
}

func (blkM *BlockManager) CheckBlockByHash(
	hash *common.Hash,
) (
	bool,
	error,
) {
	if config.Config().EnableFFStorage {
		keyIdx := rawdbv2.GetHashToBlockIndexKey(*hash)
		_, err := blkM.sRDB.Get(keyIdx)
		if err != nil {
			return false, err
		}
		return true, nil
	}
	if blkM.chainID == common.BeaconChainID {
		return rawdbv2.HasBeaconBlock(blkM.rDB, *hash)
	} else {
		return rawdbv2.HasShardBlock(blkM.rDB, *hash)
	}
}

func (blkM *BlockManager) getBlkIndex(blkHash *common.Hash) (uint64, error) {
	keyIdx := rawdbv2.GetHashToBlockIndexKey(*blkHash)
	blkIdBytes, err := blkM.sRDB.Get(keyIdx)
	if (err != nil) || (len(blkIdBytes) == 0) {
		return 0, errors.Errorf("Can not get index for block hash %v, got %v, error %v", blkHash.String(), blkIdBytes, err)
	}
	return common.BytesToUint64(blkIdBytes)
}

func (blkM *BlockManager) storeBlkIndex(index uint64, blkHash *common.Hash) error {
	key := rawdbv2.GetHashToBlockIndexKey(*blkHash)
	return blkM.sRDB.Put(key, common.Uint64ToBytes(index))
}

func (blkM *BlockManager) GetBlockByHash(
	hash *common.Hash,
	blkType proto.BlkType,
) (
	types.BlockInterface,
	[]byte,
	[]byte,
	error,
) {
	if config.Config().EnableFFStorage {
		if blk, cached := blkM.cacher.Get(hash.String()); (blk != nil) && (cached) {
			if blkI, ok := blk.(types.BlockInterface); ok {
				return blkI, nil, nil, nil
			}
		}
		blkID, err := blkM.getBlkIndex(hash)
		if err != nil {
			err = errors.Wrap(err, "Can not get block index")
			return nil, nil, nil, err
		}
		headerBytes, err := blkM.fDB.Read(blkID)
		if err != nil {
			err = errors.Wrap(err, "Can not get header data")
			return nil, nil, nil, err
		}
		bodyBytes := []byte{}
		if (config.Config().SyncMode != common.STATEDB_LITE_MODE) || (blkType != proto.BlkType_BlkBc) {
			if bodyBytes, err = blkM.rmDB.Read(blkID); err != nil {
				err = errors.Wrap(err, "Can not get body data")
				return nil, nil, nil, err
			}
		}
		if headerBytes, err = common.GZipToBytes(headerBytes); err != nil {
			err = errors.Wrap(err, "Can not unzip header data")
			return nil, nil, nil, err
		} else {
			if len(bodyBytes) > 0 {
				if bodyBytes, err := common.GZipToBytes(bodyBytes); err != nil {
					err = errors.Wrapf(err, "Can not unzip body %+v", bodyBytes)
					return nil, nil, nil, err
				}
			}
			return nil, headerBytes, bodyBytes, nil
		}
	}
	if blkM.chainID == common.BeaconChainID {
		blk, err := rawdbv2.GetBeaconBlockByHash(blkM.rDB, *hash)
		return nil, blk, nil, err
	}
	blk, err := rawdbv2.GetShardBlockByHash(blkM.rDB, *hash)
	return nil, blk, nil, err
}

func (blkM *BlockManager) StoreBlock(
	blkType proto.BlkType,
	blkData types.BlockInterface,
) error {
	var err error
	blkHash := blkData.Hash()
	if config.Config().EnableFFStorage {
		var (
			blkHeaderBytes = []byte{}
			blkIndex       uint64
			blkBodyBytes   = []byte{}
			blkBodyIdx     = uint64(0)
			sizeBlk        = uint64(0)
			vData          = string("")
		)
		if blkData.GetHeight() > 1 {
			vData = blkData.GetValidationField()
			if err = blkM.StoreBlockValidation(*blkHash, vData); err != nil {
				return err
			}
			sizeBlk += uint64(len([]byte(vData)))
		}
		if blkHeaderBytes, err = blkData.GetHeaderBytes(); err != nil {
			return err
		}
		sizeBlk += uint64(len(blkHeaderBytes))
		if blkHeaderBytes, err = common.GZipFromBytesWithLvl(blkHeaderBytes, config.Param().FlatFileParam.CompLevel); err != nil {
			return err
		}
		if (config.Config().SyncMode != common.STATEDB_LITE_MODE) || (blkType != proto.BlkType_BlkBc) {
			if blkBodyBytes, err = blkData.GetBodyBytes(); err != nil {
				return err
			} else {
				fmt.Printf("block %v %+v", blkData.Hash().String(), blkBodyBytes)
				if len(blkBodyBytes) > 0 {
					if blkBodyBytes, err = common.GZipFromBytesWithLvl(blkBodyBytes, config.Param().FlatFileParam.CompLevel); err != nil {
						return err
					}
					sizeBlk += uint64(len(blkBodyBytes))
				}
				blkBodyIdx, err = blkM.rmDB.Append(blkBodyBytes)
				if err != nil {
					return err
				}
			}
		}
		if blkIndex, err = blkM.fDB.Append(blkHeaderBytes); err != nil {
			return err
		}
		if (blkBodyIdx != blkIndex) && (blkBodyIdx != 0) {
			panic(fmt.Errorf("Flatfile system corruption, %v - %v", blkBodyIdx, blkIndex))
		}
		if err = blkM.storeBlkIndex(blkIndex, blkHash); err != nil {
			return err
		}
		blkM.cacher.Set(blkHash.String(), blkData, int64(sizeBlk))
	} else {
		if blkType == proto.BlkType_BlkShard {
			err = rawdbv2.StoreShardBlock(blkM.rDB, *blkHash, blkData)
		} else {
			err = rawdbv2.StoreBeaconBlockByHash(blkM.rDB, *blkHash, blkData)
		}
		if err != nil {
			return err
		}
	}
	return nil
}

func (blkM *BlockManager) StoreBlockValidation(
	blkHash common.Hash,
	valData string,
) error {
	vData, err := consensustypes.DecodeValidationData(valData)
	if err != nil {
		return err
	}
	vDataBytes, err := vData.ToBytes()
	if err != nil {
		return err
	}
	key := rawdbv2.GetHashToBlockValidationKey(blkHash)
	err = blkM.sRDB.Put(key, vDataBytes)
	if err != nil {
		return err
	}
	blkM.cacher.Set(key, valData, int64(len([]byte(valData))))
	return nil
}

func (blkM *BlockManager) GetBlockValidation(
	blkHash common.Hash,
) (string, error) {
	key := rawdbv2.GetHashToBlockValidationKey(blkHash)
	rawValData, existed := blkM.cacher.Get(key)
	if existed {
		if valData, ok := rawValData.(string); ok {
			return valData, nil
		}
	}
	valDataBytes, err := blkM.sRDB.Get(key)
	if err != nil {
		return "", err
	}
	valData := consensustypes.ValidationData{}
	err = valData.FromBytes(valDataBytes)
	if err != nil {
		return "", err
	}

	valDataStr, err := consensustypes.EncodeValidationData(valData)
	if err != nil {
		return "", err
	}
	blkM.cacher.Set(key, valDataStr, int64(len([]byte(valDataStr))))
	return valDataStr, nil
}
