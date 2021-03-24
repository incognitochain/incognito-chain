package portalprocess

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/metadata"
	pCommon "github.com/incognitochain/incognito-chain/portal/portalv3/common"
	"github.com/incognitochain/incognito-chain/portal/portalv4"
	portalcommonv4 "github.com/incognitochain/incognito-chain/portal/portalv4/common"
	"github.com/incognitochain/incognito-chain/portal/portalv4/portaltokens"
	"strconv"
)

/* =======
Portal Unshield Request Batching Processor
======= */
type PortalUnshieldBatchingProcessor struct {
	*PortalInstProcessorV4
}

func (p *PortalUnshieldBatchingProcessor) GetActions() map[byte][][]string {
	return p.Actions
}

func (p *PortalUnshieldBatchingProcessor) PutAction(action []string, shardID byte) {
	_, found := p.Actions[shardID]
	if !found {
		p.Actions[shardID] = [][]string{action}
	} else {
		p.Actions[shardID] = append(p.Actions[shardID], action)
	}
}

func (p *PortalUnshieldBatchingProcessor) PrepareDataForBlockProducer(stateDB *statedb.StateDB, contentStr string) (map[string]interface{}, error) {
	return nil, nil
}

// beacon build new instruction from instruction received from ShardToBeaconBlock
func buildUnshieldBatchingInst(
	batchID string,
	rawExtTx string,
	tokenID string,
	unshieldIDs []string,
	utxos map[string][]*statedb.UTXO,
	networkFee map[uint64]uint,
	metaType int,
	status string,
) []string {
	unshieldBatchContent := metadata.PortalUnshieldRequestBatchContent{
		BatchID:       batchID,
		RawExternalTx: rawExtTx,
		TokenID:       tokenID,
		UnshieldIDs:   unshieldIDs,
		UTXOs:         utxos,
		NetworkFee:    networkFee,
	}
	unshieldBatchContentBytes, _ := json.Marshal(unshieldBatchContent)
	return []string{
		strconv.Itoa(metaType),
		strconv.Itoa(-1),
		status,
		string(unshieldBatchContentBytes),
	}
}

// batchID is hash of current beacon height and unshieldIDs that processed
func GetBatchID(beaconHeight uint64, unshieldIDs []string) string {
	dataBytes := []byte(fmt.Sprintf("%d", beaconHeight))
	for _, id := range unshieldIDs {
		dataBytes = append(dataBytes, []byte(id)...)
	}
	dataHash := common.HashH(dataBytes)
	return dataHash.String()
}

