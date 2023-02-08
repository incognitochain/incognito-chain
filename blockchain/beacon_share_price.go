package blockchain

import (
	"errors"
	"github.com/incognitochain/incognito-chain/blockchain/committeestate"
	"github.com/incognitochain/incognito-chain/config"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/instruction"
)

//todo: big number calculation
func (bestView *BeaconBestState) CalculateDelegationSharePrice(bc *BlockChain, delegationReward uint64) ([][]string, error) {
	//check if start of epoch
	if bestView.GetHeight()%config.Param().EpochParam.NumberOfBlockInEpoch != 1 {
		return nil, errors.New("Not new epoch")
	}

	beaconCommitteeReward := map[string]uint64{}
	totalDelegationAmount := uint64(0)

	//get reward with performance
	beaconConsensusStateRootHash, err := bc.GetBeaconRootsHashFromBlockHeight(
		bestView.GetHeight() - 1,
	)
	if err != nil {
		return nil, err
	}
	stateDB, err := statedb.NewWithPrefixTrie(beaconConsensusStateRootHash.ConsensusStateDBRootHash,
		statedb.NewDatabaseAccessWarper(bc.GetBeaconChainDatabase()))
	if err != nil {
		return nil, err
	}

	committeeData := statedb.GetCommitteeData(stateDB)
	oldPrice := map[string]uint64{}

	oldPriceDelegationAmount := map[string]uint64{}
	//get total delegation of this epoch
	for k, v := range committeeData.LastCommitteeEpochInfo {
		stakeID := v.BeaconStakeID
		sharePrice, _, _ := statedb.GetBeaconSharePrice(stateDB, stakeID)
		if sharePrice == nil {
			return nil, errors.New("cannot find share price of beacon " + k + " " + stakeID)
		}
		oldPrice[k] = sharePrice.GetPrice()
		oldPriceDelegationAmount[k] = v.DelegationAmount
		totalDelegationAmount += oldPriceDelegationAmount[k]
	}

	//get beacon delegation reward with performance
	for k, v := range committeeData.LastCommitteeEpochInfo {
		beaconCommitteeReward[k] = delegationReward * oldPriceDelegationAmount[k] / totalDelegationAmount
		beaconCommitteeReward[k] = uint64(float64(beaconCommitteeReward[k]) *
			(float64(v.Performance) / float64(bestView.GetBeaconCommitteeState().(*committeestate.BeaconCommitteeStateV4).GetConfig().MAX_SCORE)))
	}

	//increase share price
	sharePriceInsts := instruction.NewSharePriceInstruction()
	for cpkStr, v := range committeeData.LastCommitteeEpochInfo {
		price := oldPriceDelegationAmount[cpkStr]
		price = price * (beaconCommitteeReward[cpkStr] + oldPriceDelegationAmount[cpkStr]) / (oldPriceDelegationAmount[cpkStr])
		stakeID := v.BeaconStakeID
		sharePriceInsts.AddPrice(stakeID, price)
	}

	return [][]string{sharePriceInsts.ToString()}, nil

}
