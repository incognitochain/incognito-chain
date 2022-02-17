package pdex

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"reflect"

	"github.com/incognitochain/incognito-chain/blockchain/pdex/v2utils"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
)

type StakingPoolState struct {
	liquidity       uint64
	stakers         map[string]*Staker // nft -> amount staking
	rewardsPerShare map[common.Hash]*big.Int
}

func NewStakingPoolState() *StakingPoolState {
	return &StakingPoolState{
		stakers:         make(map[string]*Staker),
		rewardsPerShare: make(map[common.Hash]*big.Int),
	}
}

func NewStakingPoolStateWithValue(
	liquidity uint64,
	stakers map[string]*Staker,
	rewardsPerShare map[common.Hash]*big.Int,
) *StakingPoolState {
	return &StakingPoolState{
		liquidity:       liquidity,
		stakers:         stakers,
		rewardsPerShare: rewardsPerShare,
	}
}

func (stakingPoolState *StakingPoolState) Liquidity() uint64 {
	return stakingPoolState.liquidity
}

func (stakingPoolState *StakingPoolState) Stakers() map[string]*Staker {
	res := make(map[string]*Staker)
	for k, v := range stakingPoolState.stakers {
		res[k] = v.Clone()
	}
	return res
}

func (stakingPoolState *StakingPoolState) RewardsPerShare() map[common.Hash]*big.Int {
	res := make(map[common.Hash]*big.Int)
	for k, v := range stakingPoolState.rewardsPerShare {
		res[k] = new(big.Int).Set(v)
	}
	return res
}

func (stakingPoolState *StakingPoolState) SetRewardsPerShare(rewardsPerShare map[common.Hash]*big.Int) {
	stakingPoolState.rewardsPerShare = rewardsPerShare
}

func (stakingPoolState *StakingPoolState) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(struct {
		Liquidity       uint64                   `json:"Liquidity"`
		Stakers         map[string]*Staker       `json:"Stakers,omitempty"`
		RewardsPerShare map[common.Hash]*big.Int `json:"RewardsPerShare"`
	}{
		Liquidity:       stakingPoolState.liquidity,
		Stakers:         stakingPoolState.stakers,
		RewardsPerShare: stakingPoolState.rewardsPerShare,
	})
	if err != nil {
		return []byte{}, err
	}
	return data, nil
}

func (stakingPoolState *StakingPoolState) UnmarshalJSON(data []byte) error {
	temp := struct {
		Liquidity uint64                   `json:"Liquidity"`
		Stakers   map[string]*Staker       `json:"Stakers"`
		Rewards   map[common.Hash]*big.Int `json:"Rewards"`
	}{}
	err := json.Unmarshal(data, &temp)
	if err != nil {
		return err
	}
	stakingPoolState.liquidity = temp.Liquidity
	stakingPoolState.stakers = temp.Stakers
	stakingPoolState.rewardsPerShare = temp.Rewards
	return nil
}

func (s *StakingPoolState) Clone() *StakingPoolState {
	res := NewStakingPoolState()
	res.liquidity = s.liquidity
	for k, v := range s.stakers {
		res.stakers[k] = v.Clone()
	}
	for k, v := range s.rewardsPerShare {
		res.rewardsPerShare[k] = new(big.Int).Set(v)
	}
	return res
}

