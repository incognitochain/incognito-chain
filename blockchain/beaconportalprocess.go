package blockchain

import (
	"encoding/json"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/rawdbv2"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/metadata"
	"sort"
	"strconv"
)

func (blockchain *BlockChain) processPortalInstructions(portalStateDB *statedb.StateDB, block *BeaconBlock) error {
	beaconHeight := block.Header.Height - 1

	currentPortalState, err := InitCurrentPortalStateFromDB(portalStateDB, beaconHeight)
	if err != nil {
		Logger.log.Error(err)
		return nil
	}

	// re-use update info of bridge
	updatingInfoByTokenID := map[common.Hash]UpdatingInfo{}

	for _, inst := range block.Body.Instructions {
		if len(inst) < 4 {
			continue // Not error, just not Portal instruction
		}

		var err error

		switch inst[0] {
			//porting request
			case strconv.Itoa(metadata.PortalUserRegisterMeta):
				err = blockchain.processPortalUserRegister(portalStateDB, beaconHeight, inst, currentPortalState)
			//exchange rates
			case strconv.Itoa(metadata.PortalExchangeRatesMeta):
				err = blockchain.processPortalExchangeRates(portalStateDB, beaconHeight, inst, currentPortalState)
			//custodian withdraw
			case strconv.Itoa(metadata.PortalCustodianWithdrawRequestMeta):
				err = blockchain.processPortalCustodianWithdrawRequest(portalStateDB, beaconHeight, inst, currentPortalState)
			//liquidation exchange rates
			case strconv.Itoa(metadata.PortalLiquidateTPExchangeRatesMeta):
				err = blockchain.processLiquidationTopPercentileExchangeRates(portalStateDB, beaconHeight, inst, currentPortalState)
			//liquidation custodian deposit
			case strconv.Itoa(metadata.PortalLiquidationCustodianDepositMeta):
				err = blockchain.processPortalLiquidationCustodianDeposit(portalStateDB, beaconHeight, inst, currentPortalState)
			//liquidation user redeem
			case strconv.Itoa(metadata.PortalRedeemLiquidateExchangeRatesMeta):
				err = blockchain.processPortalRedeemLiquidateExchangeRates(portalStateDB, beaconHeight, inst, currentPortalState, updatingInfoByTokenID)
			//custodian deposit
			case strconv.Itoa(metadata.PortalCustodianDepositMeta):
				err = blockchain.processPortalCustodianDeposit(portalStateDB, beaconHeight, inst, currentPortalState)
			case strconv.Itoa(metadata.PortalUserRequestPTokenMeta):
				err = blockchain.processPortalUserReqPToken(portalStateDB, beaconHeight, inst, currentPortalState, updatingInfoByTokenID)
			case strconv.Itoa(metadata.PortalRedeemRequestMeta):
				err = blockchain.processPortalRedeemRequest(portalStateDB, beaconHeight, inst, currentPortalState, updatingInfoByTokenID)
			case strconv.Itoa(metadata.PortalRequestUnlockCollateralMeta):
				err = blockchain.processPortalUnlockCollateral(portalStateDB, beaconHeight, inst, currentPortalState)
			case strconv.Itoa(metadata.PortalLiquidateCustodianMeta):
				err = blockchain.processPortalLiquidateCustodian(portalStateDB, beaconHeight, inst, currentPortalState)
			case strconv.Itoa(metadata.PortalRewardMeta):
				err = blockchain.processPortalReward(portalStateDB, beaconHeight, inst, currentPortalState)
			case strconv.Itoa(metadata.PortalRequestWithdrawRewardMeta):
				err = blockchain.processPortalWithdrawReward(portalStateDB, beaconHeight, inst, currentPortalState)
			case strconv.Itoa(metadata.PortalExpiredWaitingPortingReqMeta):
				err = blockchain.processPortalExpiredPortingRequest(portalStateDB, beaconHeight, inst, currentPortalState)
		}

		if err != nil {
			Logger.log.Error(err)
			return nil
		}
	}

	//save final exchangeRates
	err = blockchain.pickExchangesRatesFinal(beaconHeight, currentPortalState)
	if err != nil {
		Logger.log.Error(err)
		return nil
	}

	// update info of bridge portal token
	for _, updatingInfo := range updatingInfoByTokenID {
		var updatingAmt uint64
		var updatingType string
		if updatingInfo.countUpAmt > updatingInfo.deductAmt {
			updatingAmt = updatingInfo.countUpAmt - updatingInfo.deductAmt
			updatingType = "+"
		}
		if updatingInfo.countUpAmt < updatingInfo.deductAmt {
			updatingAmt = updatingInfo.deductAmt - updatingInfo.countUpAmt
			updatingType = "-"
		}
		err := statedb.UpdateBridgeTokenInfo(
			portalStateDB,
			updatingInfo.tokenID,
			updatingInfo.externalTokenID,
			updatingInfo.isCentralized,
			updatingAmt,
			updatingType,
		)
		if err != nil {
			return err
		}
	}

	// store updated currentPortalState to leveldb with new beacon height
	err = storePortalStateToDB(portalStateDB, beaconHeight + 1, currentPortalState)
	if err != nil {
		Logger.log.Error(err)
	}

	return nil
}

