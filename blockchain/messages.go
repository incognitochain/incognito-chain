package blockchain

import (
	"sync"

	"github.com/incognitochain/incognito-chain/common"
)

func (blockchain *BlockChain) OnPeerStateReceived(
	beacon *ChainState,
	shard *map[byte]ChainState,
	shardToBeaconPool *map[byte][]uint64,
	crossShardPool *map[byte]map[byte][]uint64,
	peerMiningKey string,
) {
	if blockchain.IsTest {
		return
	}
	if beacon.Timestamp < blockchain.GetBeaconBestState().BestBlock.Header.Timestamp && beacon.Height > blockchain.GetBeaconBestState().BestBlock.Header.Height {
		return
	}

	var (
		userRole    string
		userShardID byte
	)

	userRole, userShardIDInt := blockchain.config.ConsensusEngine.GetUserLayer()
	if userRole == common.ShardRole {
		userShardID = byte(userShardIDInt)
	}

	pState := &PeerState{
		Shard:               make(map[byte]*ChainState),
		Beacon:              beacon,
		PeerMiningPublicKey: peerMiningKey,
	}
	nodeMode := blockchain.config.NodeMode
	if userRole == common.BeaconRole {
		pState.ShardToBeaconPool = shardToBeaconPool
		for shardID := byte(0); shardID < byte(common.MaxShardNumber); shardID++ {
			if shardState, ok := (*shard)[shardID]; ok {
				if shardState.Height > blockchain.GetBeaconBestState().GetBestHeightOfShard(shardID) {
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
	for shardID := 0; shardID < blockchain.BestState.Beacon.ActiveShards; shardID++ {
		if shardState, ok := (*shard)[byte(shardID)]; ok {
			if shardState.Height > blockchain.GetBestStateShard(byte(shardID)).ShardHeight && (*shard)[byte(shardID)].Timestamp > blockchain.GetBestStateShard(byte(shardID)).BestBlock.Header.Timestamp {
				pState.Shard[byte(shardID)] = &shardState
			}
		}
	}
	blockchain.Synker.States.Lock()
	if blockchain.Synker.States.PeersState != nil {
		blockchain.Synker.States.PeersState[peerMiningKey] = pState
	}
	blockchain.Synker.States.Unlock()
}

func (blockchain *BlockChain) OnBlockShardReceived(newBlk *ShardBlock) {
	if blockchain.IsTest {
		return
	}
	Logger.log.Debugf("Received shard block  message from shard %d block %d", newBlk.Header.ShardID, newBlk.Header.Height)

	if _, ok := blockchain.Synker.Status.Shards[newBlk.Header.ShardID]; ok {
		if _, ok := currentInsert.Shards[newBlk.Header.ShardID]; !ok {
			currentInsert.Shards[newBlk.Header.ShardID] = &sync.Mutex{}
		}

		currentInsert.Shards[newBlk.Header.ShardID].Lock()
		defer currentInsert.Shards[newBlk.Header.ShardID].Unlock()
		currentShardBestState := blockchain.BestState.Shard[newBlk.Header.ShardID]

		if currentShardBestState.ShardHeight <= newBlk.Header.Height {
			//layer, role, _ := blockchain.config.ConsensusEngine.GetUserRole()
			currentShardBestState := blockchain.BestState.Shard[newBlk.Header.ShardID]

			if currentShardBestState.ShardHeight == newBlk.Header.Height && currentShardBestState.BestBlock.Header.Timestamp < newBlk.Header.Timestamp && currentShardBestState.BestBlock.Header.Round < newBlk.Header.Round {
				err := blockchain.InsertShardBlock(newBlk, false)
				if err != nil {
					Logger.log.Error(err)
				}
				return
			}

			err := blockchain.config.ShardPool[newBlk.Header.ShardID].AddShardBlock(newBlk)
			if err != nil {
				Logger.log.Warnf("Shard block pool add block %d from shard %d with error %s: \n", newBlk.Header.Height, newBlk.Header.ShardID, err.Error())
			}
		}
	}
}

func (blockchain *BlockChain) OnBlockBeaconReceived(newBlk *BeaconBlock) {
	if blockchain.IsTest {
		return
	}
	if blockchain.Synker.Status.Beacon {
		Logger.log.Debugf("Beacon block received at height: %d, current  beacon best state: %d, received block time %+v", newBlk.Header.Height, blockchain.BestState.Beacon.BeaconHeight, newBlk.Header.Timestamp)
		if newBlk.Header.Timestamp < blockchain.BestState.Beacon.BestBlock.Header.Timestamp { // not receive block older than current latest block
			return
		}
		if blockchain.BestState.Beacon.BeaconHeight <= newBlk.Header.Height {
			currentBeaconBestState := blockchain.BestState.Beacon
			if currentBeaconBestState.BeaconHeight == newBlk.Header.Height && currentBeaconBestState.BestBlock.Header.Timestamp < newBlk.Header.Timestamp && currentBeaconBestState.BestBlock.Header.Round < newBlk.Header.Round {
				Logger.log.Debugf("Beacon block insert %d", newBlk.Header.Height)
				err := blockchain.InsertBeaconBlock(newBlk, false)
				if err != nil {
					Logger.log.Error(err)
					return
				}
				return
			}
			Logger.log.Debugf("Beacon block prepare add to pool, new block height: %d", newBlk.Header.Height)
			err := blockchain.config.BeaconPool.AddBeaconBlock(newBlk)
			if err != nil {
				Logger.log.Debugf("Beacon block add pool err %+v", err)
			}
		}

	}
}

func (blockchain *BlockChain) OnShardToBeaconBlockReceived(block *ShardToBeaconBlock) {
	if blockchain.IsTest {
		return
	}
	Logger.log.Infof("[sync] OnShardToBeaconBlockReceived NodeMode: %+v", blockchain.config.NodeMode)
	if blockchain.config.NodeMode == common.NodeModeBeacon || blockchain.config.NodeMode == common.NodeModeAuto {
		layer, role, _ := blockchain.config.ConsensusEngine.GetUserRole()
		Logger.log.Infof("OnShardToBeaconBlockReceived layer && role: %+v %+v", layer, role)
		if layer != common.BeaconRole || role != common.CommitteeRole {
			return
		}
	} else {
		return
	}

	Logger.log.Infof("[sync] OnShardToBeaconBlockReceived IsLatest: %+v", blockchain.Synker.IsLatest(false, 0))
	if blockchain.Synker.IsLatest(false, 0) {
		Logger.log.Info("[sync] OnShardToBeaconBlockReceived IsLatest!")
		if block.Header.Version != SHARD_BLOCK_VERSION {
			Logger.log.Info("[sync] Damn it, wrong block version!")
			Logger.log.Debugf("[sync] Invalid Verion of block height %+v in Shard %+v", block.Header.Height, block.Header.ShardID)
			return
		}

		from, to, err := blockchain.config.ShardToBeaconPool.AddShardToBeaconBlock(block)
		Logger.log.Infof("[sync] AddShardToBeaconBlock return from:%v to:%v err:%v!", from, to, err)
		if err != nil {
			if err.Error() != "receive old block" && err.Error() != "receive duplicate block" {
				Logger.log.Error(err)
				return
			}
		}
		if from != 0 && to != 0 {
			Logger.log.Infof("[sync] Message/SyncBlkShardToBeacon, from %+v to %+v \n", from, to)
			blockchain.Synker.SyncBlkShardToBeacon(block.Header.ShardID, false, false, false, nil, nil, from, to, "")
		}
	} else {
		Logger.log.Info("[sync] OnShardToBeaconBlockReceived Is not Latest!")
	}
}

func (blockchain *BlockChain) OnCrossShardBlockReceived(block *CrossShardBlock) {
	if blockchain.IsTest {
		return
	}
	Logger.log.Infof("[sync] Received CrossShardBlock blk Height %v shardID %v", block.Header.Height, block.Header.ShardID)
	if blockchain.IsTest {
		return
	}
	if blockchain.config.NodeMode == common.NodeModeShard || blockchain.config.NodeMode == common.NodeModeAuto {
		layer, role, _ := blockchain.config.ConsensusEngine.GetUserRole()
		if layer != common.ShardRole || role != common.CommitteeRole {
			return
		}
	} else {
		return
	}
	expectedHeight, toShardID, err := blockchain.config.CrossShardPool[block.ToShardID].AddCrossShardBlock(block)
	if err != nil {
		if err.Error() != "receive old block" && err.Error() != "receive duplicate block" {
			Logger.log.Error(err)
			return
		}
	}
	Logger.log.Infof("[sync] Shard %+v After insert cross shard block %v: expectedHeight %v, toShardID %v \n", block.ToShardID, block.GetHeight(), expectedHeight, toShardID)
}
