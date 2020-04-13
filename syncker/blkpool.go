package syncker

import (
	"time"

	"github.com/incognitochain/incognito-chain/common"
)

type BlkPool struct {
	action            chan func()
	blkPoolByHash     map[string]common.BlockPoolInterface // hash -> block
	blkPoolByPrevHash map[string][]string                  // prevhash -> []nexthash
}

func NewBlkPool(name string, IsOutdatedBlk func(interface{}) bool) *BlkPool {

	pool := new(BlkPool)
	pool.action = make(chan func())
	pool.blkPoolByHash = make(map[string]common.BlockPoolInterface)
	pool.blkPoolByPrevHash = make(map[string][]string)
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

func (pool *BlkPool) GetPoolSize() int {
	return len(pool.blkPoolByHash)
}

func (pool *BlkPool) GetBlockList() []common.BlockPoolInterface {
	res := make(chan []common.BlockPoolInterface)
	pool.action <- func() {
		blkList := make([]common.BlockPoolInterface, len(pool.blkPoolByHash))
		for _, blk := range pool.blkPoolByHash {
			blkList = append(blkList, blk)
		}
		res <- blkList
	}
	return <-res
}

func (pool *BlkPool) AddBlock(blk common.BlockPoolInterface) {
	pool.action <- func() {
		prevHash := blk.GetPrevHash().String()
		hash := blk.Hash().String()
		//if exists, return
		if _, ok := pool.blkPoolByHash[hash]; ok {
			return
		}
		pool.blkPoolByHash[hash] = blk
		//insert into prehash datastructure
		if common.IndexOfStr(hash, pool.blkPoolByPrevHash[prevHash]) > -1 {
			return
		}
		pool.blkPoolByPrevHash[prevHash] = append(pool.blkPoolByPrevHash[prevHash], hash)
		//fmt.Println("Syncker: add block to pool", blk.GetHeight())
	}
}

func (pool *BlkPool) GetBlock(hash common.Hash) common.BlockPoolInterface {
	res := make(chan common.BlockPoolInterface)
	pool.action <- func() {
		blk, _ := pool.blkPoolByHash[hash.String()]
		res <- blk
	}
	return <-res
}

func (pool *BlkPool) HasHash(hash common.Hash) bool {
	res := make(chan bool)
	pool.action <- func() {
		_, ok := pool.blkPoolByHash[hash.String()]
		res <- ok
	}
	return <-res
}

func (pool *BlkPool) RemoveBlock(hash *common.Hash) {
	pool.action <- func() {
		if _, ok := pool.blkPoolByHash[hash.String()]; ok {
			delete(pool.blkPoolByHash, hash.String())
		}
	}
}

func (pool *BlkPool) GetPoolInfo() []common.BlockPoolInterface {
	res := make(chan []common.BlockPoolInterface)
	pool.action <- func() {
		res <- GetPoolInfo(pool.blkPoolByHash)
	}
	return <-res
}

// END OF COMMON FUNCTION =======================================================================

// START OF SPECIAL CASE FUNCTION =======================================================================

//When get s2b block for producer
//Get Block from current hash to final block
func (pool *BlkPool) GetFinalBlockFromBlockHash(currentHash string) []common.BlockPoolInterface {
	res := make(chan []common.BlockPoolInterface)
	pool.action <- func() {
		res <- GetFinalBlockFromBlockHash_v1(currentHash, pool.blkPoolByHash, pool.blkPoolByPrevHash)
	}
	return <-res
}

//When get last block for s2b synchronization
//Get longest branch in pool
func (pool *BlkPool) GetLongestChain(currentHash string) []common.BlockPoolInterface {
	res := make(chan []common.BlockPoolInterface)
	pool.action <- func() {
		res <- GetLongestChain(currentHash, pool.blkPoolByHash, pool.blkPoolByPrevHash)
	}
	return <-res
}
