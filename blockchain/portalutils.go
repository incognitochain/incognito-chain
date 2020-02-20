package blockchain

import (
	"encoding/json"
	"fmt"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/database"
	"github.com/incognitochain/incognito-chain/database/lvdb"
	"github.com/pkg/errors"
	"strings"
)

type CurrentPortalState struct {
	CustodianPoolState     map[string]*lvdb.CustodianState       // key : beaconHeight || custodian_address
	ExchangeRatesRequests  map[string]*lvdb.ExchangeRatesRequest // key : beaconHeight | TxID
	WaitingPortingRequests map[string]*lvdb.PortingRequest       // key : beaconHeight || UniquePortingID
	WaitingRedeemRequests  map[string]*lvdb.RedeemRequest        // key : beaconHeight || UniqueRedeemID
}

func NewCustodianState(
	incognitoAddress string,
	totalColl uint64,
	freeColl uint64,
	holdingPubTokens map[string]uint64,
	lockedAmountCollateral map[string]uint64,
	remoteAddresses map[string]string,
) (*lvdb.CustodianState, error) {
	return &lvdb.CustodianState{
		IncognitoAddress:       incognitoAddress,
		TotalCollateral:        totalColl,
		FreeCollateral:         freeColl,
		HoldingPubTokens:       holdingPubTokens,
		LockedAmountCollateral: lockedAmountCollateral,
		RemoteAddresses:        remoteAddresses,
	}, nil
}

func NewPortingRequestState(
	uniquePortingID string,
	txReqID common.Hash,
	tokenID string,
	porterAddress string,
	amount uint64,
	custodians map[string]lvdb.MatchingPortingCustodianDetail,
	portingFee uint64,
) (*lvdb.PortingRequest, error) {
	return &lvdb.PortingRequest{
		UniquePortingID: uniquePortingID,
		TxReqID:         txReqID,
		TokenID:         tokenID,
		PorterAddress:   porterAddress,
		Amount:          amount,
		Custodians:      custodians,
		PortingFee:      portingFee,
	}, nil
}

func NewExchangeRatesState(
	senderAddress string,
	rates map[string]lvdb.ExchangeRatesDetail,
) (*lvdb.ExchangeRatesRequest, error) {
	return &lvdb.ExchangeRatesRequest{
		SenderAddress: senderAddress,
		Rates:         rates,
	}, nil
}

func InitCurrentPortalStateFromDB(
	db database.DatabaseInterface,
	beaconHeight uint64,
) (*CurrentPortalState, error) {
	custodianPoolState, err := getCustodianPoolState(db, beaconHeight)
	if err != nil {
		return nil, err
	}
	waitingPortingReqs, err := getWaitingPortingRequests(db, beaconHeight)
	if err != nil {
		return nil, err
	}
	waitingRedeemReqs, err := getWaitingRedeemRequests(db, beaconHeight)
	if err != nil {
		return nil, err
	}

	return &CurrentPortalState{
		CustodianPoolState:     custodianPoolState,
		WaitingPortingRequests: waitingPortingReqs,
		WaitingRedeemRequests:  waitingRedeemReqs,
	}, nil
}

func storePortalStateToDB(
	db database.DatabaseInterface,
	beaconHeight uint64,
	currentPortalState *CurrentPortalState,
) error {
	err := storeCustodianState(db, beaconHeight, currentPortalState.CustodianPoolState)
	if err != nil {
		return err
	}
	err = storeWaitingPortingRequests(db, beaconHeight, currentPortalState.WaitingPortingRequests)
	if err != nil {
		return err
	}
	err = storeWaitingRedeemRequests(db, beaconHeight, currentPortalState.WaitingRedeemRequests)
	if err != nil {
		return err
	}

	return nil
}

// storeCustodianState stores custodian state at beaconHeight
func storeCustodianState(db database.DatabaseInterface,
	beaconHeight uint64,
	custodianState map[string]*lvdb.CustodianState) error {
	for custodianStateKey, custodian := range custodianState {
		newKey := replaceKeyByBeaconHeight(custodianStateKey, beaconHeight)
		custodianBytes, err := json.Marshal(custodian)
		if err != nil {
			return err
		}
		err = db.Put([]byte(newKey), custodianBytes)
		if err != nil {
			return database.NewDatabaseError(database.StoreCustodianDepositStateError, errors.Wrap(err, "db.lvdb.put"))
		}
	}
	return nil
}

// storeWaitingPortingRequests stores waiting porting requests at beaconHeight
func storeWaitingPortingRequests(db database.DatabaseInterface,
	beaconHeight uint64,
	waitingPortingReqs map[string]*lvdb.PortingRequest) error {
	for waitingReqKey, waitingReq := range waitingPortingReqs {
		newKey := replaceKeyByBeaconHeight(waitingReqKey, beaconHeight)
		waitingReqBytes, err := json.Marshal(waitingReq)
		if err != nil {
			return err
		}
		err = db.Put([]byte(newKey), waitingReqBytes)
		if err != nil {
			return database.NewDatabaseError(database.StoreWaitingPortingRequestError, errors.Wrap(err, "db.lvdb.put"))
		}
	}
	return nil
}

