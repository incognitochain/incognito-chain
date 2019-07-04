package lvdb

import (
	"encoding/json"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/database"
	"github.com/pkg/errors"
	lvdberr "github.com/syndtr/goleveldb/leveldb/errors"
	"github.com/syndtr/goleveldb/leveldb/util"
)

type TokenWithAmount struct {
	TokenID *common.Hash `json:"tokenId"`
	Amount  uint64       `json:"amount"`
}

func (db *db) GetBridgeTokensAmounts() ([][]byte, error) {
	iter := db.lvdb.NewIterator(util.BytesPrefix(centralizedBridgePrefix), nil)
	results := [][]byte{}
	for iter.Next() {
		value := iter.Value()
		bridgedTokensAmountBytes := make([]byte, len(value))
		copy(bridgedTokensAmountBytes, value)
		results = append(results, bridgedTokensAmountBytes)
	}

	iter.Release()
	err := iter.Error()
	if err != nil && err != lvdberr.ErrNotFound {
		return nil, database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "db.lvdb.Get"))
	}
	return results, nil
}

func (db *db) IsBridgeTokenExisted(
	tokenID common.Hash,
) (bool, error) {
	key := append(centralizedBridgePrefix, tokenID[:]...)
	tokenWithAmtBytes, dbErr := db.Get(key)
	if dbErr != nil && dbErr != lvdberr.ErrNotFound {
		return false, database.NewDatabaseError(database.UnexpectedError, errors.Wrap(dbErr, "db.lvdb.Get"))
	}
	if len(tokenWithAmtBytes) == 0 {
		return false, nil
	}
	return true, nil
}

func (db *db) UpdateAmtByTokenID(
	tokenID common.Hash,
	amount uint64,
	updateType string,
) error {
	// don't need to have atomic operation here since instructions on beacon would be processed one by one, not in parallel
	key := append(centralizedBridgePrefix, tokenID[:]...)

	tokenWithAmtBytes, dbErr := db.Get(key)
	if dbErr != nil && dbErr != lvdberr.ErrNotFound {
		return database.NewDatabaseError(database.UnexpectedError, errors.Wrap(dbErr, "db.lvdb.Get"))
	}

	var newTokenWithAmount TokenWithAmount
	if len(tokenWithAmtBytes) == 0 {
		newTokenWithAmount = TokenWithAmount{
			TokenID: &tokenID,
		}
		if updateType == "-" {
			newTokenWithAmount.Amount = 0
		} else {
			newTokenWithAmount.Amount = amount
		}

	} else { // found existing amount info
		var existingTokenWithAmount TokenWithAmount
		unmarshalErr := json.Unmarshal(tokenWithAmtBytes, &existingTokenWithAmount)
		if unmarshalErr != nil {
			return unmarshalErr
		}
		newTokenWithAmount = TokenWithAmount{
			TokenID: existingTokenWithAmount.TokenID,
		}
		if updateType == "+" {
			newTokenWithAmount.Amount = existingTokenWithAmount.Amount + amount
		} else if existingTokenWithAmount.Amount <= amount {
			newTokenWithAmount.Amount = 0
		} else {
			newTokenWithAmount.Amount = existingTokenWithAmount.Amount - amount
		}
	}

	contentBytes, marshalErr := json.Marshal(newTokenWithAmount)
	if marshalErr != nil {
		return marshalErr
	}

	dbErr = db.Put(key, contentBytes)
	if dbErr != nil {
		return database.NewDatabaseError(database.UnexpectedError, errors.Wrap(dbErr, "db.lvdb.put"))
	}
	return nil
}
