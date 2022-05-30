package statedb

import (
	"fmt"
	"math/big"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject"
	"github.com/incognitochain/incognito-chain/dataaccessobject/rawdbv2"
)

func TrackPdexv3Status(stateDB *StateDB, statusType []byte, statusSuffix []byte, statusContent []byte) error {
	key := GeneratePdexv3StatusObjectKey(statusType, statusSuffix)
	value := NewPdexv3StatusStateWithValue(statusType, statusSuffix, statusContent)
	err := stateDB.SetStateObject(Pdexv3StatusObjectType, key, value)
	if err != nil {
		return NewStatedbError(StorePdexv3StatusError, err)
	}
	return nil
}

func GetPdexv3Status(stateDB *StateDB, statusType []byte, statusSuffix []byte) ([]byte, error) {
	key := GeneratePdexv3StatusObjectKey(statusType, statusSuffix)
	s, has, err := stateDB.getPdexv3StatusByKey(key)
	if err != nil {
		return []byte{}, NewStatedbError(GetPdexv3StatusError, err)
	}
	if !has {
		return []byte{}, NewStatedbError(GetPdexv3StatusError, fmt.Errorf("status %+v with prefix %+v not found", string(statusType), statusSuffix))
	}
	return s.statusContent, nil
}

func StorePdexv3Params(
	stateDB *StateDB,
	defaultFeeRateBPS uint,
	feeRateBPS map[string]uint,
	prvDiscountPercent uint,
	tradingProtocolFeePercent uint,
	tradingStakingPoolRewardPercent uint,
	pdexRewardPoolPairsShare map[string]uint,
	stakingPoolsShare map[string]uint,
	stakingRewardTokens []common.Hash,
	mintNftRequireAmount uint64,
	maxOrdersPerNft uint,
	autoWithdrawOrderLimitAmount uint,
	minPRVReserveTradingRate uint64,
	defaultOrderTradingRewardRatioBPS uint,
	orderTradingRewardRatioBPS map[string]uint,
	orderLiquidityMiningBPS map[string]uint,
	daoContributingPercent uint,
	miningRewardPendingBlocks uint64,
	orderMiningRewardRatioBPS map[string]uint,
) error {
	key := GeneratePdexv3ParamsObjectKey()
	value := NewPdexv3ParamsWithValue(
		defaultFeeRateBPS,
		feeRateBPS,
		prvDiscountPercent,
		tradingProtocolFeePercent,
		tradingStakingPoolRewardPercent,
		pdexRewardPoolPairsShare,
		stakingPoolsShare,
		stakingRewardTokens,
		mintNftRequireAmount,
		maxOrdersPerNft,
		autoWithdrawOrderLimitAmount,
		minPRVReserveTradingRate,
		defaultOrderTradingRewardRatioBPS,
		orderTradingRewardRatioBPS,
		orderLiquidityMiningBPS,
		daoContributingPercent,
		miningRewardPendingBlocks,
		orderMiningRewardRatioBPS,
	)
	err := stateDB.SetStateObject(Pdexv3ParamsObjectType, key, value)
	if err != nil {
		return NewStatedbError(StorePdexv3ParamsError, err)
	}
	return nil
}

func GetPdexv3Params(stateDB *StateDB) (*Pdexv3Params, error) {
	key := GeneratePdexv3ParamsObjectKey()
	s, has, err := stateDB.getPdexv3ParamsByKey(key)
	if err != nil {
		return nil, NewStatedbError(GetPdexv3ParamsError, err)
	}
	if !has {
		return nil, NewStatedbError(GetPdexv3ParamsError, fmt.Errorf("could not found pDex v3 params in statedb"))
	}
	return s, nil
}

func StorePdexv3WaitingContributions(
	stateDB *StateDB,
	contributions map[string]rawdbv2.Pdexv3Contribution,
) error {
	for k, v := range contributions {
		key := GeneratePdexv3ContributionObjectKey(k)
		value := NewPdexv3ContributionStateWithValue(
			v, k,
		)
		err := stateDB.SetStateObject(Pdexv3ContributionObjectType, key, value)
		if err != nil {
			return NewStatedbError(StorePdexv3ContributionError, err)
		}
	}
	return nil
}

