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
	CustodianPoolState      map[string]*statedb.CustodianState        // key : hash(custodian_address)
	WaitingPortingRequests  map[string]*statedb.WaitingPortingRequest // key : hash(UniquePortingID)
	WaitingRedeemRequests   map[string]*statedb.RedeemRequest         // key : hash(UniqueRedeemID)
	MatchedRedeemRequests   map[string]*statedb.RedeemRequest         // key : hash(UniquePortingID)
	FinalExchangeRatesState *statedb.FinalExchangeRatesState
	LiquidationPool         map[string]*statedb.LiquidationPool // key : hash(beaconHeight || TxID)
	// it used for calculate reward for custodian at the end epoch
	LockedCollateralForRewards *statedb.LockedCollateralState
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
	matchedRedeemReqs, err := statedb.GetMatchedRedeemRequests(stateDB)
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
		MatchedRedeemRequests:      matchedRedeemReqs,
		FinalExchangeRatesState:    finalExchangeRates,
		ExchangeRatesRequests:      make(map[string]*metadata.ExchangeRatesRequestStatus),
		LiquidationPool:            liquidateExchangeRatesPool,
		LockedCollateralForRewards: lockedCollateralState,
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
	err = statedb.StoreMatchedRedeemRequests(stateDB, currentPortalState.MatchedRedeemRequests)
	if err != nil {
		return err
	}
	err = statedb.StoreBulkFinalExchangeRatesState(stateDB, currentPortalState.FinalExchangeRatesState)
	if err != nil {
		return err
	}
	err = statedb.StoreBulkLiquidateExchangeRatesPool(stateDB, currentPortalState.LiquidationPool)
	if err != nil {
		return err
	}
	err = statedb.StoreLockedCollateralState(stateDB, currentPortalState.LockedCollateralForRewards)
	if err != nil {
		return err
	}

	return nil
}

// convertAllFreeCollateralsToUSDT converts all collaterals of custodian to USDT
func convertAllFreeCollateralsToUSDT(convertRateTool *PortalExchangeRateTool, custodian *statedb.CustodianState) (uint64, error) {
	res := uint64(0)
	prvCollateralInUSDT, err := convertRateTool.ConvertToUSDT(common.PRVIDStr, custodian.GetFreeCollateral())
	if err != nil {
		return 0, err
	}
	res += prvCollateralInUSDT

	tokenCollaterals := custodian.GetFreeTokenCollaterals()
	for tokenID, amount := range tokenCollaterals {
		amountInUSDT, err := convertRateTool.ConvertToUSDT(tokenID, amount)
		if err != nil {
			return 0, err
		}

		res += amountInUSDT
	}
	return res, nil
}

func calHoldPubTokenAmountAndLockCollaterals(
	portingAmount uint64,
	totalLockCollateralInUSDT uint64, matchLockCollateralInUSDT uint64,
	convertRateTool *PortalExchangeRateTool, custodianState *statedb.CustodianState) (uint64, uint64, map[string]uint64) {
	// hold public token amount by percent of matchLockCollateralInUSDT
	tmp := new(big.Int).Mul(new(big.Int).SetUint64(matchLockCollateralInUSDT), new(big.Int).SetUint64(portingAmount))
	pubTokenAmountCanBeHold := tmp.Div(tmp, new(big.Int).SetUint64(totalLockCollateralInUSDT)).Uint64()

	lockPRVCollateral := uint64(0)
	lockTokenCollaterals := map[string]uint64{}

	remainLockCollateralInUSDT := matchLockCollateralInUSDT

	// lock collateral PRV first
	freePRVCollateralInUSDT, _ := convertRateTool.ConvertToUSDT(common.PRVIDStr, custodianState.GetFreeCollateral())
	if freePRVCollateralInUSDT >= matchLockCollateralInUSDT {
		lockPRVCollateral, _ = convertRateTool.ConvertFromUSDT(common.PRVIDStr, matchLockCollateralInUSDT)
		return pubTokenAmountCanBeHold, lockPRVCollateral, lockTokenCollaterals
	} else {
		lockPRVCollateral = custodianState.GetFreeCollateral()
		remainLockCollateralInUSDT = matchLockCollateralInUSDT - freePRVCollateralInUSDT
	}

	// lock other token collaterals
	freeTokenCollaterals := custodianState.GetFreeTokenCollaterals()
	sortedTokenIDs := []string{}
	for tokenID := range freeTokenCollaterals {
		sortedTokenIDs = append(sortedTokenIDs, tokenID)
	}
	sort.Strings(sortedTokenIDs)
	for _, tokenID := range sortedTokenIDs {
		amount := freeTokenCollaterals[tokenID]
		if amount == 0 {
			continue
		}
		freeTokenInUSDT, _ := convertRateTool.ConvertToUSDT(tokenID, amount)
		lockTokenCollateralAmt := uint64(0)
		if freeTokenInUSDT >= remainLockCollateralInUSDT {
			lockTokenCollateralAmt, _ = convertRateTool.ConvertFromUSDT(tokenID, remainLockCollateralInUSDT)
			remainLockCollateralInUSDT = 0
		} else {
			lockTokenCollateralAmt = amount
			remainLockCollateralInUSDT -= freeTokenInUSDT
		}
		lockTokenCollaterals[tokenID] = lockTokenCollateralAmt
		if remainLockCollateralInUSDT == 0 {
			break
		}
	}

	return pubTokenAmountCanBeHold, lockPRVCollateral, lockTokenCollaterals
}

