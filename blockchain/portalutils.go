package blockchain

import (
	"encoding/json"
	"fmt"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/database"
	"github.com/incognitochain/incognito-chain/database/lvdb"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/pkg/errors"
	"sort"
	"strings"
)

type CurrentPortalState struct {
	CustodianPoolState     map[string]*lvdb.CustodianState       // key : beaconHeight || custodian_address
	ExchangeRatesRequests  map[string]*lvdb.ExchangeRatesRequest // key : beaconHeight | TxID
	WaitingPortingRequests map[string]*lvdb.PortingRequest       // key : beaconHeight || UniquePortingID
	WaitingRedeemRequests  map[string]*lvdb.RedeemRequest        // key : beaconHeight || UniqueRedeemID
	FinalExchangeRates     map[string]*lvdb.FinalExchangeRates      // key : beaconHeight || TxID
}


type CustodianStateSlice struct {
	Key   string
	Value *lvdb.CustodianState
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
	status string,
	beaconHeight uint64,
) (*lvdb.PortingRequest, error) {
	return &lvdb.PortingRequest{
		UniquePortingID: uniquePortingID,
		TxReqID:         txReqID,
		TokenID:         tokenID,
		PorterAddress:   porterAddress,
		Amount:          amount,
		Custodians:      custodians,
		PortingFee:      portingFee,
		Status:          status,
		BeaconHeight:    beaconHeight,
	}, nil
}

func NewExchangeRatesState(
	senderAddress string,
	rates map[string]uint64,
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

	finalExchangeRates, err := getFinalExchangeRates(db, beaconHeight)
	if err != nil {
		return nil, err
	}

	return &CurrentPortalState{
		CustodianPoolState:     custodianPoolState,
		WaitingPortingRequests: waitingPortingReqs,
		WaitingRedeemRequests:  waitingRedeemReqs,
		FinalExchangeRates:     finalExchangeRates,
		ExchangeRatesRequests:  make(map[string]*lvdb.ExchangeRatesRequest),
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

	err = storeFinalExchangeRates(db, beaconHeight, currentPortalState.FinalExchangeRates)
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
		//todo: review new key
		newKey := replaceKeyByBeaconHeight(waitingReqKey, beaconHeight)

		Logger.log.Infof("Porting request, save waiting db with key %v", newKey)

		waitingReqBytes, err := json.Marshal(waitingReq)
		if err != nil {
			return err
		}
		err = db.Put([]byte(waitingReqKey), waitingReqBytes)
		if err != nil {
			return database.NewDatabaseError(database.StoreWaitingPortingRequestError, errors.Wrap(err, "db.lvdb.put"))
		}
	}

	return nil
}

