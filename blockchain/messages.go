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
