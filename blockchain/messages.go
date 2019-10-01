package blockchain

import (
	"fmt"
	"sync"

	"github.com/incognitochain/incognito-chain/common"
	libp2p "github.com/libp2p/go-libp2p-peer"
)

func (blockchain *BlockChain) OnPeerStateReceived(beacon *ChainState, shard *map[byte]ChainState, shardToBeaconPool *map[byte][]uint64, crossShardPool *map[byte]map[byte][]uint64, peerID libp2p.ID) {
	if blockchain.IsTest {
		return
	}
	if beacon.Timestamp < GetBeaconBestState().BestBlock.Header.Timestamp && beacon.Height > GetBeaconBestState().BestBlock.Header.Height {
		return
	}

	var (
		userRole    string
		userShardID byte
	)

	userRole, userShardIDInt := blockchain.config.ConsensusEngine.GetUserRole()
	if userRole == common.ShardRole {
		userShardID = byte(userShardIDInt)
	}
	// miningKey, _ := blockchain.config.ConsensusEngine.GetCurrentMiningPublicKey()
	// if miningKey != "" {
	// 	userRole, userShardID = blockchain.BestState.Beacon.GetPubkeyRole(miningKey, blockchain.BestState.Beacon.BestBlock.Header.Round)
	// }
	pState := &peerState{
		Shard:  make(map[byte]*ChainState),
		Beacon: beacon,
		Peer:   peerID,
	}
	nodeMode := blockchain.config.NodeMode
	if userRole == common.BeaconRole {
		pState.ShardToBeaconPool = shardToBeaconPool
		for shardID := byte(0); shardID < byte(common.MaxShardNumber); shardID++ {
			if shardState, ok := (*shard)[shardID]; ok {
				if shardState.Height > GetBeaconBestState().GetBestHeightOfShard(shardID) {
					pState.Shard[shardID] = &shardState
				}
			}
		}
	}
	if userRole == common.ShardRole && (nodeMode == common.NodeModeAuto || nodeMode == common.NodeModeBeacon) {
		// userShardRole = blockchain.BestState.Shard[userShardID].GetPubkeyRole(miningKey, blockchain.BestState.Shard[userShardID].BestBlock.Header.Round)
		// if userShardRole == common.ProposerRole || userShardRole == common.ValidatorRole {
		if shardState, ok := (*shard)[userShardID]; ok && shardState.Height >= blockchain.BestState.Shard[userShardID].ShardHeight {
			pState.Shard[userShardID] = &shardState
			if pool, ok := (*crossShardPool)[userShardID]; ok {
				pState.CrossShardPool = make(map[byte]*map[byte][]uint64)
				pState.CrossShardPool[userShardID] = &pool
			}
		}
		// }
	}
	blockchain.Synker.Status.Lock()
	for shardID := 0; shardID < blockchain.BestState.Beacon.ActiveShards; shardID++ {
		if shardState, ok := (*shard)[byte(shardID)]; ok {
			if shardState.Height > GetBestStateShard(byte(shardID)).ShardHeight && (*shard)[byte(shardID)].Timestamp > GetBestStateShard(byte(shardID)).BestBlock.Header.Timestamp {
				pState.Shard[byte(shardID)] = &shardState
			}
		}
	}
	blockchain.Synker.Status.Unlock()
	//TODO hy
	blockchain.Synker.States.Lock()
	if blockchain.Synker.States.PeersState != nil {
		blockchain.Synker.States.PeersState[pState.Peer] = pState
	}
	blockchain.Synker.States.Unlock()
}

