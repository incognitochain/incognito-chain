package blockchain

import (
	"encoding/base64"
	"encoding/json"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"math/big"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/metadata"
	"strconv"
)

func buildReqUnlockOverRateCollateralsInst(
	custodianAddresStr string,
	tokenID string,
	unlockedAmount uint64,
	metaType int,
	shardID byte,
	status string,
) []string {
	unlockOverRateCollateralsContent := metadata.PortalUnlockOverRateCollateralsContent{
		CustodianAddressStr: custodianAddresStr,
		TokenID:             tokenID,
		UnlockedAmount:      unlockedAmount,
	}
	unlockOverRateCollateralsContentBytes, _ := json.Marshal(unlockOverRateCollateralsContent)
	return []string{
		strconv.Itoa(metaType),
		strconv.Itoa(int(shardID)),
		status,
		string(unlockOverRateCollateralsContentBytes),
	}
}

type portalCusUnlockOverRateCollateralsProcessor struct {
	*portalInstProcessor
}

func (p *portalCusUnlockOverRateCollateralsProcessor) getActions() map[byte][][]string {
	return p.actions
}

func (p *portalCusUnlockOverRateCollateralsProcessor) putAction(action []string, shardID byte) {
	_, found := p.actions[shardID]
	if !found {
		p.actions[shardID] = [][]string{action}
	} else {
		p.actions[shardID] = append(p.actions[shardID], action)
	}
}

func (p *portalCusUnlockOverRateCollateralsProcessor) prepareDataBeforeProcessing(stateDB *statedb.StateDB, contentStr string) (map[string]interface{}, error) {
	return nil, nil
}

func (p *portalCusUnlockOverRateCollateralsProcessor) buildNewInsts(
	bc *BlockChain,
	contentStr string,
	shardID byte,
	currentPortalState *CurrentPortalState,
	beaconHeight uint64,
	portalParams PortalParams,
	optionalData map[string]interface{},
) ([][]string, error) {
	actionContentBytes, err := base64.StdEncoding.DecodeString(contentStr)
	if err != nil {
		Logger.log.Errorf("ERROR: an error occurred while decoding content string of portal exchange rates action: %+v", err)
		return [][]string{}, nil
	}

	var actionData metadata.PortalUnlockOverRateCollateralsAction
	err = json.Unmarshal(actionContentBytes, &actionData)
	if err != nil {
		Logger.log.Errorf("ERROR: an error occurred while unmarshal portal exchange rates action: %+v", err)
		return [][]string{}, nil
	}

	metaType := actionData.Meta.Type

	rejectInst := buildReqUnlockOverRateCollateralsInst(
		actionData.Meta.CustodianAddressStr,
		actionData.Meta.TokenID,
		uint64(0),
		metaType,
		shardID,
		common.PortalCusUnlockOverRateCollateralsRejectedChainStatus,
	)
	//check key from db
	if currentPortalState.UnlockOverRateCollaterals != nil {
		_, ok := currentPortalState.UnlockOverRateCollaterals[actionData.TxReqID.String()]
		if ok {
			Logger.log.Errorf("ERROR: unlock over rate collaterals rates key is duplicated")
			return [][]string{rejectInst}, nil
		}
	}
	exchangeTool := NewPortalExchangeRateTool(currentPortalState.FinalExchangeRatesState, portalParams.SupportedCollateralTokens)
	custodianStateKey := statedb.GenerateCustodianStateObjectKey(actionData.Meta.CustodianAddressStr).String()
	custodianState := currentPortalState.CustodianPoolState[custodianStateKey]
	tokenAmountListInWaitingPoring := GetTotalLockedCollateralAmountInWaitingPortingsV3(currentPortalState, custodianState, actionData.Meta.TokenID)
	lockedCollaters := custodianState.GetLockedTokenCollaterals()[actionData.Meta.TokenID]
	lockedCollaters[common.PRVIDStr] = custodianState.GetLockedAmountCollateral()[actionData.Meta.TokenID]
	lockedCollatersExceptPorting := make(map[string]uint64, 0)
	totalAmountInUSD := uint64(0)
	for collateralID, tokenValue := range lockedCollaters {
		if tokenValue < tokenAmountListInWaitingPoring[collateralID] {
			Logger.log.Errorf("ERROR: total %v locked less than amount lock in porting", collateralID)
			return [][]string{rejectInst}, nil
		}
		lockedCollatersExceptPorting[collateralID] = lockedCollaters[collateralID] - tokenAmountListInWaitingPoring[collateralID]
		// convert to usd
		pubTokenAmountInUSDT, err := exchangeTool.ConvertToUSD(collateralID, lockedCollatersExceptPorting[collateralID])
		if err != nil {
			Logger.log.Errorf("Error when converting locked public token to prv: %v", err)
			return [][]string{rejectInst}, nil
		}
		totalAmountInUSD = totalAmountInUSD + pubTokenAmountInUSDT
	}

	// convert holding token to usd
	hodTokenAmountInUSDT, err := exchangeTool.ConvertToUSD(actionData.Meta.TokenID, custodianState.GetHoldingPublicTokens()[actionData.Meta.TokenID])
	if err != nil {
		Logger.log.Errorf("Error when converting holding public token to prv: %v", err)
		return [][]string{rejectInst}, nil
	}
	totalHoldAmountInUSDBigInt := new(big.Int).Mul(new(big.Int).SetUint64(hodTokenAmountInUSDT), new(big.Int).SetUint64(portalParams.MinUnlockOverRateCollaterals))
	minHoldUnlockedAmountInBigInt := new(big.Int).Div(totalHoldAmountInUSDBigInt, big.NewInt(10))
	if minHoldUnlockedAmountInBigInt.Cmp(new(big.Int).SetUint64(totalAmountInUSD)) >= 0 {
		Logger.log.Errorf("Error locked collaterals amount not enough to unlock")
		return [][]string{rejectInst}, nil
	}
	amountToUnlock := big.NewInt(0).Sub(new(big.Int).SetUint64(totalAmountInUSD), minHoldUnlockedAmountInBigInt).Uint64()
	err = updateCustodianStateAfterReqUnlockCollateralV3(custodianState, amountToUnlock, actionData.Meta.TokenID, portalParams, currentPortalState)
	if err != nil {
		Logger.log.Errorf("Error when converting holding public token to prv: %v", err)
		return [][]string{rejectInst}, nil
	}

	unlockCollateralsResult := metadata.NewUnlockOverRateCollateralsRequestStatus(
		common.PortalUnlockOverRateCollateralsAcceptedStatus,
		actionData.Meta.CustodianAddressStr,
		actionData.Meta.TokenID,
		amountToUnlock,
	)
	if currentPortalState.UnlockOverRateCollaterals != nil {
		currentPortalState.UnlockOverRateCollaterals[actionData.TxReqID.String()] = unlockCollateralsResult
	} else {
		//new object
		newUnlockOverRateCollaterals := make(map[string]*metadata.UnlockOverRateCollateralsRequestStatus)
		newUnlockOverRateCollaterals[actionData.TxReqID.String()] = unlockCollateralsResult

		currentPortalState.UnlockOverRateCollaterals = newUnlockOverRateCollaterals
	}

	inst := buildReqUnlockOverRateCollateralsInst(
		actionData.Meta.CustodianAddressStr,
		actionData.Meta.TokenID,
		amountToUnlock,
		metaType,
		shardID,
		common.PortalCusUnlockOverRateCollateralsAcceptedChainStatus,
	)

	return [][]string{inst}, nil
}