func (blockchain *BlockChain) processPortalCustodianDeposit(
	stateDB *statedb.StateDB, beaconHeight uint64, instructions []string, currentPortalState *CurrentPortalState) error {
	if currentPortalState == nil {
		Logger.log.Errorf("current portal state is nil")
		return nil
	}
	if len(instructions) != 4 {
		return nil // skip the instruction
	}

	// unmarshal instructions content
	var actionData metadata.PortalCustodianDepositContent
	err := json.Unmarshal([]byte(instructions[3]), &actionData)
	if err != nil {
		return err
	}

	depositStatus := instructions[2]
	if depositStatus == common.PortalCustodianDepositAcceptedChainStatus {
		keyCustodianState := statedb.GenerateCustodianStateObjectKey(beaconHeight, actionData.IncogAddressStr)
		keyCustodianStateStr := string(keyCustodianState[:])

		if currentPortalState.CustodianPoolState[keyCustodianStateStr] == nil {
			// new custodian
			newCustodian := statedb.NewCustodianStateWithValue(
				actionData.IncogAddressStr, actionData.DepositedAmount, actionData.DepositedAmount,
				nil, nil,
				actionData.RemoteAddresses, 0)
			currentPortalState.CustodianPoolState[keyCustodianStateStr] = newCustodian
		} else {
			// custodian deposited before
			// update state of the custodian
			custodian := currentPortalState.CustodianPoolState[keyCustodianStateStr]
			totalCollateral := custodian.GetTotalCollateral() + actionData.DepositedAmount
			freeCollateral := custodian.GetFreeCollateral() + actionData.DepositedAmount
			holdingPubTokens := custodian.GetHoldingPublicTokens()
			lockedAmountCollateral := custodian.GetLockedAmountCollateral()
			rewardAmount := custodian.GetRewardAmount()
			remoteAddresses := custodian.GetRemoteAddresses()
			for _, address := range actionData.RemoteAddresses {
				if existedAddr, _ := statedb.GetRemoteAddressByTokenID(remoteAddresses, address.GetPTokenID()); existedAddr == "" {
					remoteAddresses = append(remoteAddresses, address)
				}
			}

			newCustodian := statedb.NewCustodianStateWithValue(actionData.IncogAddressStr, totalCollateral, freeCollateral,
				holdingPubTokens, lockedAmountCollateral, remoteAddresses, rewardAmount)
			currentPortalState.CustodianPoolState[keyCustodianStateStr] = newCustodian
		}

		// store custodian deposit status into DB
		custodianDepositTrackData := metadata.PortalCustodianDepositStatus{
			Status:          common.PortalCustodianDepositAcceptedStatus,
			IncogAddressStr: actionData.IncogAddressStr,
			DepositedAmount: actionData.DepositedAmount,
		}
		custodianDepositDataBytes, _ := json.Marshal(custodianDepositTrackData)

		err = statedb.StoreCustodianDepositStatus(
			stateDB,
			actionData.TxReqID.String(),
			custodianDepositDataBytes,
		)
		if err != nil {
			Logger.log.Errorf("ERROR: an error occured while tracking custodian deposit collateral: %+v", err)
			return nil
		}
	} else if depositStatus == common.PortalCustodianDepositRefundChainStatus {
		// store custodian deposit status into DB
		custodianDepositTrackData := metadata.PortalCustodianDepositStatus{
			Status:          common.PortalCustodianDepositRefundStatus,
			IncogAddressStr: actionData.IncogAddressStr,
			DepositedAmount: actionData.DepositedAmount,
		}
		custodianDepositDataBytes, _ := json.Marshal(custodianDepositTrackData)

		err = statedb.StoreCustodianDepositStatus(
			stateDB,
			actionData.TxReqID.String(),
			custodianDepositDataBytes,
		)
		if err != nil {
			Logger.log.Errorf("ERROR: an error occured while tracking custodian deposit collateral: %+v", err)
			return nil
		}
	}

	return nil
}

