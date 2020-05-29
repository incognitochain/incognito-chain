package blockchain

import (
	"encoding/json"
	"fmt"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/metadata"
)

func (blockchain *BlockChain) processPortalLiquidateCustodian(
	stateDB *statedb.StateDB,
	beaconHeight uint64,
	instructions []string,
	currentPortalState *CurrentPortalState,
	portalParams PortalParams, ) error {

	// unmarshal instructions content
	var actionData metadata.PortalLiquidateCustodianContent
	err := json.Unmarshal([]byte(instructions[3]), &actionData)
	if err != nil {
		Logger.log.Errorf("Can not unmarshal instruction content %v - Error %v\n", instructions[3], err)
		return nil
	}

	reqStatus := instructions[2]
	if reqStatus == common.PortalLiquidateCustodianSuccessChainStatus {
		// update custodian state
		Logger.log.Infof("[processPortalLiquidateCustodian] actionData.CustodianIncAddressStr = %s in beaconHeight=%d", actionData.CustodianIncAddressStr, beaconHeight)
		cusStateKey := statedb.GenerateCustodianStateObjectKey(actionData.CustodianIncAddressStr)
		cusStateKeyStr := cusStateKey.String()
		custodianState, ok := currentPortalState.CustodianPoolState[cusStateKeyStr]
		if !ok {
			Logger.log.Errorf("[processPortalLiquidateCustodian] cusStateKeyStr %s can not found", cusStateKeyStr)
			return nil
		}

		err := updateCustodianStateAfterLiquidateCustodian(custodianState, actionData.LiquidatedCollateralAmount, actionData.RemainUnlockAmountForCustodian, actionData.TokenID)
		if err != nil {
			Logger.log.Errorf("[processPortalLiquidateCustodian] Error when update custodian state after liquidation %v", err)
			return nil
		}

		// remove matching custodian from matching custodians list in matched redeem request
		matchedRedeemReqKey := statedb.GenerateMatchedRedeemRequestObjectKey(actionData.UniqueRedeemID)
		matchedRedeemReqKeyStr := matchedRedeemReqKey.String()

		updatedCustodians, err := removeCustodianFromMatchingRedeemCustodians(
			currentPortalState.MatchedRedeemRequests[matchedRedeemReqKeyStr].GetCustodians(), actionData.CustodianIncAddressStr)
		currentPortalState.MatchedRedeemRequests[matchedRedeemReqKeyStr].SetCustodians(updatedCustodians)
		if err != nil {
			Logger.log.Errorf("[processPortalLiquidateCustodian] Error when removing custodian from matching custodians %v", err)
			return nil
		}

		// remove redeem request from matched redeem requests list
		if len(currentPortalState.MatchedRedeemRequests[matchedRedeemReqKeyStr].GetCustodians()) == 0 {
			deleteMatchedRedeemRequest(currentPortalState, matchedRedeemReqKeyStr)
			statedb.DeleteMatchedRedeemRequest(stateDB,  actionData.UniqueRedeemID)

			// update status of redeem request with redeemID to liquidated status
			err = updateRedeemRequestStatusByRedeemId(actionData.UniqueRedeemID, common.PortalRedeemReqLiquidatedStatus, stateDB)
			if err != nil {
				Logger.log.Errorf("ERROR: an error occurred while updating redeem request status by redeemID: %+v", err)
				return nil
			}
		}

		// track liquidation custodian status by redeemID and custodian address into DB
		custodianLiquidationTrackData := metadata.PortalLiquidateCustodianStatus{
			Status:                         common.PortalLiquidateCustodianSuccessStatus,
			UniqueRedeemID:                 actionData.UniqueRedeemID,
			TokenID:                        actionData.TokenID,
			RedeemPubTokenAmount:           actionData.RedeemPubTokenAmount,
			LiquidatedCollateralAmount:     actionData.LiquidatedCollateralAmount,
			RemainUnlockAmountForCustodian: actionData.RemainUnlockAmountForCustodian,
			RedeemerIncAddressStr:          actionData.RedeemerIncAddressStr,
			CustodianIncAddressStr:         actionData.CustodianIncAddressStr,
			LiquidatedByExchangeRate:       actionData.LiquidatedByExchangeRate,
			ShardID:                        actionData.ShardID,
			LiquidatedBeaconHeight:         beaconHeight + 1,
		}
		custodianLiquidationTrackDataBytes, _ := json.Marshal(custodianLiquidationTrackData)
		err = statedb.StorePortalLiquidationCustodianRunAwayStatus(
			stateDB,
			actionData.UniqueRedeemID,
			actionData.CustodianIncAddressStr,
			custodianLiquidationTrackDataBytes,
		)
		if err != nil {
			Logger.log.Errorf("ERROR: an error occured while tracking liquidation custodian: %+v", err)
			return nil
		}

	} else if reqStatus == common.PortalLiquidateCustodianFailedChainStatus {
		// track liquidation custodian status by redeemID and custodian address into DB
		custodianLiquidationTrackData := metadata.PortalLiquidateCustodianStatus{
			Status:                         common.PortalLiquidateCustodianFailedStatus,
			UniqueRedeemID:                 actionData.UniqueRedeemID,
			TokenID:                        actionData.TokenID,
			RedeemPubTokenAmount:           actionData.RedeemPubTokenAmount,
			LiquidatedCollateralAmount:     actionData.LiquidatedCollateralAmount,
			RemainUnlockAmountForCustodian: actionData.RemainUnlockAmountForCustodian,
			RedeemerIncAddressStr:          actionData.RedeemerIncAddressStr,
			CustodianIncAddressStr:         actionData.CustodianIncAddressStr,
			LiquidatedByExchangeRate:       actionData.LiquidatedByExchangeRate,
			ShardID:                        actionData.ShardID,
			LiquidatedBeaconHeight:         beaconHeight + 1,
		}
		custodianLiquidationTrackDataBytes, _ := json.Marshal(custodianLiquidationTrackData)
		err = statedb.StorePortalLiquidationCustodianRunAwayStatus(
			stateDB,
			actionData.UniqueRedeemID,
			actionData.CustodianIncAddressStr,
			custodianLiquidationTrackDataBytes,
		)
		if err != nil {
			Logger.log.Errorf("ERROR: an error occured while tracking liquidation custodian: %+v", err)
			return nil
		}
	}

	return nil
}

