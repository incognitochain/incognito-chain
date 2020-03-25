package blockchain

import (
	"encoding/json"
	"fmt"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/incdb/lvdb"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/pkg/errors"
	"math"
	"math/rand"
	"sort"
	"time"
)

const (
	PercentPortingFeeAmount = 0.01
	PercentRedeemFeeAmount  = 0.01
)

type CurrentPortalState struct {
	CustodianPoolState     map[string]*statedb.CustodianState         // key : hash(beaconHeight || custodian_address)
	WaitingPortingRequests map[string]*statedb.PortingRequest         // key : hash(beaconHeight || UniquePortingID)
	WaitingRedeemRequests  map[string]*statedb.WaitingRedeemRequest          // key : hash(beaconHeight || UniqueRedeemID)
	FinalExchangeRates     map[string]*statedb.FinalExchangeRatesState     // key : hash(beaconHeight || TxID)
	LiquidateExchangeRates map[string]*statedb.LiquidateExchangeRates // key : hash(beaconHeight || TxID)
	//Store temporary exchange rates requests
	ExchangeRatesRequests map[string]*statedb.ExchangeRatesRequest // key : hash(beaconHeight | TxID)

	DeletedWaitingRedeemRequests map[string]*statedb.WaitingRedeemRequest // key : hash(beaconHeight || UniqueRedeemID)
}

type CustodianStateSlice struct {
	Key   string
	Value *statedb.CustodianState
}

type RedeemMemoBNB struct {
	RedeemID                  string `json:"RedeemID"`
	CustodianIncognitoAddress string `json:"CustodianIncognitoAddress"`
}

type PortingMemoBNB struct {
	PortingID string `json:"PortingID"`
}

func InitCurrentPortalStateFromDB(
	stateDB *statedb.StateDB,
	beaconHeight uint64,
) (*CurrentPortalState, error) {
	custodianPoolState, err := statedb.GetCustodianPoolState(stateDB, beaconHeight)
	if err != nil {
		return nil, err
	}
	waitingPortingReqs, err := statedb.GetWaitingPortingRequests(stateDB, beaconHeight)
	if err != nil {
		return nil, err
	}
	waitingRedeemReqs, err := statedb.GetWaitingRedeemRequests(stateDB, beaconHeight)
	if err != nil {
		return nil, err
	}

	finalExchangeRates, err := statedb.GetFinalExchangeRatesState(stateDB, beaconHeight)
	if err != nil {
		return nil, err
	}

	liquidateExchangeRates, err := statedb.GetLiquidateExchangeRates(stateDB, beaconHeight)
	if err != nil {
		return nil, err
	}

	return &CurrentPortalState{
		CustodianPoolState:     custodianPoolState,
		WaitingPortingRequests: waitingPortingReqs,
		WaitingRedeemRequests:  waitingRedeemReqs,
		FinalExchangeRates:     finalExchangeRates,
		ExchangeRatesRequests:  make(map[string]*statedb.ExchangeRatesRequest),
		LiquidateExchangeRates: liquidateExchangeRates,
	}, nil
}

func storePortalStateToDB(
	stateDB *statedb.StateDB,
	beaconHeight uint64,
	currentPortalState *CurrentPortalState,
) error {
	err := statedb.StoreCustodianState(stateDB, beaconHeight, currentPortalState.CustodianPoolState)
	if err != nil {
		return err
	}
	err = statedb.StoreWaitingPortingRequests(stateDB, beaconHeight, currentPortalState.WaitingPortingRequests)
	if err != nil {
		return err
	}
	err = statedb.StoreWaitingRedeemRequests(stateDB, beaconHeight, currentPortalState.WaitingRedeemRequests)
	if err != nil {
		return err
	}

	err = statedb.StoreFinalExchangeRatesState(stateDB, beaconHeight, currentPortalState.FinalExchangeRates)
	if err != nil {
		return err
	}

	err = statedb.StoreLiquidateExchangeRates(stateDB, beaconHeight, currentPortalState.LiquidateExchangeRates)
	if err != nil {
		return err
	}

	return nil
}

