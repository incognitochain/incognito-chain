package blockchain

import (
	"encoding/json"
	"fmt"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/database/lvdb"
	"github.com/incognitochain/incognito-chain/metadata"
)

func (blockchain *BlockChain) processPortalLiquidateCustodian(
	beaconHeight uint64, instructions []string,
	currentPortalState *CurrentPortalState) error {

	db := blockchain.GetDatabase()

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
		cusStateKey := lvdb.NewCustodianStateKey(beaconHeight, actionData.CustodianIncAddressStr)
		custodianState := currentPortalState.CustodianPoolState[cusStateKey]

		if custodianState.TotalCollateral < actionData.MintedCollateralAmount ||
			custodianState.LockedAmountCollateral[pTokenID] < actionData.MintedCollateralAmount {
			Logger.log.Errorf("[checkAndBuildInstForCustodianLiquidation] Total collateral %v, locked amount %v "+
				"should be greater than minted amount %v\n: ",
				custodianState.TotalCollateral, custodianState.LockedAmountCollateral[pTokenID], actionData.MintedCollateralAmount)
			return fmt.Errorf("[checkAndBuildInstForCustodianLiquidation] Total collateral %v, locked amount %v "+
				"should be greater than minted amount %v\n: ",
				custodianState.TotalCollateral, custodianState.LockedAmountCollateral[pTokenID], actionData.MintedCollateralAmount)
		}

		err = updateCustodianStateAfterLiquidateCustodian(custodianState, actionData.MintedCollateralAmount, pTokenID)
		if err != nil {
			Logger.log.Errorf("[checkAndBuildInstForCustodianLiquidation] Error when updating %v\n: ", err)
			return err
		}

		// remove matching custodian from matching custodians list in waiting redeem request
		waitingRedeemReqKey := lvdb.NewWaitingRedeemReqKey(beaconHeight, actionData.UniqueRedeemID)
		currentPortalState.WaitingRedeemRequests[waitingRedeemReqKey].Custodians, _ = removeCustodianFromMatchingRedeemCustodians(
			currentPortalState.WaitingRedeemRequests[waitingRedeemReqKey].Custodians, actionData.CustodianIncAddressStr)

		// remove redeem request from waiting redeem requests list
		if len(currentPortalState.WaitingRedeemRequests[waitingRedeemReqKey].Custodians) == 0 {
			delete(currentPortalState.WaitingRedeemRequests, waitingRedeemReqKey)

			// update status of redeem request with redeemID to liquidated status
			err = updateRedeemRequestStatusByRedeemId(actionData.UniqueRedeemID, common.PortalRedeemReqLiquidatedStatus, db)
			if err != nil {
				Logger.log.Errorf("ERROR: an error occurred while updating redeem request status by redeemID: %+v", err)
				return nil
			}
		}

		// track liquidation custodian status by redeemID and custodian address into DB
		custodianLiquidationTrackKey := lvdb.NewPortalLiquidationCustodianKey(actionData.UniqueRedeemID, actionData.CustodianIncAddressStr)
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
		err = db.TrackLiquidateCustodian(
			[]byte(custodianLiquidationTrackKey),
			custodianLiquidationTrackDataBytes,
		)
		if err != nil {
			Logger.log.Errorf("ERROR: an error occured while tracking liquidation custodian: %+v", err)
			return nil
		}

	} else if reqStatus == common.PortalLiquidateCustodianFailedChainStatus {
		// track liquidation custodian status by redeemID and custodian address into DB
		custodianLiquidationTrackKey := lvdb.NewPortalLiquidationCustodianKey(actionData.UniqueRedeemID, actionData.CustodianIncAddressStr)
		custodianLiquidationTrackData := metadata.PortalLiquidateCustodianStatus{
			Status:                   common.PortalLiquidateCustodianFailedStatus,
			UniqueRedeemID:           actionData.UniqueRedeemID,
			CustodianIncAddressStr:   actionData.CustodianIncAddressStr,
			LiquidatedByExchangeRate: actionData.LiquidatedByExchangeRate,
			LiquidatedBeaconHeight:   beaconHeight + 1,
		}
		custodianLiquidationTrackDataBytes, _ := json.Marshal(custodianLiquidationTrackData)
		err = db.TrackLiquidateCustodian(
			[]byte(custodianLiquidationTrackKey),
			custodianLiquidationTrackDataBytes,
		)
		if err != nil {
			Logger.log.Errorf("ERROR: an error occured while tracking liquidation custodian: %+v", err)
			return nil
		}
	}

	return nil
}