func DeletePdexv3WaitingContributions(
	stateDB *StateDB,
	pairHashes []string,
) error {
	for _, pairHash := range pairHashes {
		key := GeneratePdexv3ContributionObjectKey(pairHash)
		if !stateDB.MarkDeleteStateObject(Pdexv3ContributionObjectType, key) {
			dataaccessobject.Logger.Log.Infof("Can't delete contributions with pair hash %v maybe 2 contributions have been matched\n", pairHash)
		}
	}
	return nil
}

func StorePdexv3Share(
	stateDB *StateDB, poolPairID string, nftID common.Hash,
	state *Pdexv3ShareState,
) error {
	key := GeneratePdexv3ShareObjectKey(poolPairID, nftID.String())
	return stateDB.SetStateObject(Pdexv3ShareObjectType, key, state)
}

func DeletePdexv3Share(
	stateDB *StateDB, poolPairID, nftID string,
) error {
	key := GeneratePdexv3ShareObjectKey(poolPairID, nftID)
	if !stateDB.MarkDeleteStateObject(Pdexv3ShareObjectType, key) {
		return fmt.Errorf("Cannot delete share with ID %v - %v", poolPairID, nftID)
	}
	return nil
}

func StorePdexv3PoolPair(
	stateDB *StateDB,
	poolPairID string,
	poolPair rawdbv2.Pdexv3PoolPair,
) error {
	state := NewPdexv3PoolPairStateWithValue(poolPairID, poolPair)
	key := GeneratePdexv3PoolPairObjectKey(poolPairID)
	return stateDB.SetStateObject(Pdexv3PoolPairObjectType, key, state)
}

func StorePdexv3NftIDs(stateDB *StateDB, nftIDs map[string]uint64) error {
	for k, v := range nftIDs {
		key := GeneratePdexv3NftObjectKey(k)
		nftHash, err := common.Hash{}.NewHashFromStr(k)
		if err != nil {
			return err
		}
		value := NewPdexv3NftStateWithValue(*nftHash, v)
		err = stateDB.SetStateObject(Pdexv3NftObjectType, key, value)
		if err != nil {
			return NewStatedbError(StorePdexv3NftsError, err)
		}
	}
	return nil
}

func StorePdexv3Staker(stateDB *StateDB, stakingPoolID, nftID string, state *Pdexv3StakerState) error {
	key := GeneratePdexv3StakerObjectKey(stakingPoolID, nftID)
	return stateDB.SetStateObject(Pdexv3StakerObjectType, key, state)
}

func DeletePdexv3Staker(
	stateDB *StateDB, stakingPoolID string, nftID common.Hash,
) error {
	key := GeneratePdexv3StakerObjectKey(stakingPoolID, nftID.String())
	if !stateDB.MarkDeleteStateObject(Pdexv3StakerObjectType, key) {
		return fmt.Errorf("Cannot delete staker with ID %v - %v", stakingPoolID, nftID.String())
	}
	return nil
}

func StorePdexv3Order(stateDB *StateDB, orderState Pdexv3OrderState) error {
	v := orderState.Value()
	key := GeneratePdexv3OrderObjectKey(orderState.PoolPairID(), v.Id())
	return stateDB.SetStateObject(Pdexv3OrderObjectType, key, &orderState)
}

func StorePdexv3PoolPairLpFeePerShare(
	stateDB *StateDB, poolPairID string, state *Pdexv3PoolPairLpFeePerShareState,
) error {
	key := GeneratePdexv3PoolPairLpFeePerShareObjectKey(poolPairID, state.tokenID.String())
	return stateDB.SetStateObject(Pdexv3PoolPairLpFeePerShareObjectType, key, state)
}