func (blockchain *BlockChain) processLiquidationTopPercentileExchangeRates(
	portalStateDB *statedb.StateDB,
	beaconHeight uint64,
	instructions []string,
	currentPortalState *CurrentPortalState,
	portalParams PortalParams) error {

	// unmarshal instructions content
	var actionData metadata.PortalLiquidateTopPercentileExchangeRatesContent
	err := json.Unmarshal([]byte(instructions[3]), &actionData)
	if err != nil {
		Logger.log.Errorf("Can not unmarshal instruction content %v - Error %v\n", instructions[3], err)
		return nil
	}

	Logger.log.Infof("start processLiquidationTopPercentileExchangeRates with data %#v", actionData)

	cusStateKey := statedb.GenerateCustodianStateObjectKey(actionData.CustodianAddress)
	cusStateKeyStr := cusStateKey.String()
	custodianState, ok := currentPortalState.CustodianPoolState[cusStateKeyStr]
	if !ok || custodianState == nil {
		Logger.log.Errorf("Custodian not found")
		return nil
	}

	reqStatus := instructions[2]
	if reqStatus == common.PortalLiquidateTPExchangeRatesSuccessChainStatus {
		//validation
		Logger.log.Infof("custodian address %v, hold ptoken %+v, lock amount %+v", custodianState.GetIncognitoAddress(), custodianState.GetHoldingPublicTokens(), custodianState.GetLockedAmountCollateral())

		detectTp := actionData.TP
		if len(detectTp) > 0 {
			//update current portal state
			Logger.log.Infof("start update liquidation %#v", currentPortalState)
			updateCurrentPortalStateOfLiquidationExchangeRates(currentPortalState, cusStateKeyStr, custodianState, detectTp, actionData.RemainUnlockAmount)
			Logger.log.Infof("end update liquidation %#v", currentPortalState)

			//save db
			beaconHeightBytes := []byte(fmt.Sprintf("%d-", beaconHeight))
			newTPKey := beaconHeightBytes
			newTPKey = append(newTPKey, []byte(custodianState.GetIncognitoAddress())...)
			newTPExchangeRates := metadata.NewLiquidateTopPercentileExchangeRatesStatus(
				custodianState.GetIncognitoAddress(),
				common.PortalLiquidationTPExchangeRatesSuccessStatus,
				detectTp,
			)

			contentStatusBytes, _ := json.Marshal(newTPExchangeRates)
			err = statedb.TrackPortalStateStatusMultiple(
				portalStateDB,
				statedb.PortalLiquidationTpExchangeRatesStatusPrefix(),
				newTPKey,
				contentStatusBytes,
				beaconHeight,
			)

			if err != nil {
				Logger.log.Errorf("ERROR: an error occurred while store liquidation TP exchange rates %v", err)
				return nil
			}
		}
	} else if reqStatus == common.PortalLiquidateTPExchangeRatesFailedChainStatus {
		beaconHeightBytes := []byte(fmt.Sprintf("%d-", beaconHeight))
		newTPKey := beaconHeightBytes
		newTPKey = append(newTPKey, []byte(custodianState.GetIncognitoAddress())...)
		newTPExchangeRates := metadata.NewLiquidateTopPercentileExchangeRatesStatus(
			custodianState.GetIncognitoAddress(),
			common.PortalLiquidationTPExchangeRatesFailedStatus,
			nil,
		)
		contentStatusBytes, _ := json.Marshal(newTPExchangeRates)
		err = statedb.TrackPortalStateStatusMultiple(
			portalStateDB,
			statedb.PortalLiquidationTpExchangeRatesStatusPrefix(),
			newTPKey,
			contentStatusBytes,
			beaconHeight,
		)

		if err != nil {
			Logger.log.Errorf("ERROR: an error occurred while store liquidation TP exchange rates %v", err)
			return nil
		}
	}

	return nil
}

