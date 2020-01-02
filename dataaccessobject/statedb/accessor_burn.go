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
	burningConfrimState, has, err := stateDB.GetBurningConfirmState(key)
	if err != nil {
		return 0, NewStatedbError(GetBurningConfirmError, err)
	}
	if !has {
		return 0, NewStatedbError(GetBurningConfirmError, fmt.Errorf("burning confirm with txID %+v not found", txID))
	}
	tempTxID := burningConfrimState.TxID()
	if !tempTxID.IsEqual(&txID) {
		panic("burning confirm state")
	}
	return burningConfrimState.Height(), nil
}
