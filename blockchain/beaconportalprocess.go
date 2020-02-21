package blockchain

import (
	"encoding/json"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/database"
	"github.com/incognitochain/incognito-chain/database/lvdb"
	"github.com/incognitochain/incognito-chain/metadata"
	"sort"
	"strconv"
)

func (blockchain *BlockChain) processPortalInstructions(block *BeaconBlock, bd *[]database.BatchData) error {
	beaconHeight := block.Header.Height - 1
	db := blockchain.GetDatabase()

	currentPortalState, err := InitCurrentPortalStateFromDB(db, beaconHeight)
	if err != nil {
		Logger.log.Error(err)
		return nil
	}

	for _, inst := range block.Body.Instructions {
		if len(inst) < 4 {
			continue // Not error, just not Portal instruction
		}

		var err error

		switch inst[0] {
		case strconv.Itoa(metadata.PortalCustodianDepositMeta):
			err = blockchain.processPortalCustodianDeposit(beaconHeight, inst, currentPortalState)
		case strconv.Itoa(metadata.PortalUserRegisterMeta):
			err = blockchain.processPortalUserRegister(beaconHeight, inst, currentPortalState)
		case strconv.Itoa(metadata.PortalUserRequestPTokenMeta):
			err = blockchain.processPortalUserReqPToken(beaconHeight, inst, currentPortalState)
		case strconv.Itoa(metadata.PortalExchangeRatesMeta):
			err = blockchain.processPortalExchangeRates(beaconHeight, inst, currentPortalState)
		}

		if err != nil {
			Logger.log.Error(err)
			return nil
		}
	}

	//todo: check timeout register porting via beacon height
	// all request timeout ? unhold

	//save final exchangeRates
	if len(currentPortalState.ExchangeRatesRequests) > 0 {
		err = blockchain.pickExchangesRatesFinal(beaconHeight, currentPortalState)
		if err != nil {
			Logger.log.Error(err)
			return nil
		}
	}

	// store updated currentPortalState to leveldb with new beacon height
	err = storePortalStateToDB(db, beaconHeight+1, currentPortalState)
	if err != nil {
		Logger.log.Error(err)
	}

	return nil
}