func pickUpCustodianForPorting(
	portingAmount uint64, portalTokenID string,
	custodianPool map[string]*statedb.CustodianState,
	exchangeRate *statedb.FinalExchangeRatesState,
	portalParams PortalParams) ([]*statedb.MatchingPortingCustodianDetail, error) {
	if len(custodianPool) == 0 {
		return nil, errors.New("pickUpCustodianForPorting: Custodian pool is empty")
	}
	if exchangeRate == nil {
		return nil, errors.New("pickUpCustodianForPorting: Current exchange rate is nil")
	}

	// convert free collaterals of custodians to usdt to compare and sort descending
	convertRateTool := NewPortalExchangeRateTool(exchangeRate, portalParams.SupportedCollateralTokens)
	type custodianTotalCollateral struct {
		custodianKey string
		amountInUSDT uint64
	}
	sortedCusCollaterals := make([]custodianTotalCollateral, 0)
	for cusKey, cusDetail := range custodianPool {
		if cusDetail.GetRemoteAddresses() != nil && cusDetail.GetRemoteAddresses()[portalTokenID] != "" {
			collateralInUSDT, err := convertAllFreeCollateralsToUSDT(convertRateTool, cusDetail)
			if err != nil {
				return nil, errors.New("pickUpCustodianForPorting: Error when converting free collateral to USDT")
			}

			sortedCusCollaterals = append(sortedCusCollaterals,
				custodianTotalCollateral{
					custodianKey: cusKey,
					amountInUSDT: collateralInUSDT,
				})
		}
	}
	if len(sortedCusCollaterals) == 0 {
		return nil, errors.New("pickUpCustodianForPorting: There is no custodian supply remote address for porting tokenID")
	}
	sort.SliceStable(sortedCusCollaterals, func(i, j int) bool {
		return sortedCusCollaterals[i].amountInUSDT > sortedCusCollaterals[j].amountInUSDT
	})

	// convert porting amount (up to percent) to USDT
	portAmtInUSDT, _ := convertRateTool.ConvertToUSDT(portalTokenID, upPercent(portingAmount, portalParams.MinPercentLockedCollateral))
	fmt.Println("portAmtInUSDT: ", portAmtInUSDT)
	fmt.Println("sortedCusCollaterals: ", sortedCusCollaterals)
	for _, cus := range sortedCusCollaterals {
		fmt.Println("custodianKey: ", cus.custodianKey)
		fmt.Println("amountInUSDT: ", cus.amountInUSDT)
	}

	// choose the custodian that has free collateral
	matchCustodians := make([]*statedb.MatchingPortingCustodianDetail, 0)

	isChooseOneCustodian := false
	if sortedCusCollaterals[0].amountInUSDT >= portAmtInUSDT {
		isChooseOneCustodian = true
	}

	actualHoldPubToken := uint64(0)
	remainPortAmtInUSDT := portAmtInUSDT
	for i, cus := range sortedCusCollaterals {
		pickedCus := cus
		if cus.amountInUSDT > portAmtInUSDT && i != len(sortedCusCollaterals)-1 {
			continue
		} else if cus.amountInUSDT < portAmtInUSDT && isChooseOneCustodian && i > 0 {
			pickedCus = sortedCusCollaterals[i-1]
		}

		custodianState := custodianPool[pickedCus.custodianKey]
		lockPRVCollateral := uint64(0)
		lockTokenColaterals := map[string]uint64{}
		holdPublicToken := uint64(0)

		matchPortAmtInUSDT := pickedCus.amountInUSDT
		if pickedCus.amountInUSDT > remainPortAmtInUSDT {
			matchPortAmtInUSDT = remainPortAmtInUSDT
		}

		holdPublicToken, lockPRVCollateral, lockTokenColaterals = calHoldPubTokenAmountAndLockCollaterals(
			portingAmount, portAmtInUSDT, matchPortAmtInUSDT, convertRateTool, custodianState)
		actualHoldPubToken += holdPublicToken

		matchCus := statedb.MatchingPortingCustodianDetail{
			IncAddress:             custodianState.GetIncognitoAddress(),
			RemoteAddress:          custodianState.GetRemoteAddresses()[portalTokenID],
			Amount:                 holdPublicToken,
			LockedAmountCollateral: lockPRVCollateral,
			LockedTokenCollaterals: lockTokenColaterals,
		}
		matchCustodians = append(matchCustodians, &matchCus)

		remainPortAmtInUSDT -= matchPortAmtInUSDT
		if remainPortAmtInUSDT == 0 {
			if actualHoldPubToken < portingAmount {
				matchCustodians[0].Amount = matchCustodians[0].Amount + portingAmount - actualHoldPubToken
			}
			break
		}
	}
	if remainPortAmtInUSDT > 0 {
		return nil, errors.New("pickUpCustodianForPorting: Not enough custodians for matching to porting request")
	}

	return matchCustodians, nil
}