// TODO: need to move to stateDB
// start
func GetLiquidateTPExchangeRatesByKey(
	db database.DatabaseInterface,
	key []byte,
) (*lvdb.LiquidateTopPercentileExchangeRates, error) {
	item, err := db.GetItemPortalByKey(key)

	if err != nil {
		return nil, err
	}

	var object lvdb.LiquidateTopPercentileExchangeRates

	if item == nil {
		return &object, nil
	}

	//get value via idx
	err = json.Unmarshal(item, &object)
	if err != nil {
		return nil, err
	}

	return &object, nil
}

func GetLiquidateExchangeRatesByKey(
	db database.DatabaseInterface,
	key []byte,
) (*lvdb.LiquidateExchangeRates, error) {
	item, err := db.GetItemPortalByKey(key)

	if err != nil {
		return nil, err
	}

	var object lvdb.LiquidateExchangeRates

	if item == nil {
		return &object, nil
	}

	//get value via idx
	err = json.Unmarshal(item, &object)
	if err != nil {
		return nil, err
	}

	return &object, nil
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

	if finalExchangeRatesItem == nil {
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

	if portingRequest == nil {
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

	if custodianWithdrawItem == nil {
		return &custodianWithdraw, nil
	}

	//get value via idx
	err = json.Unmarshal(custodianWithdrawItem, &custodianWithdraw)
	if err != nil {
		return nil, err
	}

	return &custodianWithdraw, nil
}

func GetCustodianByKey(
	db database.DatabaseInterface,
	key []byte,
) (*lvdb.CustodianState, error) {
	custodianItem, err := db.GetItemPortalByKey(key)

	if err != nil {
		return nil, err
	}

	var custodianState lvdb.CustodianState

	if custodianItem == nil {
		return &custodianState, nil
	}

	//get value via idx
	err = json.Unmarshal(custodianItem, &custodianState)
	if err != nil {
		return nil, err
	}

	return &custodianState, nil
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

// end

func removeWaitingPortingReqByKey(key string, state *CurrentPortalState) bool {
	if state.WaitingPortingRequests[key] != nil {
		delete(state.WaitingPortingRequests, key)
		return true
	}

	return false
}

func sortCustodianByAmountAscent(metadata metadata.PortalUserRegister, custodianState map[string]*statedb.CustodianState, custodianStateSlice *[]CustodianStateSlice) error {
	//convert to slice

	var result []CustodianStateSlice
	for k, v := range custodianState {
		//check pTokenId, select only ptokenid
		tokenIdExist := false
		for _, remoteAddr := range v.GetRemoteAddresses() {
			if remoteAddr.PTokenID == metadata.PTokenId {
				tokenIdExist = true
				break
			}
		}
		if !tokenIdExist {
			continue
		}

		item := CustodianStateSlice{
			Key:   k,
			Value: v,
		}
		result = append(result, item)
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].Value.GetFreeCollateral() <= result[j].Value.GetFreeCollateral()
	})

	*custodianStateSlice = result
	return nil
}

func pickSingleCustodian(
	metadata metadata.PortalUserRegister,
	exchangeRate *statedb.FinalExchangeRatesState,
	custodianStateSlice []CustodianStateSlice,
	currentPortalState *CurrentPortalState) ([]*statedb.MatchingPortingCustodianDetail, error) {
	//sort random slice
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(custodianStateSlice), func(i, j int) {
		custodianStateSlice[i],
			custodianStateSlice[j] = custodianStateSlice[j],
			custodianStateSlice[i]
	})

	//pToken to PRV
	totalPTokenAfterUp150PercentUnit64 := up150Percent(metadata.RegisterAmount) //return nano pBTC, pBNB
	totalPRV, err := exchangeRate.ExchangePToken2PRVByTokenId(metadata.PTokenId, totalPTokenAfterUp150PercentUnit64)

	if err != nil {
		Logger.log.Errorf("Convert PToken is error %v", err)
		return nil, err
	}

	Logger.log.Infof("Porting request, pick single custodian ptoken: %v,  need prv %v for %v ptoken", metadata.PTokenId, totalPRV, metadata.RegisterAmount)

	for _, kv := range custodianStateSlice {
		Logger.log.Infof("Porting request,  pick single custodian key %v, free collateral: %v", kv.Key, kv.Value.FreeCollateral)
		if kv.Value.GetFreeCollateral() >= totalPRV {
			result := make([]*statedb.MatchingPortingCustodianDetail, 1)

			remoteAddr, err := statedb.GetRemoteAddressByTokenID(kv.Value.GetRemoteAddresses(), metadata.PTokenId)
			if err != nil {
				Logger.log.Errorf("Error when get remote address by tokenID %v", err)
				return nil, err
			}
			result[0] = &statedb.MatchingPortingCustodianDetail{
				IncAddress:             kv.Value.GetIncognitoAddress(),
				RemoteAddress:          remoteAddr,
				Amount:                 metadata.RegisterAmount,
				LockedAmountCollateral: totalPRV,
				RemainCollateral:       kv.Value.GetFreeCollateral() - totalPRV,
			}

			//update custodian state
			err = UpdateCustodianWithNewAmount(currentPortalState, kv.Key, metadata.PTokenId, metadata.RegisterAmount, totalPRV)

			if err != nil {
				return nil, err
			}

			return result, nil
		}
	}

	return nil, nil
}

