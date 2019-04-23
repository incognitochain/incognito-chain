package blockchain

import (
	"fmt"

	"github.com/constant-money/constant-chain/cashec"
	"github.com/constant-money/constant-chain/common"
	"github.com/constant-money/constant-chain/common/base58"
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
	blockchain.syncStatus.Lock()
	for shardID := range blockchain.syncStatus.Shards {
		if shardState, ok := (*shard)[shardID]; ok {
			if shardState.Height > blockchain.BestState.Shard[shardID].ShardHeight {
				pState.Shard[shardID] = &shardState
			}
		}
	}
	blockchain.syncStatus.Unlock()

	blockchain.syncStatus.PeersStateLock.Lock()
	blockchain.syncStatus.PeersState[pState.Peer] = pState
	blockchain.syncStatus.PeersStateLock.Unlock()
}

func (blockchain *BlockChain) OnBlockShardReceived(newBlk *ShardBlock) {
	if _, ok := blockchain.syncStatus.Shards[newBlk.Header.ShardID]; ok {
		fmt.Printf("Shard block received from shard %+v \n", newBlk.Header.ShardID)
		if blockchain.BestState.Shard[newBlk.Header.ShardID].ShardHeight < newBlk.Header.Height {
			blkHash := newBlk.Header.Hash()
			err := cashec.ValidateDataB58(base58.Base58Check{}.Encode(newBlk.Header.ProducerAddress.Pk, common.ZeroByte), newBlk.ProducerSig, blkHash.GetBytes())
			if err != nil {
				Logger.log.Error(err)
				return
			} else {
				if blockchain.BestState.Shard[newBlk.Header.ShardID].ShardHeight == newBlk.Header.Height-1 && blockchain.config.UserKeySet != nil {
					userRole := blockchain.BestState.Shard[newBlk.Header.ShardID].GetPubkeyRole(blockchain.config.UserKeySet.GetPublicKeyB58(), 0)
					if !blockchain.ConsensusOngoing {
						if userRole == common.PROPOSER_ROLE || userRole == common.VALIDATOR_ROLE {
							fmt.Println("Shard block insert", newBlk.Header.Height)
							err = blockchain.InsertShardBlock(newBlk, false)
							if err != nil {
								Logger.log.Error(err)
								return
							}
						}
					} else {
						if userRole == common.PROPOSER_ROLE || userRole == common.VALIDATOR_ROLE {
							return
						}
					}
				}
				err = blockchain.config.ShardPool[newBlk.Header.ShardID].AddShardBlock(newBlk)
				if err != nil {
					fmt.Println("Shard block add pool err", err)
				}
			}
		}
	}
}

func (blockchain *BlockChain) OnBlockBeaconReceived(newBlk *BeaconBlock) {
	if blockchain.syncStatus.Beacon {
		fmt.Println("Beacon block received", newBlk.Header.Height, blockchain.BestState.Beacon.BeaconHeight)
		if blockchain.BestState.Beacon.BeaconHeight < newBlk.Header.Height {
			blkHash := newBlk.Header.Hash()
			err := cashec.ValidateDataB58(base58.Base58Check{}.Encode(newBlk.Header.ProducerAddress.Pk, common.ZeroByte), newBlk.ProducerSig, blkHash.GetBytes())
			if err != nil {
				fmt.Println("Beacon block validate err", err)
				Logger.log.Error(err)
				return
			} else {
				if blockchain.BestState.Beacon.BeaconHeight == newBlk.Header.Height-1 && blockchain.config.UserKeySet != nil {
					if !blockchain.ConsensusOngoing {
						userRole, _ := blockchain.BestState.Beacon.GetPubkeyRole(blockchain.config.UserKeySet.GetPublicKeyB58(), 0)
						if userRole == common.PROPOSER_ROLE || userRole == common.VALIDATOR_ROLE {
							fmt.Println("Beacon block insert", newBlk.Header.Height)
							err = blockchain.InsertBeaconBlock(newBlk, false)
							if err != nil {
								Logger.log.Error(err)
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

	if blockchain.IsReady(false, 0) {

		fmt.Println("Blockchain Message/OnShardToBeaconBlockReceived: Block Height", block.Header.Height)
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

		if err = ValidateAggSignature(block.ValidatorsIdx, blockchain.BestState.Beacon.ShardCommittee[block.Header.ShardID], block.AggregatedSig, block.R, block.Hash()); err != nil {
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
			blockchain.SyncBlkShardToBeacon(block.Header.ShardID, false, false, false, nil, nil, from, to, "")
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

	expectedHeight, toShardID, err := blockchain.config.CrossShardPool[block.ToShardID].AddCrossShardBlock(block)
	for fromShardID, height := range expectedHeight {
		// fmt.Printf("Shard %+v request CrossShardBlock with Height %+v from shard %+v \n", toShardID, height, fromShardID)
		blockchain.SyncBlkCrossShard(false, false, []common.Hash{}, []uint64{height}, fromShardID, toShardID, "")
	}

	if err != nil {
		if err.Error() != "receive old block" && err.Error() != "receive duplicate block" {
			Logger.log.Error(err)
			return
		}
	}
}
