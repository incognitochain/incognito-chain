package portalprocess

//func (blockchain *BlockChain) processPortalLiquidateCustodian(
//	stateDB *statedb.StateDB,
//	beaconHeight uint64,
//	instructions []string,
//	currentPortalState *CurrentPortalState,
//	portalParams portal.PortalParams) error {
//
//	// unmarshal instructions content
//	var actionData metadata2.PortalLiquidateCustodianContent
//	err := json.Unmarshal([]byte(instructions[3]), &actionData)
//	if err != nil {
//		Logger.log.Errorf("Can not unmarshal instruction content %v - Error %v\n", instructions[3], err)
//		return nil
//	}
//
//	reqStatus := instructions[2]
//	if reqStatus == common.PortalLiquidateCustodianSuccessChainStatus {
//		// update custodian state
//		Logger.log.Infof("[processPortalLiquidateCustodian] actionData.CustodianIncAddressStr = %s in beaconHeight=%d", actionData.CustodianIncAddressStr, beaconHeight)
//		cusStateKey := statedb.GenerateCustodianStateObjectKey(actionData.CustodianIncAddressStr)
//		cusStateKeyStr := cusStateKey.String()
//		custodianState, ok := currentPortalState.CustodianPoolState[cusStateKeyStr]
//		if !ok {
//			Logger.log.Errorf("[processPortalLiquidateCustodian] cusStateKeyStr %s can not found", cusStateKeyStr)
//			return nil
//		}
//
//		if instructions[0] == strconv.Itoa(metadata.PortalLiquidateCustodianMeta) {
//			err = updateCustodianStateAfterLiquidateCustodian(custodianState, actionData.LiquidatedCollateralAmount, actionData.RemainUnlockAmountForCustodian, actionData.TokenID)
//		} else {
//			err = updateCustodianStateAfterLiquidateCustodianV3(custodianState, actionData.LiquidatedCollateralAmount, actionData.RemainUnlockAmountForCustodian, actionData.LiquidatedCollateralAmounts, actionData.RemainUnlockAmountsForCustodian, actionData.TokenID)
//		}
//
//		if err != nil {
//			Logger.log.Errorf("[processPortalLiquidateCustodian] Error when update custodian state after liquidation %v", err)
//			return nil
//		}
//
//		// remove matching custodian from matching custodians list in matched redeem request
//		matchedRedeemReqKey := statedb.GenerateMatchedRedeemRequestObjectKey(actionData.UniqueRedeemID)
//		matchedRedeemReqKeyStr := matchedRedeemReqKey.String()
//
//		updatedCustodians, err := removeCustodianFromMatchingRedeemCustodians(
//			currentPortalState.MatchedRedeemRequests[matchedRedeemReqKeyStr].GetCustodians(), actionData.CustodianIncAddressStr)
//		currentPortalState.MatchedRedeemRequests[matchedRedeemReqKeyStr].SetCustodians(updatedCustodians)
//		if err != nil {
//			Logger.log.Errorf("[processPortalLiquidateCustodian] Error when removing custodian from matching custodians %v", err)
//			return nil
//		}
//
//		// remove redeem request from matched redeem requests list
//		if len(currentPortalState.MatchedRedeemRequests[matchedRedeemReqKeyStr].GetCustodians()) == 0 {
//			deleteMatchedRedeemRequest(currentPortalState, matchedRedeemReqKeyStr)
//			statedb.DeleteMatchedRedeemRequest(stateDB, actionData.UniqueRedeemID)
//
//			// update status of redeem request with redeemID to liquidated status
//			err = updateRedeemRequestStatusByRedeemId(actionData.UniqueRedeemID, common.PortalRedeemReqLiquidatedStatus, stateDB)
//			if err != nil {
//				Logger.log.Errorf("ERROR: an error occurred while updating redeem request status by redeemID: %+v", err)
//				return nil
//			}
//		}
//
//		// track liquidation custodian status by redeemID and custodian address into DB
//		custodianLiquidationTrackData := metadata2.PortalLiquidateCustodianStatus{
//			Status:                          common.PortalLiquidateCustodianSuccessStatus,
//			UniqueRedeemID:                  actionData.UniqueRedeemID,
//			TokenID:                         actionData.TokenID,
//			RedeemPubTokenAmount:            actionData.RedeemPubTokenAmount,
//			LiquidatedCollateralAmount:      actionData.LiquidatedCollateralAmount,
//			RemainUnlockAmountForCustodian:  actionData.RemainUnlockAmountForCustodian,
//			LiquidatedCollateralAmounts:     actionData.LiquidatedCollateralAmounts,
//			RemainUnlockAmountsForCustodian: actionData.RemainUnlockAmountsForCustodian,
//			RedeemerIncAddressStr:           actionData.RedeemerIncAddressStr,
//			CustodianIncAddressStr:          actionData.CustodianIncAddressStr,
//			LiquidatedByExchangeRate:        actionData.LiquidatedByExchangeRate,
//			ShardID:                         actionData.ShardID,
//			LiquidatedBeaconHeight:          beaconHeight + 1,
//		}
//		custodianLiquidationTrackDataBytes, _ := json.Marshal(custodianLiquidationTrackData)
//		err = statedb.StorePortalLiquidationCustodianRunAwayStatus(
//			stateDB,
//			actionData.UniqueRedeemID,
//			actionData.CustodianIncAddressStr,
//			custodianLiquidationTrackDataBytes,
//		)
//		if err != nil {
//			Logger.log.Errorf("ERROR: an error occured while tracking liquidation custodian: %+v", err)
//			return nil
//		}
//
//	} else if reqStatus == common.PortalLiquidateCustodianFailedChainStatus {
//		// track liquidation custodian status by redeemID and custodian address into DB
//		custodianLiquidationTrackData := metadata2.PortalLiquidateCustodianStatus{
//			Status:                          common.PortalLiquidateCustodianFailedStatus,
//			UniqueRedeemID:                  actionData.UniqueRedeemID,
//			TokenID:                         actionData.TokenID,
//			RedeemPubTokenAmount:            actionData.RedeemPubTokenAmount,
//			LiquidatedCollateralAmount:      actionData.LiquidatedCollateralAmount,
//			RemainUnlockAmountForCustodian:  actionData.RemainUnlockAmountForCustodian,
//			LiquidatedCollateralAmounts:     actionData.LiquidatedCollateralAmounts,
//			RemainUnlockAmountsForCustodian: actionData.RemainUnlockAmountsForCustodian,
//			RedeemerIncAddressStr:           actionData.RedeemerIncAddressStr,
//			CustodianIncAddressStr:          actionData.CustodianIncAddressStr,
//			LiquidatedByExchangeRate:        actionData.LiquidatedByExchangeRate,
//			ShardID:                         actionData.ShardID,
//			LiquidatedBeaconHeight:          beaconHeight + 1,
//		}
//		custodianLiquidationTrackDataBytes, _ := json.Marshal(custodianLiquidationTrackData)
//		err = statedb.StorePortalLiquidationCustodianRunAwayStatus(
//			stateDB,
//			actionData.UniqueRedeemID,
//			actionData.CustodianIncAddressStr,
//			custodianLiquidationTrackDataBytes,
//		)
//		if err != nil {
//			Logger.log.Errorf("ERROR: an error occured while tracking liquidation custodian: %+v", err)
//			return nil
//		}
//	}
//
//	return nil
//}

