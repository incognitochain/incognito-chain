package blockchain

import (
	"time"

	"github.com/ninjadotorg/constant/cashec"
)

func (self *BlockChain) SyncShard(shardID byte, stopCh chan struct{}) {

}

func (self *BlockChain) SyncBeacon(stopCh chan struct{}) error {
	var pendingBlock map[uint64]*BeaconBlock
	pendingBlock = make(map[uint64]*BeaconBlock)
	go func() {
		for {
			select {
			case <-stopCh:
				return
			default:
				time.Sleep(5 * time.Second)
				//TODO send get chain state of beacon
			}
		}
	}()

	for {
		select {
		case <-stopCh:
			return nil
		case newBlk := <-self.newBeaconBlkCh:
			if self.BestState.Beacon.BeaconHeight < newBlk.Header.Height {
				err := cashec.ValidateDataB58(newBlk.Header.Producer, newBlk.ProducerSig, []byte(newBlk.Header.Hash().String()))
				if err != nil {
					return err
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
