package blockchain

import (
	"encoding/json"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/metadata"
)

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

