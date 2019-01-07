package blockchain

import (
	"time"

	"github.com/ninjadotorg/constant/cashec"
	"github.com/ninjadotorg/constant/common"
)

type ShardChainState struct {
	Height    uint64
	BlockHash common.Hash
}

type BeaconChainState struct {
	Height    uint64
	BlockHash common.Hash
}

func (self *BlockChain) SyncShard(shardID byte) {

}

func (self *BlockChain) SyncBeacon() {
	if self.syncStatus.Beacon {
		Logger.log.Error("Beacon synchronzation is already started")
		return
	}
	Logger.log.Info("Beacon synchronzation started")
	self.syncStatus.Beacon = true
	var pendingBlock map[uint64]*BeaconBlock
	pendingBlock = make(map[uint64]*BeaconBlock)
	go func() {
		for {
			select {
			case <-self.cQuitSync:
				return
			default:
				time.Sleep(5 * time.Second)
				self.config.Server.PushMessageGetBeaconState()
			}
		}
	}()

	for {
		select {
		case <-self.cQuitSync:
			return
		case newBlk := <-self.newBeaconBlkCh:
			if self.BestState.Beacon.BeaconHeight < newBlk.Header.Height {
				err := cashec.ValidateDataB58(newBlk.Header.Producer, newBlk.ProducerSig, []byte(newBlk.Header.Hash().String()))
				if err != nil {
					continue
				} else {
					if self.BestState.Beacon.BeaconHeight == newBlk.Header.Height-1 {
						err = self.InsertBeaconBlock(newBlk)
						if err != nil {
							Logger.log.Error(err)
							continue
						}
					} else {
						if _, ok := pendingBlock[newBlk.Header.Height]; !ok {
							pendingBlock[newBlk.Header.Height] = newBlk
						}
					}
				}

			}
		}
	}
}

func (self *BlockChain) RequestSyncShard(shardID byte) {

}

func (self *BlockChain) StopSyncShard(shardID byte) {

}

func (self *BlockChain) GetCurrentSyncShards() []byte {

	return []byte{}
}

func (self *BlockChain) StopSync() error {
	return nil
}
