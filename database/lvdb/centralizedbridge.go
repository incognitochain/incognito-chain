package lvdb

import (
	"encoding/json"

	"github.com/constant-money/constant-chain/common"
	"github.com/constant-money/constant-chain/database"
	"github.com/pkg/errors"
	lvdberr "github.com/syndtr/goleveldb/leveldb/errors"
	"github.com/syndtr/goleveldb/leveldb/util"
)

type TokenWithAmount struct {
	TokenID      *common.Hash `json:"tokenId"`
	Amount       uint64       `json:"amount"`
}

func (db *db) CountUpDepositedAmtByTokenID(
	tokenID *common.Hash,
	amount uint64,
) error {
	// don't need to have atomic operation here since instructions on beacon would be processed one by one, not in parallel
	key := append(centralizedBridgePrefix, tokenID[:]...)

	tokenWithAmtBytes, dbErr := db.lvdb.Get(key, nil)
	if dbErr != nil && dbErr != lvdberr.ErrNotFound {
		return database.NewDatabaseError(database.UnexpectedError, errors.Wrap(dbErr, "db.lvdb.Get"))
	}

	var newTokenWithAmount TokenWithAmount
	if len(tokenWithAmtBytes) == 0 {
		newTokenWithAmount = TokenWithAmount{
			TokenID:      tokenID,
			Amount:       amount,
		}
	} else { // found existing amount info
		var existingTokenWithAmount TokenWithAmount
		unmarshalErr := json.Unmarshal(tokenWithAmtBytes, &existingTokenWithAmount)
		if unmarshalErr != nil {
			return unmarshalErr
		}
		newTokenWithAmount = TokenWithAmount{
			TokenID:      existingTokenWithAmount.TokenID,
			Amount:       existingTokenWithAmount.Amount + amount,
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

func (db *db) GetBridgeTokensAmounts() ([][]byte, error) {
	iter := db.NewIterator(util.BytesPrefix(centralizedBridgePrefix), nil)
	results := [][]byte{}
	for iter.Next() {
		value := iter.Value()
		results = append(results, value)
	}
	iter.Release()
	err := iter.Error()
	if err != nil && err != lvdberr.ErrNotFound {
		return nil, database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "db.lvdb.Get"))
	}
	return results, nil
}

func (db *db) DeductAmtByTokenID(
	tokenID *common.Hash,
	amount uint64,
) error {
	// don't need to have atomic operation here since instructions on beacon would be processed one by one, not in parallel
	key := append(centralizedBridgePrefix, tokenID[:]...)

	tokenWithAmtBytes, dbErr := db.lvdb.Get(key, nil)
	if dbErr != nil && dbErr != lvdberr.ErrNotFound {
		return database.NewDatabaseError(database.UnexpectedError, errors.Wrap(dbErr, "db.lvdb.Get"))
	}

	if len(tokenWithAmtBytes) == 0 { // not found
		return nil
	} 
	// found existing amount info
	var existingTokenWithAmount TokenWithAmount
	unmarshalErr := json.Unmarshal(tokenWithAmtBytes, &existingTokenWithAmount)
	if unmarshalErr != nil {
		return unmarshalErr
	}
	newTokenWithAmount := TokenWithAmount{
		TokenID:      existingTokenWithAmount.TokenID,
	}
	if newTokenWithAmount.Amount <= amount {
		newTokenWithAmount.Amount = 0
	} else {
		newTokenWithAmount.Amount -= amount
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

func (db *db) IsBridgeTokenExisted(
	tokenID *common.Hash,
) (bool, error) {
	key := append(centralizedBridgePrefix, tokenID[:]...)
	tokenWithAmtBytes, dbErr := db.lvdb.Get(key, nil)
	if dbErr != nil && dbErr != lvdberr.ErrNotFound {
		return false, database.NewDatabaseError(database.UnexpectedError, errors.Wrap(dbErr, "db.lvdb.Get"))
	}
	if len(tokenWithAmtBytes) == 0 {
		return false, nil
	}
	return true, nil
}