// TODO:
//func (blockchain *BlockChain) processLiquidationTopPercentileExchangeRates(
////	portalStateDB *statedb.StateDB,
////	beaconHeight uint64,
////	instructions []string,
////	currentPortalState *CurrentPortalState,
////	portalParams portal.PortalParams) error {
////
////	// unmarshal instructions content
////	var actionData metadata2.PortalLiquidateTopPercentileExchangeRatesContent
////	err := json.Unmarshal([]byte(instructions[3]), &actionData)
////	if err != nil {
////		Logger.log.Errorf("Can not unmarshal instruction content %v - Error %v\n", instructions[3], err)
////		return nil
////	}
////
////	Logger.log.Infof("start processLiquidationTopPercentileExchangeRates with data %#v", actionData)
////
////	cusStateKey := statedb.GenerateCustodianStateObjectKey(actionData.CustodianAddress)
////	cusStateKeyStr := cusStateKey.String()
////	custodianState, ok := currentPortalState.CustodianPoolState[cusStateKeyStr]
////	if !ok || custodianState == nil {
////		Logger.log.Errorf("Custodian not found")
////		return nil
////	}
////
////	reqStatus := instructions[2]
////	if reqStatus == common.PortalLiquidateTPExchangeRatesSuccessChainStatus {
////		//validation
////		Logger.log.Infof("custodian address %v, hold ptoken %+v, lock amount %+v", custodianState.GetIncognitoAddress(), custodianState.GetHoldingPublicTokens(), custodianState.GetLockedAmountCollateral())
////
////		detectTp := actionData.TP
////		if len(detectTp) > 0 {
////			//update current portal state
////			Logger.log.Infof("start update liquidation %#v", currentPortalState)
////			updateCurrentPortalStateOfLiquidationExchangeRates(currentPortalState, cusStateKeyStr, custodianState, detectTp, actionData.RemainUnlockAmount)
////			Logger.log.Infof("end update liquidation %#v", currentPortalState)
////
////			//save db
////			newTPExchangeRates := metadata2.NewLiquidateTopPercentileExchangeRatesStatus(
////				custodianState.GetIncognitoAddress(),
////				common.PortalLiquidationTPExchangeRatesSuccessStatus,
////				detectTp,
////			)
////			contentStatusBytes, _ := json.Marshal(newTPExchangeRates)
////			err = statedb.StoreLiquidationByExchangeRateStatus(
////				portalStateDB,
////				beaconHeight,
////				custodianState.GetIncognitoAddress(),
////				contentStatusBytes)
////			if err != nil {
////				Logger.log.Errorf("ERROR: an error occurred while store liquidation TP exchange rates %v", err)
////				return nil
////			}
////		}
////	} else if reqStatus == common.PortalLiquidateTPExchangeRatesFailedChainStatus {
////		newTPExchangeRates := metadata2.NewLiquidateTopPercentileExchangeRatesStatus(
////			custodianState.GetIncognitoAddress(),
////			common.PortalLiquidationTPExchangeRatesFailedStatus,
////			nil,
////		)
////		contentStatusBytes, _ := json.Marshal(newTPExchangeRates)
////		err = statedb.StoreLiquidationByExchangeRateStatus(
////			portalStateDB,
////			beaconHeight,
////			custodianState.GetIncognitoAddress(),
////			contentStatusBytes)
////
////		if err != nil {
////			Logger.log.Errorf("ERROR: an error occurred while store liquidation TP exchange rates %v", err)
////			return nil
////		}
////	}
////
////	return nil
////}

