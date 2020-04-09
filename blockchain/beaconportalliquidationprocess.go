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
	beaconHeight uint64, instructions []string,
	currentPortalState *CurrentPortalState) error {

	// unmarshal instructions content
	var actionData metadata.PortalLiquidateCustodianContent
	err := json.Unmarshal([]byte(instructions[3]), &actionData)
	if err != nil {
		Logger.log.Errorf("Can not unmarshal instruction content %v\n", err)
		return nil
	}

	// get pTokenID from actionData
	pTokenID := actionData.TokenID

	reqStatus := instructions[2]
	if reqStatus == common.PortalLiquidateCustodianSuccessChainStatus {
		// update custodian state (total collateral, holding public tokens, locked amount, free collateral)
		cusStateKey := statedb.GenerateCustodianStateObjectKey(beaconHeight, actionData.CustodianIncAddressStr)
		cusStateKeyStr := cusStateKey.String()
		custodianState := currentPortalState.CustodianPoolState[cusStateKeyStr]

		if custodianState.GetTotalCollateral() < actionData.MintedCollateralAmount ||
			custodianState.GetLockedAmountCollateral()[pTokenID] < actionData.MintedCollateralAmount {
			Logger.log.Errorf("[checkAndBuildInstForCustodianLiquidation] Total collateral %v, locked amount %v "+
				"should be greater than minted amount %v\n: ",
				custodianState.GetTotalCollateral(), custodianState.GetLockedAmountCollateral()[pTokenID], actionData.MintedCollateralAmount)
			return fmt.Errorf("[checkAndBuildInstForCustodianLiquidation] Total collateral %v, locked amount %v "+
				"should be greater than minted amount %v\n: ",
				custodianState.GetTotalCollateral(), custodianState.GetLockedAmountCollateral()[pTokenID], actionData.MintedCollateralAmount)
		}

		err = updateCustodianStateAfterLiquidateCustodian(custodianState, actionData.MintedCollateralAmount, pTokenID)
		if err != nil {
			Logger.log.Errorf("[checkAndBuildInstForCustodianLiquidation] Error when updating %v\n: ", err)
			return err
		}

		// remove matching custodian from matching custodians list in waiting redeem request
		waitingRedeemReqKey := statedb.GenerateWaitingRedeemRequestObjectKey(beaconHeight, actionData.UniqueRedeemID)
		waitingRedeemReqKeyStr := waitingRedeemReqKey.String()

		updatedCustodians, _ := removeCustodianFromMatchingRedeemCustodians(
			currentPortalState.WaitingRedeemRequests[waitingRedeemReqKeyStr].GetCustodians(), actionData.CustodianIncAddressStr)
		currentPortalState.WaitingRedeemRequests[waitingRedeemReqKeyStr].SetCustodians(updatedCustodians)

		// remove redeem request from waiting redeem requests list
		if len(currentPortalState.WaitingRedeemRequests[waitingRedeemReqKeyStr].GetCustodians()) == 0 {
			deleteWaitingRedeemRequest(currentPortalState, waitingRedeemReqKeyStr)

			// update status of redeem request with redeemID to liquidated status
			err = updateRedeemRequestStatusByRedeemId(actionData.UniqueRedeemID, common.PortalRedeemReqLiquidatedStatus, stateDB)
			if err != nil {
				Logger.log.Errorf("ERROR: an error occurred while updating redeem request status by redeemID: %+v", err)
				return nil
			}
		}

		// track liquidation custodian status by redeemID and custodian address into DB
		custodianLiquidationTrackData := metadata.PortalLiquidateCustodianStatus{
			Status:                   common.PortalLiquidateCustodianSuccessStatus,
			UniqueRedeemID:           actionData.UniqueRedeemID,
			TokenID:                  actionData.TokenID,
			RedeemPubTokenAmount:     actionData.RedeemPubTokenAmount,
			MintedCollateralAmount:   actionData.MintedCollateralAmount,
			RedeemerIncAddressStr:    actionData.RedeemerIncAddressStr,
			CustodianIncAddressStr:   actionData.CustodianIncAddressStr,
			LiquidatedByExchangeRate: actionData.LiquidatedByExchangeRate,
			ShardID:                  actionData.ShardID,
			LiquidatedBeaconHeight:   beaconHeight + 1,
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
			Status:                   common.PortalLiquidateCustodianFailedStatus,
			UniqueRedeemID:           actionData.UniqueRedeemID,
			TokenID:                  actionData.TokenID,
			RedeemPubTokenAmount:     actionData.RedeemPubTokenAmount,
			MintedCollateralAmount:   actionData.MintedCollateralAmount,
			RedeemerIncAddressStr:    actionData.RedeemerIncAddressStr,
			CustodianIncAddressStr:   actionData.CustodianIncAddressStr,
			LiquidatedByExchangeRate: actionData.LiquidatedByExchangeRate,
			ShardID:                  actionData.ShardID,
			LiquidatedBeaconHeight:   beaconHeight + 1,
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

func (blockchain *BlockChain) processLiquidationTopPercentileExchangeRates(portalStateDB *statedb.StateDB, beaconHeight uint64, instructions []string,
	currentPortalState *CurrentPortalState) error {

	// unmarshal instructions content
	var actionData metadata.PortalLiquidateTopPercentileExchangeRatesContent
	err := json.Unmarshal([]byte(instructions[3]), &actionData)
	if err != nil {
		Logger.log.Errorf("Can not unmarshal instruction content %v\n", err)
		return nil
	}

	keyExchangeRate := statedb.GeneratePortalFinalExchangeRatesStateObjectKey(beaconHeight)
	exchangeRate, ok := currentPortalState.FinalExchangeRatesState[keyExchangeRate.String()]
	if !ok {
		Logger.log.Errorf("Exchange rate not found", err)
		return nil
	}

	cusStateKey := statedb.GenerateCustodianStateObjectKey(beaconHeight, actionData.CustodianAddress)
	cusStateKeyStr := cusStateKey.String()
	custodianState, ok := currentPortalState.CustodianPoolState[cusStateKeyStr]
	//todo: check custodian exist on db
	if !ok {
		Logger.log.Errorf("Custodian not found")
		return nil
	}

	reqStatus := instructions[2]
	if reqStatus == common.PortalLiquidateTPExchangeRatesSuccessChainStatus {
		//validation
		detectTPExchangeRates, err := calculateTPRatio(custodianState.GetHoldingPublicTokens(), custodianState.GetLockedAmountCollateral(), exchangeRate)
		if err != nil {
			Logger.log.Errorf("Detect tp ratio error %v", err)
			return nil
		}

		detectTp, err := detectTopPercentileLiquidation(custodianState, detectTPExchangeRates)

		if err != nil {
			Logger.log.Errorf("ERROR: an error occurred while Get liquidate exchange rates change error %v", err)
			return nil
		}

		if len(detectTp) > 0 {
			//update current portal state
			updateCurrentPortalStateOfLiquidationExchangeRates(beaconHeight, currentPortalState, cusStateKeyStr, custodianState, detectTp)

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
		)

		if err != nil {
			Logger.log.Errorf("ERROR: an error occurred while store liquidation TP exchange rates %v", err)
			return nil
		}
	}

	return nil
}

func (blockchain *BlockChain) processPortalRedeemLiquidateExchangeRates(portalStateDB *statedb.StateDB, beaconHeight uint64, instructions []string, currentPortalState *CurrentPortalState, updatingInfoByTokenID map[common.Hash]UpdatingInfo) error {
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
		Logger.log.Errorf("Can not unmarshal instruction content %v\n", err)
		return nil
	}

	reqStatus := instructions[2]
	if reqStatus == common.PortalRedeemLiquidateExchangeRatesSuccessChainStatus {
		keyExchangeRate := statedb.GeneratePortalFinalExchangeRatesStateObjectKey(beaconHeight)
		_, ok := currentPortalState.FinalExchangeRatesState[keyExchangeRate.String()]
		if !ok {
			Logger.log.Errorf("Exchange rate not found", err)
			return nil
		}

		liquidateExchangeRatesKey := statedb.GeneratePortalLiquidateExchangeRatesPoolObjectKey(beaconHeight)
		liquidateExchangeRates, ok := currentPortalState.LiquidateExchangeRatesPool[liquidateExchangeRatesKey.String()]

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

		liquidateExchangeRates.Rates()[actionData.TokenID] = statedb.LiquidateExchangeRatesDetail{
			HoldAmountFreeCollateral: liquidateByTokenID.HoldAmountFreeCollateral - totalPrv,
			HoldAmountPubToken:       liquidateByTokenID.HoldAmountPubToken - actionData.RedeemAmount,
		}

		currentPortalState.LiquidateExchangeRatesPool[liquidateExchangeRatesKey.String()] = liquidateExchangeRates

		Logger.log.Infof("Redeem Liquidation: Amount refund to user amount ptoken %v, amount prv %v", actionData.RedeemAmount, totalPrv)

		redeem := metadata.NewRedeemLiquidateExchangeRatesStatus(
			actionData.TxReqID,
			actionData.TokenID,
			actionData.RedeemerIncAddressStr,
			actionData.RemoteAddress,
			actionData.RedeemAmount,
			actionData.RedeemFee,
			common.PortalRedeemLiquidateExchangeRatesSuccessStatus,
			totalPrv,
		)

		contentStatusBytes, _ := json.Marshal(redeem)
		err = statedb.TrackPortalStateStatusMultiple(
			portalStateDB,
			statedb.PortalLiquidationRedeemRequestStatusPrefix(),
			[]byte(actionData.TxReqID.String()),
			contentStatusBytes,
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
			actionData.RemoteAddress,
			actionData.RedeemAmount,
			actionData.RedeemFee,
			common.PortalRedeemLiquidateExchangeRatesRejectedStatus,
			0,
		)

		contentStatusBytes, _ := json.Marshal(redeem)
		err = statedb.TrackPortalStateStatusMultiple(
			portalStateDB,
			statedb.PortalLiquidationRedeemRequestStatusPrefix(),
			[]byte(actionData.TxReqID.String()),
			contentStatusBytes,
		)
		if err != nil {
			Logger.log.Errorf("Store redeem liquidate exchange rates error %v\n", err)
			return nil
		}
	}

	return nil
}

func (blockchain *BlockChain) processPortalLiquidationCustodianDeposit(portalStateDB *statedb.StateDB, beaconHeight uint64, instructions []string, currentPortalState *CurrentPortalState) error {
	if currentPortalState == nil {
		Logger.log.Errorf("current portal state is nil")
		return nil
	}
	if len(instructions) != 4 {
		return nil // skip the instruction
	}

	// unmarshal instructions content
	var actionData metadata.PortalLiquidationCustodianDepositContent
	err := json.Unmarshal([]byte(instructions[3]), &actionData)
	if err != nil {
		return err
	}

	depositStatus := instructions[2]

	if depositStatus == common.PortalLiquidationCustodianDepositSuccessChainStatus {
		keyExchangeRate := statedb.GeneratePortalFinalExchangeRatesStateObjectKey(beaconHeight)
		exchangeRate := currentPortalState.FinalExchangeRatesState[keyExchangeRate.String()]

		keyCustodianState := statedb.GenerateCustodianStateObjectKey(beaconHeight, actionData.IncogAddressStr)
		keyCustodianStateStr := keyCustodianState.String()

		custodian, ok := currentPortalState.CustodianPoolState[keyCustodianStateStr]
		if !ok {
			Logger.log.Errorf("Custodian not found")
			return nil
		}

		amountNeeded, totalFreeCollateralNeeded, remainFreeCollateral, err := CalAmountNeededDepositLiquidate(custodian, exchangeRate, actionData.PTokenId, actionData.FreeCollateralSelected)

		if err != nil {
			Logger.log.Errorf("Calculate amount needed deposit err %v", err)
			return nil
		}

		if actionData.DepositedAmount < amountNeeded {
			Logger.log.Errorf("Deposited amount is not enough, expect %v, data sent %v", amountNeeded, actionData.DepositedAmount)
			return nil
		}

		Logger.log.Infof("Deposited amount: expect %v, data sent %v", amountNeeded, actionData.DepositedAmount)

		remainDepositAmount := actionData.DepositedAmount - amountNeeded
		custodian.SetTotalCollateral(custodian.GetTotalCollateral() + actionData.DepositedAmount)

		if actionData.FreeCollateralSelected == false {
			lockedAmountTmp := custodian.GetLockedAmountCollateral()
			lockedAmountTmp[actionData.PTokenId] += amountNeeded
			custodian.SetLockedAmountCollateral(lockedAmountTmp)

			//update remain
			custodian.SetFreeCollateral(custodian.GetFreeCollateral() + remainDepositAmount)
		} else {
			//deposit from free collateral DepositedAmount
			lockedAmountTmp := custodian.GetLockedAmountCollateral()
			lockedAmountTmp[actionData.PTokenId] = lockedAmountTmp[actionData.PTokenId] +  amountNeeded + totalFreeCollateralNeeded
			custodian.SetLockedAmountCollateral(lockedAmountTmp)

			custodian.SetFreeCollateral(remainFreeCollateral + remainDepositAmount)
		}

		currentPortalState.CustodianPoolState[keyCustodianStateStr] = custodian

		newLiquidationCustodianDeposit := metadata.NewLiquidationCustodianDepositStatus(
			actionData.TxReqID,
			actionData.IncogAddressStr,
			actionData.PTokenId,
			actionData.DepositedAmount,
			actionData.FreeCollateralSelected,
			common.PortalLiquidationCustodianDepositSuccessStatus,
		)

		contentStatusBytes, _ := json.Marshal(newLiquidationCustodianDeposit)
		err = statedb.TrackPortalStateStatusMultiple(
			portalStateDB,
			statedb.PortalLiquidationCustodianDepositStatusPrefix(),
			[]byte(actionData.TxReqID.String()),
			contentStatusBytes,
		)

		if err != nil {
			Logger.log.Errorf("ERROR: an error occurred while store liquidation custodian deposit error %v", err)
			return nil
		}
	} else if depositStatus == common.PortalLiquidationCustodianDepositRejectedChainStatus {
		newLiquidationCustodianDeposit := metadata.NewLiquidationCustodianDepositStatus(
			actionData.TxReqID,
			actionData.IncogAddressStr,
			actionData.PTokenId,
			actionData.DepositedAmount,
			actionData.FreeCollateralSelected,
			common.PortalLiquidationCustodianDepositRejectedStatus,
		)

		contentStatusBytes, _ := json.Marshal(newLiquidationCustodianDeposit)
		err = statedb.TrackPortalStateStatusMultiple(
			portalStateDB,
			statedb.PortalLiquidationCustodianDepositStatusPrefix(),
			[]byte(actionData.TxReqID.String()),
			contentStatusBytes,
		)

		if err != nil {
			Logger.log.Errorf("ERROR: an error occurred while store liquidation custodian deposit error %v", err)
			return nil
		}
	}

	return nil
}

func (blockchain *BlockChain) processPortalExpiredPortingRequest(
	stateDB *statedb.StateDB, beaconHeight uint64, instructions []string, currentPortalState *CurrentPortalState) error {
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
		return err
	}

	status := instructions[2]
	waitingPortingID := actionData.UniquePortingID

	if status == common.PortalExpiredWaitingPortingReqSuccessChainStatus {
		waitingPortingKey := statedb.GeneratePortalWaitingPortingRequestObjectKey(beaconHeight, waitingPortingID)
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
			cusStateKey := statedb.GenerateCustodianStateObjectKey(beaconHeight, matchCusDetail.IncAddress)
			cusStateKeyStr := cusStateKey.String()
			custodianState := currentPortalState.CustodianPoolState[cusStateKeyStr]
			if custodianState == nil {
				Logger.log.Errorf("[checkAndBuildInstForExpiredWaitingPortingRequest] Error when get custodian state with key %v\n: ", cusStateKey)
				continue
			}
			_ = updateCustodianStateAfterExpiredPortingReq(
				custodianState, matchCusDetail.LockedAmountCollateral, matchCusDetail.Amount, tokenID)
		}

		// remove waiting porting request from waiting list
		// TODO:
		delete(currentPortalState.WaitingPortingRequests, waitingPortingKeyStr)

		// update status of porting ID  => expired/liquidated
		portingReqStatus := common.PortalPortingReqExpiredStatus
		if actionData.ExpiredByLiquidation {
			portingReqStatus = common.PortalPortingReqLiquidatedStatus
		}

		newPortingRequestStatus := statedb.NewWaitingPortingRequestWithValue(
			waitingPortingReq.UniquePortingID(),
			waitingPortingReq.TxReqID(),
			tokenID,
			waitingPortingReq.PorterAddress(),
			waitingPortingReq.Amount(),
			waitingPortingReq.Custodians(),
			waitingPortingReq.PortingFee(),
			portingReqStatus,
			waitingPortingReq.BeaconHeight(),
		)

		err = statedb.StoreWaitingPortingRequests(stateDB, beaconHeight, waitingPortingReq.UniquePortingID(), newPortingRequestStatus)
		if err != nil {
			Logger.log.Errorf("ERROR: an error occurred while store porting request item: %+v", err)
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
