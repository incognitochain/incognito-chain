package lvdb

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/database"
	"github.com/pkg/errors"
	lvdberr "github.com/syndtr/goleveldb/leveldb/errors"
	"github.com/syndtr/goleveldb/leveldb/util"
)

// TODO - change json to CamelCase
type BridgeTokenInfo struct {
	TokenID         *common.Hash `json:"tokenId"`
	Amount          uint64       `json:"amount"`
	ExternalTokenID []byte       `json:"externalTokenId"`
	Network         string       `json:"network"`
	IsCentralized   bool         `json:"isCentralized"`
}

func (db *db) InsertETHTxHashIssued(
	uniqETHTx []byte,
) error {
	key := append(ethTxHashIssuedPrefix, uniqETHTx...)
	dbErr := db.Put(key, []byte{1})
	if dbErr != nil {
		return database.NewDatabaseError(database.InsertETHTxHashIssuedError, errors.Wrap(dbErr, "db.lvdb.put"))
	}
	return nil
}

func (db *db) IsETHTxHashIssued(
	uniqETHTx []byte,
) (bool, error) {
	key := append(ethTxHashIssuedPrefix, uniqETHTx...)
	contentBytes, dbErr := db.lvdb.Get(key, nil)
	if dbErr != nil && dbErr != lvdberr.ErrNotFound {
		return false, database.NewDatabaseError(database.IsETHTxHashIssuedError, errors.Wrap(dbErr, "db.lvdb.Get"))
	}
	if len(contentBytes) == 0 {
		return false, nil
	}
	return true, nil
}

func (db *db) CanProcessCIncToken(
	incTokenID common.Hash,
) (bool, error) {
	dBridgeTokenExisted, err := db.IsBridgeTokenExistedByType(incTokenID, false)
	if err != nil {
		return false, database.NewDatabaseError(database.BridgeUnexpectedError, err)
	}
	if dBridgeTokenExisted {
		return false, nil
	}

	cBridgeTokenExisted, err := db.IsBridgeTokenExistedByType(incTokenID, true)
	if err != nil {
		return false, database.NewDatabaseError(database.BridgeUnexpectedError, err)
	}
	privacyCustomTokenExisted := db.PrivacyTokenIDExisted(incTokenID)
	privacyCustomTokenCrossShardExisted := db.PrivacyTokenIDCrossShardExisted(incTokenID)
	if !cBridgeTokenExisted && (privacyCustomTokenExisted || privacyCustomTokenCrossShardExisted) {
		return false, nil
	}
	return true, nil
}

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
		return false, database.NewDatabaseError(database.BridgeUnexpectedError, err)
	}
	if cBridgeTokenExisted {
		fmt.Println("WARNING: inc token was existed in centralized token set")
		return false, nil
	}

	dBridgeTokenExisted, err := db.IsBridgeTokenExistedByType(incTokenID, false)
	if err != nil {
		return false, database.NewDatabaseError(database.BridgeUnexpectedError, err)
	}
	fmt.Println("INFO: whether inc token was existed in decentralized token set: ", dBridgeTokenExisted)
	privacyCustomTokenExisted := db.PrivacyTokenIDExisted(incTokenID)
	privacyCustomTokenCrossShardExisted := db.PrivacyTokenIDCrossShardExisted(incTokenID)
	if !dBridgeTokenExisted && (privacyCustomTokenExisted || privacyCustomTokenCrossShardExisted) {
		fmt.Println("WARNING: failed at condition 1: ", dBridgeTokenExisted, privacyCustomTokenExisted, privacyCustomTokenCrossShardExisted)
		return false, nil
	}

	key := append(decentralizedBridgePrefix, incTokenID[:]...)
	contentBytes, dbErr := db.lvdb.Get(key, nil)
	if dbErr != nil && dbErr != lvdberr.ErrNotFound {
		return false, database.NewDatabaseError(database.BridgeUnexpectedError, dbErr)
	}
	if len(contentBytes) > 0 {
		var bridgeTokenInfo BridgeTokenInfo
		err := json.Unmarshal(contentBytes, &bridgeTokenInfo)
		if err != nil {
			return false, database.NewDatabaseError(database.BridgeUnexpectedError, err)
		}
		if bytes.Equal(bridgeTokenInfo.ExternalTokenID[:], externalTokenID[:]) {
			return true, nil
		}
		fmt.Println("WARNING: failed at condition 2:", bridgeTokenInfo.ExternalTokenID[:], externalTokenID[:])
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
			return false, database.NewDatabaseError(database.BridgeUnexpectedError, err)
		}
		if !bytes.Equal(bridgeTokenInfo.ExternalTokenID, externalTokenID) {
			continue
		}

		fmt.Println("WARNING: failed at condition 3:", bridgeTokenInfo.ExternalTokenID[:], externalTokenID[:])
		return false, nil
	}

	iter.Release()
	err = iter.Error()
	if err != nil && err != lvdberr.ErrNotFound {
		return false, database.NewDatabaseError(database.BridgeUnexpectedError, err)
	}
	// both tokens are not existed -> can create new one
	return true, nil
}