func (blockchain *BlockChain) processLiquidationTopPercentileExchangeRates(beaconHeight uint64, instructions []string,
	currentPortalState *CurrentPortalState) error {
	db := blockchain.GetDatabase()

	// unmarshal instructions content
	var actionData metadata.PortalLiquidateTopPercentileExchangeRatesContent
	err := json.Unmarshal([]byte(instructions[3]), &actionData)
	if err != nil {
		Logger.log.Errorf("Can not unmarshal instruction content %v\n", err)
		return nil
	}

	keyExchangeRate := lvdb.NewFinalExchangeRatesKey(beaconHeight)
	exchangeRate, ok := currentPortalState.FinalExchangeRates[keyExchangeRate]
	if !ok {
		Logger.log.Errorf("Exchange rate not found", err)
		return nil
	}

	cusStateKey := lvdb.NewCustodianStateKey(beaconHeight, actionData.CustodianAddress)
	custodianState, ok := currentPortalState.CustodianPoolState[cusStateKey]
	//todo: check custodian exist on db
	if !ok {
		Logger.log.Errorf("Custodian not found")
		return nil
	}

	reqStatus := instructions[2]

	Logger.log.Infof("status %v", reqStatus)

	if reqStatus == common.PortalLiquidateTPExchangeRatesSuccessChainStatus {
		//validation
		detectTPExchangeRates, err := calculateTPRatio(custodianState.HoldingPubTokens, custodianState.LockedAmountCollateral, exchangeRate)
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
			for ptoken, liquidateTopPercentileExchangeRatesDetail := range detectTp {
				custodianState.LockedAmountCollateral[ptoken] = custodianState.LockedAmountCollateral[ptoken] - liquidateTopPercentileExchangeRatesDetail.HoldAmountFreeCollateral
				custodianState.HoldingPubTokens[ptoken] = custodianState.HoldingPubTokens[ptoken] - liquidateTopPercentileExchangeRatesDetail.HoldAmountPubToken
				custodianState.TotalCollateral = custodianState.TotalCollateral - liquidateTopPercentileExchangeRatesDetail.HoldAmountFreeCollateral
			}

			//update custodian
			Logger.log.Infof("update custodian key %v", cusStateKey)
			currentPortalState.CustodianPoolState[cusStateKey] = custodianState

			//update LiquidateExchangeRates
			liquidateExchangeRatesKey := lvdb.NewPortalLiquidateExchangeRatesKey(beaconHeight)
			liquidateExchangeRates, ok := currentPortalState.LiquidateExchangeRates[liquidateExchangeRatesKey]

			Logger.log.Infof("update liquidateExchangeRatesKey key %v", liquidateExchangeRatesKey)
			if !ok {
				item := make(map[string]lvdb.LiquidateExchangeRatesDetail)

				for ptoken, liquidateTopPercentileExchangeRatesDetail := range detectTp {
					item[ptoken] = lvdb.LiquidateExchangeRatesDetail{
						HoldAmountFreeCollateral: liquidateTopPercentileExchangeRatesDetail.HoldAmountFreeCollateral,
						HoldAmountPubToken:       liquidateTopPercentileExchangeRatesDetail.HoldAmountPubToken,
					}
				}
				currentPortalState.LiquidateExchangeRates[liquidateExchangeRatesKey], _ = NewLiquidateExchangeRates(item)
			} else {
				for ptoken, liquidateTopPercentileExchangeRatesDetail := range detectTp {
					if _, ok := liquidateExchangeRates.Rates[ptoken]; !ok {
						liquidateExchangeRates.Rates[ptoken] = lvdb.LiquidateExchangeRatesDetail{
							HoldAmountFreeCollateral: liquidateTopPercentileExchangeRatesDetail.HoldAmountFreeCollateral,
							HoldAmountPubToken:       liquidateTopPercentileExchangeRatesDetail.HoldAmountPubToken,
						}
					} else {
						liquidateExchangeRates.Rates[ptoken] = lvdb.LiquidateExchangeRatesDetail{
							HoldAmountFreeCollateral: liquidateExchangeRates.Rates[ptoken].HoldAmountFreeCollateral + liquidateTopPercentileExchangeRatesDetail.HoldAmountFreeCollateral,
							HoldAmountPubToken:       liquidateExchangeRates.Rates[ptoken].HoldAmountPubToken + liquidateTopPercentileExchangeRatesDetail.HoldAmountPubToken,
						}
					}
				}

				currentPortalState.LiquidateExchangeRates[liquidateExchangeRatesKey] = liquidateExchangeRates
			}

			newTPKey := lvdb.NewPortalLiquidateTPExchangeRatesKey(beaconHeight, custodianState.IncognitoAddress)
			newTPExchangeRates, _ := NewLiquidateTopPercentileExchangeRates(
				custodianState.IncognitoAddress,
				detectTp,
				common.PortalLiquidationTPExchangeRatesSuccessStatus,
			)

			Logger.log.Infof("update liquidateTPExchangeRatesKey key %v", newTPKey)
			err := db.StoreLiquidateTopPercentileExchangeRates([]byte(newTPKey), newTPExchangeRates)
			if err != nil {
				Logger.log.Errorf("ERROR: an error occurred while store liquidation TP exchange rates %v", err)
				return nil
			}
		}
	} else if reqStatus == common.PortalLiquidateTPExchangeRatesFailedChainStatus {
		newTPKey := lvdb.NewPortalLiquidateTPExchangeRatesKey(beaconHeight, custodianState.IncognitoAddress)
		newTPExchangeRates, _ := NewLiquidateTopPercentileExchangeRates(
			custodianState.IncognitoAddress,
			nil,
			common.PortalLiquidationTPExchangeRatesFailedStatus,
		)

		err := db.StoreLiquidateTopPercentileExchangeRates([]byte(newTPKey), newTPExchangeRates)
		if err != nil {
			Logger.log.Errorf("ERROR: an error occurred while store liquidation TP exchange rates %v", err)
			return nil
		}
	}

	return nil
}

