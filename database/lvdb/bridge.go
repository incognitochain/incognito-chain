package lvdb

import (
	"bytes"
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

type BridgeTokenInfo struct {
	TokenID         *common.Hash `json:"tokenId"`
	Amount          uint64       `json:"amount"`
	IsCentralized   bool         `json:"isCentralized"`
	ExternalTokenID []byte       `json:"externalTokenId"`
	Network         string       `json:"network"`
}

func buildBridgedTokensAmounts(item []byte, results [][]byte) [][]byte {
	results = append(results, item)
	return results
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
	tokenWithAmtBytes, dbErr := db.lvdb.Get(key, nil)
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

	tokenWithAmtBytes, dbErr := db.lvdb.Get(key, nil)
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

func (db *db) InsertETHTxHashIssued(
	uniqETHTx []byte,
) error {
	key := append(decentralizedBridgePrefix, uniqETHTx...)
	dbErr := db.Put(key, []byte{1})
	if dbErr != nil {
		return database.NewDatabaseError(database.UnexpectedError, errors.Wrap(dbErr, "db.lvdb.put"))
	}
	return nil
}

func (db *db) IsETHTxHashIssued(
	uniqETHTx []byte,
) (bool, error) {
	key := append(decentralizedBridgePrefix, uniqETHTx...)
	contentBytes, dbErr := db.lvdb.Get(key, nil)
	if dbErr != nil && dbErr != lvdberr.ErrNotFound {
		return false, database.NewDatabaseError(database.UnexpectedError, errors.Wrap(dbErr, "db.lvdb.Get"))
	}
	if len(contentBytes) == 0 {
		return false, nil
	}
	return true, nil
}

//////////////////////////////////////////////
func (db *db) CanProcessTokenPair(
	externalTokenID []byte,
	incTokenID common.Hash,
) (bool, error) {
	if len(externalTokenID) == 0 || len(incTokenID[:]) == 0 {
		return false, nil
	}
	// check incognito bridge token is existed in centralized bridge tokens or not
	cBridgeTokenExisted, err := db.IsBridgeTokenExistedByType(incTokenID, true)
	if err != nil {
		return false, err
	}
	if cBridgeTokenExisted {
		return false, nil
	}

	dBridgeTokenExisted, err := db.IsBridgeTokenExistedByType(incTokenID, false)
	if err != nil {
		return false, err
	}
	privacyCustomTokenExisted := db.PrivacyCustomTokenIDExisted(incTokenID)
	privacyCustomTokenCrossShardExisted := db.PrivacyCustomTokenIDCrossShardExisted(incTokenID)
	if !dBridgeTokenExisted && (privacyCustomTokenExisted || privacyCustomTokenCrossShardExisted) {
		return false, nil
	}

	key := append(decentralizedBridgePrefix, incTokenID[:]...)
	contentBytes, dbErr := db.lvdb.Get(key, nil)
	if dbErr != nil && dbErr != lvdberr.ErrNotFound {
		return false, database.NewDatabaseError(database.UnexpectedError, errors.Wrap(dbErr, "db.lvdb.Get"))
	}
	if len(contentBytes) > 0 {
		var bridgeTokenInfo BridgeTokenInfo
		err := json.Unmarshal(contentBytes, &bridgeTokenInfo)
		if err != nil {
			return false, err
		}
		if bytes.Equal(bridgeTokenInfo.ExternalTokenID[:], externalTokenID[:]) {
			return true, nil
		}
		return false, nil
	}
	// else: could not find incTokenID out
	iter := db.lvdb.NewIterator(util.BytesPrefix(decentralizedBridgePrefix), nil)
	for iter.Next() {
		value := iter.Value()
		itemBytes := make([]byte, len(value))
		copy(itemBytes, value)
		var bridgeTokenInfo BridgeTokenInfo
		err := json.Unmarshal(itemBytes, &bridgeTokenInfo)
		if err != nil {
			return false, err
		}
		if !bytes.Equal(bridgeTokenInfo.ExternalTokenID, externalTokenID) {
			continue
		}
		return false, nil
	}

	iter.Release()
	err = iter.Error()
	if err != nil && err != lvdberr.ErrNotFound {
		return false, database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "db.lvdb.Get"))
	}
	// both tokens are not existed -> can create new one
	return true, nil
}

func getBridgePrefix(isCentralized bool) []byte {
	prefix := []byte{}
	if isCentralized {
		prefix = centralizedBridgePrefix
	} else {
		prefix = decentralizedBridgePrefix
	}
	return prefix
}

func (db *db) UpdateBridgeTokenPairInfo(
	incTokenID common.Hash,
	externalTokenID []byte,
	isCentralized bool,
) error {
	prefix := getBridgePrefix(isCentralized)
	key := append(prefix, incTokenID[:]...)
	bridgeTokenInfo := BridgeTokenInfo{
		TokenID:         &incTokenID,
		IsCentralized:   isCentralized,
		ExternalTokenID: externalTokenID,
	}
	bridgeTokenInfoBytes, err := json.Marshal(bridgeTokenInfo)
	if err != nil {
		return err
	}

	dbErr := db.Put(key, bridgeTokenInfoBytes)
	if dbErr != nil {
		return database.NewDatabaseError(database.UnexpectedError, errors.Wrap(dbErr, "db.lvdb.put"))
	}
	return nil
}

func (db *db) IsBridgeTokenExistedByType(
	incTokenID common.Hash,
	isCentralized bool,
) (bool, error) {
	prefix := getBridgePrefix(isCentralized)
	key := append(prefix, incTokenID[:]...)
	tokenInfoBytes, dbErr := db.lvdb.Get(key, nil)
	if dbErr != nil && dbErr != lvdberr.ErrNotFound {
		return false, database.NewDatabaseError(database.UnexpectedError, errors.Wrap(dbErr, "db.lvdb.Get"))
	}
	if len(tokenInfoBytes) == 0 {
		return false, nil
	}
	return true, nil
}
