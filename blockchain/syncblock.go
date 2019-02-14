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
	Height               uint64
	ShardID              byte
	CrossShardsPoolState map[byte]map[uint64][]common.Hash // how many distinct block of each height of each shard
	BlockHash            common.Hash
}

type BeaconChainState struct {
	Height          uint64
	ShardsPoolState map[byte][]common.Hash // how many distinct block of each height of each shard
	BlockHash       common.Hash
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
	self.knownChainState.Shards[shardID] = ShardChainState{
		Height:  self.BestState.Shard[shardID].ShardHeight,
		ShardID: shardID,
	}
	var shardStateCh chan *PeerShardChainState
	var newShardBlkCh chan *ShardBlock
	shardStateCh = make(chan *PeerShardChainState)
	newShardBlkCh = make(chan *ShardBlock)

	self.ShardStateCh[shardID] = shardStateCh
	self.newShardBlkCh[shardID] = &newShardBlkCh
	go func(shardID byte) {
		//used for fancy block retriever but not implement that now :p
		var peerChainState map[libp2p.ID]PeerShardChainState
		peerChainState = make(map[libp2p.ID]PeerShardChainState)
		_ = peerChainState
		getStateWaitTime := time.Duration(defaultGetStateWaitTime)
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
				if self.BestState.Shard[shardID].ShardHeight < shardState.State.Height {
					if self.knownChainState.Shards[shardID].Height < shardState.State.Height {
						self.knownChainState.Shards[shardID] = *shardState.State
						if getStateWaitTime == defaultGetStateWaitTime*2 {
							getStateWaitTime -= defaultGetStateWaitTime
						}
						go self.config.Server.PushMessageGetBlockShard(shardID, self.BestState.Shard[shardID].ShardHeight+1, shardState.State.Height, shardState.Peer)
					} else {
						if getStateWaitTime == defaultGetStateWaitTime {
							getStateWaitTime += defaultGetStateWaitTime
						}
					}
				} else {
					if getStateWaitTime == defaultGetStateWaitTime {
						getStateWaitTime += defaultGetStateWaitTime
					}
				}

				// check shardToBeacon pool state
				if len(shardState.State.CrossShardsPoolState) > 0 {
					// myPoolPending := self.config.CrossShardPool.GetPendingBlockHashes()
					// for shardID, peerPoolBlks := range shardState.State.CrossShardsPoolState {
					// 	myShardPoolBlks, ok := myPoolPending[shardID]
					// 	if ok {
					// 		blksNeedToSync := GetDiffHashesOf(peerPoolBlks, myShardPoolBlks)
					// 		for _, blkHash := range blksNeedToSync {
					// 			go self.config.Server.PushMessageGetShardToBeacon(shardID, blkHash)
					// 		}
					// 	} else {
					// 		// sync all blks of this shard
					// 		for _, blkHash := range peerPoolBlks {
					// 			go self.config.Server.PushMessageGetShardToBeacon(shardID, blkHash)
					// 		}
					// 	}
					// }
				}
			case newBlk := <-newShardBlkCh:
				fmt.Println("Shard block received")
				if self.BestState.Shard[shardID].ShardHeight < newBlk.Header.Height {
					blkHash := newBlk.Header.Hash()
					err := cashec.ValidateDataB58(newBlk.Header.Producer, newBlk.ProducerSig, blkHash.GetBytes())
					if err != nil {
						Logger.log.Error(err)
						continue
					} else {
						if self.BestState.Shard[shardID].ShardHeight == newBlk.Header.Height-1 {
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
				if self.knownChainState.Shards[shardID].Height > self.BestState.Shard[shardID].ShardHeight {
					needToSync := self.knownChainState.Beacon.Height - self.BestState.Beacon.BeaconHeight
					for offset := uint64(0); offset <= needToSync; offset++ {
						blks, err := self.config.NodeShardPool.GetBlocks(shardID, self.BestState.Shard[shardID].ShardHeight+1)
						if err != nil {
							Logger.log.Error(err)
							continue
						}
						for _, newBlk := range blks {
							err = self.InsertShardBlock(&newBlk)
							if err != nil {
								Logger.log.Error(err)
								continue
							}
						}
					}
				}
			}
		}
	}(shardID)

	return nil
}

func (self *BlockChain) SyncShardToBeacon() {
	for range self.cQuitSync {
		return
	}
}

func (self *BlockChain) SyncBeacon() error {
	if self.syncStatus.Beacon {
		return errors.New("beacon synchronzation is already started")
	}
	Logger.log.Info("Beacon synchronzation started")
	self.BeaconStateCh = make(chan *PeerBeaconChainState)
	self.newBeaconBlkCh = make(chan *BeaconBlock)
	self.knownChainState.Beacon.Height = self.BestState.Beacon.BeaconHeight
	self.syncStatus.Beacon = true

	// go func() {
	// 	for {
	// 		select {
	// 		case <-self.cQuitSync:
	// 			return
	// 		default:
	// 			if self.config.NodeMode == "auto" || self.config.NodeMode == "beacon" {
	// 				if role, _ := self.BestState.Beacon.GetPubkeyRole(self.config.UserKeySet.GetPublicKeyB58()); role == "beacon-proposer" || role == "beacon-validator" {
	// 					for shardID := 0; shardID < common.SHARD_NUMBER; shardID++ {

	// 					}
	// 				}
	// 			}
	// 		}
	// 	}
	// }()

	go func() {
		//used for fancy block retriever but too lazy to implement that now :p
		var peerChainState map[libp2p.ID]PeerBeaconChainState
		peerChainState = make(map[libp2p.ID]PeerBeaconChainState)
		_ = peerChainState

		getStateWaitTime := time.Duration(defaultGetStateWaitTime)
		for {
			select {
			case <-self.cQuitSync:
				return
			case beaconState := <-self.BeaconStateCh:
				if self.BestState.Beacon.BeaconHeight < beaconState.State.Height {
					if self.knownChainState.Beacon.Height < beaconState.State.Height {
						self.knownChainState.Beacon = *beaconState.State
						if getStateWaitTime == defaultGetStateWaitTime*2 {
							getStateWaitTime -= defaultGetStateWaitTime
						}
						go self.config.Server.PushMessageGetBlockBeacon(self.BestState.Beacon.BeaconHeight+1, beaconState.State.Height, beaconState.Peer)
					} else {
						if getStateWaitTime == defaultGetStateWaitTime {
							getStateWaitTime += defaultGetStateWaitTime
						}
					}
				} else {
					if getStateWaitTime == defaultGetStateWaitTime {
						getStateWaitTime += defaultGetStateWaitTime
					}
				}

				// check shardToBeacon pool state
				if len(beaconState.State.ShardsPoolState) > 0 {
					myPoolPending := self.config.ShardToBeaconPool.GetValidPendingBlockHash()
					for shardID, peerPoolBlks := range beaconState.State.ShardsPoolState {
						myShardPoolBlks, ok := myPoolPending[shardID]
						if ok {
							blksNeedToSync := GetDiffHashesOf(peerPoolBlks, myShardPoolBlks)
							for _, blkHash := range blksNeedToSync {
								go self.config.Server.PushMessageGetShardToBeacon(shardID, blkHash)
							}
						} else {
							// sync all blks of this shard
							for _, blkHash := range peerPoolBlks {
								go self.config.Server.PushMessageGetShardToBeacon(shardID, blkHash)
							}
						}
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
							err = self.InsertBeaconBlock(newBlk, false)
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
				if self.knownChainState.Beacon.Height > self.BestState.Beacon.BeaconHeight {
					needToSync := self.knownChainState.Beacon.Height - self.BestState.Beacon.BeaconHeight
					for offset := uint64(0); offset <= needToSync; offset++ {
						blks, err := self.config.NodeBeaconPool.GetBlocks(self.BestState.Beacon.BeaconHeight + 1)
						if err != nil {
							Logger.log.Error(err)
							continue
						}
						for _, newBlk := range blks {
							err = self.InsertBeaconBlock(&newBlk, false)
							if err != nil {
								Logger.log.Error(err)
								continue
							}
						}
					}
				}
			}
		}
	}()
	return nil
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

//GetDiffHashesOf Get unique hashes of 1st slice compare to 2nd slice
func GetDiffHashesOf(slice1 []common.Hash, slice2 []common.Hash) []common.Hash {
	var diff []common.Hash

	for _, s1 := range slice1 {
		found := false
		for _, s2 := range slice2 {
			if s1 == s2 {
				found = true
				break
			}
		}
		if !found {
			diff = append(diff, s1)
		}
	}

	return diff
}