func (blockchain *BlockChain) processPortalRedeemLiquidateExchangeRates(beaconHeight uint64, instructions []string, currentPortalState *CurrentPortalState, updatingInfoByTokenID map[common.Hash]UpdatingInfo) error {
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

	db := blockchain.GetDatabase()

	reqStatus := instructions[2]
	if reqStatus == common.PortalRedeemLiquidateExchangeRatesSuccessChainStatus {
		keyExchangeRate := lvdb.NewFinalExchangeRatesKey(beaconHeight)
		_, ok := currentPortalState.FinalExchangeRates[keyExchangeRate]
		if !ok {
			Logger.log.Errorf("Exchange rate not found", err)
			return nil
		}

		liquidateExchangeRatesKey := lvdb.NewPortalLiquidateExchangeRatesKey(beaconHeight)
		liquidateExchangeRates, ok := currentPortalState.LiquidateExchangeRates[liquidateExchangeRatesKey]

		if !ok {
			Logger.log.Errorf("Liquidate exchange rates not found")
			return nil
		}

		liquidateByTokenID, ok := liquidateExchangeRates.Rates[actionData.TokenID]
		if !ok {
			Logger.log.Errorf("Liquidate exchange rates not found")
			return nil
		}

		totalPrv := actionData.TotalPTokenReceived

		liquidateExchangeRates.Rates[actionData.TokenID] = lvdb.LiquidateExchangeRatesDetail{
			HoldAmountFreeCollateral: liquidateByTokenID.HoldAmountFreeCollateral - totalPrv,
			HoldAmountPubToken:       liquidateByTokenID.HoldAmountPubToken - actionData.RedeemAmount,
		}

		currentPortalState.LiquidateExchangeRates[liquidateExchangeRatesKey] = liquidateExchangeRates

		redeemKey := lvdb.NewRedeemLiquidateExchangeRatesKey(actionData.TxReqID.String())
		redeem, _ := NewRedeemLiquidateExchangeRates(
			actionData.TxReqID,
			actionData.TokenID,
			actionData.RedeemerIncAddressStr,
			actionData.RemoteAddress,
			actionData.RedeemAmount,
			actionData.RedeemFee,
			totalPrv,
			common.PortalRedeemLiquidateExchangeRatesSuccessStatus,
		)

		err = db.StoreRedeemLiquidationExchangeRates([]byte(redeemKey), redeem)
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
		redeemKey := lvdb.NewRedeemLiquidateExchangeRatesKey(actionData.TxReqID.String())
		redeem, _ := NewRedeemLiquidateExchangeRates(
			actionData.TxReqID,
			actionData.TokenID,
			actionData.RedeemerIncAddressStr,
			actionData.RemoteAddress,
			actionData.RedeemAmount,
			actionData.RedeemFee,
			actionData.TotalPTokenReceived,
			common.PortalRedeemLiquidateExchangeRatesRejectedStatus,
		)

		err = db.StoreRedeemLiquidationExchangeRates([]byte(redeemKey), redeem)
		if err != nil {
			Logger.log.Errorf("Store redeem liquidate exchange rates error %v\n", err)
			return nil
		}
	}

	return nil
}