func (blockchain *BlockChain) processPortalRedeemLiquidateExchangeRates(
	portalStateDB *statedb.StateDB,
	beaconHeight uint64,
	instructions []string,
	currentPortalState *CurrentPortalState,
	portalParams PortalParams,
	updatingInfoByTokenID map[common.Hash]UpdatingInfo) error {
	if currentPortalState == nil {
		Logger.log.Errorf("current portal state is nil")
		return nil
	}

	if len(instructions) != 4 {
		return nil // skip the instruction
	}

	// unmarshal instructions content
	var actionData metadata.PortalRedeemLiquidateExchangeRatesContent
	err := json.Unmarshal([]byte(instructions[3]), &actionData)
	if err != nil {
		Logger.log.Errorf("Can not unmarshal instruction content %v - Error %v\n", instructions[3], err)
		return nil
	}

	reqStatus := instructions[2]
	if reqStatus == common.PortalRedeemLiquidateExchangeRatesSuccessChainStatus {
		liquidateExchangeRatesKey := statedb.GeneratePortalLiquidationPoolObjectKey()
		liquidateExchangeRates, ok := currentPortalState.LiquidationPool[liquidateExchangeRatesKey.String()]

		if !ok {
			Logger.log.Errorf("Liquidate exchange rates not found")
			return nil
		}

		liquidateByTokenID, ok := liquidateExchangeRates.Rates()[actionData.TokenID]
		if !ok {
			Logger.log.Errorf("Liquidate exchange rates not found")
			return nil
		}

		totalPrv := actionData.TotalPTokenReceived

		liquidateExchangeRates.Rates()[actionData.TokenID] = statedb.LiquidationPoolDetail{
			CollateralAmount: liquidateByTokenID.CollateralAmount - totalPrv,
			PubTokenAmount:   liquidateByTokenID.PubTokenAmount - actionData.RedeemAmount,
		}

		currentPortalState.LiquidationPool[liquidateExchangeRatesKey.String()] = liquidateExchangeRates

		Logger.log.Infof("Redeem Liquidation: Amount refund to user amount ptoken %v, amount prv %v", actionData.RedeemAmount, totalPrv)

		redeem := metadata.NewRedeemLiquidateExchangeRatesStatus(
			actionData.TxReqID,
			actionData.TokenID,
			actionData.RedeemerIncAddressStr,
			actionData.RedeemAmount,
			common.PortalRedeemLiquidateExchangeRatesSuccessStatus,
			totalPrv,
		)

		contentStatusBytes, _ := json.Marshal(redeem)
		err = statedb.TrackPortalStateStatusMultiple(
			portalStateDB,
			statedb.PortalLiquidationRedeemRequestStatusPrefix(),
			[]byte(actionData.TxReqID.String()),
			contentStatusBytes,
			beaconHeight,
		)

		if err != nil {
			Logger.log.Errorf("Store redeem liquidate exchange rates error %v\n", err)
			return nil
		}

		// update bridge/portal token info
		incTokenID, err := common.Hash{}.NewHashFromStr(actionData.TokenID)
		if err != nil {
			Logger.log.Errorf("ERROR: Can not new hash from porting incTokenID: %+v", err)
			return nil
		}
		updatingInfo, found := updatingInfoByTokenID[*incTokenID]
		if found {
			updatingInfo.deductAmt += actionData.RedeemAmount
		} else {
			updatingInfo = UpdatingInfo{
				countUpAmt:      0,
				deductAmt:       actionData.RedeemAmount,
				tokenID:         *incTokenID,
				externalTokenID: nil,
				isCentralized:   false,
			}
		}
		updatingInfoByTokenID[*incTokenID] = updatingInfo
	} else if reqStatus == common.PortalRedeemLiquidateExchangeRatesRejectedChainStatus {
		redeem := metadata.NewRedeemLiquidateExchangeRatesStatus(
			actionData.TxReqID,
			actionData.TokenID,
			actionData.RedeemerIncAddressStr,
			actionData.RedeemAmount,
			common.PortalRedeemLiquidateExchangeRatesRejectedStatus,
			0,
		)

		contentStatusBytes, _ := json.Marshal(redeem)
		err = statedb.TrackPortalStateStatusMultiple(
			portalStateDB,
			statedb.PortalLiquidationRedeemRequestStatusPrefix(),
			[]byte(actionData.TxReqID.String()),
			contentStatusBytes,
			beaconHeight,
		)
		if err != nil {
			Logger.log.Errorf("Store redeem liquidate exchange rates error %v\n", err)
			return nil
		}
	}

	return nil
}

