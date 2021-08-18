package pdex

import (
	"encoding/json"
	"math/big"
	"reflect"

	"github.com/incognitochain/incognito-chain/common"
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
	res.tradingFees = make(map[common.Hash]uint64)
	for k, v := range share.tradingFees {
		res.tradingFees[k] = v
	}
	res.lastLPFeesPerShare = make(map[common.Hash]*big.Int)
	for k, v := range share.lastLPFeesPerShare {
		res.lastLPFeesPerShare[k] = v
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
	stateChange *StateChange,
) *StateChange {
	newStateChange := stateChange
	if compareShare == nil {
		newStateChange.shares[nftID] = true
	} else {
		if share.amount != compareShare.amount {
			newStateChange.shares[nftID] = true
		}
		for k, v := range share.tradingFees {
			if m, ok := compareShare.tradingFees[k]; !ok || !reflect.DeepEqual(m, v) {
				newStateChange.shares[nftID] = true
			}
		}
		for k, v := range share.lastLPFeesPerShare {
			if m, ok := compareShare.lastLPFeesPerShare[k]; !ok || !reflect.DeepEqual(m, v) {
				newStateChange.shares[nftID] = true
			}
		}
	}
	return newStateChange
}

type StakingInfo struct {
	amount                  uint64
	uncollectedReward       uint64
	lastUpdatedBeaconHeight uint64
}

type StateChange struct {
	poolPairIDs map[string]bool
	shares      map[string]bool
	orders      map[string]map[int]bool
}

func NewStateChange() *StateChange {
	return &StateChange{
		poolPairIDs: make(map[string]bool),
		shares:      make(map[string]bool),
		orders:      make(map[string]map[int]bool),
	}
}

type StakingPoolState struct {
	liquidity        uint64
	stakers          map[string]*StakingInfo // nfst -> amount staking
	currentStakingID uint64
}

func NewStakingPoolState() *StakingPoolState {
	return &StakingPoolState{
		stakers: make(map[string]*StakingInfo),
	}
}

func NewStakingPoolStateWithValue(
	liquidity uint64,
	stakers map[string]*StakingInfo,
	currentStakingID uint64,
) *StakingPoolState {
	return &StakingPoolState{
		liquidity:        liquidity,
		stakers:          stakers,
		currentStakingID: currentStakingID,
	}
}

func (s *StakingPoolState) Clone() *StakingPoolState {
	res := NewStakingPoolState()
	return res
}