func (p *PortalUnshieldBatchingProcessor) BuildNewInsts(
	bc metadata.ChainRetriever,
	contentStr string,
	shardID byte,
	CurrentPortalStateV4 *CurrentPortalStateV4,
	beaconHeight uint64,
	shardHeights map[byte]uint64,
	portalParams portalv4.PortalParams,
	optionalData map[string]interface{},
) ([][]string, error) {
	if CurrentPortalStateV4 == nil {
		Logger.log.Warn("WARN - [Batch Unshield Request]: Current Portal state V4 is null.")
		return [][]string{}, nil
	}

	newInsts := [][]string{}
	wUnshieldRequests := CurrentPortalStateV4.WaitingUnshieldRequests
	for tokenID, wReqs := range wUnshieldRequests {
		portalTokenProcessor := portalParams.PortalTokens[tokenID]
		if portalTokenProcessor == nil {
			Logger.log.Errorf("[Batch Unshield Request]: Portal token ID %v is null.", tokenID)
			continue
		}

		// use default unshield fee in nano ptoken
		feeUnshield := portalParams.DefaultFeeUnshields[tokenID]

		// only process for waiting unshield request has enough minimum confirmation incognito blocks (avoid fork beacon chain)
		wReqForProcess := map[string]*statedb.WaitingUnshieldRequest{}
		for key, wR := range wReqs {
			if wR.GetBeaconHeight()+uint64(portalParams.MinConfirmationIncBlockNum) > beaconHeight {
				continue
			}
			wReqForProcess[key] = statedb.NewWaitingUnshieldRequestStateWithValue(
				wR.GetRemoteAddress(),
				wR.GetAmount(),
				wR.GetUnshieldID(),
				wR.GetBeaconHeight())
		}

		// choose waiting unshield IDs to process with current UTXOs
		utxos := CurrentPortalStateV4.UTXOs[tokenID]
		broadCastTxs := portalTokenProcessor.ChooseUnshieldIDsFromCandidates(utxos, wReqForProcess)

		// create raw external txs
		for _, bcTx := range broadCastTxs {
			// prepare outputs for tx in pbtc amount (haven't paid network fee)
			outputTxs := []*portaltokens.OutputTx{}
			for _, chosenUnshieldID := range bcTx.UnshieldIDs {
				keyWaitingUnshieldRequest := statedb.GenerateWaitingUnshieldRequestObjectKey(tokenID, chosenUnshieldID).String()
				wUnshieldReq := wUnshieldRequests[tokenID][keyWaitingUnshieldRequest]
				outputTxs = append(outputTxs, &portaltokens.OutputTx{
					ReceiverAddress: wUnshieldReq.GetRemoteAddress(),
					Amount:          wUnshieldReq.GetAmount(),
				})
			}

			// memo in tx: batchId: combine beacon height and list of unshieldIDs
			batchID := GetBatchID(beaconHeight+1, bcTx.UnshieldIDs)
			memo := batchID

			// create raw tx
			hexRawExtTxStr, _, err := portalTokenProcessor.CreateRawExternalTx(
				bcTx.UTXOs, outputTxs, feeUnshield, memo, bc)
			if err != nil {
				Logger.log.Errorf("[Batch Unshield Request]: Error when creating raw external tx %v", err)
				continue
			}

			externalFees := map[uint64]uint{
				beaconHeight + 1: uint(feeUnshield),
			}
			chosenUTXOs := map[string][]*statedb.UTXO{
				portalParams.MultiSigAddresses[tokenID]: bcTx.UTXOs,
			}
			// update current portal state
			// remove chosen waiting unshield requests from waiting list
			// remove utxos
			UpdatePortalStateAfterProcessBatchUnshieldRequest(
				CurrentPortalStateV4, batchID, chosenUTXOs, externalFees, bcTx.UnshieldIDs, tokenID)

			// build new instruction with new raw external tx
			newInst := buildUnshieldBatchingInst(
				batchID, hexRawExtTxStr, tokenID, bcTx.UnshieldIDs, chosenUTXOs, externalFees,
				metadata.PortalV4UnshieldBatchingMeta, portalcommonv4.PortalV4RequestAcceptedChainStatus)
			newInsts = append(newInsts, newInst)
		}
	}
	return newInsts, nil
}

func (p *PortalUnshieldBatchingProcessor) ProcessInsts(
	stateDB *statedb.StateDB,
	beaconHeight uint64,
	instructions []string,
	CurrentPortalStateV4 *CurrentPortalStateV4,
	portalParams portalv4.PortalParams,
	updatingInfoByTokenID map[common.Hash]metadata.UpdatingInfo,
) error {
	if CurrentPortalStateV4 == nil {
		Logger.log.Errorf("current portal state is nil")
		return nil
	}

	if len(instructions) != 4 {
		return nil // skip the instruction
	}

	// unmarshal instructions content
	var actionData metadata.PortalUnshieldRequestBatchContent
	err := json.Unmarshal([]byte(instructions[3]), &actionData)
	if err != nil {
		Logger.log.Errorf("Can not unmarshal instruction content %v - Error %v\n", instructions[3], err)
		return nil
	}

	reqStatus := instructions[2]
	if reqStatus == portalcommonv4.PortalV4RequestAcceptedChainStatus {
		// add new processed batch unshield request to batch unshield list
		// remove waiting unshield request from waiting list
		UpdatePortalStateAfterProcessBatchUnshieldRequest(
			CurrentPortalStateV4, actionData.BatchID, actionData.UTXOs, actionData.NetworkFee, actionData.UnshieldIDs, actionData.TokenID)
		RemoveListUtxoFromDB(stateDB, actionData.UTXOs, actionData.TokenID)
		RemoveListWaitingUnshieldFromDB(stateDB, actionData.UnshieldIDs, actionData.TokenID)

		for _, unshieldID := range actionData.UnshieldIDs {
			// update status of unshield request that processed
			err := UpdateNewStatusUnshieldRequest(unshieldID, portalcommonv4.PortalUnshieldReqProcessedStatus, stateDB)
			if err != nil {
				Logger.log.Errorf("[processPortalBatchUnshieldRequest] Error when updating status of unshielding request with unshieldID %v: %v\n", unshieldID, err)
				return nil
			}
		}

		// store status of batch unshield by batchID
		batchUnshieldRequestStatus := metadata.PortalUnshieldRequestBatchStatus{
			BatchID:       actionData.BatchID,
			RawExternalTx: actionData.RawExternalTx,
			BeaconHeight:  beaconHeight + 1,
			TokenID:       actionData.TokenID,
			UnshieldIDs:   actionData.UnshieldIDs,
			UTXOs:         actionData.UTXOs,
			NetworkFee:    actionData.NetworkFee,
			Status:        portalcommonv4.PortalBatchUnshieldReqProcessedStatus,
		}
		batchUnshieldRequestStatusBytes, _ := json.Marshal(batchUnshieldRequestStatus)
		err := statedb.StorePortalBatchUnshieldRequestStatus(
			stateDB,
			actionData.BatchID,
			batchUnshieldRequestStatusBytes)
		if err != nil {
			Logger.log.Errorf("[processPortalBatchUnshieldRequest] Error when storing status of redeem request by redeemID: %v\n", err)
			return nil
		}
	}

	return nil
}

