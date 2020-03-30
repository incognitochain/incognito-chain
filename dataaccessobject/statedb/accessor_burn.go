package statedb

import (
	"fmt"
	"github.com/incognitochain/incognito-chain/common"
)

func StoreBurningConfirm(stateDB *StateDB, txID common.Hash, height uint64) error {
	key := GenerateBurningConfirmObjectKey(txID)
	value := NewBurningConfirmStateWithValue(txID, height)
	err := stateDB.SetStateObject(BurningConfirmObjectType, key, value)
	if err != nil {
		return NewStatedbError(StoreBurningConfirmError, err)
	}
	return nil
}

func GetBurningConfirm(stateDB *StateDB, txID common.Hash) (uint64, error) {
	key := GenerateBurningConfirmObjectKey(txID)
	burningConfirmState, has, err := stateDB.getBurningConfirmState(key)
	if err != nil {
		return 0, NewStatedbError(GetBurningConfirmError, err)
	}
	if !has {
		return 0, NewStatedbError(GetBurningConfirmError, fmt.Errorf("burning confirm with txID %+v not found", txID))
	}
	tempTxID := burningConfirmState.TxID()
	if !tempTxID.IsEqual(&txID) {
		panic("burning confirm state")
	}
	return burningConfirmState.Height(), nil
}