//func (blockchain *BlockChain) processPortalRedeemLiquidateExchangeRates(
//	portalStateDB *statedb.StateDB,
//	beaconHeight uint64,
//	instructions []string,
//	currentPortalState *CurrentPortalState,
//	portalParams portal.PortalParams,
//	updatingInfoByTokenID map[common.Hash]basemeta.UpdatingInfo) error {
//	if currentPortalState == nil {
//		Logger.log.Errorf("current portal state is nil")
//		return nil
//	}
//
//	if len(instructions) != 4 {
//		return nil // skip the instruction
//	}
//
//	// unmarshal instructions content
//	var actionData metadata2.PortalRedeemLiquidateExchangeRatesContent
//	err := json.Unmarshal([]byte(instructions[3]), &actionData)
//	if err != nil {
//		Logger.log.Errorf("Can not unmarshal instruction content %v - Error %v\n", instructions[3], err)
//		return nil
//	}
//
//	reqStatus := instructions[2]
//	if reqStatus == common.PortalRedeemFromLiquidationPoolSuccessChainStatus {
//		liquidateExchangeRatesKey := statedb.GeneratePortalLiquidationPoolObjectKey()
//		liquidateExchangeRates, ok := currentPortalState.LiquidationPool[liquidateExchangeRatesKey.String()]
//
//		if !ok {
//			Logger.log.Errorf("Liquidate exchange rates not found")
//			return nil
//		}
//
//		liquidateByTokenID, ok := liquidateExchangeRates.Rates()[actionData.TokenID]
//		if !ok {
//			Logger.log.Errorf("Liquidate exchange rates not found")
//			return nil
//		}
//
//		totalPrv := actionData.TotalPTokenReceived
//
//		liquidateExchangeRates.Rates()[actionData.TokenID] = statedb.LiquidationPoolDetail{
//			CollateralAmount: liquidateByTokenID.CollateralAmount - totalPrv,
//			PubTokenAmount:   liquidateByTokenID.PubTokenAmount - actionData.RedeemAmount,
//		}
//
//		currentPortalState.LiquidationPool[liquidateExchangeRatesKey.String()] = liquidateExchangeRates
//
//		Logger.log.Infof("Redeem Liquidation: Amount refund to user amount ptoken %v, amount prv %v", actionData.RedeemAmount, totalPrv)
//
//		redeem := metadata2.NewRedeemLiquidateExchangeRatesStatus(
//			actionData.TxReqID,
//			actionData.TokenID,
//			actionData.RedeemerIncAddressStr,
//			actionData.RedeemAmount,
//			common.PortalRedeemFromLiquidationPoolSuccessStatus,
//			totalPrv,
//		)
//
//		contentStatusBytes, _ := json.Marshal(redeem)
//		err = statedb.StoreRedeemRequestFromLiquidationPoolByTxIDStatus(
//			portalStateDB,
//			actionData.TxReqID.String(),
//			contentStatusBytes,
//		)
//
//		if err != nil {
//			Logger.log.Errorf("Store redeem liquidate exchange rates error %v\n", err)
//			return nil
//		}
//
//		// update bridge/portal token info
//		incTokenID, err := common.Hash{}.NewHashFromStr(actionData.TokenID)
//		if err != nil {
//			Logger.log.Errorf("ERROR: Can not new hash from porting incTokenID: %+v", err)
//			return nil
//		}
//		updatingInfo, found := updatingInfoByTokenID[*incTokenID]
//		if found {
//			updatingInfo.deductAmt += actionData.RedeemAmount
//		} else {
//			updatingInfo = UpdatingInfo{
//				countUpAmt:      0,
//				deductAmt:       actionData.RedeemAmount,
//				tokenID:         *incTokenID,
//				externalTokenID: nil,
//				isCentralized:   false,
//			}
//		}
//		updatingInfoByTokenID[*incTokenID] = updatingInfo
//	} else if reqStatus == common.PortalRedeemFromLiquidationPoolRejectedChainStatus {
//		redeem := metadata2.NewRedeemLiquidateExchangeRatesStatus(
//			actionData.TxReqID,
//			actionData.TokenID,
//			actionData.RedeemerIncAddressStr,
//			actionData.RedeemAmount,
//			common.PortalRedeemFromLiquidationPoolRejectedStatus,
//			0,
//		)
//
//		contentStatusBytes, _ := json.Marshal(redeem)
//		err = statedb.StoreRedeemRequestFromLiquidationPoolByTxIDStatus(
//			portalStateDB,
//			actionData.TxReqID.String(),
//			contentStatusBytes,
//		)
//		if err != nil {
//			Logger.log.Errorf("Store redeem liquidate exchange rates error %v\n", err)
//			return nil
//		}
//	}
//
//	return nil
//}