/* =======
Portal Replacement Processor
======= */

type PortalFeeReplacementRequestProcessor struct {
	*PortalInstProcessorV4
}

func (p *PortalFeeReplacementRequestProcessor) GetActions() map[byte][][]string {
	return p.Actions
}

func (p *PortalFeeReplacementRequestProcessor) PutAction(action []string, shardID byte) {
	_, found := p.Actions[shardID]
	if !found {
		p.Actions[shardID] = [][]string{action}
	} else {
		p.Actions[shardID] = append(p.Actions[shardID], action)
	}
}

func (p *PortalFeeReplacementRequestProcessor) PrepareDataForBlockProducer(stateDB *statedb.StateDB, contentStr string) (map[string]interface{}, error) {
	actionContentBytes, err := base64.StdEncoding.DecodeString(contentStr)
	if err != nil {
		Logger.log.Errorf("Replace fee request: an error occurred while decoding content string of replace fee unshield request action: %+v", err)
		return nil, fmt.Errorf("Replace fee request: an error occurred while decoding content string of replace fee unshield request action: %+v", err)
	}

	var actionData metadata.PortalReplacementFeeRequestAction
	err = json.Unmarshal(actionContentBytes, &actionData)
	if err != nil {
		Logger.log.Errorf("Replace fee: an error occurred while unmarshal replace fee unshield request action: %+v", err)
		return nil, fmt.Errorf("Replace fee: an error occurred while unmarshal replace fee unshield request action: %+v", err)
	}

	unshieldBatchBytes, err := statedb.GetPortalBatchUnshieldRequestStatus(stateDB, actionData.Meta.BatchID)
	if err != nil {
		Logger.log.Error("Can not get unshield batch: %v", err)
		return nil, err
	}

	var processedUnshieldRequestBatch metadata.PortalUnshieldRequestBatchStatus
	err = json.Unmarshal(unshieldBatchBytes, &processedUnshieldRequestBatch)
	if err != nil {
		Logger.log.Errorf("Replace fee: an error occurred while unmarshal processedUnshieldRequestBatch status: %+v", err)
		return nil, fmt.Errorf("Replace fee: an error occurred while unmarshal processedUnshieldRequestBatch status: %+v", err)
	}

	var outputs []*portaltokens.OutputTx
	for _, v := range processedUnshieldRequestBatch.UnshieldIDs {
		unshieldBytes, err := statedb.GetPortalUnshieldRequestStatus(stateDB, v)
		if err != nil {
			Logger.log.Error("Can not get unshield batch: %v", err)
			return nil, err
		}
		var portalUnshieldRequestStatus metadata.PortalUnshieldRequestStatus
		err = json.Unmarshal(unshieldBytes, &portalUnshieldRequestStatus)
		if err != nil {
			Logger.log.Errorf("Replace fee: an error occurred while unmarshal PortalUnshieldRequestStatus: %+v", err)
			return nil, fmt.Errorf("Replace fee: an error occurred while unmarshal PortalUnshieldRequestStatus: %+v", err)
		}
		outputs = append(outputs, &portaltokens.OutputTx{ReceiverAddress: portalUnshieldRequestStatus.RemoteAddress, Amount: portalUnshieldRequestStatus.UnshieldAmount})
	}

	optionalData := make(map[string]interface{})
	optionalData["outputs"] = outputs
	return optionalData, nil
}