func StorePdexv3PoolPairLmRewardPerShare(
	stateDB *StateDB, poolPairID string, state *Pdexv3PoolPairLmRewardPerShareState,
) error {
	key := GeneratePdexv3PoolPairLmRewardPerShareObjectKey(poolPairID, state.tokenID.String())
	return stateDB.SetStateObject(Pdexv3PoolPairLmRewardPerShareObjectType, key, state)
}

func DeletePdexv3PoolPairLpFeePerShare(
	stateDB *StateDB, poolPairID, tokenID string,
) error {
	key := GeneratePdexv3PoolPairLpFeePerShareObjectKey(poolPairID, tokenID)
	if !stateDB.MarkDeleteStateObject(Pdexv3PoolPairLpFeePerShareObjectType, key) {
		return fmt.Errorf("Cannot delete poolPairLpFeePerShare with ID %v - %v", poolPairID, tokenID)
	}
	return nil
}

func DeletePdexv3PoolPairLmRewardPerShare(
	stateDB *StateDB, poolPairID, tokenID string,
) error {
	key := GeneratePdexv3PoolPairLmRewardPerShareObjectKey(poolPairID, tokenID)
	if !stateDB.MarkDeleteStateObject(Pdexv3PoolPairLmRewardPerShareObjectType, key) {
		return fmt.Errorf("Cannot delete poolPairLmRewardPerShare with ID %v - %v", poolPairID, tokenID)
	}
	return nil
}

func StorePdexv3PoolPairProtocolFee(
	stateDB *StateDB, poolPairID string, state *Pdexv3PoolPairProtocolFeeState,
) error {
	key := GeneratePdexv3PoolPairProtocolFeeObjectKey(poolPairID, state.tokenID.String())
	return stateDB.SetStateObject(Pdexv3PoolPairProtocolFeeObjectType, key, state)
}

func DeletePdexv3PoolPairProtocolFee(
	stateDB *StateDB, poolPairID, tokenID string,
) error {
	key := GeneratePdexv3PoolPairProtocolFeeObjectKey(poolPairID, tokenID)
	if !stateDB.MarkDeleteStateObject(Pdexv3PoolPairProtocolFeeObjectType, key) {
		return fmt.Errorf("Cannot delete poolPair protool fee with ID %v - %v", poolPairID, tokenID)
	}
	return nil
}

func StorePdexv3PoolPairStakingPoolFee(
	stateDB *StateDB, poolPairID string, state *Pdexv3PoolPairStakingPoolFeeState,
) error {
	key := GeneratePdexv3PoolPairStakingPoolFeeObjectKey(poolPairID, state.tokenID.String())
	return stateDB.SetStateObject(Pdexv3PoolPairStakingPoolFeeObjectType, key, state)
}

func DeletePdexv3PoolPairStakingPoolFee(
	stateDB *StateDB, poolPairID, tokenID string,
) error {
	key := GeneratePdexv3PoolPairStakingPoolFeeObjectKey(poolPairID, tokenID)
	if !stateDB.MarkDeleteStateObject(Pdexv3PoolPairStakingPoolFeeObjectType, key) {
		return fmt.Errorf("Cannot delete share with ID %v - %v", poolPairID, tokenID)
	}
	return nil
}

func StorePdexv3PoolPairOrderReward(
	stateDB *StateDB, poolPairID string, state *Pdexv3PoolPairOrderRewardState,
) error {
	key := GeneratePdexv3PoolPairOrderRewardObjectPrefix(poolPairID, state.nftID, state.tokenID)
	return stateDB.SetStateObject(Pdexv3PoolPairOrderRewardObjectType, key, state)
}

func DeletePdexv3PoolPairOrderReward(
	stateDB *StateDB, poolPairID, nftID string, tokenID common.Hash,
) error {
	key := GeneratePdexv3PoolPairOrderRewardObjectPrefix(poolPairID, nftID, tokenID)
	if !stateDB.MarkDeleteStateObject(Pdexv3PoolPairOrderRewardObjectType, key) {
		return fmt.Errorf("Cannot delete pool pair order reward with ID %v - %v - %v", poolPairID, nftID, tokenID.String())
	}
	return nil
}

