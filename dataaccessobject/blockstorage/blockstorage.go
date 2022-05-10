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
	rDB     incdb.Database
	sRDB    incdb.Database //mini raw db
	fDB     flatfile.FlatFile
	rmDB    flatfile.FlatFile
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

func (blkM *BlockManager) GetBlockByHash(
	hash *common.Hash,
	blkType proto.BlkType,
) (
	[]byte,
	[]byte,
	error,
) {
	if config.Config().EnableFFStorage {
		keyIdx := rawdbv2.GetHashToBlockIndexKey(*hash)
		// if v, has := blkM.cacher.Get(hash.String()); has {
		// 	if res, ok := v.([]byte); ok && (len(res) > 0) {
		// 		return res, nil
		// 	}
		// }
		blkIdBytes, err := blkM.sRDB.Get(keyIdx)
		if (err != nil) || (len(blkIdBytes) == 0) {
			return nil, nil, errors.Errorf("Can not get index for block hash %v, got %v, error %v", hash.String(), blkIdBytes, err)
		}
		blkID, err := common.BytesToUint64(blkIdBytes)
		if err != nil {
			return nil, nil, err
		}
		compBytes, err := blkM.fDB.Read(blkID)
		if err != nil {
			panic(err)
			return nil, nil, err
		}
		compBodyBytes := []byte{}
		if (config.Config().SyncMode != common.STATEDB_LITE_MODE) || (blkType != proto.BlkType_BlkBc) {
			compBodyBytes, err = blkM.rmDB.Read(blkID)
			if err != nil {
				panic(err)
				return nil, nil, err
			}
		}
		blkWithoutBody, err := common.GZipToBytes(compBytes)
		if err != nil {
			panic(err)
			return nil, nil, err
		}
		if len(compBodyBytes) > 0 {
			blkBody, err := common.GZipToBytes(compBodyBytes)
			if err != nil {
				panic(err)
				return nil, nil, err
			}
			return blkWithoutBody, blkBody, nil
		} else {
			return blkWithoutBody, compBodyBytes, nil
		}
	}
	if blkM.chainID == common.BeaconChainID {
		blk, err := rawdbv2.GetBeaconBlockByHash(blkM.rDB, *hash)
		return blk, nil, err
	}
	blk, err := rawdbv2.GetShardBlockByHash(blkM.rDB, *hash)
	return blk, nil, err
}

func (blkM *BlockManager) StoreBlock(
	blkType proto.BlkType,
	blkData types.BlockInterface,
) error {
	var err error
	blkHash := blkData.Hash()
	if config.Config().EnableFFStorage {
		var (
			compBytes    []byte
			blkIndex     uint64
			blkBodyBytes = []byte{}
			sizeBlk      = uint64(0)
			vData        = string("")
			blkBodyIdx   = uint64(0)
		)
		if blkData.GetHeight() > 1 {
			vData = blkData.GetValidationField()
			err = blkM.StoreBlockValidation(*blkHash, vData)
			if err != nil {
				return err
			}
			sizeBlk += uint64(len([]byte(vData)))
			blkData.AddValidationField("")
		}

		if blkBodyBytes, err = blkData.GetBodyBytes(); err != nil {
			return err
		}
		if (config.Config().SyncMode != common.STATEDB_LITE_MODE) || (blkType != proto.BlkType_BlkBc) {
			if len(blkBodyBytes) > 0 {
				if compBytes, err = common.GZipFromBytesWithLvl(blkBodyBytes, config.Param().FlatFileParam.CompLevel); err != nil {
					return err
				}
			} else {
				compBytes = blkBodyBytes
			}
			blkBodyIdx, err = blkM.rmDB.Append(compBytes)
			if err != nil {
				return err
			}
			sizeBlk += uint64(len(blkBodyBytes))
		}
		if blkData.GetHeight() != 1 {
			blkData.RemoveBody()
		}
		blkBytes, err := blkData.ToBytes()
		if err != nil {
			return err
		}
		if compBytes, err = common.GZipFromBytesWithLvl(blkBytes, config.Param().FlatFileParam.CompLevel); err != nil {
			return err
		}
		blkIndex, err = blkM.fDB.Append(compBytes)
		if err != nil {
			return err
		}
		if (blkBodyIdx != blkIndex) && (blkBodyIdx != 0) {
			panic(fmt.Errorf("Flatfile system corruption, %v - %v", blkBodyIdx, blkIndex))
		}
		sizeBlk += uint64(len(blkBytes))
		key := rawdbv2.GetHashToBlockIndexKey(*blkHash)
		if err = blkM.sRDB.Put(key, common.Uint64ToBytes(blkIndex)); err != nil {
			return err
		}
		blkData.AddValidationField(vData)
		if len(blkBodyBytes) != 0 {
			blkData.SetBodyFromBytes(blkBodyBytes)
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