func (blockchain *BlockChain) processPortalLiquidationCustodianDeposit(beaconHeight uint64, instructions []string, currentPortalState *CurrentPortalState) error {
	if currentPortalState == nil {
		Logger.log.Errorf("current portal state is nil")
		return nil
	}
	if len(instructions) != 4 {
		return nil // skip the instruction
	}

	db := blockchain.GetDatabase()

	// unmarshal instructions content
	var actionData metadata.PortalLiquidationCustodianDepositContent
	err := json.Unmarshal([]byte(instructions[3]), &actionData)
	if err != nil {
		return err
	}

	depositStatus := instructions[2]

	if depositStatus == common.PortalLiquidationCustodianDepositSuccessChainStatus {
		keyExchangeRate := lvdb.NewFinalExchangeRatesKey(beaconHeight)
		exchangeRate := currentPortalState.FinalExchangeRates[keyExchangeRate]

		keyCustodianState := lvdb.NewCustodianStateKey(beaconHeight, actionData.IncogAddressStr)

		custodian, ok := currentPortalState.CustodianPoolState[keyCustodianState]
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
		custodian.TotalCollateral = custodian.TotalCollateral + actionData.DepositedAmount

		if actionData.FreeCollateralSelected == false {
			custodian.LockedAmountCollateral[actionData.PTokenId] = custodian.LockedAmountCollateral[actionData.PTokenId] + amountNeeded

			//update remain
			custodian.FreeCollateral = custodian.FreeCollateral + remainDepositAmount
		} else {
			//deposit from free collateral DepositedAmount
			custodian.LockedAmountCollateral[actionData.PTokenId] = custodian.LockedAmountCollateral[actionData.PTokenId] + amountNeeded + totalFreeCollateralNeeded

			custodian.FreeCollateral = remainFreeCollateral + remainDepositAmount
		}

		currentPortalState.CustodianPoolState[keyCustodianState] = custodian

		liquidationCustodianDepositKey := lvdb.NewLiquidationCustodianDepositKey(actionData.TxReqID.String())
		newLiquidationCustodianDeposit, _ := NewLiquidationCustodianDeposit(
			actionData.TxReqID,
			actionData.PTokenId,
			actionData.IncogAddressStr,
			actionData.DepositedAmount,
			actionData.FreeCollateralSelected,
			common.PortalLiquidationCustodianDepositSuccessStatus,
		)

		err = db.StoredLiquidationCustodianDeposit([]byte(liquidationCustodianDepositKey), newLiquidationCustodianDeposit)

		if err != nil {
			Logger.log.Errorf("ERROR: an error occurred while store liquidation custodian deposit error %v", err)
			return nil
		}
	} else if depositStatus == common.PortalLiquidationCustodianDepositRejectedChainStatus {
		liquidationCustodianDepositKey := lvdb.NewLiquidationCustodianDepositKey(actionData.TxReqID.String())
		newLiquidationCustodianDeposit, _ := NewLiquidationCustodianDeposit(
			actionData.TxReqID,
			actionData.PTokenId,
			actionData.IncogAddressStr,
			actionData.DepositedAmount,
			actionData.FreeCollateralSelected,
			common.PortalLiquidationCustodianDepositRejectedStatus,
		)

		err = db.StoredLiquidationCustodianDeposit([]byte(liquidationCustodianDepositKey), newLiquidationCustodianDeposit)

		if err != nil {
			Logger.log.Errorf("ERROR: an error occurred while store liquidation custodian deposit error %v", err)
			return nil
		}
	}

	return nil
}

