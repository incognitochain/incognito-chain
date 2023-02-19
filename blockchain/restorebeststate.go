package blockchain

import (
	"fmt"

	"github.com/incognitochain/incognito-chain/blockchain/bridgeagg"
	"github.com/incognitochain/incognito-chain/blockchain/bridgehub"
	"github.com/incognitochain/incognito-chain/blockchain/pdex"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/config"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/incdb"
)

//RestoreBeaconViewStateFromHash ...
func (beaconBestState *BeaconBestState) RestoreBeaconViewStateFromHash(
	blockchain *BlockChain, includeCommittee, includePdexv3, includeBridgeAgg bool, includeBridgeHub bool,
) error {
	Logger.log.Infof("Start restore beaconBestState")
	err := beaconBestState.InitStateRootHash(blockchain)
	if err != nil {
		return err
	}
	//best block
	block, _, err := blockchain.GetBeaconBlockByHashWithLatestValidationData(beaconBestState.BestBlockHash)
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
		Logger.log.Infof("Start restore pdexv3 state")
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
		Logger.log.Infof("Finish restore pdexv3 state")
	}
	if includeBridgeAgg {
		beaconBestState.bridgeAggManager, err = bridgeagg.InitManager(beaconBestState.featureStateDB)
	}
	if includeBridgeHub {
		beaconBestState.bridgeHubManager, err = bridgehub.InitManager(beaconBestState.featureStateDB)
	}
	Logger.log.Infof("Finish restore beaconBestState")
	return err
}

func (blockchain *BlockChain) GetPdexv3Cached(blockHash common.Hash) interface{} {
	beaconViewCached, ok := blockchain.beaconViewCache.Get(blockHash.String())
	if !ok || beaconViewCached == nil {
		return nil
	}
	return beaconViewCached.(*BeaconBestState).pdeStates[pdex.AmplifierVersion]
}

func (beaconBestState *BeaconBestState) IsValidNftID(db incdb.Database, pdexv3StateCached interface{}, nftID string) error {
	state, ok := pdexv3StateCached.(pdex.State)
	if ok && state != nil {
		return state.Validator().IsValidNftID(nftID)
	}
	if !ok || state == nil {
		var dbAccessWarper = statedb.NewDatabaseAccessWarper(db)
		sDB, err := statedb.NewWithPrefixTrie(beaconBestState.FeatureStateDBRootHash, dbAccessWarper)
		if err != nil {
			return err
		}
		nftIDState, err := statedb.GetPdexv3NftID(sDB, nftID)
		if err != nil || nftIDState == nil {
			if nftIDState == nil && err != nil {
				err = fmt.Errorf("Not found nftID %s", nftID)
			}
			return err
		}
		return nil
	}
	return fmt.Errorf("Cannot recognize pdex cache format")
}

func (beaconBestState *BeaconBestState) IsValidPoolPairID(db incdb.Database, pdexv3StateCached interface{}, poolPairID string) error {
	state, ok := pdexv3StateCached.(pdex.State)
	if ok && state != nil {
		return state.Validator().IsValidPoolPairID(poolPairID)
	}
	if !ok || state == nil {
		var dbAccessWarper = statedb.NewDatabaseAccessWarper(db)
		sDB, err := statedb.NewWithPrefixTrie(beaconBestState.FeatureStateDBRootHash, dbAccessWarper)
		if err != nil {
			return err
		}
		poolpair, err := statedb.GetPdexv3PoolPair(sDB, poolPairID)
		if err != nil || poolpair == nil {
			if poolpair == nil && err != nil {
				err = fmt.Errorf("Not found poolPairID %s", poolPairID)
			}
			return err
		}
		return nil
	}
	return fmt.Errorf("Cannot recognize pdex cache format")

}

func (beaconBestState *BeaconBestState) IsValidMintNftRequireAmount(db incdb.Database, pdexv3StateCached interface{}, amount uint64) error {
	state, ok := pdexv3StateCached.(pdex.State)
	if ok && state != nil {
		return state.Validator().IsValidMintNftRequireAmount(amount)
	}
	if !ok || state == nil {
		var dbAccessWarper = statedb.NewDatabaseAccessWarper(db)
		sDB, err := statedb.NewWithPrefixTrie(beaconBestState.FeatureStateDBRootHash, dbAccessWarper)
		if err != nil {
			return err
		}
		params, err := statedb.GetPdexv3Params(sDB)
		if err != nil {
			return err
		}
		if params.MintNftRequireAmount() != amount {
			return fmt.Errorf("Expect mint nft amount to be %v but get %v", params.MintNftRequireAmount(), amount)
		}
		return nil
	}
	return fmt.Errorf("Cannot recognize pdex cache format")
}