func (blockchain *BlockChain) OnBlockShardReceived(newBlk *ShardBlock) {
	if blockchain.IsTest {
		return
	}
	fmt.Println("Shard block received from shard", newBlk.Header.ShardID, newBlk.Header.Height)
	if newBlk.Header.Timestamp < GetBestStateShard(newBlk.Header.ShardID).BestBlock.Header.Timestamp { // not receive block older than current latest block
		return
	}

	if _, ok := blockchain.Synker.Status.Shards[newBlk.Header.ShardID]; ok {
		if _, ok := currentInsert.Shards[newBlk.Header.ShardID]; !ok {
			currentInsert.Shards[newBlk.Header.ShardID] = &sync.Mutex{}
		}

		currentInsert.Shards[newBlk.Header.ShardID].Lock()
		defer currentInsert.Shards[newBlk.Header.ShardID].Unlock()
		currentShardBestState := blockchain.BestState.Shard[newBlk.Header.ShardID]
		if currentShardBestState.ShardHeight <= newBlk.Header.Height {
			userPubKey, _ := blockchain.config.ConsensusEngine.GetCurrentMiningPublicKey()
			if userPubKey != "" {

				userRole := currentShardBestState.GetPubkeyRole(userPubKey, 0)
				fmt.Println("Shard block received 1", userRole)
				if userRole == common.ProposerRole || userRole == common.ValidatorRole {
					// Revert beststate
					// @NOTICE: Choose block with highest round, because we assume that most of node state is at the highest round
					if currentShardBestState.ShardHeight == newBlk.Header.Height && currentShardBestState.BestBlock.Header.Timestamp < newBlk.Header.Timestamp && currentShardBestState.BestBlock.Header.Round < newBlk.Header.Round {
						fmt.Println("FORK SHARD", newBlk.Header.ShardID, newBlk.Header.Height)
						if err := blockchain.ValidateBlockWithPrevShardBestState(newBlk); err != nil {
							Logger.log.Error(err)
							return
						}
						if err := blockchain.RevertShardState(newBlk.Header.ShardID); err != nil {
							Logger.log.Error(err)
							return
						}
						fmt.Println("REVERTED SHARD", newBlk.Header.ShardID, newBlk.Header.Height)
						err := blockchain.InsertShardBlock(newBlk, true)
						if err != nil {
							Logger.log.Error(err)
						}
						return
					}

					isConsensusOngoing := blockchain.config.ConsensusEngine.IsOngoing(common.GetShardChainKey(newBlk.Header.ShardID))
					fmt.Println("Shard block received 2", currentShardBestState.ShardHeight, newBlk.Header.Height)
					if currentShardBestState.ShardHeight == newBlk.Header.Height-1 {
						fmt.Println("Shard block received 3", isConsensusOngoing, blockchain.Synker.IsLatest(true, newBlk.Header.ShardID))
						if blockchain.Synker.IsLatest(true, newBlk.Header.ShardID) == false {
							Logger.log.Info("Insert New Shard Block to pool", newBlk.Header.Height)
							err := blockchain.config.ShardPool[newBlk.Header.ShardID].AddShardBlock(newBlk)
							if err != nil {
								Logger.log.Errorf("Add block %+v from shard %+v error %+v: \n", newBlk.Header.Height, newBlk.Header.ShardID, err)
								return
							}
						} else if !isConsensusOngoing {
							Logger.log.Infof("Insert New Shard Block %+v, ShardID %+v \n", newBlk.Header.Height, newBlk.Header.ShardID)
							err := blockchain.InsertShardBlock(newBlk, true)
							if err != nil {
								Logger.log.Error(err)
								return
							}
						}
					}
				}
			}

			err := blockchain.config.ShardPool[newBlk.Header.ShardID].AddShardBlock(newBlk)
			if err != nil {
				Logger.log.Errorf("Add block %+v from shard %+v error %+v: \n", newBlk.Header.Height, newBlk.Header.ShardID, err)
			}
		}
	}
}