func pickMultipleCustodian(
	metadata metadata.PortalUserRegister,
	exchangeRate *statedb.FinalExchangeRatesState,
	custodianStateSlice []CustodianStateSlice,
	currentPortalState *CurrentPortalState) ([]*statedb.MatchingPortingCustodianDetail, error) {
	//get multiple custodian
	var holdPToken uint64 = 0

	multipleCustodian := make([]*statedb.MatchingPortingCustodianDetail, 0)

	for i := len(custodianStateSlice) - 1; i >= 0; i-- {
		custodianItem := custodianStateSlice[i]
		if holdPToken >= metadata.RegisterAmount {
			break
		}
		Logger.log.Infof("Porting request, pick multiple custodian key: %v, has collateral %v", custodianItem.Key, custodianItem.Value.FreeCollateral)

		//base on current FreeCollateral find PToken can use
		totalPToken, err := exchangeRate.ExchangePRV2PTokenByTokenId(metadata.PTokenId, custodianItem.Value.GetFreeCollateral())
		if err != nil {
			Logger.log.Errorf("Convert PToken is error %v", err)
			return nil, err
		}

		pTokenCanUseUint64 := down150Percent(totalPToken)

		remainPToken := metadata.RegisterAmount - holdPToken // 1000 - 833 = 167
		if pTokenCanUseUint64 > remainPToken {
			pTokenCanUseUint64 = remainPToken
			Logger.log.Infof("Porting request, custodian key: %v, ptoken amount is more larger than remain so custodian can keep ptoken  %v", custodianItem.Key, pTokenCanUseUint64)
		} else {
			Logger.log.Infof("Porting request, pick multiple custodian key: %v, can keep ptoken %v", custodianItem.Key, pTokenCanUseUint64)
		}

		totalPTokenAfterUp150PercentUnit64 := up150Percent(pTokenCanUseUint64)
		totalPRV, err := exchangeRate.ExchangePToken2PRVByTokenId(metadata.PTokenId, totalPTokenAfterUp150PercentUnit64)

		if err != nil {
			Logger.log.Errorf("Convert PToken is error %v", err)
			return nil, err
		}

		Logger.log.Infof("Porting request, custodian key: %v, to keep ptoken %v need prv %v", custodianItem.Key, pTokenCanUseUint64, totalPRV)

		if custodianItem.Value.GetFreeCollateral() >= totalPRV {
			remoteAddr, err := lvdb.GetRemoteAddressByTokenID(custodianItem.Value.GetRemoteAddresses(), metadata.PTokenId)
			if err != nil {
				Logger.log.Errorf("Error when get remote address by tokenID %v", err)
				return nil, err
			}
			multipleCustodian = append(
				multipleCustodian,
				&statedb.MatchingPortingCustodianDetail{
					IncAddress:             custodianItem.Value.GetIncognitoAddress(),
					RemoteAddress:          remoteAddr,
					Amount:                 pTokenCanUseUint64,
					LockedAmountCollateral: totalPRV,
					RemainCollateral:       custodianItem.Value.GetFreeCollateral() - totalPRV,
				},
			)

			holdPToken = holdPToken + pTokenCanUseUint64

			//update custodian state
			err = UpdateCustodianWithNewAmount(currentPortalState, custodianItem.Key, metadata.PTokenId, pTokenCanUseUint64, totalPRV)
			if err != nil {
				return nil, err
			}
		}
	}

	return multipleCustodian, nil
}

