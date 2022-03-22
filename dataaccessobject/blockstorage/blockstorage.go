package blockstorage

import (
	"github.com/incognitochain/incognito-chain/blockchain/types"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/config"
	"github.com/incognitochain/incognito-chain/dataaccessobject/flatfile"
	"github.com/incognitochain/incognito-chain/dataaccessobject/rawdbv2"
	"github.com/incognitochain/incognito-chain/incdb"
	"github.com/incognitochain/incognito-chain/proto"
	"github.com/pkg/errors"
)

type BlockService interface {
	GetBlockByHash(
		hash *common.Hash,
	) (
		[]byte,
		error,
	)
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
	fDB     flatfile.FlatFile
	cacher  common.Cacher
}

func NewBlockService(
	rawDB incdb.Database,
	flatfileManager flatfile.FlatFile,
	chainID int,
) (
	BlockService,
	error,
) {
	mCache, err := common.NewRistrettoMemCache(common.CacheMaxCost)
	if err != nil {
		return nil, err
	}
	res := &BlockManager{
		chainID: chainID,
		rDB:     rawDB,
		fDB:     flatfileManager,
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
		_, err := blkM.rDB.Get(keyIdx)
		if err != nil {
			return false, err
		}
		return true, nil
	}

	//else
	if blkM.chainID == common.BeaconChainID {
		return rawdbv2.HasBeaconBlock(blkM.rDB, *hash)
	} else {
		return rawdbv2.HasShardBlock(blkM.rDB, *hash)
	}
}

func (blkM *BlockManager) GetBlockByHash(
	hash *common.Hash,
) (
	[]byte,
	error,
) {
	if config.Config().EnableFFStorage {
		keyIdx := rawdbv2.GetHashToBlockIndexKey(*hash)
		if v, has := blkM.cacher.Get(hash.String()); has {
			if res, ok := v.([]byte); ok && (len(res) > 0) {
				return res, nil
			}
		}
		blkIdBytes, err := blkM.rDB.Get(keyIdx)
		if (err != nil) || (len(blkIdBytes) == 0) {
			return nil, errors.Errorf("Can not get index for block hash %v, got %v, error %v", hash.String(), blkIdBytes, err)
		}
		blkID, err := common.BytesToUint64(blkIdBytes)
		if err != nil {
			return nil, err
		}
		return blkM.fDB.Read(blkID)
	}
	if blkM.chainID == common.BeaconChainID {
		return rawdbv2.GetBeaconBlockByHash(blkM.rDB, *hash)
	}
	return rawdbv2.GetShardBlockByHash(blkM.rDB, *hash)
}

func (blkM *BlockManager) StoreBlock(
	blkType proto.BlkType,
	blkData types.BlockInterface,
) error {
	blkBytes, err := blkData.ToBytes()
	if err != nil {
		return err
	}
	blkHash := blkData.Hash()
	if config.Config().EnableFFStorage {
		blkIndex, err := blkM.fDB.Append(blkBytes)
		if err != nil {
			return err
		}
		key := rawdbv2.GetHashToBlockIndexKey(*blkHash)
		err = blkM.rDB.Put(key, common.Uint64ToBytes(blkIndex))
		if err != nil {
			return err
		}
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
	blkM.cacher.Set(blkHash.String(), blkBytes, int64(len(blkBytes)))
	return nil
}
