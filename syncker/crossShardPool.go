package syncker

import (
	"github.com/incognitochain/incognito-chain/common"
	"time"
)

type CrossShardBlkPool struct {
	action        chan func()
	BlkPoolByHash map[string]common.CrossShardBlkPoolInterface // hash -> block
}

func NewCrossShardBlkPool(name string) *CrossShardBlkPool {
	pool := new(CrossShardBlkPool)
	pool.action = make(chan func())
	pool.BlkPoolByHash = make(map[string]common.CrossShardBlkPoolInterface)
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

func (pool *CrossShardBlkPool) AddBlock(blk common.CrossShardBlkPoolInterface) {
	pool.action <- func() {
		hash := blk.Hash()
		if _, ok := pool.BlkPoolByHash[hash.String()]; ok {
			return
		}
		pool.BlkPoolByHash[hash.String()] = blk

	}
}

func (pool *CrossShardBlkPool) HasBlock(hash common.Hash) bool {
	res := make(chan bool)
	pool.action <- func() {
		_, ok := pool.BlkPoolByHash[hash.String()]
		res <- ok
	}
	return <-res
}

func (pool *CrossShardBlkPool) GetBlock(hash common.Hash) common.CrossShardBlkPoolInterface {
	res := make(chan common.CrossShardBlkPoolInterface)
	pool.action <- func() {
		b, _ := pool.BlkPoolByHash[hash.String()]
		res <- b
	}
	return <-res
}

func (pool *CrossShardBlkPool) RemoveBlock(hash string) {
	pool.action <- func() {
		if _, ok := pool.BlkPoolByHash[hash]; ok {
			delete(pool.BlkPoolByHash, hash)
		}
	}
}