func (beaconBestState *BeaconBestState) IsValidPdexv3StakingPool(db incdb.Database, pdexv3StateCached interface{}, tokenID string) error {
	state, ok := pdexv3StateCached.(pdex.State)
	if ok && state != nil {
		return state.Validator().IsValidStakingPool(tokenID)
	}
	if !ok || state == nil {
		var dbAccessWarper = statedb.NewDatabaseAccessWarper(db)
		sDB, err := statedb.NewWithPrefixTrie(beaconBestState.FeatureStateDBRootHash, dbAccessWarper)
		if err != nil {
			return err
		}
		params, err := statedb.GetPdexv3Params(sDB)
		if err != nil || params == nil {
			if params == nil && err != nil {
				err = fmt.Errorf("Not found paramss")
			}
			return err
		}
		if _, found := params.StakingPoolsShare()[tokenID]; !found {
			return fmt.Errorf("Not found stakingPoolID %s", tokenID)
		}
		return nil
	}
	return fmt.Errorf("Cannot recognize pdex cache format")
}

func (beaconBestState *BeaconBestState) IsValidPdexv3UnstakingAmount(db incdb.Database, pdexv3StateCached interface{}, tokenID, nftID string, unstakingAmount uint64) error {
	state, ok := pdexv3StateCached.(pdex.State)
	if ok && state != nil {
		return state.Validator().IsValidUnstakingAmount(tokenID, nftID, unstakingAmount)
	}
	if !ok || state == nil {
		var dbAccessWarper = statedb.NewDatabaseAccessWarper(db)
		sDB, err := statedb.NewWithPrefixTrie(beaconBestState.FeatureStateDBRootHash, dbAccessWarper)
		if err != nil {
			return err
		}
		params, err := statedb.GetPdexv3Params(sDB)
		if err != nil || params == nil {
			if params == nil && err != nil {
				err = fmt.Errorf("Not found paramss")
			}
			return err
		}
		if _, found := params.StakingPoolsShare()[tokenID]; !found {
			return fmt.Errorf("Not found stakingPoolID %s", tokenID)
		}
		staker, err := statedb.GetPdexv3Staker(sDB, tokenID, nftID)
		if err != nil {
			return err
		}
		if staker.Liquidity() < unstakingAmount {
			return fmt.Errorf("unstakingAmount > current staker liquidity")
		}
		if staker.Liquidity() == 0 || unstakingAmount == 0 {
			return fmt.Errorf("unstakingAmount or staker.Liquidity is 0")
		}
		return nil
	}
	return fmt.Errorf("Cannot recognize pdex cache format")

}

func (beaconBestState *BeaconBestState) IsValidPdexv3ShareAmount(db incdb.Database, pdexv3StateCached interface{}, poolPairID, nftID string, shareAmount uint64) error {
	state, ok := pdexv3StateCached.(pdex.State)
	if ok && state != nil {
		return state.Validator().IsValidShareAmount(poolPairID, nftID, shareAmount)
	}
	if !ok || state == nil {
		var dbAccessWarper = statedb.NewDatabaseAccessWarper(db)
		sDB, err := statedb.NewWithPrefixTrie(beaconBestState.FeatureStateDBRootHash, dbAccessWarper)
		if err != nil {
			return err
		}
		_, err = statedb.GetPdexv3PoolPair(sDB, poolPairID)
		if err != nil {
			return err
		}
		share, err := statedb.GetPdexv3Share(sDB, poolPairID, nftID)
		if err != nil {
			return err
		}
		if share.Amount() < shareAmount {
			return fmt.Errorf("shareAmount > current share amount")
		}
		if shareAmount == 0 || share.Amount() == 0 {
			return fmt.Errorf("share amount or share.Amount() is 0")
		}
		return nil
	}
	return fmt.Errorf("Cannot recognize pdex cache format")
}
