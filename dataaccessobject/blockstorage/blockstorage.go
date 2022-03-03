package blockstorage

import (
	"fmt"
	"log"
	"time"

	"github.com/incognitochain/incognito-chain/blockchain/types"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/flatfile"
	"github.com/incognitochain/incognito-chain/dataaccessobject/rawdbv2"
	"github.com/incognitochain/incognito-chain/incdb"
	"github.com/incognitochain/incognito-chain/proto"
	"github.com/pkg/errors"
)

type BlockService interface {
	GetBlockByHash(hash *common.Hash) ([]byte, error)
	StoreBlock(blkType proto.BlkType, blkData types.BlockInterface) error
	CheckBlockExist(hash *common.Hash) (bool, error)
}

type BlockManager struct {
	rDB incdb.Database
	fDB flatfile.FlatFile
}

func NewBlockService(rawDB incdb.Database, flatfileManager flatfile.FlatFile) (BlockService, error) {
	res := &BlockManager{
		rDB: rawDB,
		fDB: flatfileManager,
	}
	return res, nil
}

func (blkM *BlockManager) GetBlockFinalByHeight(height uint64, cID int) ([]byte, error) {
	key := rawdbv2.GetHeightToBlockIndexKey(height, cID)
	blkHashBytes, err := blkM.rDB.Get(key)
	if err != nil {
		return nil, errors.Wrapf(err, "can not get block %v of cID %v", height, cID)
	}
	blkHash := common.Hash{}
	copy(blkHash[:], blkHashBytes)
	return blkM.GetBlockByHash(&blkHash)
}

func (blkM *BlockManager) CheckBlockExist(hash *common.Hash) (bool, error) {
	keyIdx := rawdbv2.GetHashToBlockIndexKey(*hash)
	_, err := blkM.rDB.Get(keyIdx)
	if err != nil {
		return false, err
	}

	return true, nil
}

func (blkM *BlockManager) GetBlockByHash(hash *common.Hash) ([]byte, error) {
	keyIdx := rawdbv2.GetHashToBlockIndexKey(*hash)
	log.Println("*hash", hash.String(), keyIdx)
	blkIdBytes, err := blkM.rDB.Get(keyIdx)
	if err != nil {
		return nil, err
	}
	blkID, _ := common.BytesToUint64(blkIdBytes)
	log.Println("blkID", blkID, blkIdBytes)
	rawBlk, err := blkM.fDB.Read(int(blkID))
	return rawBlk, err

}

func (blkM *BlockManager) StoreBlock(
	blkType proto.BlkType,
	blkData types.BlockInterface,
) error {
	st := time.Now()
	blkBytes, err := blkData.ToBytes()
	if err != nil {
		return err
	}
	blkHash := blkData.Hash()
	st1 := time.Now()
	blkIndex, err := blkM.fDB.Append(blkBytes)
	if err != nil {
		return err
	}

	fmt.Printf("ffFile Append to DB cost %v\n", time.Since(st1))
	key := rawdbv2.GetHashToBlockIndexKey(*blkHash)
	if blkType == proto.BlkType_BlkShard {
		fmt.Printf("[testFF] store blk %v, key %v index %v\n", blkHash.String(), common.HashH(key).String(), blkIndex)
	}

	err = blkM.rDB.Put(key, common.Int64ToBytes(int64(blkIndex)))
	if err != nil {
		panic(err)
		return err
	}
	fmt.Printf("ffFile StoreBlock %v cost %v _ %v\n", blkData.Hash().String(), time.Since(st1), time.Since(st))
	return nil
}

//
//func (blkM *BlockManager) storeBlockHeightFinalized(
//	cID int,
//	blkHeight uint64,
//	blkHash common.Hash,
//) error {
//	key := rawdbv2.GetHeightToBlockIndexKey(blkHeight, cID)
//	err := blkM.rDB.Put(key, blkHash[:])
//	return err
//}

//func (blkM *BlockManager) MarkFinalized(
//	blkHeight uint64,
//	blkHash common.Hash,
//	cID byte,
//) error {
//	curFinalHeight := blkM.finalHeight[cID]
//	for height := blkHeight; height >= curFinalHeight; height-- {
//		err := blkM.storeBlockHeightFinalized(int(cID), height, blkHash)
//		if err != nil {
//			return err
//		}
//		pHash, ok := blkM.prevHashByHash[blkHash]
//		fmt.Printf("cID %v height %v testdelete %v prev %v \n", cID, height, blkHash.String(), pHash.String())
//		needToRemove := blkM.hashByHeight[cID][height]
//		for _, hash := range needToRemove {
//			delete(blkM.prevHashByHash, hash)
//			if hash != blkHash {
//				fmt.Printf("cID %v height %v testdelete delete %v - %v\n", cID, height, hash.String(), blkHash.String())
//				key := rawdbv2.GetHashToBlockIndexKey(hash)
//				err := blkM.rDB.Delete(key)
//				if err != nil {
//					panic(err)
//				}
//			}
//		}
//		if ok {
//			blkHash = pHash
//		}
//	}
//	blkM.finalHeight[cID] = blkHeight
//	return nil
//}
//
//func (blkM *BlockManager) GetPrevHashByHash(
//	hash *common.Hash,
//) (
//	common.Hash,
//	error,
//) {
//	blkM.locker.RLock()
//	prevH, existed := blkM.prevHashByHash[*hash]
//	blkM.locker.RUnlock()
//	if !existed {
//		return common.Hash{}, errors.Errorf("Can not found prev Hash for non-finalize hash %v ", hash.String())
//	}
//	return prevH, nil
//}