// beacon build new instruction from instruction received from ShardToBeaconBlock
func buildReplacementFeeRequestInst(
	tokenID string,
	incAddressStr string,
	fee uint,
	batchID string,
	metaType int,
	shardID byte,
	externalRawTx string,
	txReqID common.Hash,
	status string,
) []string {
	replacementRequestContent := metadata.PortalReplacementFeeRequestContent{
		TokenID:       tokenID,
		IncAddressStr: incAddressStr,
		Fee:           fee,
		BatchID:       batchID,
		TxReqID:       txReqID,
		ExternalRawTx: externalRawTx,
	}
	replacementRequestContentBytes, _ := json.Marshal(replacementRequestContent)
	return []string{
		strconv.Itoa(metaType),
		strconv.Itoa(int(shardID)),
		status,
		string(replacementRequestContentBytes),
	}
}

func (p *PortalFeeReplacementRequestProcessor) BuildNewInsts(
	bc metadata.ChainRetriever,
	contentStr string,
	shardID byte,
	currentPortalV4State *CurrentPortalStateV4,
	beaconHeight uint64,
	shardHeights map[byte]uint64,
	portalParams portalv4.PortalParams,
	optionalData map[string]interface{},
) ([][]string, error) {
	// parse instruction
	actionContentBytes, err := base64.StdEncoding.DecodeString(contentStr)
	if err != nil {
		Logger.log.Errorf("ERROR: an error occured while decoding content string of portal replacement fee request action: %+v", err)
		return nil, fmt.Errorf("ERROR: an error occured while decoding content string of portal replacement fee request action: %+v", err)
	}
	var actionData metadata.PortalReplacementFeeRequestAction
	err = json.Unmarshal(actionContentBytes, &actionData)
	if err != nil {
		Logger.log.Errorf("ERROR: an error occured while unmarshal portal replacement fee request action: %+v", err)
		return nil, fmt.Errorf("ERROR: an error occured while unmarshal portal replacement fee request action: %+v", err)
	}

	if currentPortalV4State == nil {
		Logger.log.Warn("WARN - [Unshield Request]: Current Portal state V4 is null.")
		return [][]string{}, nil
	}

	meta := actionData.Meta
	rejectInst := buildReplacementFeeRequestInst(
		meta.TokenID,
		meta.IncAddressStr,
		meta.Fee,
		meta.BatchID,
		meta.Type,
		actionData.ShardID,
		"",
		actionData.TxReqID,
		pCommon.PortalRequestRejectedChainStatus,
	)

	tokenIDStr := meta.TokenID
	keyUnshieldBatch := statedb.GenerateProcessedUnshieldRequestBatchObjectKey(tokenIDStr, meta.BatchID).String()
	unshieldBatch, ok := currentPortalV4State.ProcessedUnshieldRequests[tokenIDStr][keyUnshieldBatch]
	if !ok {
		Logger.log.Errorf("Error: Replace a non-exist unshield batch with tokenID: %v, batchid : %v.", tokenIDStr, meta.BatchID)
		return [][]string{rejectInst}, nil
	}
	latestBeaconHeight := GetMaxKeyValue(unshieldBatch.GetExternalFees())
	if latestBeaconHeight == 0 || !bc.CheckBlockTimeIsReachedByBeaconHeight(beaconHeight, latestBeaconHeight, portalParams.TimeSpaceForFeeReplacement) {
		Logger.log.Errorf("Error: Can not replace unshield batch with tokenID: %v, batchid : %v.", tokenIDStr, meta.BatchID)
		return [][]string{rejectInst}, nil
	}
	latestFee := unshieldBatch.GetExternalFees()[latestBeaconHeight]

	maxFeeTemp := latestFee * portalParams.MaxFeePercentageForEachStep
	if maxFeeTemp < latestFee {
		Logger.log.Errorf("Error: Invalid fee request with latest fee: %v, MaxFeeForEachStep : %v.", latestFee, portalParams.MaxFeePercentageForEachStep)
		return [][]string{rejectInst}, nil
	}

	if meta.Fee < latestFee || meta.Fee-latestFee > (maxFeeTemp/100) {
		Logger.log.Errorf("Error: Replace unshield batch with invalid fee: %v", meta.Fee)
		return [][]string{rejectInst}, nil
	}

	portalTokenProcessor := portalParams.PortalTokens[tokenIDStr]
	multisigAddress := portalParams.MultiSigAddresses[tokenIDStr]
	if unshieldBatch.GetUTXOs() == nil || unshieldBatch.GetUTXOs()[multisigAddress] == nil {
		Logger.log.Errorf("Error: Can not get utxos from unshield batch with multisig address: %v", multisigAddress)
		return [][]string{rejectInst}, nil
	}
	hexRawExtTxStr, _, err := portalTokenProcessor.CreateRawExternalTx(unshieldBatch.GetUTXOs()[multisigAddress], optionalData["outputs"].([]*portaltokens.OutputTx), uint64(meta.Fee), meta.BatchID, bc)
	if err != nil {
		Logger.log.Errorf("ERROR: an error occured create new raw transaction portal replacement fee: %+v", err)
		return nil, fmt.Errorf("ERROR: an error occured create new raw transaction portal replacement fee: %+v", err)
	}

	// build accept instruction
	newInst := buildReplacementFeeRequestInst(
		meta.TokenID,
		meta.IncAddressStr,
		meta.Fee,
		meta.BatchID,
		meta.Type,
		actionData.ShardID,
		hexRawExtTxStr,
		actionData.TxReqID,
		pCommon.PortalRequestAcceptedChainStatus,
	)

	// add new waiting unshield request to waiting list
	UpdatePortalStateAfterReplaceFeeRequest(currentPortalV4State, unshieldBatch, beaconHeight, meta.Fee, tokenIDStr, meta.BatchID)

	return [][]string{newInst}, nil
}

