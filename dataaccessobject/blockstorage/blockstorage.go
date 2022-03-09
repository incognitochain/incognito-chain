package blockstorage

import (
	"fmt"
	"sync"

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
	GetBlockByHeight(
		height uint64,
		cID byte,
	) (
		[]byte,
		error,
	)
	GetBlockByHash(
		hash *common.Hash,
		cID byte,
	) (
		[]byte,
		error,
	)
	GetPrevHashByHash(
		hash *common.Hash,
	) (
		common.Hash,
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
	MarkFinalized(
		height uint64,
		hash common.Hash,
		cID byte,
	) error
}

type BlockInfor struct {
	Height   uint64
	Hash     common.Hash
	PrevHash common.Hash
}

type BlockManager struct {
	rDB            incdb.Database
	fDB            flatfile.FlatFile
	cacher         common.Cacher
	locker         *sync.RWMutex
	hashByHeight   map[uint64][]common.Hash
	prevHashByHash map[common.Hash]common.Hash
	finalHeight    uint64
}

func NewBlockService(
	rawDB incdb.Database,
	flatfileManager flatfile.FlatFile,
) (
	BlockService,
	error,
) {
	mCache, err := common.NewRistrettoMemCache(common.CacheMaxCost)
	if err != nil {
		return nil, err
	}
	res := &BlockManager{
		rDB:            rawDB,
		fDB:            flatfileManager,
		cacher:         mCache,
		locker:         &sync.RWMutex{},
		hashByHeight:   map[uint64][]common.Hash{},
		prevHashByHash: map[common.Hash]common.Hash{},
		finalHeight:    1,
	}
	return res, nil
}

// func (blkM *BlockManager) GetBlockFinalByHeight(
// 	height uint64,
// 	cID int,
// ) (
// 	[]byte,
// 	error,
// ) {
// 	key := rawdbv2.GetHeightToBlockIndexKey(height, cID)
// 	_, err := rawdbv2.GetFinalizedShardBlockHashByIndex(blkM.rDB, byte(cID), height)
// 	if err != nil {
// 		return nil, err
// 	}
// 	blkHashBytes, err := blkM.rDB.Get(key)
// 	if err != nil {
// 		return nil, errors.Wrapf(err, "can not get block %v of cID %v", height, cID)
// 	}
// 	blkHash := common.Hash{}
// 	copy(blkHash[:], blkHashBytes)
// 	return blkM.GetBlockByHash(&blkHash)
// }

func (blkM *BlockManager) GetBlockByHeight(
	height uint64,
	cID byte,
) (
	[]byte,
	error,
) {
	blkHash := &common.Hash{}
	var err error
	if height <= blkM.finalHeight {
		if cID == common.BeaconChainSyncID {
			blkHash, err = rawdbv2.GetFinalizedBeaconBlockHashByIndex(blkM.rDB, height)
		} else {
			blkHash, err = rawdbv2.GetFinalizedShardBlockHashByIndex(blkM.rDB, cID, height)
		}
	} else {
		if blkHashs, ok := blkM.hashByHeight[height]; ok && (len(blkHashs) > 0) {
			blkHash = &blkHashs[0]
			err = nil
		} else {
			err = errors.Errorf("Can not found blk hash for blk height %v of cID %v", cID)
		}
	}
	if err != nil {
		return nil, err
	}
	return blkM.GetBlockByHash(blkHash, cID)
}

func (blkM *BlockManager) CheckBlockByHash(
	hash *common.Hash,
) (
	bool,
	error,
) {
	keyIdx := rawdbv2.GetHashToBlockIndexKey(*hash)
	fmt.Printf("[testFF] Get blk %v, key %v\n", hash.String(), common.HashH(keyIdx).String())
	_, err := blkM.rDB.Get(keyIdx)
	if err != nil {
		return false, err
	}

	return true, nil
}

func (blkM *BlockManager) GetBlockByHash(
	hash *common.Hash,
	cID byte,
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
	if cID == common.BeaconChainSyncID {
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
	blkHeight := blkData.GetHeight()
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
			err = rawdbv2.StoreShardBlock(blkM.rDB, *blkHash, blkBytes)
		} else {
			err = rawdbv2.StoreBeaconBlockByHash(blkM.rDB, *blkHash, blkBytes)
		}
		if err != nil {
			return err
		}
	}
	blkM.locker.Lock()
	blkM.hashByHeight[blkHeight] = append(blkM.hashByHeight[blkHeight], *blkHash)
	blkM.prevHashByHash[*blkHash] = blkData.GetPrevHash()
	blkM.locker.Unlock()
	blkM.cacher.Set(blkHash.String(), blkBytes, int64(len(blkBytes)))
	return nil
}

// func (blkM *BlockManager) storeBlockHeightFinalized(
// 	cID int,
// 	blkHeight uint64,
// 	blkHash common.Hash,
// ) error {
// 	key := rawdbv2.GetHeightToBlockIndexKey(blkHeight, cID)
// 	err := blkM.rDB.Put(key, blkHash[:])
// 	return err
// }
func (blkM *BlockManager) MarkFinalized(
	blkHeight uint64,
	blkHash common.Hash,
	cID byte,
) error {
	// curFinalHeight := blkM.finalHeight
	// for height := blkHeight; height >= curFinalHeight; height-- {
	// 	err := blkM.storeBlockHeightFinalized(int(cID), height, blkHash)
	// 	if err != nil {
	// 		return err
	// 	}
	// 	pHash, ok := blkM.prevHashByHash[blkHash]
	// 	fmt.Printf("cID %v height %v testdelete %v prev %v \n", cID, height, blkHash.String(), pHash.String())
	// 	needToRemove := blkM.hashByHeight[height]

	// 	for _, hash := range needToRemove {
	// 		delete(blkM.prevHashByHash, hash)
	// 		if hash != blkHash {
	// 			fmt.Printf("cID %v height %v testdelete delete %v - %v\n", cID, height, hash.String(), blkHash.String())
	// 			key := rawdbv2.GetHashToBlockIndexKey(hash)
	// 			err := blkM.rDB.Delete(key)
	// 			if err != nil {
	// 				panic(err)
	// 			}
	// 		}
	// 	}
	// 	if ok {
	// 		blkHash = pHash
	// 	}
	// }
	blkM.finalHeight = blkHeight
	return nil
}

func (blkM *BlockManager) GetPrevHashByHash(
	hash *common.Hash,
) (
	common.Hash,
	error,
) {
	blkM.locker.RLock()
	prevH, existed := blkM.prevHashByHash[*hash]
	blkM.locker.RUnlock()
	if !existed {
		return common.Hash{}, errors.Errorf("Can not found prev Hash for non-finalize hash %v ", hash.String())
	}
	return prevH, nil
}