func UpdateCustodianWithNewAmount(currentPortalState *CurrentPortalState, custodianKey string, PTokenId string, amountPToken uint64, lockedAmountCollateral uint64) error {
	custodian, ok := currentPortalState.CustodianPoolState[custodianKey]
	if !ok {
		return errors.New("Custodian not found")
	}

	freeCollateral := custodian.GetFreeCollateral() - lockedAmountCollateral

	//update ptoken holded
	holdingPubTokensMapping := make(map[string]uint64)
	if custodian.GetHoldingPublicTokens() == nil {
		holdingPubTokensMapping[PTokenId] = amountPToken
	} else {
		for ptokenId, value := range custodian.GetHoldingPublicTokens() {
			holdingPubTokensMapping[ptokenId] = value + amountPToken
		}
	}
	holdingPubTokens := holdingPubTokensMapping

	//update collateral holded
	totalLockedAmountCollateral := make(map[string]uint64)
	if custodian.GetLockedAmountCollateral() == nil {
		totalLockedAmountCollateral[PTokenId] = lockedAmountCollateral
	} else {
		for ptokenId, value := range custodian.GetLockedAmountCollateral() {
			totalLockedAmountCollateral[ptokenId] = value + lockedAmountCollateral
		}
	}

	custodian.SetFreeCollateral(freeCollateral)
	custodian.SetHoldingPublicTokens(holdingPubTokens)
	custodian.SetLockedAmountCollateral(totalLockedAmountCollateral)

	currentPortalState.CustodianPoolState[custodianKey] = custodian

	return nil
}

func CalculatePortingFees(totalPToken uint64) uint64 {
	result := 0.01 * float64(totalPToken) / 100
	roundNumber := math.Round(result)
	return uint64(roundNumber)
}

func CalMinPortingFee(portingAmountInPToken uint64, tokenSymbol string, exchangeRate *statedb.FinalExchangeRatesState) (uint64, error) {
	portingAmountInPRV, err := exchangeRate.ExchangePToken2PRVByTokenId(tokenSymbol, portingAmountInPToken)
	if err != nil {
		Logger.log.Errorf("Error when calculating minimum porting fee %v", err)
		return 0, err
	}

	portingFee := math.Ceil(float64(portingAmountInPRV) * PercentPortingFeeAmount / 100)
	return uint64(portingFee), nil
}

func calMinRedeemFee(redeemAmountInPToken uint64, tokenSymbol string, exchangeRate *statedb.FinalExchangeRatesState) (uint64, error) {
	redeemAmountInPRV, err := exchangeRate.ExchangePToken2PRVByTokenId(tokenSymbol, redeemAmountInPToken)
	if err != nil {
		Logger.log.Errorf("Error when calculating minimum redeem fee %v", err)
		return 0, err
	}

	redeemFee := math.Ceil(float64(redeemAmountInPRV) * PercentRedeemFeeAmount / 100)
	return uint64(redeemFee), nil
}

/*
	up 150%
*/
func up150Percent(amount uint64) uint64 {
	result := float64(amount) * 1.5 //return nano pBTC, pBNB
	roundNumber := math.Ceil(result)
	return uint64(roundNumber) //return nano pBTC, pBNB
}

func down150Percent(amount uint64) uint64 {
	result := float64(amount) / 1.5
	roundNumber := math.Ceil(result)
	return uint64(roundNumber)
}

func upByPercent(amount uint64, percent int) uint64 {
	result := float64(amount) * (float64(percent) / 100) //return nano pBTC, pBNB
	roundNumber := math.Ceil(result)
	return uint64(roundNumber) //return nano pBTC, pBNB
}

func calTotalLiquidationByExchangeRates(RedeemAmount uint64, liquidateExchangeRates statedb.LiquidateExchangeRatesDetail) (uint64, error) {
	//todo: need review divide operator

	// prv  ------   total token
	// ?		     amount token
	totalPrv := liquidateExchangeRates.HoldAmountFreeCollateral * RedeemAmount / liquidateExchangeRates.HoldAmountPubToken
	return totalPrv, nil
}

//check value is tp120 or tp130
func IsTP120(tpValue int) (bool, bool) {
	if tpValue > common.TP120 && tpValue <= common.TP130 {
		return false, true
	}

	if tpValue <= common.TP120 {
		return true, true
	}

	//not found
	return false, false
}

