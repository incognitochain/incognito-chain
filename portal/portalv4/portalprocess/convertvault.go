package portalprocess

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/incognitochain/incognito-chain/portal/portalv4"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/metadata"
	portalcommonv4 "github.com/incognitochain/incognito-chain/portal/portalv4/common"
)

/* =======
Portal Converting Vault Request Processor V4
======= */

type PortalConvertVaultRequestProcessor struct {
	*PortalInstProcessorV4
}

func (p *PortalConvertVaultRequestProcessor) GetActions() map[byte][][]string {
	return p.Actions
}

func (p *PortalConvertVaultRequestProcessor) PutAction(action []string, shardID byte) {
	_, found := p.Actions[shardID]
	if !found {
		p.Actions[shardID] = [][]string{action}
	} else {
		p.Actions[shardID] = append(p.Actions[shardID], action)
	}
}

func (p *PortalConvertVaultRequestProcessor) PrepareDataForBlockProducer(stateDB *statedb.StateDB, contentStr string) (map[string]interface{}, error) {
	actionContentBytes, err := base64.StdEncoding.DecodeString(contentStr)
	if err != nil {
		Logger.log.Errorf("Converting request: an error occurred while decoding content string of converting vault request action - Error: %v", err)
		return nil, fmt.Errorf("Converting request: an error occurred while decoding content string of converting vault request action - Error: %v", err)
	}

	var actionData metadata.PortalConvertVaultRequestAction
	err = json.Unmarshal(actionContentBytes, &actionData)
	if err != nil {
		Logger.log.Errorf("Converting request: an error occurred while unmarshal converting vault request action - Error: %v", err)
		return nil, fmt.Errorf("Converting request: an error occurred while unmarshal converting vault request action - Error: %v", err)
	}

	proofHash := hashProof(actionData.Meta.ConvertProof, portalcommonv4.PortalConvertVaultChainCode)
	isExistProofTxHash, err := statedb.IsExistsShieldingRequest(stateDB, actionData.Meta.TokenID, proofHash)
	if err != nil {
		Logger.log.Errorf("Converting request: an error occurred while get converting vault request proof from DB - Error: %v", err)
		return nil, fmt.Errorf("Converting request: an error occurred while get converting vault request proof from DB - Error: %v", err)
	}

	optionalData := make(map[string]interface{})
	optionalData["isExistProofTxHash"] = isExistProofTxHash

	return optionalData, nil
}

// beacon build new instruction from instruction received from ShardToBeaconBlock
func buildPortalConvertVaultRequestInstV4(
	tokenID string,
	proofHash string,
	shieldingUTXO []*statedb.UTXO,
	mintingAmt uint64,
	metaType int,
	shardID byte,
	txReqID common.Hash,
	status string,
	errorStr string,
) []string {
	convertingReqContent := metadata.PortalConvertVaultRequestContent{
		TokenID:          tokenID,
		ConvertProofHash: proofHash,
		ConvertingUTXO:   shieldingUTXO,
		ConvertingAmount: mintingAmt,
		TxReqID:          txReqID,
		ShardID:          shardID,
	}
	convertingReqContentBytes, _ := json.Marshal(convertingReqContent)
	return []string{
		strconv.Itoa(metaType),
		strconv.Itoa(int(shardID)),
		status,
		string(convertingReqContentBytes),
		errorStr,
	}
}

