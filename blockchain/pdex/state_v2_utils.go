package pdex

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"

	"github.com/incognitochain/incognito-chain/blockchain/pdex/v2utils"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	metadataPdexv3 "github.com/incognitochain/incognito-chain/metadata/pdexv3"
	"github.com/incognitochain/incognito-chain/privacy"
)

type Share struct {
	amount                uint64
	accessOTA             []byte
	lmLockedAmount        uint64
	tradingFees           map[common.Hash]uint64
	lastLPFeesPerShare    map[common.Hash]*big.Int
	lastLmRewardsPerShare map[common.Hash]*big.Int
}

func (share *Share) AccessOTA() []byte {
	return share.accessOTA
}

func (share *Share) Amount() uint64 {
	return share.amount
}

func (share *Share) LmLockedShareAmount() uint64 {
	return share.lmLockedAmount
}

func (share *Share) LastLPFeesPerShare() map[common.Hash]*big.Int {
	res := make(map[common.Hash]*big.Int)
	for k, v := range share.lastLPFeesPerShare {
		res[k] = big.NewInt(0).Set(v)
	}
	return res
}

func (share *Share) LastLmRewardsPerShare() map[common.Hash]*big.Int {
	res := make(map[common.Hash]*big.Int)
	for k, v := range share.lastLmRewardsPerShare {
		res[k] = big.NewInt(0).Set(v)
	}
	return res
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
		amount:                0,
		tradingFees:           map[common.Hash]uint64{},
		lastLPFeesPerShare:    map[common.Hash]*big.Int{},
		lastLmRewardsPerShare: map[common.Hash]*big.Int{},
	}
}

func NewShareWithValue(
	amount, lmLockedAmount uint64,
	accessOTA []byte,
	tradingFees map[common.Hash]uint64,
	lastLPFeesPerShare, lastLmRewardsPerShare map[common.Hash]*big.Int,
) *Share {
	return &Share{
		amount:                amount,
		accessOTA:             accessOTA,
		lmLockedAmount:        lmLockedAmount,
		tradingFees:           tradingFees,
		lastLPFeesPerShare:    lastLPFeesPerShare,
		lastLmRewardsPerShare: lastLmRewardsPerShare,
	}
}

func (share *Share) Clone() *Share {
	res := NewShare()
	res.accessOTA = share.accessOTA
	res.amount = share.amount
	res.lmLockedAmount = share.lmLockedAmount
	res.tradingFees = map[common.Hash]uint64{}
	for k, v := range share.tradingFees {
		res.tradingFees[k] = v
	}
	res.lastLPFeesPerShare = map[common.Hash]*big.Int{}
	for k, v := range share.lastLPFeesPerShare {
		res.lastLPFeesPerShare[k] = new(big.Int).Set(v)
	}
	res.lastLmRewardsPerShare = map[common.Hash]*big.Int{}
	for k, v := range share.lastLmRewardsPerShare {
		res.lastLmRewardsPerShare[k] = new(big.Int).Set(v)
	}
	return res
}

func (share *Share) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(struct {
		Amount                uint64                   `json:"Amount"`
		AccessOTA             []byte                   `json:"AccessOTA,omitempty"`
		TradingFees           map[common.Hash]uint64   `json:"TradingFees"`
		LastLPFeesPerShare    map[common.Hash]*big.Int `json:"LastLPFeesPerShare"`
		LmLockedAmount        uint64                   `json:"LmLockedAmount,omitempty"`
		LastLmRewardsPerShare map[common.Hash]*big.Int `json:"LastLmRewardsPerShare,omitempty"`
	}{
		AccessOTA:             share.accessOTA,
		Amount:                share.amount,
		TradingFees:           share.tradingFees,
		LastLPFeesPerShare:    share.lastLPFeesPerShare,
		LmLockedAmount:        share.lmLockedAmount,
		LastLmRewardsPerShare: share.lastLmRewardsPerShare,
	})
	if err != nil {
		return []byte{}, err
	}
	return data, nil
}

