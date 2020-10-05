package blockchain

import (
	"encoding/json"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/metadata"
)

func (blockchain *BlockChain) processPortalRedeemRequest(
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
	var actionData metadata.PortalRedeemRequestContent
	err := json.Unmarshal([]byte(instructions[3]), &actionData)
	if err != nil {
		Logger.log.Errorf("Can not unmarshal instruction content %v - Error %v\n", instructions[3], err)
		return nil
	}

	reqStatus := instructions[2]
	if reqStatus == common.PortalRedeemRequestAcceptedChainStatus {
		// add waiting redeem request into waiting redeems list
		keyWaitingRedeemRequest := statedb.GenerateWaitingRedeemRequestObjectKey(actionData.UniqueRedeemID)
		keyWaitingRedeemRequestStr := keyWaitingRedeemRequest.String()
		redeemRequest := statedb.NewRedeemRequestWithValue(
			actionData.UniqueRedeemID,
			actionData.TokenID,
			actionData.RedeemerIncAddressStr,
			actionData.RemoteAddress,
			actionData.RedeemAmount,
			actionData.MatchingCustodianDetail,
			actionData.RedeemFee,
			beaconHeight+1,
			actionData.TxReqID,
		)
		currentPortalState.WaitingRedeemRequests[keyWaitingRedeemRequestStr] = redeemRequest

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
			Status:                  common.PortalRedeemRequestTxAcceptedStatus,
			UniqueRedeemID:          actionData.UniqueRedeemID,
			TokenID:                 actionData.TokenID,
			RedeemAmount:            actionData.RedeemAmount,
			RedeemerIncAddressStr:   actionData.RedeemerIncAddressStr,
			RemoteAddress:           actionData.RemoteAddress,
			RedeemFee:               actionData.RedeemFee,
			MatchingCustodianDetail: actionData.MatchingCustodianDetail,
			TxReqID:                 actionData.TxReqID,
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
			Status:                  common.PortalRedeemRequestTxRejectedStatus,
			UniqueRedeemID:          actionData.UniqueRedeemID,
			TokenID:                 actionData.TokenID,
			RedeemAmount:            actionData.RedeemAmount,
			RedeemerIncAddressStr:   actionData.RedeemerIncAddressStr,
			RemoteAddress:           actionData.RemoteAddress,
			RedeemFee:               actionData.RedeemFee,
			MatchingCustodianDetail: actionData.MatchingCustodianDetail,
			TxReqID:                 actionData.TxReqID,
		}
		redeemRequestByTxIDStatusBytes, _ := json.Marshal(redeemRequestByTxIDStatus)
		err = statedb.StorePortalRedeemRequestByTxIDStatus(
			stateDB, actionData.TxReqID.String(), redeemRequestByTxIDStatusBytes)
		if err != nil {
			Logger.log.Errorf("[processPortalRedeemRequest] Error when tracking status of redeem request by txReqID: %v\n", err)
			return nil
		}
	} else if reqStatus == common.PortalRedeemReqCancelledByLiquidationChainStatus {
		keyWaitingRedeemRequest := statedb.GenerateWaitingRedeemRequestObjectKey(actionData.UniqueRedeemID)
		keyWaitingRedeemRequestStr := keyWaitingRedeemRequest.String()
		redeemReq := currentPortalState.WaitingRedeemRequests[keyWaitingRedeemRequestStr]
		if redeemReq == nil {
			Logger.log.Errorf("[processPortalRedeemRequest] redeemReq with ID %v not found: %v\n", actionData.UniqueRedeemID)
			return nil
		}

		// reject waiting redeem request, return ptoken and redeem fee for users
		// update custodian state (return holding public token amount)
		err := UpdateCustodianStateAfterRejectRedeemRequestByLiquidation(currentPortalState, redeemReq, beaconHeight)
		if err != nil {
			Logger.log.Errorf("[processPortalRedeemRequest] Error when updating custodian state %v - RedeemID %v\n: ",
				err, redeemReq.GetUniqueRedeemID())
			return nil
		}

		// remove redeem request from waiting redeem requests list
		deleteWaitingRedeemRequest(currentPortalState, keyWaitingRedeemRequestStr)
		statedb.DeleteWaitingRedeemRequest(stateDB, redeemReq.GetUniqueRedeemID())

		// update status of redeem request by redeemID to rejected by liquidation
		redeemRequestStatus := metadata.PortalRedeemRequestStatus{
			Status:                  common.PortalRedeemReqCancelledByLiquidationStatus,
			UniqueRedeemID:          actionData.UniqueRedeemID,
			TokenID:                 actionData.TokenID,
			RedeemAmount:            actionData.RedeemAmount,
			RedeemerIncAddressStr:   actionData.RedeemerIncAddressStr,
			RemoteAddress:           actionData.RemoteAddress,
			RedeemFee:               actionData.RedeemFee,
			MatchingCustodianDetail: actionData.MatchingCustodianDetail,
			TxReqID:                 redeemReq.GetTxReqID(),
		}
		redeemRequestStatusBytes, _ := json.Marshal(redeemRequestStatus)
		err = statedb.StorePortalRedeemRequestStatus(
			stateDB,
			actionData.UniqueRedeemID,
			redeemRequestStatusBytes)
		if err != nil {
			Logger.log.Errorf("[processPortalRedeemRequest] Error when storing status of redeem request by redeemID: %v\n", err)
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
			updatingInfo.countUpAmt += actionData.RedeemAmount
		} else {
			updatingInfo = UpdatingInfo{
				countUpAmt:      actionData.RedeemAmount,
				deductAmt:       0,
				tokenID:         *incTokenID,
				externalTokenID: nil,
				isCentralized:   false,
			}
		}
		updatingInfoByTokenID[*incTokenID] = updatingInfo
	}

	return nil
}