func (blockchain *BlockChain) processPortalUserRegister(
	portalStateDB *statedb.StateDB,
	beaconHeight uint64, instructions []string, currentPortalState *CurrentPortalState) error {

	if currentPortalState == nil {
		Logger.log.Errorf("current portal state is nil")
		return nil
	}

	if len(instructions) != 4 {
		return nil // skip the instruction
	}

	// parse instruction
	var portingRequestContent metadata.PortalPortingRequestContent
	err := json.Unmarshal([]byte(instructions[3]), &portingRequestContent)
	if err != nil {
		Logger.log.Errorf("ERROR: an error occurred while unmarshaling content string of porting request contribution instruction: %+v", err)
		return nil
	}

	reqStatus := instructions[2]

	uniquePortingID := portingRequestContent.UniqueRegisterId
	txReqID := portingRequestContent.TxReqID
	tokenID := portingRequestContent.PTokenId

	porterAddress := portingRequestContent.IncogAddressStr
	amount := portingRequestContent.RegisterAmount

	custodiansDetail := portingRequestContent.Custodian
	portingFee := portingRequestContent.PortingFee


	switch reqStatus {
	case common.PortalPortingRequestAcceptedChainStatus:
		//verify custodian
		isCustodianAccepted := true
		for _, itemCustodian := range custodiansDetail {
			keyPortingRequestNewState := lvdb.NewCustodianStateKey(beaconHeight, itemCustodian.IncAddress)
			custodian, ok := currentPortalState.CustodianPoolState[keyPortingRequestNewState]
			if !ok {
				Logger.log.Errorf("ERROR: Custodian not found")
				isCustodianAccepted	= false
				break
			}

			if custodian.FreeCollateral < itemCustodian.LockedAmountCollateral {
				Logger.log.Errorf("ERROR: Custodian is not enough PRV, free collateral %v < lock amount %v", custodian.FreeCollateral, itemCustodian.LockedAmountCollateral)
				isCustodianAccepted	= false
				break
			}

			continue
		}

		if isCustodianAccepted == false {
			Logger.log.Errorf("ERROR: Custodian not found")
			return nil
		}

		// new request
		newPortingRequestStateWaiting := metadata.NewPortingRequestStatus(
			uniquePortingID,
			txReqID,
			tokenID,
			porterAddress,
			amount,
			custodiansDetail,
			portingFee,
			common.PortalPortingReqWaitingStatus,
			beaconHeight+1,
		)

		newPortingRequestStateAccept := metadata.NewPortingRequestStatus(
			uniquePortingID,
			txReqID,
			tokenID,
			porterAddress,
			amount,
			custodiansDetail,
			portingFee,
			common.PortalPortingRequestAcceptedStatus,
			beaconHeight+1,
		)

		newPortingTxRequestStateAccept := metadata.NewPortingRequestStatus(
			uniquePortingID,
			txReqID,
			tokenID,
			porterAddress,
			amount,
			custodiansDetail,
			portingFee,
			common.PortalPortingTxRequestAcceptedStatus,
			beaconHeight+1,
		)

		//save transaction
		newPortingTxRequestStatusBytes, _ := json.Marshal(newPortingTxRequestStateAccept)
		err = statedb.TrackPortalStateStatus(
			portalStateDB,
			statedb.PortalPortingRequestStatusPrefix(),
			[]byte(txReqID.String()),
			newPortingTxRequestStatusBytes,
		)

		if err != nil {
			Logger.log.Errorf("ERROR: an error occurred while store porting tx request item: %+v", err)
			return nil
		}

		//save success porting request
		newPortingRequestStatusBytes, _ := json.Marshal(newPortingRequestStateAccept)
		err = statedb.TrackPortalStateStatus(
			portalStateDB,
			statedb.PortalPortingRequestTxStatusPrefix(),
			[]byte(uniquePortingID),
			newPortingRequestStatusBytes,
		)
		if err != nil {
			Logger.log.Errorf("ERROR: an error occurred while store porting request item: %+v", err)
			return nil
		}

		//save custodian state
		for _, itemCustodian := range custodiansDetail {
			//update custodian state
			custodianKey := lvdb.NewCustodianStateKey(beaconHeight, itemCustodian.IncAddress)
			_ = UpdateCustodianWithNewAmount(currentPortalState, custodianKey, tokenID, itemCustodian.Amount, itemCustodian.LockedAmountCollateral)
		}

		//save waiting request porting state
		keyWaitingPortingRequest := lvdb.NewWaitingPortingReqKey(beaconHeight, portingRequestContent.UniqueRegisterId)
		Logger.log.Infof("Porting request, save waiting porting request with key %v", keyWaitingPortingRequest)
		currentPortalState.WaitingPortingRequests[keyWaitingPortingRequest] = newPortingRequestStateWaiting

		break
	case common.PortalPortingRequestRejectedChainStatus:
		txReqID := portingRequestContent.TxReqID

		newPortingRequest := statedb.NewPortingRequest(
			uniquePortingID,
			txReqID,
			tokenID,
			porterAddress,
			amount,
			custodiansDetail,
			portingFee,
			common.PortalPortingTxRequestRejectedStatus,
			beaconHeight+1,
		)

		//save transaction
		newPortingTxRequestStatusBytes, _ := json.Marshal(newPortingRequest)
		err = statedb.TrackPortalStateStatus(
			portalStateDB,
			statedb.PortalPortingRequestTxStatusPrefix(),
			[]byte(txReqID.String()),
			newPortingTxRequestStatusBytes,
		)

		if err != nil {
			Logger.log.Errorf("ERROR: an error occurred while store porting request item: %+v", err)
			return nil
		}
		break
	}

	return nil
}

