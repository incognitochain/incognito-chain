package syncker

import (
	"github.com/incognitochain/incognito-chain/common"
	"time"
)

type BlkPool struct {
	action            chan func()
	blkPoolByHash     map[string]common.BlockPoolInterface // hash -> block
	blkPoolByPrevHash map[string][]string                  // prevhash -> []nexthash
}

func NewBlkPool(name string) *BlkPool {
	pool := new(BlkPool)
	pool.action = make(chan func())
	pool.blkPoolByHash = make(map[string]common.BlockPoolInterface)
	pool.blkPoolByPrevHash = make(map[string][]string)
	go pool.Start()
	return pool
}

func (pool *BlkPool) Start() {
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

func (pool *BlkPool) AddBlock(blk common.BlockPoolInterface) {
	pool.action <- func() {
		prevHash := blk.GetPrevHash()
		hash := blk.Hash().String()
		if _, ok := pool.blkPoolByHash[hash]; ok {
			return
		}
		pool.blkPoolByHash[hash] = blk
		if common.IndexOfStr(hash, pool.blkPoolByPrevHash[prevHash]) > -1 {
			return
		}
		pool.blkPoolByPrevHash[prevHash] = append(pool.blkPoolByPrevHash[prevHash], hash)
		//fmt.Println("Syncker: add block to pool", blk.GetHeight())
	}
}

func (pool *BlkPool) HashBlock(blk common.BlockPoolInterface) bool {
	res := make(chan bool)
	pool.action <- func() {
		hash := blk.Hash().String()
		_, ok := pool.blkPoolByHash[hash]
		res <- ok
	}
	return <-res
}

func (pool *BlkPool) RemoveBlock(hash string) {
	pool.action <- func() {
		if _, ok := pool.blkPoolByHash[hash]; ok {
			delete(pool.blkPoolByHash, hash)
		}
	}
}

func (pool *BlkPool) GetNextBlock(prevhash string, shouldGetLatest bool) common.BlockPoolInterface {
	//For multichain, we need to Get a Map
	res := make(chan common.BlockPoolInterface)
	pool.action <- func() {
		hashes := pool.blkPoolByPrevHash[prevhash][:]
		for _, h := range hashes {
			blk := pool.blkPoolByHash[h]
			if _, ok := pool.blkPoolByPrevHash[blk.Hash().String()]; shouldGetLatest || ok {
				res <- pool.blkPoolByHash[h]
				return
			}
		}
		res <- nil
	}
	return (<-res)
}

func (pool *BlkPool) GetFinalBlockFromBlockHash(currentHash string) []common.BlockPoolInterface {
	res := make(chan []common.BlockPoolInterface)
	pool.action <- func() {
		res <- GetFinalBlockFromBlockHash_v1(currentHash, pool.blkPoolByHash, pool.blkPoolByPrevHash)
	}
	return <-res
}

func (pool *BlkPool) GetLongestChain(currentHash string) []common.BlockPoolInterface {
	res := make(chan []common.BlockPoolInterface)
	pool.action <- func() {
		res <- GetLongestChain(currentHash, pool.blkPoolByHash, pool.blkPoolByPrevHash)
	}
	return <-res
}