func getBridgePrefix(isCentralized bool) []byte {
	if isCentralized {
		return centralizedBridgePrefix
	}
	return decentralizedBridgePrefix
}

func (db *db) UpdateBridgeTokenInfo(
	incTokenID common.Hash,
	externalTokenID []byte,
	isCentralized bool,
	updatingAmt uint64,
	updateType string,
) error {
	prefix := getBridgePrefix(isCentralized)
	key := append(prefix, incTokenID[:]...)
	bridgeTokenInfoBytes, dbErr := db.lvdb.Get(key, nil)
	if dbErr != nil && dbErr != lvdberr.ErrNotFound {
		return database.NewDatabaseError(database.BridgeUnexpectedError, dbErr)
	}

	var newBridgeTokenInfo BridgeTokenInfo
	if len(bridgeTokenInfoBytes) == 0 {
		newBridgeTokenInfo = BridgeTokenInfo{
			TokenID:         &incTokenID,
			ExternalTokenID: externalTokenID,
			IsCentralized:   isCentralized,
		}
		if updateType == "-" {
			newBridgeTokenInfo.Amount = 0
		} else {
			newBridgeTokenInfo.Amount = updatingAmt
		}
	} else { // found existing bridge token info
		var existingBridgeTokenInfo BridgeTokenInfo
		unmarshalErr := json.Unmarshal(bridgeTokenInfoBytes, &existingBridgeTokenInfo)
		if unmarshalErr != nil {
			return unmarshalErr
		}
		newBridgeTokenInfo = BridgeTokenInfo{
			TokenID:         existingBridgeTokenInfo.TokenID,
			ExternalTokenID: existingBridgeTokenInfo.ExternalTokenID,
			IsCentralized:   existingBridgeTokenInfo.IsCentralized,
		}
		if updateType == "+" {
			newBridgeTokenInfo.Amount = existingBridgeTokenInfo.Amount + updatingAmt
		} else if existingBridgeTokenInfo.Amount <= updatingAmt {
			newBridgeTokenInfo.Amount = 0
		} else {
			newBridgeTokenInfo.Amount = existingBridgeTokenInfo.Amount - updatingAmt
		}
	}

	contentBytes, marshalErr := json.Marshal(newBridgeTokenInfo)
	if marshalErr != nil {
		return marshalErr
	}

	dbErr = db.Put(key, contentBytes)
	if dbErr != nil {
		return database.NewDatabaseError(database.BridgeUnexpectedError, dbErr)
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
		return false, database.NewDatabaseError(database.BridgeUnexpectedError, dbErr)
	}
	if len(tokenInfoBytes) == 0 {
		return false, nil
	}
	return true, nil
}

func (db *db) getBridgeTokensByType(isCentralized bool) ([]*BridgeTokenInfo, error) {
	prefix := getBridgePrefix(isCentralized)
	iter := db.lvdb.NewIterator(util.BytesPrefix(prefix), nil)
	bridgeTokenInfos := []*BridgeTokenInfo{}
	for iter.Next() {
		value := iter.Value()
		itemBytes := make([]byte, len(value))
		copy(itemBytes, value)
		var bridgeTokenInfo BridgeTokenInfo
		err := json.Unmarshal(itemBytes, &bridgeTokenInfo)
		if err != nil {
			return nil, err
		}
		bridgeTokenInfos = append(bridgeTokenInfos, &bridgeTokenInfo)
	}

	iter.Release()
	err := iter.Error()
	if err != nil && err != lvdberr.ErrNotFound {
		return nil, database.NewDatabaseError(database.BridgeUnexpectedError, err)
	}

	return bridgeTokenInfos, nil
}

func (db *db) GetAllBridgeTokens() ([]byte, error) {
	cBridgeTokenInfos, err := db.getBridgeTokensByType(true)
	if err != nil {
		return nil, err
	}
	dBridgeTokenInfos, err := db.getBridgeTokensByType(false)
	if err != nil {
		return nil, err
	}
	allBridgeTokens := append(cBridgeTokenInfos, dBridgeTokenInfos...)
	allBridgeTokensBytes, err := json.Marshal(allBridgeTokens)
	if err != nil {
		return nil, err
	}
	return allBridgeTokensBytes, nil
}

func (db *db) TrackBridgeReqWithStatus(txReqID common.Hash, status byte) error {
	key := append(bridgePrefix, txReqID[:]...)
	return db.Put(key, []byte{status})
}

func (db *db) GetBridgeReqWithStatus(txReqID common.Hash) (byte, error) {
	key := append(bridgePrefix, txReqID[:]...)
	bridgeRedStatusBytes, dbErr := db.lvdb.Get(key, nil)
	if dbErr != nil && dbErr != lvdberr.ErrNotFound {
		return common.BridgeRequestNotFoundStatus, database.NewDatabaseError(database.BridgeUnexpectedError, dbErr)
	}
	if len(bridgeRedStatusBytes) == 0 {
		return common.BridgeRequestNotFoundStatus, nil
	}
	return bridgeRedStatusBytes[0], nil
}
