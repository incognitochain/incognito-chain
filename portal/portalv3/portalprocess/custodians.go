package portalprocess

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"sort"
	"strconv"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/common/base58"
	"github.com/incognitochain/incognito-chain/config"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/metadata"
	metadataBridge "github.com/incognitochain/incognito-chain/metadata/bridge"
	"github.com/incognitochain/incognito-chain/portal/portalv3"
	pCommon "github.com/incognitochain/incognito-chain/portal/portalv3/common"
)

/* =======
Portal Custodian Deposit Collateral (PRV) Processor
======= */

type PortalCustodianDepositProcessor struct {
	*PortalInstProcessorV3
}

func (p *PortalCustodianDepositProcessor) GetActions() map[byte][][]string {
	return p.Actions
}

func (p *PortalCustodianDepositProcessor) PutAction(action []string, shardID byte) {
	_, found := p.Actions[shardID]
	if !found {
		p.Actions[shardID] = [][]string{action}
	} else {
		p.Actions[shardID] = append(p.Actions[shardID], action)
	}
}

func (p *PortalCustodianDepositProcessor) PrepareDataForBlockProducer(stateDB *statedb.StateDB, contentStr string) (map[string]interface{}, error) {
	return nil, nil
}

// beacon build new instruction from instruction received from ShardToBeaconBlock
func buildCustodianDepositInst(
	custodianAddressStr string,
	depositedAmount uint64,
	remoteAddresses map[string]string,
	metaType int,
	shardID byte,
	txReqID common.Hash,
	status string,
) []string {
	custodianDepositContent := metadata.PortalCustodianDepositContent{
		IncogAddressStr: custodianAddressStr,
		RemoteAddresses: remoteAddresses,
		DepositedAmount: depositedAmount,
		TxReqID:         txReqID,
		ShardID:         shardID,
	}
	custodianDepositContentBytes, _ := json.Marshal(custodianDepositContent)
	return []string{
		strconv.Itoa(metaType),
		strconv.Itoa(int(shardID)),
		status,
		string(custodianDepositContentBytes),
	}
}

func (p *PortalCustodianDepositProcessor) BuildNewInsts(
	bc metadata.ChainRetriever,
	contentStr string,
	shardID byte,
	currentPortalState *CurrentPortalState,
	beaconHeight uint64,
	shardHeights map[byte]uint64,
	portalParams portalv3.PortalParams,
	optionalData map[string]interface{},
) ([][]string, error) {
	// parse instruction
	actionContentBytes, err := base64.StdEncoding.DecodeString(contentStr)
	if err != nil {
		Logger.log.Errorf("ERROR: an error occured while decoding content string of portal custodian deposit action: %+v", err)
		return [][]string{}, nil
	}
	var actionData metadata.PortalCustodianDepositAction
	err = json.Unmarshal(actionContentBytes, &actionData)
	if err != nil {
		Logger.log.Errorf("ERROR: an error occured while unmarshal portal custodian deposit action: %+v", err)
		return [][]string{}, nil
	}

	if currentPortalState == nil {
		Logger.log.Errorf("WARN - [buildInstructionsForCustodianDeposit]: Current Portal state is null.")
		// need to refund collateral to custodian
		inst := buildCustodianDepositInst(
			actionData.Meta.IncogAddressStr,
			actionData.Meta.DepositedAmount,
			actionData.Meta.RemoteAddresses,
			actionData.Meta.Type,
			shardID,
			actionData.TxReqID,
			pCommon.PortalRequestRefundChainStatus,
		)
		return [][]string{inst}, nil
	}
	meta := actionData.Meta

	// add custodian to custodian pool
	newCustodian := addCustodianToPool(
		currentPortalState.CustodianPoolState,
		meta.IncogAddressStr,
		meta.DepositedAmount,
		common.PRVIDStr,
		meta.RemoteAddresses)

	keyCustodianStateStr := statedb.GenerateCustodianStateObjectKey(meta.IncogAddressStr).String()
	currentPortalState.CustodianPoolState[keyCustodianStateStr] = newCustodian

	inst := buildCustodianDepositInst(
		actionData.Meta.IncogAddressStr,
		actionData.Meta.DepositedAmount,
		newCustodian.GetRemoteAddresses(),
		actionData.Meta.Type,
		shardID,
		actionData.TxReqID,
		pCommon.PortalRequestAcceptedChainStatus,
	)
	return [][]string{inst}, nil
}