func (blockchain *BlockChain) processPortalExpiredPortingRequest(beaconHeight uint64, instructions []string, currentPortalState *CurrentPortalState) error {
	if currentPortalState == nil {
		Logger.log.Errorf("current portal state is nil")
		return nil
	}
	if len(instructions) != 4 {
		return nil // skip the instruction
	}

	db := blockchain.GetDatabase()

	// unmarshal instructions content
	var actionData metadata.PortalExpiredWaitingPortingReqContent
	err := json.Unmarshal([]byte(instructions[3]), &actionData)
	if err != nil {
		return err
	}

	status := instructions[2]
	waitingPortingID := actionData.UniquePortingID

	if status == common.PortalExpiredWaitingPortingReqSuccessChainStatus {
		waitingPortingKey := lvdb.NewWaitingPortingReqKey(beaconHeight, waitingPortingID)
		waitingPortingReq := currentPortalState.WaitingPortingRequests[waitingPortingKey]
		if waitingPortingReq == nil {
			Logger.log.Errorf("[processPortalExpiredPortingRequest] waiting porting req nil with key : %v", waitingPortingKey)
			return nil
		}

		// get tokenID from redeemTokenID
		tokenID := waitingPortingReq.TokenID

		// update custodian state in matching custodians list (holding public tokens, locked amount)
		for _, matchCusDetail := range waitingPortingReq.Custodians {
			cusStateKey := lvdb.NewCustodianStateKey(beaconHeight, matchCusDetail.IncAddress)
			custodianState := currentPortalState.CustodianPoolState[cusStateKey]
			if custodianState == nil {
				Logger.log.Errorf("[checkAndBuildInstForExpiredWaitingPortingRequest] Error when get custodian state with key %v\n: ", cusStateKey)
				continue
			}

			Logger.log.Errorf("Custodian state before : %v", custodianState.FreeCollateral)

			_ = updateCustodianStateAfterExpiredPortingReq(
				custodianState, matchCusDetail.LockedAmountCollateral, matchCusDetail.Amount, tokenID)

			Logger.log.Errorf("Custodian state after : %v", custodianState.FreeCollateral)
		}

		// remove waiting porting request from waiting list
		delete(currentPortalState.WaitingPortingRequests, waitingPortingKey)

		// update status of porting ID  => expired/liquidated
		portingReqStatus := common.PortalPortingReqExpiredStatus
		if actionData.ExpiredByLiquidation {
			portingReqStatus = common.PortalPortingReqLiquidatedStatus
		}
		newPortingRequestStatus, err := NewPortingRequestState(
			waitingPortingReq.UniquePortingID,
			waitingPortingReq.TxReqID,
			tokenID,
			waitingPortingReq.PorterAddress,
			waitingPortingReq.Amount,
			waitingPortingReq.Custodians,
			waitingPortingReq.PortingFee,
			portingReqStatus,
			waitingPortingReq.BeaconHeight,
		)
		err = db.StorePortingRequestItem([]byte(waitingPortingKey), newPortingRequestStatus)
		if err != nil {
			Logger.log.Errorf("ERROR: an error occurred while store porting request item: %+v", err)
			return nil
		}

		// track expired waiting porting request status by portingID into DB
		expiredPortingTrackKey := lvdb.NewPortalExpiredPortingReqKey(waitingPortingID)
		expiredPortingTrackData := metadata.PortalExpiredWaitingPortingReqStatus{
			Status:               common.PortalExpiredPortingReqSuccessStatus,
			UniquePortingID:      waitingPortingID,
			ShardID:              actionData.ShardID,
			ExpiredByLiquidation: actionData.ExpiredByLiquidation,
			ExpiredBeaconHeight:  beaconHeight + 1,
		}
		expiredPortingTrackDataBytes, _ := json.Marshal(expiredPortingTrackData)
		err = db.TrackExpiredPortingReq(
			[]byte(expiredPortingTrackKey),
			expiredPortingTrackDataBytes,
		)
		if err != nil {
			Logger.log.Errorf("ERROR: an error occured while tracking expired porting request: %+v", err)
			return nil
		}

	} else if status == common.PortalLiquidationCustodianDepositRejectedChainStatus {
		// track expired waiting porting request status by portingID into DB
		expiredPortingTrackKey := lvdb.NewPortalExpiredPortingReqKey(waitingPortingID)
		expiredPortingTrackData := metadata.PortalExpiredWaitingPortingReqStatus{
			Status:               common.PortalExpiredPortingReqFailedStatus,
			UniquePortingID:      waitingPortingID,
			ShardID:              actionData.ShardID,
			ExpiredByLiquidation: actionData.ExpiredByLiquidation,
			ExpiredBeaconHeight:  beaconHeight + 1,
		}
		expiredPortingTrackDataBytes, _ := json.Marshal(expiredPortingTrackData)
		err = db.TrackExpiredPortingReq(
			[]byte(expiredPortingTrackKey),
			expiredPortingTrackDataBytes,
		)
		if err != nil {
			Logger.log.Errorf("ERROR: an error occured while tracking expired porting request: %+v", err)
			return nil
		}
	}

	return nil
}