func (blockchain *BlockChain) processPortalUserReqPToken(
	stateDB *statedb.StateDB,
	beaconHeight uint64, instructions []string,
	currentPortalState *CurrentPortalState,
	updatingInfoByTokenID map[common.Hash]UpdatingInfo) error {
	if currentPortalState == nil {
		Logger.log.Errorf("current portal state is nil")
		return nil
	}

	if len(instructions) != 4 {
		return nil // skip the instruction
	}
	db := blockchain.GetDatabase()

	// unmarshal instructions content
	var actionData metadata.PortalRequestPTokensContent
	err := json.Unmarshal([]byte(instructions[3]), &actionData)
	if err != nil {
		Logger.log.Errorf("Can not unmarshal instruction content %v\n", err)
		return nil
	}

	reqStatus := instructions[2]
	if reqStatus == common.PortalReqPTokensAcceptedChainStatus {
		// remove portingRequest from waitingPortingRequests
		waitingPortingReqKey := lvdb.NewWaitingPortingReqKey(beaconHeight, actionData.UniquePortingID)
		isRemoved := removeWaitingPortingReqByKey(waitingPortingReqKey, currentPortalState)
		if !isRemoved {
			Logger.log.Errorf("Can not remove waiting porting request from portal state")
			return nil
		}

		// make sure user can not re-use proof for other portingID
		// update status of porting request with portingID
		err = db.UpdatePortingRequestStatus(actionData.UniquePortingID, common.PortalPortingReqSuccessStatus)
		if err != nil {
			Logger.log.Errorf("ERROR: an error occurred while store porting request item status: %+v", err)
			return nil
		}

		// track reqPToken status by txID into DB
		reqPTokenTrackData := metadata.PortalRequestPTokensStatus{
			Status:          common.PortalReqPTokenAcceptedStatus,
			UniquePortingID: actionData.UniquePortingID,
			TokenID:         actionData.TokenID,
			IncogAddressStr: actionData.IncogAddressStr,
			PortingAmount:   actionData.PortingAmount,
			PortingProof:    actionData.PortingProof,
		}
		reqPTokenTrackDataBytes, _ := json.Marshal(reqPTokenTrackData)
		err = statedb.StoreRequestPTokenStatus(
			stateDB,
			actionData.TxReqID.String(),
			reqPTokenTrackDataBytes,
		)
		if err != nil {
			Logger.log.Errorf("ERROR: an error occured while tracking request ptoken tx: %+v", err)
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
			updatingInfo.countUpAmt += actionData.PortingAmount
		} else {
			updatingInfo = UpdatingInfo{
				countUpAmt:      actionData.PortingAmount,
				deductAmt:       0,
				tokenID:         *incTokenID,
				externalTokenID: nil,
				isCentralized:   false,
			}
		}
		updatingInfoByTokenID[*incTokenID] = updatingInfo

	} else if reqStatus == common.PortalReqPTokensRejectedChainStatus {
		reqPTokenTrackData := metadata.PortalRequestPTokensStatus{
			Status:          common.PortalReqPTokenRejectedStatus,
			UniquePortingID: actionData.UniquePortingID,
			TokenID:         actionData.TokenID,
			IncogAddressStr: actionData.IncogAddressStr,
			PortingAmount:   actionData.PortingAmount,
			PortingProof:    actionData.PortingProof,
		}
		reqPTokenTrackDataBytes, _ := json.Marshal(reqPTokenTrackData)
		err = statedb.StoreRequestPTokenStatus(
			stateDB,
			actionData.TxReqID.String(),
			reqPTokenTrackDataBytes,
		)
		if err != nil {
			Logger.log.Errorf("ERROR: an error occured while tracking request ptoken tx: %+v", err)
			return nil
		}
	}

	return nil
}