func StorePdexv3PoolPairMakingVolume(
	stateDB *StateDB, poolPairID string, state *Pdexv3PoolPairMakingVolumeState,
) error {
	key := GeneratePdexv3PoolPairMakingVolumeObjectPrefix(poolPairID, state.tokenID, state.nftID)
	return stateDB.SetStateObject(Pdexv3PoolPairMakingVolumeObjectType, key, state)
}

func StorePdexv3PoolPairLmLockedShare(
	stateDB *StateDB, poolPairID string, state *Pdexv3PoolPairLmLockedShareState,
) error {
	key := GeneratePdexv3PoolPairLmLockedShareObjectKey(poolPairID, state.nftID, state.beaconHeight)
	return stateDB.SetStateObject(Pdexv3PoolPairLmLockedShareObjectType, key, state)
}

func DeletePdexv3PoolPairMakingVolume(
	stateDB *StateDB, poolPairID, nftID string, tokenID common.Hash,
) error {
	key := GeneratePdexv3PoolPairMakingVolumeObjectPrefix(poolPairID, tokenID, nftID)
	if !stateDB.MarkDeleteStateObject(Pdexv3PoolPairMakingVolumeObjectType, key) {
		return fmt.Errorf("Cannot delete pool pair making volume with ID %v - %v - %v", poolPairID, tokenID.String(), nftID)
	}
	return nil
}

func DeletePdexv3PoolPairLmLockedShare(
	stateDB *StateDB, poolPairID, nftID string, beaconHeight uint64,
) error {
	key := GeneratePdexv3PoolPairLmLockedShareObjectKey(poolPairID, nftID, beaconHeight)
	if !stateDB.MarkDeleteStateObject(Pdexv3PoolPairLmLockedShareObjectType, key) {
		return fmt.Errorf("Cannot delete pool pair lm locked share with ID %v - %v - %v", poolPairID, nftID, beaconHeight)
	}
	return nil
}

func StorePdexv3ShareTradingFee(
	stateDB *StateDB, poolPairID, nftID string, state *Pdexv3ShareTradingFeeState,
) error {
	key := GeneratePdexv3ShareTradingFeeObjectKey(poolPairID, nftID, state.tokenID.String())
	return stateDB.SetStateObject(Pdexv3ShareTradingFeeObjectType, key, state)
}

func DeletePdexv3ShareTradingFee(
	stateDB *StateDB, poolPairID, nftID, tokenID string,
) error {
	key := GeneratePdexv3ShareTradingFeeObjectKey(poolPairID, nftID, tokenID)
	if !stateDB.MarkDeleteStateObject(Pdexv3ShareTradingFeeObjectType, key) {
		return fmt.Errorf("Cannot delete share trading fee with ID %v - %v - %v", poolPairID, nftID, tokenID)
	}
	return nil
}

func StorePdexv3ShareLastLpFeePerShare(
	stateDB *StateDB, poolPairID, nftID string, state *Pdexv3ShareLastLpFeePerShareState,
) error {
	key := GeneratePdexv3ShareLastLpFeePerShareObjectKey(poolPairID, nftID, state.tokenID.String())
	return stateDB.SetStateObject(Pdexv3ShareLastLPFeesPerShareObjectType, key, state)
}

func DeletePdexv3ShareLastLpFeePerShare(
	stateDB *StateDB, poolPairID, nftID, tokenID string,
) error {
	key := GeneratePdexv3ShareLastLpFeePerShareObjectKey(poolPairID, nftID, tokenID)
	if !stateDB.MarkDeleteStateObject(Pdexv3ShareTradingFeeObjectType, key) {
		return fmt.Errorf("Cannot delete share last lp fee per share with ID %v - %v - %v", poolPairID, nftID, tokenID)
	}
	return nil
}

