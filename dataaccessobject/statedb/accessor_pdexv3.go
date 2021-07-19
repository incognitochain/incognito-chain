package statedb

import (
	"fmt"
)

func TrackPDexV3Status(stateDB *StateDB, statusType []byte, statusSuffix []byte, statusContent []byte) error {
	key := GeneratePDexV3StatusObjectKey(statusType, statusSuffix)
	value := NewPDexV3StatusStateWithValue(statusType, statusSuffix, statusContent)
	err := stateDB.SetStateObject(PDexV3StatusObjectType, key, value)
	if err != nil {
		return NewStatedbError(StorePDexV3StatusError, err)
	}
	return nil
}

func GetPDexV3Status(stateDB *StateDB, statusType []byte, statusSuffix []byte) ([]byte, error) {
	key := GeneratePDexV3StatusObjectKey(statusType, statusSuffix)
	s, has, err := stateDB.getPDexV3StatusByKey(key)
	if err != nil {
		return []byte{}, NewStatedbError(GetPDexV3StatusError, err)
	}
	if !has {
		return []byte{}, NewStatedbError(GetPDexV3StatusError, fmt.Errorf("status %+v with prefix %+v not found", string(statusType), string(statusSuffix)))
	}
	return s.statusContent, nil
}

func StorePDexV3Params(
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
	key := GeneratePDexV3ParamsObjectKey()
	value := NewPDexV3ParamsWithValue(
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
	err := stateDB.SetStateObject(PDexV3ParamsObjectType, key, value)
	if err != nil {
		return NewStatedbError(StorePDexV3ParamsError, err)
	}
	return nil
}

func GetPDexV3Params(stateDB *StateDB) (*PDexV3Params, error) {
	key := GeneratePDexV3ParamsObjectKey()
	s, has, err := stateDB.getPDexV3ParamsByKey(key)
	if err != nil {
		return nil, NewStatedbError(GetPDexV3ParamsError, err)
	}
	if !has {
		return nil, NewStatedbError(GetPDexV3ParamsError, fmt.Errorf("could not found pDex v3 params in statedb"))
	}
	return s, nil
}
