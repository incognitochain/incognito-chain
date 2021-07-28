package statedb

import (
	"fmt"

	"github.com/incognitochain/incognito-chain/common"
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
		return []byte{}, NewStatedbError(GetPdexv3StatusError, fmt.Errorf("status %+v with prefix %+v not found", string(statusType), string(statusSuffix)))
	}
	return s.statusContent, nil
}

func StorePdexv3Params(
	stateDB *StateDB,
	defaultFeeRateBPS uint,
	feeRateBPS map[string]uint,
	prvDiscountPercent uint,
	limitProtocolFeePercent uint,
	limitStakingPoolRewardPercent uint,
	tradingProtocolFeePercent uint,
	tradingStakingPoolRewardPercent uint,
	defaultStakingPoolsShare uint,
	stakingPoolsShare map[string]uint,
) error {
	key := GeneratePdexv3ParamsObjectKey()
	value := NewPdexv3ParamsWithValue(
		defaultFeeRateBPS,
		feeRateBPS,
		prvDiscountPercent,
		limitProtocolFeePercent,
		limitStakingPoolRewardPercent,
		tradingProtocolFeePercent,
		tradingStakingPoolRewardPercent,
		defaultStakingPoolsShare,
		stakingPoolsShare,
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
			return fmt.Errorf("Can't delete contributions with pair hash %v", pairHash)
		}
	}
	return nil
}

func StorePdexv3Share(
	stateDB *StateDB,
	poolPairID string,
	share Pdexv3ShareState,
) error {
	key := GeneratePdexv3ShareObjectKey(poolPairID, share.nfctID.String())
	return stateDB.SetStateObject(Pdexv3ShareObjectType, key, share)
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

func StorePdexv3TradingFee(
	stateDB *StateDB,
	poolPairID string,
	nfctID string,
	tokenID string,
	tradingFee uint64,
) error {
	tokenHash, err := common.Hash{}.NewHashFromStr(tokenID)
	if err != nil {
		return NewStatedbError(StorePdexv3ShareError, err)
	}
	key := GeneratePdexv3TradingFeesObjectKey(poolPairID, nfctID, tokenID)
	tradingFeeState := NewPdexv3TradingFeeStateWithValue(*tokenHash, tradingFee)
	err = stateDB.SetStateObject(Pdexv3ShareObjectType, key, tradingFeeState)
	if err != nil {
		return NewStatedbError(StorePdexv3ShareError, err)
	}
	return nil
}

func StorePdexv3StakingPools() error {
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

func GetPdexv3Shares(stateDB *StateDB, poolPairID string) (
	map[string]Pdexv3ShareState,
	error,
) {
	prefixHash := generatePdexv3ShareObjectPrefix(poolPairID)
	return stateDB.iterateWithPdexv3Shares(prefixHash)
}

func GetPdexv3TradingFees(stateDB *StateDB, poolPairID, nfctID string) (
	map[string]Pdexv3TradingFeeState,
	error,
) {
	prefixHash := generatePdexv3TradingFeesObjectPrefix(poolPairID, nfctID)
	return stateDB.iterateWithPdexv3TradingFees(prefixHash)
}
