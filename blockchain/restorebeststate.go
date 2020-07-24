package blockchain

//RestoreBeaconViewStateFromHash ...
func (beaconBestState *BeaconBestState) RestoreBeaconViewStateFromHash(blockchain *BlockChain) error {
	err := beaconBestState.InitStateRootHash(blockchain)
	if err != nil {
		return err
	}
	//best block
	block, _, err := blockchain.GetBeaconBlockByHash(beaconBestState.BestBlockHash)
	if err != nil || block == nil {
		return err
	}
	beaconBestState.BestBlock = *block
	beaconBestState.BeaconHeight = block.GetHeight()
	beaconCommitteeEngine := InitBeaconCommitteeEngineV1(beaconBestState.ActiveShards, beaconBestState.consensusStateDB, beaconBestState.BeaconHeight, beaconBestState.BestBlockHash)
	beaconBestState.beaconCommitteeEngine = beaconCommitteeEngine
	return nil
}
