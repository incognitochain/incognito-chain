package blockchain

import (
	"errors"
	"log"
	"math/big"
	"sort"

	"github.com/incognitochain/incognito-chain/blockchain/committeestate"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/config"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/instruction"
)

func (bestView *BeaconBestState) CalculateDelegationSharePrice(bc *BlockChain, delegationReward uint64) ([][]string, error) {
	if bestView.TriggeredFeature[BEACON_STAKING_FLOW_V4] == 0 {
		return nil, nil
	}

	//check if end of epoch
	if bestView.GetHeight()%config.Param().EpochParam.NumberOfBlockInEpoch != 0 {
		return nil, errors.New("Not new epoch")
	}

	beaconCommitteeReward := map[string]uint64{}
	totalDelegationAmount := uint64(0)

	//get reward with performance
	stateDB := bestView.GetBeaconConsensusStateDB()

	committeeData := statedb.GetCommitteeData(stateDB)
	oldPrice := map[string]uint64{}

	oldPriceDelegationAmount := map[string]uint64{}
	validators := []string{}
	for k, _ := range committeeData.LastEpoch {
		validators = append(validators, k)
	}
	sort.Slice(validators, func(i, j int) bool {
		return validators[i] > validators[j]
	})

	//get total delegation of this epoch
	for k, v := range committeeData.LastEpoch {
		stakeID := v.BeaconStakeID
		sharePrice, _, _ := statedb.GetBeaconSharePrice(stateDB, stakeID)
		if sharePrice == nil || sharePrice.GetPrice() == 0 {
			return nil, errors.New("cannot find share price of beacon " + k + " ")
		}
		oldPrice[k] = sharePrice.GetPrice()
		oldPriceDelegationAmount[k] = v.Delegators * common.SHARD_STAKING_AMOUNT
		totalDelegationAmount += oldPriceDelegationAmount[k]
	}

	if totalDelegationAmount == 0 {
		log.Println("No delegation to reward!")
		return nil, nil
	}

	//get beacon delegation reward with performance
	for k, v := range committeeData.LastEpoch {
		a := new(big.Int).SetUint64(delegationReward)
		b := new(big.Int).SetUint64(oldPriceDelegationAmount[k])
		c := new(big.Int).SetUint64(totalDelegationAmount)
		tmp := new(big.Int).Div(new(big.Int).Mul(a, b), c).Uint64()
		beaconCommitteeReward[k], _ = new(big.Float).Mul(
			new(big.Float).SetUint64(tmp),
			new(big.Float).SetFloat64((float64(v.Performance) / float64(bestView.GetBeaconCommitteeState().(*committeestate.BeaconCommitteeStateV4).GetConfig().MAX_SCORE)))).Uint64()
	}

	//increase share price
	sharePriceInsts := instruction.NewSharePriceInstruction()
	for _, cpkStr := range validators {
		if committeeData.LastEpoch[cpkStr].Delegators == 0 {
			log.Println("No delegation reward for", cpkStr)
			continue
		}
		price := new(big.Int).SetUint64(oldPrice[cpkStr])
		c := new(big.Int).SetUint64(oldPriceDelegationAmount[cpkStr])
		b := new(big.Int).SetUint64(beaconCommitteeReward[cpkStr])
		newprice := new(big.Int).Div(
			new(big.Int).Mul(
				price,
				new(big.Int).Add(b, c)),
			c).Uint64()
		stakeID := committeeData.LastEpoch[cpkStr].BeaconStakeID
		sharePriceInsts.AddPrice(stakeID, newprice)
		log.Println(stakeID, beaconCommitteeReward[cpkStr], price.Uint64(), newprice)
	}
	log.Println(sharePriceInsts.ToString())
	return [][]string{sharePriceInsts.ToString()}, nil
}

func (bestView *BeaconBestState) GetCommitteeData(bc *BlockChain, delegationReward uint64) (*statedb.CommitteeData, error) {
	if bestView.TriggeredFeature[BEACON_STAKING_FLOW_V4] == 0 {
		return nil, nil
	}

	//get reward with performance
	stateDB := bestView.GetBeaconConsensusStateDB()

	committeeData := statedb.GetCommitteeData(stateDB)

	return committeeData, nil
}