func (share *Share) UnmarshalJSON(data []byte) error {
	temp := struct {
		Amount                uint64                   `json:"Amount"`
		AccessOTA             []byte                   `json:"AccessOTA,omitempty"`
		TradingFees           map[common.Hash]uint64   `json:"TradingFees"`
		LastLPFeesPerShare    map[common.Hash]*big.Int `json:"LastLPFeesPerShare"`
		LmLockedAmount        uint64                   `json:"LmLockedAmount,omitempty"`
		LastLmRewardsPerShare map[common.Hash]*big.Int `json:"LastLmRewardsPerShare,omitempty"`
	}{}
	err := json.Unmarshal(data, &temp)
	if err != nil {
		return err
	}
	share.accessOTA = temp.AccessOTA
	share.amount = temp.Amount
	share.lmLockedAmount = temp.LmLockedAmount
	share.tradingFees = temp.TradingFees
	share.lastLPFeesPerShare = temp.LastLPFeesPerShare
	share.lastLmRewardsPerShare = temp.LastLmRewardsPerShare
	return nil
}

func (share *Share) getDiff(
	compareShare *Share,
	shareChange *v2utils.ShareChange,
) *v2utils.ShareChange {
	newShareChange := shareChange
	if newShareChange == nil {
		newShareChange = v2utils.NewShareChange()
	}
	if compareShare == nil {
		newShareChange.IsChanged = true
		for tokenID := range share.tradingFees {
			newShareChange.TradingFees[tokenID.String()] = true
		}
		for tokenID := range share.lastLPFeesPerShare {
			newShareChange.LastLPFeesPerShare[tokenID.String()] = true
		}
		for tokenID := range share.lastLmRewardsPerShare {
			newShareChange.LastLmRewardsPerShare[tokenID.String()] = true
		}
	} else {
		if share.amount != compareShare.amount || !bytes.Equal(share.accessOTA, compareShare.accessOTA) || share.lmLockedAmount != compareShare.lmLockedAmount {
			newShareChange.IsChanged = true
		}
		newShareChange.TradingFees = v2utils.DifMapHashUint64(share.tradingFees).GetDiff(v2utils.DifMapHashUint64(compareShare.tradingFees))
		newShareChange.LastLPFeesPerShare = v2utils.DifMapHashBigInt(share.lastLPFeesPerShare).GetDiff(v2utils.DifMapHashBigInt(compareShare.lastLPFeesPerShare))
		newShareChange.LastLmRewardsPerShare = v2utils.DifMapHashBigInt(share.lastLmRewardsPerShare).GetDiff(v2utils.DifMapHashBigInt(compareShare.lastLmRewardsPerShare))
	}

	return newShareChange
}

func (share *Share) setAccessOTA(accessOTA []byte) {
	share.accessOTA = accessOTA
}

func (share *Share) isValidAccessOTA(burntOTA metadataPdexv3.AccessOTA) (bool, error) {
	if !bytes.Equal(share.accessOTA, burntOTA.ToBytesS()) {
		return false, errors.New("Not valid access OTA")
	}
	return true, nil
}

func (share *Share) isEmpty() bool {
	if share.amount != 0 {
		return false
	}
	for _, v := range share.tradingFees {
		if v != 0 {
			return false
		}
	}
	return true
}

type Staker struct {
	liquidity           uint64
	accessOTA           []byte
	rewards             map[common.Hash]uint64
	lastRewardsPerShare map[common.Hash]*big.Int
}

func (staker *Staker) Liquidity() uint64 {
	return staker.liquidity
}

func (staker *Staker) LastRewardsPerShare() map[common.Hash]*big.Int {
	res := make(map[common.Hash]*big.Int)
	for k, v := range staker.lastRewardsPerShare {
		res[k] = big.NewInt(0).Set(v)
	}
	return res
}

func (staker *Staker) Rewards() map[common.Hash]uint64 {
	res := make(map[common.Hash]uint64)
	for k, v := range staker.rewards {
		res[k] = v
	}
	return res
}

func (staker *Staker) AccessOTA() []byte {
	return staker.accessOTA
}