func (p *PortalFeeReplacementRequestProcessor) ProcessInsts(
	stateDB *statedb.StateDB,
	beaconHeight uint64,
	instructions []string,
	currentPortalV4State *CurrentPortalStateV4,
	portalParams portalv4.PortalParams,
	updatingInfoByTokenID map[common.Hash]metadata.UpdatingInfo,
) error {
	if currentPortalV4State == nil {
		Logger.log.Errorf("current portal state is nil")
		return nil
	}

	if len(instructions) != 4 {
		return nil // skip the instruction
	}

	// unmarshal instructions content
	var actionData metadata.PortalReplacementFeeRequestContent
	err := json.Unmarshal([]byte(instructions[3]), &actionData)
	if err != nil {
		Logger.log.Errorf("Can not unmarshal instruction content %v - Error %v\n", instructions[3], err)
		return nil
	}

	reqStatus := instructions[2]
	var unshieldBatchRequestStatus metadata.PortalReplacementFeeRequestStatus

	if reqStatus == pCommon.PortalRequestAcceptedChainStatus {
		// update unshield batch
		keyUnshieldBatch := statedb.GenerateProcessedUnshieldRequestBatchObjectKey(actionData.TokenID, actionData.BatchID).String()
		unshieldBatch := currentPortalV4State.ProcessedUnshieldRequests[actionData.TokenID][keyUnshieldBatch]
		UpdatePortalStateAfterReplaceFeeRequest(currentPortalV4State, unshieldBatch, beaconHeight, actionData.Fee, actionData.TokenID, actionData.BatchID)

		// track status of unshield batch request by batchID
		unshieldBatchRequestStatus = metadata.PortalReplacementFeeRequestStatus{
			IncAddressStr: actionData.IncAddressStr,
			TokenID:       actionData.TokenID,
			BatchID:       actionData.BatchID,
			Fee:           actionData.Fee,
			ExternalRawTx: actionData.ExternalRawTx,
			BeaconHeight:  beaconHeight + 1,
			TxHash:        actionData.TxReqID.String(),
			Status:        pCommon.PortalRequestAcceptedStatus,
		}
	} else if reqStatus == pCommon.PortalRequestRejectedChainStatus {

		unshieldBatchRequestStatus = metadata.PortalReplacementFeeRequestStatus{
			IncAddressStr: actionData.IncAddressStr,
			TokenID:       actionData.TokenID,
			BatchID:       actionData.BatchID,
			ExternalRawTx: actionData.ExternalRawTx,
			Fee:           actionData.Fee,
			BeaconHeight:  beaconHeight + 1,
			TxHash:        actionData.TxReqID.String(),
			Status:        pCommon.PortalRequestRejectedStatus,
		}
	} else {
		return nil
	}
	unshieldBatchStatusBytes, _ := json.Marshal(unshieldBatchRequestStatus)
	err = statedb.StorePortalUnshieldBatchReplacementRequestStatus(
		stateDB,
		actionData.TxReqID.String(),
		unshieldBatchStatusBytes)
	if err != nil {
		Logger.log.Errorf("[processPortalReplacementRequest] Error when storing status of replacement request: %v\n", err)
		return nil
	}

	return nil
}

