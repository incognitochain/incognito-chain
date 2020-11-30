package instructions

import (
	"bytes"
	"encoding/json"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/incognitochain/incognito-chain/portal"
	metadata2 "github.com/incognitochain/incognito-chain/portal/metadata"
	"math/big"
)

func (blockchain *BlockChain) processPortalCustodianDeposit(
	stateDB *statedb.StateDB,
	beaconHeight uint64,
	instructions []string,
	currentPortalState *CurrentPortalState,
	portalParams portal.PortalParams) error {
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

func (blockchain *BlockChain) processPortalCustodianWithdrawRequest(
	portalStateDB *statedb.StateDB,
	beaconHeight uint64,
	instructions []string,
	currentPortalState *CurrentPortalState,
	portalParams portal.PortalParams) error {
	if currentPortalState == nil {
		Logger.log.Errorf("current portal state is nil")
		return nil
	}

	if len(instructions) != 4 {
		return nil // skip the instruction
	}
	// parse instruction
	var reqContent = metadata2.PortalCustodianWithdrawRequestContent{}
	err := json.Unmarshal([]byte(instructions[3]), &reqContent)
	if err != nil {
		Logger.log.Errorf("ERROR: an error occurred while unmarshaling content string of custodian withdraw request instruction: %+v", err)
		return nil
	}

	reqStatus := instructions[2]
	paymentAddress := reqContent.PaymentAddress
	amount := reqContent.Amount
	freeCollateral := reqContent.RemainFreeCollateral
	txHash := reqContent.TxReqID.String()

	switch reqStatus {
	case common.PortalCustodianWithdrawRequestAcceptedChainStatus:
		// update custodian state
		custodianKeyStr := statedb.GenerateCustodianStateObjectKey(paymentAddress).String()
		custodian, ok := currentPortalState.CustodianPoolState[custodianKeyStr]
		if !ok || custodian == nil {
			Logger.log.Errorf("ERROR: Custodian not found ")
			return nil
		}

		//check free collateral
		if amount > custodian.GetFreeCollateral() {
			Logger.log.Errorf("ERROR: Free collateral is not enough to withdraw")
			return nil
		}
		updatedCustodian := UpdateCustodianStateAfterWithdrawCollateral(custodian, common.PRVIDStr, amount)
		currentPortalState.CustodianPoolState[custodianKeyStr] = updatedCustodian

		//store status req into db
		newCustodianWithdrawRequest := metadata2.NewCustodianWithdrawRequestStatus(
			paymentAddress,
			amount,
			common.PortalCustodianWithdrawReqAcceptedStatus,
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
	case common.PortalCustodianWithdrawRequestRejectedChainStatus:
		newCustodianWithdrawRequest := metadata2.NewCustodianWithdrawRequestStatus(
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

func (blockchain *BlockChain) processPortalCustodianWithdrawV3(
	portalStateDB *statedb.StateDB,
	beaconHeight uint64,
	instructions []string,
	currentPortalState *CurrentPortalState,
	portalParams portal.PortalParams) error {
	if currentPortalState == nil {
		Logger.log.Errorf("current portal state is nil")
		return nil
	}

	if len(instructions) != 4 {
		return nil // skip the instruction
	}
	// parse instruction
	var instContent = metadata2.PortalCustodianWithdrawRequestContentV3{}
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
	statusData := metadata2.NewCustodianWithdrawRequestStatusV3(
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

func (blockchain *BlockChain) processPortalCustodianDepositV3(
	stateDB *statedb.StateDB,
	beaconHeight uint64,
	instructions []string,
	currentPortalState *CurrentPortalState,
	portalParams portal.PortalParams) error {
	if currentPortalState == nil {
		Logger.log.Errorf("current portal state is nil")
		return nil
	}
	if len(instructions) != 4 {
		return nil // skip the instruction
	}

	// unmarshal instructions content
	var actionData metadata2.PortalCustodianDepositContentV3
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
		custodianDepositTrackData := metadata2.PortalCustodianDepositStatusV3{
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
		custodianDepositTrackData := metadata2.PortalCustodianDepositStatusV3{
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