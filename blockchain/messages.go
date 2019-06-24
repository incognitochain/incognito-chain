package blockchain

import (
	"fmt"
	"sync"

	"github.com/incognitochain/incognito-chain/cashec"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/common/base58"
	libp2p "github.com/libp2p/go-libp2p-peer"
)

func (blockchain *BlockChain) OnPeerStateReceived(beacon *ChainState, shard *map[byte]ChainState, shardToBeaconPool *map[byte][]uint64, crossShardPool *map[byte]map[byte][]uint64, peerID libp2p.ID) {
	var (
		userRole      string
		userShardID   byte
		userShardRole string
	)
	if blockchain.config.UserKeySet != nil {
		userRole, userShardID = blockchain.BestState.Beacon.GetPubkeyRole(blockchain.config.UserKeySet.GetPublicKeyB58(), blockchain.BestState.Beacon.BestBlock.Header.Round)
	}
	pState := &peerState{
		Shard:  make(map[byte]*ChainState),
		Beacon: beacon,
		Peer:   peerID,
	}
	nodeMode := blockchain.config.NodeMode
	if userRole == common.PROPOSER_ROLE || userRole == common.VALIDATOR_ROLE {
		pState.ShardToBeaconPool = shardToBeaconPool
		for shardID := byte(0); shardID < byte(common.MAX_SHARD_NUMBER); shardID++ {
			if shardState, ok := (*shard)[shardID]; ok {
				if shardState.Height > GetBestStateBeacon().GetBestHeightOfShard(shardID) {
					pState.Shard[shardID] = &shardState
				}
			}
		}
	}
	if userRole == common.SHARD_ROLE && (nodeMode == common.NODEMODE_AUTO || nodeMode == common.NODEMODE_BEACON) {
		userShardRole = blockchain.BestState.Shard[userShardID].GetPubkeyRole(blockchain.config.UserKeySet.GetPublicKeyB58(), blockchain.BestState.Shard[userShardID].BestBlock.Header.Round)
		if userShardRole == common.PROPOSER_ROLE || userShardRole == common.VALIDATOR_ROLE {
			if shardState, ok := (*shard)[userShardID]; ok && shardState.Height >= blockchain.BestState.Shard[userShardID].ShardHeight {
				pState.Shard[userShardID] = &shardState
				if pool, ok := (*crossShardPool)[userShardID]; ok {
					pState.CrossShardPool = make(map[byte]*map[byte][]uint64)
					pState.CrossShardPool[userShardID] = &pool
				}
			}
		}
	}
	blockchain.Synker.Status.Lock()
	for shardID := range blockchain.Synker.Status.Shards {
		if shardState, ok := (*shard)[shardID]; ok {
			if shardState.Height > blockchain.BestState.Shard[shardID].ShardHeight {
				pState.Shard[shardID] = &shardState
			}
		}
	}
	blockchain.Synker.Status.Unlock()

	blockchain.Synker.States.Lock()
	blockchain.Synker.States.PeersState[pState.Peer] = pState
	blockchain.Synker.States.Unlock()
}

