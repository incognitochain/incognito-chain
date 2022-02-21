package blockstorage

import (
	"sync"

	"github.com/incognitochain/incognito-chain/blockchain/types"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/flatfile"
	"github.com/incognitochain/incognito-chain/dataaccessobject/rawdbv2"
	"github.com/incognitochain/incognito-chain/incdb"
	"github.com/incognitochain/incognito-chain/proto"
	"github.com/patrickmn/go-cache"
	"github.com/pkg/errors"
)

type BlockService interface {
	GetBlockByHeight(
		blkType proto.BlkType,
		height uint64,
		fromcID byte,
		tocID byte,
	) (
		interface{},
		error,
	)
	GetBlockByHash(
		// blkType proto.BlkType,
		hash *common.Hash,
		// fromcID byte,
		// tocID byte,
	) (
		[]byte,
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
	)
}

type BlockManager struct {
	rDB            incdb.Database
	fDB            flatfile.FlatFile
	cacher         *cache.Cache
	locker         *sync.RWMutex
	hashByHeight   map[byte]map[uint64][]common.Hash
	prevHashByHash map[common.Hash]common.Hash
	finalHeight    map[byte]uint64
}

func NewBlockService(
	rawDB incdb.Database,
	flatfileManager flatfile.FlatFile,
) (
	BlockService,
	error,
) {
	return nil, errors.New("Please implement me")
}

func (blkM *BlockManager) GetBlockByHeight(
	blkType proto.BlkType,
	height uint64,
	fromcID byte,
	tocID byte,
) (
	interface{},
	error,
) {
	return nil, errors.New("Please implement me")
}
func (blkM *BlockManager) GetBlockByHash(
	// blkType proto.BlkType,
	hash *common.Hash,
	// fromcID byte,
	// tocID byte,
) (
	[]byte,
	error,
) {
	key := rawdbv2.GetShardHashToBlockKey(*hash)
	rawBlk, err := blkM.rDB.Get(key)
	if err != nil {
		keyIdx := rawdbv2.GetHashToBlockIndexKey(*hash)
		blkIdBytes, err := blkM.rDB.Get(keyIdx)
		if err != nil {
			return nil, err
		}
		blkID := common.BytesToInt(blkIdBytes)
		rawBlk, err := blkM.fDB.Read(blkID)
		return rawBlk, err
	}
	return rawBlk, err
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
	blkCID := blkData.GetShardID()
	if err != nil {
		return err
	}
	key := rawdbv2.GetShardHashToBlockKey(*blkHash)
	err = blkM.rDB.Put(key, blkBytes)
	if err != nil {
		return err
	}
	blkM.locker.Lock()
	blkM.hashByHeight[byte(blkCID)][blkHeight] = append(blkM.hashByHeight[byte(blkCID)][blkHeight], *blkHash)
	blkM.prevHashByHash[*blkHash] = blkData.GetPrevHash()
	blkM.locker.Unlock()
	return nil
}

func (blkM *BlockManager) storeBlockFinalized(
	// blkType proto.BlkType,
	blkBytes []byte,
	blkHash *common.Hash,
) error {
	blkIndex, err := blkM.fDB.Append(blkBytes)
	if err != nil {
		return err
	}
	key := rawdbv2.GetHashToBlockIndexKey(*blkHash)
	err = blkM.rDB.Put(key, common.IntToBytes(blkIndex))
	return err
}
func (blkM *BlockManager) MarkFinalized(
	height uint64,
	hash common.Hash,
	cID byte,
) error {
	curFinalHeight := blkM.finalHeight[cID]
	for h := height; h >= curFinalHeight; h-- {
		key := rawdbv2.GetShardHashToBlockKey(hash)
		blkBytes, err := blkM.rDB.Get(key)
		if err != nil {
			return err
		}
		err = blkM.storeBlockFinalized(blkBytes, &hash)
		if err != nil {
			return err
		}
		pHash, ok := blkM.prevHashByHash[hash]
		if ok {
			hash = pHash
		}
		delete(blkM.prevHashByHash, hash)
		needToRemove := blkM.hashByHeight[cID][height]
		for _, h := range needToRemove {
			key := rawdbv2.GetShardHashToBlockKey(h)
			err := blkM.rDB.Delete(key)
			if err != nil {
				panic(err)
			}
		}
	}
	return nil
}
