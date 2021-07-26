package statedb

import (
	"fmt"
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
	contributions map[string]Pdexv3ContributionState,
) error {
	for k, v := range contributions {
		key := GeneratePdexv3ContributionObjectKey(k)
		value := NewPdexv3ContributionState()
		*value = v
		err := stateDB.SetStateObject(Pdexv3ContributionObjectType, key, value)
		if err != nil {
			return NewStatedbError(StorePdexv3ContributionError, err)
		}
	}
	return nil
}

func DeletePdexv3WaitingContributions(
	stateDB *StateDB,
	contributions map[string]Pdexv3ContributionState,
) error {
	for k := range contributions {
		key := GeneratePdexv3ContributionObjectKey(k)
		if !stateDB.MarkDeleteStateObject(Pdexv3ContributionObjectType, key) {
			return fmt.Errorf("Can't delete contributions with pair hash %v", k)
		}
	}
	return nil
}

func StorePdexv3PoolPair(
	stateDB *StateDB,
	poolPairID string,
	poolPair Pdexv3PoolPairState,
	shares map[string]Pdexv3ShareState,
) error {
	key := GeneratePdexv3PoolPairObjectKey(poolPairID)
	value := NewPdexv3PoolPairState()
	*value = poolPair
	err := stateDB.SetStateObject(Pdexv3PoolPairObjectType, key, value)
	if err != nil {
		return NewStatedbError(StorePdexv3PoolPairError, err)
	}
	for k, v := range shares {
		key := GeneratePdexv3ShareObjectKey(poolPairID, k)
		value := NewPdexv3ShareState()
		*value = v
		err := stateDB.SetStateObject(Pdexv3ShareObjectType, key, value)
		if err != nil {
			return NewStatedbError(StorePdexv3ShareError, err)
		}
	}
	return nil
}

func StorePdexv3StakingPools() error {
	return nil
}

func GetPdexv3WaitingContributions(stateDB *StateDB) (map[string]Pdexv3ContributionState, error) {
	prefixHash := GetPdexv3WaitingContributionsPrefix()
	return stateDB.iterateWithPdexv3Contributions(prefixHash)
}

func GetPdexv3PoolPairs(stateDB *StateDB) (map[string]Pdexv3PoolPairState, error) {
	prefixHash := GetPdexv3PoolPairsPrefix()
	return stateDB.iterateWithPdexv3PoolPairs(prefixHash)
}

func GetPdexv3Shares(stateDB *StateDB, poolPairID string) (map[string]Pdexv3ShareState, error) {
	prefixHash := generatePdexv3ShareObjectPrefix(poolPairID)
	return stateDB.iterateWithPdexv3Shares(prefixHash)
}
