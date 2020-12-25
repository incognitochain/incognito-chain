package statedb

import (
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

func GetPTokenInit(stateDB *StateDB, tokenID string) (*rawdbv2.PTokenInitInfo, error) {
	key := GeneratePTokenInitObjecKey(tokenID)
	ptiState, has, err := stateDB.getPTokenInitState(key)

	if err != nil {
		return nil, NewStatedbError(GetPTokenInitError, err)
	}
	if !has || ptiState == nil {
		return nil, nil
	}

	return rawdbv2.NewPTokenInitInfo(
		ptiState.TokenID(),
		ptiState.TokenName(),
		ptiState.TokenSymbol(),
		ptiState.Amount(),
	), nil
}

func GetAllPTokenInits(stateDB *StateDB) ([]*rawdbv2.PTokenInitInfo, error) {
	pTokenInits := []*rawdbv2.PTokenInitInfo{}
	pTokenInitStates, err := stateDB.getAllPTokenInits()
	if err != nil {
		return pTokenInits, err
	}
	for _, ptiState := range pTokenInitStates {
		if ptiState == nil {
			continue
		}
		pTokenInitInfo := rawdbv2.NewPTokenInitInfo(
			ptiState.TokenID(),
			ptiState.TokenName(),
			ptiState.TokenSymbol(),
			ptiState.Amount(),
		)
		pTokenInits = append(pTokenInits, pTokenInitInfo)
	}
	return pTokenInits, nil
}
