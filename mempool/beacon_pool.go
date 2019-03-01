package mempool

import (
	"errors"
	"github.com/ninjadotorg/constant/blockchain"
	"github.com/ninjadotorg/constant/common"
	"sort"
)

const (
	MAX_VALID_BEACON_BLK_IN_POOL   = 100
	MAX_INVALID_BEACON_BLK_IN_POOL = 200
)

type BeaconPool struct {
	pool              []*blockchain.BeaconBlock // block
	latestValidHeight uint64
}

var beaconPool *BeaconPool = nil

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
	self.latestValidHeight = lastestBeaconHeight

	//Remove pool base on new shardstate
	self.RemoveBlock(lastestBeaconHeight)
	self.UpdateLatestBeaconState()
}

func (self *BeaconPool) GetBeaconState() uint64 {
	return self.latestValidHeight
}

func (self *BeaconPool) AddBeaconBlock(blk blockchain.BeaconBlock) error {
	//TODO: validate aggregated signature

	blkHeight := blk.Header.Height

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

	//Check if satisfy pool capacity (for valid and invalid)
	if len(self.pool) != 0 {
		numValidPedingBlk := int(self.latestValidHeight - self.pool[0].Header.Height)
		numInValidPedingBlk := len(self.pool) - numValidPedingBlk
		if numValidPedingBlk > MAX_VALID_BEACON_BLK_IN_POOL {
			return errors.New("exceed max valid pending block")
		}

		lastBlkInPool := self.pool[len(self.pool)-1]
		if numInValidPedingBlk > MAX_INVALID_BEACON_BLK_IN_POOL {
			//If invalid block is better than current invalid block
			if lastBlkInPool.Header.Height > blkHeight {
				//remove latest block and add better invalid to pool
				self.pool = self.pool[:len(self.pool)-1]
			} else {
				return errors.New("exceed invalid pending block")
			}
		}
	}

	// add to pool
	self.pool = append(self.pool, &blk)

	//sort pool
	sort.Slice(self.pool, func(i, j int) bool {
		return self.pool[i].Header.Height < self.pool[j].Header.Height
	})

	//update last valid pending ShardState
	self.UpdateLatestBeaconState()
	if self.pool[0].Header.Height > self.latestValidHeight {
		offset := self.pool[0].Header.Height - self.latestValidHeight
		if offset > MAX_VALID_BEACON_BLK_IN_POOL {
			offset = MAX_VALID_BEACON_BLK_IN_POOL
		}
		return nil
	}
	return nil
}
func (self *BeaconPool) UpdateLatestBeaconState() {
	lastHeight := self.latestValidHeight
	for _, blk := range self.pool {
		if blk.Header.Height > lastHeight && lastHeight+1 != blk.Header.Height {
			break
		}
		lastHeight = blk.Header.Height
	}
	self.latestValidHeight = lastHeight
}

//@Notice: Remove should set latest valid height
//Because normal beacon node may not have these block to remove
func (self *BeaconPool) RemoveBlock(lastBlockHeight uint64) {
	for index, block := range self.pool {
		if block.Header.Height <= lastBlockHeight {
			if index == len(self.pool)-1 {
				self.pool = self.pool[index+1:]
			}
			continue
		} else {
			self.pool = self.pool[index:]
			break
		}
	}

}

func (self *BeaconPool) GetValidBlock() []*blockchain.BeaconBlock {
	finalBlocks := []*blockchain.BeaconBlock{}
	for _, blk := range self.pool {
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

func (self *BeaconPool) GetAllBlockHeight() []uint64 {
	finalBlocks := []uint64{}
	for _, blk := range self.pool {
		finalBlocks = append(finalBlocks, blk.Header.Height)
	}
	return finalBlocks
}
