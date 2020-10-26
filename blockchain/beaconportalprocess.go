package blockchain

import (
	"bytes"
	"encoding/json"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/metadata"
	"math/big"
	"sort"
	"strconv"
)

func (blockchain *BlockChain) processPortalInstructions(portalStateDB *statedb.StateDB, block *BeaconBlock) error {
	//TODO: uncomment before push code
	//if blockchain.config.ChainParams.Net == Testnet && block.Header.Height < 1580600 {
	//	return nil
	//}
	beaconHeight := block.Header.Height - 1
	currentPortalState, err := InitCurrentPortalStateFromDB(portalStateDB)
	if err != nil {
		Logger.log.Error(err)
		return nil
	}

	portalParams := blockchain.GetPortalParams(block.GetHeight())

	// re-use update info of bridge
	updatingInfoByTokenID := map[common.Hash]UpdatingInfo{}

	for _, inst := range block.Body.Instructions {
		if len(inst) < 4 {
			continue // Not error, just not Portal instruction
		}

		var err error
		switch inst[0] {
		// ============ Exchange rate ============
		case strconv.Itoa(metadata.PortalExchangeRatesMeta):
			err = blockchain.processPortalExchangeRates(portalStateDB, beaconHeight, inst, currentPortalState, portalParams)

		// ============ Custodian ============
		// custodian deposit collateral
		case strconv.Itoa(metadata.PortalCustodianDepositMeta):
			err = blockchain.processPortalCustodianDeposit(portalStateDB, beaconHeight, inst, currentPortalState, portalParams)
		// custodian withdraw collateral
		case strconv.Itoa(metadata.PortalCustodianWithdrawRequestMeta):
			err = blockchain.processPortalCustodianWithdrawRequest(portalStateDB, beaconHeight, inst, currentPortalState, portalParams)
		// custodian deposit collateral v3
		case strconv.Itoa(metadata.PortalCustodianDepositMetaV3):
			err = blockchain.processPortalCustodianDepositV3(portalStateDB, beaconHeight, inst, currentPortalState, portalParams)
		// custodian request withdraw collateral v3
		case strconv.Itoa(metadata.PortalCustodianWithdrawRequestMetaV3):
			err = blockchain.processPortalCustodianWithdrawV3(portalStateDB, beaconHeight, inst, currentPortalState, portalParams)

		// ============ Porting flow ============
		// porting request
		case strconv.Itoa(metadata.PortalUserRegisterMeta):
			err = blockchain.processPortalUserRegister(portalStateDB, beaconHeight, inst, currentPortalState, portalParams)
		// request ptoken
		case strconv.Itoa(metadata.PortalUserRequestPTokenMeta):
			err = blockchain.processPortalUserReqPToken(portalStateDB, beaconHeight, inst, currentPortalState, portalParams, updatingInfoByTokenID)

		// ============ Redeem flow ============
		// redeem request
		case strconv.Itoa(metadata.PortalRedeemRequestMeta):
			err = blockchain.processPortalRedeemRequest(portalStateDB, beaconHeight, inst, currentPortalState, portalParams, updatingInfoByTokenID)
		// custodian request matching waiting redeem requests
		case strconv.Itoa(metadata.PortalReqMatchingRedeemMeta):
			err = blockchain.processPortalReqMatchingRedeem(portalStateDB, beaconHeight, inst, currentPortalState, portalParams)
		case strconv.Itoa(metadata.PortalPickMoreCustodianForRedeemMeta):
			err = blockchain.processPortalPickMoreCustodiansForTimeOutWaitingRedeemReq(portalStateDB, beaconHeight, inst, currentPortalState, portalParams)
		// request unlock collateral
		case strconv.Itoa(metadata.PortalRequestUnlockCollateralMeta), strconv.Itoa(metadata.PortalRequestUnlockCollateralMetaV3):
			err = blockchain.processPortalUnlockCollateral(portalStateDB, beaconHeight, inst, currentPortalState, portalParams)

		// ============ Liquidation ============
		// liquidation custodian run away
		case strconv.Itoa(metadata.PortalLiquidateCustodianMeta), strconv.Itoa(metadata.PortalLiquidateCustodianMetaV3):
			err = blockchain.processPortalLiquidateCustodian(portalStateDB, beaconHeight, inst, currentPortalState, portalParams)
		//liquidation exchange rates
		case strconv.Itoa(metadata.PortalLiquidateTPExchangeRatesMeta):
			err = blockchain.processLiquidationTopPercentileExchangeRates(portalStateDB, beaconHeight, inst, currentPortalState, portalParams)
		// custodian topup
		case strconv.Itoa(metadata.PortalCustodianTopupMetaV2):
			err = blockchain.processPortalCustodianTopup(portalStateDB, beaconHeight, inst, currentPortalState, portalParams)
		// top up for waiting porting
		case strconv.Itoa(metadata.PortalTopUpWaitingPortingRequestMeta):
			err = blockchain.processPortalTopUpWaitingPorting(portalStateDB, beaconHeight, inst, currentPortalState, portalParams)
		// redeem from liquidation pool
		case strconv.Itoa(metadata.PortalRedeemFromLiquidationPoolMeta):
			err = blockchain.processPortalRedeemLiquidateExchangeRates(portalStateDB, beaconHeight, inst, currentPortalState, portalParams, updatingInfoByTokenID)
		// expired waiting porting request
		case strconv.Itoa(metadata.PortalExpiredWaitingPortingReqMeta):
			err = blockchain.processPortalExpiredPortingRequest(portalStateDB, beaconHeight, inst, currentPortalState, portalParams)

		// liquidation by exchange rate v3
		case strconv.Itoa(metadata.PortalLiquidateByRatesMetaV3):
			err = blockchain.processLiquidationByExchangeRatesV3(portalStateDB, beaconHeight, inst, currentPortalState, portalParams)
		// redeem from liquidation pool v3
		case strconv.Itoa(metadata.PortalRedeemFromLiquidationPoolMetaV3):
			err = blockchain.processPortalRedeemFromLiquidationPoolV3(portalStateDB, beaconHeight, inst, currentPortalState, portalParams, updatingInfoByTokenID)
		// custodian topup v3
		case strconv.Itoa(metadata.PortalCustodianTopupMetaV3):
			err = blockchain.processPortalCustodianTopupV3(portalStateDB, beaconHeight, inst, currentPortalState, portalParams)
		// top up for waiting porting v3
		case strconv.Itoa(metadata.PortalTopUpWaitingPortingRequestMetaV3):
			err = blockchain.processPortalTopUpWaitingPortingV3(portalStateDB, beaconHeight, inst, currentPortalState, portalParams)

		// ============ Reward ============
		// portal reward
		case strconv.Itoa(metadata.PortalRewardMeta), strconv.Itoa(metadata.PortalRewardMetaV3):
			err = blockchain.processPortalReward(portalStateDB, beaconHeight, inst, currentPortalState, portalParams)
		// request withdraw reward
		case strconv.Itoa(metadata.PortalRequestWithdrawRewardMeta):
			err = blockchain.processPortalWithdrawReward(portalStateDB, beaconHeight, inst, currentPortalState, portalParams)
		// total custodian reward instruction
		case strconv.Itoa(metadata.PortalTotalRewardCustodianMeta):
			err = blockchain.processPortalTotalCustodianReward(portalStateDB, beaconHeight, inst, currentPortalState, portalParams)

		// ============ Portal smart contract ============
		// todo: add more metadata need to unlock token from sc
		case strconv.Itoa(metadata.PortalCustodianWithdrawConfirmMetaV3),
			strconv.Itoa(metadata.PortalRedeemFromLiquidationPoolConfirmMetaV3),
			strconv.Itoa(metadata.PortalLiquidateRunAwayCustodianConfirmMetaV3):
			err = blockchain.processPortalConfirmWithdrawInstV3(portalStateDB, beaconHeight, inst, currentPortalState, portalParams)
		}

		if err != nil {
			Logger.log.Error(err)
			return nil
		}
	}

	//save final exchangeRates
	blockchain.pickExchangesRatesFinal(currentPortalState)

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
	err = storePortalStateToDB(portalStateDB, currentPortalState)
	if err != nil {
		Logger.log.Error(err)
	}

	return nil
}