func (blockchain *BlockChain) OnBlockBeaconReceived(newBlk *BeaconBlock) {
	if blockchain.IsTest {
		return
	}
	if blockchain.Synker.Status.Beacon {
		fmt.Println("Beacon block received", newBlk.Header.Height, blockchain.BestState.Beacon.BeaconHeight, newBlk.Header.Timestamp)
		if newBlk.Header.Timestamp < blockchain.BestState.Beacon.BestBlock.Header.Timestamp { // not receive block older than current latest block
			return
		}
		if blockchain.BestState.Beacon.BeaconHeight <= newBlk.Header.Height {

			publicKey, _ := blockchain.config.ConsensusEngine.GetCurrentMiningPublicKey()
			if publicKey != "" {
				// Revert beststate

				userRole, _ := blockchain.BestState.Beacon.GetPubkeyRole(publicKey, 0)
				if userRole == common.ProposerRole || userRole == common.ValidatorRole {
					currentBeaconBestState := blockchain.BestState.Beacon
					if currentBeaconBestState.BeaconHeight == newBlk.Header.Height && currentBeaconBestState.BestBlock.Header.Timestamp < newBlk.Header.Timestamp && currentBeaconBestState.BestBlock.Header.Round < newBlk.Header.Round {
						fmt.Println("FORK BEACON", newBlk.Header.Height)
						if err := blockchain.ValidateBlockWithPrevBeaconBestState(newBlk); err != nil {
							Logger.log.Error(err)
							return
						}
						if err := blockchain.RevertBeaconState(); err != nil {
							Logger.log.Error(err)
							return
						}
						fmt.Println("REVERTED BEACON", newBlk.Header.Height)
						err := blockchain.InsertBeaconBlock(newBlk, false)
						if err != nil {
							Logger.log.Error(err)
						}
						return
					}

					if blockchain.BestState.Beacon.BeaconHeight == newBlk.Header.Height-1 {
						if !blockchain.config.ConsensusEngine.IsOngoing(common.BeaconChainKey) {
							fmt.Println("Beacon block insert", newBlk.Header.Height)
							err := blockchain.InsertBeaconBlock(newBlk, false)
							if err != nil {
								Logger.log.Error(err)
								return
							}
							return
						}
					}
				}
			}
			fmt.Println("Beacon block prepare add to pool", newBlk.Header.Height)
			err := blockchain.config.BeaconPool.AddBeaconBlock(newBlk)
			if err != nil {
				fmt.Println("Beacon block add pool err", err)
			}
		}

	}
}

func (blockchain *BlockChain) OnShardToBeaconBlockReceived(block *ShardToBeaconBlock) {
	if blockchain.IsTest {
		return
	}
	if blockchain.config.NodeMode == common.NodeModeBeacon || blockchain.config.NodeMode == common.NodeModeAuto {
		publicKey, _ := blockchain.config.ConsensusEngine.GetCurrentMiningPublicKey()
		beaconRole, _ := blockchain.BestState.Beacon.GetPubkeyRole(publicKey, 0)
		if beaconRole != common.ProposerRole && beaconRole != common.ValidatorRole {
			return
		}
	} else {
		return
	}

	if blockchain.Synker.IsLatest(false, 0) {
		if block.Header.Version != SHARD_BLOCK_VERSION {
			Logger.log.Debugf("Invalid Verion of block height %+v in Shard %+v", block.Header.Height, block.Header.ShardID)
			return
		}

		//err := blockchain.config.ConsensusEngine.ValidateProducerSig(block, block.Header.ConsensusType)
		//if err != nil {
		//	Logger.log.Error(err)
		//	return
		//}
		
		from, to, err := blockchain.config.ShardToBeaconPool.AddShardToBeaconBlock(block)
		if err != nil {
			if err.Error() != "receive old block" && err.Error() != "receive duplicate block" {
				Logger.log.Error(err)
				return
			}
		}
		if from != 0 && to != 0 {
			fmt.Printf("Message/SyncBlkShardToBeacon, from %+v to %+v \n", from, to)
			blockchain.Synker.SyncBlkShardToBeacon(block.Header.ShardID, false, false, false, nil, nil, from, to, "")
		}
	}
}

func (blockchain *BlockChain) OnCrossShardBlockReceived(block *CrossShardBlock) {
	Logger.log.Info("Received CrossShardBlock", block.Header.Height, block.Header.ShardID)
	if blockchain.IsTest {
		return
	}
	if blockchain.config.NodeMode == common.NodeModeShard || blockchain.config.NodeMode == common.NodeModeAuto {
		publickey, _ := blockchain.config.ConsensusEngine.GetCurrentMiningPublicKey()
		shardRole := blockchain.BestState.Shard[block.ToShardID].GetPubkeyRole(publickey, 0)
		if shardRole != common.ProposerRole && shardRole != common.ValidatorRole {
			return
		}
	} else {
		return
	}
	expectedHeight, toShardID, err := blockchain.config.CrossShardPool[block.ToShardID].AddCrossShardBlock(block)
	for fromShardID, height := range expectedHeight {
		// fmt.Printf("Shard %+v request CrossShardBlock with Height %+v from shard %+v \n", toShardID, height, fromShardID)
		blockchain.Synker.SyncBlkCrossShard(false, false, []common.Hash{}, []uint64{height}, fromShardID, toShardID, "")
	}
	if err != nil {
		if err.Error() != "receive old block" && err.Error() != "receive duplicate block" {
			Logger.log.Error(err)
			return
		}
	}

}
