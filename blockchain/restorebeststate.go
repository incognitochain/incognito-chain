package blockchain

import (
	"github.com/incognitochain/incognito-chain/blockchain/committeestate"
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
		var beaconCommitteeEngine committeestate.BeaconCommitteeState
		if beaconBestState.BeaconHeight > config.Param().ConsensusParam.StakingFlowV2Height {
			beaconCommitteeEngine = InitBeaconCommitteeEngineV2(
				beaconBestState,
				blockchain,
			)
		} else {
			beaconCommitteeEngine = InitBeaconCommitteeEngineV1(
				beaconBestState,
			)
		}
		beaconBestState.beaconCommitteeState = beaconCommitteeEngine
		if beaconBestState.BeaconHeight == config.Param().ConsensusParam.StakingFlowV2Height {
			beaconBestState.upgradeCommitteeEngineV2(blockchain)
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
