package blockchain

import (
	"fmt"
	"time"

	libp2p "github.com/libp2p/go-libp2p-peer"
	"github.com/ninjadotorg/constant/cashec"
	"github.com/ninjadotorg/constant/common"
)

type PeerBeaconChainState struct {
	State *BeaconChainState
	Peer  libp2p.ID
}
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
	self.BeaconStateCh = make(chan *PeerBeaconChainState)
	self.syncStatus.Beacon = true
	self.newBeaconBlkCh = make(chan *BeaconBlock)
	var pendingBlock map[uint64]*BeaconBlock
	pendingBlock = make(map[uint64]*BeaconBlock)
	go func() {
		getStateWaitTime := time.Duration(5)
		for {
			select {
			case beaconState := <-self.BeaconStateCh:
				fmt.Println()
				fmt.Println(self.BestState.Beacon.BeaconHeight, beaconState.State)
				fmt.Println()
				if self.BestState.Beacon.BeaconHeight < beaconState.State.Height {
					if self.knownChainState.Beacon.Height < beaconState.State.Height {
						self.knownChainState.Beacon = *beaconState.State
						if getStateWaitTime > 5 {
							getStateWaitTime -= 5
						}
						self.config.Server.PushMessageGetBlockBeacon(self.BestState.Beacon.BeaconHeight+1, beaconState.State.Height, beaconState.Peer)
					} else {
						if getStateWaitTime < 10 {
							getStateWaitTime += 5
						}
					}
				} else {
					if getStateWaitTime < 10 {
						getStateWaitTime += 5
					}
				}
			case <-self.cQuitSync:
				return
			default:
				time.Sleep(getStateWaitTime * time.Second)
				self.config.Server.PushMessageGetBeaconState()
			}
		}
	}()

	for {
		select {
		case <-self.cQuitSync:
			return
		case newBlk := <-self.newBeaconBlkCh:
			fmt.Println("Beacon block received")
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
