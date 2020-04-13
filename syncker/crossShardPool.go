package syncker

import (
	"time"

	"github.com/incognitochain/incognito-chain/common"
)

type CrossShardBlkPool struct {
	action        chan func()
	blkPoolByHash map[string]common.CrossShardBlkPoolInterface // hash -> block
}

func NewCrossShardBlkPool(name string, IsOutdatedBlk func(interface{}) bool) *CrossShardBlkPool {
	pool := new(CrossShardBlkPool)
	pool.action = make(chan func())
	pool.blkPoolByHash = make(map[string]common.CrossShardBlkPoolInterface)
	go pool.Start()

	//remove outdated block in pool, only trigger if pool has more than 1000 blocks
	go func() {
		ticker := time.NewTicker(15 * time.Second)
		for {
			<-ticker.C
			if pool.GetPoolSize() > 1000 {
				blkList := pool.GetBlockList()
				for _, blk := range blkList {
					if IsOutdatedBlk(blk) {
						pool.RemoveBlock(blk.Hash())
					}
				}
			}
		}
	}()

	return pool
}

func (pool *CrossShardBlkPool) Start() {
	ticker := time.NewTicker(time.Millisecond * 500)
	for {
		select {
		case f := <-pool.action:
			f()
		case <-ticker.C:
			//TODO: loop through all prevhash, delete if all nextHash is deleted
		}
	}
}

func (pool *CrossShardBlkPool) GetPoolSize() int {
	return len(pool.blkPoolByHash)
}

func (pool *CrossShardBlkPool) GetBlockList() []common.BlockPoolInterface {
	res := make(chan []common.BlockPoolInterface)
	pool.action <- func() {
		blkList := make([]common.BlockPoolInterface, len(pool.blkPoolByHash))
		for _, blk := range pool.blkPoolByHash {
			blkList = append(blkList, blk.(common.BlockPoolInterface))
		}
		res <- blkList
	}
	return <-res
}

func (pool *CrossShardBlkPool) AddBlock(blk common.CrossShardBlkPoolInterface) {
	pool.action <- func() {
		hash := blk.Hash()
		if _, ok := pool.blkPoolByHash[hash.String()]; ok {
			return
		}
		pool.blkPoolByHash[hash.String()] = blk
	}
}

func (pool *CrossShardBlkPool) HasBlock(hash common.Hash) bool {
	res := make(chan bool)
	pool.action <- func() {
		_, ok := pool.blkPoolByHash[hash.String()]
		res <- ok
	}
	return <-res
}

func (pool *CrossShardBlkPool) GetBlock(hash common.Hash) common.CrossShardBlkPoolInterface {
	res := make(chan common.CrossShardBlkPoolInterface)
	pool.action <- func() {
		b, _ := pool.blkPoolByHash[hash.String()]
		res <- b
	}
	return <-res
}

func (pool *CrossShardBlkPool) RemoveBlock(hash *common.Hash) {
	pool.action <- func() {
		if _, ok := pool.blkPoolByHash[hash.String()]; ok {
			delete(pool.blkPoolByHash, hash.String())
		}
	}
}