//func (blockchain *BlockChain) processPortalTopUpWaitingPorting(
//	portalStateDB *statedb.StateDB,
//	beaconHeight uint64,
//	instructions []string,
//	currentPortalState *CurrentPortalState,
//	portalParams portal.PortalParams,
//) error {
//	if currentPortalState == nil {
//		Logger.log.Errorf("current portal state is nil")
//		return nil
//	}
//	if len(instructions) != 4 {
//		return nil // skip the instruction
//	}
//
//	var actionData metadata2.PortalTopUpWaitingPortingRequestContent
//	err := json.Unmarshal([]byte(instructions[3]), &actionData)
//	if err != nil {
//		Logger.log.Errorf("Error when unmarshaling portal top up waiting porting action %v - %v", instructions[3], err)
//		return nil
//	}
//
//	depositStatus := instructions[2]
//	if depositStatus == common.PortalTopUpWaitingPortingRejectedChainStatus {
//		topUpWaitingPortingReq := metadata2.NewPortalTopUpWaitingPortingRequestStatus(
//			actionData.TxReqID,
//			actionData.PortingID,
//			actionData.IncogAddressStr,
//			actionData.PTokenID,
//			actionData.DepositedAmount,
//			actionData.FreeCollateralAmount,
//			common.PortalTopUpWaitingPortingRejectedStatus,
//		)
//		statusContentBytes, _ := json.Marshal(topUpWaitingPortingReq)
//		err = statedb.StoreCustodianTopupWaitingPortingStatus(
//			portalStateDB,
//			actionData.TxReqID.String(),
//			statusContentBytes,
//		)
//		if err != nil {
//			Logger.log.Errorf("ERROR: an error occurred while storing waiting porting top up error %v", err)
//		}
//	} else if depositStatus == common.PortalTopUpWaitingPortingSuccessChainStatus {
//		custodianStateKey := statedb.GenerateCustodianStateObjectKey(actionData.IncogAddressStr)
//		custodian, ok := currentPortalState.CustodianPoolState[custodianStateKey.String()]
//		if !ok {
//			Logger.log.Errorf("Custodian not found")
//			return nil
//		}
//
//		waitingPortingRequestKey := statedb.GeneratePortalWaitingPortingRequestObjectKey(actionData.PortingID)
//		waitingPortingReq, ok := currentPortalState.WaitingPortingRequests[waitingPortingRequestKey.String()]
//		if !ok || waitingPortingReq == nil || waitingPortingReq.TokenID() != actionData.PTokenID {
//			Logger.log.Errorf("Waiting porting request with portingID (%s) not found", actionData.PortingID)
//			return nil
//		}
//
//		err = UpdateCustodianAfterTopupWaitingPorting(currentPortalState, waitingPortingReq, custodian, actionData.PTokenID, actionData.DepositedAmount, actionData.FreeCollateralAmount, common.PRVIDStr)
//		if err != nil {
//			Logger.log.Errorf("Update portal state error: %+v", err)
//			return nil
//		}
//
//		topUpWaitingPortingReq := metadata2.NewPortalTopUpWaitingPortingRequestStatus(
//			actionData.TxReqID,
//			actionData.PortingID,
//			actionData.IncogAddressStr,
//			actionData.PTokenID,
//			actionData.DepositedAmount,
//			actionData.FreeCollateralAmount,
//			common.PortalTopUpWaitingPortingSuccessStatus,
//		)
//		statusContentBytes, _ := json.Marshal(topUpWaitingPortingReq)
//		err = statedb.StoreCustodianTopupWaitingPortingStatus(
//			portalStateDB,
//			actionData.TxReqID.String(),
//			statusContentBytes,
//		)
//		if err != nil {
//			Logger.log.Errorf("ERROR: an error occurred while storing waiting porting top up error %v", err)
//			return nil
//		}
//
//		// update state of porting request by portingID
//		newPortingRequestState := metadata2.NewPortingRequestStatus(
//			waitingPortingReq.UniquePortingID(),
//			waitingPortingReq.TxReqID(),
//			waitingPortingReq.TokenID(),
//			waitingPortingReq.PorterAddress(),
//			waitingPortingReq.Amount(),
//			waitingPortingReq.Custodians(),
//			waitingPortingReq.PortingFee(),
//			common.PortalPortingReqWaitingStatus,
//			beaconHeight+1,
//			waitingPortingReq.ShardHeight(),
//			waitingPortingReq.ShardID(),
//		)
//		newPortingRequestStatusBytes, _ := json.Marshal(newPortingRequestState)
//		err = statedb.StorePortalPortingRequestStatus(
//			portalStateDB,
//			waitingPortingReq.UniquePortingID(),
//			newPortingRequestStatusBytes,
//		)
//		if err != nil {
//			Logger.log.Errorf("ERROR: an error occurred while store porting request item: %+v", err)
//			return nil
//		}
//	}
//	return nil
//}

//func (blockchain *BlockChain) processPortalCustodianTopup(
//	portalStateDB *statedb.StateDB,
//	beaconHeight uint64,
//	instructions []string,
//	currentPortalState *CurrentPortalState,
//	portalParams portal.PortalParams) error {
//	if currentPortalState == nil {
//		Logger.log.Errorf("current portal state is nil")
//		return nil
//	}
//	if len(instructions) != 4 {
//		return nil // skip the instruction
//	}
//
//	// unmarshal instructions content
//	var actionData metadata2.PortalLiquidationCustodianDepositContentV2
//	err := json.Unmarshal([]byte(instructions[3]), &actionData)
//	if err != nil {
//		Logger.log.Errorf("Error when unmarshaling portal liquidation custodian deposit content %v - %v", instructions[3], err)
//		return nil
//	}
//
//	depositStatus := instructions[2]
//
//	if depositStatus == common.PortalCustodianTopupSuccessChainStatus {
//		custodianStateKey := statedb.GenerateCustodianStateObjectKey(actionData.IncogAddressStr)
//		custodianStateKeyStr := custodianStateKey.String()
//		custodian, ok := currentPortalState.CustodianPoolState[custodianStateKeyStr]
//		if !ok {
//			Logger.log.Errorf("Custodian not found")
//			return nil
//		}
//
//		_, err = UpdateCustodianAfterTopup(currentPortalState, custodian, actionData.PTokenId, actionData.DepositedAmount, actionData.FreeCollateralAmount, common.PRVIDStr)
//		if err != nil {
//			Logger.log.Errorf("Update custodians state error : %+v", err)
//			return nil
//		}
//
//		newLiquidationCustodianDeposit := metadata2.NewLiquidationCustodianDepositStatusV2(
//			actionData.TxReqID,
//			actionData.IncogAddressStr,
//			actionData.PTokenId,
//			actionData.DepositedAmount,
//			actionData.FreeCollateralAmount,
//			common.PortalCustodianTopupSuccessStatus,
//		)
//
//		contentStatusBytes, _ := json.Marshal(newLiquidationCustodianDeposit)
//		err = statedb.StoreCustodianTopupStatus(
//			portalStateDB,
//			actionData.TxReqID.String(),
//			contentStatusBytes,
//		)
//		if err != nil {
//			Logger.log.Errorf("ERROR: an error occurred while store liquidation custodian deposit error %v", err)
//			return nil
//		}
//	} else if depositStatus == common.PortalCustodianTopupRejectedChainStatus {
//		newLiquidationCustodianDeposit := metadata2.NewLiquidationCustodianDepositStatusV2(
//			actionData.TxReqID,
//			actionData.IncogAddressStr,
//			actionData.PTokenId,
//			actionData.DepositedAmount,
//			actionData.FreeCollateralAmount,
//			common.PortalCustodianTopupRejectedStatus,
//		)
//
//		contentStatusBytes, _ := json.Marshal(newLiquidationCustodianDeposit)
//		err = statedb.StoreCustodianTopupStatus(
//			portalStateDB,
//			actionData.TxReqID.String(),
//			contentStatusBytes,
//		)
//		if err != nil {
//			Logger.log.Errorf("ERROR: an error occurred while store liquidation custodian deposit error %v", err)
//			return nil
//		}
//	}
//
//	return nil
//}