//filter TP for ptoken each custodian
func detectTopPercentileLiquidation(custodian *statedb.CustodianState, tpList map[string]int) (map[string]lvdb.LiquidateTopPercentileExchangeRatesDetail, error) {
	if custodian == nil {
		return nil, errors.New("Custodian not found")
	}

	liquidateExchangeRatesList := make(map[string]lvdb.LiquidateTopPercentileExchangeRatesDetail)
	for ptoken, tpValue := range tpList {
		if tp20, ok := IsTP120(tpValue); ok {
			if tp20 {
				liquidateExchangeRatesList[ptoken] = lvdb.LiquidateTopPercentileExchangeRatesDetail{
					TPKey:                    common.TP120,
					TPValue:                  tpValue,
					HoldAmountFreeCollateral: custodian.GetLockedAmountCollateral()[ptoken],
					HoldAmountPubToken:       custodian.GetHoldingPublicTokens()[ptoken],
				}
			} else {
				liquidateExchangeRatesList[ptoken] = lvdb.LiquidateTopPercentileExchangeRatesDetail{
					TPKey:                    common.TP130,
					TPValue:                  tpValue,
					HoldAmountFreeCollateral: 0,
					HoldAmountPubToken:       0,
				}
			}
		}
	}

	return liquidateExchangeRatesList, nil
}

//detect tp by hold ptoken and hold prv each custodian
func calculateTPRatio(holdPToken map[string]uint64, holdPRV map[string]uint64, finalExchange *lvdb.FinalExchangeRates) (map[string]int, error) {
	result := make(map[string]int)
	for key, amountPToken := range holdPToken {
		amountPRV, ok := holdPRV[key]
		if !ok {
			return nil, errors.New("Ptoken not found")
		}

		if amountPRV <= 0 || amountPToken <= 0 {
			return nil, errors.New("TokenId is must larger than 0")
		}

		//(1): convert amount PToken to PRV
		amountPTokenConverted, err := finalExchange.ExchangePToken2PRVByTokenId(key, amountPToken)

		if err != nil {
			return nil, errors.New("Exchange rates error")
		}

		//(2): calculate % up-down from amount PRV and (1)
		// total1: total ptoken was converted ex: 1BNB = 1000 PRV
		// total2: total prv (was up 150%)
		// 1500 ------ ?
		//1000 ------ 100%
		// => 1500 * 100 / 1000 = 150%
		if amountPTokenConverted <= 0 {
			return nil, errors.New("Can not divide zero")
		}
		//todo: calculate
		percentUp := amountPRV * 100 / amountPTokenConverted
		roundNumber := math.Ceil(float64(percentUp))
		result[key] = int(roundNumber)
	}

	return result, nil
}

func CalAmountNeededDepositLiquidate(custodian *statedb.CustodianState, exchangeRates *lvdb.FinalExchangeRates, pTokenId string, isFreeCollateralSelected bool) (uint64, uint64, uint64, error) {
	totalPToken := up150Percent(custodian.GetHoldingPublicTokens()[pTokenId])
	totalPRV, err := exchangeRates.ExchangePToken2PRVByTokenId(pTokenId, totalPToken)

	if err != nil {
		return 0, 0, 0, err
	}

	totalAmountNeeded := totalPRV - custodian.GetLockedAmountCollateral()[pTokenId]
	var remainAmountFreeCollateral uint64
	var totalFreeCollateralNeeded uint64

	if isFreeCollateralSelected {
		if custodian.GetFreeCollateral() >= totalAmountNeeded {
			remainAmountFreeCollateral = custodian.GetFreeCollateral() - totalAmountNeeded
			totalFreeCollateralNeeded = totalAmountNeeded
			totalAmountNeeded = 0
		} else {
			remainAmountFreeCollateral = 0
			totalFreeCollateralNeeded = custodian.GetFreeCollateral()
			totalAmountNeeded = totalAmountNeeded - custodian.GetFreeCollateral()
		}

		return totalAmountNeeded, totalFreeCollateralNeeded, remainAmountFreeCollateral, nil
	}

	return totalAmountNeeded, 0, 0, nil
}