func (blockchain *BlockChain) processPortalExchangeRates(portalStateDB *statedb.StateDB, beaconHeight uint64, instructions []string, currentPortalState *CurrentPortalState) error {
	if currentPortalState == nil {
		Logger.log.Errorf("current portal state is nil")
		return nil
	}

	// parse instruction
	var portingExchangeRatesContent metadata.PortalExchangeRatesContent
	err := json.Unmarshal([]byte(instructions[3]), &portingExchangeRatesContent)
	if err != nil {
		Logger.log.Errorf("ERROR: an error occurred while unmarshaling content string of portal exchange rates instruction: %+v", err)
		return nil
	}

	Logger.log.Infof("Portal exchange rates, data input %v", portingExchangeRatesContent)

	reqStatus := instructions[2]

	switch reqStatus {
	case common.PortalExchangeRatesAcceptedChainStatus:
		//save db
		newExchangeRates := metadata.NewExchangeRatesRequestStatus(
			common.PortalExchangeRatesAcceptedStatus,
			portingExchangeRatesContent.SenderAddress,
			portingExchangeRatesContent.Rates,
		)

		newExchangeRatesStatusBytes, _ := json.Marshal(newExchangeRates)
		err = statedb.TrackPortalStateStatus(
			portalStateDB,
			statedb.PortalExchangeRatesRequestStatusPrefix(),
			[]byte(portingExchangeRatesContent.TxReqID.String()),
			newExchangeRatesStatusBytes,
		)

		if err != nil {
			Logger.log.Errorf("ERROR: Save exchange rates error: %+v", err)
			return err
		}

		currentPortalState.ExchangeRatesRequests[portingExchangeRatesContent.TxReqID.String()] = newExchangeRates

		Logger.log.Infof("Portal exchange rates, exchange rates request: count final exchange rate %v , exchange rate request %v", len(currentPortalState.FinalExchangeRatesState), len(currentPortalState.ExchangeRatesRequests))

	case common.PortalExchangeRatesRejectedChainStatus:
		//save db
		newExchangeRates := metadata.NewExchangeRatesRequestStatus(
			common.PortalExchangeRatesRejectedStatus,
			portingExchangeRatesContent.SenderAddress,
			nil,
		)

		newExchangeRatesStatusBytes, _ := json.Marshal(newExchangeRates)
		err = statedb.TrackPortalStateStatus(
			portalStateDB,
			statedb.PortalExchangeRatesRequestStatusPrefix(),
			[]byte(portingExchangeRatesContent.TxReqID.String()),
			newExchangeRatesStatusBytes,
		)

		if err != nil {
			Logger.log.Errorf("ERROR: Save exchange rates error: %+v", err)
			return err
		}

		break
	}

	return nil
}

func (blockchain *BlockChain) pickExchangesRatesFinal(beaconHeight uint64, currentPortalState *CurrentPortalState) error {
	exchangeRatesKey := lvdb.NewFinalExchangeRatesKey(beaconHeight)

	//convert to slice
	var btcExchangeRatesSlice []uint64
	var bnbExchangeRatesSlice []uint64
	var prvExchangeRatesSlice []uint64
	for _, v := range currentPortalState.ExchangeRatesRequests {
		for _, rate := range v.Rates {
			switch rate.PTokenID {
			case common.PortalBTCIDStr:
				btcExchangeRatesSlice = append(btcExchangeRatesSlice, rate.Rate)
				break
			case common.PortalBNBIDStr:
				bnbExchangeRatesSlice = append(bnbExchangeRatesSlice, rate.Rate)
				break
			case common.PRVIDStr:
				prvExchangeRatesSlice = append(prvExchangeRatesSlice, rate.Rate)
				break
			}
		}
	}

	//sort
	sort.SliceStable(btcExchangeRatesSlice, func(i, j int) bool {
		return btcExchangeRatesSlice[i] < btcExchangeRatesSlice[j]
	})

	sort.SliceStable(bnbExchangeRatesSlice, func(i, j int) bool {
		return bnbExchangeRatesSlice[i] < bnbExchangeRatesSlice[j]
	})

	sort.SliceStable(prvExchangeRatesSlice, func(i, j int) bool {
		return prvExchangeRatesSlice[i] < prvExchangeRatesSlice[j]
	})

	exchangeRatesList := make(map[string]lvdb.FinalExchangeRatesDetail)

	var btcAmount uint64
	var bnbAmount uint64
	var prvAmount uint64

	//get current value
	if len(btcExchangeRatesSlice) > 0 {
		btcAmount = calcMedian(btcExchangeRatesSlice)
	}

	if len(bnbExchangeRatesSlice) > 0 {
		bnbAmount = calcMedian(bnbExchangeRatesSlice)

	}

	if len(prvExchangeRatesSlice) > 0 {
		prvAmount = calcMedian(prvExchangeRatesSlice)
	}

	//if pre state exist

	//todo: need refactor code, not need write this code
	//update value when has exchange

	if exchangeRatesState, ok := currentPortalState.FinalExchangeRates[exchangeRatesKey]; ok {
		Logger.log.Infof("Portal final exchange rates, exits  key %v", exchangeRatesKey)

		var btcAmountPreState uint64
		var bnbAmountPreState uint64
		var prvAmountPreState uint64
		if value, ok := exchangeRatesState.Rates[common.PortalBTCIDStr]; ok {
			btcAmountPreState = value.Amount
		}

		if value, ok := exchangeRatesState.Rates[common.PortalBNBIDStr]; ok {
			bnbAmountPreState = value.Amount
		}

		if value, ok := exchangeRatesState.Rates[common.PRVIDStr]; ok {
			prvAmountPreState = value.Amount
		}

		//pick current value and pre value state
		btcAmount = choicePrice(btcAmount, btcAmountPreState)
		bnbAmount = choicePrice(bnbAmount, bnbAmountPreState)
		prvAmount = choicePrice(prvAmount, prvAmountPreState)
	}

	//select
	if btcAmount > 0 {
		exchangeRatesList[common.PortalBTCIDStr] = lvdb.FinalExchangeRatesDetail{
			Amount: btcAmount,
		}
	}

	if bnbAmount > 0 {
		exchangeRatesList[common.PortalBNBIDStr] = lvdb.FinalExchangeRatesDetail{
			Amount: bnbAmount,
		}
	}

	if prvAmount > 0 {
		exchangeRatesList[common.PRVIDStr] = lvdb.FinalExchangeRatesDetail{
			Amount: prvAmount,
		}
	}

	if len(exchangeRatesList) > 0 {
		currentPortalState.FinalExchangeRates[exchangeRatesKey] = &lvdb.FinalExchangeRates{
			Rates: exchangeRatesList,
		}

		Logger.log.Infof("Portal final exchange rates, picked key %v", exchangeRatesKey)
	}

	return nil
}