func (blockchain *BlockChain) processPortalReqMatchingRedeem(
	stateDB *statedb.StateDB,
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

	// unmarshal instructions content
	var actionData metadata.PortalReqMatchingRedeemContent
	err := json.Unmarshal([]byte(instructions[3]), &actionData)
	if err != nil {
		Logger.log.Errorf("Can not unmarshal instruction content %v - Error %v\n", instructions[3], err)
		return nil
	}

	reqStatus := instructions[2]
	if reqStatus == common.PortalReqMatchingRedeemAcceptedChainStatus {
		updatedRedeemRequest, err := UpdatePortalStateAfterCustodianReqMatchingRedeem(
			actionData.CustodianAddressStr,
			actionData.RedeemID,
			actionData.MatchingAmount,
			actionData.IsFullCustodian,
			currentPortalState)
		if err != nil {
			Logger.log.Errorf("Error when updating portal state of request matching redeem request %v", err)
			return nil
		}

		newStatus := common.PortalRedeemReqWaitingStatus
		if actionData.IsFullCustodian {
			statedb.DeleteWaitingRedeemRequest(stateDB, actionData.RedeemID)
			newStatus = common.PortalRedeemReqMatchedStatus
		}

		// update status of redeem ID by redeemID and matching custodians
		redeemRequest := metadata.PortalRedeemRequestStatus{
			Status:                  byte(newStatus),
			UniqueRedeemID:          updatedRedeemRequest.GetUniqueRedeemID(),
			TokenID:                 updatedRedeemRequest.GetTokenID(),
			RedeemAmount:            updatedRedeemRequest.GetRedeemAmount(),
			RedeemerIncAddressStr:   updatedRedeemRequest.GetRedeemerAddress(),
			RemoteAddress:           updatedRedeemRequest.GetRedeemerRemoteAddress(),
			RedeemFee:               updatedRedeemRequest.GetRedeemFee(),
			MatchingCustodianDetail: updatedRedeemRequest.GetCustodians(),
			TxReqID:                 updatedRedeemRequest.GetTxReqID(),
		}
		newRedeemRequest, err := json.Marshal(redeemRequest)
		if err != nil {
			Logger.log.Errorf("Error when marshaling status of redeem request %v", err)
			return nil
		}
		err = statedb.StorePortalRedeemRequestStatus(stateDB, actionData.RedeemID, newRedeemRequest)
		if err != nil {
			Logger.log.Errorf("Error when storing status of redeem request %v", err)
			return err
		}

		// track status of req matching redeem request by txReqID
		redeemRequestByTxIDStatus := metadata.PortalReqMatchingRedeemStatus{
			CustodianAddressStr: actionData.CustodianAddressStr,
			RedeemID:            actionData.RedeemID,
			MatchingAmount:      actionData.MatchingAmount,
			Status:              common.PortalReqMatchingRedeemAcceptedStatus,
		}
		redeemRequestByTxIDStatusBytes, _ := json.Marshal(redeemRequestByTxIDStatus)
		err = statedb.StorePortalReqMatchingRedeemByTxIDStatus(
			stateDB, actionData.TxReqID.String(), redeemRequestByTxIDStatusBytes)
		if err != nil {
			Logger.log.Errorf("[processPortalReqMatchingRedeem] Error when tracking status of redeem request by txReqID: %v\n", err)
			return nil
		}

	} else if reqStatus == common.PortalRedeemRequestRejectedChainStatus {
		// track status of req matching redeem request by txReqID
		redeemRequestByTxIDStatus := metadata.PortalReqMatchingRedeemStatus{
			CustodianAddressStr: actionData.CustodianAddressStr,
			RedeemID:            actionData.RedeemID,
			MatchingAmount:      actionData.MatchingAmount,
			Status:              common.PortalReqMatchingRedeemRejectedStatus,
		}
		redeemRequestByTxIDStatusBytes, _ := json.Marshal(redeemRequestByTxIDStatus)
		err = statedb.StorePortalReqMatchingRedeemByTxIDStatus(
			stateDB, actionData.TxReqID.String(), redeemRequestByTxIDStatusBytes)
		if err != nil {
			Logger.log.Errorf("[processPortalReqMatchingRedeem] Error when tracking status of redeem request by txReqID: %v\n", err)
			return nil
		}

	}
	return nil
}

