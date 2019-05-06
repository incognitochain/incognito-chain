package mempool

import (
	"errors"
	"fmt"
	"github.com/constant-money/constant-chain/blockchain"
	"github.com/constant-money/constant-chain/common"
	"sort"
	"sync"
	"time"
)

const (
	MAX_VALID_BEACON_BLK_IN_POOL   = 10000
	MAX_INVALID_BEACON_BLK_IN_POOL = 10000
)

type BeaconPool struct {
	pool              []*blockchain.BeaconBlock // block
	latestValidHeight uint64
	poolMu            sync.RWMutex
}

var beaconPool *BeaconPool = nil

func init() {
	go func() {
		ticker := time.Tick(5 * time.Second)
		for _ = range ticker {
			GetBeaconPool().RemoveBlock(blockchain.GetBestStateBeacon().BeaconHeight)
		}
	}()
}

func InitBeaconPool() {
	//do nothing
	GetBeaconPool().SetBeaconState(blockchain.GetBestStateBeacon().BeaconHeight)
}

// get singleton instance of ShardToBeacon pool
func GetBeaconPool() *BeaconPool {
	if beaconPool == nil {
		beaconPool = new(BeaconPool)
		beaconPool.pool = []*blockchain.BeaconBlock{}
		beaconPool.latestValidHeight = 1
	}
	return beaconPool
}

func (self *BeaconPool) SetBeaconState(lastestBeaconHeight uint64) {
	if self.latestValidHeight < lastestBeaconHeight {
		self.latestValidHeight = lastestBeaconHeight
	}
}

func (self *BeaconPool) GetBeaconState() uint64 {
	return self.latestValidHeight
}

func (self *BeaconPool) AddBeaconBlock(blk *blockchain.BeaconBlock) error {

	blkHeight := blk.Header.Height

	self.poolMu.Lock()
	defer self.poolMu.Unlock()

	//If receive old block, it will ignore
	if blkHeight <= self.latestValidHeight {
		return errors.New("receive old block")
	}

	//If block already in pool, it will ignore
	for _, blkItem := range self.pool {
		if blkItem.Header.Height == blkHeight {
			return errors.New("receive duplicate block")
		}
	}
	//fmt.Println("BP: will add to beacon pool", time.Since(time1).Seconds())

	//Check if satisfy pool capacity (for valid and invalid)

	if len(self.pool) != 0 {
		numValidPedingBlk := int(self.latestValidHeight - self.pool[0].Header.Height)
		if numValidPedingBlk < 0 {
			numValidPedingBlk = 0
		}
		numInValidPedingBlk := len(self.pool) - numValidPedingBlk
		if numValidPedingBlk > MAX_VALID_BEACON_BLK_IN_POOL {
			fmt.Println("cannot add to beacon pool (exceed valid pending block)", blk.Header.Height, self.latestValidHeight, blockchain.GetBestStateBeacon().BeaconHeight)
			return errors.New("exceed max valid pending block")
		}
		lastBlkInPool := self.pool[len(self.pool)-1]
		if numInValidPedingBlk > MAX_INVALID_BEACON_BLK_IN_POOL {
			//If invalid block is better than current invalid block
			if lastBlkInPool.Header.Height > blkHeight {
				//remove latest block and add better invalid to pool
				fmt.Println("swap out beacon pool ", self.pool[len(self.pool)-1].Header.Height, self.pool[0].Header.Height, self.latestValidHeight, blockchain.GetBestStateBeacon().BeaconHeight)
				self.pool = self.pool[:len(self.pool)-1]
			} else {
				fmt.Println("cannot add to beacon pool (exceed pending block)", blk.Header.Height, self.latestValidHeight, blockchain.GetBestStateBeacon().BeaconHeight)
				return errors.New("exceed invalid pending block")
			}
		}
	}

	//fmt.Println("BP: add to beacon pool ", blk.Header.Height, self.latestValidHeight, blockchain.GetBestStateBeacon().BeaconHeight, time.Since(time1).Seconds())
	// add to pool
	self.pool = append(self.pool, blk)

	//sort pool
	sort.Slice(self.pool, func(i, j int) bool {
		return self.pool[i].Header.Height < self.pool[j].Header.Height
	})

	//update last valid pending ShardState
	self.updateLatestBeaconState()
	//fmt.Println("BP: finish update", time.Since(time1).Seconds())
	return nil
}

func (self *BeaconPool) updateLatestBeaconState() {
	lastHeight := self.latestValidHeight
	for i, blk := range self.pool {
		if blk.Header.Height <= lastHeight {
			continue
		}

		if lastHeight+1 != blk.Header.Height {
			break
		}

		if i == len(self.pool)-1 {
			break
		}

		if self.pool[i+1].Header.PrevBlockHash != *blk.Hash() {
			break
		}

		lastHeight = blk.Header.Height
	}
	self.latestValidHeight = lastHeight
}

func (self *BeaconPool) RemoveBlock(lastBlockHeight uint64) {
	self.poolMu.Lock()
	defer self.poolMu.Unlock()
	self.removeBlock(lastBlockHeight)
}

//@Notice: Remove should set latest valid height
//Because normal beacon node may not have these block to remove
func (self *BeaconPool) removeBlock(lastBlockHeight uint64) {
	for index, block := range self.pool {
		if block.Header.Height <= lastBlockHeight {
			if index == len(self.pool)-1 {
				self.pool = []*blockchain.BeaconBlock{}
			}
			continue
		} else {
			self.pool = self.pool[index:]
			break
		}
	}
}

func (self *BeaconPool) GetValidBlock() []*blockchain.BeaconBlock {
	self.poolMu.RLock()
	defer self.poolMu.RUnlock()

	finalBlocks := []*blockchain.BeaconBlock{}
	for _, blk := range self.pool {
		if blk.Header.Height <= blockchain.GetBestStateBeacon().BeaconHeight {
			continue
		}
		if blk.Header.Height > self.latestValidHeight {
			break
		}

		finalBlocks = append(finalBlocks, blk)
	}
	return finalBlocks
}

func (self *BeaconPool) GetValidBlockHash() []common.Hash {
	finalBlocks := []common.Hash{}
	blks := self.GetValidBlock()
	for _, blk := range blks {
		finalBlocks = append(finalBlocks, *blk.Hash())
	}
	return finalBlocks
}

func (self *BeaconPool) GetValidBlockHeight() []uint64 {
	finalBlocks := []uint64{}
	blks := self.GetValidBlock()
	for _, blk := range blks {
		finalBlocks = append(finalBlocks, blk.Header.Height)
	}
	return finalBlocks
}

func (self *BeaconPool) GetLatestValidBlockHeight() uint64 {
	finalBlocks := uint64(0)
	blks := self.GetValidBlock()
	for _, blk := range blks {
		finalBlocks = blk.Header.Height
	}
	return finalBlocks
}

func (self *BeaconPool) GetPoolLen() uint64 {
	self.poolMu.RLock()
	defer self.poolMu.RUnlock()
	return uint64(len(self.pool))
}

func (self *BeaconPool) GetAllBlockHeight() []uint64 {
	self.poolMu.RLock()
	defer self.poolMu.RUnlock()

	finalBlocks := []uint64{}
	for _, blk := range self.pool {
		finalBlocks = append(finalBlocks, blk.Header.Height)
	}
	return finalBlocks
}

func (self *BeaconPool) GetBlockByHeight(height uint64) *blockchain.BeaconBlock {
	self.poolMu.RLock()
	defer self.poolMu.RUnlock()
	for _, blk := range self.pool {
		if blk.Header.Height == height {
			return blk
		}
	}
	return nil
}
