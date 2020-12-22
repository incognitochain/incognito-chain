package statedb

import (
	"encoding/json"
	"fmt"

	"github.com/incognitochain/incognito-chain/dataaccessobject/rawdbv2"
)

func StorePTokenInit(
	stateDB *StateDB,
	tokenID string,
	tokenName string,
	tokenSymbol string,
	amount uint64,
) error {
	key := GeneratePTokenInitObjecKey(tokenID)
	value := NewPTokenInitStateWithValue(tokenID, tokenName, tokenSymbol, amount)
	err := stateDB.SetStateObject(PTokenInitObjectType, key, value)
	if err != nil {
		return NewStatedbError(StorePTokenInitError, err)
	}
	return nil
}

func GetPTokenInit(stateDB *StateDB, tokenID string) ([]byte, error) {
	key := GeneratePTokenInitObjecKey(tokenID)
	ptiState, has, err := stateDB.getPTokenInitState(key)

	if err != nil {
		return []byte{}, NewStatedbError(GetPTokenInitError, err)
	}
	if !has {
		return []byte{}, NewStatedbError(GetPTokenInitError, fmt.Errorf("key with ptoken id %+v not found", tokenID))
	}
	res, err := json.Marshal(
		rawdbv2.NewPTokenInitInfo(
			ptiState.TokenID(),
			ptiState.TokenName(),
			ptiState.TokenSymbol(),
			ptiState.Amount(),
		),
	)
	if err != nil {
		return []byte{}, NewStatedbError(GetPTokenInitError, err)
	}
	return res, nil
}
