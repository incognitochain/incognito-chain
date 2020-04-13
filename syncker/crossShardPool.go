package syncker

import (
	"time"

	"github.com/incognitochain/incognito-chain/common"
)

type CrossShardBlkPool struct {
	action        chan func()
	blkPoolByHash map[string]common.CrossShardBlkPoolInterface // hash -> block
}

func NewCrossShardBlkPool(name string) *CrossShardBlkPool {
	pool := new(CrossShardBlkPool)
	pool.action = make(chan func())
	pool.blkPoolByHash = make(map[string]common.CrossShardBlkPoolInterface)
	go pool.Start()
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

func (pool *CrossShardBlkPool) GetPoolLength() int {
	res := make(chan int)
	pool.action <- func() {
		res <- len(pool.blkPoolByHash)
	}
	return <-res
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

func (pool *CrossShardBlkPool) RemoveBlock(hash string) {
	pool.action <- func() {
		if _, ok := pool.blkPoolByHash[hash]; ok {
			delete(pool.blkPoolByHash, hash)
		}
	}
}