// storeWaitingRedeemRequests stores waiting redeem requests at beaconHeight
func storeWaitingRedeemRequests(db database.DatabaseInterface,
	beaconHeight uint64,
	waitingRedeemReqs map[string]*lvdb.RedeemRequest) error {
	for waitingReqKey, waitingReq := range waitingRedeemReqs {
		newKey := replaceKeyByBeaconHeight(waitingReqKey, beaconHeight)
		waitingReqBytes, err := json.Marshal(waitingReq)
		if err != nil {
			return err
		}
		err = db.Put([]byte(newKey), waitingReqBytes)
		if err != nil {
			return database.NewDatabaseError(database.StoreWaitingRedeemRequestError, errors.Wrap(err, "db.lvdb.put"))
		}
	}
	return nil
}

func replaceKeyByBeaconHeight(key string, newBeaconHeight uint64) string {
	parts := strings.Split(key, "-")
	if len(parts) <= 1 {
		return key
	}
	// part beaconHeight
	parts[1] = fmt.Sprintf("%d", newBeaconHeight)
	newKey := ""
	for idx, part := range parts {
		if idx == len(parts)-1 {
			newKey += part
			continue
		}
		newKey += (part + "-")
	}
	return newKey
}

// getCustodianPoolState gets custodian pool state at beaconHeight
func getCustodianPoolState(
	db database.DatabaseInterface,
	beaconHeight uint64,
) (map[string]*lvdb.CustodianState, error) {
	custodianPoolState := make(map[string]*lvdb.CustodianState)
	custodianPoolStateKeysBytes, custodianPoolStateValuesBytes, err := db.GetAllRecordsPortalByPrefix(beaconHeight, lvdb.PortalCustodianStatePrefix)
	if err != nil {
		return nil, err
	}
	for idx, custodianStateKeyBytes := range custodianPoolStateKeysBytes {
		var custodianState lvdb.CustodianState
		err = json.Unmarshal(custodianPoolStateValuesBytes[idx], &custodianState)
		if err != nil {
			return nil, err
		}
		custodianPoolState[string(custodianStateKeyBytes)] = &custodianState
	}
	return custodianPoolState, nil
}


// getWaitingPortingRequests gets waiting porting requests list at beaconHeight
func getWaitingPortingRequests(
	db database.DatabaseInterface,
	beaconHeight uint64,
) (map[string]*lvdb.PortingRequest, error) {
	waitingPortingReqs := make(map[string]*lvdb.PortingRequest)
	waitingPortingReqsKeyBytes, waitingPortingReqsValueBytes, err := db.GetAllRecordsPortalByPrefix(beaconHeight, lvdb.PortalWaitingPortingRequestsPrefix)
	if err != nil {
		return nil, err
	}
	for idx, waitingPortingReqKeyBytes := range waitingPortingReqsKeyBytes {
		var portingReq lvdb.PortingRequest
		err = json.Unmarshal(waitingPortingReqsValueBytes[idx], &portingReq)
		if err != nil {
			return nil, err
		}
		waitingPortingReqs[string(waitingPortingReqKeyBytes)] = &portingReq
	}
	return waitingPortingReqs, nil
}

// getWaitingRedeemRequests gets waiting redeem requests list at beaconHeight
func getWaitingRedeemRequests(
	db database.DatabaseInterface,
	beaconHeight uint64,
) (map[string]*lvdb.RedeemRequest, error) {
	waitingRedeemReqs := make(map[string]*lvdb.RedeemRequest)
	waitingRedeemReqsKeyBytes, waitingRedeemReqsValueBytes, err := db.GetAllRecordsPortalByPrefix(beaconHeight, lvdb.PortalWaitingRedeemRequestsPrefix)
	if err != nil {
		return nil, err
	}
	for idx, waitingRedeemReqKeyBytes := range waitingRedeemReqsKeyBytes {
		var redeemReq lvdb.RedeemRequest
		err = json.Unmarshal(waitingRedeemReqsValueBytes[idx], &redeemReq)
		if err != nil {
			return nil, err
		}
		waitingRedeemReqs[string(waitingRedeemReqKeyBytes)] = &redeemReq
	}
	return waitingRedeemReqs, nil
}

func GetFinalExchangeRatesByKey(
	db database.DatabaseInterface,
	key []byte,
) (*lvdb.FinalExchangeRates, error) {
	finalExchangeRatesItem, err := db.GetItemPortalByPrefix(key)

	if err != nil {
		return nil, err
	}

	var finalExchangeRatesState lvdb.FinalExchangeRates

	//get value via idx
	var slice []byte
	slice = append(slice, finalExchangeRatesItem)

	err = json.Unmarshal(slice, &finalExchangeRatesState)
	if err != nil {
		return nil, err
	}

	return &finalExchangeRatesState, nil
}

func getAmountAdaptable(amount uint64, exchangeRate uint64) (uint64, error) {
	convertPubTokenToPRVFloat64 := (float64(amount) * 1.5) * float64(exchangeRate)
	convertPubTokenToPRVInt64 := uint64(convertPubTokenToPRVFloat64) // 2.2 -> 2

	return convertPubTokenToPRVInt64, nil
}

func getPubTokenByTotalCollateral(total uint64, exchangeRate uint64) (uint64, error) {
	pubToken := float64(total) / float64(exchangeRate) / 1.5
	pubTokenByCollateral := uint64(pubToken) // 2.2 -> 2

	return pubTokenByCollateral, nil
}

func removeWaitingPortingReqByKey(key string, state *CurrentPortalState) bool {
	if state.WaitingPortingRequests[key] != nil {
		delete(state.WaitingPortingRequests, key)
		return true
	}

	return false
}