func (staker *Staker) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(struct {
		Liquidity           uint64                   `json:"Liquidity"`
		AccessOTA           []byte                   `json:"AccessOTA,omitempty"`
		Rewards             map[common.Hash]uint64   `json:"Rewards"`
		LastRewardsPerShare map[common.Hash]*big.Int `json:"LastRewardsPerShare"`
	}{
		Liquidity:           staker.liquidity,
		AccessOTA:           staker.accessOTA,
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
		AccessOTA          []byte                   `json:"AccessOTA,omitempty"`
		Rewards            map[common.Hash]uint64   `json:"Rewards"`
		LastLPFeesPerShare map[common.Hash]*big.Int `json:"LastRewardsPerShare"`
	}{}
	err := json.Unmarshal(data, &temp)
	if err != nil {
		return err
	}
	staker.accessOTA = temp.AccessOTA
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

func NewStakerWithValue(
	liquidity uint64, accessOTA []byte,
	rewards map[common.Hash]uint64, lastLPFeesPerShare map[common.Hash]*big.Int,
) *Staker {
	return &Staker{
		accessOTA:           accessOTA,
		liquidity:           liquidity,
		rewards:             rewards,
		lastRewardsPerShare: lastLPFeesPerShare,
	}
}

func (staker *Staker) Clone() *Staker {
	res := NewStaker()
	res.liquidity = staker.liquidity
	res.accessOTA = staker.accessOTA
	for k, v := range staker.rewards {
		res.rewards[k] = v
	}
	for k, v := range staker.lastRewardsPerShare {
		res.lastRewardsPerShare[k] = new(big.Int).Set(v)
	}
	return res
}

func (staker *Staker) getDiff(
	compareStaker *Staker, stakerChange *v2utils.StakerChange,
) *v2utils.StakerChange {
	newStakerChange := stakerChange
	if newStakerChange == nil {
		newStakerChange = v2utils.NewStakerChange()
	}
	if compareStaker == nil {
		newStakerChange.IsChanged = true
		for tokenID := range staker.rewards {
			newStakerChange.Rewards[tokenID.String()] = true
		}
		for tokenID := range staker.lastRewardsPerShare {
			newStakerChange.LastRewardsPerShare[tokenID.String()] = true
		}
	} else {
		if staker.liquidity != compareStaker.liquidity || !bytes.Equal(staker.accessOTA, compareStaker.accessOTA) {
			newStakerChange.IsChanged = true
		}
		newStakerChange.LastRewardsPerShare = v2utils.DifMapHashBigInt(staker.lastRewardsPerShare).GetDiff(v2utils.DifMapHashBigInt(compareStaker.lastRewardsPerShare))
		newStakerChange.Rewards = v2utils.DifMapHashUint64(staker.rewards).GetDiff(v2utils.DifMapHashUint64(compareStaker.rewards))
	}
	return newStakerChange
}

func (staker *Staker) setAccessOTA(accessOTA []byte) {
	staker.accessOTA = accessOTA
}

func (staker *Staker) isValidAccessOTA(burntOTA metadataPdexv3.AccessOTA) (bool, error) {
	if !bytes.Equal(staker.accessOTA, burntOTA.ToBytesS()) {
		return false, errors.New("Not valid access OTA")
	}
	return true, nil
}

func (staker *Staker) isEmpty() bool {
	if staker.liquidity != 0 {
		return false
	}
	for _, v := range staker.rewards {
		if v != 0 {
			return false
		}
	}
	return true
}

func (staker *Staker) deleteFromDB(
	env StateEnvironment, stakingPoolID string, nftID common.Hash,
	stakerChange *v2utils.StakerChange,
) error {
	err := statedb.DeletePdexv3Staker(env.StateDB(), stakingPoolID, nftID)
	if err != nil {
		return err
	}
	for tokenID := range stakerChange.LastRewardsPerShare {
		err = statedb.DeletePdexv3StakerLastRewardPerShare(env.StateDB(), stakingPoolID, nftID.String(), tokenID)
		if err != nil {
			return err
		}
	}
	for tokenID := range stakerChange.Rewards {
		err = statedb.DeletePdexv3StakerReward(env.StateDB(), stakingPoolID, nftID.String(), tokenID)
		if err != nil {
			return err
		}
	}
	return nil
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
			env.StateDB(), poolPairID, *nftID,
			statedb.NewPdexv3ShareStateWithValue(
				*nftID, share.amount, share.accessOTA, share.lmLockedAmount,
			),
		)
		if err != nil {
			return err
		}
	}
	for tokenID, isChanged := range shareChange.TradingFees {
		if isChanged {
			tokenHash, err := common.Hash{}.NewHashFromStr(tokenID)
			if err != nil {
				return err
			}
			if tradingFee, found := share.tradingFees[*tokenHash]; found {
				err := statedb.StorePdexv3ShareTradingFee(
					env.StateDB(), poolPairID, nftID,
					statedb.NewPdexv3ShareTradingFeeStateWithValue(*tokenHash, tradingFee),
				)
				if err != nil {
					return err
				}
			} else {
				err := statedb.DeletePdexv3ShareTradingFee(
					env.StateDB(), poolPairID, nftID, tokenID,
				)
				if err != nil {
					return err
				}
			}
		}
	}
	for tokenID, isChanged := range shareChange.LastLPFeesPerShare {
		if isChanged {
			tokenHash, err := common.Hash{}.NewHashFromStr(tokenID)
			if err != nil {
				return err
			}
			if lastLPFeesPerShare, found := share.lastLPFeesPerShare[*tokenHash]; found {
				err := statedb.StorePdexv3ShareLastLpFeePerShare(
					env.StateDB(), poolPairID, nftID,
					statedb.NewPdexv3ShareLastLpFeePerShareStateWithValue(*tokenHash, lastLPFeesPerShare),
				)
				if err != nil {
					return err
				}
			} else {
				err := statedb.DeletePdexv3ShareLastLpFeePerShare(
					env.StateDB(), poolPairID, nftID, tokenID,
				)
				if err != nil {
					return err
				}
			}
		}
	}
	for tokenID, isChanged := range shareChange.LastLmRewardsPerShare {
		if isChanged {
			tokenHash, err := common.Hash{}.NewHashFromStr(tokenID)
			if err != nil {
				return err
			}
			if lastLmRewardsPerShare, found := share.lastLmRewardsPerShare[*tokenHash]; found {
				err := statedb.StorePdexv3ShareLastLmRewardsPerShare(
					env.StateDB(), poolPairID, nftID,
					statedb.NewPdexv3ShareLastLmRewardPerShareStateWithValue(*tokenHash, lastLmRewardsPerShare),
				)
				if err != nil {
					return err
				}
			} else {
				err := statedb.DeletePdexv3ShareLastLmRewardsPerShare(
					env.StateDB(), poolPairID, nftID, tokenID,
				)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (share *Share) deleteFromDB(
	env StateEnvironment,
	poolPairID, nftID string,
	shareChange *v2utils.ShareChange,
) error {
	err := statedb.DeletePdexv3Share(env.StateDB(), poolPairID, nftID)
	if err != nil {
		return err
	}
	for tokenID, isChanged := range shareChange.TradingFees {
		if isChanged {
			err := statedb.DeletePdexv3ShareTradingFee(
				env.StateDB(), poolPairID, nftID, tokenID,
			)
			if err != nil {
				return err
			}
		}
	}
	for tokenID, isChanged := range shareChange.LastLPFeesPerShare {
		if isChanged {
			err = statedb.DeletePdexv3ShareLastLpFeePerShare(
				env.StateDB(), poolPairID, nftID, tokenID,
			)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

type OrderRewardDetail struct {
	receiver *privacy.OTAReceiver
	amount   uint64
}

func NewOrderRewardDetail() *OrderRewardDetail {
	return &OrderRewardDetail{}
}

func (o *OrderRewardDetail) Clone() *OrderRewardDetail {
	return &OrderRewardDetail{
		receiver: o.receiver,
		amount:   o.amount,
	}
}

func NewOrderRewardDetailWithValue(
	receiver *privacy.OTAReceiver, amount uint64,
) *OrderRewardDetail {
	return &OrderRewardDetail{
		amount:   amount,
		receiver: receiver,
	}
}

func (o *OrderRewardDetail) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(struct {
		Receiver *privacy.OTAReceiver `json:"Receiver,omitempty"`
		Amount   uint64               `json:"Amount"`
	}{
		Receiver: o.receiver,
		Amount:   o.amount,
	})
	if err != nil {
		return []byte{}, err
	}
	return data, nil
}

func (o *OrderRewardDetail) UnmarshalJSON(data []byte) error {
	temp := struct {
		Receiver *privacy.OTAReceiver `json:"Receiver,omitempty"`
		Amount   uint64               `json:"Amount"`
	}{}
	err := json.Unmarshal(data, &temp)
	if err != nil {
		return err
	}
	o.receiver = temp.Receiver
	o.amount = temp.Amount
	return nil
}

func (o *OrderRewardDetail) Amount() uint64 {
	return o.amount
}

func (o *OrderRewardDetail) Receiver() privacy.OTAReceiver {
	return *o.receiver
}

type OrderReward struct {
	uncollectedRewards map[common.Hash]*OrderRewardDetail
	withdrawnStatus    byte
	txReqID            *common.Hash
}

func (orderReward *OrderReward) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(struct {
		UncollectedRewards map[common.Hash]*OrderRewardDetail `json:"UncollectedRewards"`
		WithdrawnStatus    byte                               `json:"WithdrawnStatus"`
		TxReqID            *common.Hash                       `json:"TxReqID,omitempty"`
	}{
		UncollectedRewards: orderReward.uncollectedRewards,
		WithdrawnStatus:    orderReward.withdrawnStatus,
		TxReqID:            orderReward.txReqID,
	})
	if err != nil {
		return []byte{}, err
	}
	return data, nil
}

func (orderReward *OrderReward) UnmarshalJSON(data []byte) error {
	temp := struct {
		UncollectedRewards map[common.Hash]*OrderRewardDetail `json:"UncollectedRewards"`
		WithdrawnStatus    byte                               `json:"WithdrawnStatus"`
		TxReqID            *common.Hash                       `json:"TxReqID,omitempty"`
	}{}
	err := json.Unmarshal(data, &temp)
	if err != nil {
		return err
	}
	orderReward.withdrawnStatus = temp.WithdrawnStatus
	orderReward.uncollectedRewards = temp.UncollectedRewards
	orderReward.txReqID = temp.TxReqID
	return nil
}

func (orderReward *OrderReward) isEmpty() bool {
	if orderReward.withdrawnStatus == WaitToWithdrawOrderReward {
		return false
	}
	if len(orderReward.uncollectedRewards) != 0 {
		return false
	}
	return true
}

func (orderReward *OrderReward) TxReqID() *common.Hash {
	return orderReward.txReqID
}

func (orderReward *OrderReward) WithdrawnStatus() byte {
	return orderReward.withdrawnStatus
}

func (orderReward *OrderReward) UncollectedRewards() map[common.Hash]*OrderRewardDetail {
	res := map[common.Hash]*OrderRewardDetail{}
	for k, v := range orderReward.uncollectedRewards {
		res[k] = v
	}
	return res
}

func (orderReward *OrderReward) AddReward(tokenID common.Hash, amount uint64) {
	oldAmount := uint64(0)
	if _, ok := orderReward.uncollectedRewards[tokenID]; ok {
		oldAmount = orderReward.uncollectedRewards[tokenID].amount
	}
	if orderReward.uncollectedRewards[tokenID] == nil {
		orderReward.uncollectedRewards[tokenID] = NewOrderRewardDetail()
	}
	temp := orderReward.uncollectedRewards[tokenID]
	temp.amount = oldAmount + amount
	orderReward.uncollectedRewards[tokenID] = temp
}

func NewOrderReward() *OrderReward {
	return &OrderReward{
		uncollectedRewards: make(map[common.Hash]*OrderRewardDetail),
	}
}

func NewOrderRewardWithValue(
	withdrawnStatus byte, uncollectedRewards map[common.Hash]*OrderRewardDetail, txReqID *common.Hash,
) *OrderReward {
	return &OrderReward{
		withdrawnStatus:    withdrawnStatus,
		uncollectedRewards: uncollectedRewards,
		txReqID:            txReqID,
	}
}

func (orderReward *OrderReward) Clone() *OrderReward {
	res := NewOrderReward()
	for k, v := range orderReward.uncollectedRewards {
		res.uncollectedRewards[k] = v.Clone()
	}
	res.withdrawnStatus = orderReward.withdrawnStatus
	res.txReqID = orderReward.txReqID
	return res
}

func (orderReward *OrderReward) getDiff(
	compareOrderReward *OrderReward,
	orderRewardChange *v2utils.OrderRewardChange,
) *v2utils.OrderRewardChange {
	newOrderRewardChange := orderRewardChange
	if newOrderRewardChange == nil {
		newOrderRewardChange = v2utils.NewOrderRewardChange()
	}
	newOrderRewardChange.UncollectedReward = make(map[string]bool)
	if compareOrderReward == nil {
		for tokenID := range orderReward.uncollectedRewards {
			newOrderRewardChange.UncollectedReward[tokenID.String()] = true
		}
	} else {
		orderRewardAmount := make(map[common.Hash]uint64)
		compareOrderRewardAmount := make(map[common.Hash]uint64)
		for k, v := range orderReward.uncollectedRewards {
			orderRewardAmount[k] = v.amount
		}
		for k, v := range compareOrderReward.uncollectedRewards {
			compareOrderRewardAmount[k] = v.amount
		}
		newOrderRewardChange.UncollectedReward = v2utils.DifMapHashUint64(orderRewardAmount).GetDiff(v2utils.DifMapHashUint64(compareOrderRewardAmount))
		if orderReward.withdrawnStatus != compareOrderReward.withdrawnStatus ||
			orderReward.txReqID != compareOrderReward.txReqID {
			for k := range orderReward.uncollectedRewards {
				newOrderRewardChange.UncollectedReward[k.String()] = true
			}
		}
	}
	return newOrderRewardChange
}

type MakingVolume struct {
	volume map[string]*big.Int // accessID -> amount
}

func (makingVolume *MakingVolume) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(struct {
		Volume map[string]*big.Int `json:"Volume"`
	}{
		Volume: makingVolume.volume,
	})
	if err != nil {
		return []byte{}, err
	}
	return data, nil
}

func (makingVolume *MakingVolume) UnmarshalJSON(data []byte) error {
	temp := struct {
		Volume map[string]*big.Int `json:"Volume"`
	}{}
	err := json.Unmarshal(data, &temp)
	if err != nil {
		return err
	}
	makingVolume.volume = temp.Volume
	return nil
}

func (makingVolume *MakingVolume) AddVolume(nftID string, amount *big.Int) {
	oldAmount := big.NewInt(0)
	if _, ok := makingVolume.volume[nftID]; ok {
		oldAmount = makingVolume.volume[nftID]
	}
	makingVolume.volume[nftID] = big.NewInt(0).Add(oldAmount, amount)
}

func NewMakingVolume() *MakingVolume {
	return &MakingVolume{
		volume: make(map[string]*big.Int),
	}
}

func (makingVolume *MakingVolume) Clone() *MakingVolume {
	res := NewMakingVolume()
	for k, v := range makingVolume.volume {
		res.volume[k] = new(big.Int).Set(v)
	}
	return res
}

func (makingVolume *MakingVolume) getDiff(
	compareMakingVolume *MakingVolume,
	makingVolumeChange *v2utils.MakingVolumeChange,
) *v2utils.MakingVolumeChange {
	newMakingVolumeChange := makingVolumeChange
	if newMakingVolumeChange == nil {
		newMakingVolumeChange = v2utils.NewMakingVolumeChange()
	}
	if compareMakingVolume == nil {
		for nftID := range makingVolume.volume {
			newMakingVolumeChange.Volume[nftID] = true
		}
	} else {
		newMakingVolumeChange.Volume = v2utils.DifMapStringBigInt(makingVolume.volume).GetDiff(v2utils.DifMapStringBigInt(compareMakingVolume.volume))
	}
	return newMakingVolumeChange
}

func unlockLmLockedShareAmount(
	pairs map[string]*PoolPairState,
	params *Params,
	beaconHeight uint64,
) (map[string]*PoolPairState, error) {
	for _, poolPair := range pairs {
		for shareID, lockedRecord := range poolPair.lmLockedShare {
			// check if this share exists
			if _, ok := poolPair.shares[shareID]; !ok {
				delete(poolPair.lmLockedShare, shareID)
				continue
			}
			for lockedHeight, lockedAmount := range lockedRecord {
				if lockedHeight+params.MiningRewardPendingBlocks < beaconHeight {
					delete(lockedRecord, lockedHeight)

					// releaseAmount = min(lockedAmount, poolPair.shares[shareID].lmLockedAmount)
					releaseAmount := lockedAmount
					if releaseAmount > poolPair.shares[shareID].lmLockedAmount {
						releaseAmount = poolPair.shares[shareID].lmLockedAmount
					}

					shareIDBytes, err := common.Hash{}.NewHashFromStr(shareID)
					if err != nil {
						return pairs, fmt.Errorf("Invalid shareID: %s", shareID)
					}
					poolPair.shares[shareID].tradingFees, err = poolPair.RecomputeLPRewards(*shareIDBytes)
					if err != nil {
						return pairs, fmt.Errorf("Error when tracking LP reward: %v", err)
					}

					poolPair.shares[shareID].lastLPFeesPerShare = poolPair.LpFeesPerShare()
					poolPair.shares[shareID].lastLmRewardsPerShare = poolPair.LmRewardsPerShare()

					poolPair.shares[shareID].lmLockedAmount, err = executeOperationUint64(poolPair.shares[shareID].lmLockedAmount, releaseAmount, subOperator)
					if err != nil {
						return pairs, fmt.Errorf("newShare.lmLockedAmount is out of range")
					}
					poolPairLmLockedShareAmount, err := executeOperationUint64(poolPair.state.LmLockedShareAmount(), releaseAmount, subOperator)
					if err != nil {
						return pairs, fmt.Errorf("poolPairLmLockedShareAmount is out of range")
					}
					poolPair.state.SetLmLockedShareAmount(poolPairLmLockedShareAmount)
				}
			}
		}
	}
	return pairs, nil
}