func ValidationExchangeRates(exchangeRates *lvdb.FinalExchangeRates) error {
	if exchangeRates == nil || exchangeRates.Rates == nil {
		return errors.New("Exchange rates not found")
	}

	if _, ok := exchangeRates.Rates[common.PortalBTCIDStr]; !ok {
		return errors.New("BTC rates is not exist")
	}

	if _, ok := exchangeRates.Rates[common.PortalBNBIDStr]; !ok {
		return errors.New("BNB rates is not exist")
	}

	if _, ok := exchangeRates.Rates[common.PRVIDStr]; !ok {
		return errors.New("PRV rates is not exist")
	}

	return nil
}

func sortCustodiansByAmountHoldingPubTokenAscent(tokenSymbol string, custodians map[string]*lvdb.CustodianState) ([]*CustodianStateSlice, error) {
	sortedCustodians := make([]*CustodianStateSlice, 0)
	for key, value := range custodians {
		if value.HoldingPubTokens[tokenSymbol] > 0 {
			item := CustodianStateSlice{
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

func pickupCustodianForRedeem(redeemAmount uint64, tokenID string, portalState *CurrentPortalState) ([]*lvdb.MatchingRedeemCustodianDetail, error) {
	custodianPoolState := portalState.CustodianPoolState

	// case 1: pick one custodian
	// filter custodians
	// bigCustodians who holding amount public token greater than or equal to redeem amount
	// smallCustodians who holding amount public token less than redeem amount
	bigCustodians := make(map[string]*lvdb.CustodianState, 0)
	bigCustodianKeys := make([]string, 0)
	smallCustodians := make(map[string]*lvdb.CustodianState, 0)
	matchedCustodians := make([]*lvdb.MatchingRedeemCustodianDetail, 0)

	for key, cus := range custodianPoolState {
		if cus.HoldingPubTokens[tokenID] >= redeemAmount {
			bigCustodians[key] = new(lvdb.CustodianState)
			bigCustodians[key] = cus
			bigCustodianKeys = append(bigCustodianKeys, key)
		} else if cus.HoldingPubTokens[tokenID] > 0 {
			smallCustodians[key] = new(lvdb.CustodianState)
			smallCustodians[key] = cus
		}
	}

	// random to pick-up one custodian in bigCustodians
	if len(bigCustodians) > 0 {
		randomIndexCus := rand.Intn(len(bigCustodians))
		custodianKey := bigCustodianKeys[randomIndexCus]
		matchingCustodian := bigCustodians[custodianKey]

		remoteAddr, err := lvdb.GetRemoteAddressByTokenID(matchingCustodian.RemoteAddresses, tokenID)
		if err != nil {
			Logger.log.Errorf("Error when get remote address of custodian: %v", err)
			return nil, err
		}
		matchedCustodians = append(
			matchedCustodians,
			&lvdb.MatchingRedeemCustodianDetail{
				IncAddress:    custodianPoolState[custodianKey].IncognitoAddress,
				Amount:        redeemAmount,
				RemoteAddress: remoteAddr,
			})
		return matchedCustodians, nil
	}

	// case 2: pick-up multiple custodians in smallCustodians
	if len(smallCustodians) == 0 {
		Logger.log.Errorf("there is no custodian in custodian pool")
		return nil, errors.New("there is no custodian in custodian pool")
	}
	// sort smallCustodians by amount holding public token
	sortedCustodianSlice, err := sortCustodiansByAmountHoldingPubTokenAscent(tokenID, smallCustodians)
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

		matchedAmount := custodianValue.HoldingPubTokens[tokenID]
		Logger.log.Errorf("[portal] matchedAmount: %v\n", matchedAmount)
		amountNeedToBeMatched := redeemAmount - totalMatchedAmount
		if matchedAmount > amountNeedToBeMatched {
			matchedAmount = amountNeedToBeMatched
		}

		remoteAddr, err := lvdb.GetRemoteAddressByTokenID(custodianValue.RemoteAddresses, tokenID)
		if err != nil {
			Logger.log.Errorf("Error when get remote address of custodian: %v", err)
			return nil, err
		}

		matchedCustodians = append(
			matchedCustodians,
			&lvdb.MatchingRedeemCustodianDetail{
				IncAddress:    custodianPoolState[custodianKey].IncognitoAddress,
				Amount:        matchedAmount,
				RemoteAddress: remoteAddr,
			})

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
	return externalBNBAmount * 10 // externalBNBAmount / 1^8 * 1^9
}

// convertIncPBNBAmountToExternalBNBAmount converts amount in inc chain (decimal 9) to amount in bnb chain (decimal 8)
func convertIncPBNBAmountToExternalBNBAmount(incPBNBAmount int64) int64 {
	return incPBNBAmount / 10 // incPBNBAmount / 1^9 * 1^8
}

// updateFreeCollateralCustodian updates custodian state (amount collaterals) when custodian returns redeemAmount public token to user
func updateFreeCollateralCustodian(custodianState *lvdb.CustodianState, redeemAmount uint64, tokenID string, exchangeRate *lvdb.FinalExchangeRates) (uint64, error) {
	// calculate unlock amount for custodian
	// if custodian returns redeem amount that is all amount holding of token => unlock full amount
	// else => return 120% redeem amount

	unlockedAmount := uint64(0)
	if custodianState.HoldingPubTokens[tokenID] == 0 {
		unlockedAmount = custodianState.LockedAmountCollateral[tokenID]
		custodianState.LockedAmountCollateral[tokenID] = 0
		custodianState.FreeCollateral += unlockedAmount
	} else {
		unlockedAmountInPToken := uint64(math.Floor(float64(redeemAmount) * 1.2))
		unlockedAmount, err := exchangeRate.ExchangePToken2PRVByTokenId(tokenID, unlockedAmountInPToken)

		if err != nil {
			Logger.log.Errorf("Convert PToken is error %v", err)
			return 0, errors.New("[portal-updateFreeCollateralCustodian] error convert amount ptoken to amount in prv ")
		}

		if unlockedAmount == 0 {
			return 0, errors.New("[portal-updateFreeCollateralCustodian] error convert amount ptoken to amount in prv ")
		}
		if custodianState.LockedAmountCollateral[tokenID] <= unlockedAmount {
			return 0, errors.New("[portal-updateFreeCollateralCustodian] Locked amount must be greater than amount need to unlocked")
		}
		custodianState.LockedAmountCollateral[tokenID] -= unlockedAmount
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

func updateCustodianStateAfterLiquidateCustodian(custodianState *lvdb.CustodianState, mintedAmountInPRV uint64, tokenID string) error {
	custodianState.TotalCollateral -= mintedAmountInPRV

	if custodianState.HoldingPubTokens[tokenID] > 0 {
		custodianState.LockedAmountCollateral[tokenID] -= mintedAmountInPRV
	} else {
		unlockedCollateralAmount := custodianState.LockedAmountCollateral[tokenID] - mintedAmountInPRV
		custodianState.FreeCollateral += unlockedCollateralAmount
		custodianState.LockedAmountCollateral[tokenID] = 0
	}
	return nil
}

func updateCustodianStateAfterExpiredPortingReq(
	custodianState *statedb.CustodianState, unlockedAmount uint64, unholdingPublicToken uint64, tokenID string) error {
	custodianState.HoldingPubTokens[tokenID] -= unholdingPublicToken
	custodianState.FreeCollateral += unlockedAmount
	custodianState.LockedAmountCollateral[tokenID] -= unlockedAmount
	return nil
}

func removeCustodianFromMatchingPortingCustodians(matchingCustodians []*lvdb.MatchingPortingCustodianDetail, custodianIncAddr string) bool {
	for i, cus := range matchingCustodians {
		if cus.IncAddress == custodianIncAddr {
			if i == len(matchingCustodians)-1 {
				matchingCustodians = matchingCustodians[:i]
			} else {
				matchingCustodians = append(matchingCustodians[:i], matchingCustodians[i+1:]...)
			}
			return true
		}
	}

	return false
}

func removeCustodianFromMatchingRedeemCustodians(matchingCustodians []*statedb.MatchingRedeemCustodianDetail, custodianIncAddr string) ([]*lvdb.MatchingRedeemCustodianDetail, bool) {
	for i, cus := range matchingCustodians {
		if cus.GetIncognitoAddress() == custodianIncAddr {
			if i == len(matchingCustodians)-1 {
				matchingCustodians = matchingCustodians[:i]
			} else {
				matchingCustodians = append(matchingCustodians[:i], matchingCustodians[i+1:]...)
			}
			return matchingCustodians, true
		}
	}

	return matchingCustodians, false
}