func (s *StakingPoolState) getDiff(
	stakingPoolID string, compareStakingPoolState *StakingPoolState,
	stakingPoolChange *v2utils.StakingPoolChange,
) *v2utils.StakingPoolChange {
	newStakingPoolChange := stakingPoolChange
	if compareStakingPoolState == nil {
		for tokenID := range s.rewardsPerShare {
			newStakingPoolChange.RewardsPerShare[tokenID.String()] = true
		}
		for nftID, staker := range s.stakers {
			stakerChange := v2utils.NewStakerChange()
			stakerChange = staker.getDiff(nil, stakerChange)
			stakingPoolChange.Stakers[nftID] = stakerChange
		}
	} else {
		newStakingPoolChange.RewardsPerShare = v2utils.DifMapHashBigInt(s.rewardsPerShare).GetDiff(v2utils.DifMapHashBigInt(compareStakingPoolState.rewardsPerShare))
		for nftID, staker := range s.stakers {
			if m, ok := compareStakingPoolState.stakers[nftID]; !ok || !reflect.DeepEqual(m, staker) {
				stakerChange := v2utils.NewStakerChange()
				stakerChange = staker.getDiff(m, stakerChange)
				stakingPoolChange.Stakers[nftID] = stakerChange
			}
		}
	}
	return newStakingPoolChange
}

func (s *StakingPoolState) updateLiquidity(nftID string, liquidity, beaconHeight uint64, operator byte) error {
	staker, found := s.stakers[nftID]
	if !found {
		if operator == subOperator {
			return errors.New("remove liquidity from invalid staker")
		}
		s.stakers[nftID] = NewStakerWithValue(liquidity, make(map[common.Hash]uint64), s.RewardsPerShare())
	} else {
		tempLiquidity, err := executeOperationUint64(staker.liquidity, liquidity, operator)
		if err != nil {
			return err
		}
		nftIDHash, err := new(common.Hash).NewHashFromStr(nftID)
		if err != nil {
			return fmt.Errorf("Invalid nftID: %v", nftID)
		}
		staker.rewards, err = s.RecomputeStakingRewards(*nftIDHash)
		if err != nil {
			return fmt.Errorf("Recompute staking rewards failed: %v", err)
		}
		staker.lastRewardsPerShare = s.RewardsPerShare()
		staker.liquidity = tempLiquidity
	}
	var tempLiquidity uint64
	switch operator {
	case subOperator:
		tempLiquidity = s.liquidity - liquidity
		if tempLiquidity >= s.liquidity {
			return errors.New("Staking pool liquidity is out of range")
		}
	case addOperator:
		tempLiquidity = s.liquidity + liquidity
		if tempLiquidity < s.liquidity {
			return errors.New("Staking pool liquidity is out of range")
		}
	}
	s.liquidity = tempLiquidity
	return nil
}

func (s *StakingPoolState) withLiquidity(liquidity uint64) {
	s.liquidity = liquidity
}

func (s *StakingPoolState) withStakers(stakers map[string]*Staker) {
	s.stakers = stakers
}

func (s *StakingPoolState) withRewardsPerShare(rewardsPerShare map[common.Hash]*big.Int) {
	s.rewardsPerShare = rewardsPerShare
}

func (s *StakingPoolState) cloneStaker(nftID string) map[string]*Staker {
	res := make(map[string]*Staker)
	for k, v := range s.stakers {
		if k == nftID {
			res[k] = v.Clone()
		} else {
			res[k] = v
		}
	}
	return res
}

func (s *StakingPoolState) RecomputeStakingRewards(
	nftID common.Hash,
) (map[common.Hash]uint64, error) {
	result := map[common.Hash]uint64{}

	curStaker, ok := s.stakers[nftID.String()]
	if !ok {
		return nil, fmt.Errorf("Stakers not found")
	}

	curLPFeesPerShare := s.RewardsPerShare()
	oldLPFeesPerShare := curStaker.LastRewardsPerShare()

	for tokenID := range curLPFeesPerShare {
		tradingFee, isExisted := curStaker.rewards[tokenID]
		if !isExisted {
			tradingFee = 0
		}
		oldFees, isExisted := oldLPFeesPerShare[tokenID]
		if !isExisted {
			oldFees = big.NewInt(0)
		}
		newFees := curLPFeesPerShare[tokenID]

		reward := new(big.Int).Mul(new(big.Int).Sub(newFees, oldFees), new(big.Int).SetUint64(curStaker.liquidity))
		reward = new(big.Int).Div(reward, BaseLPFeesPerShare)
		reward = new(big.Int).Add(reward, new(big.Int).SetUint64(tradingFee))

		if !reward.IsUint64() {
			return nil, fmt.Errorf("Reward of token %v is out of range", tokenID)
		}
		if reward.Uint64() > 0 {
			result[tokenID] = reward.Uint64()
		}
	}
	return result, nil
}

