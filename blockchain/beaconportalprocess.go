package blockchain

import (
	"encoding/json"
	"errors"
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
		if len(inst) < 2 {
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

		// track custodian deposit at beaconHeight + 1 into DB
		custodianDepositTrackKey := lvdb.NewCustodianDepositKey(beaconHeight + 1, actionData.IncogAddressStr)
		custodianDepositTrackData := metadata.PortalCustodianDepositStatus{
			Status: common.PortalCustodianDepositAcceptedStatus,
			DepositedAmount : actionData.DepositedAmount,
			TxReqID: actionData.TxReqID,
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
		// track custodian deposit at beaconHeight + 1 into DB
		custodianDepositTrackKey := lvdb.NewCustodianDepositKey(beaconHeight + 1, actionData.IncogAddressStr)
		custodianDepositTrackData := metadata.PortalCustodianDepositStatus{
			Status: common.PortalCustodianDepositRefundStatus,
			DepositedAmount : actionData.DepositedAmount,
			TxReqID: actionData.TxReqID,
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

	// parse instruction
	var portingRequestContent metadata.PortalPortingRequestContent
	err := json.Unmarshal([]byte(instructions[3]), &portingRequestContent)
	if err != nil {
		Logger.log.Errorf("ERROR: an error occurred while unmarshaling content string of porting request contribution instruction: %+v", err)
		return nil
	}

	//check unique id from record from db
	keyPortingRequest := append(lvdb.PortalPortingRequestsPrefix, []byte(portingRequestContent.UniqueRegisterId)...)
	portingRequestExist, err := db.GetItemPortalByPrefix(keyPortingRequest)
	if err != nil {
		return err
	}

	if portingRequestExist != 0 {
		Logger.log.Errorf("Unique porting id is duplicated")
		return nil
	}

	//find custodian
	//todo: get exchangeRate via tokenid
	pickCustodian, err := pickCustodian(portingRequestContent, 1, currentPortalState.CustodianPoolState)

	if err != nil {
		return err
	}

	uniquePortingID := portingRequestContent.UniqueRegisterId
	txReqID := portingRequestContent.TxReqID
	tokenID := portingRequestContent.PTokenId

	porterAddress := portingRequestContent.IncogAddressStr
	amount := portingRequestContent.RegisterAmount

	custodians := pickCustodian
	portingFee := portingRequestContent.PortingFee

	// new request
	newPortingRequestState, err := NewPortingRequestState(
		uniquePortingID,
		txReqID,
		tokenID,
		porterAddress,
		amount,
		custodians,
		portingFee,
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
	for address, itemCustodian := range custodians {
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

	return nil
}

func pickCustodian(metadata metadata.PortalPortingRequestContent, exchangeRate uint64, custodianState map[string]*lvdb.CustodianState) (map[string]lvdb.MatchingPortingCustodianDetail, error) {

	type custodianStateSlice struct {
		Key   string
		Value *lvdb.CustodianState
	}

	//convert to slice
	var sortCustodianStateByFreeCollateral []custodianStateSlice
	for k, v := range custodianState {
		_, tokenIdExist := v.RemoteAddresses[metadata.PTokenId]
		if !tokenIdExist {
			continue
		}

		sortCustodianStateByFreeCollateral = append(sortCustodianStateByFreeCollateral, custodianStateSlice{k, v})
	}

	sort.Slice(sortCustodianStateByFreeCollateral, func(i, j int) bool {
		return sortCustodianStateByFreeCollateral[i].Value.FreeCollateral <= sortCustodianStateByFreeCollateral[j].Value.FreeCollateral
	})


	if len(sortCustodianStateByFreeCollateral) == 0 {
		return map[string]lvdb.MatchingPortingCustodianDetail{}, errors.New("Custodian not found")
	}

	//pick custodian
	amountAdaptable, _ := getAmountAdaptable(metadata.RegisterAmount, exchangeRate)

	//get only a custodian
	for _, kv := range sortCustodianStateByFreeCollateral {
		if kv.Value.FreeCollateral >= amountAdaptable {
			result := make(map[string]lvdb.MatchingPortingCustodianDetail)
			result[kv.Key] = lvdb.MatchingPortingCustodianDetail{
				RemoteAddress: metadata.PTokenAddress,
				Amount: metadata.RegisterAmount,
				LockedAmountCollateral: amountAdaptable,
			}

			return result, nil
		}
	}

	if len(sortCustodianStateByFreeCollateral) == 1 {
		return map[string]lvdb.MatchingPortingCustodianDetail{}, errors.New("Custodian not found")
	}

	//get multiple custodian
	var totalPubTokenAfterPick uint64

	multipleCustodian := make(map[string]lvdb.MatchingPortingCustodianDetail)
	for i := len(sortCustodianStateByFreeCollateral)-1; i >= 0; i-- {
		custodianItem := sortCustodianStateByFreeCollateral[i]
		if totalPubTokenAfterPick >= metadata.RegisterAmount {
			break
		}

		pricePubToken, _ := getPubTokenByTotalCollateral(custodianItem.Value.FreeCollateral, exchangeRate)
		collateral, _ := getAmountAdaptable(pricePubToken, exchangeRate)

		amountAdaptableEachCustodian, _ := getAmountAdaptable(pricePubToken, exchangeRate)
		//verify collateral
		if custodianItem.Value.FreeCollateral >= amountAdaptableEachCustodian {
			multipleCustodian[custodianItem.Key] = lvdb.MatchingPortingCustodianDetail{
				RemoteAddress: metadata.PTokenAddress,
				Amount: pricePubToken,
				LockedAmountCollateral: collateral,
			}

			totalPubTokenAfterPick = totalPubTokenAfterPick + pricePubToken

			continue
		}

		Logger.log.Errorf("current portal state is nil")
		return map[string]lvdb.MatchingPortingCustodianDetail{}, errors.New("Pick mulitple custodian is fail")
	}

	//verify total amount group custodian
	var verifyTotalPubTokenAfterPick uint64
	for _, eachCustodian := range multipleCustodian {
		verifyTotalPubTokenAfterPick = verifyTotalPubTokenAfterPick + eachCustodian.Amount
	}

	if verifyTotalPubTokenAfterPick != metadata.RegisterAmount {
		return map[string]lvdb.MatchingPortingCustodianDetail{}, errors.New("Total public token do not match")
	}

	return multipleCustodian, nil
}

func (blockchain *BlockChain) processPortalUserReqPToken(
	beaconHeight uint64, instructions []string, currentPortalState *CurrentPortalState) error {
	if currentPortalState == nil {
		Logger.log.Errorf("current portal state is nil")
		return nil
	}

	//// parse instruction
	//actionContentB64Str := instructions[1]
	//actionContentBytes, err := base64.StdEncoding.DecodeString(actionContentB64Str)
	//if err != nil {
	//	return err
	//}
	//var actionData metadata.PortalRequestPTokensAction
	//err = json.Unmarshal(actionContentBytes, &actionData)
	//if err != nil {
	//	return err
	//}
	//
	//meta := actionData.Meta
	//// check meta.UniquePortingID is in PortingRequests list in portal state or not
	//portingID := meta.UniquePortingID
	//keyWaitingPortingRequest := lvdb.NewWaitingPortingReqKey(beaconHeight, portingID)
	//waitingPortingRequest := currentPortalState.WaitingPortingRequests[keyWaitingPortingRequest]
	//
	//if waitingPortingRequest == nil {
	//	return errors.New("PortingID is not existed in waiting porting requests list")
	//}
	//
	//// check tokenID
	//if meta.TokenID != waitingPortingRequest.TokenID {
	//	return errors.New("TokenID is not correct in portingID req")
	//}
	//
	//// check porting amount
	//if meta.PortingAmount != waitingPortingRequest.Amount {
	//	return errors.New("PortingAmount is not correct in portingID req")
	//}
	//
	//if meta.TokenID == "BTC" {
	//	//todo:
	//} else if meta.TokenID == "BNB" {
	//	// parse txproof in meta
	//	txProofBNB, err := relaying.ParseProofFromB64EncodeJsonStr(meta.PortingProof)
	//	if err != nil {
	//		return errors.New("PortingProof is invalid")
	//	}
	//
	//	// parse Tx from Data in txProofBNB
	//	txBNB, err := relaying.ParseTxFromData(txProofBNB.Data)
	//	if err != nil {
	//		return errors.New("Data in PortingProof is invalid")
	//	}
	//
	//	// check whether amount transfer in txBNB is equal porting amount or not
	//	// check receiver and amount in tx
	//	// get list matching custodians in waitingPortingRequest
	//	custodians := waitingPortingRequest.Custodians
	//	outputs := txBNB.Msgs[0].(msg.SendMsg).Outputs
	//
	//	for _, cusDetail := range custodians {
	//		remoteAddressNeedToBeTransfer := cusDetail.RemoteAddress
	//		amountNeedToBeTransfer := cusDetail.Amount
	//
	//		for _, out := range outputs {
	//			addr := string(out.Address)
	//			if addr != remoteAddressNeedToBeTransfer {
	//				continue
	//			}
	//
	//			// calculate amount that was transferred to custodian's remote address
	//			amountTransfer := int64(0)
	//			for _, coin := range out.Coins {
	//				if coin.Denom == relaying.DenomBNB {
	//					amountTransfer += coin.Amount
	//				}
	//			}
	//
	//			if amountTransfer != int64(amountNeedToBeTransfer) {
	//				return fmt.Errorf("TxProof-BNB is invalid - Amount transfer to %s must be equal %d, but got %d",
	//					addr, amountNeedToBeTransfer, amountTransfer)
	//			}
	//		}
	//
	//	}
	//} else {
	//	return errors.New("TokenID is not supported currently on Portal")
	//}
	//
	//// create instruction mint ptoken to IncogAddressStr and send to shard
	////todo:


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

	exchangeRatesKey := lvdb.NewExchangeRatesRequestKey(
		beaconHeight + 1, portingExchangeRatesContent.TxReqID.String(),
		strconv.FormatInt(portingExchangeRatesContent.LockTime, 10),
	)

    exchangeRatesDetail := make(map[string]lvdb.ExchangeRatesDetail)
    for pTokenId, rates := range portingExchangeRatesContent.Rates {
		exchangeRatesDetail[pTokenId] = lvdb.ExchangeRatesDetail {
			Amount: rates.Amount,
		}
	}

	//check key from db
	exchangeRatesKeyExist, err := db.GetItemPortalByPrefix([]byte(exchangeRatesKey))
	if err != nil {
		return err
	}

	if exchangeRatesKeyExist != 0 {
		Logger.log.Errorf(" id is duplicated")
		return nil
	}

	//save db
	newExchangeRates, err := NewExchangeRatesState(
		portingExchangeRatesContent.SenderAddress,
		exchangeRatesDetail,
	)

	if err != nil {
		return err
	}

	err = db.StoreExchangeRatesRequestItem([]byte(exchangeRatesKey), newExchangeRates)

	if err != nil {
		return err
	}

	currentPortalState.ExchangeRatesRequests[exchangeRatesKey] = newExchangeRates

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
	exchangeRatesPreState, err := GetFinalExchangeRatesByKey(db, []byte(exchangeRatesKey))
	if err != nil {
		return  err
	}

	//pick
	exchangeRatesList := make(map[string]lvdb.FinalExchangeRatesDetail)
	btcAmount := exchangeRatesPreState.Rates["BTC"].Amount
	bnbAmount := exchangeRatesPreState.Rates["BNB"].Amount
	prvAmount := exchangeRatesPreState.Rates["PRV"].Amount

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