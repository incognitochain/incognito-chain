package blockchain

import (
	libp2p "github.com/libp2p/go-libp2p-peer"
)

func (self *BlockChain) OnBlockShardReceived(block *ShardBlock) {
	if self.newShardBlkCh[block.Header.ShardID] != nil {
		*self.newShardBlkCh[block.Header.ShardID] <- block
	}
}
func (self *BlockChain) OnBlockBeaconReceived(block *BeaconBlock) {
	if self.syncStatus.Beacon {
		self.newBeaconBlkCh <- block
	}
}

func (self *BlockChain) GetBeaconState() (*BeaconChainState, error) {
	state := &BeaconChainState{
		Height:          self.BestState.Beacon.BeaconHeight,
		BlockHash:       self.BestState.Beacon.BestBlockHash,
		ShardsPoolState: self.config.ShardToBeaconPool.GetPendingBlockHashes(),
	}
	return state, nil
}

func (self *BlockChain) OnBeaconStateReceived(state *BeaconChainState, peerID libp2p.ID) {
	if self.syncStatus.Beacon {
		self.BeaconStateCh <- &PeerBeaconChainState{
			state, peerID,
		}
	}
}

func (self *BlockChain) GetShardState(shardID byte) *ShardChainState {
	state := &ShardChainState{
		Height:    self.BestState.Shard[shardID].ShardHeight,
		ShardID:   shardID,
		BlockHash: self.BestState.Shard[shardID].BestShardBlockHash,
	}
	return state
}

func (self *BlockChain) OnShardStateReceived(state *ShardChainState, peerID libp2p.ID) {
	if self.newShardBlkCh[state.ShardID] != nil {
		self.ShardStateCh[state.ShardID] <- &PeerShardChainState{
			state, peerID,
		}
	}
}

func (self *BlockChain) OnShardToBeaconBlockReceived(block ShardToBeaconBlock) {
	err := self.config.ShardToBeaconPool.ValidateShardToBeaconBlock(block)
	if err != nil {
		Logger.log.Error(err)
	} else {
		// Add to pending or queue
		// Add into pending?
		// Add into queue?
		err = self.config.ShardToBeaconPool.AddShardBeaconBlock(block)
		if err != nil {
			Logger.log.Error(err)
		}
	}
}

func (self *BlockChain) OnCrossShardBlockReceived(block CrossShardBlock) {
	err := self.config.CrossShardPool.AddCrossShardBlock(block)
	if err != nil {
		Logger.log.Error(err)
	}
}
