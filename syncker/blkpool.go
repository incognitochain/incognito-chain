package syncker

import (
	"github.com/incognitochain/incognito-chain/common"
	"sync"
	"time"
)

type BlkPool struct {
	action          chan func()
	BlkPoolByHash   sync.Map // hash -> block
	BlkPoolByHeight sync.Map // height -> []hash
}

func NewBlkPool(name string) *BlkPool {
	pool := new(BlkPool)
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
		}
	}
}

func (pool *BlkPool) AddBlock(blk common.BlockInterface) {
	pool.action <- func() {
		height := blk.GetHeight()
		hash := blk.Hash().String()
		pool.BlkPoolByHash.Store(hash, blk)
		if loadValue, isLoad := pool.BlkPoolByHeight.LoadOrStore(height, []string{hash}); isLoad {
			if common.IndexOfStr(hash, loadValue.([]string)) > -1 {
				return
			}
			pool.BlkPoolByHeight.Store(height, append(loadValue.([]string), hash))
		}

	}
}