func (blockchain *BlockChain) processPortalTopUpWaitingPorting(
	portalStateDB *statedb.StateDB,
	beaconHeight uint64,
	instructions []string,
	currentPortalState *CurrentPortalState,
	portalParams PortalParams,
) error {
	if currentPortalState == nil {
		Logger.log.Errorf("current portal state is nil")
		return nil
	}
	if len(instructions) != 4 {
		return nil // skip the instruction
	}

	var actionData metadata.PortalTopUpWaitingPortingRequestContent
	err := json.Unmarshal([]byte(instructions[3]), &actionData)
	if err != nil {
		Logger.log.Errorf("Error when unmarshaling portal top up waiting porting action %v - %v", instructions[3], err)
		return nil
	}

	depositStatus := instructions[2]
	if depositStatus == common.PortalTopUpWaitingPortingRejectedChainStatus {
		err = trackPortalStateStatus(
			beaconHeight,
			portalStateDB,
			actionData,
			common.PortalTopUpWaitingPortingRejectedStatus,
		)
		if err != nil {
			Logger.log.Errorf("ERROR: an error occurred while storing waiting porting top up error %v", err)
		}
	} else if depositStatus == common.PortalTopUpWaitingPortingSuccessChainStatus {
		custodianStateKey := statedb.GenerateCustodianStateObjectKey(actionData.IncogAddressStr)
		custodian, ok := currentPortalState.CustodianPoolState[custodianStateKey.String()]
		if !ok {
			Logger.log.Errorf("Custodian not found")
			return nil
		}

		waitingPortingRequestKey := statedb.GeneratePortalWaitingPortingRequestObjectKey(actionData.PortingID)
		waitingPortingReq, ok := currentPortalState.WaitingPortingRequests[waitingPortingRequestKey.String()]
		if !ok || waitingPortingReq == nil || waitingPortingReq.TokenID() != actionData.PTokenID {
			Logger.log.Errorf("Waiting porting request with portingID (%s) not found", actionData.PortingID)
			return nil
		}

		custodian.SetTotalCollateral(custodian.GetTotalCollateral() + actionData.DepositedAmount)
		
		topUpAmt := actionData.DepositedAmount
		if actionData.FreeCollateralAmount > 0 {
			topUpAmt += actionData.FreeCollateralAmount
			custodian.SetFreeCollateral(custodian.GetFreeCollateral() - actionData.FreeCollateralAmount)
		}

		lockedAmountCollateral := custodian.GetLockedAmountCollateral()
		lockedAmountCollateral[actionData.PTokenID] += topUpAmt
		custodian.SetLockedAmountCollateral(lockedAmountCollateral)
		custodiansByPortingID := waitingPortingReq.Custodians()
		for _, cus := range custodiansByPortingID {
			if cus.IncAddress == actionData.IncogAddressStr {
				cus.LockedAmountCollateral += topUpAmt
			}
		}
		waitingPortingReq.SetCustodians(custodiansByPortingID)
		currentPortalState.CustodianPoolState[custodianStateKey.String()] = custodian
		currentPortalState.WaitingPortingRequests[waitingPortingRequestKey.String()] = waitingPortingReq

		err = trackPortalStateStatus(
			beaconHeight,
			portalStateDB,
			actionData,
			common.PortalTopUpWaitingPortingSuccessStatus,
		)
		if err != nil {
			Logger.log.Errorf("ERROR: an error occurred while storing waiting porting top up error %v", err)
			return nil
		}

		// update state of porting request by portingID
		newPortingRequestState := metadata.NewPortingRequestStatus(
			waitingPortingReq.UniquePortingID(),
			waitingPortingReq.TxReqID(),
			waitingPortingReq.TokenID(),
			waitingPortingReq.PorterAddress(),
			waitingPortingReq.Amount(),
			waitingPortingReq.Custodians(),
			waitingPortingReq.PortingFee(),
			common.PortalPortingReqWaitingStatus,
			beaconHeight+1,
		)
		newPortingRequestStatusBytes, _ := json.Marshal(newPortingRequestState)
		err = statedb.TrackPortalStateStatusMultiple(
			portalStateDB,
			statedb.PortalPortingRequestStatusPrefix(),
			[]byte(waitingPortingReq.UniquePortingID()),
			newPortingRequestStatusBytes,
			beaconHeight,
		)
		if err != nil {
			Logger.log.Errorf("ERROR: an error occurred while store porting request item: %+v", err)
			return nil
		}
	}
	return nil
}

