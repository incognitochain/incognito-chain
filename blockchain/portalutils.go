package blockchain

import (
	"encoding/json"
	"fmt"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/incognitochain/incognito-chain/metadata/rpccaller"
	"github.com/pkg/errors"
	"github.com/tendermint/tendermint/rpc/client"
	"github.com/tendermint/tendermint/types"
	"math"
	"math/big"
	"sort"
)

type CurrentPortalState struct {
	CustodianPoolState         map[string]*statedb.CustodianState        // key : hash(custodian_address)
	WaitingPortingRequests     map[string]*statedb.WaitingPortingRequest // key : hash(UniquePortingID)
	WaitingRedeemRequests      map[string]*statedb.WaitingRedeemRequest  // key : hash(UniqueRedeemID)
	FinalExchangeRatesState    *statedb.FinalExchangeRatesState
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
) (*CurrentPortalState, error) {
	custodianPoolState, err := statedb.GetCustodianPoolState(stateDB)
	if err != nil {
		return nil, err
	}
	waitingPortingReqs, err := statedb.GetWaitingPortingRequests(stateDB)
	if err != nil {
		return nil, err
	}
	waitingRedeemReqs, err := statedb.GetWaitingRedeemRequests(stateDB)
	if err != nil {
		return nil, err
	}
	finalExchangeRates, err := statedb.GetFinalExchangeRatesState(stateDB)
	if err != nil {
		return nil, err
	}
	liquidateExchangeRatesPool, err := statedb.GetLiquidateExchangeRatesPool(stateDB)
	if err != nil {
		return nil, err
	}
	lockedCollateralState, err := statedb.GetLockedCollateralStateByBeaconHeight(stateDB)
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
	currentPortalState *CurrentPortalState,
) error {
	err := statedb.StoreCustodianState(stateDB, currentPortalState.CustodianPoolState)
	if err != nil {
		return err
	}
	err = statedb.StoreBulkWaitingPortingRequests(stateDB, currentPortalState.WaitingPortingRequests)
	if err != nil {
		return err
	}
	err = statedb.StoreWaitingRedeemRequests(stateDB, currentPortalState.WaitingRedeemRequests)
	if err != nil {
		return err
	}
	err = statedb.StoreBulkFinalExchangeRatesState(stateDB, currentPortalState.FinalExchangeRatesState)
	if err != nil {
		return err
	}
	err = statedb.StoreBulkLiquidateExchangeRatesPool(stateDB, currentPortalState.LiquidateExchangeRatesPool)
	if err != nil {
		return err
	}
	err = statedb.StoreLockedCollateralState(stateDB, currentPortalState.LockedCollateralState)
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
		if v.GetRemoteAddresses()[metadata.PTokenId] == "" {
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

func pickUpCustodians(
	metadata metadata.PortalUserRegister,
	exchangeRate *statedb.FinalExchangeRatesState,
	custodianStateSlice []CustodianStateSlice,
	currentPortalState *CurrentPortalState,
) ([]*statedb.MatchingPortingCustodianDetail, error) {
	//get multiple custodian
	custodians := make([]*statedb.MatchingPortingCustodianDetail, 0)
	remainPTokens := metadata.RegisterAmount
	convertExchangeRatesObj := NewConvertExchangeRatesObject(exchangeRate)
	for i := len(custodianStateSlice) - 1; i >= 0; i-- {
		custodianItem := custodianStateSlice[i]
		freeCollaterals := custodianItem.Value.GetFreeCollateral()
		if freeCollaterals == 0 {
			continue
		}

		Logger.log.Infof("Porting request, pick multiple custodian key: %v, has collateral %v", custodianItem.Key, freeCollaterals)

		collateralsInPToken, err := convertExchangeRatesObj.ExchangePRV2PTokenByTokenId(metadata.PTokenId, freeCollaterals)
		if err != nil {
			Logger.log.Errorf("Failed to convert prv collaterals to PToken - with error %v", err)
			return nil, err
		}

		pTokenCustodianCanHold := down150Percent(collateralsInPToken)
		if pTokenCustodianCanHold > remainPTokens {
			pTokenCustodianCanHold = remainPTokens
			Logger.log.Infof("Porting request, custodian key: %v, ptoken amount is more larger than remain so custodian can keep ptoken  %v", custodianItem.Key, pTokenCustodianCanHold)
		} else {
			Logger.log.Infof("Porting request, pick multiple custodian key: %v, can keep ptoken %v", custodianItem.Key, pTokenCustodianCanHold)
		}

		totalPTokenAfterUp150Percent := up150Percent(pTokenCustodianCanHold)
		neededCollaterals, err := convertExchangeRatesObj.ExchangePToken2PRVByTokenId(metadata.PTokenId, totalPTokenAfterUp150Percent)
		if err != nil {
			Logger.log.Errorf("Failed to convert PToken to prv - with error %v", err)
			return nil, err
		}
		Logger.log.Infof("Porting request, custodian key: %v, to keep ptoken %v need prv %v", custodianItem.Key, pTokenCustodianCanHold, neededCollaterals)

		if freeCollaterals >= neededCollaterals {
			remoteAddr := custodianItem.Value.GetRemoteAddresses()[metadata.PTokenId]
			if remoteAddr == "" {
				Logger.log.Errorf("Remote address in tokenID %v of custodian %v is null", metadata.PTokenId, custodianItem.Value.GetIncognitoAddress())
				return nil, fmt.Errorf("Remote address in tokenID %v of custodian %v is null", metadata.PTokenId, custodianItem.Value.GetIncognitoAddress())
			}
			custodians = append(
				custodians,
				&statedb.MatchingPortingCustodianDetail{
					IncAddress:             custodianItem.Value.GetIncognitoAddress(),
					RemoteAddress:          remoteAddr,
					Amount:                 pTokenCustodianCanHold,
					LockedAmountCollateral: neededCollaterals,
					RemainCollateral:       freeCollaterals - neededCollaterals,
				},
			)

			//update custodian state
			err = UpdateCustodianStateAfterMatchingPortingRequest(currentPortalState, custodianItem.Key, metadata.PTokenId, neededCollaterals)
			if err != nil {
				return nil, err
			}
			if pTokenCustodianCanHold == remainPTokens {
				break
			}
			remainPTokens = remainPTokens - pTokenCustodianCanHold
		}
	}
	return custodians, nil
}

func UpdateCustodianStateAfterMatchingPortingRequest(currentPortalState *CurrentPortalState, custodianKey string, PTokenId string, lockedAmountCollateral uint64) error {
	custodian, ok := currentPortalState.CustodianPoolState[custodianKey]
	if !ok {
		return errors.New("Custodian not found")
	}

	freeCollateral := custodian.GetFreeCollateral() - lockedAmountCollateral
	custodian.SetFreeCollateral(freeCollateral)

	// don't update holding public tokens to avoid this custodian match to redeem request before receiving pubtokens from users

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
func UpdateCustodianStateAfterUserRequestPToken(currentPortalState *CurrentPortalState, custodianKey string, PTokenId string, amountPToken uint64) error {
	custodian, ok := currentPortalState.CustodianPoolState[custodianKey]
	if !ok {
		return errors.New("[UpdateCustodianStateAfterUserRequestPToken] Custodian not found")
	}

	holdingPubTokensTmp := custodian.GetHoldingPublicTokens()
	if holdingPubTokensTmp == nil {
		holdingPubTokensTmp = make(map[string]uint64)
		holdingPubTokensTmp[PTokenId] = amountPToken
	} else {
		holdingPubTokensTmp[PTokenId] += amountPToken
	}
	currentPortalState.CustodianPoolState[custodianKey].SetHoldingPublicTokens(holdingPubTokensTmp)
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
	redeemFee := uint64(math.Round(float64(redeemAmountInPRV) * common.PercentRedeemFeeAmount / 100))

	return redeemFee, nil
}

/*
	up 150%
*/
func up150Percent(amount uint64) uint64 {
	tmp := new(big.Int).Mul(new(big.Int).SetUint64(amount), new(big.Int).SetUint64(common.MinPercentLockCollateral))
	result := new(big.Int).Div(tmp, new(big.Int).SetUint64(100)).Uint64()
	return result //return nano pBTC, pBNB
}

func down150Percent(amount uint64) uint64 {
	tmp := new(big.Int).Mul(new(big.Int).SetUint64(amount), new(big.Int).SetUint64(100))
	result := new(big.Int).Div(tmp, new(big.Int).SetUint64(common.MinPercentLockCollateral)).Uint64()
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
func checkTPRatio(tpValue uint64) (bool, bool) {
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
	tpListKeys := make([]string, 0)
	for key := range tpList {
		tpListKeys = append(tpListKeys, key)
	}
	sort.Strings(tpListKeys)
	for _, ptoken := range tpListKeys {
		tpValue := tpList[ptoken]
		if tp20, ok := checkTPRatio(tpValue); ok {
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
			Logger.log.Info("total PToken of custodian is zero")
			return nil, nil
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
		percentUp := new(big.Int).Mul(big.NewInt(int64(amountPRV)), big.NewInt(100))         //amountPRV * 100 / amountPTokenConverted
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

func sortCustodiansByAmountHoldingPubTokenAscent(tokenID string, custodians map[string]*statedb.CustodianState) []*CustodianStateSlice {
	sortedCustodians := make([]*CustodianStateSlice, 0)
	for key, value := range custodians {
		if value.GetHoldingPublicTokens()[tokenID] > 0 {
			item := CustodianStateSlice{
				Key:   key,
				Value: value,
			}
			sortedCustodians = append(sortedCustodians, &item)
		}
	}

	sort.Slice(sortedCustodians, func(i, j int) bool {
		return sortedCustodians[i].Value.GetHoldingPublicTokens()[tokenID] <= sortedCustodians[j].Value.GetHoldingPublicTokens()[tokenID]
	})

	return sortedCustodians
}

func pickupCustodianForRedeem(redeemAmount uint64, tokenID string, portalState *CurrentPortalState) ([]*statedb.MatchingRedeemCustodianDetail, error) {
	custodianPoolState := portalState.CustodianPoolState
	matchedCustodians := make([]*statedb.MatchingRedeemCustodianDetail, 0)

	// sort smallCustodians by amount holding public token
	sortedCustodianSlice := sortCustodiansByAmountHoldingPubTokenAscent(tokenID, custodianPoolState)
	if len(sortedCustodianSlice) == 0 {
		Logger.log.Errorf("There is no suitable custodian in pool for redeem request")
		return nil, errors.New("There is no suitable custodian in pool for redeem request")
	}

	totalMatchedAmount := uint64(0)
	for i := len(sortedCustodianSlice) - 1; i >= 0; i-- {
		custodianKey := sortedCustodianSlice[i].Key
		custodianValue := sortedCustodianSlice[i].Value

		matchedAmount := custodianValue.GetHoldingPublicTokens()[tokenID]
		amountNeedToBeMatched := redeemAmount - totalMatchedAmount
		if matchedAmount > amountNeedToBeMatched {
			matchedAmount = amountNeedToBeMatched
		}

		remoteAddr := custodianValue.GetRemoteAddresses()[tokenID]
		if remoteAddr == "" {
			Logger.log.Errorf("Remote address in tokenID %v of custodian %v is null", tokenID, custodianValue.GetIncognitoAddress())
			return nil, fmt.Errorf("Remote address in tokenID %v of custodian %v is null", tokenID, custodianValue.GetIncognitoAddress())
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

// updateCustodianStateAfterReqUnlockCollateral updates custodian state (amount collaterals) when custodian returns redeemAmount public token to user
func updateCustodianStateAfterReqUnlockCollateral(custodianState *statedb.CustodianState, unlockedAmount uint64, tokenID string) error {
	lockedAmount := custodianState.GetLockedAmountCollateral()
	if lockedAmount == nil {
		return errors.New("[portal-updateCustodianStateAfterReqUnlockCollateral] Locked amount is nil")
	}
	if lockedAmount[tokenID] < unlockedAmount {
		return errors.New("[portal-updateCustodianStateAfterReqUnlockCollateral] Locked amount is less than amount need to unlocked")
	}

	lockedAmount[tokenID] -= unlockedAmount
	custodianState.SetLockedAmountCollateral(lockedAmount)
	custodianState.SetFreeCollateral(custodianState.GetFreeCollateral() + unlockedAmount)
	return nil
}

// CalUnlockCollateralAmount returns unlock collateral amount by percentage of redeem amount
func CalUnlockCollateralAmount(
	portalState *CurrentPortalState,
	custodianStateKey string,
	redeemAmount uint64,
	tokenID string) (uint64, error) {
	custodianState := portalState.CustodianPoolState[custodianStateKey]
	if custodianState == nil {
		Logger.log.Errorf("Custodian not found %v\n", custodianStateKey)
		return 0, fmt.Errorf("Custodian not found %v\n", custodianStateKey)
	}

	totalHoldingPubToken := custodianState.GetHoldingPublicTokens()[tokenID]
	for _, waitingRedeemReq := range portalState.WaitingRedeemRequests {
		for _, cus := range waitingRedeemReq.GetCustodians() {
			if cus.GetIncognitoAddress() == custodianState.GetIncognitoAddress() {
				totalHoldingPubToken += cus.GetAmount()
				break
			}
		}
	}

	totalLockedAmountInWaitingPortings := uint64(0)
	for _, waitingPortingReq := range portalState.WaitingPortingRequests {
		for _, cus := range waitingPortingReq.Custodians() {
			if cus.IncAddress == custodianState.GetIncognitoAddress() {
				totalLockedAmountInWaitingPortings += cus.LockedAmountCollateral
				break
			}
		}
	}

	tmp := new(big.Int).Mul(
		new(big.Int).SetUint64(redeemAmount),
		new(big.Int).SetUint64(custodianState.GetLockedAmountCollateral()[tokenID]-totalLockedAmountInWaitingPortings))
	unlockAmount := new(big.Int).Div(tmp, new(big.Int).SetUint64(totalHoldingPubToken)).Uint64()
	if unlockAmount <= 0 {
		Logger.log.Errorf("Can not calculate unlock amount for custodian %v\n", unlockAmount)
		return 0, errors.New("Can not calculate unlock amount for custodian")
	}
	return unlockAmount, nil
}

func CalUnlockCollateralAmountAfterLiquidation(
	portalState *CurrentPortalState,
	liquidatedCustodianStateKey string,
	amountPubToken uint64,
	tokenID string,
	exchangeRate *statedb.FinalExchangeRatesState) (uint64, uint64, error) {
	totalUnlockCollateralAmount, err := CalUnlockCollateralAmount(portalState, liquidatedCustodianStateKey, amountPubToken, tokenID)
	if err != nil {
		return 0, 0, err
	}
	convertExchangeRatesObj := NewConvertExchangeRatesObject(exchangeRate)
	tmp := new(big.Int).Mul(new(big.Int).SetUint64(amountPubToken), new(big.Int).SetUint64(common.PercentReceivedCollateralAmount))
	liquidatedAmountInPToken := new(big.Int).Div(tmp, new(big.Int).SetUint64(100)).Uint64()
	liquidatedAmountInPRV, err := convertExchangeRatesObj.ExchangePToken2PRVByTokenId(tokenID, liquidatedAmountInPToken)
	if err != nil {
		return 0, 0, err
	}
	if liquidatedAmountInPRV > totalUnlockCollateralAmount {
		liquidatedAmountInPRV = totalUnlockCollateralAmount
	}

	remainUnlockAmountForCustodian := totalUnlockCollateralAmount - liquidatedAmountInPRV
	return liquidatedAmountInPRV, remainUnlockAmountForCustodian, nil
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

func updateCustodianStateAfterLiquidateCustodian(custodianState *statedb.CustodianState, liquidatedAmount uint64, remainUnlockAmountForCustodian uint64, tokenID string) error {
	if custodianState == nil {
		Logger.log.Errorf("[updateCustodianStateAfterLiquidateCustodian] custodian not found")
		return errors.New("[updateCustodianStateAfterLiquidateCustodian] custodian not found")
	}
	if custodianState.GetTotalCollateral() < liquidatedAmount {
		Logger.log.Errorf("[updateCustodianStateAfterLiquidateCustodian] total collateral less than liquidated amount")
		return errors.New("[updateCustodianStateAfterLiquidateCustodian] total collateral less than liquidated amount")
	}
	custodianState.SetTotalCollateral(custodianState.GetTotalCollateral() - liquidatedAmount)

	lockedAmountTmp := custodianState.GetLockedAmountCollateral()
	if lockedAmountTmp[tokenID] < liquidatedAmount+remainUnlockAmountForCustodian {
		Logger.log.Errorf("[updateCustodianStateAfterLiquidateCustodian] locked amount less than total unlock amount")
		return errors.New("[updateCustodianStateAfterLiquidateCustodian] locked amount less than total unlock amount")
	}
	lockedAmountTmp[tokenID] = lockedAmountTmp[tokenID] - liquidatedAmount - remainUnlockAmountForCustodian
	custodianState.SetLockedAmountCollateral(lockedAmountTmp)

	custodianState.SetFreeCollateral(custodianState.GetFreeCollateral() + remainUnlockAmountForCustodian)

	return nil
}

func updateCustodianStateAfterExpiredPortingReq(
	custodianState *statedb.CustodianState, unlockedAmount uint64, tokenID string) {
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
	total := new(big.Int).Mul(new(big.Int).SetUint64(value), new(big.Int).SetUint64(ratesFrom))

	if RatesTo <= 0 {
		return 0, errors.New("Can not divide zero")
	}

	//pusdt -> new coin
	roundNumber := new(big.Int).Div(total, new(big.Int).SetUint64(RatesTo))
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

func updateCurrentPortalStateOfLiquidationExchangeRates(
	currentPortalState *CurrentPortalState,
	custodianKey string,
	custodianState *statedb.CustodianState,
	tpRatios map[string]metadata.LiquidateTopPercentileExchangeRatesDetail,
	remainUnlockAmounts map[string]uint64,
) {
	//update custodian state
	for pTokenId, tpRatioDetail := range tpRatios {
		holdingPubTokenTmp := custodianState.GetHoldingPublicTokens()
		holdingPubTokenTmp[pTokenId] -= tpRatioDetail.HoldAmountPubToken
		custodianState.SetHoldingPublicTokens(holdingPubTokenTmp)

		lockedAmountTmp := custodianState.GetLockedAmountCollateral()
		lockedAmountTmp[pTokenId] = lockedAmountTmp[pTokenId] - tpRatioDetail.HoldAmountFreeCollateral - remainUnlockAmounts[pTokenId]
		custodianState.SetLockedAmountCollateral(lockedAmountTmp)

		custodianState.SetTotalCollateral(custodianState.GetTotalCollateral() - tpRatioDetail.HoldAmountFreeCollateral)
	}

	totalRemainUnlockAmount := uint64(0)
	for _, amount := range remainUnlockAmounts {
		totalRemainUnlockAmount += amount
	}

	custodianState.SetFreeCollateral(custodianState.GetFreeCollateral() + totalRemainUnlockAmount)
	currentPortalState.CustodianPoolState[custodianKey] = custodianState
	//end

	//update LiquidateExchangeRates
	liquidateExchangeRatesKey := statedb.GeneratePortalLiquidateExchangeRatesPoolObjectKey()
	liquidateExchangeRates, ok := currentPortalState.LiquidateExchangeRatesPool[liquidateExchangeRatesKey.String()]

	Logger.log.Infof("update LiquidateExchangeRatesPool with liquidateExchangeRatesKey %v value %#v", liquidateExchangeRatesKey, tpRatios)
	if !ok {
		item := make(map[string]statedb.LiquidateExchangeRatesDetail)

		for ptoken, liquidateTopPercentileExchangeRatesDetail := range tpRatios {
			item[ptoken] = statedb.LiquidateExchangeRatesDetail{
				HoldAmountFreeCollateral: liquidateTopPercentileExchangeRatesDetail.HoldAmountFreeCollateral,
				HoldAmountPubToken:       liquidateTopPercentileExchangeRatesDetail.HoldAmountPubToken,
			}
		}
		currentPortalState.LiquidateExchangeRatesPool[liquidateExchangeRatesKey.String()] = statedb.NewLiquidateExchangeRatesPoolWithValue(item)
	} else {
		for ptoken, liquidateTopPercentileExchangeRatesDetail := range tpRatios {
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

func getTotalLockedCollateralInEpoch(featureStateDB *statedb.StateDB) (uint64, error) {
	currentPortalState, err := InitCurrentPortalStateFromDB(featureStateDB)
	if err != nil {
		return 0, nil
	}

	return currentPortalState.LockedCollateralState.GetTotalLockedCollateralInEpoch(), nil
}

// GetBNBHeader calls RPC to fullnode bnb to get bnb header by block height
func (blockchain *BlockChain) GetBNBHeader(
	blockHeight int64,
) (*types.Header, error) {
	bnbFullNodeAddress := rpccaller.BuildRPCServerAddress(
		blockchain.GetConfig().ChainParams.BNBFullNodeProtocol,
		blockchain.GetConfig().ChainParams.BNBFullNodeHost,
		blockchain.GetConfig().ChainParams.BNBFullNodePort,
	)
	bnbClient := client.NewHTTP(bnbFullNodeAddress, "/websocket")
	result, err := bnbClient.Block(&blockHeight)
	if err != nil {
		Logger.log.Errorf("An error occured during calling status method: %s", err)
		return nil, fmt.Errorf("error occured during calling status method: %s", err)
	}
	return &result.Block.Header, nil
}

// GetBNBHeader calls RPC to fullnode bnb to get bnb data hash in header
func (blockchain *BlockChain) GetBNBDataHash(
	blockHeight int64,
) ([]byte, error) {
	header, err := blockchain.GetBNBHeader(blockHeight)
	if err != nil {
		return nil, err
	}
	if header.DataHash == nil {
		return nil, errors.New("Data hash is nil")
	}
	return header.DataHash, nil
}

// GetBNBHeader calls RPC to fullnode bnb to get latest bnb block height
func (blockchain *BlockChain) GetLatestBNBBlkHeight() (int64, error) {
	bnbFullNodeAddress := rpccaller.BuildRPCServerAddress(
		blockchain.GetConfig().ChainParams.BNBFullNodeProtocol,
		blockchain.GetConfig().ChainParams.BNBFullNodeHost,
		blockchain.GetConfig().ChainParams.BNBFullNodePort)
	bnbClient := client.NewHTTP(bnbFullNodeAddress, "/websocket")
	result, err := bnbClient.Status()
	if err != nil {
		Logger.log.Errorf("An error occured during calling status method: %s", err)
		return 0, fmt.Errorf("error occured during calling status method: %s", err)
	}
	return result.SyncInfo.LatestBlockHeight, nil
}

func calAndCheckTPRatio(
	portalState *CurrentPortalState,
	custodianState *statedb.CustodianState,
	finalExchange *statedb.FinalExchangeRatesState) (map[string]metadata.LiquidateTopPercentileExchangeRatesDetail, error) {
	result := make(map[string]metadata.LiquidateTopPercentileExchangeRatesDetail)
	convertExchangeRatesObj := NewConvertExchangeRatesObject(finalExchange)

	holdingPubToken := custodianState.GetHoldingPublicTokens()
	lockedAmount := custodianState.GetLockedAmountCollateral()

	for _, waitingPortingReq := range portalState.WaitingPortingRequests {
		for _, matchingCus := range waitingPortingReq.Custodians() {
			if matchingCus.IncAddress == custodianState.GetIncognitoAddress() {
				lockedAmount[waitingPortingReq.TokenID()] -= matchingCus.LockedAmountCollateral
				break
			}
		}
	}

	for _, waitingRedeemReq := range portalState.WaitingRedeemRequests {
		for _, matchingCus := range waitingRedeemReq.GetCustodians() {
			if matchingCus.GetIncognitoAddress() == custodianState.GetIncognitoAddress() {
				holdingPubToken[waitingRedeemReq.GetTokenID()] += matchingCus.GetAmount()
				break
			}
		}
	}

	tpListKeys := make([]string, 0)
	for key := range holdingPubToken {
		tpListKeys = append(tpListKeys, key)
	}
	sort.Strings(tpListKeys)
	for _, tokenID := range tpListKeys {
		amountPubToken := holdingPubToken[tokenID]
		amountPRV, ok := lockedAmount[tokenID]
		if !ok {
			Logger.log.Errorf("Invalid locked amount with tokenID %v\n", tokenID)
			return nil, fmt.Errorf("Invalid locked amount with tokenID %v", tokenID)
		}
		if amountPRV <= 0 || amountPubToken <= 0 {
			continue
		}

		// convert amountPubToken to PRV
		amountPTokenInPRV, err := convertExchangeRatesObj.ExchangePToken2PRVByTokenId(tokenID, amountPubToken)
		if err != nil || amountPTokenInPRV == 0 {
			Logger.log.Errorf("Error when convert exchange rate %v\n", err)
			return nil, fmt.Errorf("Error when convert exchange rate %v", err)
		}

		// amountPRV * 100 / amountPTokenInPRV
		tmp := new(big.Int).Mul(new(big.Int).SetUint64(amountPRV), big.NewInt(100))
		percent := new(big.Int).Div(tmp, new(big.Int).SetUint64(amountPTokenInPRV)).Uint64()

		if tp20, ok := checkTPRatio(percent); ok {
			if tp20 {
				result[tokenID] = metadata.LiquidateTopPercentileExchangeRatesDetail{
					TPKey:                    common.TP120,
					TPValue:                  percent,
					HoldAmountFreeCollateral: lockedAmount[tokenID],
					HoldAmountPubToken:       holdingPubToken[tokenID],
				}
			} else {
				result[tokenID] = metadata.LiquidateTopPercentileExchangeRatesDetail{
					TPKey:                    common.TP130,
					TPValue:                  percent,
					HoldAmountFreeCollateral: 0,
					HoldAmountPubToken:       0,
				}
			}
		}
	}

	return result, nil
}

func UpdateCustodianStateAfterRejectRedeemRequestByLiquidation(portalState *CurrentPortalState, rejectedRedeemReq *statedb.WaitingRedeemRequest, beaconHeight uint64) error {
	tokenID := rejectedRedeemReq.GetTokenID()
	for _, matchingCus := range rejectedRedeemReq.GetCustodians() {
		custodianStateKey := statedb.GenerateCustodianStateObjectKey(matchingCus.GetIncognitoAddress())
		custodianStateKeyStr := custodianStateKey.String()
		custodianState := portalState.CustodianPoolState[custodianStateKeyStr]
		if custodianState == nil {
			return fmt.Errorf("Custodian not found %v", custodianStateKeyStr)
		}

		holdPubTokens := custodianState.GetHoldingPublicTokens()
		if holdPubTokens == nil {
			holdPubTokens = make(map[string]uint64, 0)
			holdPubTokens[tokenID] = matchingCus.GetAmount()
		} else {
			holdPubTokens[tokenID] += matchingCus.GetAmount()
		}

		portalState.CustodianPoolState[custodianStateKeyStr].SetHoldingPublicTokens(holdPubTokens)
	}

	return nil
}

func UpdateCustodianRewards(currentPortalState *CurrentPortalState, rewardInfos map[string]*statedb.PortalRewardInfo) {
	for custodianKey, custodianState := range currentPortalState.CustodianPoolState {
		custodianAddr := custodianState.GetIncognitoAddress()
		if rewardInfos[custodianAddr] == nil {
			continue
		}

		custodianReward := custodianState.GetRewardAmount()
		if custodianReward == nil {
			custodianReward = map[string]uint64{}
		}

		for tokenID, amount := range rewardInfos[custodianAddr].GetRewards() {
			custodianReward[tokenID] += amount
		}
		currentPortalState.CustodianPoolState[custodianKey].SetRewardAmount(custodianReward)
	}
}
