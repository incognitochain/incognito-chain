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
	// @NOTICE: beaconBestState.NumberOfShardBlock this field is initialized with zero value only
	// DO NOT use data beaconBestState.NumberOfShardBlock when init from this process
	beaconBestState.NumberOfShardBlock = make(map[byte]uint, beaconBestState.ActiveShards)
	if includeCommittee {
		var beaconCommitteeEngine committeestate.BeaconCommitteeEngine
		if beaconBestState.BeaconHeight > config.Param().ConsensusParam.StakingFlowV2Height {
			beaconCommitteeEngine = initBeaconCommitteeEngineV2(
				beaconBestState,
				blockchain,
			)
		} else {
			beaconCommitteeEngine = initBeaconCommitteeEngineV1(
				beaconBestState,
			)
		}
		beaconBestState.beaconCommitteeEngine = beaconCommitteeEngine
		beaconBestState.tryUpgradeConsensusRule(block)
		if blockchain.BeaconChain.GetBestView() != nil {
			err = initMissingSignatureCounter(blockchain, beaconBestState, block)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
