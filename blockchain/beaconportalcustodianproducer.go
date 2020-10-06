package blockchain

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/metadata"
	"math/big"
	"strconv"
	"errors"
)

/* =======
Portal Custodian Deposit Collateral (PRV) Processor
======= */

type portalCustodianDepositProcessor struct {
	*portalInstProcessor
}

func (p *portalCustodianDepositProcessor) getActions() map[byte][][]string {
	return p.actions
}

func (p *portalCustodianDepositProcessor) putAction(action []string, shardID byte) {
	_, found := p.actions[shardID]
	if !found {
		p.actions[shardID] = [][]string{action}
	} else {
		p.actions[shardID] = append(p.actions[shardID], action)
	}
}

func (p *portalCustodianDepositProcessor) prepareDataBeforeProcessing(stateDB *statedb.StateDB, contentStr string) (map[string]interface{}, error) {
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

func (p *portalCustodianDepositProcessor) buildNewInsts(
	bc *BlockChain,
	contentStr string,
	shardID byte,
	currentPortalState *CurrentPortalState,
	beaconHeight uint64,
	portalParams PortalParams,
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

/* =======
Portal Custodian Request Withdraw Free Collaterals Processor
======= */

type portalRequestWithdrawCollateralProcessor struct {
	*portalInstProcessor
}

func (p *portalRequestWithdrawCollateralProcessor) getActions() map[byte][][]string {
	return p.actions
}

func (p *portalRequestWithdrawCollateralProcessor) putAction(action []string, shardID byte) {
	_, found := p.actions[shardID]
	if !found {
		p.actions[shardID] = [][]string{action}
	} else {
		p.actions[shardID] = append(p.actions[shardID], action)
	}
}

func (p *portalRequestWithdrawCollateralProcessor) prepareDataBeforeProcessing(stateDB *statedb.StateDB, contentStr string) (map[string]interface{}, error) {
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

func (p *portalRequestWithdrawCollateralProcessor) buildNewInsts(
	bc *BlockChain,
	contentStr string,
	shardID byte,
	currentPortalState *CurrentPortalState,
	beaconHeight uint64,
	portalParams PortalParams,
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
	if !ok {
		Logger.log.Errorf("Custodian not found")
		return [][]string{rejectInst}, nil
	}

	if actionData.Meta.Amount > custodian.GetFreeCollateral() {
		Logger.log.Errorf("Free Collateral is not enough PRV")
		return [][]string{rejectInst}, nil
	}
	//withdraw
	remainFreeCollateral := custodian.GetFreeCollateral() - actionData.Meta.Amount
	totalFreeCollateral := custodian.GetTotalCollateral() - actionData.Meta.Amount

	inst := buildCustodianWithdrawInst(
		actionData.Meta.Type,
		shardID,
		common.PortalCustodianWithdrawRequestAcceptedChainStatus,
		actionData.Meta.PaymentAddress,
		actionData.Meta.Amount,
		remainFreeCollateral,
		actionData.TxReqID,
	)

	//update free collateral custodian
	custodian.SetFreeCollateral(remainFreeCollateral)
	custodian.SetTotalCollateral(totalFreeCollateral)
	currentPortalState.CustodianPoolState[custodianKeyStr] = custodian
	return [][]string{inst}, nil
}

/* =======
Portal Custodian Deposit Collaterals V3 (ETH and ERC20) Processor
======= */
type portalCustodianDepositProcessorV3 struct {
	*portalInstProcessor
}

func (p *portalCustodianDepositProcessorV3) getActions() map[byte][][]string {
	return p.actions
}

func (p *portalCustodianDepositProcessorV3) putAction(action []string, shardID byte) {
	_, found := p.actions[shardID]
	if !found {
		p.actions[shardID] = [][]string{action}
	} else {
		p.actions[shardID] = append(p.actions[shardID], action)
	}
}

func (p *portalCustodianDepositProcessorV3) prepareDataBeforeProcessing(stateDB *statedb.StateDB, contentStr string) (map[string]interface{}, error) {
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
	uniqExternalTxID := metadata.GetUniqExternalTxID(common.ETHChainName, meta.BlockHash, meta.TxIndex)
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

func (p *portalCustodianDepositProcessorV3) buildNewInsts(
	bc *BlockChain,
	contentStr string,
	shardID byte,
	currentPortalState *CurrentPortalState,
	beaconHeight uint64,
	portalParams PortalParams,
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
	ethReceipt, err := metadata.VerifyProofAndParseReceipt(meta.BlockHash, meta.TxIndex, meta.ProofStrs)
	if err != nil {
		Logger.log.Errorf("Custodian deposit v3: Verify eth proof error: %+v", err)
		return [][]string{rejectedInst}, nil
	}
	if ethReceipt == nil {
		Logger.log.Errorf("The eth proof's receipt could not be null.")
		return [][]string{rejectedInst}, nil
	}

	logMap, err := metadata.PickAndParseLogMapFromReceiptByContractAddr(ethReceipt, bc.GetPortalETHContractAddrStr(), "Deposit")
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
	if ok, err := common.SliceExists(bc.GetSupportedCollateralTokenIDs(beaconHeight), externalTokenIDStr); !ok || err != nil {
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

type portalRequestWithdrawCollateralProcessorV3 struct {
	*portalInstProcessor
}

func (p *portalRequestWithdrawCollateralProcessorV3) getActions() map[byte][][]string {
	return p.actions
}

func (p *portalRequestWithdrawCollateralProcessorV3) putAction(action []string, shardID byte) {
	_, found := p.actions[shardID]
	if !found {
		p.actions[shardID] = [][]string{action}
	} else {
		p.actions[shardID] = append(p.actions[shardID], action)
	}
}

func (p *portalRequestWithdrawCollateralProcessorV3) prepareDataBeforeProcessing(stateDB *statedb.StateDB, contentStr string) (map[string]interface{}, error) {
	return nil, nil
}

// buildCustodianWithdrawCollateralInstV3 builds new instructions to allow custodian withdraw collateral from Portal SC
func buildCustodianWithdrawCollateralInstV3(
	metaType int,
	shardID byte,
	custodianIncAddress string,
	custodianExtAddress string,
	extTokenID string,
	amount uint64,
	txReqID common.Hash,
) []string {
	return []string{
		strconv.Itoa(metaType),
		strconv.Itoa(int(shardID)),
		custodianIncAddress,
		custodianExtAddress,
		extTokenID,
		new(big.Int).SetUint64(amount).String(),
		txReqID.String(),
	}
}

func buildPortalCustodianWithdrawStatusFromInstV3(
	inst []string,
) (*metadata.CustodianWithdrawRequestStatusV3, error) {
	if len(inst) != 7 {
		return nil, errors.New("Portal custodian withdraw confirm instruction should have len = 7")
	}

	metaType := inst[0]
	paymentAddress := inst[2]
	externalAddress := inst[3]
	externalTokenID:= inst[4]
	amountStr := inst[5]
	amount, _ := strconv.ParseUint(amountStr, 10, 64)
	txIDStr := inst[6]
	txID, _ := common.Hash{}.NewHashFromStr(txIDStr)

	statusType := common.PortalCustodianWithdrawReqV3RejectStatus
	if metaType == strconv.Itoa(metadata.PortalCustodianWithdrawConfirmMetaV3) {
		statusType = common.PortalCustodianWithdrawReqV3AcceptedStatus
	}

	status := metadata.NewCustodianWithdrawRequestStatusV3(
		paymentAddress,
		externalAddress,
		externalTokenID,
		amount,
		*txID,
		statusType,
	)
	return status, nil
}
func (p *portalRequestWithdrawCollateralProcessorV3) buildNewInsts(
	bc *BlockChain,
	contentStr string,
	shardID byte,
	currentPortalState *CurrentPortalState,
	beaconHeight uint64,
	portalParams PortalParams,
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

	rejectInst := buildCustodianWithdrawCollateralInstV3(
		actionData.Meta.Type,
		shardID,
		actionData.Meta.CustodianIncAddress,
		actionData.Meta.CustodianExternalAddress,
		actionData.Meta.ExternalTokenID,
		actionData.Meta.Amount,
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

	freeTokenCollaterals := custodian.GetFreeTokenCollaterals()
	externalTokenID := actionData.Meta.ExternalTokenID
	if freeTokenCollaterals == nil || freeTokenCollaterals[externalTokenID] == 0 {
		Logger.log.Errorf("Custodian has no free token collaterals")
		return [][]string{rejectInst}, nil
	}

	if actionData.Meta.Amount > freeTokenCollaterals[externalTokenID] {
		Logger.log.Errorf("Amount request withdraw greater than available free token collaterals")
		return [][]string{rejectInst}, nil
	}

	inst := buildCustodianWithdrawCollateralInstV3(
		metadata.PortalCustodianWithdrawConfirmMetaV3,
		shardID,
		actionData.Meta.CustodianIncAddress,
		actionData.Meta.CustodianExternalAddress,
		actionData.Meta.ExternalTokenID,
		actionData.Meta.Amount,
		actionData.TxReqID,
	)

	// update custodian state
	newCustodian := UpdateCustodianStateAfterWithdrawCollateral(custodian, externalTokenID, actionData.Meta.Amount)
	currentPortalState.CustodianPoolState[custodianKeyStr] = newCustodian
	return [][]string{inst}, nil
}