//func (blockchain *BlockChain) processPortalExpiredPortingRequest(
//	stateDB *statedb.StateDB,
//	beaconHeight uint64,
//	instructions []string,
//	currentPortalState *CurrentPortalState,
//	portalParams portal.PortalParams) error {
//	if currentPortalState == nil {
//		Logger.log.Errorf("current portal state is nil")
//		return nil
//	}
//	if len(instructions) != 4 {
//		return nil // skip the instruction
//	}
//
//	// unmarshal instructions content
//	var actionData metadata2.PortalExpiredWaitingPortingReqContent
//	err := json.Unmarshal([]byte(instructions[3]), &actionData)
//	if err != nil {
//		Logger.log.Errorf("Error when unmarshaling portal expired waiting porting content %v - %v", instructions[3], err)
//		return nil
//	}
//
//	status := instructions[2]
//	waitingPortingID := actionData.UniquePortingID
//
//	if status == common.PortalExpiredWaitingPortingReqSuccessChainStatus {
//		waitingPortingKey := statedb.GeneratePortalWaitingPortingRequestObjectKey(waitingPortingID)
//		waitingPortingKeyStr := waitingPortingKey.String()
//		waitingPortingReq := currentPortalState.WaitingPortingRequests[waitingPortingKeyStr]
//		if waitingPortingReq == nil {
//			Logger.log.Errorf("[processPortalExpiredPortingRequest] waiting porting req nil with key : %v", waitingPortingKey)
//			return nil
//		}
//
//		// get tokenID from redeemTokenID
//		tokenID := waitingPortingReq.TokenID()
//
//		// update custodian state in matching custodians list (holding public tokens, locked amount)
//		for _, matchCusDetail := range waitingPortingReq.Custodians() {
//			cusStateKey := statedb.GenerateCustodianStateObjectKey(matchCusDetail.IncAddress)
//			cusStateKeyStr := cusStateKey.String()
//			custodianState := currentPortalState.CustodianPoolState[cusStateKeyStr]
//			if custodianState == nil {
//				Logger.log.Errorf("[checkAndBuildInstForExpiredWaitingPortingRequest] Error when get custodian state with key %v\n: ", cusStateKey)
//				continue
//			}
//			err = updateCustodianStateAfterExpiredPortingReq(custodianState, matchCusDetail.LockedAmountCollateral, matchCusDetail.LockedTokenCollaterals, tokenID)
//			if err != nil {
//				Logger.log.Errorf("ERROR: an error occured while update state for expired porting request: %+v", err)
//				return nil
//			}
//		}
//
//		// remove waiting porting request from waiting list
//		delete(currentPortalState.WaitingPortingRequests, waitingPortingKeyStr)
//		statedb.DeleteWaitingPortingRequest(stateDB, waitingPortingReq.UniquePortingID())
//
//		// update status of porting ID  => expired/liquidated
//		portingReqStatus := common.PortalPortingReqExpiredStatus
//		if actionData.ExpiredByLiquidation {
//			portingReqStatus = common.PortalPortingReqLiquidatedStatus
//		}
//
//		newPortingRequestStatus := metadata2.NewPortingRequestStatus(
//			waitingPortingReq.UniquePortingID(),
//			waitingPortingReq.TxReqID(),
//			tokenID,
//			waitingPortingReq.PorterAddress(),
//			waitingPortingReq.Amount(),
//			waitingPortingReq.Custodians(),
//			waitingPortingReq.PortingFee(),
//			portingReqStatus,
//			waitingPortingReq.BeaconHeight(),
//			waitingPortingReq.ShardHeight(),
//			waitingPortingReq.ShardID())
//
//		newPortingRequestStatusBytes, _ := json.Marshal(newPortingRequestStatus)
//		err = statedb.StorePortalPortingRequestStatus(
//			stateDB,
//			actionData.UniquePortingID,
//			newPortingRequestStatusBytes,
//		)
//		if err != nil {
//			Logger.log.Errorf("ERROR: an error occurred while store porting request item status: %+v", err)
//			return nil
//		}
//
//		// track expired waiting porting request status by portingID into DB
//		expiredPortingTrackData := metadata2.PortalExpiredWaitingPortingReqStatus{
//			Status:               common.PortalExpiredPortingReqSuccessStatus,
//			UniquePortingID:      waitingPortingID,
//			ShardID:              actionData.ShardID,
//			ExpiredByLiquidation: actionData.ExpiredByLiquidation,
//			ExpiredBeaconHeight:  beaconHeight + 1,
//		}
//		expiredPortingTrackDataBytes, _ := json.Marshal(expiredPortingTrackData)
//		err = statedb.StorePortalExpiredPortingRequestStatus(
//			stateDB,
//			waitingPortingID,
//			expiredPortingTrackDataBytes,
//		)
//		if err != nil {
//			Logger.log.Errorf("ERROR: an error occured while tracking expired porting request: %+v", err)
//			return nil
//		}
//
//	} else if status == common.PortalCustodianTopupRejectedChainStatus {
//		// track expired waiting porting request status by portingID into DB
//		expiredPortingTrackData := metadata2.PortalExpiredWaitingPortingReqStatus{
//			Status:               common.PortalExpiredPortingReqFailedStatus,
//			UniquePortingID:      waitingPortingID,
//			ShardID:              actionData.ShardID,
//			ExpiredByLiquidation: actionData.ExpiredByLiquidation,
//			ExpiredBeaconHeight:  beaconHeight + 1,
//		}
//		expiredPortingTrackDataBytes, _ := json.Marshal(expiredPortingTrackData)
//		err = statedb.StorePortalExpiredPortingRequestStatus(
//			stateDB,
//			waitingPortingID,
//			expiredPortingTrackDataBytes,
//		)
//		if err != nil {
//			Logger.log.Errorf("ERROR: an error occured while tracking expired porting request: %+v", err)
//			return nil
//		}
//	}
//
//	return nil
//}