func (blockchain *BlockChain) processPortalPickMoreCustodiansForTimeOutWaitingRedeemReq(
	stateDB *statedb.StateDB,
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

	// unmarshal instructions content
	var actionData PortalPickMoreCustodiansForRedeemReqContent
	err := json.Unmarshal([]byte(instructions[3]), &actionData)
	if err != nil {
		Logger.log.Errorf("Can not unmarshal instruction content %v - Error %v\n", instructions[3], err)
		return nil
	}

	reqStatus := instructions[2]
	if reqStatus == common.PortalPickMoreCustodianRedeemSuccessChainStatus {
		waitingRedeemKey := statedb.GenerateWaitingRedeemRequestObjectKey(actionData.RedeemID).String()
		waitingRedeem := currentPortalState.WaitingRedeemRequests[waitingRedeemKey]
		updatedRedeemRequest, err := UpdatePortalStateAfterPickMoreCustodiansForWaitingRedeemReq(
			actionData.Custodians,
			waitingRedeem,
			currentPortalState)
		if err != nil {
			Logger.log.Errorf("Error when updating portal state of request matching redeem request %v", err)
			return nil
		}
		// delete waiting redeem request
		statedb.DeleteWaitingRedeemRequest(stateDB, actionData.RedeemID)

		// update status of redeem ID by redeemID and matching custodians
		newStatus := common.PortalRedeemReqMatchedStatus
		redeemRequest := metadata.PortalRedeemRequestStatus{
			Status:                  byte(newStatus),
			UniqueRedeemID:          updatedRedeemRequest.GetUniqueRedeemID(),
			TokenID:                 updatedRedeemRequest.GetTokenID(),
			RedeemAmount:            updatedRedeemRequest.GetRedeemAmount(),
			RedeemerIncAddressStr:   updatedRedeemRequest.GetRedeemerAddress(),
			RemoteAddress:           updatedRedeemRequest.GetRedeemerRemoteAddress(),
			RedeemFee:               updatedRedeemRequest.GetRedeemFee(),
			MatchingCustodianDetail: updatedRedeemRequest.GetCustodians(),
			TxReqID:                 updatedRedeemRequest.GetTxReqID(),
		}
		newRedeemRequest, err := json.Marshal(redeemRequest)
		if err != nil {
			Logger.log.Errorf("Error when marshaling status of redeem request %v", err)
			return nil
		}
		err = statedb.StorePortalRedeemRequestStatus(stateDB, actionData.RedeemID, newRedeemRequest)
		if err != nil {
			Logger.log.Errorf("Error when storing status of redeem request %v", err)
			return err
		}
	}
	return nil
}

