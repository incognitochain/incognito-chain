package portalprocess

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/incognitochain/incognito-chain/basemeta"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/common/base58"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/portal"
	pCommon "github.com/incognitochain/incognito-chain/portal/common"
	portalMeta "github.com/incognitochain/incognito-chain/portal/metadata"
	"math/big"
	"sort"
	"strconv"
)

/* =======
Portal Custodian Deposit Collateral (PRV) Processor
======= */

type portalCustodianDepositProcessor struct {
	*portalInstProcessor
}

func (p *portalCustodianDepositProcessor) GetActions() map[byte][][]string {
	return p.actions
}

func (p *portalCustodianDepositProcessor) PutAction(action []string, shardID byte) {
	_, found := p.actions[shardID]
	if !found {
		p.actions[shardID] = [][]string{action}
	} else {
		p.actions[shardID] = append(p.actions[shardID], action)
	}
}

func (p *portalCustodianDepositProcessor) PrepareDataForBlockProducer(stateDB *statedb.StateDB, contentStr string) (map[string]interface{}, error) {
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
	custodianDepositContent := portalMeta.PortalCustodianDepositContent{
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

func (p *portalCustodianDepositProcessor) BuildNewInsts(
	bc basemeta.ChainRetriever,
	contentStr string,
	shardID byte,
	currentPortalState *CurrentPortalState,
	beaconHeight uint64,
	shardHeights map[byte]uint64,
	portalParams portal.PortalParams,
	optionalData map[string]interface{},
) ([][]string, error) {
	// parse instruction
	actionContentBytes, err := base64.StdEncoding.DecodeString(contentStr)
	if err != nil {
		Logger.log.Errorf("ERROR: an error occured while decoding content string of portal custodian deposit action: %+v", err)
		return [][]string{}, nil
	}
	var actionData portalMeta.PortalCustodianDepositAction
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
			common.PortalCustodianDepositRefundChainStatus,
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
		common.PortalCustodianDepositAcceptedChainStatus,
	)
	return [][]string{inst}, nil
}

func (p *portalCustodianDepositProcessor) ProcessInsts(
	stateDB *statedb.StateDB,
	beaconHeight uint64,
	instructions []string,
	currentPortalState *CurrentPortalState,
	portalParams portal.PortalParams,
	updatingInfoByTokenID map[common.Hash]basemeta.UpdatingInfo,
) error {
	if currentPortalState == nil {
		Logger.log.Errorf("current portal state is nil")
		return nil
	}
	if len(instructions) != 4 {
		return nil // skip the instruction
	}

	// unmarshal instructions content
	var actionData portalMeta.PortalCustodianDepositContent
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
		custodianDepositTrackData := portalMeta.PortalCustodianDepositStatus{
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
		custodianDepositTrackData := portalMeta.PortalCustodianDepositStatus{
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


/* =======
Portal Custodian Request Withdraw Free Collaterals Processor
======= */

type portalRequestWithdrawCollateralProcessor struct {
	*portalInstProcessor
}

func (p *portalRequestWithdrawCollateralProcessor) GetActions() map[byte][][]string {
	return p.actions
}

func (p *portalRequestWithdrawCollateralProcessor) PutAction(action []string, shardID byte) {
	_, found := p.actions[shardID]
	if !found {
		p.actions[shardID] = [][]string{action}
	} else {
		p.actions[shardID] = append(p.actions[shardID], action)
	}
}

func (p *portalRequestWithdrawCollateralProcessor) PrepareDataForBlockProducer(stateDB *statedb.StateDB, contentStr string) (map[string]interface{}, error) {
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
	content := portalMeta.PortalCustodianWithdrawRequestContent{
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

func (p *portalRequestWithdrawCollateralProcessor) BuildNewInsts(
	bc basemeta.ChainRetriever,
	contentStr string,
	shardID byte,
	currentPortalState *CurrentPortalState,
	beaconHeight uint64,
	shardHeights map[byte]uint64,
	portalParams portal.PortalParams,
	optionalData map[string]interface{},
) ([][]string, error) {
	actionContentBytes, err := base64.StdEncoding.DecodeString(contentStr)
	if err != nil {
		Logger.log.Errorf("Have an error occurred while decoding content string of custodian withdraw request action: %+v", err)
		return [][]string{}, nil
	}

	var actionData portalMeta.PortalCustodianWithdrawRequestAction
	err = json.Unmarshal(actionContentBytes, &actionData)
	if err != nil {
		Logger.log.Errorf("Have an error occurred while unmarshal  custodian withdraw request action: %+v", err)
		return [][]string{}, nil
	}

	rejectInst := buildCustodianWithdrawInst(
		actionData.Meta.Type,
		shardID,
		common.PortalCustodianWithdrawRequestRejectedChainStatus,
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
		common.PortalCustodianWithdrawRequestAcceptedChainStatus,
		actionData.Meta.PaymentAddress,
		actionData.Meta.Amount,
		updatedCustodian.GetFreeCollateral(),
		actionData.TxReqID,
	)

	return [][]string{inst}, nil
}

func (p *portalRequestWithdrawCollateralProcessor) ProcessInsts(
	stateDB *statedb.StateDB,
	beaconHeight uint64,
	instructions []string,
	currentPortalState *CurrentPortalState,
	portalParams portal.PortalParams,
	updatingInfoByTokenID map[common.Hash]basemeta.UpdatingInfo,
) error {
	if currentPortalState == nil {
		Logger.log.Errorf("current portal state is nil")
		return nil
	}

	if len(instructions) != 4 {
		return nil // skip the instruction
	}
	// parse instruction
	var reqContent = portalMeta.PortalCustodianWithdrawRequestContent{}
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
		newCustodianWithdrawRequest := portalMeta.NewCustodianWithdrawRequestStatus(
			paymentAddress,
			amount,
			common.PortalCustodianWithdrawReqAcceptedStatus,
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
	case common.PortalCustodianWithdrawRequestRejectedChainStatus:
		newCustodianWithdrawRequest := portalMeta.NewCustodianWithdrawRequestStatus(
			paymentAddress,
			amount,
			common.PortalCustodianWithdrawReqRejectStatus,
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
type portalCustodianDepositProcessorV3 struct {
	*portalInstProcessor
}

func (p *portalCustodianDepositProcessorV3) GetActions() map[byte][][]string {
	return p.actions
}

func (p *portalCustodianDepositProcessorV3) PutAction(action []string, shardID byte) {
	_, found := p.actions[shardID]
	if !found {
		p.actions[shardID] = [][]string{action}
	} else {
		p.actions[shardID] = append(p.actions[shardID], action)
	}
}

func (p *portalCustodianDepositProcessorV3) PrepareDataForBlockProducer(stateDB *statedb.StateDB, contentStr string) (map[string]interface{}, error) {
	// parse instruction
	actionContentBytes, err := base64.StdEncoding.DecodeString(contentStr)
	if err != nil {
		Logger.log.Errorf("ERROR: an error occured while decoding content string of portal custodian deposit action v3: %+v", err)
		return nil, fmt.Errorf("ERROR: an error occured while decoding content string of portal custodian deposit action v3: %+v", err)
	}
	var actionData portalMeta.PortalCustodianDepositActionV3
	err = json.Unmarshal(actionContentBytes, &actionData)
	if err != nil {
		Logger.log.Errorf("ERROR: an error occured while unmarshal portal custodian deposit action v3: %+v", err)
		return nil, fmt.Errorf("ERROR: an error occured while unmarshal portal custodian deposit action v3: %+v", err)
	}
	meta := actionData.Meta
	// NOTE: since TxHash from constructedReceipt is always '0x0000000000000000000000000000000000000000000000000000000000000000'
	// so must build unique external tx as combination of chain name and block hash and tx index.
	uniqExternalTxID := pCommon.GetUniqExternalTxID(common.ETHChainName, meta.BlockHash, meta.TxIndex)
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
	custodianDepositContent := portalMeta.PortalCustodianDepositContentV3{
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

func (p *portalCustodianDepositProcessorV3) BuildNewInsts(
	bc basemeta.ChainRetriever,
	contentStr string,
	shardID byte,
	currentPortalState *CurrentPortalState,
	beaconHeight uint64,
	shardHeights map[byte]uint64,
	portalParams portal.PortalParams,
	optionalData map[string]interface{},
) ([][]string, error) {
	// parse instruction
	actionContentBytes, err := base64.StdEncoding.DecodeString(contentStr)
	if err != nil {
		Logger.log.Errorf("ERROR: an error occured while decoding content string of portal custodian deposit action v3: %+v", err)
		return [][]string{}, nil
	}
	var actionData portalMeta.PortalCustodianDepositActionV3
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
		common.PortalCustodianDepositV3RejectedChainStatus,
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
	ethReceipt, err := pCommon.VerifyProofAndParseReceipt(meta.BlockHash, meta.TxIndex, meta.ProofStrs)
	if err != nil {
		Logger.log.Errorf("Custodian deposit v3: Verify eth proof error: %+v", err)
		return [][]string{rejectedInst}, nil
	}
	if ethReceipt == nil {
		Logger.log.Errorf("The eth proof's receipt could not be null.")
		return [][]string{rejectedInst}, nil
	}

	logMap, err := pCommon.PickAndParseLogMapFromReceiptByContractAddr(ethReceipt, portalParams.PortalETHContractAddressStr, "Deposit")
	if err != nil {
		Logger.log.Errorf("WARNING: an error occured while parsing log map from receipt: ", err)
		return [][]string{rejectedInst}, nil
	}
	if logMap == nil {
		Logger.log.Errorf("WARNING: could not find log map out from receipt")
		return [][]string{rejectedInst}, nil
	}

	// parse info from log map and validate info
	custodianIncAddr, externalTokenIDStr, depositAmount, err := portalMeta.ParseInfoFromLogMap(logMap)
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
		common.PortalCustodianDepositV3RejectedChainStatus,
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
		common.PortalCustodianDepositV3AcceptedChainStatus,
	)
	return [][]string{inst}, nil
}

func (p *portalCustodianDepositProcessorV3) ProcessInsts(
	stateDB *statedb.StateDB,
	beaconHeight uint64,
	instructions []string,
	currentPortalState *CurrentPortalState,
	portalParams portal.PortalParams,
	updatingInfoByTokenID map[common.Hash]basemeta.UpdatingInfo,
) error {
	if currentPortalState == nil {
		Logger.log.Errorf("current portal state is nil")
		return nil
	}
	if len(instructions) != 4 {
		return nil // skip the instruction
	}

	// unmarshal instructions content
	var actionData portalMeta.PortalCustodianDepositContentV3
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
		custodianDepositTrackData := portalMeta.PortalCustodianDepositStatusV3{
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
		custodianDepositTrackData := portalMeta.PortalCustodianDepositStatusV3{
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

type portalRequestWithdrawCollateralProcessorV3 struct {
	*portalInstProcessor
}

func (p *portalRequestWithdrawCollateralProcessorV3) GetActions() map[byte][][]string {
	return p.actions
}

func (p *portalRequestWithdrawCollateralProcessorV3) PutAction(action []string, shardID byte) {
	_, found := p.actions[shardID]
	if !found {
		p.actions[shardID] = [][]string{action}
	} else {
		p.actions[shardID] = append(p.actions[shardID], action)
	}
}

func (p *portalRequestWithdrawCollateralProcessorV3) PrepareDataForBlockProducer(stateDB *statedb.StateDB, contentStr string) (map[string]interface{}, error) {
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
	custodianWithdrawContent := portalMeta.PortalCustodianWithdrawRequestContentV3{
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

func (p *portalRequestWithdrawCollateralProcessorV3) BuildNewInsts(
	bc basemeta.ChainRetriever,
	contentStr string,
	shardID byte,
	currentPortalState *CurrentPortalState,
	beaconHeight uint64,
	shardHeights map[byte]uint64,
	portalParams portal.PortalParams,
	optionalData map[string]interface{},
) ([][]string, error) {
	actionContentBytes, err := base64.StdEncoding.DecodeString(contentStr)
	if err != nil {
		Logger.log.Errorf("Have an error occurred while decoding content string of custodian withdraw request action: %+v", err)
		return [][]string{}, nil
	}

	var actionData portalMeta.PortalCustodianWithdrawRequestActionV3
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
		common.PortalCustodianWithdrawRequestV3RejectedChainStatus,
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
		common.PortalCustodianWithdrawRequestV3AcceptedChainStatus,
		actionData.Meta.CustodianIncAddress,
		actionData.Meta.CustodianExternalAddress,
		actionData.Meta.ExternalTokenID,
		amount,
		actionData.TxReqID,
	)

	confirmInst := buildConfirmWithdrawCollateralInstV3(
		basemeta.PortalCustodianWithdrawConfirmMetaV3,
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

func (p *portalRequestWithdrawCollateralProcessorV3) ProcessInsts(
	stateDB *statedb.StateDB,
	beaconHeight uint64,
	instructions []string,
	currentPortalState *CurrentPortalState,
	portalParams portal.PortalParams,
	updatingInfoByTokenID map[common.Hash]basemeta.UpdatingInfo,
) error {
	if currentPortalState == nil {
		Logger.log.Errorf("current portal state is nil")
		return nil
	}

	if len(instructions) != 4 {
		return nil // skip the instruction
	}
	// parse instruction
	var instContent = portalMeta.PortalCustodianWithdrawRequestContentV3{}
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
	statusData := portalMeta.NewCustodianWithdrawRequestStatusV3(
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