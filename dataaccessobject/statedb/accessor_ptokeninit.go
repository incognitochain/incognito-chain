package statedb

import (
	"bytes"
	"encoding/json"
	"log"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject"
	"github.com/incognitochain/incognito-chain/dataaccessobject/rawdbv2"
)

func TrackPTokenInitReqWithStatus(stateDB *StateDB, txReqID common.Hash, status byte) error {
	GeneratePTokenInitObjecKey()

	key := GenerateBridgeStatusObjectKey(txReqID)
	value := NewBridgeStatusStateWithValue(txReqID, status)
	err := stateDB.SetStateObject(BridgeStatusObjectType, key, value)
	if err != nil {
		return NewStatedbError(TrackBridgeReqWithStatusError, err)
	}
	return nil
}

func GetBridgeReqWithStatus(stateDB *StateDB, txReqID common.Hash) (byte, error) {
	key := GenerateBridgeStatusObjectKey(txReqID)
	bridgeStatusState, has, err := stateDB.getBridgeStatusState(key)
	if err != nil {
		return 0, NewStatedbError(GetBridgeReqWithStatusError, err)
	}
	if !has {
		return common.BridgeRequestNotFoundStatus, nil
	}
	tempTxReqID := bridgeStatusState.TxReqID()
	if !tempTxReqID.IsEqual(&txReqID) {
		panic("same key wrong value")
	}
	return bridgeStatusState.Status(), nil
}
