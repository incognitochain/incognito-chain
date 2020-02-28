package syncker

import (
	"github.com/incognitochain/incognito-chain/common"
	"time"
)

type BlkPool struct {
	action            chan func()
	BlkPoolByHash     map[string]common.BlockPoolInterface // hash -> block
	BlkPoolByPrevHash map[string][]string                  // prevhash -> []nexthash
}

func NewBlkPool(name string) *BlkPool {
	pool := new(BlkPool)
	pool.action = make(chan func())
	pool.BlkPoolByHash = make(map[string]common.BlockPoolInterface)
	pool.BlkPoolByPrevHash = make(map[string][]string)
	go pool.Start()
	return pool
}

func (pool *BlkPool) Start() {
	ticker := time.NewTicker(time.Millisecond * 500)
	for {
		select {
		case f := <-pool.action:
			f()
		default:
			<-ticker.C
			//TODO: loop through all prevhash, delete if all nextHash is deleted
		}
	}
}

func (pool *BlkPool) AddBlock(blk common.BlockPoolInterface) {
	pool.action <- func() {
		prevHash := blk.GetPrevHash()
		hash := blk.GetHash()
		if _, ok := pool.BlkPoolByHash[hash]; ok {
			return
		}
		pool.BlkPoolByHash[hash] = blk
		if common.IndexOfStr(hash, pool.BlkPoolByPrevHash[prevHash]) > -1 {
			return
		}
		pool.BlkPoolByPrevHash[prevHash] = append(pool.BlkPoolByPrevHash[prevHash], hash)
		//fmt.Println("Syncker: add block to pool", blk.GetHeight())
	}
}

func (pool *BlkPool) RemoveBlock(hash string) {
	pool.action <- func() {
		if _, ok := pool.BlkPoolByHash[hash]; ok {
			delete(pool.BlkPoolByHash, hash)
		}
	}
}

func (pool *BlkPool) GetNextBlock(prevhash string, shouldGetLatest bool) common.BlockPoolInterface {
	//For multichain, we need to Get a Map
	res := make(chan common.BlockPoolInterface)
	pool.action <- func() {
		hashes := pool.BlkPoolByPrevHash[prevhash][:]
		for _, h := range hashes {
			blk := pool.BlkPoolByHash[h]
			if _, ok := pool.BlkPoolByPrevHash[blk.GetHash()]; shouldGetLatest || ok {
				res <- pool.BlkPoolByHash[h]
				return
			}
		}
		res <- nil
	}
	return (<-res)
}
