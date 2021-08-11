package pdex

import (
	"encoding/json"
	"reflect"
)

type Share struct {
	amount                  uint64
	tradingFees             map[string]uint64
	lastUpdatedBeaconHeight uint64
}

func NewShare() *Share {
	return &Share{
		tradingFees: make(map[string]uint64),
	}
}

func NewShareWithValue(
	amount uint64,
	tradingFees map[string]uint64,
	lastUpdatedBeaconHeight uint64,
) *Share {
	return &Share{
		amount:                  amount,
		tradingFees:             tradingFees,
		lastUpdatedBeaconHeight: lastUpdatedBeaconHeight,
	}
}

func (share *Share) Clone() *Share {
	res := NewShare()
	res.amount = share.amount
	res.lastUpdatedBeaconHeight = share.lastUpdatedBeaconHeight
	res.tradingFees = make(map[string]uint64)
	for k, v := range share.tradingFees {
		res.tradingFees[k] = v
	}
	return res
}

func (share *Share) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(struct {
		Amount                  uint64            `json:"Amount"`
		TradingFees             map[string]uint64 `json:"TradingFees"`
		LastUpdatedBeaconHeight uint64            `json:"LastUpdatedBeaconHeight"`
	}{
		Amount:                  share.amount,
		TradingFees:             share.tradingFees,
		LastUpdatedBeaconHeight: share.lastUpdatedBeaconHeight,
	})
	if err != nil {
		return []byte{}, err
	}
	return data, nil
}

func (share *Share) UnmarshalJSON(data []byte) error {
	temp := struct {
		Amount                  uint64            `json:"Amount"`
		TradingFees             map[string]uint64 `json:"TradingFees"`
		LastUpdatedBeaconHeight uint64            `json:"LastUpdatedBeaconHeight"`
	}{}
	err := json.Unmarshal(data, &temp)
	if err != nil {
		return err
	}
	share.amount = temp.Amount
	share.lastUpdatedBeaconHeight = temp.LastUpdatedBeaconHeight
	share.tradingFees = temp.TradingFees
	return nil
}

func (share *Share) getDiff(
	nftID string,
	txHash string,
	compareShare *Share,
	stateChange *StateChange,
) *StateChange {
	newStateChange := stateChange
	if stateChange.shares[nftID] == nil {
		stateChange.shares[nftID] = make(map[string]*ShareChange)
	}
	if compareShare == nil {
		newStateChange.shares[nftID][txHash] = &ShareChange{
			isChanged: true,
		}
		for tokenID := range share.tradingFees {
			if newStateChange.shares[nftID][txHash].tokenIDs == nil {
				newStateChange.shares[nftID][txHash].tokenIDs = make(map[string]bool)
			}
			newStateChange.shares[nftID][txHash].tokenIDs[tokenID] = true
		}
	} else {
		if share.amount != compareShare.amount || share.lastUpdatedBeaconHeight != compareShare.lastUpdatedBeaconHeight {
			newStateChange.shares[nftID][txHash] = &ShareChange{
				isChanged: true,
			}
		}
		for k, v := range share.tradingFees {
			if m, ok := compareShare.tradingFees[k]; !ok || !reflect.DeepEqual(m, v) {
				if newStateChange.shares[nftID][txHash].tokenIDs == nil {
					newStateChange.shares[nftID][txHash].tokenIDs = make(map[string]bool)
				}
				newStateChange.shares[nftID][txHash].tokenIDs[k] = true
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

type ShareChange struct {
	isChanged bool
	tokenIDs  map[string]bool
}

type StateChange struct {
	poolPairIDs map[string]bool
	shares      map[string]map[string]*ShareChange
	orders      map[string]map[int]bool
}

func NewStateChange() *StateChange {
	return &StateChange{
		poolPairIDs: make(map[string]bool),
		shares:      make(map[string]map[string]*ShareChange),
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
