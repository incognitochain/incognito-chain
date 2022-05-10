package blockchain

import (
	"fmt"

	"github.com/incognitochain/incognito-chain/blockchain/bridgeagg"
	"github.com/incognitochain/incognito-chain/blockchain/pdex"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/config"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/incdb"
)

//RestoreBeaconViewStateFromHash ...
func (beaconBestState *BeaconBestState) RestoreBeaconViewStateFromHash(
	blockchain *BlockChain, includeCommittee, includePdexv3, includeBridgeAgg bool,
) error {
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
	newMaxCommitteeSize := GetMaxCommitteeSize(beaconBestState.MaxShardCommitteeSize,
		beaconBestState.TriggeredFeature, block.Header.Height)
	if newMaxCommitteeSize != beaconBestState.MaxShardCommitteeSize {
		Logger.log.Infof("Beacon Height %+v, Hash %+v, found new max committee size %+v", block.Header.Height, block.Header.Hash(), newMaxCommitteeSize)
		beaconBestState.MaxShardCommitteeSize = newMaxCommitteeSize
	}
	if includeCommittee {
		err := beaconBestState.restoreCommitteeState(blockchain)
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
	}

	if includePdexv3 {
		beaconBestState.pdeStates = make(map[uint]pdex.State)
		beaconViewCached, ok := blockchain.beaconViewCache.Get(beaconBestState.BestBlockHash.String())
		if !ok || beaconViewCached == nil {
			state, err := pdex.InitStateFromDB(beaconBestState.GetBeaconFeatureStateDB(), beaconBestState.BeaconHeight, pdex.AmplifierVersion)
			if err != nil {
				return err
			}
			beaconBestState.pdeStates[pdex.AmplifierVersion] = state
		} else {
			beaconBestState.pdeStates = beaconViewCached.(*BeaconBestState).pdeStates
		}
	}
	if includeBridgeAgg {
		beaconBestState.bridgeAggState, err = bridgeagg.InitStateFromDB(beaconBestState.featureStateDB)
	}
	return err
}

func (blockchain *BlockChain) GetPdexv3Cached(blockHash common.Hash) interface{} {
	beaconViewCached, ok := blockchain.beaconViewCache.Get(blockHash.String())
	if !ok || beaconViewCached == nil {
		return nil
	}
	return beaconViewCached.(*BeaconBestState).pdeStates[pdex.AmplifierVersion]
}

func (beaconBestState *BeaconBestState) IsValidPoolPairID(poolPairID string) error {
	return beaconBestState.pdeStates[pdex.AmplifierVersion].Validator().IsValidPoolPairID(poolPairID)
}

func (beaconBestState *BeaconBestState) IsValidNftID(db incdb.Database, pdexv3StateCached interface{}, nftID string) error {
	state, ok := pdexv3StateCached.(pdex.State)
	if ok && state != nil {
		return state.Validator().IsValidNftID(nftID)
	}
	if !ok || state == nil {
		if beaconBestState.pdeStates[pdex.AmplifierVersion] != nil {
			err := beaconBestState.pdeStates[pdex.AmplifierVersion].Validator().IsValidNftID(nftID)
			if err == nil {
				return nil
			}
		}
		beaconBestState.pdeStates[pdex.AmplifierVersion] = pdex.NewStatev2()
		var dbAccessWarper = statedb.NewDatabaseAccessWarper(db)
		sDB, err := statedb.NewWithPrefixTrie(beaconBestState.FeatureStateDBRootHash, dbAccessWarper)
		if err != nil {
			return err
		}

	}
	return fmt.Errorf("Cannot recognize pdex cache format")
}

func (beaconBestState *BeaconBestState) IsValidMintNftRequireAmount(pdexv3StateCached interface{}, amount uint64) error {
	state, ok := pdexv3StateCached.(pdex.State)
	if ok && state != nil {
		return state.Validator().IsValidMintNftRequireAmount(amount)
	}
	return fmt.Errorf("Cannot recognize pdex cache format")
}

func (beaconBestState *BeaconBestState) IsValidPdexv3StakingPool(tokenID string) error {
	return beaconBestState.pdeStates[pdex.AmplifierVersion].Validator().IsValidStakingPool(tokenID)
}

func (beaconBestState *BeaconBestState) IsValidPdexv3UnstakingAmount(
	tokenID, nftID string, unstakingAmount uint64,
) error {
	return beaconBestState.pdeStates[pdex.AmplifierVersion].Validator().IsValidUnstakingAmount(tokenID, nftID, unstakingAmount)
}

func (beaconBestState *BeaconBestState) IsValidPdexv3ShareAmount(
	poolPairID, nftID string, shareAmount uint64,
) error {
	return beaconBestState.pdeStates[pdex.AmplifierVersion].Validator().IsValidShareAmount(poolPairID, nftID, shareAmount)
}
