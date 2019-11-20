package rawdb

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/incdb"
	lvdberr "github.com/syndtr/goleveldb/leveldb/errors"
)

func InsertETHTxHashIssued(db incdb.Database, uniqETHTx []byte) error {
	key := append(ethTxHashIssuedPrefix, uniqETHTx...)
	dbErr := db.Put(key, []byte{1})
	if dbErr != nil {
		return incdb.NewDatabaseError(incdb.InsertETHTxHashIssuedError, errors.Wrap(dbErr, "db.lvdb.put"))
	}
	return nil
}

func IsETHTxHashIssued(db incdb.Database, uniqETHTx []byte) (bool, error) {
	key := append(ethTxHashIssuedPrefix, uniqETHTx...)
	contentBytes, dbErr := db.Get(key)
	if dbErr != nil && dbErr != lvdberr.ErrNotFound {
		return false, incdb.NewDatabaseError(incdb.IsETHTxHashIssuedError, errors.Wrap(dbErr, "db.lvdb.Get"))
	}
	if len(contentBytes) == 0 {
		return false, nil
	}
	return true, nil
}

func CanProcessCIncToken(db incdb.Database, incTokenID common.Hash) (bool, error) {
	dBridgeTokenExisted, err := IsBridgeTokenExistedByType(db, incTokenID, false)
	if err != nil {
		return false, incdb.NewDatabaseError(incdb.BridgeUnexpectedError, err)
	}
	if dBridgeTokenExisted {
		return false, nil
	}

	cBridgeTokenExisted, err := IsBridgeTokenExistedByType(db, incTokenID, true)
	if err != nil {
		return false, incdb.NewDatabaseError(incdb.BridgeUnexpectedError, err)
	}
	privacyCustomTokenExisted := PrivacyTokenIDExisted(db, incTokenID)
	privacyCustomTokenCrossShardExisted := PrivacyTokenIDCrossShardExisted(db, incTokenID)
	if !cBridgeTokenExisted && (privacyCustomTokenExisted || privacyCustomTokenCrossShardExisted) {
		return false, nil
	}
	return true, nil
}