//func (blockchain *BlockChain) processLiquidationByExchangeRatesV3(
//	portalStateDB *statedb.StateDB,
//	beaconHeight uint64,
//	instructions []string,
//	currentPortalState *CurrentPortalState,
//	portalParams portal.PortalParams) error {
//
//	// unmarshal instructions content
//	var actionData metadata2.PortalLiquidationByRatesContentV3
//	err := json.Unmarshal([]byte(instructions[3]), &actionData)
//	if err != nil {
//		Logger.log.Errorf("Can not unmarshal instruction content %v - Error %v\n", instructions[3], err)
//		return nil
//	}
//
//	cusStateKeyStr := statedb.GenerateCustodianStateObjectKey(actionData.CustodianIncAddress).String()
//	custodianState, ok := currentPortalState.CustodianPoolState[cusStateKeyStr]
//	if !ok || custodianState == nil {
//		Logger.log.Errorf("Custodian not found")
//		return nil
//	}
//
//	reqStatus := instructions[2]
//	if reqStatus == common.PortalLiquidateTPExchangeRatesSuccessChainStatus {
//		liquidationInfo := actionData.Details
//
//		//update current portal state
//		updateCurrentPortalStateAfterLiquidationByRatesV3(currentPortalState, cusStateKeyStr, liquidationInfo, actionData.RemainUnlockCollaterals)
//
//		// store db
//		status := metadata2.PortalLiquidationByRatesStatusV3{
//			CustodianIncAddress: actionData.CustodianIncAddress,
//			Details:             actionData.Details,
//		}
//		statusBytes, _ := json.Marshal(status)
//		err = statedb.StoreLiquidationByExchangeRateStatusV3(
//			portalStateDB,
//			beaconHeight,
//			custodianState.GetIncognitoAddress(),
//			statusBytes)
//		if err != nil {
//			Logger.log.Errorf("ERROR: an error occurred while store liquidation by exchange rates v3 %v", err)
//			return nil
//		}
//	}
//	return nil
//}