func UpdateCustodianStateAfterMatchingPortingRequest(
	currentPortalState *CurrentPortalState,
	matchCus *statedb.MatchingPortingCustodianDetail,
	portalTokenID string) error {
	custodianKey := statedb.GenerateCustodianStateObjectKey(matchCus.IncAddress).String()
	custodian, ok := currentPortalState.CustodianPoolState[custodianKey]
	if !ok {
		return errors.New("Custodian not found")
	}
	// lock PRV collateral
	if matchCus.LockedAmountCollateral > 0 {
		freeCollateral := custodian.GetFreeCollateral() - matchCus.LockedAmountCollateral

		lockPRVCollateral := custodian.GetLockedAmountCollateral()
		if lockPRVCollateral == nil {
			lockPRVCollateral = make(map[string]uint64)
		}
		lockPRVCollateral[portalTokenID] += matchCus.LockedAmountCollateral

		custodian.SetFreeCollateral(freeCollateral)
		custodian.SetLockedAmountCollateral(lockPRVCollateral)
	}

	// lock token collaterals
	if len(matchCus.LockedTokenCollaterals) > 0 {
		freeTokenCollaterals := custodian.GetFreeTokenCollaterals()
		lockTokenCollaterals := custodian.GetLockedTokenCollaterals()
		if lockTokenCollaterals == nil {
			lockTokenCollaterals = map[string]map[string]uint64{}
		}
		if lockTokenCollaterals[portalTokenID] == nil {
			lockTokenCollaterals[portalTokenID] = map[string]uint64{}
		}
		for collateralTokenID, amount := range matchCus.LockedTokenCollaterals {
			freeTokenCollaterals[collateralTokenID] -= amount
			lockTokenCollaterals[portalTokenID][collateralTokenID] += amount
		}

		custodian.SetFreeTokenCollaterals(freeTokenCollaterals)
		custodian.SetLockedTokenCollaterals(lockTokenCollaterals)
	}

	// Note: don't update holding public tokens to avoid this custodian match to redeem request before receiving pubtokens from users
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

func CalMinPortingFee(portingAmountInPToken uint64, portalTokenID string, exchangeRate *statedb.FinalExchangeRatesState, portalParam PortalParams) (uint64, error) {
	exchangeTool := NewPortalExchangeRateTool(exchangeRate, portalParam.SupportedCollateralTokens)
	portingAmountInPRV, err := exchangeTool.Convert(portalTokenID, common.PRVIDStr, portingAmountInPToken)
	if err != nil {
		Logger.log.Errorf("Error when calculating minimum porting fee %v", err)
		return 0, err
	}

	// can't use big int to calculate porting fee because of common.MinPercentPortingFee < 1
	portingFee := uint64(math.Round(float64(portingAmountInPRV) * portalParam.MinPercentPortingFee / 100))

	return portingFee, nil
}

func CalMinRedeemFee(redeemAmountInPToken uint64, portalTokenID string, exchangeRate *statedb.FinalExchangeRatesState, portalParam PortalParams) (uint64, error) {
	exchangeTool := NewPortalExchangeRateTool(exchangeRate, portalParam.SupportedCollateralTokens)
	redeemAmountInPRV, err := exchangeTool.Convert(portalTokenID, common.PRVIDStr, redeemAmountInPToken)
	if err != nil {
		Logger.log.Errorf("Error when calculating minimum redeem fee %v", err)
		return 0, err
	}

	// can't use big int to calculate porting fee because of common.MinPercentRedeemFee < 1
	redeemFee := uint64(math.Round(float64(redeemAmountInPRV) * portalParam.MinPercentRedeemFee / 100))

	return redeemFee, nil
}

/*
	up 150%
*/
func upPercent(amount uint64, percent uint64) uint64 {
	tmp := new(big.Int).Mul(new(big.Int).SetUint64(amount), new(big.Int).SetUint64(percent))
	result := new(big.Int).Div(tmp, new(big.Int).SetUint64(100)).Uint64()
	return result //return nano pBTC, pBNB
}

func downPercent(amount uint64, percent uint64) uint64 {
	tmp := new(big.Int).Mul(new(big.Int).SetUint64(amount), new(big.Int).SetUint64(100))
	result := new(big.Int).Div(tmp, new(big.Int).SetUint64(percent)).Uint64()
	return result
}

func calTotalLiquidationByExchangeRates(RedeemAmount uint64, liquidateExchangeRates statedb.LiquidationPoolDetail) (uint64, error) {
	//todo: need review divide operator
	// prv  ------   total token
	// ?		     amount token

	if liquidateExchangeRates.PubTokenAmount <= 0 {
		return 0, errors.New("Can not divide 0")
	}

	tmp := new(big.Int).Mul(
		new(big.Int).SetUint64(liquidateExchangeRates.CollateralAmount),
		new(big.Int).SetUint64(RedeemAmount),
	)
	totalPrv := new(big.Int).Div(
		tmp,
		new(big.Int).SetUint64(liquidateExchangeRates.PubTokenAmount),
	)
	return totalPrv.Uint64(), nil
}

//check value is tp120 or tp130
func checkTPRatio(tpValue uint64, portalParams PortalParams) (bool, bool) {
	if tpValue > portalParams.TP120 && tpValue <= portalParams.TP130 {
		return false, true
	}

	if tpValue <= portalParams.TP120 {
		return true, true
	}

	//not found
	return false, false
}

func CalAmountNeededDepositLiquidate(currentPortalState *CurrentPortalState, custodian *statedb.CustodianState, exchangeRates *statedb.FinalExchangeRatesState, pTokenId string, portalParams PortalParams) (uint64, error) {
	if custodian.GetHoldingPublicTokens() == nil {
		return 0, nil
	}
	totalHoldingPublicToken := GetTotalHoldPubTokenAmount(currentPortalState, custodian, pTokenId)
	totalPToken := upPercent(totalHoldingPublicToken, portalParams.MinPercentLockedCollateral)
	convertExchangeRatesObj := NewConvertExchangeRatesObject(exchangeRates)
	totalPRV, err := convertExchangeRatesObj.ExchangePToken2PRVByTokenId(pTokenId, totalPToken)
	if err != nil {
		return 0, err
	}

	// convert free collaterals of custodians to usdt to compare and sort descending
	convertRateTool := NewPortalExchangeRateTool(currentPortalState.FinalExchangeRatesState, portalParams.SupportedCollateralTokens)
	// TODO: update this function from prv to usdt
	totalLockedAmountInWaitingPorting, err := GetTotalLockedCollateralAmountInWaitingPortings(currentPortalState, custodian, pTokenId, convertRateTool)
	if err != nil {
		return 0, errors.New("[CalAmountNeededDepositLiquidate] got error while get total token valued in USDT locked ")
	}

	lockedAmountMap := custodian.GetLockedAmountCollateral()
	if lockedAmountMap == nil {
		return 0, errors.New("Locked amount is nil")
	}
	lockedAmount := lockedAmountMap[pTokenId] - totalLockedAmountInWaitingPorting
	if lockedAmount >= totalPRV {
		return 0, nil
	}

	totalAmountNeeded := totalPRV - lockedAmount

	return totalAmountNeeded, nil
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
		if sortedCustodians[i].Value.GetHoldingPublicTokens()[tokenID] < sortedCustodians[j].Value.GetHoldingPublicTokens()[tokenID] {
			return true
		} else if (sortedCustodians[i].Value.GetHoldingPublicTokens()[tokenID] == sortedCustodians[j].Value.GetHoldingPublicTokens()[tokenID]) &&
			(sortedCustodians[i].Value.GetIncognitoAddress() < sortedCustodians[j].Value.GetIncognitoAddress()) {
			return true
		}
		return false
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
func updateCustodianStateAfterReqUnlockCollateral(custodianState *statedb.CustodianState, unlockedAmount uint64, tokenID string, portalParams PortalParams, portalState *CurrentPortalState) error {
	lockedTokenAmounts := custodianState.GetLockedTokenCollaterals()
	lockedPrvAmount := custodianState.GetLockedAmountCollateral()
	if lockedTokenAmounts == nil && lockedPrvAmount == nil {
		return errors.New("[portal-updateCustodianStateAfterReqUnlockCollateral] Locked amount is nil")
	}

	// convert free collaterals of custodians to usdt to compare and sort descending
	convertRateTool := NewPortalExchangeRateTool(portalState.FinalExchangeRatesState, portalParams.SupportedCollateralTokens)
	var amountToUnlockTemp uint64
	if lockedPrvAmount != nil {
		tokenAmtInUSD, err := convertRateTool.ConvertToUSDT(common.PRVIDStr, lockedPrvAmount[tokenID])
		if err != nil {
			return errors.New("[portal-updateCustodianStateAfterReqUnlockCollateral] Can not convert prv to usd")
		}
		if unlockedAmount > tokenAmtInUSD {
			if lockedTokenAmounts == nil {
				return errors.New("[portal-updateCustodianStateAfterReqUnlockCollateral] can not unlock nil tokens")
			}
			unlockedAmount -= tokenAmtInUSD
			amountToUnlockTemp = tokenAmtInUSD
		} else {
			amountToUnlockTemp = unlockedAmount
			unlockedAmount = 0
		}
		prvCollateralAmountToUpdate, err := convertRateTool.ConvertFromUSDT(common.PRVIDStr, amountToUnlockTemp)
		if err != nil || prvCollateralAmountToUpdate > lockedPrvAmount[tokenID] {
			return errors.New("[portal-updateCustodianStateAfterReqUnlockCollateral] Can not convert usd to collateral prv")
		}

		// update collateral prv token
		lockedPrvAmount[tokenID] -= prvCollateralAmountToUpdate
		custodianState.SetLockedAmountCollateral(lockedPrvAmount)
		// update free prv token
		custodianState.SetFreeCollateral(custodianState.GetFreeCollateral() + prvCollateralAmountToUpdate)
	}

	amountToUnlockTemp = uint64(0)
	freeTokenCollaterals := custodianState.GetFreeTokenCollaterals()

	if unlockedAmount > 0 {
		// lock other token collaterals
		sortedTokenIDs := []string{}
		for tokenCollateralID := range lockedTokenAmounts[tokenID] {
			sortedTokenIDs = append(sortedTokenIDs, tokenCollateralID)
		}
		sort.Strings(sortedTokenIDs)

		for _, tokenCollateralID := range sortedTokenIDs {
			tokenValueLocked, err := convertRateTool.ConvertToUSDT(tokenCollateralID, lockedTokenAmounts[tokenID][tokenCollateralID])
			if err != nil {
				Logger.log.Errorf("[portal-updateCustodianStateAfterReqUnlockCollateral] got error %v", err.Error())
				return errors.New("[portal-updateCustodianStateAfterReqUnlockCollateral] got error while get convert from collateral to USDT ")
			}
			if unlockedAmount > tokenValueLocked {
				amountToUnlockTemp = tokenValueLocked
				unlockedAmount -= tokenValueLocked
			} else {
				amountToUnlockTemp = unlockedAmount
				unlockedAmount = 0
			}
			tokenCollateralAmountToUpdate, err := convertRateTool.ConvertFromUSDT(tokenCollateralID, amountToUnlockTemp)
			if err != nil || tokenCollateralAmountToUpdate > lockedTokenAmounts[tokenID][tokenCollateralID] {
				return errors.New("[portal-updateCustodianStateAfterReqUnlockCollateral] Can not convert usd to collateral token")
			}
			lockedTokenAmounts[tokenID][tokenCollateralID] -= tokenCollateralAmountToUpdate
			freeTokenCollaterals[tokenCollateralID] += tokenCollateralAmountToUpdate

			if unlockedAmount == 0 {
				break
			}
		}

		if unlockedAmount > 0 {
			return errors.New("[portal-updateCustodianStateAfterReqUnlockCollateral] not enough collateral tokens to unlock for custodian")
		}
		custodianState.SetLockedTokenCollaterals(lockedTokenAmounts)
		custodianState.SetFreeTokenCollaterals(freeTokenCollaterals)
	}
	return nil
}

// CalUnlockCollateralAmount returns unlock collateral amount by percentage of redeem amount in usd
func CalUnlockCollateralAmount(
	portalState *CurrentPortalState,
	custodianStateKey string,
	redeemAmount uint64,
	tokenID string,
	portalParams PortalParams) (uint64, error) {
	custodianState := portalState.CustodianPoolState[custodianStateKey]
	if custodianState == nil {
		Logger.log.Errorf("[test][CalUnlockCollateralAmount] Custodian not found %v\n", custodianStateKey)
		return 0, fmt.Errorf("Custodian not found %v\n", custodianStateKey)
	}

	totalHoldingPubToken := GetTotalHoldPubTokenAmount(portalState, custodianState, tokenID)

	// convert free collaterals of custodians to usdt to compare and sort descending
	convertRateTool := NewPortalExchangeRateTool(portalState.FinalExchangeRatesState, portalParams.SupportedCollateralTokens)

	totalLockedAmountInWaitingPortings, err := GetTotalLockedCollateralAmountInWaitingPortings(portalState, custodianState, tokenID, convertRateTool)
	if err != nil {
		Logger.log.Errorf("[GetTotalLockedCollateralAmountInWaitingPortings] got error %v", err.Error())
		return 0, errors.New("[CalUnlockCollateralAmount] got error while get total token valued in USDT locked ")
	}

	lockedAmountCollateral := uint64(0)
	listLockedTokens := custodianState.GetLockedTokenCollaterals()[tokenID]
	listLockedTokens[common.PRVIDStr] = custodianState.GetLockedAmountCollateral()[tokenID]
	for tokenID, token := range listLockedTokens {
		tokenValueLocked, err := convertRateTool.ConvertToUSDT(tokenID, token)
		if err != nil {
			Logger.log.Errorf("[GetTotalLockedCollateralAmountInWaitingPortings] got error %v", err.Error())
			return 0, errors.New("[CalUnlockCollateralAmount] got error while get convert from collateral to USDT ")
		}
		lockedAmountCollateral += tokenValueLocked
	}

	if lockedAmountCollateral < totalLockedAmountInWaitingPortings {
		Logger.log.Errorf("custodianState.GetLockedAmountCollateral()[tokenID] %v\n", custodianState.GetLockedAmountCollateral()[tokenID])
		Logger.log.Errorf("totalLockedAmountInWaitingPortings %v\n", totalLockedAmountInWaitingPortings)
		return 0, errors.New("[CalUnlockCollateralAmount] Lock amount is invalid")
	}

	if totalHoldingPubToken == 0 {
		Logger.log.Errorf("[CalUnlockCollateralAmount] Total holding public token amount of custodianAddr %v is zero", custodianState.GetIncognitoAddress())
		return 0, errors.New("[CalUnlockCollateralAmount] Total holding public token amount is zero")
	}

	tmp := new(big.Int).Mul(
		new(big.Int).SetUint64(redeemAmount),
		new(big.Int).SetUint64(lockedAmountCollateral-totalLockedAmountInWaitingPortings))
	unlockAmount := new(big.Int).Div(tmp, new(big.Int).SetUint64(totalHoldingPubToken)).Uint64()
	if unlockAmount <= 0 || unlockAmount > (lockedAmountCollateral-totalLockedAmountInWaitingPortings) {
		Logger.log.Errorf("[CalUnlockCollateralAmount] Can not calculate unlock amount for custodian %v\n", unlockAmount)
		return 0, errors.New("[CalUnlockCollateralAmount] Can not calculate unlock amount for custodian")
	}
	return unlockAmount, nil
}

func CalUnlockCollateralAmountAfterLiquidation(
	portalState *CurrentPortalState,
	liquidatedCustodianStateKey string,
	amountPubToken uint64,
	tokenID string,
	exchangeRate *statedb.FinalExchangeRatesState,
	portalParams PortalParams) (uint64, uint64, error) {
	// TODO: update this function from prv amount to usdt
	totalUnlockCollateralAmount, err := CalUnlockCollateralAmount(portalState, liquidatedCustodianStateKey, amountPubToken, tokenID, portalParams)
	if err != nil {
		Logger.log.Errorf("CalUnlockCollateralAmountAfterLiquidation error : %v\n", err)
		return 0, 0, err
	}
	convertExchangeRatesObj := NewConvertExchangeRatesObject(exchangeRate)
	tmp := new(big.Int).Mul(new(big.Int).SetUint64(amountPubToken), new(big.Int).SetUint64(uint64(portalParams.MaxPercentLiquidatedCollateralAmount)))
	liquidatedAmountInPToken := new(big.Int).Div(tmp, new(big.Int).SetUint64(100)).Uint64()
	liquidatedAmountInPRV, err := convertExchangeRatesObj.ExchangePToken2PRVByTokenId(tokenID, liquidatedAmountInPToken)
	if err != nil {
		Logger.log.Errorf("CalUnlockCollateralAmountAfterLiquidation error converting rate : %v\n", err)
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
	lockedAmountTmp := custodianState.GetLockedAmountCollateral()
	if lockedAmountTmp[tokenID] < liquidatedAmount+remainUnlockAmountForCustodian {
		Logger.log.Errorf("[updateCustodianStateAfterLiquidateCustodian] locked amount less than total unlock amount")
		return errors.New("[updateCustodianStateAfterLiquidateCustodian] locked amount less than total unlock amount")
	}

	custodianState.SetTotalCollateral(custodianState.GetTotalCollateral() - liquidatedAmount)

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

func removeCustodianFromMatchingRedeemCustodians(
	matchingCustodians []*statedb.MatchingRedeemCustodianDetail,
	custodianIncAddr string) ([]*statedb.MatchingRedeemCustodianDetail, error) {
	matchingCustodiansRes := make([]*statedb.MatchingRedeemCustodianDetail, len(matchingCustodians))
	copy(matchingCustodiansRes, matchingCustodians)

	for i, cus := range matchingCustodiansRes {
		if cus.GetIncognitoAddress() == custodianIncAddr {
			matchingCustodiansRes = append(matchingCustodiansRes[:i], matchingCustodiansRes[i+1:]...)
			return matchingCustodiansRes, nil
		}
	}
	return matchingCustodiansRes, errors.New("Custodian not found in matching redeem custodians")
}

func deleteWaitingRedeemRequest(state *CurrentPortalState, waitingRedeemRequestKey string) {
	delete(state.WaitingRedeemRequests, waitingRedeemRequestKey)
}

func deleteMatchedRedeemRequest(state *CurrentPortalState, matchedRedeemRequestKey string) {
	delete(state.MatchedRedeemRequests, matchedRedeemRequestKey)
}

func deleteWaitingPortingRequest(state *CurrentPortalState, waitingPortingRequestKey string) {
	delete(state.WaitingPortingRequests, waitingPortingRequestKey)
}

// todo: replace it by PortalExchangeRateTool
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
	liquidateExchangeRatesKey := statedb.GeneratePortalLiquidationPoolObjectKey()
	liquidateExchangeRates, ok := currentPortalState.LiquidationPool[liquidateExchangeRatesKey.String()]
	if !ok {
		item := make(map[string]statedb.LiquidationPoolDetail)

		for ptoken, liquidateTopPercentileExchangeRatesDetail := range tpRatios {
			item[ptoken] = statedb.LiquidationPoolDetail{
				CollateralAmount: liquidateTopPercentileExchangeRatesDetail.HoldAmountFreeCollateral,
				PubTokenAmount:   liquidateTopPercentileExchangeRatesDetail.HoldAmountPubToken,
			}
		}
		currentPortalState.LiquidationPool[liquidateExchangeRatesKey.String()] = statedb.NewLiquidationPoolWithValue(item)
	} else {
		for ptoken, liquidateTopPercentileExchangeRatesDetail := range tpRatios {
			if _, ok := liquidateExchangeRates.Rates()[ptoken]; !ok {
				liquidateExchangeRates.Rates()[ptoken] = statedb.LiquidationPoolDetail{
					CollateralAmount: liquidateTopPercentileExchangeRatesDetail.HoldAmountFreeCollateral,
					PubTokenAmount:   liquidateTopPercentileExchangeRatesDetail.HoldAmountPubToken,
				}
			} else {
				liquidateExchangeRates.Rates()[ptoken] = statedb.LiquidationPoolDetail{
					CollateralAmount: liquidateExchangeRates.Rates()[ptoken].CollateralAmount + liquidateTopPercentileExchangeRatesDetail.HoldAmountFreeCollateral,
					PubTokenAmount:   liquidateExchangeRates.Rates()[ptoken].PubTokenAmount + liquidateTopPercentileExchangeRatesDetail.HoldAmountPubToken,
				}
			}
		}

		currentPortalState.LiquidationPool[liquidateExchangeRatesKey.String()] = liquidateExchangeRates
	}
	//end
}

func getTotalLockedCollateralInEpoch(featureStateDB *statedb.StateDB) (uint64, error) {
	currentPortalState, err := InitCurrentPortalStateFromDB(featureStateDB)
	if err != nil {
		return 0, nil
	}

	return currentPortalState.LockedCollateralForRewards.GetTotalLockedCollateralForRewards(), nil
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
	finalExchange *statedb.FinalExchangeRatesState,
	portalParams PortalParams) (map[string]metadata.LiquidateTopPercentileExchangeRatesDetail, error) {
	result := make(map[string]metadata.LiquidateTopPercentileExchangeRatesDetail)
	convertExchangeRatesObj := NewConvertExchangeRatesObject(finalExchange)

	lockedAmount := make(map[string]uint64)
	for tokenID, amount := range custodianState.GetLockedAmountCollateral() {
		lockedAmount[tokenID] = amount
	}

	holdingPubToken := make(map[string]uint64)
	for tokenID := range custodianState.GetHoldingPublicTokens() {
		holdingPubToken[tokenID] = GetTotalHoldPubTokenAmount(portalState, custodianState, tokenID)
	}

	for _, waitingPortingReq := range portalState.WaitingPortingRequests {
		for _, matchingCus := range waitingPortingReq.Custodians() {
			if matchingCus.IncAddress == custodianState.GetIncognitoAddress() {
				lockedAmount[waitingPortingReq.TokenID()] -= matchingCus.LockedAmountCollateral
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

		if tp20, ok := checkTPRatio(percent, portalParams); ok {
			if tp20 {
				result[tokenID] = metadata.LiquidateTopPercentileExchangeRatesDetail{
					TPKey:                    int(portalParams.TP120),
					TPValue:                  percent,
					HoldAmountFreeCollateral: lockedAmount[tokenID],
					HoldAmountPubToken:       holdingPubToken[tokenID],
				}
			} else {
				result[tokenID] = metadata.LiquidateTopPercentileExchangeRatesDetail{
					TPKey:                    int(portalParams.TP130),
					TPValue:                  percent,
					HoldAmountFreeCollateral: 0,
					HoldAmountPubToken:       0,
				}
			}
		}
	}

	return result, nil
}

func UpdateCustodianStateAfterRejectRedeemRequestByLiquidation(portalState *CurrentPortalState, rejectedRedeemReq *statedb.RedeemRequest, beaconHeight uint64) error {
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

// MatchCustodianToWaitingRedeemReq returns amount matching of custodian in redeem request if valid
func MatchCustodianToWaitingRedeemReq(
	custodianAddr string,
	redeemID string,
	portalState *CurrentPortalState) (uint64, bool, error) {
	// check redeemID is in waiting redeem requests or not
	keyWaitingRedeemRequest := statedb.GenerateWaitingRedeemRequestObjectKey(redeemID).String()
	waitingRedeemRequest := portalState.WaitingRedeemRequests[keyWaitingRedeemRequest]
	if waitingRedeemRequest == nil {
		return 0, false, fmt.Errorf("RedeemID is not existed in waiting matching redeem requests list %v\n", redeemID)
	}

	// check Incognito Address is an custodian or not
	custodianStateKey := statedb.GenerateCustodianStateObjectKey(custodianAddr).String()
	custodianState := portalState.CustodianPoolState[custodianStateKey]
	if custodianState == nil {
		return 0, false, fmt.Errorf("custodianState not found %v\n", custodianAddr)
	}

	// calculate amount need to be matched
	totalMatchedAmount := uint64(0)
	for _, cus := range waitingRedeemRequest.GetCustodians() {
		totalMatchedAmount += cus.GetAmount()
	}
	neededMatchingAmountInPToken := waitingRedeemRequest.GetRedeemAmount() - totalMatchedAmount
	if neededMatchingAmountInPToken <= 0 {
		return 0, false, errors.New("Amount need to be matched is not greater than zero")
	}

	holdPubTokenMap := custodianState.GetHoldingPublicTokens()
	if holdPubTokenMap == nil || len(holdPubTokenMap) == 0 {
		return 0, false, errors.New("Holding public token amount of custodian is not valid")
	}
	holdPubTokenAmount := holdPubTokenMap[waitingRedeemRequest.GetTokenID()]
	if holdPubTokenAmount == 0 {
		return 0, false, errors.New("Holding public token amount of custodian is not available")
	}

	if holdPubTokenAmount >= neededMatchingAmountInPToken {
		return neededMatchingAmountInPToken, true, nil
	} else {
		return holdPubTokenAmount, false, nil
	}
}

func UpdatePortalStateAfterCustodianReqMatchingRedeem(
	custodianAddr string,
	redeemID string,
	matchedAmount uint64,
	isEnoughCustodians bool,
	portalState *CurrentPortalState) (*statedb.RedeemRequest, error) {
	// check redeemID is in waiting redeem requests or not
	keyWaitingRedeemRequest := statedb.GenerateWaitingRedeemRequestObjectKey(redeemID).String()
	waitingRedeemRequest := portalState.WaitingRedeemRequests[keyWaitingRedeemRequest]
	if waitingRedeemRequest == nil {
		return nil, fmt.Errorf("RedeemID is not existed in waiting matching redeem requests list %v\n", redeemID)
	}

	// update custodian state
	custodianStateKey := statedb.GenerateCustodianStateObjectKey(custodianAddr).String()
	custodianState := portalState.CustodianPoolState[custodianStateKey]
	err := UpdateCustodianStateAfterMatchingRedeemReq(custodianState, matchedAmount, waitingRedeemRequest.GetTokenID())
	if err != nil {
		return nil, fmt.Errorf("Error wahne updating custodian state %v\n", err)
	}

	// update matching custodians in waiting redeem request
	matchingCus := waitingRedeemRequest.GetCustodians()
	if matchingCus == nil {
		matchingCus = make([]*statedb.MatchingRedeemCustodianDetail, 0)
	}
	matchingCus = append(matchingCus,
		statedb.NewMatchingRedeemCustodianDetailWithValue(custodianAddr, custodianState.GetRemoteAddresses()[waitingRedeemRequest.GetTokenID()], matchedAmount))
	waitingRedeemRequest.SetCustodians(matchingCus)

	if isEnoughCustodians {
		deleteWaitingRedeemRequest(portalState, keyWaitingRedeemRequest)
		keyMatchedRedeemRequest := statedb.GenerateMatchedRedeemRequestObjectKey(redeemID).String()
		portalState.MatchedRedeemRequests[keyMatchedRedeemRequest] = waitingRedeemRequest
	}

	return waitingRedeemRequest, nil
}

func UpdateCustodianStateAfterMatchingRedeemReq(custodianState *statedb.CustodianState, matchingAmount uint64, tokenID string) error {
	// check Incognito Address is an custodian or not
	if custodianState == nil {
		return fmt.Errorf("custodianState not found %v\n", custodianState)
	}

	// update custodian state
	holdingPubTokenTmp := custodianState.GetHoldingPublicTokens()
	if holdingPubTokenTmp == nil {
		return errors.New("Holding public token of custodian is null")
	}
	if holdingPubTokenTmp[tokenID] < matchingAmount {
		return fmt.Errorf("Holding public token %v is less than matching amount %v : ", holdingPubTokenTmp[tokenID], matchingAmount)
	}
	holdingPubTokenTmp[tokenID] -= matchingAmount
	custodianState.SetHoldingPublicTokens(holdingPubTokenTmp)

	return nil
}

func UpdatePortalStateAfterPickMoreCustodiansForWaitingRedeemReq(
	moreCustodians []*statedb.MatchingRedeemCustodianDetail,
	waitingRedeem *statedb.RedeemRequest,
	portalState *CurrentPortalState) (*statedb.RedeemRequest, error) {
	// update custodian state
	for _, cus := range moreCustodians {
		custodianStateKey := statedb.GenerateCustodianStateObjectKey(cus.GetIncognitoAddress()).String()
		err := UpdateCustodianStateAfterMatchingRedeemReq(portalState.CustodianPoolState[custodianStateKey], cus.GetAmount(), waitingRedeem.GetTokenID())
		if err != nil {
			Logger.log.Errorf("Error when update custodian state for timeout redeem request %v\n", err)
			return nil, err
		}
	}

	// move waiting redeem request from waiting list to matched list
	waitingRedeemKey := statedb.GenerateWaitingRedeemRequestObjectKey(waitingRedeem.GetUniqueRedeemID()).String()
	deleteWaitingRedeemRequest(portalState, waitingRedeemKey)

	matchedCustodians := waitingRedeem.GetCustodians()
	if matchedCustodians == nil {
		matchedCustodians = make([]*statedb.MatchingRedeemCustodianDetail, 0)
	}
	matchedCustodians = append(matchedCustodians, moreCustodians...)
	waitingRedeem.SetCustodians(matchedCustodians)

	matchedRedeemKey := statedb.GenerateMatchedRedeemRequestObjectKey(waitingRedeem.GetUniqueRedeemID()).String()
	portalState.MatchedRedeemRequests[matchedRedeemKey] = waitingRedeem

	return waitingRedeem, nil
}

// convert all tokens to usd
func GetTotalLockedCollateralAmountInWaitingPortings(portalState *CurrentPortalState, custodianState *statedb.CustodianState, tokenID string, convertRateTool *PortalExchangeRateTool) (uint64, error) {
	totalLockedAmountInWaitingPortings := uint64(0)
	for _, waitingPortingReq := range portalState.WaitingPortingRequests {
		if waitingPortingReq.TokenID() != tokenID {
			continue
		}
		for _, cus := range waitingPortingReq.Custodians() {
			if cus.IncAddress == custodianState.GetIncognitoAddress() {
				tokenList := cus.LockedTokenCollaterals
				tokenList[common.PRVIDStr] = cus.LockedAmountCollateral
				for tokenId, tokenValue := range tokenList {
					tokenValueConverted, err := convertRateTool.ConvertToUSDT(tokenId, tokenValue)
					if err != nil {
						return 0, err
					}
					totalLockedAmountInWaitingPortings += tokenValueConverted
				}
				break
			}
		}
	}

	return totalLockedAmountInWaitingPortings, nil
}

func GetTotalHoldPubTokenAmount(portalState *CurrentPortalState, custodianState *statedb.CustodianState, tokenID string) uint64 {
	holdPubToken := custodianState.GetHoldingPublicTokens()
	totalHoldingPubTokenAmount := uint64(0)
	if holdPubToken != nil {
		totalHoldingPubTokenAmount += holdPubToken[tokenID]
	}

	for _, waitingRedeemReq := range portalState.WaitingRedeemRequests {
		if waitingRedeemReq.GetTokenID() != tokenID {
			continue
		}

		for _, cus := range waitingRedeemReq.GetCustodians() {
			if cus.GetIncognitoAddress() != custodianState.GetIncognitoAddress() {
				continue
			}
			totalHoldingPubTokenAmount += cus.GetAmount()
			break
		}
	}

	for _, matchedRedeemReq := range portalState.MatchedRedeemRequests {
		if matchedRedeemReq.GetTokenID() != tokenID {
			continue
		}

		for _, cus := range matchedRedeemReq.GetCustodians() {
			if cus.GetIncognitoAddress() != custodianState.GetIncognitoAddress() {
				continue
			}
			totalHoldingPubTokenAmount += cus.GetAmount()
			break
		}
	}

	return totalHoldingPubTokenAmount
}

func GetTotalHoldPubTokenAmountExcludeMatchedRedeemReqs(portalState *CurrentPortalState, custodianState *statedb.CustodianState, tokenID string) uint64 {
	//todo: check
	holdPubToken := custodianState.GetHoldingPublicTokens()
	totalHoldingPubTokenAmount := uint64(0)
	if holdPubToken != nil {
		totalHoldingPubTokenAmount += holdPubToken[tokenID]
	}

	for _, waitingRedeemReq := range portalState.WaitingRedeemRequests {
		if waitingRedeemReq.GetTokenID() != tokenID {
			continue
		}

		for _, cus := range waitingRedeemReq.GetCustodians() {
			if cus.GetIncognitoAddress() != custodianState.GetIncognitoAddress() {
				continue
			}
			totalHoldingPubTokenAmount += cus.GetAmount()
			break
		}
	}

	return totalHoldingPubTokenAmount
}

func GetTotalMatchingPubTokenInWaitingPortings(portalState *CurrentPortalState, custodianState *statedb.CustodianState, tokenID string) uint64 {
	totalMatchingPubTokenAmount := uint64(0)

	for _, waitingPortingReq := range portalState.WaitingPortingRequests {
		if waitingPortingReq.TokenID() != tokenID {
			continue
		}

		for _, cus := range waitingPortingReq.Custodians() {
			if cus.IncAddress != custodianState.GetIncognitoAddress() {
				continue
			}
			totalMatchingPubTokenAmount += cus.Amount
			break
		}
	}

	return totalMatchingPubTokenAmount
}

func UpdateLockedCollateralForRewards(currentPortalState *CurrentPortalState, portalParam PortalParams) {
	exchangeTool := NewPortalExchangeRateTool(currentPortalState.FinalExchangeRatesState, portalParam.SupportedCollateralTokens)

	totalLockedCollateralAmount := currentPortalState.LockedCollateralForRewards.GetTotalLockedCollateralForRewards()
	lockedCollateralDetails := currentPortalState.LockedCollateralForRewards.GetLockedCollateralDetail()

	for _, custodianState := range currentPortalState.CustodianPoolState {
		lockedCollaterals := custodianState.GetLockedAmountCollateral()
		if lockedCollaterals == nil || len(lockedCollaterals) == 0 {
			continue
		}

		for tokenID := range lockedCollaterals {
			holdPubTokenAmount := GetTotalHoldPubTokenAmount(currentPortalState, custodianState, tokenID)
			matchingPubTokenAmount := GetTotalMatchingPubTokenInWaitingPortings(currentPortalState, custodianState, tokenID)
			totalPubToken := holdPubTokenAmount + matchingPubTokenAmount
			pubTokenAmountInPRV, err := exchangeTool.Convert(tokenID, common.PRVIDStr, totalPubToken)
			if err != nil {
				Logger.log.Errorf("Error when converting public token to prv: %v", err)
			}
			lockedCollateralDetails[custodianState.GetIncognitoAddress()] += pubTokenAmountInPRV
			totalLockedCollateralAmount += pubTokenAmountInPRV
		}
	}

	currentPortalState.LockedCollateralForRewards.SetTotalLockedCollateralForRewards(totalLockedCollateralAmount)
	currentPortalState.LockedCollateralForRewards.SetLockedCollateralDetail(lockedCollateralDetails)
}

func UpdateLockedCollateralForRewardsV3(currentPortalState *CurrentPortalState, portalParam PortalParams) {
	exchangeTool := NewPortalExchangeRateTool(currentPortalState.FinalExchangeRatesState, portalParam.SupportedCollateralTokens)

	totalLockedCollateralAmount := currentPortalState.LockedCollateralForRewards.GetTotalLockedCollateralForRewards()
	lockedCollateralDetails := currentPortalState.LockedCollateralForRewards.GetLockedCollateralDetail()
	portalTokenIDs := common.PortalSupportedIncTokenIDs
	for _, custodianState := range currentPortalState.CustodianPoolState {
		for _, tokenID := range portalTokenIDs {
			holdPubTokenAmount := GetTotalHoldPubTokenAmount(currentPortalState, custodianState, tokenID)
			matchingPubTokenAmount := GetTotalMatchingPubTokenInWaitingPortings(currentPortalState, custodianState, tokenID)
			totalPubToken := holdPubTokenAmount + matchingPubTokenAmount
			if totalPubToken == 0 {
				continue
			}
			pubTokenAmountInUSDT, err := exchangeTool.ConvertToUSDT(tokenID, totalPubToken)
			if err != nil {
				Logger.log.Errorf("Error when converting public token to prv: %v", err)
			}
			lockedCollateralDetails[custodianState.GetIncognitoAddress()] += pubTokenAmountInUSDT
			totalLockedCollateralAmount += pubTokenAmountInUSDT
		}
	}

	currentPortalState.LockedCollateralForRewards.SetTotalLockedCollateralForRewards(totalLockedCollateralAmount)
	currentPortalState.LockedCollateralForRewards.SetLockedCollateralDetail(lockedCollateralDetails)
}

func CalAmountTopUpWaitingPortings(
	portalState *CurrentPortalState,
	custodianState *statedb.CustodianState, portalParam PortalParams) (map[string]uint64, error) {

	result := make(map[string]uint64)
	convertExchangeRatesObj := NewConvertExchangeRatesObject(portalState.FinalExchangeRatesState)

	for _, waitingPorting := range portalState.WaitingPortingRequests {
		for _, cus := range waitingPorting.Custodians() {
			if cus.IncAddress != custodianState.GetIncognitoAddress() {
				continue
			}

			minCollateralAmount, err := convertExchangeRatesObj.ExchangePToken2PRVByTokenId(
				waitingPorting.TokenID(),
				upPercent(cus.Amount, portalParam.MinPercentLockedCollateral))
			if err != nil {
				Logger.log.Errorf("[calAmountTopUpWaitingPortings] Error when converting ptoken to PRV %v", err)
				return result, err
			}

			if minCollateralAmount <= cus.LockedAmountCollateral {
				break
			}

			result[waitingPorting.UniquePortingID()] = minCollateralAmount - cus.LockedAmountCollateral
		}
	}

	return result, nil
}

func cloneRedeemRequests(redeemReqs map[string]*statedb.RedeemRequest) map[string]*statedb.RedeemRequest {
	newReqs := make(map[string]*statedb.RedeemRequest, len(redeemReqs))
	for key, req := range redeemReqs {
		newReqs[key] = statedb.NewRedeemRequestWithValue(
			req.GetUniqueRedeemID(),
			req.GetTokenID(),
			req.GetRedeemerAddress(),
			req.GetRedeemerRemoteAddress(),
			req.GetRedeemAmount(),
			req.GetCustodians(),
			req.GetRedeemFee(),
			req.GetBeaconHeight(),
			req.GetTxReqID(),
		)
	}
	return newReqs
}

func getNewMatchedRedeemReqIDs(
	existRedeemReqs map[string]*statedb.RedeemRequest,
	newRedeemReqs map[string]*statedb.RedeemRequest) []string {
	newIDs := []string{}

	m := map[string]bool{}
	for _, req := range existRedeemReqs {
		m[req.GetUniqueRedeemID()] = true
	}

	for _, req := range newRedeemReqs {
		newID := req.GetUniqueRedeemID()
		if m[newID] == false {
			newIDs = append(newIDs, newID)
		}
	}

	return newIDs
}

func addCustodianToPool(
	custodianPool map[string]*statedb.CustodianState,
	custodianIncAddr string,
	depositAmount uint64,
	collateralTokenID string,
	remoteAddresses map[string]string,
) *statedb.CustodianState {
	keyCustodianState := statedb.GenerateCustodianStateObjectKey(custodianIncAddr)
	keyCustodianStateStr := keyCustodianState.String()
	existCustodian := custodianPool[keyCustodianStateStr]

	// check collateral token ID
	isPRVCollateral := collateralTokenID == common.PRVIDStr

	// the custodian hasn't deposited before
	newCustodian := statedb.NewCustodianState()
	if existCustodian == nil {
		newCustodian.SetIncognitoAddress(custodianIncAddr)
		newCustodian.SetRemoteAddresses(remoteAddresses)

		if isPRVCollateral {
			newCustodian.SetTotalCollateral(depositAmount)
			newCustodian.SetFreeCollateral(depositAmount)
		} else {
			totalTokenColaterals := map[string]uint64{
				collateralTokenID: depositAmount,
			}
			newCustodian.SetTotalTokenCollaterals(totalTokenColaterals)
			newCustodian.SetFreeTokenCollaterals(totalTokenColaterals)
		}
	} else {
		newCustodian.SetIncognitoAddress(custodianIncAddr)
		newCustodian.SetHoldingPublicTokens(existCustodian.GetHoldingPublicTokens())
		newCustodian.SetLockedAmountCollateral(existCustodian.GetLockedAmountCollateral())
		newCustodian.SetLockedTokenCollaterals(existCustodian.GetLockedTokenCollaterals())
		newCustodian.SetRewardAmount(existCustodian.GetRewardAmount())

		updateRemoteAddresses := existCustodian.GetRemoteAddresses()
		if len(remoteAddresses) > 0 {
			// if total collateral is zero, custodians are able to update remote addresses
			if existCustodian.IsEmptyCollaterals() {
				updateRemoteAddresses = remoteAddresses
			} else {
				sortedTokenIDs := make([]string, 0)
				for tokenID := range remoteAddresses {
					sortedTokenIDs = append(sortedTokenIDs, tokenID)
				}

				for _, tokenID := range sortedTokenIDs {
					if updateRemoteAddresses[tokenID] != "" {
						continue
					}
					updateRemoteAddresses[tokenID] = remoteAddresses[tokenID]
				}
			}
		}
		newCustodian.SetRemoteAddresses(updateRemoteAddresses)

		if isPRVCollateral {
			newCustodian.SetTotalCollateral(existCustodian.GetTotalCollateral() + depositAmount)
			newCustodian.SetFreeCollateral(existCustodian.GetFreeCollateral() + depositAmount)
			newCustodian.SetTotalTokenCollaterals(existCustodian.GetTotalTokenCollaterals())
			newCustodian.SetFreeTokenCollaterals(existCustodian.GetFreeTokenCollaterals())
		} else {
			newCustodian.SetTotalCollateral(existCustodian.GetTotalCollateral())
			newCustodian.SetFreeCollateral(existCustodian.GetFreeCollateral())

			tmpTotalTokenCollaterals := existCustodian.GetTotalTokenCollaterals()
			tmpTotalTokenCollaterals[collateralTokenID] += depositAmount
			tmpFreeTokenCollaterals := existCustodian.GetFreeTokenCollaterals()
			tmpFreeTokenCollaterals[collateralTokenID] += depositAmount
			newCustodian.SetTotalTokenCollaterals(tmpTotalTokenCollaterals)
			newCustodian.SetFreeTokenCollaterals(existCustodian.GetFreeTokenCollaterals())
		}
	}

	return newCustodian
}

func UpdateCustodianStateAfterWithdrawCollateral(custodian *statedb.CustodianState, collateralTokenID string, amount uint64) *statedb.CustodianState {
	if collateralTokenID == common.PRVIDStr {
		custodian.SetTotalCollateral(custodian.GetTotalCollateral() - amount)
		custodian.SetFreeCollateral(custodian.GetFreeCollateral() - amount)
	} else {
		freeTokenCollaterals := custodian.GetFreeTokenCollaterals()
		freeTokenCollaterals[collateralTokenID] -= amount
		totalTokenCollaterals := custodian.GetTotalTokenCollaterals()
		totalTokenCollaterals[collateralTokenID] -= amount

		custodian.SetTotalTokenCollaterals(totalTokenCollaterals)
		custodian.SetFreeTokenCollaterals(freeTokenCollaterals)
	}

	return custodian
}