func storeFinalExchangeRates(db database.DatabaseInterface,
	beaconHeight uint64,
	finalExchangeRates map[string]*lvdb.FinalExchangeRates) error {

	Logger.log.Infof("Portal exchange rates, save exchange rates: count final exchange rate %v", len(finalExchangeRates))

	for key, exchangeRates := range finalExchangeRates {
		newKey := replaceKeyByBeaconHeight(key, beaconHeight)

		Logger.log.Infof("Portal exchange rates, generate new key %v", newKey)

		exchangeRatesBytes, err := json.Marshal(exchangeRates)
		if err != nil {
			return err
		}

		err = db.Put([]byte(newKey), exchangeRatesBytes)
		if err != nil {
			return database.NewDatabaseError(database.StoreFinalExchangeRatesStateError, errors.Wrap(err, "db.lvdb.put"))
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

func getFinalExchangeRates(
	db database.DatabaseInterface,
	beaconHeight uint64,
) (map[string]*lvdb.FinalExchangeRates, error) {
	finalExchangeRates := make(map[string]*lvdb.FinalExchangeRates)

	finalExchangeRatesKeysBytes, finalExchangeRatesValueBytes, err := db.GetAllRecordsPortalByPrefix(beaconHeight, lvdb.PortalFinalExchangeRatesPrefix)

	if err != nil {
		return nil, err
	}

	for idx, finalExchangeRatesKeyBytes := range finalExchangeRatesKeysBytes {
		var items lvdb.FinalExchangeRates
		err = json.Unmarshal(finalExchangeRatesValueBytes[idx], &items)
		if err != nil {
			return nil, err
		}
		finalExchangeRates[string(finalExchangeRatesKeyBytes)] = &items
	}

	Logger.log.Infof("Portal exchange rates, init db %v, raw data %v", len(finalExchangeRates), finalExchangeRatesValueBytes)

	return finalExchangeRates, nil
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

	if  finalExchangeRatesItem == nil {
		return &finalExchangeRatesState, nil
	}

	//get value via idx
	err = json.Unmarshal(finalExchangeRatesItem, &finalExchangeRatesState)
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

func removeWaitingPortingReqByKey (key string, state *CurrentPortalState) bool {
	if state.WaitingPortingRequests[key] != nil {
		delete(state.WaitingPortingRequests, key)
		return true
	}

	return false
}

func sortCustodianByAmountAscent(metadata metadata.PortalUserRegister, custodianState map[string]*lvdb.CustodianState, custodianStateSlice *[]CustodianStateSlice) error {
	//convert to slice

	var result []CustodianStateSlice
	for k, v := range custodianState {
		//check pTokenId, select only ptokenid
		_, tokenIdExist := v.RemoteAddresses[metadata.PTokenId]
		if !tokenIdExist {
			continue
		}

		item := CustodianStateSlice {
			Key: k,
			Value: v,
		}
		result = append(result, item)
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].Value.FreeCollateral <= result[j].Value.FreeCollateral
	})

	*custodianStateSlice = result
	return nil
}

func pickSingleCustodian(metadata metadata.PortalUserRegister, exchangeRate *lvdb.FinalExchangeRates, custodianStateSlice []CustodianStateSlice) (map[string]lvdb.MatchingPortingCustodianDetail, error) {
	//pToken to PRV
	totalPTokenAfterUp150Percent := float64(metadata.RegisterAmount) * 1.5
	totalPTokenAfterUp150PercentUnit64 := uint64(totalPTokenAfterUp150Percent)

	totalPRV := exchangeRate.ExchangePToken2PRVByTokenId(metadata.PTokenId, totalPTokenAfterUp150PercentUnit64)

	Logger.log.Infof("Porting request, pick single custodian ptoken: %v, total prv %v", metadata.PTokenId, totalPRV)

	for _, kv := range custodianStateSlice {
		if kv.Value.FreeCollateral >= totalPRV {
			result := make(map[string]lvdb.MatchingPortingCustodianDetail)
			result[kv.Key] = lvdb.MatchingPortingCustodianDetail{
				RemoteAddress: metadata.PTokenAddress,
				Amount: metadata.RegisterAmount,
				LockedAmountCollateral: totalPRV,
			}

			return result, nil
		}
	}

	return  nil, nil
}

func pickMultipleCustodian (metadata metadata.PortalUserRegister, exchangeRate *lvdb.FinalExchangeRates, custodianStateSlice []CustodianStateSlice) (map[string]lvdb.MatchingPortingCustodianDetail, error){
	//get multiple custodian
	var totalPubTokenAfterPick uint64

	multipleCustodian := make(map[string]lvdb.MatchingPortingCustodianDetail)
	for i := len(custodianStateSlice) - 1; i >= 0; i-- {
		custodianItem := custodianStateSlice[i]
		if totalPubTokenAfterPick >= metadata.RegisterAmount {
			break
		}

		totalPTokenExchange := exchangeRate.ExchangePRV2PTokenByTokenId(metadata.PTokenId, custodianItem.Value.FreeCollateral)

		//temp
		totalPTokenAdaptable := float64(totalPTokenExchange) / 1.5
		totalPTokenCustodianCouldBeKept := uint64(totalPTokenAdaptable) //final 2.2 -> 2

		totalPTokenAfterUp150Percent := float64(totalPTokenCustodianCouldBeKept) * 1.5
		totalPTokenAfterUp150PercentUnit64 := uint64(totalPTokenAfterUp150Percent)

		totalPRV := exchangeRate.ExchangePToken2PRVByTokenId(metadata.PTokenId, totalPTokenAfterUp150PercentUnit64) //final

		//verify collateral
		if custodianItem.Value.FreeCollateral >= totalPRV {
			multipleCustodian[custodianItem.Key] = lvdb.MatchingPortingCustodianDetail{
				RemoteAddress: metadata.PTokenAddress,
				Amount: totalPTokenCustodianCouldBeKept,
				LockedAmountCollateral: totalPRV,
			}

			totalPubTokenAfterPick = totalPubTokenAfterPick + totalPTokenCustodianCouldBeKept

			continue
		}

		Logger.log.Errorf("current portal state is nil")
		return nil, errors.New("Pick multiple custodian is fail")
	}

	//verify total amount
	var verifyTotalPubTokenAfterPick uint64
	for _, eachCustodian := range multipleCustodian {
		verifyTotalPubTokenAfterPick = verifyTotalPubTokenAfterPick + eachCustodian.Amount
	}

	if verifyTotalPubTokenAfterPick != metadata.RegisterAmount {
		return nil, errors.New("Total public token do not match")
	}

	return multipleCustodian, nil
}

func calculatePortingFees(totalPToken uint64) uint64  {
	result := 0.01 * float64(totalPToken) / 100
	return uint64(result)
}