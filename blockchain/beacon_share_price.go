package blockchain

import (
	"errors"
	"fmt"
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
		a := new(big.Float).SetUint64(delegationReward)
		b := new(big.Float).SetUint64(oldPriceDelegationAmount[k])
		c := new(big.Float).SetUint64(totalDelegationAmount)
		p := new(big.Float).SetUint64(v.Performance)
		maxScore := new(big.Float).SetUint64(bestView.GetBeaconCommitteeState().(*committeestate.BeaconCommitteeStateV4).GetConfig().MAX_SCORE)
		tmp := new(big.Float).Mul(new(big.Float).Mul(a, b), p)
		fmt.Printf("tmp %v c*maxscore %v\n", tmp.String(), new(big.Float).Mul(c, maxScore).String())
		beaconCommitteeReward[k], _ = new(big.Float).Quo(tmp, new(big.Float).Mul(c, maxScore)).Uint64()
		aU, _ := a.Uint64()
		bU, _ := b.Uint64()
		pU, _ := p.Uint64()
		cU, _ := c.Uint64()
		fmt.Printf("Value to calculate delegation reward: total %v, oldPriceDAmount %v performance %v c %v\n", aU, bU, pU, cU)
		fmt.Printf("beaconCommitteeReward %v\n", beaconCommitteeReward[k])
	}

	//increase share price
	sharePriceInsts := instruction.NewSharePriceInstruction()
	for _, cpkStr := range validators {
		if committeeData.LastEpoch[cpkStr].Delegators == 0 {
			log.Println("No delegation reward for", cpkStr)
			continue
		}
		price := new(big.Float).SetUint64(oldPrice[cpkStr])
		c := new(big.Float).SetUint64(oldPriceDelegationAmount[cpkStr])
		b := new(big.Float).SetUint64(beaconCommitteeReward[cpkStr])
		newprice := new(big.Float).Quo(
			new(big.Float).Mul(
				price,
				b,
			),
			c,
		)
		newprice = new(big.Float).Add(newprice, price)
		stakeID := committeeData.LastEpoch[cpkStr].BeaconStakeID
		newPrU, _ := newprice.Uint64()
		prU, _ := price.Uint64()
		sharePriceInsts.AddPrice(stakeID, newPrU)
		fmt.Println("New price ", stakeID, beaconCommitteeReward[cpkStr], prU, newPrU)
	}
	fmt.Println("New Price Insts ", sharePriceInsts.ToString())
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