//func (blockchain *BlockChain) processPortalRedeemFromLiquidationPoolV3(
//	portalStateDB *statedb.StateDB,
//	beaconHeight uint64,
//	instructions []string,
//	currentPortalState *CurrentPortalState,
//	portalParams portal.PortalParams,
//	updatingInfoByTokenID map[common.Hash]basemeta.UpdatingInfo) error {
//	if currentPortalState == nil {
//		Logger.log.Errorf("current portal state is nil")
//		return nil
//	}
//
//	if len(instructions) != 4 {
//		return nil // skip the instruction
//	}
//
//	// unmarshal instructions content
//	var actionData metadata2.PortalRedeemFromLiquidationPoolContentV3
//	err := json.Unmarshal([]byte(instructions[3]), &actionData)
//	if err != nil {
//		Logger.log.Errorf("Can not unmarshal instruction content %v - Error %v\n", instructions[3], err)
//		return nil
//	}
//
//	reqStatus := instructions[2]
//
//	status := byte(common.PortalRedeemFromLiquidationPoolRejectedStatus)
//	if reqStatus == common.PortalRedeemFromLiquidationPoolSuccessChainStatus {
//		liquidateExchangeRatesKey := statedb.GeneratePortalLiquidationPoolObjectKey()
//		liquidateExchangeRates := currentPortalState.LiquidationPool[liquidateExchangeRatesKey.String()]
//
//		UpdateLiquidationPoolAfterRedeemFrom(
//			currentPortalState, liquidateExchangeRates, actionData.TokenID, actionData.RedeemAmount,
//			actionData.MintedPRVCollateral, actionData.UnlockedTokenCollaterals)
//		status = byte(common.PortalRedeemFromLiquidationPoolSuccessStatus)
//
//		// update bridge/portal token info
//		incTokenID, err := common.Hash{}.NewHashFromStr(actionData.TokenID)
//		if err != nil {
//			Logger.log.Errorf("ERROR: Can not new hash from porting incTokenID: %+v", err)
//			return nil
//		}
//		updatingInfo, found := updatingInfoByTokenID[*incTokenID]
//		if found {
//			updatingInfo.deductAmt += actionData.RedeemAmount
//		} else {
//			updatingInfo = UpdatingInfo{
//				countUpAmt:      0,
//				deductAmt:       actionData.RedeemAmount,
//				tokenID:         *incTokenID,
//				externalTokenID: nil,
//				isCentralized:   false,
//			}
//		}
//		updatingInfoByTokenID[*incTokenID] = updatingInfo
//	}
//
//	// store db status
//	redeem := metadata2.PortalRedeemFromLiquidationPoolStatusV3{
//		TokenID:                  actionData.TokenID,
//		RedeemAmount:             actionData.RedeemAmount,
//		RedeemerIncAddressStr:    actionData.RedeemerIncAddressStr,
//		RedeemerExtAddressStr:    actionData.RedeemerExtAddressStr,
//		TxReqID:                  actionData.TxReqID,
//		MintedPRVCollateral:      actionData.MintedPRVCollateral,
//		UnlockedTokenCollaterals: actionData.UnlockedTokenCollaterals,
//		Status:                   status,
//	}
//
//	contentStatusBytes, _ := json.Marshal(redeem)
//	err = statedb.StoreRedeemRequestFromLiquidationPoolByTxIDStatusV3(
//		portalStateDB,
//		actionData.TxReqID.String(),
//		contentStatusBytes,
//	)
//
//	if err != nil {
//		Logger.log.Errorf("Store redeem liquidate exchange rates error %v\n", err)
//		return nil
//	}
//
//	return nil
//}

