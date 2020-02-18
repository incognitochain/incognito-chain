package statedb

import (
	"fmt"
	"github.com/incognitochain/incognito-chain/common"
	"strings"
)

func StorePrivacyToken(stateDB *StateDB, tokenID common.Hash, name string, symbol string, tokenType int, mintable bool, amount uint64, info []byte, txHash common.Hash) error {
	key := GenerateTokenObjectKey(tokenID)
	_, has, err := stateDB.GetTokenState(key)
	if err != nil {
		return NewStatedbError(StorePrivacyTokenError, err)
	}
	if has {
		return nil
	}
	value := NewTokenStateWithValue(tokenID, name, symbol, tokenType, mintable, amount, info, txHash, []common.Hash{})
	err = stateDB.SetStateObject(TokenObjectType, key, value)
	if err != nil {
		return NewStatedbError(StorePrivacyTokenError, err)
	}
	return nil
}

func StorePrivacyTokenTx(stateDB *StateDB, tokenID common.Hash, txHash common.Hash) error {
	key := GenerateTokenObjectKey(tokenID)
	t, has, err := stateDB.GetTokenState(key)
	if err != nil {
		return NewStatedbError(GetPrivacyTokenError, err)
	}
	if !has {
		err := StorePrivacyToken(stateDB, tokenID, "", "", UnknownToken, false, 0, []byte{}, txHash)
		if err != nil {
			return err
		}
		return nil
	}
	t.AddTxs([]common.Hash{txHash})
	err = stateDB.SetStateObject(TokenObjectType, key, t)
	if err != nil {
		return NewStatedbError(StorePrivacyTokenError, err)
	}
	return nil
}

func HasPrivacyTokenID(stateDB *StateDB, tokenID common.Hash) (bool, error) {
	key := GenerateTokenObjectKey(tokenID)
	t, has, err := stateDB.GetTokenState(key)
	if err != nil {
		return false, NewStatedbError(GetPrivacyTokenError, err)
	}
	if strings.Compare(t.TokenID().String(), tokenID.String()) != 0 {
		panic("same key wrong value")
	}
	return has, nil
}

func ListPrivacyToken(stateDB *StateDB) (map[common.Hash]*TokenState, error) {
	return stateDB.GetAllToken(), nil
}

func GetPrivacyTokenTxs(stateDB *StateDB, tokenID common.Hash) ([]common.Hash, error) {
	txs, has, err := stateDB.GetTokenTxs(tokenID)
	if err != nil {
		return []common.Hash{}, NewStatedbError(GetPrivacyTokenTxsError, err)
	}
	if !has {
		return []common.Hash{}, NewStatedbError(GetPrivacyTokenTxsError, fmt.Errorf("token %+v txs not exist", tokenID))
	}
	return txs, nil
}

func PrivacyTokenIDExisted(stateDB *StateDB, tokenID common.Hash) bool {
	key := GenerateTokenObjectKey(tokenID)
	tokenState, has, err := stateDB.GetTokenState(key)
	if err != nil {
		return false
	}
	tempTokenID := tokenState.TokenID()
	if has && !tempTokenID.IsEqual(&tokenID) {
		panic("same key wrong value")
	}
	return has
}
