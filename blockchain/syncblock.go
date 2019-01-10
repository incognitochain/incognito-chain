package blockchain

import (
	"errors"
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
type PeerShardChainState struct {
	State *ShardChainState
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

func (self *BlockChain) SyncShard(shardID byte) error {
	self.syncStatus.Lock()
	defer self.syncStatus.Unlock()

	if _, ok := self.syncStatus.Shard[shardID]; ok {
		return errors.New("Shard " + fmt.Sprintf("%d", shardID) + " synchronzation is already started")
	}
	var cSyncShardQuit chan struct{}
	cSyncShardQuit = make(chan struct{})
	self.syncStatus.Shard[shardID] = cSyncShardQuit

	var shardStateCh chan *PeerShardChainState
	var newShardBlkCh chan *ShardBlock
	shardStateCh = make(chan *PeerShardChainState)
	newShardBlkCh = make(chan *ShardBlock)

	self.ShardStateCh[shardID] = shardStateCh
	self.newShardBlkCh[shardID] = newShardBlkCh
	go func(shardID byte) {
		//used for fancy block retriever but too lazy to implement that now :p
		var peerChainState map[libp2p.ID]PeerShardChainState
		peerChainState = make(map[libp2p.ID]PeerShardChainState)
		_ = peerChainState
		getStateWaitTime := time.Duration(5)
		for {
			select {
			case <-self.cQuitSync:
				return
			case <-cSyncShardQuit:
				close(shardStateCh)
				close(newShardBlkCh)
				delete(self.newShardBlkCh, shardID)
				delete(self.ShardStateCh, shardID)
				delete(self.syncStatus.Shard, shardID)
				return
			case shardState := <-shardStateCh:
				if self.BestState.Shard[shardID].Height < shardState.State.Height {
					if self.knownChainState.Shards[shardID].Height < shardState.State.Height {
						self.knownChainState.Shards[shardID] = *shardState.State
						if getStateWaitTime > 5 {
							getStateWaitTime -= 5
						}
						self.config.Server.PushMessageGetBlockShard(shardID, self.BestState.Shard[shardID].Height+1, shardState.State.Height, shardState.Peer)
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
			case newBlk := <-newShardBlkCh:
				fmt.Println("Shard block received")
				if self.BestState.Shard[shardID].Height < newBlk.Header.Height {
					blkHash := newBlk.Header.Hash()
					err := cashec.ValidateDataB58(newBlk.Header.Producer, newBlk.ProducerSig, blkHash.GetBytes())
					if err != nil {
						Logger.log.Error(err)
						continue
					} else {
						if self.BestState.Shard[shardID].Height == newBlk.Header.Height-1 {
							err = self.InsertShardBlock(newBlk)
							if err != nil {
								Logger.log.Error(err)
								continue
							}
						} else {
							self.config.NodeShardPool.PushBlock(*newBlk)
						}
					}
				}
			default:
				time.Sleep(getStateWaitTime * time.Second)
				self.config.Server.PushMessageGetShardState(shardID)
			}
		}
	}(shardID)

	return nil
}

func (self *BlockChain) SyncBeacon() error {
	if self.syncStatus.Beacon {
		return errors.New("Beacon synchronzation is already started")
	}
	Logger.log.Info("Beacon synchronzation started")
	self.BeaconStateCh = make(chan *PeerBeaconChainState)
	self.newBeaconBlkCh = make(chan *BeaconBlock)
	self.syncStatus.Beacon = true

	go func() {
		//used for fancy block retriever but too lazy to implement that now :p
		var peerChainState map[libp2p.ID]PeerBeaconChainState
		peerChainState = make(map[libp2p.ID]PeerBeaconChainState)
		_ = peerChainState

		getStateWaitTime := time.Duration(5)
		for {
			select {
			case <-self.cQuitSync:
				return
			case beaconState := <-self.BeaconStateCh:
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
			case newBlk := <-self.newBeaconBlkCh:
				fmt.Println("Beacon block received")
				if self.BestState.Beacon.BeaconHeight < newBlk.Header.Height {
					blkHash := newBlk.Header.Hash()
					err := cashec.ValidateDataB58(newBlk.Header.Producer, newBlk.ProducerSig, blkHash.GetBytes())
					if err != nil {
						Logger.log.Error(err)
						continue
					} else {
						if self.BestState.Beacon.BeaconHeight == newBlk.Header.Height-1 {
							err = self.InsertBeaconBlock(newBlk)
							if err != nil {
								Logger.log.Error(err)
								continue
							}
						} else {
							self.config.NodeBeaconPool.PushBlock(*newBlk)
						}
					}
				}
			default:
				time.Sleep(getStateWaitTime * time.Second)
				self.config.Server.PushMessageGetBeaconState()
			}
		}
	}()
	return nil
}

func (self *BlockChain) RequestSyncShard(shardID byte) {

}

func (self *BlockChain) StopSyncShard(shardID byte) {
	self.syncStatus.Lock()
	defer self.syncStatus.Unlock()
	if _, ok := self.syncStatus.Shard[shardID]; ok {
		close(self.syncStatus.Shard[shardID])
		delete(self.syncStatus.Shard, shardID)
	}
}

func (self *BlockChain) GetCurrentSyncShards() []byte {

	return []byte{}
}

func (self *BlockChain) StopSync() error {
	return nil
}