func (blockchain *BlockChain) processPortalCustodianDeposit(
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
	var actionData metadata.PortalCustodianDepositContent
	err := json.Unmarshal([]byte(instructions[3]), &actionData)
	if err != nil {
		return err
	}

	depositStatus := instructions[2]
	if depositStatus == common.PortalCustodianDepositAcceptedChainStatus {
		// add custodian to custodian pool
		newCustodian := addCustodianToPool(
			currentPortalState.CustodianPoolState,
			actionData.IncogAddressStr,
			actionData.DepositedAmount,
			common.PRVIDStr,
			actionData.RemoteAddresses)
		keyCustodianStateStr := statedb.GenerateCustodianStateObjectKey(actionData.IncogAddressStr).String()
		currentPortalState.CustodianPoolState[keyCustodianStateStr] = newCustodian

		// store custodian deposit status into DB
		custodianDepositTrackData := metadata.PortalCustodianDepositStatus{
			Status:          common.PortalCustodianDepositAcceptedStatus,
			IncogAddressStr: actionData.IncogAddressStr,
			DepositedAmount: actionData.DepositedAmount,
			RemoteAddresses: actionData.RemoteAddresses,
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
			RemoteAddresses: actionData.RemoteAddresses,
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
	beaconHeight uint64, instructions []string,
	currentPortalState *CurrentPortalState,
	portalParams PortalParams) error {

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
	shardHeight := portingRequestContent.ShardHeight
	shardId := portingRequestContent.ShardID

	switch reqStatus {
	case common.PortalPortingRequestAcceptedChainStatus:
		//verify custodian
		isCustodianAccepted := true
		for _, itemCustodian := range custodiansDetail {
			keyPortingRequestNewState := statedb.GenerateCustodianStateObjectKey(itemCustodian.IncAddress)
			keyPortingRequestNewStateStr := keyPortingRequestNewState.String()
			custodian, ok := currentPortalState.CustodianPoolState[keyPortingRequestNewStateStr]
			if !ok {
				Logger.log.Errorf("ERROR: Custodian not found")
				isCustodianAccepted = false
				break
			}

			if custodian.GetFreeCollateral() < itemCustodian.LockedAmountCollateral {
				Logger.log.Errorf("ERROR: Custodian is not enough PRV, free collateral %v < lock amount %v", custodian.GetFreeCollateral(), itemCustodian.LockedAmountCollateral)
				isCustodianAccepted = false
				break
			}

			continue
		}

		if isCustodianAccepted == false {
			Logger.log.Errorf("ERROR: Custodian not found")
			return nil
		}

		// new request
		newWaitingPortingRequestState := statedb.NewWaitingPortingRequestWithValue(
			uniquePortingID,
			txReqID,
			tokenID,
			porterAddress,
			amount,
			custodiansDetail,
			portingFee,
			beaconHeight+1,
			shardHeight,
			shardId,
		)

		newPortingRequestState := metadata.NewPortingRequestStatus(
			uniquePortingID,
			txReqID,
			tokenID,
			porterAddress,
			amount,
			custodiansDetail,
			portingFee,
			common.PortalPortingReqWaitingStatus,
			beaconHeight+1,
			shardHeight,
			shardId,
		)

		newPortingTxRequestState := metadata.NewPortingRequestStatus(
			uniquePortingID,
			txReqID,
			tokenID,
			porterAddress,
			amount,
			custodiansDetail,
			portingFee,
			common.PortalPortingTxRequestAcceptedStatus,
			beaconHeight+1,
			shardHeight,
			shardId,
		)

		//save transaction
		newPortingTxRequestStatusBytes, _ := json.Marshal(newPortingTxRequestState)
		err = statedb.StorePortalPortingRequestByTxIDStatus(
			portalStateDB,
			txReqID.String(),
			newPortingTxRequestStatusBytes,
		)
		if err != nil {
			Logger.log.Errorf("ERROR: an error occurred while store porting tx request item: %+v", err)
			return nil
		}

		//save success porting request
		newPortingRequestStatusBytes, _ := json.Marshal(newPortingRequestState)
		err = statedb.StorePortalPortingRequestStatus(
			portalStateDB,
			uniquePortingID,
			newPortingRequestStatusBytes,
		)
		if err != nil {
			Logger.log.Errorf("ERROR: an error occurred while store porting request item: %+v", err)
			return nil
		}

		//save custodian state
		for _, itemCustodian := range custodiansDetail {
			//update custodian state
			_ = UpdateCustodianStateAfterMatchingPortingRequest(currentPortalState, itemCustodian, tokenID)
		}

		//save waiting request porting state
		keyWaitingPortingRequest := statedb.GeneratePortalWaitingPortingRequestObjectKey(portingRequestContent.UniqueRegisterId)
		Logger.log.Infof("Porting request, save waiting porting request with key %v", keyWaitingPortingRequest)
		currentPortalState.WaitingPortingRequests[keyWaitingPortingRequest.String()] = newWaitingPortingRequestState

		break
	case common.PortalPortingRequestRejectedChainStatus:
		txReqID := portingRequestContent.TxReqID

		newPortingRequest := metadata.NewPortingRequestStatus(
			uniquePortingID,
			txReqID,
			tokenID,
			porterAddress,
			amount,
			custodiansDetail,
			portingFee,
			common.PortalPortingTxRequestRejectedStatus,
			beaconHeight+1,
			shardHeight,
			shardId,
		)

		//save transaction
		newPortingTxRequestStatusBytes, _ := json.Marshal(newPortingRequest)
		err = statedb.StorePortalPortingRequestByTxIDStatus(
			portalStateDB,
			txReqID.String(),
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
	var actionData metadata.PortalRequestPTokensContent
	err := json.Unmarshal([]byte(instructions[3]), &actionData)
	if err != nil {
		Logger.log.Errorf("Can not unmarshal instruction content %v - Error: %v\n", instructions[3], err)
		return nil
	}

	reqStatus := instructions[2]
	if reqStatus == common.PortalReqPTokensAcceptedChainStatus {
		waitingPortingReqKey := statedb.GeneratePortalWaitingPortingRequestObjectKey(actionData.UniquePortingID)
		waitingPortingReqKeyStr := waitingPortingReqKey.String()
		waitingPortingReq := currentPortalState.WaitingPortingRequests[waitingPortingReqKeyStr]

		// update holding public token for custodians
		for _, cusDetail := range waitingPortingReq.Custodians() {
			custodianKey := statedb.GenerateCustodianStateObjectKey(cusDetail.IncAddress)
			UpdateCustodianStateAfterUserRequestPToken(currentPortalState, custodianKey.String(), waitingPortingReq.TokenID(), cusDetail.Amount)
		}

		// remove portingRequest from waitingPortingRequests
		deleteWaitingPortingRequest(currentPortalState, waitingPortingReqKeyStr)
		statedb.DeleteWaitingPortingRequest(stateDB, waitingPortingReq.UniquePortingID())
		// make sure user can not re-use proof for other portingID
		// update status of porting request with portingID

		//update new status of porting request
		portingRequestState, err := statedb.GetPortalPortingRequestStatus(stateDB, actionData.UniquePortingID)
		if err != nil {
			Logger.log.Errorf("Has an error occurred while get porting request status: %+v", err)
			return nil
		}

		var portingRequestStatus metadata.PortingRequestStatus
		err = json.Unmarshal(portingRequestState, &portingRequestStatus)
		if err != nil {
			Logger.log.Errorf("ERROR: an error occurred while unmarshal PortingRequestStatus: %+v", err)
			return nil
		}

		portingRequestStatus.Status = common.PortalPortingReqSuccessStatus
		newPortingRequestStatusBytes, _ := json.Marshal(portingRequestStatus)
		err = statedb.StorePortalPortingRequestStatus(
			stateDB,
			actionData.UniquePortingID,
			newPortingRequestStatusBytes,
		)
		if err != nil {
			Logger.log.Errorf("ERROR: an error occurred while store porting request item status: %+v", err)
			return nil
		}
		//end

		// track reqPToken status by txID into DB
		reqPTokenTrackData := metadata.PortalRequestPTokensStatus{
			Status:          common.PortalReqPTokenAcceptedStatus,
			UniquePortingID: actionData.UniquePortingID,
			TokenID:         actionData.TokenID,
			IncogAddressStr: actionData.IncogAddressStr,
			PortingAmount:   actionData.PortingAmount,
			PortingProof:    actionData.PortingProof,
			TxReqID:         actionData.TxReqID,
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
			TxReqID:         actionData.TxReqID,
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

func (blockchain *BlockChain) processPortalExchangeRates(
	portalStateDB *statedb.StateDB,
	beaconHeight uint64,
	instructions []string,
	currentPortalState *CurrentPortalState,
	portalParams PortalParams) error {
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

	reqStatus := instructions[2]
	Logger.log.Infof("Portal exchange rates, data input: %+v, status: %+v", portingExchangeRatesContent, reqStatus)

	switch reqStatus {
	case common.PortalExchangeRatesAcceptedChainStatus:
		//save db
		newExchangeRates := metadata.NewExchangeRatesRequestStatus(
			common.PortalExchangeRatesAcceptedStatus,
			portingExchangeRatesContent.SenderAddress,
			portingExchangeRatesContent.Rates,
		)

		newExchangeRatesStatusBytes, _ := json.Marshal(newExchangeRates)
		err = statedb.StorePortalExchangeRateStatus(
			portalStateDB,
			portingExchangeRatesContent.TxReqID.String(),
			newExchangeRatesStatusBytes,
		)

		if err != nil {
			Logger.log.Errorf("ERROR: Save exchange rates error: %+v", err)
			return nil
		}

		currentPortalState.ExchangeRatesRequests[portingExchangeRatesContent.TxReqID.String()] = newExchangeRates

		Logger.log.Infof("Portal exchange rates, exchange rates request: total exchange rate request %v", len(currentPortalState.ExchangeRatesRequests))

	case common.PortalExchangeRatesRejectedChainStatus:
		//save db
		newExchangeRates := metadata.NewExchangeRatesRequestStatus(
			common.PortalExchangeRatesRejectedStatus,
			portingExchangeRatesContent.SenderAddress,
			nil,
		)

		newExchangeRatesStatusBytes, _ := json.Marshal(newExchangeRates)
		err = statedb.StorePortalExchangeRateStatus(
			portalStateDB,
			portingExchangeRatesContent.TxReqID.String(),
			newExchangeRatesStatusBytes,
		)
		if err != nil {
			Logger.log.Errorf("ERROR: Save exchange rates error: %+v", err)
			return nil
		}
	}

	return nil
}

func (blockchain *BlockChain) pickExchangesRatesFinal(currentPortalState *CurrentPortalState) {
	// sort exchange rate requests by rate
	sumRates := map[string][]uint64{}

	for _, req := range currentPortalState.ExchangeRatesRequests {
		for _, rate := range req.Rates {
			sumRates[rate.PTokenID] = append(sumRates[rate.PTokenID], rate.Rate)
		}
	}

	updateFinalExchangeRates := currentPortalState.FinalExchangeRatesState.Rates()
	if updateFinalExchangeRates == nil {
		updateFinalExchangeRates = map[string]statedb.FinalExchangeRatesDetail{}
	}
	for tokenID, rates := range sumRates {
		// sort rates
		sort.SliceStable(rates, func(i, j int) bool {
			return rates[i] < rates[j]
		})

		// pick one median rate to make final rate for tokenID
		medianRate := calcMedian(rates)

		if medianRate > 0 {
			updateFinalExchangeRates[tokenID] = statedb.FinalExchangeRatesDetail{Amount: medianRate}
		}
	}
	currentPortalState.FinalExchangeRatesState = statedb.NewFinalExchangeRatesStateWithValue(updateFinalExchangeRates)
}

func calcMedian(ratesList []uint64) uint64 {
	mNumber := len(ratesList) / 2

	if len(ratesList)%2 == 0 {
		return (ratesList[mNumber-1] + ratesList[mNumber]) / 2
	}

	return ratesList[mNumber]
}

func (blockchain *BlockChain) processPortalCustodianWithdrawRequest(
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
	case common.PortalCustodianWithdrawRequestAcceptedChainStatus:
		//save transaction
		newCustodianWithdrawRequest := metadata.NewCustodianWithdrawRequestStatus(
			paymentAddress,
			amount,
			common.PortalCustodianWithdrawReqAcceptedStatus,
			freeCollateral,
		)

		custodianKey := statedb.GenerateCustodianStateObjectKey(paymentAddress)
		custodianKeyStr := custodianKey.String()
		custodian, ok := currentPortalState.CustodianPoolState[custodianKeyStr]

		if !ok {
			Logger.log.Errorf("ERROR: Custodian not found ")
			return nil
		}

		//check free collateral
		if amount > custodian.GetFreeCollateral() {
			Logger.log.Errorf("ERROR: Free collateral is not enough to refund")
			return nil
		}

		contentStatusBytes, _ := json.Marshal(newCustodianWithdrawRequest)
		err = statedb.StorePortalCustodianWithdrawCollateralStatus(
			portalStateDB,
			txHash,
			contentStatusBytes,
		)

		if err != nil {
			Logger.log.Errorf("ERROR: an error occurred while store custodian withdraw item: %+v", err)
			return nil
		}

		//update custodian
		custodian.SetFreeCollateral(custodian.GetFreeCollateral() - amount)
		custodian.SetTotalCollateral(custodian.GetTotalCollateral() - amount)

		currentPortalState.CustodianPoolState[custodianKeyStr] = custodian

	case common.PortalCustodianWithdrawRequestRejectedChainStatus:
		newCustodianWithdrawRequest := metadata.NewCustodianWithdrawRequestStatus(
			paymentAddress,
			amount,
			common.PortalCustodianWithdrawReqRejectStatus,
			freeCollateral,
		)

		contentStatusBytes, _ := json.Marshal(newCustodianWithdrawRequest)
		err = statedb.StorePortalCustodianWithdrawCollateralStatus(
			portalStateDB,
			txHash,
			contentStatusBytes,
		)
		if err != nil {
			Logger.log.Errorf("ERROR: an error occurred while store custodian withdraw item: %+v", err)
			return nil
		}
	}

	return nil
}

func (blockchain *BlockChain) processPortalCustodianDepositV3(
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
	var actionData metadata.PortalCustodianDepositContentV3
	err := json.Unmarshal([]byte(instructions[3]), &actionData)
	if err != nil {
		return err
	}

	depositStatus := instructions[2]
	if depositStatus == common.PortalCustodianDepositV3AcceptedChainStatus {
		// add custodian to custodian pool
		newCustodian := addCustodianToPool(
			currentPortalState.CustodianPoolState,
			actionData.IncAddressStr,
			actionData.DepositAmount,
			actionData.ExternalTokenID,
			actionData.RemoteAddresses)
		keyCustodianStateStr := statedb.GenerateCustodianStateObjectKey(actionData.IncAddressStr).String()
		currentPortalState.CustodianPoolState[keyCustodianStateStr] = newCustodian

		// store custodian deposit status into DB
		custodianDepositTrackData := metadata.PortalCustodianDepositStatusV3{
			Status:           common.PortalCustodianDepositV3AcceptedStatus,
			IncAddressStr:    actionData.IncAddressStr,
			RemoteAddresses:  actionData.RemoteAddresses,
			DepositAmount:    actionData.DepositAmount,
			ExternalTokenID:  actionData.ExternalTokenID,
			UniqExternalTxID: actionData.UniqExternalTxID,
		}
		custodianDepositDataBytes, _ := json.Marshal(custodianDepositTrackData)
		err = statedb.StoreCustodianDepositStatusV3(
			stateDB,
			actionData.TxReqID.String(),
			custodianDepositDataBytes,
		)
		if err != nil {
			Logger.log.Errorf("ERROR: an error occured while tracking custodian deposit collateral: %+v", err)
			return nil
		}

		// store uniq external tx
		err := statedb.InsertPortalExternalTxHashSubmitted(stateDB, actionData.UniqExternalTxID)
		if err != nil {
			Logger.log.Errorf("ERROR: an error occured while tracking uniq external tx id: %+v", err)
			return nil
		}
	} else if depositStatus == common.PortalCustodianDepositV3RejectedChainStatus {
		// store custodian deposit status into DB
		custodianDepositTrackData := metadata.PortalCustodianDepositStatusV3{
			Status:           common.PortalCustodianDepositV3RejectedStatus,
			IncAddressStr:    actionData.IncAddressStr,
			RemoteAddresses:  actionData.RemoteAddresses,
			DepositAmount:    actionData.DepositAmount,
			ExternalTokenID:  actionData.ExternalTokenID,
			UniqExternalTxID: actionData.UniqExternalTxID,
		}
		custodianDepositDataBytes, _ := json.Marshal(custodianDepositTrackData)
		err = statedb.StoreCustodianDepositStatusV3(
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

func (blockchain *BlockChain) processPortalCustodianWithdrawV3(
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
	// parse instruction
	var instContent = metadata.PortalCustodianWithdrawRequestContentV3{}
	err := json.Unmarshal([]byte(instructions[3]), &instContent)
	if err != nil {
		Logger.log.Errorf("ERROR: an error occurred while unmarshaling content string of custodian withdraw request instruction: %+v", err)
		return nil
	}
	custodianIncAddress := instContent.CustodianIncAddress
	custodianExtAddress := instContent.CustodianExternalAddress
	externalTokenID := instContent.ExternalTokenID
	txId := instContent.TxReqID
	amountBN := instContent.Amount

	status := instructions[2]
	statusInt := common.PortalCustodianWithdrawReqV3RejectStatus
	if status == common.PortalCustodianWithdrawRequestV3AcceptedChainStatus {
		statusInt = common.PortalCustodianWithdrawReqV3AcceptedStatus

		custodianKeyStr := statedb.GenerateCustodianStateObjectKey(custodianIncAddress).String()
		custodian, ok := currentPortalState.CustodianPoolState[custodianKeyStr]
		if !ok {
			Logger.log.Errorf("ERROR: Custodian not found")
			return nil
		}

		// check free collateral
		if bytes.Equal(common.FromHex(externalTokenID), common.FromHex(common.EthAddrStr)) {
			// Convert Wei to Gwei for Ether
			amountBN = amountBN.Div(amountBN, big.NewInt(1000000000))
		}
		amount := amountBN.Uint64()
		if amount > custodian.GetFreeTokenCollaterals()[externalTokenID] {
			Logger.log.Errorf("ERROR: Free collateral is not enough to withdraw")
			return nil
		}

		updatedCustodian := UpdateCustodianStateAfterWithdrawCollateral(custodian, externalTokenID, amount)
		currentPortalState.CustodianPoolState[custodianKeyStr] = updatedCustodian
	}

	// store status of requesting withdraw collateral
	statusData := metadata.NewCustodianWithdrawRequestStatusV3(
		custodianIncAddress,
		custodianExtAddress,
		externalTokenID,
		amountBN,
		txId,
		statusInt)
	contentStatusBytes, _ := json.Marshal(statusData)
	err = statedb.StorePortalCustodianWithdrawCollateralStatusV3(
		portalStateDB,
		statusData.TxReqID.String(),
		contentStatusBytes,
	)
	if err != nil {
		Logger.log.Errorf("ERROR: an error occurred while store custodian withdraw v3 item: %+v", err)
		return nil
	}

	return nil
}

func (blockchain *BlockChain) processPortalConfirmWithdrawInstV3(
	portalStateDB *statedb.StateDB,
	beaconHeight uint64,
	instructions []string,
	currentPortalState *CurrentPortalState,
	portalParams PortalParams) error {
	if currentPortalState == nil {
		Logger.log.Errorf("current portal state is nil")
		return nil
	}

	if len(instructions) != 8 {
		return nil // skip the instruction
	}

	txReqIDStr := instructions[6]
	txReqID, _ := common.Hash{}.NewHashFromStr(txReqIDStr)

	// store withdraw confirm proof
	err := statedb.StoreWithdrawCollateralConfirmProof(portalStateDB, *txReqID, beaconHeight+1)
	if err != nil {
		Logger.log.Errorf("ERROR: an error occurred while store custodian withdraw confirm proof: %+v", err)
		return nil
	}
	return nil
}