/* =======
Portal Submit external unshield tx confirmed Processor V4
======= */

type PortalSubmitConfirmedTxProcessor struct {
	*PortalInstProcessorV4
}

func (p *PortalSubmitConfirmedTxProcessor) GetActions() map[byte][][]string {
	return p.Actions
}

func (p *PortalSubmitConfirmedTxProcessor) PutAction(action []string, shardID byte) {
	_, found := p.Actions[shardID]
	if !found {
		p.Actions[shardID] = [][]string{action}
	} else {
		p.Actions[shardID] = append(p.Actions[shardID], action)
	}
}

func (p *PortalSubmitConfirmedTxProcessor) PrepareDataForBlockProducer(stateDB *statedb.StateDB, contentStr string) (map[string]interface{}, error) {
	actionContentBytes, err := base64.StdEncoding.DecodeString(contentStr)
	if err != nil {
		Logger.log.Errorf("SubmitConfirmed request: an error occurred while decoding content string of SubmitConfirmed unshield request action: %+v", err)
		return nil, fmt.Errorf("SubmitConfirmed request: an error occurred while decoding content string of SubmitConfirmed unshield request action: %+v", err)
	}

	var actionData metadata.PortalSubmitConfirmedTxAction
	err = json.Unmarshal(actionContentBytes, &actionData)
	if err != nil {
		Logger.log.Errorf("SubmitConfirmed request: an error occurred while unmarshal SubmitConfirmed unshield request action: %+v", err)
		return nil, fmt.Errorf("SubmitConfirmed request: an error occurred while unmarshal SubmitConfirmed unshield request action: %+v", err)
	}

	unshieldBatchBytes, err := statedb.GetPortalBatchUnshieldRequestStatus(stateDB, actionData.Meta.BatchID)
	if err != nil {
		Logger.log.Error("Can not get unshield batch: %v", err)
		return nil, err
	}

	var processedUnshieldRequestBatch metadata.PortalUnshieldRequestBatchStatus
	err = json.Unmarshal(unshieldBatchBytes, &processedUnshieldRequestBatch)
	if err != nil {
		Logger.log.Errorf("SubmitConfirmed request: an error occurred while unmarshal processedUnshieldRequestBatch status: %+v", err)
		return nil, fmt.Errorf("SubmitConfirmed request: an error occurred while unmarshal processedUnshieldRequestBatch status: %+v", err)
	}

	outputs := make(map[string]uint64, 0)
	for _, v := range processedUnshieldRequestBatch.UnshieldIDs {
		unshieldBytes, err := statedb.GetPortalUnshieldRequestStatus(stateDB, v)
		if err != nil {
			Logger.log.Error("Can not get unshield batch: %v", err)
			return nil, err
		}
		var portalUnshieldRequestStatus metadata.PortalUnshieldRequestStatus
		err = json.Unmarshal(unshieldBytes, &portalUnshieldRequestStatus)
		if err != nil {
			Logger.log.Errorf("SubmitConfirmed: an error occurred while unmarshal PortalUnshieldRequestStatus: %+v", err)
			return nil, fmt.Errorf("SubmitConfirmed: an error occurred while unmarshal PortalUnshieldRequestStatus: %+v", err)
		}
		outputs[portalUnshieldRequestStatus.RemoteAddress] = portalUnshieldRequestStatus.UnshieldAmount
	}

	optionalData := make(map[string]interface{})
	optionalData["outputs"] = outputs

	return optionalData, nil
}

// beacon build new instruction from instruction received from ShardToBeaconBlock
func buildSubmitConfirmedTxInst(
	tokenID string,
	utxos []*statedb.UTXO,
	batchID string,
	metaType int,
	shardID byte,
	txReqID common.Hash,
	status string,
) []string {
	replacementRequestContent := metadata.PortalSubmitConfirmedTxContent{
		TokenID: tokenID,
		UTXOs:   utxos,
		BatchID: batchID,
		TxReqID: txReqID,
	}
	replacementRequestContentBytes, _ := json.Marshal(replacementRequestContent)
	return []string{
		strconv.Itoa(metaType),
		strconv.Itoa(int(shardID)),
		status,
		string(replacementRequestContentBytes),
	}
}

