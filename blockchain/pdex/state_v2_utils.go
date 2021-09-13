package pdex

import (
	"encoding/json"
	"math/big"
	"reflect"

	"github.com/incognitochain/incognito-chain/blockchain/pdex/v2utils"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
)

type Share struct {
	amount             uint64
	tradingFees        map[common.Hash]uint64
	lastLPFeesPerShare map[common.Hash]*big.Int
}

func (share *Share) Amount() uint64 {
	return share.amount
}

func (share *Share) LastLPFeesPerShare() map[common.Hash]*big.Int {
	return share.lastLPFeesPerShare
}

func (share *Share) TradingFees() map[common.Hash]uint64 {
	res := make(map[common.Hash]uint64)
	for k, v := range share.tradingFees {
		res[k] = v
	}
	return res
}

func NewShare() *Share {
	return &Share{
		amount:             0,
		tradingFees:        map[common.Hash]uint64{},
		lastLPFeesPerShare: map[common.Hash]*big.Int{},
	}
}

func NewShareWithValue(
	amount uint64,
	tradingFees map[common.Hash]uint64,
	lastLPFeesPerShare map[common.Hash]*big.Int,
) *Share {
	return &Share{
		amount:             amount,
		tradingFees:        tradingFees,
		lastLPFeesPerShare: lastLPFeesPerShare,
	}
}

func (share *Share) Clone() *Share {
	res := NewShare()
	res.amount = share.amount
	res.tradingFees = map[common.Hash]uint64{}
	for k, v := range share.tradingFees {
		res.tradingFees[k] = v
	}
	res.lastLPFeesPerShare = map[common.Hash]*big.Int{}
	for k, v := range share.lastLPFeesPerShare {
		res.lastLPFeesPerShare[k] = new(big.Int).Set(v)
	}
	return res
}

func (share *Share) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(struct {
		Amount             uint64                   `json:"Amount"`
		TradingFees        map[common.Hash]uint64   `json:"TradingFees"`
		LastLPFeesPerShare map[common.Hash]*big.Int `json:"LastLPFeesPerShare"`
	}{
		Amount:             share.amount,
		TradingFees:        share.tradingFees,
		LastLPFeesPerShare: share.lastLPFeesPerShare,
	})
	if err != nil {
		return []byte{}, err
	}
	return data, nil
}

func (share *Share) UnmarshalJSON(data []byte) error {
	temp := struct {
		Amount             uint64                   `json:"Amount"`
		TradingFees        map[common.Hash]uint64   `json:"TradingFees"`
		LastLPFeesPerShare map[common.Hash]*big.Int `json:"LastLPFeesPerShare"`
	}{}
	err := json.Unmarshal(data, &temp)
	if err != nil {
		return err
	}
	share.amount = temp.Amount
	share.tradingFees = temp.TradingFees
	share.lastLPFeesPerShare = temp.LastLPFeesPerShare
	return nil
}

func (share *Share) getDiff(
	nftID string,
	compareShare *Share,
	stateChange *v2utils.StateChange,
) *v2utils.StateChange {
	newStateChange := stateChange
	if compareShare == nil || !reflect.DeepEqual(share, compareShare) {
		//newStateChange.shares[nftID].IsChanged = true
	}
	return newStateChange
}

type Staker struct {
	liquidity           uint64
	rewards             map[common.Hash]uint64
	lastRewardsPerShare map[common.Hash]*big.Int
}

func (staker *Staker) Liquidity() uint64 {
	return staker.liquidity
}

func (staker *Staker) LastRewardsPerShare() map[common.Hash]*big.Int {
	return staker.lastRewardsPerShare
}

func (staker *Staker) Rewards() map[common.Hash]uint64 {
	res := make(map[common.Hash]uint64)
	for k, v := range staker.rewards {
		res[k] = v
	}
	return res
}

func (staker *Staker) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(struct {
		Liquidity           uint64                   `json:"Liquidity"`
		Rewards             map[common.Hash]uint64   `json:"Rewards"`
		LastRewardsPerShare map[common.Hash]*big.Int `json:"LastLPFeesPerShare"`
	}{
		Liquidity:           staker.liquidity,
		Rewards:             staker.rewards,
		LastRewardsPerShare: staker.lastRewardsPerShare,
	})
	if err != nil {
		return []byte{}, err
	}
	return data, nil
}

func (staker *Staker) UnmarshalJSON(data []byte) error {
	temp := struct {
		Liquidity          uint64                   `json:"Liquidity"`
		Rewards            map[common.Hash]uint64   `json:"Rewards"`
		LastLPFeesPerShare map[common.Hash]*big.Int `json:"LastLPFeesPerShare"`
	}{}
	err := json.Unmarshal(data, &temp)
	if err != nil {
		return err
	}
	staker.liquidity = temp.Liquidity
	staker.rewards = temp.Rewards
	staker.lastRewardsPerShare = temp.LastLPFeesPerShare
	return nil
}

func NewStaker() *Staker {
	return &Staker{
		rewards:             make(map[common.Hash]uint64),
		lastRewardsPerShare: make(map[common.Hash]*big.Int),
	}
}

