package blockchain

//RestoreBeaconViewStateFromHash ...
func (beaconBestState *BeaconBestState) RestoreBeaconViewStateFromHash(blockchain *BlockChain, includeCommittee bool) error {
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

	if includeCommittee {
		beaconBestState.initCommitteeState(blockchain)
		if beaconBestState.BeaconHeight == blockchain.config.ChainParams.StakingFlowV2Height {
			beaconBestState.upgradeCommitteeState(blockchain)
		}
		if blockchain.BeaconChain.GetBestView() != nil {
			err = beaconBestState.initMissingSignatureCounter(blockchain, block)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