func StorePdexv3ShareLastLmRewardsPerShare(
	stateDB *StateDB, poolPairID, nftID string, state *Pdexv3ShareLastLmRewardPerShareState,
) error {
	key := GeneratePdexv3ShareLastLmRewardPerShareObjectKey(poolPairID, nftID, state.tokenID.String())
	return stateDB.SetStateObject(Pdexv3ShareLastLmRewardPerShareObjectType, key, state)
}

func DeletePdexv3ShareLastLmRewardsPerShare(
	stateDB *StateDB, poolPairID, nftID, tokenID string,
) error {
	key := GeneratePdexv3ShareLastLmRewardPerShareObjectKey(poolPairID, nftID, tokenID)
	if !stateDB.MarkDeleteStateObject(Pdexv3ShareLastLmRewardPerShareObjectType, key) {
		return fmt.Errorf("Cannot delete share last lm reward per share with ID %v - %v - %v", poolPairID, nftID, tokenID)
	}
	return nil
}

func StorePdexv3StakingPoolRewardPerShare(
	stateDB *StateDB, stakingPoolID string, state *Pdexv3StakingPoolRewardPerShareState,
) error {
	key := GeneratePdexv3StakingPoolRewardPerShareObjectKey(stakingPoolID, state.tokenID.String())
	return stateDB.SetStateObject(Pdexv3StakingPoolRewardPerShareObjectType, key, state)
}

func DeletePdexv3StakingPoolRewardPerShare(
	stateDB *StateDB, poolPairID, tokenID string,
) error {
	key := GeneratePdexv3StakingPoolRewardPerShareObjectKey(poolPairID, tokenID)
	if !stateDB.MarkDeleteStateObject(Pdexv3StakingPoolRewardPerShareObjectType, key) {
		return fmt.Errorf("Cannot delete staking pool reward per share with ID %v - %v", poolPairID, tokenID)
	}
	return nil
}

func StorePdexv3StakerLastRewardPerShare(
	stateDB *StateDB, stakingPoolID, nftID string, state *Pdexv3StakerLastRewardPerShareState,
) error {
	key := GeneratePdexv3StakerLastRewardPerShareObjectKey(stakingPoolID, nftID, state.tokenID.String())
	return stateDB.SetStateObject(Pdexv3StakerLastRewardPerShareObjectType, key, state)
}

func DeletePdexv3StakerLastRewardPerShare(
	stateDB *StateDB, stakingPoolID, nftID, tokenID string,
) error {
	key := GeneratePdexv3StakerLastRewardPerShareObjectKey(stakingPoolID, nftID, tokenID)
	if !stateDB.MarkDeleteStateObject(Pdexv3StakerLastRewardPerShareObjectType, key) {
		return fmt.Errorf("Cannot delete staker last reward per share with ID %v - %v - %v", stakingPoolID, nftID, tokenID)
	}
	return nil
}

func StorePdexv3StakerReward(
	stateDB *StateDB, stakingPoolID, nftID string, state *Pdexv3StakerRewardState,
) error {
	key := GeneratePdexv3StakerRewardObjectKey(stakingPoolID, nftID, state.tokenID.String())
	return stateDB.SetStateObject(Pdexv3StakerRewardObjectType, key, state)
}

func DeletePdexv3StakerReward(
	stateDB *StateDB, stakingPoolID, nftID, tokenID string,
) error {
	key := GeneratePdexv3StakerRewardObjectKey(stakingPoolID, nftID, tokenID)
	if !stateDB.MarkDeleteStateObject(Pdexv3StakerRewardObjectType, key) {
		return fmt.Errorf("Cannot delete staker reward with ID %v - %v - %v", stakingPoolID, nftID, tokenID)
	}
	return nil
}

func DeletePdexv3Order(stateDB *StateDB, pairID, orderID string) error {
	key := GeneratePdexv3OrderObjectKey(pairID, orderID)
	if !stateDB.MarkDeleteStateObject(Pdexv3OrderObjectType, key) {
		return fmt.Errorf("Cannot delete order with ID %v - %v", pairID, orderID)
	}
	return nil
}