func (s *StakingPoolState) AddReward(
	tokenID common.Hash, amount uint64,
) {
	if s.Liquidity() == 0 {
		return
	}

	oldRewardsPerShare, isExisted := s.RewardsPerShare()[tokenID]
	if !isExisted {
		oldRewardsPerShare = big.NewInt(0)
	}

	// delta (reward / total share) = reward * BASE / totalShare
	deltaRewardsPerShare := new(big.Int).Mul(new(big.Int).SetUint64(amount), BaseLPFeesPerShare)
	deltaRewardsPerShare = new(big.Int).Div(deltaRewardsPerShare, new(big.Int).SetUint64(s.Liquidity()))

	// update accumulated sum of (fee / LP share)
	newLPFeesPerShare := new(big.Int).Add(oldRewardsPerShare, deltaRewardsPerShare)
	tempRewardsPerShare := s.RewardsPerShare()
	tempRewardsPerShare[tokenID] = newLPFeesPerShare

	s.SetRewardsPerShare(tempRewardsPerShare)
}

func (s *StakingPoolState) updateToDB(
	env StateEnvironment, stakingPoolID string, stakingPoolChange *v2utils.StakingPoolChange,
) error {
	for tokenID, value := range s.rewardsPerShare {
		if stakingPoolChange.RewardsPerShare[tokenID.String()] {
			err := statedb.StorePdexv3StakingPoolRewardPerShare(
				env.StateDB(), stakingPoolID,
				statedb.NewPdexv3StakingPoolRewardPerShareStateWithValue(tokenID, value),
			)
			if err != nil {
				return err
			}
		}
	}

	for nftID, stakerChange := range stakingPoolChange.Stakers {
		if stakerChange.IsChanged {
			nftHash, err := common.Hash{}.NewHashFromStr(nftID)
			if err != nil {
				return err
			}
			if staker, found := s.stakers[nftID]; found {
				if stakerChange.IsChanged {
					err := statedb.StorePdexv3Staker(env.StateDB(), stakingPoolID, nftID,
						statedb.NewPdexv3StakerStateWithValue(*nftHash, staker.liquidity),
					)
					if err != nil {
						return err
					}
				}
				for tokenID, value := range staker.lastRewardsPerShare {
					if stakerChange.LastRewardsPerShare[tokenID.String()] {
						err := statedb.StorePdexv3StakerLastRewardPerShare(
							env.StateDB(), stakingPoolID, nftID,
							statedb.NewPdexv3StakerLastRewardPerShareStateWithValue(tokenID, value),
						)
						if err != nil {
							return err
						}
					}
				}
				for tokenID, value := range staker.rewards {
					if stakerChange.Rewards[tokenID.String()] {
						err := statedb.StorePdexv3StakerReward(
							env.StateDB(), stakingPoolID, nftID,
							statedb.NewPdexv3StakerRewardStateWithValue(tokenID, value),
						)
						if err != nil {
							return err
						}
					}
				}
			} else {
				err = statedb.DeletePdexv3Staker(env.StateDB(), stakingPoolID, *nftHash)
				if err != nil {
					return err
				}
				for tokenID, _ := range staker.lastRewardsPerShare {
					err = statedb.DeletePdexv3StakerLastRewardPerShare(env.StateDB(), stakingPoolID, nftID, tokenID.String())
					if err != nil {
						return err
					}
				}
				for tokenID, _ := range staker.rewards {
					err = statedb.DeletePdexv3StakerReward(env.StateDB(), stakingPoolID, nftID, tokenID.String())
					if err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
}