func trackPortalStateStatus(
	beaconHeight uint64,
	portalStateDB *statedb.StateDB,
	actionData metadata.PortalTopUpWaitingPortingRequestContent,
	status byte,
) error {
	topUpWaitingPortingReq := metadata.NewPortalTopUpWaitingPortingRequestStatus(
		actionData.TxReqID,
		actionData.PortingID,
		actionData.IncogAddressStr,
		actionData.PTokenID,
		actionData.DepositedAmount,
		actionData.FreeCollateralAmount,
		status,
	)
	statusContentBytes, _ := json.Marshal(topUpWaitingPortingReq)
	return statedb.TrackPortalStateStatusMultiple(
		portalStateDB,
		statedb.PortalTopUpWaitingPortingStatusPrefix(),
		[]byte(actionData.TxReqID.String()),
		statusContentBytes,
		beaconHeight,
	)
}

func (blockchain *BlockChain) processPortalLiquidationCustodianDeposit(
	portalStateDB *statedb.StateDB,
	beaconHeight uint64,
	instructions []string,
	currentPortalState *CurrentPortalState,
	portalParams PortalParams) error {
	if currentPortalState == nil {
		Logger.log.Errorf("current portal state is nil")
		return nil
	}
	if len(instructions) != 4 {
		return nil // skip the instruction
	}

	// unmarshal instructions content
	var actionData metadata.PortalLiquidationCustodianDepositContentV2
	err := json.Unmarshal([]byte(instructions[3]), &actionData)
	if err != nil {
		Logger.log.Errorf("Error when unmarshaling portal liquidation custodian deposit content %v - %v", instructions[3], err)
		return nil
	}

	depositStatus := instructions[2]

	if depositStatus == common.PortalLiquidationCustodianDepositSuccessChainStatus {
		custodianStateKey := statedb.GenerateCustodianStateObjectKey(actionData.IncogAddressStr)
		custodianStateKeyStr := custodianStateKey.String()
		custodian, ok := currentPortalState.CustodianPoolState[custodianStateKeyStr]
		if !ok {
			Logger.log.Errorf("Custodian not found")
			return nil
		}

		custodian.SetTotalCollateral(custodian.GetTotalCollateral() + actionData.DepositedAmount)

		lockedAmountCollateral := custodian.GetLockedAmountCollateral()
		topUpAmt := actionData.DepositedAmount
		if actionData.FreeCollateralAmount > 0 {
			topUpAmt += actionData.FreeCollateralAmount
			custodian.SetFreeCollateral(custodian.GetFreeCollateral() - actionData.FreeCollateralAmount)
		}
		lockedAmountCollateral[actionData.PTokenId] += topUpAmt
		custodian.SetLockedAmountCollateral(lockedAmountCollateral)
		currentPortalState.CustodianPoolState[custodianStateKeyStr] = custodian

		newLiquidationCustodianDeposit := metadata.NewLiquidationCustodianDepositStatusV2(
			actionData.TxReqID,
			actionData.IncogAddressStr,
			actionData.PTokenId,
			actionData.DepositedAmount,
			actionData.FreeCollateralAmount,
			common.PortalLiquidationCustodianDepositSuccessStatus,
		)

		contentStatusBytes, _ := json.Marshal(newLiquidationCustodianDeposit)
		err = statedb.TrackPortalStateStatusMultiple(
			portalStateDB,
			statedb.PortalLiquidationCustodianDepositStatusPrefix(),
			[]byte(actionData.TxReqID.String()),
			contentStatusBytes,
			beaconHeight,
		)

		if err != nil {
			Logger.log.Errorf("ERROR: an error occurred while store liquidation custodian deposit error %v", err)
			return nil
		}
	} else if depositStatus == common.PortalLiquidationCustodianDepositRejectedChainStatus {
		newLiquidationCustodianDeposit := metadata.NewLiquidationCustodianDepositStatusV2(
			actionData.TxReqID,
			actionData.IncogAddressStr,
			actionData.PTokenId,
			actionData.DepositedAmount,
			actionData.FreeCollateralAmount,
			common.PortalLiquidationCustodianDepositRejectedStatus,
		)

		contentStatusBytes, _ := json.Marshal(newLiquidationCustodianDeposit)
		err = statedb.TrackPortalStateStatusMultiple(
			portalStateDB,
			statedb.PortalLiquidationCustodianDepositStatusPrefix(),
			[]byte(actionData.TxReqID.String()),
			contentStatusBytes,
			beaconHeight,
		)

		if err != nil {
			Logger.log.Errorf("ERROR: an error occurred while store liquidation custodian deposit error %v", err)
			return nil
		}
	}

	return nil
}

