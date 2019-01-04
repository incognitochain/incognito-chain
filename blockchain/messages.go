package blockchain

func (self *BlockChain) OnBlockShardReceived(block *ShardBlock) {
	if self.newShardBlkCh != nil {

	}
}
func (self *BlockChain) OnBlockBeaconReceived(block *BeaconBlock) {
	if self.newBeaconBlkCh != nil {
		self.newBeaconBlkCh <- block
	}
}

func (self *BlockChain) OnGetBeaconState() *BeaconChainState {
	state := &BeaconChainState{}
	return state
}

func (self *BlockChain) OnBeaconStateReceived(state *BeaconChainState) {

}

func (self *BlockChain) OnGetShardState(shardID byte) *ShardChainState {
	state := &ShardChainState{}
	return state
}

func (self *BlockChain) OnShardStateReceived(state *ShardChainState) {

}