//func (blockchain *BlockChain) processPortalCustodianTopupV3(
//	portalStateDB *statedb.StateDB,
//	beaconHeight uint64,
//	instructions []string,
//	currentPortalState *CurrentPortalState,
//	portalParams portal.PortalParams) error {
//	if currentPortalState == nil {
//		Logger.log.Errorf("current portal state is nil")
//		return nil
//	}
//	if len(instructions) != 4 {
//		return nil // skip the instruction
//	}
//
//	// unmarshal instructions content
//	var actionData metadata2.PortalLiquidationCustodianDepositContentV3
//	err := json.Unmarshal([]byte(instructions[3]), &actionData)
//	if err != nil {
//		Logger.log.Errorf("Error when unmarshaling portal liquidation custodian deposit content %v - %v", instructions[3], err)
//		return nil
//	}
//
//	depositStatus := instructions[2]
//
//	if depositStatus == common.PortalCustodianTopupSuccessChainStatus {
//		custodianStateKey := statedb.GenerateCustodianStateObjectKey(actionData.IncogAddressStr)
//		custodianStateKeyStr := custodianStateKey.String()
//		custodian, ok := currentPortalState.CustodianPoolState[custodianStateKeyStr]
//		if !ok {
//			Logger.log.Errorf("Process custodian topop v3 error: Custodian not found")
//			return nil
//		}
//
//		_, err = UpdateCustodianAfterTopup(currentPortalState, custodian, actionData.PortalTokenID, actionData.DepositAmount, actionData.FreeTokenCollateralAmount, actionData.CollateralTokenID)
//		if !ok {
//			Logger.log.Errorf("Process custodian topop v3 error: %+v", err)
//			return nil
//		}
//
//		newLiquidationCustodianDeposit := metadata2.NewLiquidationCustodianDepositStatus3(
//			actionData.IncogAddressStr,
//			actionData.PortalTokenID,
//			actionData.CollateralTokenID,
//			actionData.DepositAmount,
//			actionData.FreeTokenCollateralAmount,
//			actionData.UniqExternalTxID,
//			actionData.TxReqID,
//			common.PortalCustodianTopupSuccessStatus,
//		)
//
//		contentStatusBytes, _ := json.Marshal(newLiquidationCustodianDeposit)
//		err = statedb.StoreCustodianTopupStatusV3(
//			portalStateDB,
//			actionData.TxReqID.String(),
//			contentStatusBytes,
//		)
//		if err != nil {
//			Logger.log.Errorf("ERROR: an error occurred while store liquidation custodian deposit v3 error %v", err)
//			return nil
//		}
//
//		// store uniq external tx
//		err := statedb.InsertPortalExternalTxHashSubmitted(portalStateDB, actionData.UniqExternalTxID)
//		if err != nil {
//			Logger.log.Errorf("ERROR: an error occured while tracking uniq external tx id: %+v", err)
//			return nil
//		}
//	} else if depositStatus == common.PortalCustodianTopupRejectedChainStatus {
//		newLiquidationCustodianDeposit := metadata2.NewLiquidationCustodianDepositStatus3(
//			actionData.IncogAddressStr,
//			actionData.PortalTokenID,
//			actionData.CollateralTokenID,
//			actionData.DepositAmount,
//			actionData.FreeTokenCollateralAmount,
//			actionData.UniqExternalTxID,
//			actionData.TxReqID,
//			common.PortalCustodianTopupRejectedStatus,
//		)
//
//		contentStatusBytes, _ := json.Marshal(newLiquidationCustodianDeposit)
//		err = statedb.StoreCustodianTopupStatusV3(
//			portalStateDB,
//			actionData.TxReqID.String(),
//			contentStatusBytes,
//		)
//		if err != nil {
//			Logger.log.Errorf("ERROR: an error occurred while store liquidation custodian deposit v3 error %v", err)
//			return nil
//		}
//	}
//
//	return nil
//}
//
//func (blockchain *BlockChain) processPortalTopUpWaitingPortingV3(
//	portalStateDB *statedb.StateDB,
//	beaconHeight uint64,
//	instructions []string,
//	currentPortalState *CurrentPortalState,
//	portalParams portal.PortalParams,
//) error {
//	if currentPortalState == nil {
//		Logger.log.Errorf("current portal state is nil")
//		return nil
//	}
//	if len(instructions) != 4 {
//		return nil // skip the instruction
//	}
//
//	var actionData metadata2.PortalTopUpWaitingPortingRequestContentV3
//	err := json.Unmarshal([]byte(instructions[3]), &actionData)
//	if err != nil {
//		Logger.log.Errorf("Error when unmarshaling portal top up waiting porting action %v - %v", instructions[3], err)
//		return nil
//	}
//
//	depositStatus := instructions[2]
//	if depositStatus == common.PortalTopUpWaitingPortingRejectedChainStatus {
//		topUpWaitingPortingReq := metadata2.NewPortalTopUpWaitingPortingRequestStatusV3(
//			actionData.IncogAddressStr,
//			actionData.PortalTokenID,
//			actionData.CollateralTokenID,
//			actionData.DepositAmount,
//			actionData.FreeTokenCollateralAmount,
//			actionData.PortingID,
//			actionData.UniqExternalTxID,
//			actionData.TxReqID,
//			common.PortalTopUpWaitingPortingRejectedStatus,
//		)
//		statusContentBytes, _ := json.Marshal(topUpWaitingPortingReq)
//		err = statedb.StoreCustodianTopupWaitingPortingStatusV3(
//			portalStateDB,
//			actionData.TxReqID.String(),
//			statusContentBytes,
//		)
//		if err != nil {
//			Logger.log.Errorf("ERROR: an error occurred while storing waiting porting top up error %v", err)
//		}
//	} else if depositStatus == common.PortalTopUpWaitingPortingSuccessChainStatus {
//		custodianStateKey := statedb.GenerateCustodianStateObjectKey(actionData.IncogAddressStr)
//		custodian, ok := currentPortalState.CustodianPoolState[custodianStateKey.String()]
//		if !ok {
//			Logger.log.Errorf("Custodian not found")
//			return nil
//		}
//
//		waitingPortingRequestKey := statedb.GeneratePortalWaitingPortingRequestObjectKey(actionData.PortingID)
//		waitingPortingReq, ok := currentPortalState.WaitingPortingRequests[waitingPortingRequestKey.String()]
//		if !ok || waitingPortingReq == nil || waitingPortingReq.TokenID() != actionData.PortalTokenID {
//			Logger.log.Errorf("Waiting porting request with portingID (%s) not found", actionData.PortingID)
//			return nil
//		}
//
//		err = UpdateCustodianAfterTopupWaitingPorting(currentPortalState, waitingPortingReq, custodian, actionData.PortalTokenID, actionData.DepositAmount, actionData.FreeTokenCollateralAmount, actionData.CollateralTokenID)
//		if err != nil {
//			Logger.log.Errorf("Update portal state error: %+v", err)
//			return nil
//		}
//
//		topUpWaitingPortingReq := metadata2.NewPortalTopUpWaitingPortingRequestStatusV3(
//			actionData.IncogAddressStr,
//			actionData.PortalTokenID,
//			actionData.CollateralTokenID,
//			actionData.DepositAmount,
//			actionData.FreeTokenCollateralAmount,
//			actionData.PortingID,
//			actionData.UniqExternalTxID,
//			actionData.TxReqID,
//			common.PortalTopUpWaitingPortingSuccessStatus,
//		)
//		statusContentBytes, _ := json.Marshal(topUpWaitingPortingReq)
//		err = statedb.StoreCustodianTopupWaitingPortingStatusV3(
//			portalStateDB,
//			actionData.TxReqID.String(),
//			statusContentBytes,
//		)
//		if err != nil {
//			Logger.log.Errorf("ERROR: an error occurred while storing waiting porting top up error %v", err)
//		}
//
//		// update state of porting request by portingID
//		newPortingRequestState := metadata2.NewPortingRequestStatus(
//			waitingPortingReq.UniquePortingID(),
//			waitingPortingReq.TxReqID(),
//			waitingPortingReq.TokenID(),
//			waitingPortingReq.PorterAddress(),
//			waitingPortingReq.Amount(),
//			waitingPortingReq.Custodians(),
//			waitingPortingReq.PortingFee(),
//			common.PortalPortingReqWaitingStatus,
//			beaconHeight+1,
//			waitingPortingReq.ShardHeight(),
//			waitingPortingReq.ShardID(),
//		)
//		newPortingRequestStatusBytes, _ := json.Marshal(newPortingRequestState)
//		err = statedb.StorePortalPortingRequestStatus(
//			portalStateDB,
//			waitingPortingReq.UniquePortingID(),
//			newPortingRequestStatusBytes,
//		)
//		if err != nil {
//			Logger.log.Errorf("ERROR: an error occurred while store porting request item: %+v", err)
//			return nil
//		}
//
//		// store uniq external tx
//		err := statedb.InsertPortalExternalTxHashSubmitted(portalStateDB, actionData.UniqExternalTxID)
//		if err != nil {
//			Logger.log.Errorf("ERROR: an error occured while tracking uniq external tx id: %+v", err)
//			return nil
//		}
//	}
//	return nil
//}