func (p *PortalSubmitConfirmedTxProcessor) BuildNewInsts(
	bc metadata.ChainRetriever,
	contentStr string,
	shardID byte,
	currentPortalV4State *CurrentPortalStateV4,
	beaconHeight uint64,
	shardHeights map[byte]uint64,
	portalParams portalv4.PortalParams,
	optionalData map[string]interface{},
) ([][]string, error) {
	// parse instruction
	actionContentBytes, err := base64.StdEncoding.DecodeString(contentStr)
	if err != nil {
		Logger.log.Errorf("ERROR: an error occured while decoding content string of portal replacement fee request action: %+v", err)
		return nil, fmt.Errorf("ERROR: an error occured while decoding content string of portal replacement fee request action: %+v", err)
	}
	var actionData metadata.PortalSubmitConfirmedTxAction
	err = json.Unmarshal(actionContentBytes, &actionData)
	if err != nil {
		Logger.log.Errorf("ERROR: an error occured while unmarshal portal replacement fee request action: %+v", err)
		return nil, fmt.Errorf("ERROR: an error occured while unmarshal portal replacement fee request action: %+v", err)
	}

	if currentPortalV4State == nil {
		Logger.log.Warn("WARN - [Unshield Request]: Current Portal state V4 is null.")
		return [][]string{}, nil
	}

	meta := actionData.Meta
	listUTXO := []*statedb.UTXO{}
	rejectInst := buildSubmitConfirmedTxInst(
		meta.TokenID,
		listUTXO,
		meta.BatchID,
		meta.Type,
		actionData.ShardID,
		actionData.TxReqID,
		pCommon.PortalRequestRejectedChainStatus,
	)

	tokenIDStr := meta.TokenID
	batchIDStr := meta.BatchID
	keyUnshieldBatch := statedb.GenerateProcessedUnshieldRequestBatchObjectKey(tokenIDStr, batchIDStr).String()
	if currentPortalV4State.ProcessedUnshieldRequests == nil ||
		currentPortalV4State.ProcessedUnshieldRequests[tokenIDStr] == nil {
		Logger.log.Errorf("Error: currentPortalV4State.ProcessedUnshieldRequests not initialized yet")
		return [][]string{rejectInst}, nil
	}
	unshieldBatch, ok := currentPortalV4State.ProcessedUnshieldRequests[tokenIDStr][keyUnshieldBatch]
	if !ok {
		Logger.log.Errorf("Error: Submit non-exist unshield external transaction with tokenID: %v, batchid : %v.", tokenIDStr, batchIDStr)
		return [][]string{rejectInst}, nil
	}
	portalTokenProcessor := portalParams.PortalTokens[meta.TokenID]
	if portalTokenProcessor == nil {
		Logger.log.Errorf("TokenID is not supported currently on Portal")
		return [][]string{rejectInst}, nil
	}

	expectedMultisigAddress := portalParams.MultiSigAddresses[tokenIDStr]
	outputs := optionalData["outputs"].(map[string]uint64)
	if unshieldBatch.GetUTXOs() == nil || unshieldBatch.GetUTXOs()[expectedMultisigAddress] == nil {
		Logger.log.Errorf("Error submit external confirmed tx: can not get utxos of wallet address: %v", expectedMultisigAddress)
		return [][]string{rejectInst}, nil
	}
	isValid, listUTXO, err := portalTokenProcessor.ParseAndVerifyUnshieldProof(meta.UnshieldProof, bc, batchIDStr, expectedMultisigAddress, outputs, unshieldBatch.GetUTXOs()[expectedMultisigAddress])
	if !isValid || err != nil {
		Logger.log.Errorf("Unshield Proof is invalid")
		return [][]string{rejectInst}, nil
	}

	// build accept instruction
	newInst := buildSubmitConfirmedTxInst(
		meta.TokenID,
		listUTXO,
		meta.BatchID,
		meta.Type,
		actionData.ShardID,
		actionData.TxReqID,
		pCommon.PortalRequestAcceptedChainStatus,
	)

	// remove unshield being processed and update status
	UpdatePortalStateAfterSubmitConfirmedTx(currentPortalV4State, tokenIDStr, keyUnshieldBatch)
	if len(listUTXO) > 0 {
		UpdatePortalStateUTXOs(currentPortalV4State, tokenIDStr, listUTXO)
	}

	return [][]string{newInst}, nil
}