func (blockchain *BlockChain) OnBlockShardReceived(newBlk *ShardBlock) {
	fmt.Println("Shard block received from shard A", newBlk.Header.ShardID, newBlk.Header.Height)
	if _, ok := blockchain.Synker.Status.Shards[newBlk.Header.ShardID]; ok {
		if _, ok := currentInsert.Shards[newBlk.Header.ShardID]; !ok {
			currentInsert.Shards[newBlk.Header.ShardID] = &sync.Mutex{}
		}

		currentInsert.Shards[newBlk.Header.ShardID].Lock()
		defer currentInsert.Shards[newBlk.Header.ShardID].Unlock()
		fmt.Println("Shard block received from shard B", newBlk.Header.ShardID, newBlk.Header.Height)
		currentShardBestState := blockchain.BestState.Shard[newBlk.Header.ShardID]
		if currentShardBestState.ShardHeight <= newBlk.Header.Height {
			if blockchain.config.UserKeySet != nil {
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
				}

				userRole := currentShardBestState.GetPubkeyRole(blockchain.config.UserKeySet.GetPublicKeyB58(), 0)
				fmt.Println("Shard block received 1", userRole)

				if userRole == common.PROPOSER_ROLE || userRole == common.VALIDATOR_ROLE {
					fmt.Println("Shard block received 2", currentShardBestState.ShardHeight, newBlk.Header.Height)
					if currentShardBestState.ShardHeight == newBlk.Header.Height-1 {
						fmt.Println("Shard block received 3", blockchain.ConsensusOngoing, blockchain.Synker.IsLatest(true, newBlk.Header.ShardID))
						if blockchain.Synker.IsLatest(true, newBlk.Header.ShardID) == false {
							Logger.log.Info("Insert New Shard Block to pool", newBlk.Header.Height)
							err := blockchain.config.ShardPool[newBlk.Header.ShardID].AddShardBlock(newBlk)
							if err != nil {
								Logger.log.Errorf("Add block %+v from shard %+v error %+v: \n", newBlk.Header.Height, newBlk.Header.ShardID, err)
							}
						} else if !blockchain.ConsensusOngoing {
							Logger.log.Infof("Insert New Shard Block %+v, ShardID %+v \n", newBlk.Header.Height, newBlk.Header.ShardID)
							err := blockchain.InsertShardBlock(newBlk, false)
							if err != nil {
								Logger.log.Error(err)
							}
						}
						return
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
	if blockchain.Synker.Status.Beacon {
		fmt.Println("Beacon block received", newBlk.Header.Height, blockchain.BestState.Beacon.BeaconHeight)
		if blockchain.BestState.Beacon.BeaconHeight <= newBlk.Header.Height {
			blkHash := newBlk.Header.Hash()
			err := cashec.ValidateDataB58(base58.Base58Check{}.Encode(newBlk.Header.ProducerAddress.Pk, common.ZeroByte), newBlk.ProducerSig, blkHash.GetBytes())
			if err != nil {
				fmt.Println("Beacon block validate err", err)
				Logger.log.Error(err)
				return
			} else {
				if blockchain.config.UserKeySet != nil {
					// Revert beststate

					// currentBeaconBestState := blockchain.BestState.Beacon
					// if currentBeaconBestState.BeaconHeight == newBlk.Header.Height && currentBeaconBestState.BestBlock.Header.Timestamp < newBlk.Header.Timestamp && currentBeaconBestState.BestBlock.Header.Round < newBlk.Header.Round {
					// 	fmt.Println("FORK BEACON", newBlk.Header.Height)
					// 	if err := blockchain.ValidateBlockWithPrevBeaconBestState(newBlk); err != nil {
					// 		Logger.log.Error(err)
					// 		return
					// 	}
					// 	if err := blockchain.RevertBeaconState(); err != nil {
					// 		Logger.log.Error(err)
					// 		return
					// 	}
					// 	fmt.Println("REVERTED BEACON", newBlk.Header.Height)
					// }

					userRole, _ := blockchain.BestState.Beacon.GetPubkeyRole(blockchain.config.UserKeySet.GetPublicKeyB58(), 0)
					if userRole == common.PROPOSER_ROLE || userRole == common.VALIDATOR_ROLE {
						if blockchain.BestState.Beacon.BeaconHeight == newBlk.Header.Height-1 {
							if !blockchain.ConsensusOngoing {
								fmt.Println("Beacon block insert", newBlk.Header.Height)
								err = blockchain.InsertBeaconBlock(newBlk, false)
								if err != nil {
									Logger.log.Error(err)
									return
								}
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
}

func (blockchain *BlockChain) OnShardToBeaconBlockReceived(block ShardToBeaconBlock) {
	if blockchain.config.NodeMode == common.NODEMODE_BEACON || blockchain.config.NodeMode == common.NODEMODE_AUTO {
		beaconRole, _ := blockchain.BestState.Beacon.GetPubkeyRole(blockchain.config.UserKeySet.GetPublicKeyB58(), 0)
		if beaconRole != common.PROPOSER_ROLE && beaconRole != common.VALIDATOR_ROLE {
			return
		}
	} else {
		return
	}
	fmt.Println("Blockchain Message/OnShardToBeaconBlockReceived: Block Height", block.Header.ShardID, block.Header.Height, blockchain.Synker.IsLatest(false, 0))

	if blockchain.Synker.IsLatest(false, 0) {

		blkHash := block.Header.Hash()
		err := cashec.ValidateDataB58(base58.Base58Check{}.Encode(block.Header.ProducerAddress.Pk, common.ZeroByte), block.ProducerSig, blkHash.GetBytes())

		if err != nil {
			Logger.log.Debugf("Invalid Producer Signature of block height %+v in Shard %+v", block.Header.Height, block.Header.ShardID)
			return
		}
		if block.Header.Version != VERSION {
			Logger.log.Debugf("Invalid Verion of block height %+v in Shard %+v", block.Header.Height, block.Header.ShardID)
			return
		}

		if err = ValidateAggSignature(block.ValidatorsIdx, blockchain.BestState.Beacon.GetAShardCommittee(block.Header.ShardID), block.AggregatedSig, block.R, block.Hash()); err != nil {
			Logger.log.Error(err)
			return
		}

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

func (blockchain *BlockChain) OnCrossShardBlockReceived(block CrossShardBlock) {
	Logger.log.Info("Received CrossShardBlock", block.Header.Height, block.Header.ShardID)
	if blockchain.config.NodeMode == common.NODEMODE_SHARD || blockchain.config.NodeMode == common.NODEMODE_AUTO {
		shardRole := blockchain.BestState.Shard[block.ToShardID].GetPubkeyRole(blockchain.config.UserKeySet.GetPublicKeyB58(), 0)
		if shardRole != common.PROPOSER_ROLE && shardRole != common.VALIDATOR_ROLE {
			return
		}
	} else {
		return
	}

	if blockchain.Synker.IsLatest(true, block.ToShardID) {
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

}
