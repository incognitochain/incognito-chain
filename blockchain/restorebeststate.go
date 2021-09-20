package blockchain

import (
	"github.com/incognitochain/incognito-chain/blockchain/pdex"
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

	//TODO: @tin optimize here later by includePdex bool
	beaconBestState.pdeStates, err = pdex.InitStatesFromDB(beaconBestState.featureStateDB, beaconBestState.BeaconHeight)
	if err != nil {
		return err
	}
	if beaconBestState.BeaconHeight > config.Param().ConsensusParam.BlockProducingV3Height {
		if err := beaconBestState.checkBlockProducingV3Config(); err != nil {
			return err
		}
		if err := beaconBestState.upgradeBlockProducingV3Config(); err != nil {
			return err
		}
	}
	return nil
}
