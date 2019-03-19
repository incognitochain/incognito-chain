package blockchain

import (
	"fmt"

	"github.com/constant-money/constant-chain/cashec"
	"github.com/constant-money/constant-chain/common"
	libp2p "github.com/libp2p/go-libp2p-peer"
)

func (blockchain *BlockChain) OnPeerStateReceived(beacon *ChainState, shard *map[byte]ChainState, shardToBeaconPool *map[byte][]uint64, crossShardPool *map[byte]map[byte][]uint64, peerID libp2p.ID) {
	if beacon.Height >= blockchain.BestState.Beacon.BeaconHeight {
		pState := &peerState{
			Shard:  make(map[byte]*ChainState),
			Beacon: beacon,
			Peer:   peerID,
		}
		userRole, userShardID := blockchain.BestState.Beacon.GetPubkeyRole(blockchain.config.UserKeySet.GetPublicKeyB58(), blockchain.BestState.Beacon.BestBlock.Header.Round)
		nodeMode := blockchain.config.NodeMode
		if userRole == common.PROPOSER_ROLE || userRole == common.VALIDATOR_ROLE {
			pState.ShardToBeaconPool = shardToBeaconPool
			for shardID := byte(0); shardID < byte(common.MAX_SHARD_NUMBER); shardID++ {
				if shardState, ok := (*shard)[shardID]; ok {
					if shardState.Height > blockchain.BestState.Beacon.BestShardHeight[shardID] {
						pState.Shard[shardID] = &shardState
					}
				}
			}
		}
		if userRole == common.SHARD_ROLE && (nodeMode == common.NODEMODE_AUTO || nodeMode == common.NODEMODE_BEACON) {
			userShardRole := blockchain.BestState.Shard[userShardID].GetPubkeyRole(blockchain.config.UserKeySet.GetPublicKeyB58(), blockchain.BestState.Shard[userShardID].BestBlock.Header.Round)
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
		for shardID := range blockchain.syncStatus.Shards {
			if shardState, ok := (*shard)[shardID]; ok {
				if shardState.Height > blockchain.BestState.Shard[shardID].ShardHeight {
					pState.Shard[shardID] = &shardState
				}
			}
		}
		blockchain.syncStatus.PeersStateLock.Lock()
		blockchain.syncStatus.PeersState[pState.Peer] = pState
		blockchain.syncStatus.PeersStateLock.Unlock()
	}
}

func (blockchain *BlockChain) OnBlockShardReceived(newBlk *ShardBlock) {
	if _, ok := blockchain.syncStatus.Shards[newBlk.Header.ShardID]; ok {
		fmt.Printf("Shard block received from shard %+v \n", newBlk.Header.ShardID)
		if blockchain.BestState.Shard[newBlk.Header.ShardID].ShardHeight < newBlk.Header.Height {
			blkHash := newBlk.Header.Hash()
			err := cashec.ValidateDataB58(newBlk.Header.Producer, newBlk.ProducerSig, blkHash.GetBytes())
			if err != nil {
				Logger.log.Error(err)
				return
			} else {
				if blockchain.BestState.Shard[newBlk.Header.ShardID].ShardHeight == newBlk.Header.Height-1 {
					err = blockchain.InsertShardBlock(newBlk, false)
					if err != nil {
						if err.(*BlockChainError).Code != -26 {
							Logger.log.Error(err)
						}
						return
					}
				} else {
					blockchain.config.ShardPool[newBlk.Header.ShardID].AddShardBlock(newBlk)
				}
			}
		}
	}
}

func (blockchain *BlockChain) OnBlockBeaconReceived(newBlk *BeaconBlock) {
	if blockchain.syncStatus.Beacon {
		fmt.Println("Beacon block received", newBlk.Header.Height)
		if blockchain.BestState.Beacon.BeaconHeight < newBlk.Header.Height {
			blkHash := newBlk.Header.Hash()
			err := cashec.ValidateDataB58(newBlk.Header.Producer, newBlk.ProducerSig, blkHash.GetBytes())
			if err != nil {
				Logger.log.Error(err)
				return
			} else {
				if blockchain.BestState.Beacon.BeaconHeight == newBlk.Header.Height-1 {
					err = blockchain.InsertBeaconBlock(newBlk, false)
					if err != nil {
						if err.(*BlockChainError).Code != -26 {
							Logger.log.Error(err)
						}
						return
					}
				} else {
					blockchain.config.BeaconPool.AddBeaconBlock(newBlk)
				}
			}
		}
	}
}

func (blockchain *BlockChain) OnShardToBeaconBlockReceived(block ShardToBeaconBlock) {
	//TODO: check node mode -> node mode & role before add block to pool
	if blockchain.IsReady(false, 0) {
		fmt.Println("Blockchain Message/OnShardToBeaconBlockReceived: Block Height", block.Header.Height)
		blkHash := block.Header.Hash()
		err := cashec.ValidateDataB58(block.Header.Producer, block.ProducerSig, blkHash.GetBytes())

		if err != nil {
			Logger.log.Debugf("Invalid Producer Signature of block height %+v in Shard %+v", block.Header.Height, block.Header.ShardID)
			return
		}
		if block.Header.Version != VERSION {
			Logger.log.Debugf("Invalid Verion of block height %+v in Shard %+v", block.Header.Height, block.Header.ShardID)
			return
		}

		//TODO: what if shard to beacon from old committee
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
	Logger.log.Criticalf("Received CrossShardBlock %+v from %+v \n", block.Header.Height, block.Header.ShardID)
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