func NewStakerWithValue(liquidity uint64, rewards map[common.Hash]uint64, lastLPFeesPerShare map[common.Hash]*big.Int) *Staker {
	return &Staker{
		liquidity:           liquidity,
		rewards:             rewards,
		lastRewardsPerShare: lastLPFeesPerShare,
	}
}

func (staker *Staker) Clone() *Staker {
	res := NewStaker()
	res.liquidity = staker.liquidity
	for k, v := range staker.rewards {
		res.rewards[k] = v
	}
	for k, v := range staker.lastRewardsPerShare {
		res.lastRewardsPerShare[k] = new(big.Int).Set(v)
	}
	return res
}

func (staker *Staker) getDiff(
	stakingPoolID, nftID string, compareStaker *Staker, stateChange *v2utils.StateChange,
) *v2utils.StateChange {
	newStateChange := stateChange
	stakingChange := &v2utils.StakingPoolChange{}
	if compareStaker == nil {
		stakingChange = &v2utils.StakingPoolChange{
			IsChanged: true,
			TokenIDs:  make(map[string]bool),
		}
		newStateChange.StakingPool[stakingPoolID][nftID] = stakingChange
		for tokenID := range staker.rewards {
			newStateChange.StakingPool[stakingPoolID][nftID].TokenIDs[tokenID] = true
		}
	} else {
		if staker.liquidity != compareStaker.liquidity {
			stakingChange.IsChanged = true
		}
		newStateChange.StakingPool[stakingPoolID][nftID] = stakingChange
		for tokenID, value := range staker.rewards {
			if v, ok := compareStaker.rewards[nftID]; !ok || !reflect.DeepEqual(v, value) {
				if stakingChange.TokenIDs == nil {
					stakingChange.TokenIDs = make(map[string]bool)
				}
				newStateChange.StakingPool[stakingPoolID][nftID].TokenIDs[tokenID] = true
			}
		}
	}
	return newStateChange
}

func addStakingPoolState(
	stakingPoolStates map[string]*StakingPoolState, stakingPoolIDs map[string]uint,
) map[string]*StakingPoolState {
	for k := range stakingPoolIDs {
		if stakingPoolStates[k] == nil {
			stakingPoolStates[k] = NewStakingPoolState()
		}
	}
	return stakingPoolStates
}

func (share *Share) updateToDB(
	env StateEnvironment, poolPairID, nftID string, shareChange *v2utils.ShareChange,
) error {
	if shareChange.IsChanged {
		nftID, err := common.Hash{}.NewHashFromStr(nftID)
		err = statedb.StorePdexv3Share(
			env.StateDB(), poolPairID,
			*nftID,
			share.amount, share.tradingFees, share.lastLPFeesPerShare,
		)
		if err != nil {
			return err
		}
	}
	for tokenID, value := range share.tradingFees {
		if shareChange.TradingFees[tokenID.String()] {
			err := statedb.StorePdexv3ShareTradingFee(
				env.StateDB(), poolPairID, nftID,
				statedb.NewPdexv3ShareTradingFeeStateWithValue(tokenID, value),
			)
			if err != nil {
				return err
			}
		}
	}
	for tokenID, value := range share.lastLPFeesPerShare {
		if shareChange.LastLPFeesPerShare[tokenID.String()] {
			err := statedb.StorePdexv3ShareLastLpFeePerShare(
				env.StateDB(), poolPairID, nftID,
				statedb.NewPdexv3ShareLastLpFeePerShareStateWithValue(tokenID, value),
			)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func initShares(poolPairID string, stateDB *statedb.StateDB) (map[string]*Share, error) {
	res := make(map[string]*Share)
	shareStates, err := statedb.GetPdexv3Shares(stateDB, poolPairID)
	if err != nil {
		return nil, err
	}
	for nftID, shareState := range shareStates {
		tradingFees, err := statedb.GetPdexv3ShareTradingFees(stateDB, poolPairID, nftID)
		if err != nil {
			return nil, err
		}
		lastLPFeesPerShare, err := statedb.GetPdexv3ShareLastLpFeesPerShare(stateDB, poolPairID, nftID)
		if err != nil {
			return nil, err
		}
		res[nftID] = NewShareWithValue(shareState.Amount(), tradingFees, lastLPFeesPerShare)
	}
	return res, nil
}

func initStakers(stakingPoolID string, stateDB *statedb.StateDB) (map[string]*Staker, uint64, error) {
	res := make(map[string]*Staker)
	totalLiquidity := uint64(0)
	stakerStates, err := statedb.GetPdexv3Stakers(stateDB, stakingPoolID)
	if err != nil {
		return res, totalLiquidity, err
	}
	for nftID, stakerState := range stakerStates {
		totalLiquidity += stakerState.Liquidity()
		rewards, err := statedb.GetPdexv3StakerRewards(stateDB, stakingPoolID, nftID)
		if err != nil {
			return res, totalLiquidity, err
		}
		lastRewardsPerShare, err := statedb.GetPdexv3StakerLastRewardsPerShare(stateDB, stakingPoolID, nftID)
		if err != nil {
			return res, totalLiquidity, err
		}
		res[nftID] = NewStakerWithValue(stakerState.Liquidity(), rewards, lastRewardsPerShare)
	}
	return res, totalLiquidity, nil
}