func CanProcessTokenPair(db incdb.Database, externalTokenID []byte, incTokenID common.Hash) (bool, error) {
	if len(externalTokenID) == 0 || len(incTokenID[:]) == 0 {
		return false, nil
	}
	// check incognito bridge token is existed in centralized bridge tokens or not
	cBridgeTokenExisted, err := IsBridgeTokenExistedByType(db, incTokenID, true)
	if err != nil {
		return false, incdb.NewDatabaseError(incdb.BridgeUnexpectedError, err)
	}
	if cBridgeTokenExisted {
		fmt.Println("WARNING: inc token was existed in centralized token set")
		return false, nil
	}

	dBridgeTokenExisted, err := IsBridgeTokenExistedByType(db, incTokenID, false)
	if err != nil {
		return false, incdb.NewDatabaseError(incdb.BridgeUnexpectedError, err)
	}
	fmt.Println("INFO: whether inc token was existed in decentralized token set: ", dBridgeTokenExisted)
	privacyCustomTokenExisted := PrivacyTokenIDExisted(db, incTokenID)
	privacyCustomTokenCrossShardExisted := PrivacyTokenIDCrossShardExisted(db, incTokenID)
	if !dBridgeTokenExisted && (privacyCustomTokenExisted || privacyCustomTokenCrossShardExisted) {
		fmt.Println("WARNING: failed at condition 1: ", dBridgeTokenExisted, privacyCustomTokenExisted, privacyCustomTokenCrossShardExisted)
		return false, nil
	}

	key := append(decentralizedBridgePrefix, incTokenID[:]...)
	contentBytes, dbErr := db.Get(key)
	if dbErr != nil && dbErr != lvdberr.ErrNotFound {
		return false, incdb.NewDatabaseError(incdb.BridgeUnexpectedError, dbErr)
	}
	if len(contentBytes) > 0 {
		var bridgeTokenInfo BridgeTokenInfo
		err := json.Unmarshal(contentBytes, &bridgeTokenInfo)
		if err != nil {
			return false, incdb.NewDatabaseError(incdb.BridgeUnexpectedError, err)
		}
		if bytes.Equal(bridgeTokenInfo.ExternalTokenID[:], externalTokenID[:]) {
			return true, nil
		}
		fmt.Println("WARNING: failed at condition 2:", bridgeTokenInfo.ExternalTokenID[:], externalTokenID[:])
		return false, nil
	}
	// else: could not find incTokenID out
	iter := db.NewIteratorWithPrefix(decentralizedBridgePrefix)
	for iter.Next() {
		value := iter.Value()
		itemBytes := make([]byte, len(value))
		copy(itemBytes, value)
		var bridgeTokenInfo BridgeTokenInfo
		err := json.Unmarshal(itemBytes, &bridgeTokenInfo)
		if err != nil {
			return false, incdb.NewDatabaseError(incdb.BridgeUnexpectedError, err)
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
		return false, incdb.NewDatabaseError(incdb.BridgeUnexpectedError, err)
	}
	// both tokens are not existed -> can create new one
	return true, nil
}

func UpdateBridgeTokenInfo(
	db incdb.Database,
	incTokenID common.Hash,
	externalTokenID []byte,
	isCentralized bool,
	updatingAmt uint64,
	updateType string,
	bd *[]incdb.BatchData,
) error {
	prefix := getBridgePrefix(isCentralized)
	key := append(prefix, incTokenID[:]...)
	bridgeTokenInfoBytes, dbErr := db.Get(key)
	if dbErr != nil && dbErr != lvdberr.ErrNotFound {
		return incdb.NewDatabaseError(incdb.BridgeUnexpectedError, dbErr)
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

	if bd != nil {
		*bd = append(*bd, incdb.BatchData{key, contentBytes})
		return nil
	}
	dbErr = db.Put(key, contentBytes)
	if dbErr != nil {
		return incdb.NewDatabaseError(incdb.BridgeUnexpectedError, dbErr)
	}
	return nil
}

func IsBridgeTokenExistedByType(db incdb.Database, incTokenID common.Hash, isCentralized bool) (bool, error) {
	prefix := getBridgePrefix(isCentralized)
	key := append(prefix, incTokenID[:]...)
	tokenInfoBytes, dbErr := db.Get(key)
	if dbErr != nil && dbErr != lvdberr.ErrNotFound {
		return false, incdb.NewDatabaseError(incdb.BridgeUnexpectedError, dbErr)
	}
	if len(tokenInfoBytes) == 0 {
		return false, nil
	}
	return true, nil
}

func getBridgeTokensByType(db incdb.Database, isCentralized bool) ([]*BridgeTokenInfo, error) {
	prefix := getBridgePrefix(isCentralized)
	iter := db.NewIteratorWithPrefix(prefix)
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
		return nil, incdb.NewDatabaseError(incdb.BridgeUnexpectedError, err)
	}

	return bridgeTokenInfos, nil
}

func GetAllBridgeTokens(db incdb.Database) ([]byte, error) {
	cBridgeTokenInfos, err := getBridgeTokensByType(db, true)
	if err != nil {
		return nil, err
	}
	dBridgeTokenInfos, err := getBridgeTokensByType(db, false)
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

func TrackBridgeReqWithStatus(db incdb.Database, txReqID common.Hash, status byte, bd *[]incdb.BatchData) error {
	key := append(bridgePrefix, txReqID[:]...)

	if bd != nil {
		*bd = append(*bd, incdb.BatchData{key, []byte{status}})
		return nil
	}
	return db.Put(key, []byte{status})
}

func GetBridgeReqWithStatus(db incdb.Database, txReqID common.Hash) (byte, error) {
	key := append(bridgePrefix, txReqID[:]...)
	bridgeRedStatusBytes, dbErr := db.Get(key)
	if dbErr != nil && dbErr != lvdberr.ErrNotFound {
		return common.BridgeRequestNotFoundStatus, incdb.NewDatabaseError(incdb.BridgeUnexpectedError, dbErr)
	}
	if len(bridgeRedStatusBytes) == 0 {
		return common.BridgeRequestNotFoundStatus, nil
	}
	return bridgeRedStatusBytes[0], nil
}