func (p *PortalSubmitConfirmedTxProcessor) ProcessInsts(
	stateDB *statedb.StateDB,
	beaconHeight uint64,
	instructions []string,
	currentPortalV4State *CurrentPortalStateV4,
	portalParams portalv4.PortalParams,
	updatingInfoByTokenID map[common.Hash]metadata.UpdatingInfo,
) error {
	if currentPortalV4State == nil {
		Logger.log.Errorf("current portal state is nil")
		return nil
	}

	if len(instructions) != 4 {
		return nil // skip the instruction
	}

	// unmarshal instructions content
	var actionData metadata.PortalSubmitConfirmedTxContent
	err := json.Unmarshal([]byte(instructions[3]), &actionData)
	if err != nil {
		Logger.log.Errorf("Can not unmarshal instruction content %v - Error %v\n", instructions[3], err)
		return nil
	}

	reqStatus := instructions[2]
	var portalSubmitConfirmedStatus metadata.PortalSubmitConfirmedTxStatus

	if reqStatus == pCommon.PortalRequestAcceptedChainStatus {
		// update unshield batch
		keyUnshieldBatchHash := statedb.GenerateProcessedUnshieldRequestBatchObjectKey(actionData.TokenID, actionData.BatchID)
		keyUnshieldBatch := keyUnshieldBatchHash.String()
		unshieldRequests := currentPortalV4State.ProcessedUnshieldRequests[actionData.TokenID][keyUnshieldBatch].GetUnshieldRequests()
		UpdatePortalStateAfterSubmitConfirmedTx(currentPortalV4State, actionData.TokenID, keyUnshieldBatch)
		statedb.DeleteUnshieldBatchRequest(stateDB, keyUnshieldBatchHash)
		if len(actionData.UTXOs) > 0 {
			UpdatePortalStateUTXOs(currentPortalV4State, actionData.TokenID, actionData.UTXOs)
		}
		// track status of unshield batch request by batchID
		portalSubmitConfirmedStatus = metadata.PortalSubmitConfirmedTxStatus{
			TokenID: actionData.TokenID,
			BatchID: actionData.BatchID,
			UTXOs:   actionData.UTXOs,
			TxHash:  actionData.TxReqID.String(),
			Status:  pCommon.PortalRequestAcceptedStatus,
		}

		// update unshield list to completed
		for _, v := range unshieldRequests {
			unshieldRequestBytes, err := statedb.GetPortalUnshieldRequestStatus(stateDB, v)
			if err != nil {
				Logger.log.Errorf("[processPortalSubmitConfirmedTx] Error when query unshield tx by unshieldID: %v\n err: %v", v, err)
				return nil
			}
			var unshieldRequest metadata.PortalUnshieldRequestStatus
			err = json.Unmarshal(unshieldRequestBytes, &unshieldRequest)
			if err != nil {
				Logger.log.Errorf("Can not unmarshal instruction content %v - Error %v\n", unshieldRequestBytes, err)
				return nil
			}

			unshieldRequestStatus := metadata.PortalUnshieldRequestStatus{
				IncAddressStr:  unshieldRequest.IncAddressStr,
				RemoteAddress:  unshieldRequest.RemoteAddress,
				TokenID:        unshieldRequest.TokenID,
				UnshieldAmount: unshieldRequest.UnshieldAmount,
				UnshieldID:     unshieldRequest.UnshieldID,
				Status:         portalcommonv4.PortalUnshieldReqCompletedStatus,
			}
			redeemRequestStatusBytes, _ := json.Marshal(unshieldRequestStatus)
			err = statedb.StorePortalUnshieldRequestStatus(
				stateDB,
				actionData.TxReqID.String(),
				redeemRequestStatusBytes)
			if err != nil {
				Logger.log.Errorf("[processPortalSubmitConfirmedTx] Error store completed unshield request unshieldID: %v\n err: %v", v, err)
				return nil
			}
		}

	} else if reqStatus == pCommon.PortalRequestRejectedChainStatus {
		portalSubmitConfirmedStatus = metadata.PortalSubmitConfirmedTxStatus{
			TokenID: actionData.TokenID,
			BatchID: actionData.BatchID,
			UTXOs:   actionData.UTXOs,
			TxHash:  actionData.TxReqID.String(),
			Status:  pCommon.PortalRequestRejectedStatus,
		}
	} else {
		return nil
	}
	portalSubmitConfirmedStatusBytes, _ := json.Marshal(portalSubmitConfirmedStatus)
	err = statedb.StorePortalSubmitConfirmedTxRequestStatus(
		stateDB,
		actionData.TxReqID.String(),
		portalSubmitConfirmedStatusBytes)
	if err != nil {
		Logger.log.Errorf("[processPortalReplacementRequest] Error when storing status of replacement request: %v\n", err)
		return nil
	}

	return nil
}
