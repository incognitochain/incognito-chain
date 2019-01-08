package blockchain

import libp2p "github.com/libp2p/go-libp2p-peer"

func (self *BlockChain) OnBlockShardReceived(block *ShardBlock) {
	if self.newShardBlkCh != nil {

	}
}
func (self *BlockChain) OnBlockBeaconReceived(block *BeaconBlock) {
	if self.newBeaconBlkCh != nil {
		self.newBeaconBlkCh <- block
	}
}

func (self *BlockChain) GetBeaconState() (*BeaconChainState, error) {
	state := &BeaconChainState{
		Height:    self.BestState.Beacon.BeaconHeight,
		BlockHash: self.BestState.Beacon.BestBlockHash,
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

func (self *BlockChain) OnGetShardState(shardID byte) *ShardChainState {
	state := &ShardChainState{}
	return state
}

func (self *BlockChain) OnShardStateReceived(state *ShardChainState) {

}

func (self *BlockChain) OnShardToBeaconBlockReceived(block ShardToBeaconBlock) {
	self.config.ShardToBeaconPool.AddShardBeaconBlock(block)
}

func (self *BlockChain) OnCrossShardBlockReceived(block CrossShardBlock) {
	self.config.CrossShardPool.AddCrossShardBlock(block)

}
