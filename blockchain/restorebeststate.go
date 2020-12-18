package blockchain

import "github.com/incognitochain/incognito-chain/blockchain/committeestate"

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
	var beaconCommitteeEngine committeestate.BeaconCommitteeEngine
	if beaconBestState.BeaconHeight >= blockchain.config.ChainParams.ConsensusV3Height {
		beaconCommitteeEngine = initBeaconCommitteeEngineV2(
			beaconBestState,
			blockchain.config.ChainParams,
			blockchain,
		)
	} else {
		beaconCommitteeEngine = initBeaconCommitteeEngineV1(
			beaconBestState,
		)
	}
	beaconBestState.beaconCommitteeEngine = beaconCommitteeEngine
	if blockchain.BeaconChain.GetBestView() != nil {
		err = initMissingSignatureCounter(blockchain, beaconBestState, block)
		if err != nil {
			return err
		}
	}
	return nil
}