func (p *PortalCustodianDepositProcessor) ProcessInsts(
	stateDB *statedb.StateDB,
	beaconHeight uint64,
	instructions []string,
	currentPortalState *CurrentPortalState,
	portalParams portalv3.PortalParams,
	updatingInfoByTokenID map[common.Hash]metadata.UpdatingInfo,
) error {
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
	if depositStatus == pCommon.PortalRequestAcceptedChainStatus {
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
			Status:          pCommon.PortalRequestAcceptedStatus,
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
	} else if depositStatus == pCommon.PortalRequestRefundChainStatus {
		// store custodian deposit status into DB
		custodianDepositTrackData := metadata.PortalCustodianDepositStatus{
			Status:          pCommon.PortalRequestRejectedStatus,
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

/* =======
Portal Custodian Request Withdraw Free Collaterals Processor
======= */

type PortalRequestWithdrawCollateralProcessor struct {
	*PortalInstProcessorV3
}

func (p *PortalRequestWithdrawCollateralProcessor) GetActions() map[byte][][]string {
	return p.Actions
}

func (p *PortalRequestWithdrawCollateralProcessor) PutAction(action []string, shardID byte) {
	_, found := p.Actions[shardID]
	if !found {
		p.Actions[shardID] = [][]string{action}
	} else {
		p.Actions[shardID] = append(p.Actions[shardID], action)
	}
}

func (p *PortalRequestWithdrawCollateralProcessor) PrepareDataForBlockProducer(stateDB *statedb.StateDB, contentStr string) (map[string]interface{}, error) {
	return nil, nil
}

func buildCustodianWithdrawInst(
	metaType int,
	shardID byte,
	reqStatus string,
	paymentAddress string,
	amount uint64,
	remainFreeCollateral uint64,
	txReqID common.Hash,
) []string {
	content := metadata.PortalCustodianWithdrawRequestContent{
		PaymentAddress:       paymentAddress,
		Amount:               amount,
		RemainFreeCollateral: remainFreeCollateral,
		TxReqID:              txReqID,
		ShardID:              shardID,
	}

	contentBytes, _ := json.Marshal(content)
	return []string{
		strconv.Itoa(metaType),
		strconv.Itoa(int(shardID)),
		reqStatus,
		string(contentBytes),
	}
}

func (p *PortalRequestWithdrawCollateralProcessor) BuildNewInsts(
	bc metadata.ChainRetriever,
	contentStr string,
	shardID byte,
	currentPortalState *CurrentPortalState,
	beaconHeight uint64,
	shardHeights map[byte]uint64,
	portalParams portalv3.PortalParams,
	optionalData map[string]interface{},
) ([][]string, error) {
	actionContentBytes, err := base64.StdEncoding.DecodeString(contentStr)
	if err != nil {
		Logger.log.Errorf("Have an error occurred while decoding content string of custodian withdraw request action: %+v", err)
		return [][]string{}, nil
	}

	var actionData metadata.PortalCustodianWithdrawRequestAction
	err = json.Unmarshal(actionContentBytes, &actionData)
	if err != nil {
		Logger.log.Errorf("Have an error occurred while unmarshal  custodian withdraw request action: %+v", err)
		return [][]string{}, nil
	}

	rejectInst := buildCustodianWithdrawInst(
		actionData.Meta.Type,
		shardID,
		pCommon.PortalRequestRejectedChainStatus,
		actionData.Meta.PaymentAddress,
		actionData.Meta.Amount,
		0,
		actionData.TxReqID,
	)

	if currentPortalState == nil {
		Logger.log.Warn("Current Portal state is null")
		return [][]string{rejectInst}, nil
	}

	if len(currentPortalState.CustodianPoolState) <= 0 {
		Logger.log.Errorf("Custodian state is empty")
		return [][]string{rejectInst}, nil
	}

	custodianKey := statedb.GenerateCustodianStateObjectKey(actionData.Meta.PaymentAddress)
	custodianKeyStr := custodianKey.String()
	custodian, ok := currentPortalState.CustodianPoolState[custodianKeyStr]
	if !ok || custodian == nil {
		Logger.log.Errorf("Custodian not found")
		return [][]string{rejectInst}, nil
	}

	if actionData.Meta.Amount > custodian.GetFreeCollateral() {
		Logger.log.Errorf("Withdraw amount is greater than free collateral amount")
		return [][]string{rejectInst}, nil
	}

	updatedCustodian := UpdateCustodianStateAfterWithdrawCollateral(custodian, common.PRVIDStr, actionData.Meta.Amount)
	currentPortalState.CustodianPoolState[custodianKeyStr] = updatedCustodian

	inst := buildCustodianWithdrawInst(
		actionData.Meta.Type,
		shardID,
		pCommon.PortalRequestAcceptedChainStatus,
		actionData.Meta.PaymentAddress,
		actionData.Meta.Amount,
		updatedCustodian.GetFreeCollateral(),
		actionData.TxReqID,
	)

	return [][]string{inst}, nil
}

func (p *PortalRequestWithdrawCollateralProcessor) ProcessInsts(
	stateDB *statedb.StateDB,
	beaconHeight uint64,
	instructions []string,
	currentPortalState *CurrentPortalState,
	portalParams portalv3.PortalParams,
	updatingInfoByTokenID map[common.Hash]metadata.UpdatingInfo,
) error {
	if currentPortalState == nil {
		Logger.log.Errorf("current portal state is nil")
		return nil
	}

	if len(instructions) != 4 {
		return nil // skip the instruction
	}
	// parse instruction
	var reqContent = metadata.PortalCustodianWithdrawRequestContent{}
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
	case pCommon.PortalRequestAcceptedChainStatus:
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
		newCustodianWithdrawRequest := metadata.NewCustodianWithdrawRequestStatus(
			paymentAddress,
			amount,
			pCommon.PortalRequestAcceptedStatus,
			freeCollateral,
		)
		contentStatusBytes, _ := json.Marshal(newCustodianWithdrawRequest)
		err = statedb.StorePortalCustodianWithdrawCollateralStatus(
			stateDB,
			txHash,
			contentStatusBytes,
		)
		if err != nil {
			Logger.log.Errorf("ERROR: an error occurred while store custodian withdraw item: %+v", err)
			return nil
		}
	case pCommon.PortalRequestRejectedChainStatus:
		newCustodianWithdrawRequest := metadata.NewCustodianWithdrawRequestStatus(
			paymentAddress,
			amount,
			pCommon.PortalRequestRejectedStatus,
			freeCollateral,
		)
		contentStatusBytes, _ := json.Marshal(newCustodianWithdrawRequest)
		err = statedb.StorePortalCustodianWithdrawCollateralStatus(
			stateDB,
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

/* =======
Portal Custodian Deposit Collaterals V3 (ETH and ERC20) Processor
======= */
type PortalCustodianDepositProcessorV3 struct {
	*PortalInstProcessorV3
}

func (p *PortalCustodianDepositProcessorV3) GetActions() map[byte][][]string {
	return p.Actions
}

func (p *PortalCustodianDepositProcessorV3) PutAction(action []string, shardID byte) {
	_, found := p.Actions[shardID]
	if !found {
		p.Actions[shardID] = [][]string{action}
	} else {
		p.Actions[shardID] = append(p.Actions[shardID], action)
	}
}

func (p *PortalCustodianDepositProcessorV3) PrepareDataForBlockProducer(stateDB *statedb.StateDB, contentStr string) (map[string]interface{}, error) {
	// parse instruction
	actionContentBytes, err := base64.StdEncoding.DecodeString(contentStr)
	if err != nil {
		Logger.log.Errorf("ERROR: an error occured while decoding content string of portal custodian deposit action v3: %+v", err)
		return nil, fmt.Errorf("ERROR: an error occured while decoding content string of portal custodian deposit action v3: %+v", err)
	}
	var actionData metadata.PortalCustodianDepositActionV3
	err = json.Unmarshal(actionContentBytes, &actionData)
	if err != nil {
		Logger.log.Errorf("ERROR: an error occured while unmarshal portal custodian deposit action v3: %+v", err)
		return nil, fmt.Errorf("ERROR: an error occured while unmarshal portal custodian deposit action v3: %+v", err)
	}
	meta := actionData.Meta
	// NOTE: since TxHash from constructedReceipt is always '0x0000000000000000000000000000000000000000000000000000000000000000'
	// so must build unique external tx as combination of chain name and block hash and tx index.
	uniqExternalTxID := GetUniqExternalTxID(pCommon.ETHChainName, meta.BlockHash, meta.TxIndex)
	isSubmitted, err := statedb.IsPortalExternalTxHashSubmitted(stateDB, uniqExternalTxID)
	if err != nil {
		Logger.log.Errorf("ERROR: an error occured while checking eth tx submitted: %+v", err)
		return nil, fmt.Errorf("ERROR: an error occured while checking eth tx submitted: %+v", err)
	}

	optionalData := make(map[string]interface{})
	optionalData["isSubmitted"] = isSubmitted
	optionalData["uniqExternalTxID"] = uniqExternalTxID
	return optionalData, nil
}

// beacon build new instruction from instruction received from ShardToBeaconBlock
func buildCustodianDepositInstV3(
	custodianAddressStr string,
	depositedAmount uint64,
	remoteAddresses map[string]string,
	externalTokenID string,
	uniqExternalTxID []byte,
	metaType int,
	shardID byte,
	txReqID common.Hash,
	status string,
) []string {
	custodianDepositContent := metadata.PortalCustodianDepositContentV3{
		IncAddressStr:    custodianAddressStr,
		RemoteAddresses:  remoteAddresses,
		DepositAmount:    depositedAmount,
		ExternalTokenID:  externalTokenID,
		UniqExternalTxID: uniqExternalTxID,
		TxReqID:          txReqID,
		ShardID:          shardID,
	}
	custodianDepositContentBytes, _ := json.Marshal(custodianDepositContent)
	return []string{
		strconv.Itoa(metaType),
		strconv.Itoa(int(shardID)),
		status,
		string(custodianDepositContentBytes),
	}
}

func (p *PortalCustodianDepositProcessorV3) BuildNewInsts(
	bc metadata.ChainRetriever,
	contentStr string,
	shardID byte,
	currentPortalState *CurrentPortalState,
	beaconHeight uint64,
	shardHeights map[byte]uint64,
	portalParams portalv3.PortalParams,
	optionalData map[string]interface{},
) ([][]string, error) {
	// parse instruction
	actionContentBytes, err := base64.StdEncoding.DecodeString(contentStr)
	if err != nil {
		Logger.log.Errorf("ERROR: an error occured while decoding content string of portal custodian deposit action v3: %+v", err)
		return [][]string{}, nil
	}
	var actionData metadata.PortalCustodianDepositActionV3
	err = json.Unmarshal(actionContentBytes, &actionData)
	if err != nil {
		Logger.log.Errorf("ERROR: an error occured while unmarshal portal custodian deposit action v3: %+v", err)
		return [][]string{}, nil
	}

	meta := actionData.Meta

	rejectedInst := buildCustodianDepositInstV3(
		"",
		0,
		meta.RemoteAddresses,
		"",
		[]byte{},
		meta.Type,
		shardID,
		actionData.TxReqID,
		pCommon.PortalRequestRejectedChainStatus,
	)

	// check uniqExternalTxID from optionalData which get from statedb
	if optionalData == nil {
		Logger.log.Errorf("Custodian deposit v3: optionalData is null")
		return [][]string{rejectedInst}, nil
	}
	uniqExternalTxID, ok := optionalData["uniqExternalTxID"].([]byte)
	if !ok || len(uniqExternalTxID) == 0 {
		Logger.log.Errorf("Custodian deposit v3: optionalData uniqExternalTxID is invalid")
		return [][]string{rejectedInst}, nil
	}
	isExist, ok := optionalData["isSubmitted"].(bool)
	if !ok {
		Logger.log.Errorf("Custodian deposit v3: optionalData isSubmitted is invalid")
		return [][]string{rejectedInst}, nil
	}
	if isExist {
		Logger.log.Errorf("Custodian deposit v3: Unique external id exist in db %v", uniqExternalTxID)
		return [][]string{rejectedInst}, nil
	}

	// verify proof and parse receipt
	// Note: currently only support ETH
	ethReceipt, err := metadataBridge.VerifyProofAndParseEVMReceipt(
		meta.BlockHash, meta.TxIndex, meta.ProofStrs,
		config.Param().GethParam.Host,
		metadata.EVMConfirmationBlocks,
		"",
		true,
	)
	if err != nil {
		Logger.log.Errorf("Custodian deposit v3: Verify eth proof error: %+v", err)
		return [][]string{rejectedInst}, nil
	}
	if ethReceipt == nil {
		Logger.log.Errorf("The eth proof's receipt could not be null.")
		return [][]string{rejectedInst}, nil
	}

	logMap, err := metadataBridge.PickAndParseLogMapFromReceiptByContractAddr(ethReceipt, portalParams.PortalETHContractAddressStr, "Deposit")
	if err != nil {
		Logger.log.Errorf("WARNING: an error occured while parsing log map from receipt: ", err)
		return [][]string{rejectedInst}, nil
	}
	if logMap == nil {
		Logger.log.Errorf("WARNING: could not find log map out from receipt")
		return [][]string{rejectedInst}, nil
	}

	// parse info from log map and validate info
	custodianIncAddr, externalTokenIDStr, depositAmount, err := metadata.ParseInfoFromLogMap(logMap)
	if err != nil {
		Logger.log.Errorf("Custodian deposit v3: Error when parsing info from log map : %+v", err)
		return [][]string{rejectedInst}, err
	}
	externalTokenIDStr = common.Remove0xPrefix(externalTokenIDStr)

	rejectedInst2 := buildCustodianDepositInstV3(
		custodianIncAddr,
		depositAmount,
		meta.RemoteAddresses,
		externalTokenIDStr,
		uniqExternalTxID,
		meta.Type,
		shardID,
		actionData.TxReqID,
		pCommon.PortalRequestRejectedChainStatus,
	)

	// check externalTokenID should be one of supported collateral tokenIDs
	if !portalParams.IsSupportedTokenCollateralV3(externalTokenIDStr) {
		Logger.log.Errorf("Custodian deposit v3: external collateral tokenID is not supported on portal %v", externalTokenIDStr)
		return [][]string{rejectedInst2}, nil
	}

	// check depositAmount
	if depositAmount <= 0 {
		Logger.log.Errorf("Custodian deposit v3: depositAmount should be greater than zero %v", depositAmount)
		return [][]string{rejectedInst2}, nil
	}

	if currentPortalState == nil {
		Logger.log.Errorf("Custodian deposit V3: Current Portal state is null.")
		return [][]string{rejectedInst2}, nil
	}

	newCustodian := addCustodianToPool(
		currentPortalState.CustodianPoolState,
		custodianIncAddr,
		depositAmount,
		externalTokenIDStr,
		meta.RemoteAddresses)

	// update state of the custodian
	keyCustodianStateStr := statedb.GenerateCustodianStateObjectKey(custodianIncAddr).String()
	currentPortalState.CustodianPoolState[keyCustodianStateStr] = newCustodian

	inst := buildCustodianDepositInstV3(
		custodianIncAddr,
		depositAmount,
		newCustodian.GetRemoteAddresses(),
		externalTokenIDStr,
		uniqExternalTxID,
		actionData.Meta.Type,
		shardID,
		actionData.TxReqID,
		pCommon.PortalRequestAcceptedChainStatus,
	)
	return [][]string{inst}, nil
}

func (p *PortalCustodianDepositProcessorV3) ProcessInsts(
	stateDB *statedb.StateDB,
	beaconHeight uint64,
	instructions []string,
	currentPortalState *CurrentPortalState,
	portalParams portalv3.PortalParams,
	updatingInfoByTokenID map[common.Hash]metadata.UpdatingInfo,
) error {
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
	if depositStatus == pCommon.PortalRequestAcceptedChainStatus {
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
			Status:           pCommon.PortalRequestAcceptedStatus,
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
	} else if depositStatus == pCommon.PortalRequestRejectedChainStatus {
		// store custodian deposit status into DB
		custodianDepositTrackData := metadata.PortalCustodianDepositStatusV3{
			Status:           pCommon.PortalRequestRejectedStatus,
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

type PortalRequestWithdrawCollateralProcessorV3 struct {
	*PortalInstProcessorV3
}

func (p *PortalRequestWithdrawCollateralProcessorV3) GetActions() map[byte][][]string {
	return p.Actions
}

func (p *PortalRequestWithdrawCollateralProcessorV3) PutAction(action []string, shardID byte) {
	_, found := p.Actions[shardID]
	if !found {
		p.Actions[shardID] = [][]string{action}
	} else {
		p.Actions[shardID] = append(p.Actions[shardID], action)
	}
}

func (p *PortalRequestWithdrawCollateralProcessorV3) PrepareDataForBlockProducer(stateDB *statedb.StateDB, contentStr string) (map[string]interface{}, error) {
	return nil, nil
}

// buildConfirmWithdrawCollateralInstV3 builds new instructions to allow custodians/users withdraw collateral from Portal SC
func buildConfirmWithdrawCollateralInstV3(
	metaType int,
	shardID byte,
	incAddress string,
	extAddress string,
	extCollaterals map[string]*big.Int,
	txReqID common.Hash,
	beaconHeight uint64,
) []string {
	// convert extCollaterals to bytes (include padding)
	// the first byte is len(extCollaterals)
	extCollateralBytes := []byte{}
	tokenIDs := []string{}
	for tokenId := range extCollaterals {
		tokenIDs = append(tokenIDs, tokenId)
	}
	sort.Strings(tokenIDs)
	for _, tokenID := range tokenIDs {
		amount := extCollaterals[tokenID]
		tokenIDBytes, _ := common.DecodeETHAddr(tokenID)
		amountBytes := common.AddPaddingBigInt(amount, common.BigIntSize)
		extCollateralBytes = append(extCollateralBytes, tokenIDBytes...)
		extCollateralBytes = append(extCollateralBytes, amountBytes...)
	}
	extCollateralStrs := base58.Base58Check{}.Encode(extCollateralBytes, common.ZeroByte)
	return []string{
		strconv.Itoa(metaType),
		strconv.Itoa(int(shardID)),
		strconv.Itoa(len(extCollaterals)),
		incAddress,
		extAddress,
		extCollateralStrs,
		txReqID.String(),
		strconv.Itoa(int(beaconHeight)),
	}
}

// buildCustodianWithdrawCollateralInstV3 builds new instructions to allow custodian withdraw collateral from Portal SC
func buildCustodianWithdrawCollateralInstV3(
	metaType int,
	shardID byte,
	status string,
	custodianIncAddress string,
	custodianExtAddress string,
	extTokenID string,
	amount *big.Int,
	txReqID common.Hash,
) []string {
	custodianWithdrawContent := metadata.PortalCustodianWithdrawRequestContentV3{
		CustodianIncAddress:      custodianIncAddress,
		CustodianExternalAddress: custodianExtAddress,
		ExternalTokenID:          extTokenID,
		Amount:                   amount,
		TxReqID:                  txReqID,
		ShardID:                  shardID,
	}
	custodianWithdrawContentBytes, _ := json.Marshal(custodianWithdrawContent)
	return []string{
		strconv.Itoa(metaType),
		strconv.Itoa(int(shardID)),
		status,
		string(custodianWithdrawContentBytes),
	}
}

func (p *PortalRequestWithdrawCollateralProcessorV3) BuildNewInsts(
	bc metadata.ChainRetriever,
	contentStr string,
	shardID byte,
	currentPortalState *CurrentPortalState,
	beaconHeight uint64,
	shardHeights map[byte]uint64,
	portalParams portalv3.PortalParams,
	optionalData map[string]interface{},
) ([][]string, error) {
	actionContentBytes, err := base64.StdEncoding.DecodeString(contentStr)
	if err != nil {
		Logger.log.Errorf("Have an error occurred while decoding content string of custodian withdraw request action: %+v", err)
		return [][]string{}, nil
	}

	var actionData metadata.PortalCustodianWithdrawRequestActionV3
	err = json.Unmarshal(actionContentBytes, &actionData)
	if err != nil {
		Logger.log.Errorf("Have an error occurred while unmarshal custodian withdraw request action v3: %+v", err)
		return [][]string{}, nil
	}

	amount := big.NewInt(0).SetUint64(actionData.Meta.Amount)
	externalTokenID := actionData.Meta.ExternalTokenID
	// Convert amount to big.Int to get bytes later
	if bytes.Equal(common.FromHex(externalTokenID), common.FromHex(common.EthAddrStr)) {
		// Convert Gwei to Wei for Ether
		amount = amount.Mul(amount, big.NewInt(1000000000))
	}
	rejectInst := buildCustodianWithdrawCollateralInstV3(
		actionData.Meta.Type,
		shardID,
		pCommon.PortalRequestRejectedChainStatus,
		actionData.Meta.CustodianIncAddress,
		actionData.Meta.CustodianExternalAddress,
		actionData.Meta.ExternalTokenID,
		amount,
		actionData.TxReqID,
	)

	if currentPortalState == nil {
		Logger.log.Warn("Current Portal state is null")
		return [][]string{rejectInst}, nil
	}

	if len(currentPortalState.CustodianPoolState) <= 0 {
		Logger.log.Errorf("Custodian state is empty")
		return [][]string{rejectInst}, nil
	}

	custodianKey := statedb.GenerateCustodianStateObjectKey(actionData.Meta.CustodianIncAddress)
	custodianKeyStr := custodianKey.String()
	custodian, ok := currentPortalState.CustodianPoolState[custodianKeyStr]
	if !ok {
		Logger.log.Errorf("Custodian not found")
		return [][]string{rejectInst}, nil
	}

	// validate fee token collaterals
	freeTokenCollaterals := custodian.GetFreeTokenCollaterals()
	if freeTokenCollaterals == nil || freeTokenCollaterals[externalTokenID] == 0 {
		Logger.log.Errorf("Custodian has no free token collaterals")
		return [][]string{rejectInst}, nil
	}

	if actionData.Meta.Amount > freeTokenCollaterals[externalTokenID] {
		Logger.log.Errorf("Amount request withdraw greater than available free token collaterals")
		return [][]string{rejectInst}, nil
	}

	acceptedInst := buildCustodianWithdrawCollateralInstV3(
		actionData.Meta.Type,
		shardID,
		pCommon.PortalRequestAcceptedChainStatus,
		actionData.Meta.CustodianIncAddress,
		actionData.Meta.CustodianExternalAddress,
		actionData.Meta.ExternalTokenID,
		amount,
		actionData.TxReqID,
	)

	confirmInst := buildConfirmWithdrawCollateralInstV3(
		metadata.PortalCustodianWithdrawConfirmMetaV3,
		shardID,
		actionData.Meta.CustodianIncAddress,
		actionData.Meta.CustodianExternalAddress,
		map[string]*big.Int{
			externalTokenID: amount,
		},
		actionData.TxReqID,
		beaconHeight+1,
	)

	// update custodian state
	newCustodian := UpdateCustodianStateAfterWithdrawCollateral(custodian, externalTokenID, actionData.Meta.Amount)
	currentPortalState.CustodianPoolState[custodianKeyStr] = newCustodian
	return [][]string{acceptedInst, confirmInst}, nil
}

func (p *PortalRequestWithdrawCollateralProcessorV3) ProcessInsts(
	stateDB *statedb.StateDB,
	beaconHeight uint64,
	instructions []string,
	currentPortalState *CurrentPortalState,
	portalParams portalv3.PortalParams,
	updatingInfoByTokenID map[common.Hash]metadata.UpdatingInfo,
) error {
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
	statusInt := pCommon.PortalRequestRejectedStatus
	if status == pCommon.PortalRequestAcceptedChainStatus {
		statusInt = pCommon.PortalRequestAcceptedStatus

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
		stateDB,
		statusData.TxReqID.String(),
		contentStatusBytes,
	)
	if err != nil {
		Logger.log.Errorf("ERROR: an error occurred while store custodian withdraw v3 item: %+v", err)
		return nil
	}

	return nil
}

/* =======
Portal Custodian unlock over rate collaterals Processor
======= */

func buildReqUnlockOverRateCollateralsInst(
	custodianAddresStr string,
	tokenID string,
	unlockedAmounts map[string]uint64,
	metaType int,
	shardID byte,
	txReqID common.Hash,
	status string,
) []string {
	unlockOverRateCollateralsContent := metadata.PortalUnlockOverRateCollateralsContent{
		CustodianAddressStr: custodianAddresStr,
		TokenID:             tokenID,
		UnlockedAmounts:     unlockedAmounts,
		TxReqID:             txReqID,
	}
	unlockOverRateCollateralsContentBytes, _ := json.Marshal(unlockOverRateCollateralsContent)
	return []string{
		strconv.Itoa(metaType),
		strconv.Itoa(int(shardID)),
		status,
		string(unlockOverRateCollateralsContentBytes),
	}
}

type PortalCusUnlockOverRateCollateralsProcessor struct {
	*PortalInstProcessorV3
}

func (p *PortalCusUnlockOverRateCollateralsProcessor) GetActions() map[byte][][]string {
	return p.Actions
}

func (p *PortalCusUnlockOverRateCollateralsProcessor) PutAction(action []string, shardID byte) {
	_, found := p.Actions[shardID]
	if !found {
		p.Actions[shardID] = [][]string{action}
	} else {
		p.Actions[shardID] = append(p.Actions[shardID], action)
	}
}

func (p *PortalCusUnlockOverRateCollateralsProcessor) PrepareDataForBlockProducer(stateDB *statedb.StateDB, contentStr string) (map[string]interface{}, error) {
	return nil, nil
}

func (p *PortalCusUnlockOverRateCollateralsProcessor) BuildNewInsts(
	bc metadata.ChainRetriever,
	contentStr string,
	shardID byte,
	currentPortalState *CurrentPortalState,
	beaconHeight uint64,
	shardHeights map[byte]uint64,
	portalParams portalv3.PortalParams,
	optionalData map[string]interface{},
) ([][]string, error) {
	actionContentBytes, err := base64.StdEncoding.DecodeString(contentStr)
	if err != nil {
		Logger.log.Errorf("ERROR: an error occurred while decoding content string of portal exchange rates action: %+v", err)
		return [][]string{}, nil
	}

	var actionData metadata.PortalUnlockOverRateCollateralsAction
	err = json.Unmarshal(actionContentBytes, &actionData)
	if err != nil {
		Logger.log.Errorf("ERROR: an error occurred while unmarshal portal exchange rates action: %+v", err)
		return [][]string{}, nil
	}

	metaType := actionData.Meta.Type

	rejectInst := buildReqUnlockOverRateCollateralsInst(
		actionData.Meta.CustodianAddressStr,
		actionData.Meta.TokenID,
		map[string]uint64{},
		metaType,
		shardID,
		actionData.TxReqID,
		pCommon.PortalRequestRejectedChainStatus,
	)
	//check key from db
	exchangeTool := NewPortalExchangeRateTool(currentPortalState.FinalExchangeRatesState, portalParams)
	custodianStateKey := statedb.GenerateCustodianStateObjectKey(actionData.Meta.CustodianAddressStr).String()
	custodianState, ok := currentPortalState.CustodianPoolState[custodianStateKey]
	if !ok || custodianState == nil {
		Logger.log.Error("ERROR: custodian not found")
		return [][]string{rejectInst}, nil
	}
	tokenAmountListInWaitingPoring := GetTotalLockedCollateralAmountInWaitingPortingsV3(currentPortalState, custodianState, actionData.Meta.TokenID)
	if (custodianState.GetLockedTokenCollaterals() == nil || custodianState.GetLockedTokenCollaterals()[actionData.Meta.TokenID] == nil) && custodianState.GetLockedAmountCollateral() == nil {
		Logger.log.Error("ERROR: custodian has no collaterals to unlock")
		return [][]string{rejectInst}, nil
	}
	if custodianState.GetHoldingPublicTokens() == nil || custodianState.GetHoldingPublicTokens()[actionData.Meta.TokenID] == 0 {
		Logger.log.Error("ERROR: custodian has no holding token to unlock")
		return [][]string{rejectInst}, nil
	}
	var lockedCollaterals map[string]uint64
	if custodianState.GetLockedTokenCollaterals() != nil && custodianState.GetLockedTokenCollaterals()[actionData.Meta.TokenID] != nil {
		lockedCollaterals = cloneMap(custodianState.GetLockedTokenCollaterals()[actionData.Meta.TokenID])
	} else {
		lockedCollaterals = make(map[string]uint64, 0)
	}
	if custodianState.GetLockedAmountCollateral() != nil {
		lockedCollaterals[common.PRVIDStr] = custodianState.GetLockedAmountCollateral()[actionData.Meta.TokenID]
	}

	totalAmountInUSD := uint64(0)
	for collateralID, tokenValue := range lockedCollaterals {
		if tokenValue < tokenAmountListInWaitingPoring[collateralID] {
			Logger.log.Errorf("ERROR: total %v locked less than amount lock in porting", collateralID)
			return [][]string{rejectInst}, nil
		}
		lockedCollateralExceptPorting := tokenValue - tokenAmountListInWaitingPoring[collateralID]
		// convert to usd
		pubTokenAmountInUSDT, err := exchangeTool.ConvertToUSD(collateralID, lockedCollateralExceptPorting)
		if err != nil {
			Logger.log.Errorf("Error when converting locked public token to prv: %v", err)
			return [][]string{rejectInst}, nil
		}
		totalAmountInUSD = totalAmountInUSD + pubTokenAmountInUSDT
	}

	// convert holding token to usd
	hodTokenAmountInUSDT, err := exchangeTool.ConvertToUSD(actionData.Meta.TokenID, custodianState.GetHoldingPublicTokens()[actionData.Meta.TokenID])
	if err != nil {
		Logger.log.Errorf("Error when converting holding public token to prv: %v", err)
		return [][]string{rejectInst}, nil
	}
	totalHoldAmountInUSDBigInt := new(big.Int).Mul(new(big.Int).SetUint64(hodTokenAmountInUSDT), new(big.Int).SetUint64(portalParams.MinUnlockOverRateCollaterals))
	minHoldUnlockedAmountInBigInt := new(big.Int).Div(totalHoldAmountInUSDBigInt, big.NewInt(10))
	if minHoldUnlockedAmountInBigInt.Cmp(new(big.Int).SetUint64(totalAmountInUSD)) >= 0 {
		Logger.log.Errorf("Error locked collaterals amount not enough to unlock")
		return [][]string{rejectInst}, nil
	}
	amountToUnlock := big.NewInt(0).Sub(new(big.Int).SetUint64(totalAmountInUSD), minHoldUnlockedAmountInBigInt).Uint64()
	listUnlockTokens, err := updateCustodianStateAfterReqUnlockCollateralV3(custodianState, amountToUnlock, actionData.Meta.TokenID, portalParams, currentPortalState)
	if err != nil || len(listUnlockTokens) == 0 {
		Logger.log.Errorf("Error when updateCustodianStateAfterReqUnlockCollateralV3: %v, %v", err, len(listUnlockTokens))
		return [][]string{rejectInst}, nil
	}

	inst := buildReqUnlockOverRateCollateralsInst(
		actionData.Meta.CustodianAddressStr,
		actionData.Meta.TokenID,
		listUnlockTokens,
		metaType,
		shardID,
		actionData.TxReqID,
		pCommon.PortalRequestAcceptedChainStatus,
	)

	return [][]string{inst}, nil
}

func (p *PortalCusUnlockOverRateCollateralsProcessor) ProcessInsts(
	stateDB *statedb.StateDB,
	beaconHeight uint64,
	instructions []string,
	currentPortalState *CurrentPortalState,
	portalParams portalv3.PortalParams,
	updatingInfoByTokenID map[common.Hash]metadata.UpdatingInfo,
) error {
	if currentPortalState == nil {
		Logger.log.Errorf("current portal state is nil")
		return nil
	}

	// parse instruction
	var unlockOverRateCollateralsContent metadata.PortalUnlockOverRateCollateralsContent
	err := json.Unmarshal([]byte(instructions[3]), &unlockOverRateCollateralsContent)
	if err != nil {
		Logger.log.Errorf("ERROR: an error occurred while unmarshaling content string of portal unlock over rate collaterals instruction: %+v", err)
		return nil
	}

	reqStatus := instructions[2]
	Logger.log.Infof("Portal unlock over rate collaterals, data input: %+v, status: %+v", unlockOverRateCollateralsContent, reqStatus)

	switch reqStatus {
	case pCommon.PortalRequestAcceptedChainStatus:
		custodianStateKey := statedb.GenerateCustodianStateObjectKey(unlockOverRateCollateralsContent.CustodianAddressStr)
		custodianStateKeyStr := custodianStateKey.String()
		listTokensWithValue := cloneMap(unlockOverRateCollateralsContent.UnlockedAmounts)
		unlockPrvAmount := listTokensWithValue[common.PRVIDStr]
		delete(listTokensWithValue, common.PRVIDStr)
		err = updateCustodianStateUnlockOverRateCollaterals(currentPortalState.CustodianPoolState[custodianStateKeyStr], unlockPrvAmount, listTokensWithValue, unlockOverRateCollateralsContent.TokenID)
		if err != nil {
			Logger.log.Errorf("ERROR: an error occurred while updateCustodianStateUnlockOverRateCollaterals: %+v", err)
			return nil
		}

		//save db
		newUnlockOverRateCollaterals := metadata.NewUnlockOverRateCollateralsRequestStatus(
			pCommon.PortalRequestAcceptedStatus,
			unlockOverRateCollateralsContent.CustodianAddressStr,
			unlockOverRateCollateralsContent.TokenID,
			unlockOverRateCollateralsContent.UnlockedAmounts,
		)

		newUnlockOverRateCollateralsStatusBytes, _ := json.Marshal(newUnlockOverRateCollaterals)
		err = statedb.StorePortalUnlockOverRateCollaterals(
			stateDB,
			unlockOverRateCollateralsContent.TxReqID.String(),
			newUnlockOverRateCollateralsStatusBytes,
		)

		if err != nil {
			Logger.log.Errorf("ERROR: Save UnlockOverRateCollaterals error: %+v", err)
			return nil
		}

	case pCommon.PortalRequestRejectedChainStatus:
		//save db
		newUnlockOverRateCollaterals := metadata.NewUnlockOverRateCollateralsRequestStatus(
			pCommon.PortalRequestRejectedStatus,
			unlockOverRateCollateralsContent.CustodianAddressStr,
			unlockOverRateCollateralsContent.TokenID,
			map[string]uint64{},
		)

		newUnlockOverRateCollateralsStatusBytes, _ := json.Marshal(newUnlockOverRateCollaterals)
		err = statedb.StorePortalUnlockOverRateCollaterals(
			stateDB,
			unlockOverRateCollateralsContent.TxReqID.String(),
			newUnlockOverRateCollateralsStatusBytes,
		)
		if err != nil {
			Logger.log.Errorf("ERROR: Save UnlockOverRateCollaterals error: %+v", err)
			return nil
		}
	}

	return nil
}