func GetPdexv3WaitingContributions(stateDB *StateDB) (map[string]rawdbv2.Pdexv3Contribution, error) {
	prefixHash := GetPdexv3WaitingContributionsPrefix()
	return stateDB.iterateWithPdexv3Contributions(prefixHash)
}

func GetPdexv3PoolPairs(stateDB *StateDB) (map[string]Pdexv3PoolPairState, error) {
	prefixHash := GetPdexv3PoolPairsPrefix()
	return stateDB.iterateWithPdexv3PoolPairs(prefixHash)
}

func GetPdexv3PoolPair(stateDB *StateDB, poolPairID string) (*Pdexv3PoolPairState, error) {
	key := GeneratePdexv3PoolPairObjectKey(poolPairID)
	s, has, err := stateDB.getPdexv3PoolPairState(key)
	if err != nil {
		return nil, NewStatedbError(GetPdexv3PoolPairError, err)
	}
	if !has {
		return nil, NewStatedbError(GetPdexv3PoolPairError, fmt.Errorf("could not found pDex v3 pool pairs in statedb"))
	}
	return s, nil
}

func GetPdexv3Share(stateDB *StateDB, poolPairID, nftID string) (*Pdexv3ShareState, error) {
	key := GeneratePdexv3ShareObjectKey(poolPairID, nftID)
	s, has, err := stateDB.getPdexv3ShareState(key)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, fmt.Errorf("could not found pDex v3 pool pair share in statedb")
	}
	return s, nil
}

func GetPdexv3Shares(stateDB *StateDB, poolPairID string) (
	map[string]Pdexv3ShareState, error,
) {
	prefixHash := generatePdexv3ShareObjectPrefix(poolPairID)
	return stateDB.iterateWithPdexv3Shares(prefixHash)
}

func GetPdexv3NftIDs(stateDB *StateDB) (map[string]uint64, error) {
	prefixHash := GetPdexv3NftPrefix()
	return stateDB.iterateWithPdexv3Nfts(prefixHash)
}

func GetPdexv3NftID(stateDB *StateDB, nftID string) (*Pdexv3NftState, error) {
	key := GeneratePdexv3NftObjectKey(nftID)
	s, has, err := stateDB.getPdexv3NftIDState(key)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, fmt.Errorf("could not found pDex v3 nftID in statedb")
	}
	return s, nil
}

func GetPdexv3Orders(stateDB *StateDB, poolPairID string) (
	map[string]Pdexv3OrderState,
	error,
) {
	prefixHash := generatePdexv3OrderObjectPrefix(poolPairID)
	return stateDB.iterateWithPdexv3Orders(prefixHash)
}

func GetPdexv3Stakers(stateDB *StateDB, stakingPoolID string) (map[string]Pdexv3StakerState, error) {
	prefixHash := generatePdexv3StakerObjectPrefix(stakingPoolID)
	return stateDB.iterateWithPdexv3Stakers(prefixHash)
}

func GetPdexv3PoolPairLpFeesPerShares(stateDB *StateDB, poolPairID string) (
	map[common.Hash]*big.Int, error,
) {
	prefixHash := generatePdexv3PoolPairLpFeePerShareObjectPrefix(poolPairID)
	return stateDB.iterateWithPdexv3PoolPairLpFeesPerShare(prefixHash)
}

func GetPdexv3PoolPairLmRewardPerShares(stateDB *StateDB, poolPairID string) (
	map[common.Hash]*big.Int, error,
) {
	prefixHash := generatePdexv3PoolPairLmRewardPerShareObjectPrefix(poolPairID)
	return stateDB.iterateWithPdexv3PoolPairLmRewardPerShare(prefixHash)
}

func GetPdexv3PoolPairProtocolFees(stateDB *StateDB, poolPairID string) (
	map[common.Hash]uint64, error,
) {
	prefixHash := generatePdexv3PoolPairProtocolFeeObjectPrefix(poolPairID)
	return stateDB.iterateWithPdexv3PoolPairProtocolFees(prefixHash)
}

