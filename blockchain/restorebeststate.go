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
	beaconBestState.Epoch = block.GetCurrentEpoch()
	beaconBestState.BestBlockHash = *block.Hash()
	beaconBestState.PreviousBestBlockHash = block.PreviousBestBlockHash()
	beaconBestState.initCommitteeEngine(blockchain)

	if blockchain.BeaconChain.GetBestView() != nil || beaconBestState.BeaconHeight == 1 {
		err = initMissingSignatureCounter(blockchain, beaconBestState, block)
		if err != nil {
			return err
		}
	}
	return nil
}