func (blockchain *BlockChain) processPortalExpiredPortingRequest(
	stateDB *statedb.StateDB,
	beaconHeight uint64,
	instructions []string,
	currentPortalState *CurrentPortalState,
	portalParams PortalParams) error {
	if currentPortalState == nil {
		Logger.log.Errorf("current portal state is nil")
		return nil
	}
	if len(instructions) != 4 {
		return nil // skip the instruction
	}

	// unmarshal instructions content
	var actionData metadata.PortalExpiredWaitingPortingReqContent
	err := json.Unmarshal([]byte(instructions[3]), &actionData)
	if err != nil {
		Logger.log.Errorf("Error when unmarshaling portal expired waiting porting content %v - %v", instructions[3], err)
		return nil
	}

	status := instructions[2]
	waitingPortingID := actionData.UniquePortingID

	if status == common.PortalExpiredWaitingPortingReqSuccessChainStatus {
		waitingPortingKey := statedb.GeneratePortalWaitingPortingRequestObjectKey(waitingPortingID)
		waitingPortingKeyStr := waitingPortingKey.String()
		waitingPortingReq := currentPortalState.WaitingPortingRequests[waitingPortingKeyStr]
		if waitingPortingReq == nil {
			Logger.log.Errorf("[processPortalExpiredPortingRequest] waiting porting req nil with key : %v", waitingPortingKey)
			return nil
		}

		// get tokenID from redeemTokenID
		tokenID := waitingPortingReq.TokenID()

		// update custodian state in matching custodians list (holding public tokens, locked amount)
		for _, matchCusDetail := range waitingPortingReq.Custodians() {
			cusStateKey := statedb.GenerateCustodianStateObjectKey(matchCusDetail.IncAddress)
			cusStateKeyStr := cusStateKey.String()
			custodianState := currentPortalState.CustodianPoolState[cusStateKeyStr]
			if custodianState == nil {
				Logger.log.Errorf("[checkAndBuildInstForExpiredWaitingPortingRequest] Error when get custodian state with key %v\n: ", cusStateKey)
				continue
			}
			updateCustodianStateAfterExpiredPortingReq(custodianState, matchCusDetail.LockedAmountCollateral, tokenID)
		}

		// remove waiting porting request from waiting list
		delete(currentPortalState.WaitingPortingRequests, waitingPortingKeyStr)
		statedb.DeleteWaitingPortingRequest(stateDB, waitingPortingReq.UniquePortingID())

		// update status of porting ID  => expired/liquidated
		portingReqStatus := common.PortalPortingReqExpiredStatus
		if actionData.ExpiredByLiquidation {
			portingReqStatus = common.PortalPortingReqLiquidatedStatus
		}

		newPortingRequestStatus := metadata.NewPortingRequestStatus(
			waitingPortingReq.UniquePortingID(),
			waitingPortingReq.TxReqID(),
			tokenID,
			waitingPortingReq.PorterAddress(),
			waitingPortingReq.Amount(),
			waitingPortingReq.Custodians(),
			waitingPortingReq.PortingFee(),
			portingReqStatus,
			waitingPortingReq.BeaconHeight())

		newPortingRequestStatusBytes, _ := json.Marshal(newPortingRequestStatus)
		err = statedb.TrackPortalStateStatusMultiple(
			stateDB,
			statedb.PortalPortingRequestStatusPrefix(),
			[]byte(actionData.UniquePortingID),
			newPortingRequestStatusBytes,
			beaconHeight,
		)
		if err != nil {
			Logger.log.Errorf("ERROR: an error occurred while store porting request item status: %+v", err)
			return nil
		}

		// track expired waiting porting request status by portingID into DB
		expiredPortingTrackData := metadata.PortalExpiredWaitingPortingReqStatus{
			Status:               common.PortalExpiredPortingReqSuccessStatus,
			UniquePortingID:      waitingPortingID,
			ShardID:              actionData.ShardID,
			ExpiredByLiquidation: actionData.ExpiredByLiquidation,
			ExpiredBeaconHeight:  beaconHeight + 1,
		}
		expiredPortingTrackDataBytes, _ := json.Marshal(expiredPortingTrackData)
		err = statedb.StorePortalExpiredPortingRequestStatus(
			stateDB,
			waitingPortingID,
			expiredPortingTrackDataBytes,
		)
		if err != nil {
			Logger.log.Errorf("ERROR: an error occured while tracking expired porting request: %+v", err)
			return nil
		}

	} else if status == common.PortalLiquidationCustodianDepositRejectedChainStatus {
		// track expired waiting porting request status by portingID into DB
		expiredPortingTrackData := metadata.PortalExpiredWaitingPortingReqStatus{
			Status:               common.PortalExpiredPortingReqFailedStatus,
			UniquePortingID:      waitingPortingID,
			ShardID:              actionData.ShardID,
			ExpiredByLiquidation: actionData.ExpiredByLiquidation,
			ExpiredBeaconHeight:  beaconHeight + 1,
		}
		expiredPortingTrackDataBytes, _ := json.Marshal(expiredPortingTrackData)
		err = statedb.StorePortalExpiredPortingRequestStatus(
			stateDB,
			waitingPortingID,
			expiredPortingTrackDataBytes,
		)
		if err != nil {
			Logger.log.Errorf("ERROR: an error occured while tracking expired porting request: %+v", err)
			return nil
		}
	}

	return nil
}