func GetPdexv3PoolPairStakingPoolFees(stateDB *StateDB, poolPairID string) (
	map[common.Hash]uint64, error,
) {
	prefixHash := generatePdexv3PoolPairStakingPoolFeeObjectPrefix(poolPairID)
	return stateDB.iterateWithPdexv3PoolPairStakingPoolFees(prefixHash)
}

func GetPdexv3PoolPairMakingVolume(stateDB *StateDB, poolPairID string) (
	map[common.Hash]map[string]*big.Int, error,
) {
	prefixHash := generatePdexv3PoolPairMakingVolumeObjectPrefix(poolPairID)
	return stateDB.iterateWithPdexv3PoolPairMakingVolume(prefixHash)
}

func GetPdexv3PoolPairLmLockedShare(stateDB *StateDB, poolPairID string) (
	map[string]map[uint64]uint64, error,
) {
	prefixHash := generatePdexv3PoolPairLmLockedShareObjectPrefix(poolPairID)
	return stateDB.iterateWithPdexv3PoolPairLmLockedShare(prefixHash)
}

func GetPdexv3PoolPairOrderReward(stateDB *StateDB, poolPairID string) (
	map[string]map[common.Hash]uint64, error,
) {
	prefixHash := generatePdexv3PoolPairOrderRewardObjectPrefix(poolPairID)
	return stateDB.iterateWithPdexv3PoolPairOrderReward(prefixHash)
}

func GetPdexv3ShareTradingFees(stateDB *StateDB, poolPairID, nftID string) (
	map[common.Hash]uint64, error,
) {
	prefixHash := generatePdexv3ShareTradingFeeObjectPrefix(poolPairID, nftID)
	return stateDB.iterateWithPdexv3ShareTradingFees(prefixHash)
}

func GetPdexv3ShareLastLpFeesPerShare(stateDB *StateDB, poolPairID, nftID string) (
	map[common.Hash]*big.Int, error,
) {
	prefixHash := generatePdexv3ShareLastLpFeePerShareObjectPrefix(poolPairID, nftID)
	return stateDB.iterateWithPdexv3ShareLastLpFeesPerShare(prefixHash)
}

func GetPdexv3ShareLastLmRewardPerShare(stateDB *StateDB, poolPairID, nftID string) (
	map[common.Hash]*big.Int, error,
) {
	prefixHash := generatePdexv3ShareLastLmRewardPerShareObjectPrefix(poolPairID, nftID)
	return stateDB.iterateWithPdexv3ShareLastLmRewardPerShare(prefixHash)
}

func GetPdexv3StakingPoolRewardsPerShare(stateDB *StateDB, stakingPoolID string) (
	map[common.Hash]*big.Int, error,
) {
	prefixHash := generatePdexv3StakingPoolRewardPerShareObjectPrefix(stakingPoolID)
	return stateDB.iterateWithPdexv3StakingPoolRewardsPerShare(prefixHash)
}

func GetPdexv3StakerRewards(stateDB *StateDB, stakingPoolID, nftID string) (
	map[common.Hash]uint64, error,
) {
	prefixHash := generatePdexv3StakerRewardObjectPrefix(stakingPoolID, nftID)
	return stateDB.iterateWithPdexv3StakerRewards(prefixHash)
}

func GetPdexv3StakerLastRewardsPerShare(stateDB *StateDB, stakingPoolID, nftID string) (
	map[common.Hash]*big.Int, error,
) {
	prefixHash := generatePdexv3StakerLastRewardPerShareObjectPrefix(stakingPoolID, nftID)
	return stateDB.iterateWithPdexv3StakerLastRewardsPerShare(prefixHash)
}

func GetPdexv3Staker(stateDB *StateDB, stakingPoolID, nftID string) (*Pdexv3StakerState, error) {
	key := GeneratePdexv3StakerObjectKey(stakingPoolID, nftID)
	s, has, err := stateDB.getPdexv3StakerByKey(key)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, fmt.Errorf("could not found pDex v3 staker in statedb")
	}
	return s, nil
}