func (blockchain *BlockChain) processPortalCustodianDeposit(
	beaconHeight uint64, instructions []string, currentPortalState *CurrentPortalState) error {
	if currentPortalState == nil {
		Logger.log.Errorf("current portal state is nil")
		return nil
	}
	if len(instructions) !=  4 {
		return nil  // skip the instruction
	}
	db := blockchain.GetDatabase()

	// unmarshal instructions content
	var actionData metadata.PortalCustodianDepositContent
	err := json.Unmarshal([]byte(instructions[3]), &actionData)
	if err != nil {
		return err
	}

	depositStatus := instructions[2]
	if depositStatus == common.PortalCustodianDepositAcceptedChainStatus {
		keyCustodianState := lvdb.NewCustodianStateKey(beaconHeight, actionData.IncogAddressStr)
		// update custodian state
		if currentPortalState.CustodianPoolState[keyCustodianState] == nil {
			// new custodian
			newCustodian, err := NewCustodianState(actionData.IncogAddressStr, actionData.DepositedAmount, actionData.DepositedAmount, nil, nil, actionData.RemoteAddresses)
			if err != nil {
				return err
			}
			currentPortalState.CustodianPoolState[keyCustodianState] = newCustodian
		} else {
			// custodian deposited before
			// update state of the custodian
			custodian := currentPortalState.CustodianPoolState[keyCustodianState]
			totalCollateral := custodian.TotalCollateral + actionData.DepositedAmount
			freeCollateral := custodian.FreeCollateral + actionData.DepositedAmount
			holdingPubTokens := custodian.HoldingPubTokens
			lockedAmountCollateral := custodian.LockedAmountCollateral
			remoteAddresses := custodian.RemoteAddresses
			for tokenSymbol, address := range actionData.RemoteAddresses {
				if remoteAddresses[tokenSymbol] == "" {
					remoteAddresses[tokenSymbol] = address
				}
			}

			newCustodian, err := NewCustodianState(actionData.IncogAddressStr, totalCollateral, freeCollateral, holdingPubTokens, lockedAmountCollateral, remoteAddresses)
			if err != nil {
				return err
			}
			currentPortalState.CustodianPoolState[keyCustodianState] = newCustodian
		}

		// track custodian deposit into DB
		custodianDepositTrackKey := lvdb.NewCustodianDepositKey(actionData.TxReqID.String())
		custodianDepositTrackData := metadata.PortalCustodianDepositStatus{
			Status: common.PortalCustodianDepositAcceptedStatus,
			IncogAddressStr: actionData.IncogAddressStr,
			DepositedAmount : actionData.DepositedAmount,
		}

		custodianDepositDataBytes, _ := json.Marshal(custodianDepositTrackData)
		err = db.TrackCustodianDepositCollateral(
			[]byte(custodianDepositTrackKey),
			custodianDepositDataBytes,
		)
		if err != nil {
			Logger.log.Errorf("ERROR: an error occured while tracking custodian deposit collateral: %+v", err)
			return nil
		}
	} else if depositStatus == common.PortalCustodianDepositRefundChainStatus {
		// track custodian deposit into DB
		custodianDepositTrackKey := lvdb.NewCustodianDepositKey(actionData.TxReqID.String())
		custodianDepositTrackData := metadata.PortalCustodianDepositStatus{
			Status: common.PortalCustodianDepositRefundStatus,
			IncogAddressStr: actionData.IncogAddressStr,
			DepositedAmount : actionData.DepositedAmount,
		}

		custodianDepositDataBytes, _ := json.Marshal(custodianDepositTrackData)
		err = db.TrackCustodianDepositCollateral(
			[]byte(custodianDepositTrackKey),
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
	beaconHeight uint64, instructions []string, currentPortalState *CurrentPortalState) error {
	db := blockchain.GetDatabase()

	if currentPortalState == nil {
		Logger.log.Errorf("current portal state is nil")
		return nil
	}

	if len(instructions) !=  4 {
		return nil  // skip the instruction
	}

	// parse instruction
	var portingRequestContent metadata.PortalPortingRequestContent
	err := json.Unmarshal([]byte(instructions[3]), &portingRequestContent)
	if err != nil {
		Logger.log.Errorf("ERROR: an error occurred while unmarshaling content string of porting request contribution instruction: %+v", err)
		return nil
	}

	reqStatus := instructions[2]

	switch reqStatus {
	case common.PortalPortingRequestWaitingStatus:
		uniquePortingID := portingRequestContent.UniqueRegisterId
		txReqID := portingRequestContent.TxReqID
		tokenID := portingRequestContent.PTokenId

		porterAddress := portingRequestContent.IncogAddressStr
		amount := portingRequestContent.RegisterAmount

		custodiansDetail := portingRequestContent.Custodian
		portingFee := portingRequestContent.PortingFee

		// new request
		newPortingRequestState, err := NewPortingRequestState(
			uniquePortingID,
			txReqID,
			tokenID,
			porterAddress,
			amount,
			custodiansDetail,
			portingFee,
			reqStatus,
			beaconHeight + 1,
		)

		if err != nil {
			return err
		}

		//save porting request
		keyPortingRequestNewState := lvdb.NewPortingRequestKey(beaconHeight + 1, portingRequestContent.UniqueRegisterId)
		err = db.StorePortingRequestItem([]byte(keyPortingRequestNewState), newPortingRequestState)
		if err != nil {
			Logger.log.Errorf("ERROR: an error occurred while store porting request item: %+v", err)
			return nil
		}

		//save custodian state
		for address, itemCustodian := range custodiansDetail {
			custodian := currentPortalState.CustodianPoolState[address]
			totalCollateral := custodian.TotalCollateral
			freeCollateral := custodian.FreeCollateral - itemCustodian.LockedAmountCollateral

			holdingPubTokensMapping := make(map[string]uint64)
			holdingPubTokensMapping[tokenID] = amount
			lockedAmountCollateralMapping := make(map[string]uint64)
			lockedAmountCollateralMapping[tokenID] = itemCustodian.LockedAmountCollateral

			lockedAmountCollateral := lockedAmountCollateralMapping
			holdingPubTokens := holdingPubTokensMapping
			remoteAddresses := custodian.RemoteAddresses

			newCustodian, err := NewCustodianState(portingRequestContent.IncogAddressStr, totalCollateral, freeCollateral, holdingPubTokens, lockedAmountCollateral, remoteAddresses)
			if err != nil {
				return err
			}
			currentPortalState.CustodianPoolState[address] = newCustodian
		}

		//save waiting request porting state
		currentPortalState.WaitingPortingRequests[keyPortingRequestNewState] = newPortingRequestState

		break
	case common.PortalLoadDataFailedStatus:
	case common.PortalDuplicateKeyStatus:
	case common.PortalItemNotFoundStatus:
	case common.PortalPortingFeesNotEnoughStatus:
		txReqID := portingRequestContent.TxReqID
		newPortingRequest := lvdb.PortingRequest{
			TxReqID:        txReqID,
			Status:			reqStatus,
			BeaconHeight:	beaconHeight + 1,
		}

		//save porting request
		newKey := reqStatus + txReqID.String() + portingRequestContent.UniqueRegisterId
		keyPortingRequestNewState := lvdb.NewPortingRequestKey(beaconHeight + 1, newKey)
		err = db.StorePortingRequestItem([]byte(keyPortingRequestNewState), newPortingRequest)
		if err != nil {
			Logger.log.Errorf("ERROR: an error occurred while store porting request item: %+v", err)
			return nil
		}
		break
	}

	return nil
}

func (blockchain *BlockChain) processPortalUserReqPToken(
	beaconHeight uint64, instructions []string, currentPortalState *CurrentPortalState) error {
	if currentPortalState == nil {
		Logger.log.Errorf("current portal state is nil")
		return nil
	}

	if len(instructions) !=  4 {
		return nil  // skip the instruction
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
		// track reqPToken and deposit proof into DB
		// make sure user can not re-use proof for other portingID
		reqPTokenTrackKey := lvdb.NewPortalReqPTokenKey(actionData.UniquePortingID)
		reqPTokenTrackData := metadata.PortalRequestPTokensStatus{
			Status: common.PortalReqPTokenAcceptedStatus,
			TxReqID: actionData.TxReqID,
		}
		reqPTokenTrackDataBytes, _ := json.Marshal(reqPTokenTrackData)
		err = db.TrackReqPTokens(
			[]byte(reqPTokenTrackKey),
			reqPTokenTrackDataBytes,
		)
		if err != nil {
			Logger.log.Errorf("ERROR: an error occured while tracking custodian deposit collateral: %+v", err)
			return nil
		}
	} else if reqStatus == common.PortalReqPTokensRejectedChainStatus {
		// track reqPToken and deposit proof into DB
		reqPTokenTrackKey := lvdb.NewPortalReqPTokenKey(actionData.UniquePortingID)
		reqPTokenTrackData := metadata.PortalRequestPTokensStatus{
			Status: common.PortalReqPTokenRejectedStatus,
			TxReqID: actionData.TxReqID,
		}
		reqPTokenTrackDataBytes, _ := json.Marshal(reqPTokenTrackData)
		err = db.TrackReqPTokens(
			[]byte(reqPTokenTrackKey),
			reqPTokenTrackDataBytes,
		)
		if err != nil {
			Logger.log.Errorf("ERROR: an error occured while tracking custodian deposit collateral: %+v", err)
			return nil
		}
	}

	return nil
}

func (blockchain *BlockChain) processPortalExchangeRates(beaconHeight uint64, instructions []string, currentPortalState *CurrentPortalState) error {
	db := blockchain.GetDatabase()

	// parse instruction
	var portingExchangeRatesContent metadata.PortalExchangeRatesContent
	err := json.Unmarshal([]byte(instructions[3]), &portingExchangeRatesContent)
	if err != nil {
		Logger.log.Errorf("ERROR: an error occurred while unmarshaling content string of portal exchange rates instruction: %+v", err)
		return nil
	}

	reqStatus := instructions[2]

	switch reqStatus {
	case common.PortalExchangeRatesSuccessStatus:
		exchangeRatesDetail := make(map[string]lvdb.ExchangeRatesDetail)
		for pTokenId, rates := range portingExchangeRatesContent.Rates {
			exchangeRatesDetail[pTokenId] = lvdb.ExchangeRatesDetail {
				Amount: rates.Amount,
			}
		}

		//save db
		newExchangeRates, _ := NewExchangeRatesState(
			portingExchangeRatesContent.SenderAddress,
			exchangeRatesDetail,
		)


		err = db.StoreExchangeRatesRequestItem([]byte(portingExchangeRatesContent.UniqueRequestId), newExchangeRates)

		if err != nil {
			Logger.log.Errorf("ERROR: Save exchange rates error: %+v", err)
			return err
		}

		currentPortalState.ExchangeRatesRequests[portingExchangeRatesContent.UniqueRequestId] = newExchangeRates
	case common.PortalLoadDataFailedStatus:
	case common.PortalDuplicateKeyStatus:
		//save db
		newExchangeRates := lvdb.ExchangeRatesRequest{
			SenderAddress: portingExchangeRatesContent.SenderAddress,
		}

		//todo: verify key
		err = db.StoreExchangeRatesRequestItem([]byte(reqStatus + portingExchangeRatesContent.UniqueRequestId), newExchangeRates)

		if err != nil {
			Logger.log.Errorf("ERROR: Save exchange rates error: %+v", err)
			return err
		}

		break
	}

	return nil
}

func (blockchain *BlockChain) pickExchangesRatesFinal(beaconHeight uint64, currentPortalState *CurrentPortalState) error  {

	db := blockchain.GetDatabase()

	//convert to slice
	var btcExchangeRatesSlice []uint64
	var bnbExchangeRatesSlice []uint64
	var prvExchangeRatesSlice []uint64
	for _, v := range currentPortalState.ExchangeRatesRequests {
		for key, rates := range v.Rates {
			switch key {
			case "BTC":
				btcExchangeRatesSlice = append(btcExchangeRatesSlice, rates.Amount)
				break
			case "BNB":
				bnbExchangeRatesSlice = append(bnbExchangeRatesSlice, rates.Amount)
				break
			case "PRV":
				prvExchangeRatesSlice = append(prvExchangeRatesSlice, rates.Amount)
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

	exchangeRatesKey := lvdb.NewFinalExchangeRatesKey(beaconHeight)
	exchangeRatesState, err := GetFinalExchangeRatesByKey(db, []byte(exchangeRatesKey))
	if err != nil {
		return  err
	}

	//pick
	exchangeRatesList := make(map[string]lvdb.FinalExchangeRatesDetail)
	btcAmount := exchangeRatesState.Rates["BTC"].Amount
	bnbAmount := exchangeRatesState.Rates["BNB"].Amount
	prvAmount := exchangeRatesState.Rates["PRV"].Amount

	//pick final rates of BTC
	if len(btcExchangeRatesSlice) > 0 {
		btcAmount = calcMedian(btcExchangeRatesSlice)
	}

	exchangeRatesList["BTC"] = lvdb.FinalExchangeRatesDetail{
		Amount: btcAmount,
	}

	//pick final rates of BNB
	if len(bnbExchangeRatesSlice) > 0 {
		bnbAmount = calcMedian(bnbExchangeRatesSlice)
	}

	exchangeRatesList["BTC"] = lvdb.FinalExchangeRatesDetail{
		Amount: bnbAmount,
	}

	//pick final rates of PRV
	if len(prvExchangeRatesSlice) > 0 {
		prvAmount = calcMedian(prvExchangeRatesSlice)
	}

	exchangeRatesList["BTC"] = lvdb.FinalExchangeRatesDetail{
		Amount: prvAmount,
	}

	//save to db
	newFinalExchangeRatesKey := lvdb.NewFinalExchangeRatesKey(beaconHeight + 1)
	err = db.StoreFinalExchangeRatesItem([]byte(newFinalExchangeRatesKey), lvdb.FinalExchangeRates{
		Rates: exchangeRatesList,
	})

	if err != nil {
		return  err
	}

	return nil
}

func calcMedian(ratesList []uint64) uint64 {
	mNumber := len(ratesList) / 2

	if len(ratesList) % 2 == 0 {
		return (ratesList[mNumber-1] + ratesList[mNumber]) / 2
	}

	return ratesList[mNumber]
}