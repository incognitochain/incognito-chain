package blockchain

import (
	"encoding/json"
	"fmt"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/pkg/errors"
	"math"
	"math/big"
	"math/rand"
	"sort"
	"time"
)

type CurrentPortalState struct {
	CustodianPoolState         map[string]*statedb.CustodianState             // key : hash(beaconHeight || custodian_address)
	WaitingPortingRequests     map[string]*statedb.WaitingPortingRequest      // key : hash(beaconHeight || UniquePortingID)
	WaitingRedeemRequests      map[string]*statedb.WaitingRedeemRequest       // key : hash(beaconHeight || UniqueRedeemID)
	FinalExchangeRatesState    map[string]*statedb.FinalExchangeRatesState    // key : hash(beaconHeight || TxID)
	LiquidateExchangeRatesPool map[string]*statedb.LiquidateExchangeRatesPool // key : hash(beaconHeight || TxID)
	// it used for calculate reward for custodian at the end epoch
	LockedCollateralState *statedb.LockedCollateralState
	//Store temporary exchange rates requests
	ExchangeRatesRequests map[string]*metadata.ExchangeRatesRequestStatus // key : hash(beaconHeight | TxID)
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
	liquidateExchangeRatesPool, err := statedb.GetLiquidateExchangeRatesPool(stateDB, beaconHeight)
	if err != nil {
		return nil, err
	}
	lockedCollateralState, err := statedb.GetLockedCollateralStateByBeaconHeight(stateDB, beaconHeight)
	if err != nil {
		return nil, err
	}

	return &CurrentPortalState{
		CustodianPoolState:         custodianPoolState,
		WaitingPortingRequests:     waitingPortingReqs,
		WaitingRedeemRequests:      waitingRedeemReqs,
		FinalExchangeRatesState:    finalExchangeRates,
		ExchangeRatesRequests:      make(map[string]*metadata.ExchangeRatesRequestStatus),
		LiquidateExchangeRatesPool: liquidateExchangeRatesPool,
		LockedCollateralState:      lockedCollateralState,
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
	err = statedb.StoreBulkWaitingPortingRequests(stateDB, beaconHeight, currentPortalState.WaitingPortingRequests)
	if err != nil {
		return err
	}
	err = statedb.StoreWaitingRedeemRequests(stateDB, beaconHeight, currentPortalState.WaitingRedeemRequests)
	if err != nil {
		return err
	}
	err = statedb.StoreBulkFinalExchangeRatesState(stateDB, beaconHeight, currentPortalState.FinalExchangeRatesState)
	if err != nil {
		return err
	}
	err = statedb.StoreBulkLiquidateExchangeRatesPool(stateDB, beaconHeight, currentPortalState.LiquidateExchangeRatesPool)
	if err != nil {
		return err
	}
	err = statedb.StoreLockedCollateralState(stateDB, beaconHeight, currentPortalState.LockedCollateralState)
	if err != nil {
		return err
	}

	return nil
}

func sortCustodianByAmountAscent(
	metadata metadata.PortalUserRegister,
	custodianState map[string]*statedb.CustodianState,
	custodianStateSlice *[]CustodianStateSlice) {
	//convert to slice

	var result []CustodianStateSlice
	for k, v := range custodianState {
		//check pTokenId, select only ptokenid
		tokenIdExist := false
		for _, remoteAddr := range v.GetRemoteAddresses() {
			if remoteAddr.GetPTokenID() == metadata.PTokenId {
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
}

func pickSingleCustodian(
	metadata metadata.PortalUserRegister,
	exchangeRate *statedb.FinalExchangeRatesState,
	custodianStateSlice []CustodianStateSlice,
	currentPortalState *CurrentPortalState) ([]*statedb.MatchingPortingCustodianDetail, error) {
	//sort random slice
	var cloneCustodianList []CustodianStateSlice
	copy(cloneCustodianList, custodianStateSlice)

	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(cloneCustodianList), func(i, j int) {
		cloneCustodianList[i],
			cloneCustodianList[j] = cloneCustodianList[j],
			cloneCustodianList[i]
	})

	//pToken to PRV
	convertExchangeRatesObj := NewConvertExchangeRatesObject(exchangeRate)
	totalPTokenAfterUp150PercentUnit64 := up150Percent(metadata.RegisterAmount) //return nano pBTC, pBNB

	totalPRV, err := convertExchangeRatesObj.ExchangePToken2PRVByTokenId(metadata.PTokenId, totalPTokenAfterUp150PercentUnit64)

	if err != nil {
		Logger.log.Errorf("Convert PToken is error %v", err)
		return nil, err
	}

	Logger.log.Infof("Porting request, pick single custodian ptoken: %v,  need prv %v for %v ptoken", metadata.PTokenId, totalPRV, metadata.RegisterAmount)

	for _, kv := range cloneCustodianList {
		Logger.log.Infof("Porting request,  pick single custodian address %v, key %v, free collateral: %v", kv.Value.GetIncognitoAddress(), kv.Key, kv.Value.GetFreeCollateral())
		if kv.Value.GetFreeCollateral() > 0 && kv.Value.GetFreeCollateral() >= totalPRV {
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
	currentPortalState *CurrentPortalState,
) ([]*statedb.MatchingPortingCustodianDetail, error) {
	//get multiple custodian
	var holdPToken uint64 = 0
	multipleCustodian := make([]*statedb.MatchingPortingCustodianDetail, 0)

	convertExchangeRatesObj := NewConvertExchangeRatesObject(exchangeRate)

	for i := len(custodianStateSlice) - 1; i >= 0; i-- {
		custodianItem := custodianStateSlice[i]
		if holdPToken >= metadata.RegisterAmount {
			break
		}
		Logger.log.Infof("Porting request, pick multiple custodian key: %v, has collateral %v", custodianItem.Key, custodianItem.Value.GetFreeCollateral())

		//base on current FreeCollateral find PToken can use
		totalPToken, err := convertExchangeRatesObj.ExchangePRV2PTokenByTokenId(metadata.PTokenId, custodianItem.Value.GetFreeCollateral())
		if err != nil {
			Logger.log.Errorf("Convert PToken is error %v", err)
			return nil, err
		}

		pTokenHolded := down150Percent(totalPToken)

		remainPToken := metadata.RegisterAmount - holdPToken // 1000 - 833 = 167
		if pTokenHolded > remainPToken {
			pTokenHolded = remainPToken
			Logger.log.Infof("Porting request, custodian key: %v, ptoken amount is more larger than remain so custodian can keep ptoken  %v", custodianItem.Key, pTokenHolded)
		} else {
			Logger.log.Infof("Porting request, pick multiple custodian key: %v, can keep ptoken %v", custodianItem.Key, pTokenHolded)
		}

		totalPTokenAfterUp150PercentUnit64 := up150Percent(pTokenHolded)
		totalPRV, err := convertExchangeRatesObj.ExchangePToken2PRVByTokenId(metadata.PTokenId, totalPTokenAfterUp150PercentUnit64)

		if err != nil {
			Logger.log.Errorf("Convert PToken is error %v", err)
			return nil, err
		}

		Logger.log.Infof("Porting request, custodian key: %v, to keep ptoken %v need prv %v", custodianItem.Key, pTokenHolded, totalPRV)

		if custodianItem.Value.GetFreeCollateral() > 0 && custodianItem.Value.GetFreeCollateral() >= totalPRV {

			remoteAddr, err := statedb.GetRemoteAddressByTokenID(custodianItem.Value.GetRemoteAddresses(), metadata.PTokenId)
			if err != nil {
				Logger.log.Errorf("Error when get remote address by tokenID %v", err)
				return nil, err
			}
			multipleCustodian = append(
				multipleCustodian,
				&statedb.MatchingPortingCustodianDetail{
					IncAddress:             custodianItem.Value.GetIncognitoAddress(),
					RemoteAddress:          remoteAddr,
					Amount:                 pTokenHolded,
					LockedAmountCollateral: totalPRV,
					RemainCollateral:       custodianItem.Value.GetFreeCollateral() - totalPRV,
				},
			)

			holdPToken = holdPToken + pTokenHolded

			//update custodian state
			err = UpdateCustodianWithNewAmount(currentPortalState, custodianItem.Key, metadata.PTokenId, pTokenHolded, totalPRV)
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
	custodian.SetFreeCollateral(freeCollateral)

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
	custodian.SetHoldingPublicTokens(holdingPubTokens)

	//update collateral holded
	if custodian.GetLockedAmountCollateral() == nil {
		totalLockedAmountCollateral := make(map[string]uint64)
		totalLockedAmountCollateral[PTokenId] = lockedAmountCollateral
		custodian.SetLockedAmountCollateral(totalLockedAmountCollateral)
	} else {
		lockedAmount := custodian.GetLockedAmountCollateral()
		lockedAmount[PTokenId] = lockedAmount[PTokenId] + lockedAmountCollateral
		custodian.SetLockedAmountCollateral(lockedAmount)
	}

	currentPortalState.CustodianPoolState[custodianKey] = custodian

	return nil
}

func CalculatePortingFees(totalPToken uint64) uint64 {
	result := common.PercentPortingFeeAmount * float64(totalPToken) / 100
	roundNumber := math.Round(result)
	return uint64(roundNumber)
}

func CalMinPortingFee(portingAmountInPToken uint64, tokenSymbol string, exchangeRate *statedb.FinalExchangeRatesState) (uint64, error) {
	convertExchangeRatesObj := NewConvertExchangeRatesObject(exchangeRate)
	portingAmountInPRV, err := convertExchangeRatesObj.ExchangePToken2PRVByTokenId(tokenSymbol, portingAmountInPToken)
	if err != nil {
		Logger.log.Errorf("Error when calculating minimum porting fee %v", err)
		return 0, err
	}

	// can't use big int to calculate porting fee because of common.PercentPortingFeeAmount < 1
	portingFee := uint64(math.Round(float64(portingAmountInPRV) * common.PercentPortingFeeAmount / 100))

	return portingFee, nil
}

func CalMinRedeemFee(redeemAmountInPToken uint64, tokenSymbol string, exchangeRate *statedb.FinalExchangeRatesState) (uint64, error) {
	convertExchangeRatesObj := NewConvertExchangeRatesObject(exchangeRate)
	redeemAmountInPRV, err := convertExchangeRatesObj.ExchangePToken2PRVByTokenId(tokenSymbol, redeemAmountInPToken)
	if err != nil {
		Logger.log.Errorf("Error when calculating minimum redeem fee %v", err)
		return 0, err
	}

	// can't use big int to calculate porting fee because of common.PercentRedeemFeeAmount < 1
	redeemFee :=  uint64(math.Round(float64(redeemAmountInPRV) * common.PercentRedeemFeeAmount / 100))

	return redeemFee, nil
}

/*
	up 150%
*/
func up150Percent(amount uint64) uint64 {
	tmp := new(big.Int).Mul(new(big.Int).SetUint64(amount), new(big.Int).SetUint64(150))
	result := new(big.Int).Div(tmp, new(big.Int).SetUint64(100)).Uint64()
	return result //return nano pBTC, pBNB
}

func down150Percent(amount uint64) uint64 {
	tmp := new(big.Int).Mul(new(big.Int).SetUint64(amount), new(big.Int).SetUint64(100))
	result := new(big.Int).Div(tmp, new(big.Int).SetUint64(150)).Uint64()
	return result
}

func calTotalLiquidationByExchangeRates(RedeemAmount uint64, liquidateExchangeRates statedb.LiquidateExchangeRatesDetail) (uint64, error) {
	//todo: need review divide operator
	// prv  ------   total token
	// ?		     amount token

	if liquidateExchangeRates.HoldAmountPubToken <= 0 {
		return 0, errors.New("Can not divide 0")
	}

	tmp := new(big.Int).Mul(big.NewInt(int64(liquidateExchangeRates.HoldAmountFreeCollateral)), big.NewInt(int64(RedeemAmount)))
	totalPrv := new(big.Int).Div(tmp, big.NewInt(int64(liquidateExchangeRates.HoldAmountPubToken)))
	return totalPrv.Uint64(), nil
}

//check value is tp120 or tp130
func IsTP120(tpValue uint64) (bool, bool) {
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
func detectTopPercentileLiquidation(custodian *statedb.CustodianState, tpList map[string]uint64) (map[string]metadata.LiquidateTopPercentileExchangeRatesDetail, error) {
	if custodian == nil {
		return nil, errors.New("Custodian not found")
	}

	liquidateExchangeRatesList := make(map[string]metadata.LiquidateTopPercentileExchangeRatesDetail)
	for ptoken, tpValue := range tpList {
		if tp20, ok := IsTP120(tpValue); ok {
			if tp20 {
				liquidateExchangeRatesList[ptoken] = metadata.LiquidateTopPercentileExchangeRatesDetail{
					TPKey:                    common.TP120,
					TPValue:                  tpValue,
					HoldAmountFreeCollateral: custodian.GetLockedAmountCollateral()[ptoken],
					HoldAmountPubToken:       custodian.GetHoldingPublicTokens()[ptoken],
				}
			} else {
				liquidateExchangeRatesList[ptoken] = metadata.LiquidateTopPercentileExchangeRatesDetail{
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
func calculateTPRatio(holdPToken map[string]uint64, holdPRV map[string]uint64, finalExchange *statedb.FinalExchangeRatesState) (map[string]uint64, error) {
	result := make(map[string]uint64)
	convertExchangeRatesObj := NewConvertExchangeRatesObject(finalExchange)
	for key, amountPToken := range holdPToken {
		amountPRV, ok := holdPRV[key]
		if !ok {
			return nil, errors.New("Ptoken not found")
		}

		if amountPRV <= 0 || amountPToken <= 0 {
			return nil, errors.New("total PToken of custodian is zero")
		}

		//(1): convert amount PToken to PRV
		amountPTokenConverted, err := convertExchangeRatesObj.ExchangePToken2PRVByTokenId(key, amountPToken)

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
		percentUp := new(big.Int).Mul(big.NewInt(int64(amountPRV)), big.NewInt(100)) //amountPRV * 100 / amountPTokenConverted
		roundNumber := new(big.Int).Div(percentUp, big.NewInt(int64(amountPTokenConverted))) // math.Ceil(float64(percentUp))
		result[key] = roundNumber.Uint64()
	}

	return result, nil
}

func CalAmountNeededDepositLiquidate(custodian *statedb.CustodianState, exchangeRates *statedb.FinalExchangeRatesState, pTokenId string, isFreeCollateralSelected bool) (uint64, uint64, uint64, error) {
	totalPToken := up150Percent(custodian.GetHoldingPublicTokens()[pTokenId])
	convertExchangeRatesObj := NewConvertExchangeRatesObject(exchangeRates)
	totalPRV, err := convertExchangeRatesObj.ExchangePToken2PRVByTokenId(pTokenId, totalPToken)

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

func ValidationExchangeRates(exchangeRates *statedb.FinalExchangeRatesState) error {
	if exchangeRates == nil || exchangeRates.Rates() == nil {
		return errors.New("Exchange rates not found")
	}

	if _, ok := exchangeRates.Rates()[common.PortalBTCIDStr]; !ok {
		return errors.New("BTC rates is not exist")
	}

	if _, ok := exchangeRates.Rates()[common.PortalBNBIDStr]; !ok {
		return errors.New("BNB rates is not exist")
	}

	if _, ok := exchangeRates.Rates()[common.PRVIDStr]; !ok {
		return errors.New("PRV rates is not exist")
	}

	return nil
}

func sortCustodiansByAmountHoldingPubTokenAscent(tokenSymbol string, custodians map[string]*statedb.CustodianState) []*CustodianStateSlice {
	sortedCustodians := make([]*CustodianStateSlice, 0)
	for key, value := range custodians {
		if value.GetHoldingPublicTokens()[tokenSymbol] > 0 {
			item := CustodianStateSlice{
				Key:   key,
				Value: value,
			}
			sortedCustodians = append(sortedCustodians, &item)
		}
	}

	sort.Slice(sortedCustodians, func(i, j int) bool {
		return sortedCustodians[i].Value.GetHoldingPublicTokens()[tokenSymbol] <= sortedCustodians[j].Value.GetHoldingPublicTokens()[tokenSymbol]
	})

	return sortedCustodians
}

func pickupCustodianForRedeem(redeemAmount uint64, tokenID string, portalState *CurrentPortalState) ([]*statedb.MatchingRedeemCustodianDetail, error) {
	custodianPoolState := portalState.CustodianPoolState

	// case 1: pick one custodian
	// filter custodians
	// bigCustodians who holding amount public token greater than or equal to redeem amount
	// smallCustodians who holding amount public token less than redeem amount
	bigCustodians := make(map[string]*statedb.CustodianState, 0)
	bigCustodianKeys := make([]string, 0)
	smallCustodians := make(map[string]*statedb.CustodianState, 0)
	matchedCustodians := make([]*statedb.MatchingRedeemCustodianDetail, 0)

	for key, cus := range custodianPoolState {
		holdingPubTokenAmount := cus.GetHoldingPublicTokens()[tokenID]
		if holdingPubTokenAmount >= redeemAmount {
			bigCustodians[key] = new(statedb.CustodianState)
			bigCustodians[key] = cus
			bigCustodianKeys = append(bigCustodianKeys, key)
		} else if holdingPubTokenAmount > 0 {
			smallCustodians[key] = new(statedb.CustodianState)
			smallCustodians[key] = cus
		}
	}

	// random to pick-up one custodian in bigCustodians
	if len(bigCustodians) > 0 {
		randomIndexCus := rand.Intn(len(bigCustodians))
		custodianKey := bigCustodianKeys[randomIndexCus]
		matchingCustodian := bigCustodians[custodianKey]

		remoteAddr, err := statedb.GetRemoteAddressByTokenID(matchingCustodian.GetRemoteAddresses(), tokenID)
		if err != nil {
			Logger.log.Errorf("Error when get remote address of custodian: %v", err)
			return nil, err
		}
		matchedCustodians = append(
			matchedCustodians,
			statedb.NewMatchingRedeemCustodianDetailWithValue(
				custodianPoolState[custodianKey].GetIncognitoAddress(), remoteAddr, redeemAmount))

		return matchedCustodians, nil
	}

	// case 2: pick-up multiple custodians in smallCustodians
	if len(smallCustodians) == 0 {
		Logger.log.Errorf("there is no custodian in custodian pool")
		return nil, errors.New("there is no custodian in custodian pool")
	}
	// sort smallCustodians by amount holding public token
	sortedCustodianSlice := sortCustodiansByAmountHoldingPubTokenAscent(tokenID, smallCustodians)

	// get custodians util matching full redeemAmount
	totalMatchedAmount := uint64(0)
	for i := len(sortedCustodianSlice) - 1; i >= 0; i-- {
		custodianKey := sortedCustodianSlice[i].Key
		custodianValue := sortedCustodianSlice[i].Value

		matchedAmount := custodianValue.GetHoldingPublicTokens()[tokenID]
		amountNeedToBeMatched := redeemAmount - totalMatchedAmount
		if matchedAmount > amountNeedToBeMatched {
			matchedAmount = amountNeedToBeMatched
		}

		remoteAddr, err := statedb.GetRemoteAddressByTokenID(custodianValue.GetRemoteAddresses(), tokenID)
		if err != nil {
			Logger.log.Errorf("Error when get remote address of custodian: %v", err)
			return nil, err
		}

		matchedCustodians = append(
			matchedCustodians,
			statedb.NewMatchingRedeemCustodianDetailWithValue(
				custodianPoolState[custodianKey].GetIncognitoAddress(), remoteAddr, matchedAmount))

		totalMatchedAmount += matchedAmount
		if totalMatchedAmount >= redeemAmount {
			return matchedCustodians, nil
		}
	}

	Logger.log.Errorf("Not enough amount public token to return user")
	return nil, errors.New("Not enough amount public token to return user")
}

// convertIncPBNBAmountToExternalBNBAmount converts amount in inc chain (decimal 9) to amount in bnb chain (decimal 8)
func convertIncPBNBAmountToExternalBNBAmount(incPBNBAmount int64) int64 {
	return incPBNBAmount / 10 // incPBNBAmount / 1^9 * 1^8
}

// updateFreeCollateralCustodian updates custodian state (amount collaterals) when custodian returns redeemAmount public token to user
func updateFreeCollateralCustodian(custodianState *statedb.CustodianState, redeemAmount uint64, tokenID string, exchangeRate *statedb.FinalExchangeRatesState) (uint64, error) {
	// calculate unlock amount for custodian
	// if custodian returns redeem amount that is all amount holding of token => unlock full amount
	// else => return 120% redeem amount

	convertExchangeRatesObj := NewConvertExchangeRatesObject(exchangeRate)

	unlockedAmount := uint64(0)
	var err error
	if custodianState.GetHoldingPublicTokens()[tokenID] == 0 {
		unlockedAmount = custodianState.GetLockedAmountCollateral()[tokenID]
		lockedAmountTmp := custodianState.GetLockedAmountCollateral()
		lockedAmountTmp[tokenID] = 0
		custodianState.SetLockedAmountCollateral(lockedAmountTmp)
		custodianState.SetFreeCollateral(custodianState.GetFreeCollateral() + unlockedAmount)
	} else {
		tmp := new(big.Int).Mul(new(big.Int).SetUint64(redeemAmount), new(big.Int).SetUint64(common.MinPercentUnlockedCollateralAmount))
		unlockedAmountInPToken := new(big.Int).Div(tmp, new(big.Int).SetUint64(100)).Uint64()
		unlockedAmount, err = convertExchangeRatesObj.ExchangePToken2PRVByTokenId(tokenID, unlockedAmountInPToken)

		if err != nil {
			Logger.log.Errorf("Convert PToken is error %v", err)
			return 0, errors.New("[portal-updateFreeCollateralCustodian] error convert amount ptoken to amount in prv ")
		}

		if unlockedAmount == 0 {
			return 0, errors.New("[portal-updateFreeCollateralCustodian] error convert amount ptoken to amount in prv ")
		}
		if custodianState.GetLockedAmountCollateral()[tokenID] <= unlockedAmount {
			return 0, errors.New("[portal-updateFreeCollateralCustodian] Locked amount must be greater than amount need to unlocked")
		}
		lockedAmountTmp := custodianState.GetLockedAmountCollateral()
		lockedAmountTmp[tokenID] -= unlockedAmount
		custodianState.SetLockedAmountCollateral(lockedAmountTmp)
		custodianState.SetFreeCollateral(custodianState.GetFreeCollateral() + unlockedAmount)
	}
	return unlockedAmount, nil
}

// updateRedeemRequestStatusByRedeemId updates status of redeem request into db
func updateRedeemRequestStatusByRedeemId(redeemID string, newStatus int, db *statedb.StateDB) error {
	redeemRequestBytes, err := statedb.GetPortalRedeemRequestStatus(db, redeemID)
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
	err = statedb.StorePortalRedeemRequestStatus(db, redeemID, newRedeemRequest)
	if err != nil {
		return err
	}
	return nil
}

func updateCustodianStateAfterLiquidateCustodian(custodianState *statedb.CustodianState, mintedAmountInPRV uint64, tokenID string) {
	custodianState.SetTotalCollateral(custodianState.GetTotalCollateral() - mintedAmountInPRV)

	if custodianState.GetHoldingPublicTokens()[tokenID] > 0 {
		lockedAmountTmp := custodianState.GetLockedAmountCollateral()
		lockedAmountTmp[tokenID] -= mintedAmountInPRV
		custodianState.SetLockedAmountCollateral(lockedAmountTmp)
	} else {
		unlockedCollateralAmount := custodianState.GetLockedAmountCollateral()[tokenID] - mintedAmountInPRV
		custodianState.SetFreeCollateral(custodianState.GetFreeCollateral() + unlockedCollateralAmount)
		lockedAmountTmp := custodianState.GetLockedAmountCollateral()
		lockedAmountTmp[tokenID] = 0
		custodianState.SetLockedAmountCollateral(lockedAmountTmp)
	}
}

func updateCustodianStateAfterExpiredPortingReq(
	custodianState *statedb.CustodianState, unlockedAmount uint64, unholdingPublicToken uint64, tokenID string) {

	holdingPubTokenTmp := custodianState.GetHoldingPublicTokens()
	holdingPubTokenTmp[tokenID] -= unholdingPublicToken
	custodianState.SetHoldingPublicTokens(holdingPubTokenTmp)

	custodianState.SetFreeCollateral(custodianState.GetFreeCollateral() + unlockedAmount)

	lockedAmountTmp := custodianState.GetLockedAmountCollateral()
	lockedAmountTmp[tokenID] -= unlockedAmount
	custodianState.SetLockedAmountCollateral(lockedAmountTmp)
}

func removeCustodianFromMatchingPortingCustodians(matchingCustodians []*statedb.MatchingPortingCustodianDetail, custodianIncAddr string) bool {
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

func removeCustodianFromMatchingRedeemCustodians(
	matchingCustodians []*statedb.MatchingRedeemCustodianDetail,
	custodianIncAddr string) ([]*statedb.MatchingRedeemCustodianDetail, bool) {
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

func deleteWaitingRedeemRequest(state *CurrentPortalState, waitingRedeemRequestKey string) {
	delete(state.WaitingRedeemRequests, waitingRedeemRequestKey)
}

func deleteWaitingPortingRequest(state *CurrentPortalState, waitingPortingRequestKey string) {
	delete(state.WaitingPortingRequests, waitingPortingRequestKey)
}

type ConvertExchangeRatesObject struct {
	finalExchangeRates *statedb.FinalExchangeRatesState
}

func NewConvertExchangeRatesObject(finalExchangeRates *statedb.FinalExchangeRatesState) *ConvertExchangeRatesObject {
	return &ConvertExchangeRatesObject{finalExchangeRates: finalExchangeRates}
}

func (c ConvertExchangeRatesObject) ExchangePToken2PRVByTokenId(pTokenId string, value uint64) (uint64, error) {
	switch pTokenId {
	case common.PortalBTCIDStr:
		result, err := c.ExchangeBTC2PRV(value)
		if err != nil {
			return 0, err
		}

		return result, nil
	case common.PortalBNBIDStr:
		result, err := c.ExchangeBNB2PRV(value)
		if err != nil {
			return 0, err
		}

		return result, nil
	}

	return 0, errors.New("Ptoken is not support")
}

func (c *ConvertExchangeRatesObject) ExchangePRV2PTokenByTokenId(pTokenId string, value uint64) (uint64, error) {
	switch pTokenId {
	case common.PortalBTCIDStr:
		return c.ExchangePRV2BTC(value)
	case common.PortalBNBIDStr:
		return c.ExchangePRV2BNB(value)
	}

	return 0, errors.New("Ptoken is not support")
}

func (c *ConvertExchangeRatesObject) convert(value uint64, ratesFrom uint64, RatesTo uint64) (uint64, error) {
	//convert to pusdt
	total := new(big.Int).Mul(big.NewInt(int64(value)), big.NewInt(int64(ratesFrom)))
	pUstd := new(big.Int).Div(total, big.NewInt(int64(math.Pow10(9)))) //value of nanno

	if RatesTo <= 0 {
		return 0, errors.New("Can not divide zero")
	}

	//pusdt -> new coin
	result := new(big.Int).Mul(pUstd, big.NewInt(int64(math.Pow10(9)))) // (total * uint64(math.Pow10(9))) / RatesTo
	roundNumber := new(big.Int).Div(result, big.NewInt(int64(RatesTo))) //round up
	return roundNumber.Uint64(), nil

}

func (c *ConvertExchangeRatesObject) ExchangeBTC2PRV(value uint64) (uint64, error) {
	//input : nano
	//todo: check rates exist
	BTCRates := c.finalExchangeRates.Rates()[common.PortalBTCIDStr].Amount //return nano pUSDT
	PRVRates := c.finalExchangeRates.Rates()[common.PRVIDStr].Amount       //return nano pUSDT
	valueExchange, err := c.convert(value, BTCRates, PRVRates)

	if err != nil {
		return 0, err
	}

	Logger.log.Infof("================ Convert, BTC %d 2 PRV with BTCRates %d PRVRates %d , result %d", value, BTCRates, PRVRates, valueExchange)

	//nano
	return valueExchange, nil
}

func (c *ConvertExchangeRatesObject) ExchangeBNB2PRV(value uint64) (uint64, error) {
	BNBRates := c.finalExchangeRates.Rates()[common.PortalBNBIDStr].Amount
	PRVRates := c.finalExchangeRates.Rates()[common.PRVIDStr].Amount

	valueExchange, err := c.convert(value, BNBRates, PRVRates)

	if err != nil {
		return 0, err
	}

	Logger.log.Infof("================ Convert, BNB %v 2 PRV with BNBRates %v PRVRates %v, result %v", value, BNBRates, PRVRates, valueExchange)

	return valueExchange, nil
}

func (c *ConvertExchangeRatesObject) ExchangePRV2BTC(value uint64) (uint64, error) {
	//input nano
	BTCRates := c.finalExchangeRates.Rates()[common.PortalBTCIDStr].Amount //return nano pUSDT
	PRVRates := c.finalExchangeRates.Rates()[common.PRVIDStr].Amount       //return nano pUSDT

	valueExchange, err := c.convert(value, PRVRates, BTCRates)

	if err != nil {
		return 0, err
	}

	Logger.log.Infof("================ Convert, PRV %v 2 BTC with BTCRates %v PRVRates %v, result %v", value, BTCRates, PRVRates, valueExchange)

	return valueExchange, nil
}

func (c *ConvertExchangeRatesObject) ExchangePRV2BNB(value uint64) (uint64, error) {
	BNBRates := c.finalExchangeRates.Rates()[common.PortalBNBIDStr].Amount
	PRVRates := c.finalExchangeRates.Rates()[common.PRVIDStr].Amount

	valueExchange, err := c.convert(value, PRVRates, BNBRates)
	if err != nil {
		return 0, err
	}
	Logger.log.Infof("================ Convert, PRV %v 2 BNB with BNBRates %v PRVRates %v, result %v", value, BNBRates, PRVRates, valueExchange)
	return valueExchange, nil
}

func updateCurrentPortalStateOfLiquidationExchangeRates(beaconHeight uint64, currentPortalState *CurrentPortalState, custodianKey string, custodianState *statedb.CustodianState, detectTp map[string]metadata.LiquidateTopPercentileExchangeRatesDetail)  {
	//update custodian
	for pTokenId, liquidateTopPercentileExchangeRatesDetail := range detectTp {
		holdingPubTokenTmp := custodianState.GetHoldingPublicTokens()
		holdingPubTokenTmp[pTokenId] -= liquidateTopPercentileExchangeRatesDetail.HoldAmountPubToken
		custodianState.SetHoldingPublicTokens(holdingPubTokenTmp)

		lockedAmountTmp := custodianState.GetLockedAmountCollateral()
		lockedAmountTmp[pTokenId] -= liquidateTopPercentileExchangeRatesDetail.HoldAmountFreeCollateral
		custodianState.SetLockedAmountCollateral(lockedAmountTmp)

		custodianState.SetTotalCollateral(custodianState.GetTotalCollateral() - liquidateTopPercentileExchangeRatesDetail.HoldAmountFreeCollateral)
	}

	currentPortalState.CustodianPoolState[custodianKey] = custodianState
	//end

	//update LiquidateExchangeRates
	liquidateExchangeRatesKey := statedb.GeneratePortalLiquidateExchangeRatesPoolObjectKey(beaconHeight)
	liquidateExchangeRates, ok := currentPortalState.LiquidateExchangeRatesPool[liquidateExchangeRatesKey.String()]

	Logger.log.Infof("update LiquidateExchangeRatesPool with liquidateExchangeRatesKey %v value %#v", liquidateExchangeRatesKey, detectTp)
	if !ok {
		item := make(map[string]statedb.LiquidateExchangeRatesDetail)

		for ptoken, liquidateTopPercentileExchangeRatesDetail := range detectTp {
			item[ptoken] = statedb.LiquidateExchangeRatesDetail{
				HoldAmountFreeCollateral: liquidateTopPercentileExchangeRatesDetail.HoldAmountFreeCollateral,
				HoldAmountPubToken:       liquidateTopPercentileExchangeRatesDetail.HoldAmountPubToken,
			}
		}
		currentPortalState.LiquidateExchangeRatesPool[liquidateExchangeRatesKey.String()] = statedb.NewLiquidateExchangeRatesPoolWithValue(item)
	} else {
		for ptoken, liquidateTopPercentileExchangeRatesDetail := range detectTp {
			if _, ok := liquidateExchangeRates.Rates()[ptoken]; !ok {
				liquidateExchangeRates.Rates()[ptoken] = statedb.LiquidateExchangeRatesDetail{
					HoldAmountFreeCollateral: liquidateTopPercentileExchangeRatesDetail.HoldAmountFreeCollateral,
					HoldAmountPubToken:       liquidateTopPercentileExchangeRatesDetail.HoldAmountPubToken,
				}
			} else {
				liquidateExchangeRates.Rates()[ptoken] = statedb.LiquidateExchangeRatesDetail{
					HoldAmountFreeCollateral: liquidateExchangeRates.Rates()[ptoken].HoldAmountFreeCollateral + liquidateTopPercentileExchangeRatesDetail.HoldAmountFreeCollateral,
					HoldAmountPubToken:       liquidateExchangeRates.Rates()[ptoken].HoldAmountPubToken + liquidateTopPercentileExchangeRatesDetail.HoldAmountPubToken,
				}
			}
		}

		currentPortalState.LiquidateExchangeRatesPool[liquidateExchangeRatesKey.String()] = liquidateExchangeRates
	}
	//end
}

func getTotalLockedCollateralInEpoch(featureStateDB *statedb.StateDB, beaconHeight uint64) (uint64, error){
	currentPortalState, err := InitCurrentPortalStateFromDB(featureStateDB, beaconHeight)
	if err != nil {
		return 0, nil
	}

	return currentPortalState.LockedCollateralState.GetTotalLockedCollateralInEpoch(), nil
}