func calcMedian(ratesList []uint64) uint64 {
	mNumber := len(ratesList) / 2

	if len(ratesList)%2 == 0 {
		return (ratesList[mNumber-1] + ratesList[mNumber]) / 2
	}

	return ratesList[mNumber]
}

func choicePrice(currentPrice uint64, prePrice uint64) uint64 {
	if currentPrice > 0 {
		return currentPrice
	} else {
		if prePrice > 0 {
			return prePrice
		}
	}

	return 0
}

func (blockchain *BlockChain) processPortalRedeemRequest(
	stateDB *statedb.StateDB,
	beaconHeight uint64, instructions []string,
	currentPortalState *CurrentPortalState,
	updatingInfoByTokenID map[common.Hash]UpdatingInfo) error {
	if currentPortalState == nil {
		Logger.log.Errorf("current portal state is nil")
		return nil
	}

	if len(instructions) != 4 {
		return nil // skip the instruction
	}

	// unmarshal instructions content
	var actionData metadata.PortalRedeemRequestContent
	err := json.Unmarshal([]byte(instructions[3]), &actionData)
	if err != nil {
		Logger.log.Errorf("Can not unmarshal instruction content %v\n", err)
		return nil
	}

	// get tokenID from redeemTokenID
	tokenID :=  actionData.TokenID

	reqStatus := instructions[2]

	if reqStatus == common.PortalRedeemRequestAcceptedChainStatus {
		// add waiting redeem request into waiting redeems list
		keyWaitingRedeemRequest := statedb.GenerateWaitingRedeemRequestObjectKey(beaconHeight, actionData.UniqueRedeemID)
		keyWaitingRedeemRequestStr := string(keyWaitingRedeemRequest[:])
		redeemRequest := statedb.NewWaitingRedeemRequestWithValue(
			actionData.UniqueRedeemID,
			actionData.TokenID,
			actionData.RedeemerIncAddressStr,
			actionData.RemoteAddress,
			actionData.RedeemAmount,
			actionData.MatchingCustodianDetail,
			actionData.RedeemFee,
			beaconHeight + 1,
			actionData.TxReqID,
		)
		currentPortalState.WaitingRedeemRequests[keyWaitingRedeemRequestStr] = redeemRequest

		// update custodian state
		for _, cus := range actionData.MatchingCustodianDetail {
			custodianStateKey := statedb.GenerateCustodianStateObjectKey(beaconHeight, cus.GetIncognitoAddress())
			custodianStateKeyStr := string(custodianStateKey[:])
			holdingPubTokenTmp := currentPortalState.CustodianPoolState[custodianStateKeyStr].GetHoldingPublicTokens()
			if holdingPubTokenTmp[tokenID] < cus.GetAmount() {
				Logger.log.Errorf("[processPortalRedeemRequest] Amount holding public tokens is less than matching redeem amount")
				return nil
			}
			holdingPubTokenTmp[tokenID] -= cus.GetAmount()
			currentPortalState.CustodianPoolState[custodianStateKeyStr].SetHoldingPublicTokens(holdingPubTokenTmp)
		}

		// track status of redeem request by redeemID
		redeemRequestStatus := metadata.PortalRedeemRequestStatus{
			Status:                  common.PortalRedeemReqWaitingStatus,
			UniqueRedeemID:          actionData.UniqueRedeemID,
			TokenID:                 actionData.TokenID,
			RedeemAmount:            actionData.RedeemAmount,
			RedeemerIncAddressStr:   actionData.RedeemerIncAddressStr,
			RemoteAddress:           actionData.RemoteAddress,
			RedeemFee:               actionData.RedeemFee,
			MatchingCustodianDetail: actionData.MatchingCustodianDetail,
			TxReqID:                 actionData.TxReqID,
		}
		redeemRequestStatusBytes, _ := json.Marshal(redeemRequestStatus)
		err := statedb.StorePortalRedeemRequestStatus(
			stateDB,
			actionData.UniqueRedeemID,
			redeemRequestStatusBytes)
		if err != nil {
			Logger.log.Errorf("[processPortalRedeemRequest] Error when storing status of redeem request by redeemID: %v\n", err)
			return nil
		}

		// track status of redeem request by txReqID
		redeemRequestByTxIDStatus := metadata.PortalRedeemRequestStatus{
			Status:         common.PortalRedeemReqWaitingStatus,
			UniqueRedeemID: actionData.UniqueRedeemID,
		}
		redeemRequestByTxIDStatusBytes, _ := json.Marshal(redeemRequestByTxIDStatus)
		err = statedb.StorePortalRedeemRequestByTxIDStatus(
			stateDB, actionData.TxReqID.String(), redeemRequestByTxIDStatusBytes)
		if err != nil {
			Logger.log.Errorf("[processPortalRedeemRequest] Error when tracking status of redeem request by txReqID: %v\n", err)
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

	} else if reqStatus == common.PortalRedeemRequestRejectedChainStatus {
		// track status of redeem request by txReqID
		redeemRequestByTxIDStatus := metadata.PortalRedeemRequestStatus{
			Status:         common.PortalRedeemReqWaitingStatus,
			UniqueRedeemID: actionData.UniqueRedeemID,
		}
		redeemRequestByTxIDStatusBytes, _ := json.Marshal(redeemRequestByTxIDStatus)
		err = statedb.StorePortalRedeemRequestByTxIDStatus(
			stateDB, actionData.TxReqID.String(), redeemRequestByTxIDStatusBytes)
		if err != nil {
			Logger.log.Errorf("[processPortalRedeemRequest] Error when tracking status of redeem request by txReqID: %v\n", err)
			return nil
		}
	}

	return nil
}

func (blockchain *BlockChain) processPortalCustodianWithdrawRequest(portalStateDB *statedb.StateDB, beaconHeight uint64, instructions []string, currentPortalState *CurrentPortalState) error {
	if currentPortalState == nil {
		Logger.log.Errorf("current portal state is nil")
		return nil
	}

	if len(instructions) != 4 {
		return nil // skip the instruction
	}
	// parse instruction
	var custodianWithdrawRequestContent = metadata.PortalCustodianWithdrawRequestContent{}
	err := json.Unmarshal([]byte(instructions[3]), &custodianWithdrawRequestContent)
	if err != nil {
		Logger.log.Errorf("ERROR: an error occurred while unmarshaling content string of custodian withdraw request instruction: %+v", err)
		return nil
	}

	reqStatus := instructions[2]
	paymentAddress := custodianWithdrawRequestContent.PaymentAddress
	amount := custodianWithdrawRequestContent.Amount
	freeCollateral := custodianWithdrawRequestContent.RemainFreeCollateral
	txHash := custodianWithdrawRequestContent.TxReqID.String()

	switch reqStatus {
	case common.PortalCustodianWithdrawRequestAcceptedStatus:
		//save transaction
		newCustodianWithdrawRequest := metadata.NewCustodianWithdrawRequestStatus(
			paymentAddress,
			amount,
			common.PortalCustodianWithdrawReqAcceptedStatus,
			freeCollateral,
		)

		custodianKey := lvdb.NewCustodianStateKey(beaconHeight, paymentAddress)
		custodian, ok := currentPortalState.CustodianPoolState[custodianKey]

		if !ok {
			Logger.log.Errorf("ERROR: Custodian not found ")
			return nil
		}

		//check free collateral
		if amount > custodian.FreeCollateral {
			Logger.log.Errorf("ERROR: Free collateral is not enough to refund")
			return nil
		}

		contentStatusBytes, _ := json.Marshal(newCustodianWithdrawRequest)
		err = statedb.TrackPortalStateStatus(
			portalStateDB,
			statedb.PortalCustodianWithdrawStatusPrefix(),
			[]byte(txHash),
			contentStatusBytes,
		)

		if err != nil {
			Logger.log.Errorf("ERROR: an error occurred while store custodian withdraw item: %+v", err)
			return nil
		}

		//update custodian
		custodian.FreeCollateral = custodian.FreeCollateral - amount
		custodian.TotalCollateral = custodian.TotalCollateral - amount

		currentPortalState.CustodianPoolState[custodianKey] = custodian

	case common.PortalCustodianWithdrawRequestRejectedStatus:
		newCustodianWithdrawRequest := metadata.NewCustodianWithdrawRequestStatus(
			paymentAddress,
			amount,
			common.PortalCustodianWithdrawReqRejectStatus,
			freeCollateral,
		)

		contentStatusBytes, _ := json.Marshal(newCustodianWithdrawRequest)
		err = statedb.TrackPortalStateStatus(
			portalStateDB,
			statedb.PortalCustodianWithdrawStatusPrefix(),
			[]byte(txHash),
			contentStatusBytes,
		)

		if err != nil {
			Logger.log.Errorf("ERROR: an error occurred while store custodian withdraw item: %+v", err)
			return nil
		}
	}

	return nil
}

func (blockchain *BlockChain) processPortalUnlockCollateral(
	stateDB *statedb.StateDB,
	beaconHeight uint64, instructions []string,
	currentPortalState *CurrentPortalState) error {

	// unmarshal instructions content
	var actionData metadata.PortalRequestUnlockCollateralContent
	err := json.Unmarshal([]byte(instructions[3]), &actionData)
	if err != nil {
		Logger.log.Errorf("Can not unmarshal instruction content %v\n", err)
		return nil
	}

	// get tokenID from redeemTokenID
	tokenID :=  actionData.TokenID
	reqStatus := instructions[2]
	if reqStatus == common.PortalReqUnlockCollateralAcceptedChainStatus {
		// update custodian state (FreeCollateral, LockedAmountCollateral)
		custodianStateKey := statedb.GenerateCustodianStateObjectKey(beaconHeight, actionData.CustodianAddressStr)
		custodianStateKeyStr := string(custodianStateKey[:])
		finalExchangeRateKey := lvdb.NewFinalExchangeRatesKey(beaconHeight)
		_, err2 := updateFreeCollateralCustodian(
			currentPortalState.CustodianPoolState[custodianStateKeyStr],
			actionData.RedeemAmount, tokenID,
			currentPortalState.FinalExchangeRates[finalExchangeRateKey])
		if err2 != nil {
			Logger.log.Errorf("Error when update free collateral amount for custodian", err2)

			return nil
		}

		redeemID := actionData.UniqueRedeemID
		keyWaitingRedeemRequest := statedb.GenerateWaitingRedeemRequestObjectKey(beaconHeight, redeemID)
		keyWaitingRedeemRequestStr := string(keyWaitingRedeemRequest[:])

		// update redeem request state in WaitingRedeemRequest (remove custodian from matchingCustodianDetail)
		newCustodians, _ := removeCustodianFromMatchingRedeemCustodians(
			currentPortalState.WaitingRedeemRequests[keyWaitingRedeemRequestStr].GetCustodians(), actionData.CustodianAddressStr)
		currentPortalState.WaitingRedeemRequests[keyWaitingRedeemRequestStr].SetCustodians(newCustodians)

		// remove redeem request from WaitingRedeemRequest list when all matching custodians return public token to user
		// when list matchingCustodianDetail is empty
		if len(currentPortalState.WaitingRedeemRequests[keyWaitingRedeemRequestStr].GetCustodians()) == 0 {
			deleteWaitingRedeemRequest(currentPortalState, keyWaitingRedeemRequestStr)

			// update status of redeem request with redeemID
			err = updateRedeemRequestStatusByRedeemId(redeemID, common.PortalRedeemReqSuccessStatus, stateDB)
			if err != nil {
				Logger.log.Errorf("ERROR: an error occurred while updating redeem request status by redeemID: %+v", err)
				return nil
			}
		}

		// track reqUnlockCollateral status by txID into DB
		reqUnlockCollateralTrackData := metadata.PortalRequestUnlockCollateralStatus{
			Status:              common.PortalReqUnlockCollateralAcceptedStatus,
			UniqueRedeemID:      actionData.UniqueRedeemID,
			TokenID:             actionData.TokenID,
			CustodianAddressStr: actionData.CustodianAddressStr,
			RedeemAmount:        actionData.RedeemAmount,
			UnlockAmount:        actionData.UnlockAmount,
			RedeemProof:         actionData.RedeemProof,
		}
		reqUnlockCollateralTrackDataBytes, _ := json.Marshal(reqUnlockCollateralTrackData)
		err = statedb.StorePortalRequestUnlockCollateralStatus(
			stateDB,
			actionData.TxReqID.String(),
			reqUnlockCollateralTrackDataBytes,
		)
		if err != nil {
			Logger.log.Errorf("ERROR: an error occured while tracking request unlock collateral tx: %+v", err)
			return nil
		}

	} else if reqStatus == common.PortalReqUnlockCollateralRejectedChainStatus {
		// track reqUnlockCollateral status by txID into DB
		reqUnlockCollateralTrackData := metadata.PortalRequestUnlockCollateralStatus{
			Status:              common.PortalReqUnlockCollateralAcceptedStatus,
			UniqueRedeemID:      actionData.UniqueRedeemID,
			TokenID:             actionData.TokenID,
			CustodianAddressStr: actionData.CustodianAddressStr,
			RedeemAmount:        actionData.RedeemAmount,
			UnlockAmount:        actionData.UnlockAmount,
			RedeemProof:         actionData.RedeemProof,
		}
		reqUnlockCollateralTrackDataBytes, _ := json.Marshal(reqUnlockCollateralTrackData)
		err = statedb.StorePortalRequestUnlockCollateralStatus(
			stateDB,
			actionData.TxReqID.String(),
			reqUnlockCollateralTrackDataBytes,
		)
		if err != nil {
			Logger.log.Errorf("ERROR: an error occured while tracking request unlock collateral tx: %+v", err)
			return nil
		}
	}

	return nil
}