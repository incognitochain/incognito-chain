package blockchain

import (
	"github.com/incognitochain/incognito-chain/blockchain/finishsync"
	"github.com/incognitochain/incognito-chain/config"
)

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
	beaconBestState.PreviousBestBlockHash = block.GetPrevHash()

	if includeCommittee {
		err := beaconBestState.restoreCommitteeState(blockchain)
		if err != nil {
			return err
		}
	}

	if beaconBestState.BeaconHeight > config.Param().ConsensusParam.StakingFlowV3Height {
		if err := beaconBestState.checkAndUpgradeStakingFlowV3Config(); err != nil {
			return err
		}
	}

	beaconBestState.FinishSyncManager = finishsync.NewFinishManagerWithValue(
		beaconBestState.FinishSyncManager.FinishedSyncValidators)

	return nil
}