func (blockchain *BlockChain) processPortalUnlockCollateral(
	stateDB *statedb.StateDB,
	beaconHeight uint64, instructions []string,
	currentPortalState *CurrentPortalState,
	portalParams PortalParams) error {

	// unmarshal instructions content
	var actionData metadata.PortalRequestUnlockCollateralContent
	err := json.Unmarshal([]byte(instructions[3]), &actionData)
	if err != nil {
		Logger.log.Errorf("Can not unmarshal instruction content %v - Error %v\n", instructions[3], err)
		return nil
	}

	// get tokenID from redeemTokenID
	tokenID := actionData.TokenID
	reqStatus := instructions[2]
	if reqStatus == common.PortalReqUnlockCollateralAcceptedChainStatus {
		// update custodian state (FreeCollateral, LockedAmountCollateral)
		custodianStateKey := statedb.GenerateCustodianStateObjectKey(actionData.CustodianAddressStr)
		custodianStateKeyStr := custodianStateKey.String()
		err := updateCustodianStateAfterReqUnlockCollateral(
			currentPortalState.CustodianPoolState[custodianStateKeyStr],
			actionData.UnlockAmount, tokenID, portalParams, currentPortalState)
		if err != nil {
			Logger.log.Errorf("Error when update custodian state", err)
			return nil
		}

		redeemID := actionData.UniqueRedeemID
		keyMatchedRedeemRequest := statedb.GenerateMatchedRedeemRequestObjectKey(redeemID)
		keyMatchedRedeemRequestStr := keyMatchedRedeemRequest.String()

		// update redeem request state in WaitingRedeemRequest (remove custodian from matchingCustodianDetail)
		newCustodians, err := removeCustodianFromMatchingRedeemCustodians(
			currentPortalState.MatchedRedeemRequests[keyMatchedRedeemRequestStr].GetCustodians(), actionData.CustodianAddressStr)
		if err != nil {
			Logger.log.Errorf("ERROR: an error occurred while removing custodian %v from matching custodians", actionData.CustodianAddressStr)
			return nil
		}
		currentPortalState.MatchedRedeemRequests[keyMatchedRedeemRequestStr].SetCustodians(newCustodians)

		// remove redeem request from WaitingRedeemRequest list when all matching custodians return public token to user
		// when list matchingCustodianDetail is empty
		if len(currentPortalState.MatchedRedeemRequests[keyMatchedRedeemRequestStr].GetCustodians()) == 0 {
			deleteMatchedRedeemRequest(currentPortalState, keyMatchedRedeemRequestStr)
			statedb.DeleteMatchedRedeemRequest(stateDB, actionData.UniqueRedeemID)

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
			TxReqID:             actionData.TxReqID,
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
			Status:              common.PortalReqUnlockCollateralRejectedStatus,
			UniqueRedeemID:      actionData.UniqueRedeemID,
			TokenID:             actionData.TokenID,
			CustodianAddressStr: actionData.CustodianAddressStr,
			RedeemAmount:        actionData.RedeemAmount,
			UnlockAmount:        actionData.UnlockAmount,
			RedeemProof:         actionData.RedeemProof,
			TxReqID:             actionData.TxReqID,
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
