package blockchain

import (
	"encoding/json"
	"fmt"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/database"
	"github.com/incognitochain/incognito-chain/database/lvdb"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/pkg/errors"
	"math"
	"math/rand"
	"sort"
	"strings"
	"time"
)

type CurrentPortalState struct {
	CustodianPoolState     map[string]*lvdb.CustodianState       // key : beaconHeight || custodian_address
	ExchangeRatesRequests  map[string]*lvdb.ExchangeRatesRequest // key : beaconHeight | TxID
	WaitingPortingRequests map[string]*lvdb.PortingRequest       // key : beaconHeight || UniquePortingID
	PortingIdRequests 	   map[string]string       			     // key : UniquePortingID
	WaitingRedeemRequests  map[string]*lvdb.RedeemRequest        // key : beaconHeight || UniqueRedeemID
	FinalExchangeRates     map[string]*lvdb.FinalExchangeRates   // key : beaconHeight || TxID
}

type CustodianStateSlice struct {
	Key   string
	Value *lvdb.CustodianState
}

type RedeemMemoBNB struct {
	RedeemID string `json:"RedeemID"`
}

type PortingMemoBNB struct {
	PortingID string `json:"PortingID"`
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
	status int,
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

func NewRedeemRequestState(
	uniqueRedeemID string,
	txReqID common.Hash,
	tokenID string,
	redeemerAddress string,
	redeemerRemoteAddress string,
	redeemAmount uint64,
	custodians map[string]*lvdb.MatchingRedeemCustodianDetail,
	redeemFee uint64,
	beaconHeight uint64,
) (*lvdb.RedeemRequest, error) {
	return &lvdb.RedeemRequest{
		UniqueRedeemID: uniqueRedeemID,
		TxReqID : txReqID,
		TokenID: tokenID,
		RedeemerAddress:redeemerAddress,
		RedeemerRemoteAddress: redeemerRemoteAddress,
		RedeemAmount: redeemAmount,
		Custodians: custodians,
		RedeemFee: redeemFee,
		BeaconHeight: beaconHeight,
	}, nil
}

func NewMatchingRedeemCustodianDetail(
	remoteAddress string,
	amount uint64) (*lvdb.MatchingRedeemCustodianDetail, error){
	return &lvdb.MatchingRedeemCustodianDetail{
		RemoteAddress:remoteAddress,
		Amount:amount,
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

func NewCustodianWithdrawRequest(
	paymentAddress string,
	amount uint64,
	status int,
	remainCustodianFreeCollateral uint64,
) (*lvdb.CustodianWithdrawRequest, error) {
	return &lvdb.CustodianWithdrawRequest{
		PaymentAddress: paymentAddress,
		Amount: amount,
		Status: status,
		RemainCustodianFreeCollateral: remainCustodianFreeCollateral,
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
		PortingIdRequests: 		make(map[string]string),
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
		Logger.log.Infof("Porting request, store custodian key  %v", newKey)
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
		Logger.log.Infof("Porting request, save waiting db with key %v", newKey)

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

func storeFinalExchangeRates(db database.DatabaseInterface,
	beaconHeight uint64,
	finalExchangeRates map[string]*lvdb.FinalExchangeRates) error {
	for key, exchangeRates := range finalExchangeRates {
		newKey := replaceKeyByBeaconHeight(key, beaconHeight)
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

	//note: key for get data
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

	return finalExchangeRates, nil
}

func GetFinalExchangeRatesByKey(
	db database.DatabaseInterface,
	key []byte,
) (*lvdb.FinalExchangeRates, error) {
	finalExchangeRatesItem, err := db.GetItemPortalByKey(key)

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

func GetPortingRequestByKey(
	db database.DatabaseInterface,
	key []byte,
) (*lvdb.PortingRequest, error) {
	portingRequest, err := db.GetItemPortalByKey(key)

	if err != nil {
		return nil, err
	}

	var portingRequestResult lvdb.PortingRequest

	if  portingRequest == nil {
		return &portingRequestResult, nil
	}

	//get value via idx
	err = json.Unmarshal(portingRequest, &portingRequestResult)
	if err != nil {
		return nil, err
	}

	return &portingRequestResult, nil
}

func GetCustodianWithdrawRequestByKey(
	db database.DatabaseInterface,
	key []byte,
) (*lvdb.CustodianWithdrawRequest, error) {
	custodianWithdrawItem, err := db.GetItemPortalByKey(key)

	if err != nil {
		return nil, err
	}

	var custodianWithdraw lvdb.CustodianWithdrawRequest

	if  custodianWithdrawItem == nil {
		return &custodianWithdraw, nil
	}

	//get value via idx
	err = json.Unmarshal(custodianWithdrawItem, &custodianWithdraw)
	if err != nil {
		return nil, err
	}

	return &custodianWithdraw, nil
}

func GetAllPortingRequest(
	db database.DatabaseInterface,
	key []byte,
) (map[string]*lvdb.PortingRequest, error) {
	portingRequest := make(map[string]*lvdb.PortingRequest)
	portingRequestKeysBytes, portingRequestValueBytes, err := db.GetAllRecordsPortalByPrefixWithoutBeaconHeight(key)

	if err != nil {
		return nil, err
	}

	for idx, portingRequestKeyBytes := range portingRequestKeysBytes {
		var items lvdb.PortingRequest
		err = json.Unmarshal(portingRequestValueBytes[idx], &items)
		if err != nil {
			return nil, err
		}
		portingRequest[string(portingRequestKeyBytes)] = &items
	}

	return portingRequest, nil
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

func pickSingleCustodian(metadata metadata.PortalUserRegister, exchangeRate *lvdb.FinalExchangeRates, custodianStateSlice []CustodianStateSlice, currentPortalState *CurrentPortalState) (map[string]lvdb.MatchingPortingCustodianDetail, error) {
	//sort random slice
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(custodianStateSlice), func(i, j int) {
		custodianStateSlice[i],
		custodianStateSlice[j] = custodianStateSlice[j],
		custodianStateSlice[i]
	})

	//pToken to PRV
	totalPTokenAfterUp150Percent := float64(metadata.RegisterAmount) * 1.5 //return nano pBTC, pBNB
	totalPTokenAfterUp150PercentUnit64 := uint64(totalPTokenAfterUp150Percent) //return nano pBTC, pBNB

	totalPRV, err := exchangeRate.ExchangePToken2PRVByTokenId(metadata.PTokenId, totalPTokenAfterUp150PercentUnit64)

	if err != nil {
		Logger.log.Errorf("Convert PToken is error %v", err)
		return nil, err
	}

	Logger.log.Infof("Porting request, pick single custodian ptoken: %v,  need prv %v for %v ptoken",  metadata.PTokenId, totalPRV, metadata.RegisterAmount)

	for _, kv := range custodianStateSlice {
		Logger.log.Infof("Porting request,  pick single custodian key %v, free collateral: %v", kv.Key, kv.Value.FreeCollateral)
		if kv.Value.FreeCollateral >= totalPRV {
			result := make(map[string]lvdb.MatchingPortingCustodianDetail)

			result[kv.Value.IncognitoAddress] = lvdb.MatchingPortingCustodianDetail{
				RemoteAddress: kv.Value.RemoteAddresses[metadata.PTokenId],
				Amount: metadata.RegisterAmount,
				LockedAmountCollateral: totalPRV,
				RemainCollateral: kv.Value.FreeCollateral - totalPRV,
			}

			//update custodian state
			err := UpdateCustodianWithNewAmount(currentPortalState, kv.Key, metadata.PTokenId, metadata.RegisterAmount, totalPRV)

			if err != nil {
				return nil, err
			}

			return result, nil
		}
	}

	return  nil, nil
}

func pickMultipleCustodian (metadata metadata.PortalUserRegister, exchangeRate *lvdb.FinalExchangeRates, custodianStateSlice []CustodianStateSlice, currentPortalState *CurrentPortalState) (map[string]lvdb.MatchingPortingCustodianDetail, error){
	//get multiple custodian
	var holdPToken uint64 = 0

	multipleCustodian := make(map[string]lvdb.MatchingPortingCustodianDetail)

	for i := len(custodianStateSlice) - 1; i >= 0; i-- {
		custodianItem := custodianStateSlice[i]
		if holdPToken >= metadata.RegisterAmount {
			break
		}
		Logger.log.Infof("Porting request, pick multiple custodian key: %v, has collateral %v", custodianItem.Key, custodianItem.Value.FreeCollateral)

		//base on current FreeCollateral find PToken can use
		totalPToken, err := exchangeRate.ExchangePRV2PTokenByTokenId(metadata.PTokenId, custodianItem.Value.FreeCollateral)
		if err != nil {
			Logger.log.Errorf("Convert PToken is error %v", err)
			return nil, err
		}

		pTokenCanUse := float64(totalPToken) / 1.5
		pTokenCanUseUint64 := uint64(pTokenCanUse)

		remainPToken := metadata.RegisterAmount - holdPToken // 1000 - 833 = 167
		if pTokenCanUseUint64 >  remainPToken {
			pTokenCanUseUint64 = remainPToken
			Logger.log.Infof("Porting request, custodian key: %v, ptoken amount is more larger than remain so custodian can keep ptoken  %v", custodianItem.Key, pTokenCanUseUint64)
		} else {
			Logger.log.Infof("Porting request, pick multiple custodian key: %v, can keep ptoken %v", custodianItem.Key, pTokenCanUseUint64)
		}

		totalPTokenAfterUp150Percent := float64(pTokenCanUseUint64) * 1.5
		totalPTokenAfterUp150PercentUnit64 := uint64(totalPTokenAfterUp150Percent)

		totalPRV, err := exchangeRate.ExchangePToken2PRVByTokenId(metadata.PTokenId, totalPTokenAfterUp150PercentUnit64) //final

		if err != nil {
			Logger.log.Errorf("Convert PToken is error %v", err)
			return nil, err
		}

		Logger.log.Infof("Porting request, custodian key: %v, to keep ptoken %v need prv %v", custodianItem.Key, pTokenCanUseUint64, totalPRV)


		if custodianItem.Value.FreeCollateral >= totalPRV {
			multipleCustodian[custodianItem.Value.IncognitoAddress] = lvdb.MatchingPortingCustodianDetail{
				RemoteAddress: custodianItem.Value.RemoteAddresses[metadata.PTokenId],
				Amount: pTokenCanUseUint64,
				LockedAmountCollateral: totalPRV,
				RemainCollateral: custodianItem.Value.FreeCollateral - totalPRV,
			}

			holdPToken = holdPToken + pTokenCanUseUint64

			//update custodian state
			err := UpdateCustodianWithNewAmount(currentPortalState, custodianItem.Key, metadata.PTokenId, pTokenCanUseUint64, totalPRV)
			if err != nil {
				return nil, err
			}
		}
	}

	return multipleCustodian, nil
}

func UpdateCustodianWithNewAmount(currentPortalState *CurrentPortalState, custodianKey string, PTokenId string,  amountPToken uint64, lockedAmountCollateral uint64) error  {
	custodian, ok := currentPortalState.CustodianPoolState[custodianKey]
	if !ok {
		return errors.New("Custodian not found")
	}

	freeCollateral := custodian.FreeCollateral - lockedAmountCollateral

	//update ptoken holded
	holdingPubTokensMapping := make(map[string]uint64)
	if custodian.HoldingPubTokens == nil {
		holdingPubTokensMapping[PTokenId] = amountPToken
	} else {
		for ptokenId, value := range custodian.HoldingPubTokens {
			holdingPubTokensMapping[ptokenId] = value + amountPToken
		}
	}
	holdingPubTokens := holdingPubTokensMapping

	//update collateral holded
	totalLockedAmountCollateral := make(map[string]uint64)
	if custodian.LockedAmountCollateral == nil {
		totalLockedAmountCollateral[PTokenId] = lockedAmountCollateral
	} else {
		for ptokenId, value := range custodian.LockedAmountCollateral {
			totalLockedAmountCollateral[ptokenId] = value + lockedAmountCollateral
		}
	}

	custodian.FreeCollateral = freeCollateral
	custodian.HoldingPubTokens = holdingPubTokens
	custodian.LockedAmountCollateral = totalLockedAmountCollateral

	currentPortalState.CustodianPoolState[custodianKey] = custodian

	return nil
}

func CalculatePortingFees(totalPToken uint64) uint64  {
	result := 0.01 * float64(totalPToken) / 100
	integer, _  := math.Modf(result)
	return  uint64(integer)
}

func ValidationExchangeRates(exchangeRates *lvdb.FinalExchangeRates) error  {
	if exchangeRates == nil || exchangeRates.Rates == nil {
		return errors.New("Exchange rates not found")
	}

	if _, ok := exchangeRates.Rates[metadata.PortalTokenSymbolBTC]; !ok {
		return errors.New("BTC rates is not exist")
	}

	if _, ok := exchangeRates.Rates[metadata.PortalTokenSymbolBNB]; !ok {
		return errors.New("BNB rates is not exist")
	}

	if _, ok := exchangeRates.Rates[metadata.PortalTokenSymbolPRV]; !ok {
		return errors.New("PRV rates is not exist")
	}

	return  nil
}

func sortCustodiansByAmountHoldingPubTokenAscent(tokenSymbol string, custodians map[string]*lvdb.CustodianState) ([]*CustodianStateSlice, error) {
	sortedCustodians := make([]*CustodianStateSlice, 0)
	for key, value := range custodians {
		 if value.HoldingPubTokens[tokenSymbol] > 0 {
			 item := CustodianStateSlice {
				 Key:   key,
				 Value: value,
			 }
			 sortedCustodians = append(sortedCustodians, &item)
		 }
	}

	sort.Slice(sortedCustodians, func(i, j int) bool {
		return sortedCustodians[i].Value.HoldingPubTokens[tokenSymbol] <= sortedCustodians[j].Value.HoldingPubTokens[tokenSymbol]
	})

	return sortedCustodians, nil
}


func pickupCustodianForRedeem(redeemAmount uint64, tokenSymbol string, portalState *CurrentPortalState) (map[string]*lvdb.MatchingRedeemCustodianDetail, error) {
	custodianPoolState := portalState.CustodianPoolState

	// case 1: pick one custodian
	// filter custodians
	// bigCustodians who holding amount public token greater than or equal to redeem amount
	// smallCustodians who holding amount public token less than redeem amount
	bigCustodians := make(map[string]*lvdb.CustodianState, 0)
	bigCustodianKeys  := make([]string, 0)
	smallCustodians := make(map[string]*lvdb.CustodianState, 0)
	matchedCustodians := make(map[string]*lvdb.MatchingRedeemCustodianDetail, 0)

	for key, cus := range custodianPoolState {
		if cus.HoldingPubTokens[tokenSymbol] >= redeemAmount {
			bigCustodians[key] = new(lvdb.CustodianState)
			bigCustodians[key] = cus
			bigCustodianKeys = append(bigCustodianKeys, key)
		} else if cus.HoldingPubTokens[tokenSymbol] > 0 {
			smallCustodians[key] = new(lvdb.CustodianState)
			smallCustodians[key] = cus
		}
	}

	// random to pick-up one custodian in bigCustodians
	if len(bigCustodians) > 0 {
		randomIndexCus := rand.Intn(len(bigCustodians))
		custodianKey := bigCustodianKeys[randomIndexCus]
		matchingCustodian := bigCustodians[custodianKey]

		matchedCustodians[custodianKey] = new(lvdb.MatchingRedeemCustodianDetail)
		matchedCustodians[custodianKey], _ = NewMatchingRedeemCustodianDetail(
			matchingCustodian.RemoteAddresses[tokenSymbol],
			redeemAmount)

		return matchedCustodians, nil
	}

	// case 2: pick-up multiple custodians in smallCustodians
	if len(smallCustodians) == 0 {
		Logger.log.Errorf("there is no custodian in custodian pool")
		return nil, errors.New("there is no custodian in custodian pool")
	}
	// sort smallCustodians by amount holding public token
	sortedCustodianSlice, err := sortCustodiansByAmountHoldingPubTokenAscent(tokenSymbol, smallCustodians)
	if err != nil {
		Logger.log.Errorf("Error when sorting custodians by amount holding public token %v", err)
		return nil, err
	}

	Logger.log.Errorf("[portal] sortedCustodianSlice: %v\n", sortedCustodianSlice)

	// get custodians util matching full redeemAmount
	totalMatchedAmount := uint64(0)
	for i := len(sortedCustodianSlice) - 1; i >= 0; i-- {
		Logger.log.Errorf("[portal] sortedCustodianSlice[i].Value: %v\n", sortedCustodianSlice[i].Value)
		custodianKey := sortedCustodianSlice[i].Key
		custodianValue := sortedCustodianSlice[i].Value

		matchedAmount := custodianValue.HoldingPubTokens[tokenSymbol]
		Logger.log.Errorf("[portal] matchedAmount: %v\n", matchedAmount)
		amountNeedToBeMatched := redeemAmount - totalMatchedAmount
		if matchedAmount >  amountNeedToBeMatched {
			matchedAmount = amountNeedToBeMatched
		}

		matchedCustodians[custodianKey] = new(lvdb.MatchingRedeemCustodianDetail)
		matchedCustodians[custodianKey], _ = NewMatchingRedeemCustodianDetail(
			custodianValue.RemoteAddresses[tokenSymbol],
			matchedAmount)

		totalMatchedAmount += matchedAmount
		Logger.log.Errorf("[portal] totalMatchedAmount: %v\n", totalMatchedAmount)
		if totalMatchedAmount >= redeemAmount {
			return matchedCustodians, nil
		}
	}

	Logger.log.Errorf("Not enough amount public token to return user")
	return nil, errors.New("Not enough amount public token to return user")
}

// convertExternalBNBAmountToIncAmount converts amount in bnb chain (decimal 8) to amount in inc chain (decimal 9)
func convertExternalBNBAmountToIncAmount(externalBNBAmount int64) int64 {
	return externalBNBAmount * 10   // externalBNBAmount / 1^8 * 1^9
}

// convertIncPBNBAmountToExternalBNBAmount converts amount in inc chain (decimal 9) to amount in bnb chain (decimal 8)
func convertIncPBNBAmountToExternalBNBAmount(incPBNBAmount int64) int64 {
	return incPBNBAmount / 10   // incPBNBAmount / 1^9 * 1^8
}

// updateFreeCollateralCustodian updates custodian state (amount collaterals) when custodian returns redeemAmount public token to user
func updateFreeCollateralCustodian(custodianState * lvdb.CustodianState, redeemAmount uint64, tokenSymbol string, exchangeRate *lvdb.FinalExchangeRates) (uint64, error){
	// calculate unlock amount for custodian
	// if custodian returns redeem amount that is all amount holding of token => unlock full amount
	// else => return 120% redeem amount

	unlockedAmount := uint64(0)
	if custodianState.HoldingPubTokens[tokenSymbol] == 0 {
		unlockedAmount = custodianState.LockedAmountCollateral[tokenSymbol]
		custodianState.LockedAmountCollateral[tokenSymbol] = 0
		custodianState.FreeCollateral += unlockedAmount
	} else {
		unlockedAmountInPToken := uint64(math.Floor(float64(redeemAmount) * 1.2))
		Logger.log.Errorf("updateFreeCollateralCustodian - unlockedAmountInPToken: %v\n", unlockedAmountInPToken)
		unlockedAmount, err := exchangeRate.ExchangePToken2PRVByTokenId(tokenSymbol, unlockedAmountInPToken)

		Logger.log.Errorf("updateFreeCollateralCustodian - unlockedAmount: %v\n", unlockedAmount)

		if err != nil {
			Logger.log.Errorf("Convert PToken is error %v", err)
			return 0, errors.New("[portal-updateFreeCollateralCustodian] error convert amount ptoken to amount in prv ")
		}

		if unlockedAmount == 0 {
			return 0, errors.New("[portal-updateFreeCollateralCustodian] error convert amount ptoken to amount in prv ")
		}
		if custodianState.LockedAmountCollateral[tokenSymbol] <= unlockedAmount {
			return 0, errors.New("[portal-updateFreeCollateralCustodian] Locked amount must be greater than amount need to unlocked")
		}
		custodianState.LockedAmountCollateral[tokenSymbol] -= unlockedAmount
		custodianState.FreeCollateral += unlockedAmount
	}
	return unlockedAmount, nil
}

// updateRedeemRequestStatusByRedeemId updates status of redeem request into db
func updateRedeemRequestStatusByRedeemId(redeemID string, newStatus int, db database.DatabaseInterface) error {
	redeemRequestBytes, err := db.GetRedeemRequestByRedeemID(redeemID)
	if err != nil {
		return err
	}
	if len(redeemRequestBytes) == 0 {
		return fmt.Errorf("Not found redeem request from db with redeemId %v\n", redeemID)
	}

	var redeemRequest metadata.PortalRedeemRequestStatus
	err = json.Unmarshal(redeemRequestBytes, &redeemRequest)
	if err != nil {
		return err
	}

	redeemRequest.Status = byte(newStatus)
	newRedeemRequest, err := json.Marshal(redeemRequest)
	if err != nil {
		return err
	}
	redeemRequestKey := lvdb.NewRedeemReqKey(redeemID)
	err = db.StoreRedeemRequest([]byte(redeemRequestKey), newRedeemRequest)
	if err != nil {
		return err
	}
	return nil
}

func updateCustodianStateAfterLiquidateCustodian(custodianState * lvdb.CustodianState, mintedAmountInPRV uint64, tokenSymbol string) error {
	custodianState.TotalCollateral -= mintedAmountInPRV

	if custodianState.HoldingPubTokens[tokenSymbol] > 0 {
		custodianState.LockedAmountCollateral[tokenSymbol] -= mintedAmountInPRV
	} else {
		unlockedCollateralAmount := custodianState.LockedAmountCollateral[tokenSymbol] - mintedAmountInPRV
		custodianState.FreeCollateral += unlockedCollateralAmount
		custodianState.LockedAmountCollateral[tokenSymbol] = 0
	}
	return nil
}