func (p *PortalConvertVaultRequestProcessor) BuildNewInsts(
	bc metadata.ChainRetriever,
	contentStr string,
	shardID byte,
	currentPortalState *CurrentPortalStateV4,
	beaconHeight uint64,
	shardHeights map[byte]uint64,
	portalParams portalv4.PortalParams,
	optionalData map[string]interface{},
) ([][]string, error) {
	Logger.log.Infof("Converting Vault Request Producing ...")
	// parse instruction
	actionContentBytes, err := base64.StdEncoding.DecodeString(contentStr)
	if err != nil {
		Logger.log.Errorf("Converting request: an error occurred while decoding content string of portal converting vault request action - Error: %v", err)
		return [][]string{}, nil
	}
	var actionData metadata.PortalConvertVaultRequestAction
	err = json.Unmarshal(actionContentBytes, &actionData)
	if err != nil {
		Logger.log.Errorf("Converting request: an error occurred while unmarshal portal converting vault request action - Error: %v", err)
		return [][]string{}, nil
	}
	meta := actionData.Meta

	rejectInst := buildPortalConvertVaultRequestInstV4(
		meta.TokenID,
		"",
		[]*statedb.UTXO{},
		0,
		meta.Type,
		shardID,
		actionData.TxReqID,
		portalcommonv4.PortalV4RequestRejectedChainStatus,
		"isInvalidProof",
	)

	if currentPortalState == nil {
		Logger.log.Warn("Converting Request: Current Portal state is null.")
		return [][]string{rejectInst}, nil
	}

	portalTokenProcessor := portalParams.PortalTokens[meta.TokenID]
	if portalTokenProcessor == nil {
		Logger.log.Errorf("Converting Request: TokenID is not supported currently on Portal")
		return [][]string{rejectInst}, nil
	}

	// check unique external proof from optionalData which get from statedb
	if optionalData == nil {
		Logger.log.Errorf("Converting Request: optionalData is null")
		return [][]string{rejectInst}, nil
	}
	isExistInStateDB, ok := optionalData["isExistProofTxHash"].(bool)
	if !ok {
		Logger.log.Errorf("Converting Request: optionalData isExistProofTxHash is invalid")
		return [][]string{rejectInst}, nil
	}

	proofHash := hashProof(meta.ConvertProof, portalcommonv4.PortalConvertVaultChainCode)

	// check unique external proof from portal state
	if currentPortalState.IsExistedShieldingExternalTx(meta.TokenID, proofHash) || isExistInStateDB {
		rejectInst[4] = "IsExistedProof"
		Logger.log.Errorf("Converting Request: Shielding request proof exist in db %v", meta.ConvertProof)
		return [][]string{rejectInst}, nil
	}

	// generate expected multisig address from master pubkeys and user payment address
	_, expectedReceivedMultisigAddress, err := portalTokenProcessor.GenerateOTMultisigAddress(
		portalParams.MasterPubKeys[meta.TokenID], int(portalParams.NumRequiredSigs), portalcommonv4.PortalConvertVaultChainCode)
	if err != nil {
		Logger.log.Error("Converting Request: Could not generate multisig address - Error: %v", err)
		return [][]string{rejectInst}, nil
	}

	// verify shielding proof
	isValid, listUTXO, err := portalTokenProcessor.ParseAndVerifyShieldProof(
		meta.ConvertProof, bc, expectedReceivedMultisigAddress, portalcommonv4.PortalConvertVaultChainCode, 0)
	if !isValid || err != nil {
		Logger.log.Error("Converting Request: Parse proof and verify converting proof failed - Error: %v", err)
		return [][]string{rejectInst}, nil
	}

	// calculate and verify converting amount
	convertingAmountInExtAmt := uint64(0)
	for _, utxo := range listUTXO {
		convertingAmountInExtAmt += utxo.GetOutputAmount()
	}
	convertingAmount := portalTokenProcessor.ConvertExternalToIncAmount(convertingAmountInExtAmt)

	// update portal state
	// add utxos for portal v4 and shielding external tx
	currentPortalState.AddUTXOs(listUTXO, meta.TokenID)
	currentPortalState.AddShieldingExternalTx(meta.TokenID, proofHash,
		listUTXO[0].GetTxHash(), portalcommonv4.PortalConvertVaultChainCode, convertingAmountInExtAmt)

	inst := buildPortalConvertVaultRequestInstV4(
		actionData.Meta.TokenID,
		proofHash,
		listUTXO,
		convertingAmount,
		actionData.Meta.Type,
		shardID,
		actionData.TxReqID,
		portalcommonv4.PortalV4RequestAcceptedChainStatus,
		"None",
	)
	return [][]string{inst}, nil
}

func (p *PortalConvertVaultRequestProcessor) ProcessInsts(
	stateDB *statedb.StateDB,
	beaconHeight uint64,
	instructions []string,
	currentPortalState *CurrentPortalStateV4,
	portalParams portalv4.PortalParams,
	updatingInfoByTokenID map[common.Hash]metadata.UpdatingInfo,
) error {
	Logger.log.Infof("Converting Vault Request Processing ...")
	if currentPortalState == nil {
		Logger.log.Errorf("Converting Request: Current Portal state is nil")
		return nil
	}

	if len(instructions) != 5 {
		Logger.log.Errorf("Converting Request: Instructions are in wrong format")
		return nil // skip the instruction
	}

	// unmarshal instructions content
	var actionData metadata.PortalConvertVaultRequestContent
	err := json.Unmarshal([]byte(instructions[3]), &actionData)
	if err != nil {
		Logger.log.Errorf("Converting Request: Could not unmarshal instruction content %v - Error: %v\n", instructions[3], err)
		return nil
	}

	var shieldStatus byte
	reqStatus := instructions[2]
	if reqStatus == portalcommonv4.PortalV4RequestAcceptedChainStatus {
		shieldStatus = portalcommonv4.PortalV4RequestAcceptedStatus

		convertingAmountInExtAmt := portalParams.PortalTokens[actionData.TokenID].ConvertIncToExternalAmount(actionData.ConvertingAmount)
		currentPortalState.AddUTXOs(actionData.ConvertingUTXO, actionData.TokenID)
		currentPortalState.AddShieldingExternalTx(actionData.TokenID, actionData.ConvertProofHash,
			actionData.ConvertingUTXO[0].GetTxHash(), portalcommonv4.PortalConvertVaultChainCode, convertingAmountInExtAmt)
	} else if reqStatus == portalcommonv4.PortalV4RequestRejectedChainStatus {
		shieldStatus = portalcommonv4.PortalV4RequestRejectedStatus
	}

	// track shieldingReq status by txID into DB
	shieldingReqTrackData := metadata.PortalConvertVaultRequestStatus{
		Status:           shieldStatus,
		ErrorMsg:         instructions[4],
		TokenID:          actionData.TokenID,
		ConvertProofHash: actionData.ConvertProofHash,
		ConvertingUTXO:   actionData.ConvertingUTXO,
		ConvertingAmount: actionData.ConvertingAmount,
		TxReqID:          actionData.TxReqID,
	}
	shieldingReqTrackDataBytes, _ := json.Marshal(shieldingReqTrackData)
	err = statedb.StorePortalConvertVaultRequestStatus(
		stateDB,
		actionData.TxReqID.String(),
		shieldingReqTrackDataBytes,
	)
	if err != nil {
		Logger.log.Errorf("Converting Request: An error occurred while tracking convert vault request tx - Error: %v", err)
	}

	return nil
}
