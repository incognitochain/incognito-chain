package portalprocess

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/incognitochain/incognito-chain/portal/portalv4"
	portalcommonv4 "github.com/incognitochain/incognito-chain/portal/portalv4/common"
	"github.com/incognitochain/incognito-chain/portal/portalv4/portaltokens"
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
	utxos []*statedb.UTXO,
	networkFee map[uint64]uint,
	beaconHeight uint64,
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
		BeaconHeight:  beaconHeight,
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
	currentPortalStateV4 *CurrentPortalStateV4,
	beaconHeight uint64,
	shardHeights map[byte]uint64,
	portalParams portalv4.PortalParams,
	optionalData map[string]interface{},
) ([][]string, error) {
	if currentPortalStateV4 == nil {
		Logger.log.Warn("[BatchUnshieldRequest]: Current Portal state V4 is null.")
		return [][]string{}, nil
	}

	newInsts := [][]string{}
	wUnshieldRequests := currentPortalStateV4.WaitingUnshieldRequests
	if len(wUnshieldRequests) > 0 && currentPortalStateV4.UTXOs == nil {
		Logger.log.Errorf("[BatchUnshieldRequest]: List utxos in current portal state is null.")
		return [][]string{}, nil
	}
	for tokenID, wReqs := range wUnshieldRequests {
		portalTokenProcessor := portalParams.PortalTokens[tokenID]
		if portalTokenProcessor == nil {
			Logger.log.Errorf("[BatchUnshieldRequest]: Portal token ID %v is null.", tokenID)
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
		if len(wReqForProcess) == 0 {
			Logger.log.Infof("[BatchUnshieldRequest]: List unshield request need to be processed of token ID %v is null.", tokenID)
			continue
		}

		// choose waiting unshield IDs to process with current UTXOs
		utxos := currentPortalStateV4.UTXOs[tokenID]
		dustAmount := portalTokenProcessor.ConvertIncToExternalAmount(portalParams.DustValueThreshold[tokenID])
		batchTxs := portalTokenProcessor.MatchUTXOsAndUnshieldIDs(utxos, wReqForProcess, dustAmount)

		// create raw external txs
		for _, bcTx := range batchTxs {
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

			// batchId: combine beacon height and list of unshieldIDs
			batchID := GetBatchID(beaconHeight+1, bcTx.UnshieldIDs)

			// create raw tx
			hexRawExtTxStr, _, err := portalTokenProcessor.CreateRawExternalTx(
				bcTx.UTXOs, outputTxs, feeUnshield, bc)
			if err != nil {
				Logger.log.Errorf("[BatchUnshieldRequest]: Error when creating raw external tx %v", err)
				continue
			}

			// update current portal state
			// remove chosen waiting unshield requests from waiting list
			// remove utxos
			externalFees := map[uint64]uint{
				beaconHeight + 1: uint(feeUnshield),
			}
			chosenUTXOs := bcTx.UTXOs
			currentPortalStateV4.UpdatePortalStateAfterProcessBatchUnshieldRequest(
				batchID, chosenUTXOs, externalFees, bcTx.UnshieldIDs, tokenID)

			// build new instruction with new raw external tx
			newInst := buildUnshieldBatchingInst(
				batchID, hexRawExtTxStr, tokenID, bcTx.UnshieldIDs, chosenUTXOs, externalFees, beaconHeight+1,
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
	currentPortalStateV4 *CurrentPortalStateV4,
	portalParams portalv4.PortalParams,
	updatingInfoByTokenID map[common.Hash]metadata.UpdatingInfo,
) error {
	if currentPortalStateV4 == nil {
		Logger.log.Errorf("[ProcessBatchUnshieldRequest] current portal state is nil")
		return nil
	}

	if len(instructions) != 4 {
		return nil // skip the instruction
	}

	// unmarshal instructions content
	var actionData metadata.PortalUnshieldRequestBatchContent
	err := json.Unmarshal([]byte(instructions[3]), &actionData)
	if err != nil {
		Logger.log.Errorf("[ProcessBatchUnshieldRequest] Can not unmarshal instruction content %v - Error %v\n", instructions[3], err)
		return nil
	}

	reqStatus := instructions[2]
	if reqStatus == portalcommonv4.PortalV4RequestAcceptedChainStatus {
		// store status of batch unshield by batchID
		err := statedb.StorePortalBatchUnshieldRequestStatus(
			stateDB,
			actionData.BatchID,
			[]byte(instructions[3]))
		if err != nil {
			Logger.log.Errorf("[ProcessBatchUnshieldRequest] Error when storing status of batch unshield requests: %v\n", err)
			return nil
		}

		// add new processed batch unshield request to batch unshield list
		// remove waiting unshield request from waiting list
		currentPortalStateV4.UpdatePortalStateAfterProcessBatchUnshieldRequest(
			actionData.BatchID, actionData.UTXOs, actionData.NetworkFee, actionData.UnshieldIDs, actionData.TokenID)

		for _, unshieldID := range actionData.UnshieldIDs {
			// update status of unshield request that processed
			err := UpdateNewStatusUnshieldRequest(unshieldID, portalcommonv4.PortalUnshieldReqProcessedStatus, "", 0, stateDB)
			if err != nil {
				Logger.log.Errorf("[ProcessBatchUnshieldRequest] Error when updating status of unshielding request with unshieldID %v: %v\n", unshieldID, err)
				return nil
			}
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
		Logger.log.Errorf("[ReplaceFeeRequest]: an error occurred while decoding content string of replace fee unshield request action: %+v", err)
		return nil, fmt.Errorf("[ReplaceFeeRequest]: an error occurred while decoding content string of replace fee unshield request action: %+v", err)
	}

	var actionData metadata.PortalReplacementFeeRequestAction
	err = json.Unmarshal(actionContentBytes, &actionData)
	if err != nil {
		Logger.log.Errorf("[ReplaceFeeRequest]: an error occurred while unmarshal replace fee unshield request action: %+v", err)
		return nil, fmt.Errorf("[ReplaceFeeRequest]: an error occurred while unmarshal replace fee unshield request action: %+v", err)
	}

	unshieldBatchBytes, err := statedb.GetPortalBatchUnshieldRequestStatus(stateDB, actionData.Meta.BatchID)
	if err != nil {
		Logger.log.Error("[ReplaceFeeRequest]: can not query unshield batch with BatchID - %v, ERROR - %+v", actionData.Meta.BatchID, err)
		return nil, err
	}

	var processedUnshieldRequestBatch metadata.PortalUnshieldRequestBatchContent
	err = json.Unmarshal(unshieldBatchBytes, &processedUnshieldRequestBatch)
	if err != nil {
		Logger.log.Errorf("[ReplaceFeeRequest]: an error occurred while unmarshal processedUnshieldRequestBatch status: %+v", err)
		return nil, fmt.Errorf("[ReplaceFeeRequest]: an error occurred while unmarshal processedUnshieldRequestBatch status: %+v", err)
	}

	var outputs []*portaltokens.OutputTx
	for _, v := range processedUnshieldRequestBatch.UnshieldIDs {
		unshieldBytes, err := statedb.GetPortalUnshieldRequestStatus(stateDB, v)
		if err != nil {
			Logger.log.Error("[ReplaceFeeRequest]: can not query unshield request with UnshieldID- %v, ERROR - %+v", v, err)
			return nil, err
		}
		var portalUnshieldRequestStatus metadata.PortalUnshieldRequestStatus
		err = json.Unmarshal(unshieldBytes, &portalUnshieldRequestStatus)
		if err != nil {
			Logger.log.Errorf("[ReplaceFeeRequest]: an error occurred while unmarshal PortalUnshieldRequestStatus: %+v", err)
			return nil, fmt.Errorf("[ReplaceFeeRequest]: an error occurred while unmarshal PortalUnshieldRequestStatus: %+v", err)
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
	fee uint,
	batchID string,
	metaType int,
	shardID byte,
	externalRawTx string,
	utxos []*statedb.UTXO,
	txReqID common.Hash,
	status string,
) []string {
	replacementRequestContent := metadata.PortalReplacementFeeRequestContent{
		TokenID:       tokenID,
		Fee:           fee,
		BatchID:       batchID,
		TxReqID:       txReqID,
		ExternalRawTx: externalRawTx,
		UTXOs:         utxos,
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
		Logger.log.Errorf("[ReplaceFeeRequest]: an error occured while decoding content string of portal replacement fee request action: %+v", err)
		return nil, fmt.Errorf("[ReplaceFeeRequest]: an error occured while decoding content string of portal replacement fee request action: %+v", err)
	}
	var actionData metadata.PortalReplacementFeeRequestAction
	err = json.Unmarshal(actionContentBytes, &actionData)
	if err != nil {
		Logger.log.Errorf("[ReplaceFeeRequest]: an error occured while unmarshal portal replacement fee request action: %+v", err)
		return nil, fmt.Errorf("[ReplaceFeeRequest]: an error occured while unmarshal portal replacement fee request action: %+v", err)
	}

	if currentPortalV4State == nil {
		Logger.log.Warn("WARN - [ReplaceFeeRequest]: current Portal state V4 is null.")
		return [][]string{}, nil
	}

	meta := actionData.Meta
	rejectInst := buildReplacementFeeRequestInst(
		meta.TokenID,
		meta.Fee,
		meta.BatchID,
		meta.Type,
		actionData.ShardID,
		"",
		nil,
		actionData.TxReqID,
		portalcommonv4.PortalV4RequestRejectedChainStatus,
	)

	tokenIDStr := meta.TokenID
	keyUnshieldBatch := statedb.GenerateProcessedUnshieldRequestBatchObjectKey(tokenIDStr, meta.BatchID).String()
	unshieldBatch, ok := currentPortalV4State.ProcessedUnshieldRequests[tokenIDStr][keyUnshieldBatch]
	if !ok {
		Logger.log.Errorf("[ReplaceFeeRequest]: replace a non-exist unshield batch with TokenID - %v, BatchID - %v.", tokenIDStr, meta.BatchID)
		return [][]string{rejectInst}, nil
	}
	latestBeaconHeight := GetMaxKeyValue(unshieldBatch.GetExternalFees())
	if latestBeaconHeight == 0 ||
		!bc.CheckBlockTimeIsReachedByBeaconHeight(beaconHeight+1, latestBeaconHeight, portalParams.TimeSpaceForFeeReplacement) {
		Logger.log.Errorf("[ReplaceFeeRequest]: can not replace unshield batch with TokenID - %v, BatchID - %v.", tokenIDStr, meta.BatchID)
		return [][]string{rejectInst}, nil
	}
	latestFee := unshieldBatch.GetExternalFees()[latestBeaconHeight]

	maxFeeTemp := latestFee * portalParams.MaxFeePercentageForEachStep
	if maxFeeTemp < latestFee {
		Logger.log.Errorf("[ReplaceFeeRequest]: invalid fee request with LatestFee - %v, MaxFeeForEachStep - %v.",
			latestFee, portalParams.MaxFeePercentageForEachStep)
		return [][]string{rejectInst}, nil
	}

	if meta.Fee < latestFee || meta.Fee-latestFee > (maxFeeTemp/100) {
		Logger.log.Errorf("[ReplaceFeeRequest]: replace unshield batch with invalid Fee - %v", meta.Fee)
		return [][]string{rejectInst}, nil
	}
	if uint64(meta.Fee)%portalParams.PortalTokens[meta.TokenID].GetMultipleTokenAmount() != 0 {
		Logger.log.Errorf("[ReplaceFeeRequest]: replace fee amount has to be divisible by 10 with Fee - %v", meta.Fee)
		return [][]string{rejectInst}, nil
	}

	portalTokenProcessor := portalParams.PortalTokens[tokenIDStr]
	if len(unshieldBatch.GetUTXOs()) == 0 {
		Logger.log.Errorf("[ReplaceFeeRequest]: UTXOs of unshield batchID - %v is empty: ", meta.BatchID)
		return [][]string{rejectInst}, nil
	}
	hexRawExtTxStr, _, err := portalTokenProcessor.CreateRawExternalTx(
		unshieldBatch.GetUTXOs(), optionalData["outputs"].([]*portaltokens.OutputTx), uint64(meta.Fee), bc)
	if err != nil {
		Logger.log.Errorf("[ReplaceFeeRequest]: an error occured create new raw transaction portal replacement fee: %+v", err)
		return nil, fmt.Errorf("[ReplaceFeeRequest]: an error occured create new raw transaction portal replacement fee: %+v", err)
	}

	// build accept instruction
	newInst := buildReplacementFeeRequestInst(
		meta.TokenID,
		meta.Fee,
		meta.BatchID,
		meta.Type,
		actionData.ShardID,
		hexRawExtTxStr,
		unshieldBatch.GetUTXOs(),
		actionData.TxReqID,
		portalcommonv4.PortalV4RequestAcceptedChainStatus,
	)

	// update external fee for batch processed unshield requests
	currentPortalV4State.AddExternalFeeForBatchProcessedUnshieldRequest(
		unshieldBatch.GetBatchID(), tokenIDStr, meta.Fee, beaconHeight+1)

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
		Logger.log.Errorf("[ReplaceFeeRequest]: current portal state is nil")
		return nil
	}

	if len(instructions) != 4 {
		return nil // skip the instruction
	}

	// unmarshal instructions content
	var actionData metadata.PortalReplacementFeeRequestContent
	err := json.Unmarshal([]byte(instructions[3]), &actionData)
	if err != nil {
		Logger.log.Errorf("[ReplaceFeeRequest]: can not unmarshal instruction content %v - Error %+v\n", instructions[3], err)
		return nil
	}

	reqStatus := instructions[2]
	var rbfStatus int
	if reqStatus == portalcommonv4.PortalV4RequestAcceptedChainStatus {
		rbfStatus = portalcommonv4.PortalV4RequestAcceptedStatus

		// update unshield batch
		currentPortalV4State.AddExternalFeeForBatchProcessedUnshieldRequest(
			actionData.BatchID, actionData.TokenID, actionData.Fee, beaconHeight+1)
	} else if reqStatus == portalcommonv4.PortalV4RequestRejectedChainStatus {
		rbfStatus = portalcommonv4.PortalV4RequestRejectedStatus
	}

	// track status of unshield batch request by batchID
	unshieldBatchRequestStatus := metadata.PortalReplacementFeeRequestStatus{
		TokenID:       actionData.TokenID,
		BatchID:       actionData.BatchID,
		Fee:           actionData.Fee,
		ExternalRawTx: actionData.ExternalRawTx,
		BeaconHeight:  beaconHeight + 1,
		TxHash:        actionData.TxReqID.String(),
		Status:        rbfStatus,
	}
	unshieldBatchStatusBytes, _ := json.Marshal(unshieldBatchRequestStatus)
	err = statedb.StorePortalUnshieldBatchReplacementRequestStatus(
		stateDB,
		actionData.TxReqID.String(),
		unshieldBatchStatusBytes)
	if err != nil {
		Logger.log.Errorf("[ReplaceFeeRequest]: Error when storing status of replacement request: %+v\n", err)
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
		Logger.log.Errorf("[SubmitConfirmedRequest]: an error occurred while decoding content string of SubmitConfirmed unshield request action: %+v", err)
		return nil, fmt.Errorf("[SubmitConfirmedRequest]: an error occurred while decoding content string of SubmitConfirmed unshield request action: %+v", err)
	}

	var actionData metadata.PortalSubmitConfirmedTxAction
	err = json.Unmarshal(actionContentBytes, &actionData)
	if err != nil {
		Logger.log.Errorf("[SubmitConfirmedRequest]: an error occurred while unmarshal SubmitConfirmed unshield request action: %+v", err)
		return nil, fmt.Errorf("[SubmitConfirmedRequest]: an error occurred while unmarshal SubmitConfirmed unshield request action: %+v", err)
	}

	unshieldBatchBytes, err := statedb.GetPortalBatchUnshieldRequestStatus(stateDB, actionData.Meta.BatchID)
	if err != nil {
		Logger.log.Error("[SubmitConfirmedRequest]: can not query unshield batch with BatchID - %v, ERROR - %+v", actionData.Meta.BatchID, err)
		return nil, err
	}

	var processedUnshieldRequestBatch metadata.PortalUnshieldRequestBatchContent
	err = json.Unmarshal(unshieldBatchBytes, &processedUnshieldRequestBatch)
	if err != nil {
		Logger.log.Errorf("[SubmitConfirmedRequest]: an error occurred while unmarshal processedUnshieldRequestBatch status: %+v", err)
		return nil, fmt.Errorf("[SubmitConfirmedRequest]: an error occurred while unmarshal processedUnshieldRequestBatch status: %+v", err)
	}

	outputs := []*portaltokens.OutputTx{}
	for _, v := range processedUnshieldRequestBatch.UnshieldIDs {
		unshieldBytes, err := statedb.GetPortalUnshieldRequestStatus(stateDB, v)
		if err != nil {
			Logger.log.Error("[SubmitConfirmedRequest]: can not query unshield request with UnshieldID - %v, ERROR - %+v", v, err)
			return nil, err
		}
		var portalUnshieldRequestStatus metadata.PortalUnshieldRequestStatus
		err = json.Unmarshal(unshieldBytes, &portalUnshieldRequestStatus)
		if err != nil {
			Logger.log.Errorf("[SubmitConfirmedRequest]: an error occurred while unmarshal PortalUnshieldRequestStatus: %+v", err)
			return nil, fmt.Errorf("[SubmitConfirmedRequest]: an error occurred while unmarshal PortalUnshieldRequestStatus: %+v", err)
		}
		outputs = append(outputs, &portaltokens.OutputTx{
			ReceiverAddress: portalUnshieldRequestStatus.RemoteAddress,
			Amount:          portalUnshieldRequestStatus.UnshieldAmount,
		})
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
	externalTxID string,
	externalFee uint64,
	status string,
) []string {
	replacementRequestContent := metadata.PortalSubmitConfirmedTxContent{
		TokenID:      tokenID,
		UTXOs:        utxos,
		BatchID:      batchID,
		TxReqID:      txReqID,
		ExternalTxID: externalTxID,
		ExternalFee:  externalFee,
		ShardID:      shardID,
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
		Logger.log.Errorf("[SubmitConfirmedRequest]: an error occured while decoding content string of portal replacement fee request action: %+v", err)
		return nil, fmt.Errorf("[SubmitConfirmedRequest]: an error occured while decoding content string of portal replacement fee request action: %+v", err)
	}
	var actionData metadata.PortalSubmitConfirmedTxAction
	err = json.Unmarshal(actionContentBytes, &actionData)
	if err != nil {
		Logger.log.Errorf("[SubmitConfirmedRequest]: an error occured while unmarshal portal replacement fee request action: %+v", err)
		return nil, fmt.Errorf("[SubmitConfirmedRequest]: an error occured while unmarshal portal replacement fee request action: %+v", err)
	}

	if currentPortalV4State == nil {
		Logger.log.Warn("WARN - [SubmitConfirmedRequest]: current Portal state V4 is null.")
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
		"", 0,
		portalcommonv4.PortalV4RequestRejectedChainStatus,
	)

	tokenIDStr := meta.TokenID
	batchIDStr := meta.BatchID
	keyUnshieldBatchHash := statedb.GenerateProcessedUnshieldRequestBatchObjectKey(tokenIDStr, batchIDStr)
	keyUnshieldBatch := keyUnshieldBatchHash.String()
	if currentPortalV4State.ProcessedUnshieldRequests == nil ||
		currentPortalV4State.ProcessedUnshieldRequests[tokenIDStr] == nil {
		Logger.log.Errorf("[SubmitConfirmedRequest]: currentPortalV4State.ProcessedUnshieldRequests not initialized yet")
		return [][]string{rejectInst}, nil
	}
	unshieldBatch, ok := currentPortalV4State.ProcessedUnshieldRequests[tokenIDStr][keyUnshieldBatch]
	if !ok {
		Logger.log.Errorf("[SubmitConfirmedRequest]: submit non-exist unshield external transaction with tokenID: %v, batchid : %v.", tokenIDStr, batchIDStr)
		return [][]string{rejectInst}, nil
	}
	portalTokenProcessor := portalParams.PortalTokens[meta.TokenID]
	if portalTokenProcessor == nil {
		Logger.log.Errorf("[SubmitConfirmedRequest]: tokenID - %v is currently not supported on Portal", meta.TokenID)
		return [][]string{rejectInst}, nil
	}

	expectedReceivedMultisigAddress := portalParams.GeneralMultiSigAddresses[tokenIDStr]
	outputs := optionalData["outputs"].([]*portaltokens.OutputTx)
	if len(unshieldBatch.GetUTXOs()) == 0 {
		Logger.log.Errorf("[SubmitConfirmedRequest]: UTXOs of unshield batchID - %v is empty: ", meta.BatchID)
		return [][]string{rejectInst}, nil
	}
	isValid, listUTXO, externalTxID, externalFee, err := portalTokenProcessor.ParseAndVerifyUnshieldProof(
		meta.UnshieldProof, bc, expectedReceivedMultisigAddress, "", outputs, unshieldBatch.GetUTXOs())
	if !isValid || err != nil {
		Logger.log.Errorf("[SubmitConfirmedRequest]: unshield Proof is invalid with Proof - %v, Error - %+v", meta.UnshieldProof, err)
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
		externalTxID,
		externalFee,
		portalcommonv4.PortalV4RequestAcceptedChainStatus,
	)

	// remove unshield being processed and update status
	currentPortalV4State.RemoveBatchProcessedUnshieldRequest(tokenIDStr, keyUnshieldBatchHash)
	if len(listUTXO) > 0 {
		currentPortalV4State.AddUTXOs(listUTXO, tokenIDStr)
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
		Logger.log.Errorf("[SubmitConfirmedRequest]: current portal state is nil")
		return nil
	}

	if len(instructions) != 4 {
		return nil // skip the instruction
	}

	// unmarshal instructions content
	var actionData metadata.PortalSubmitConfirmedTxContent
	err := json.Unmarshal([]byte(instructions[3]), &actionData)
	if err != nil {
		Logger.log.Errorf("[SubmitConfirmedRequest]: can not unmarshal instruction content %v - Error %+v\n", instructions[3], err)
		return nil
	}

	reqStatus := instructions[2]
	// var portalSubmitConfirmedStatus metadata.PortalSubmitConfirmedTxStatus
	var submitConfirmStatus int
	if reqStatus == portalcommonv4.PortalV4RequestAcceptedChainStatus {
		submitConfirmStatus = portalcommonv4.PortalV4RequestAcceptedStatus
		// update unshield batch
		keyUnshieldBatchHash := statedb.GenerateProcessedUnshieldRequestBatchObjectKey(actionData.TokenID, actionData.BatchID)
		keyUnshieldBatch := keyUnshieldBatchHash.String()
		unshieldRequests := currentPortalV4State.ProcessedUnshieldRequests[actionData.TokenID][keyUnshieldBatch].GetUnshieldRequests()
		currentPortalV4State.RemoveBatchProcessedUnshieldRequest(actionData.TokenID, keyUnshieldBatchHash)
		if len(actionData.UTXOs) > 0 {
			currentPortalV4State.AddUTXOs(actionData.UTXOs, actionData.TokenID)
		}

		// update unshield list to completed
		for _, unshieldID := range unshieldRequests {
			// update status of unshield request that processed
			err := UpdateNewStatusUnshieldRequest(unshieldID, portalcommonv4.PortalUnshieldReqCompletedStatus, actionData.ExternalTxID, actionData.ExternalFee, stateDB)
			if err != nil {
				Logger.log.Errorf("[SubmitConfirmedRequest]: error occur when updating status of unshielding request with UnshieldID - %v, ERROR - %+v\n", unshieldID, err)
				return nil
			}
		}
	} else if reqStatus == portalcommonv4.PortalV4RequestRejectedChainStatus {
		submitConfirmStatus = portalcommonv4.PortalV4RequestRejectedStatus
	}

	portalSubmitConfirmedStatus := metadata.PortalSubmitConfirmedTxStatus{
		TokenID:      actionData.TokenID,
		UTXOs:        actionData.UTXOs,
		BatchID:      actionData.BatchID,
		TxHash:       actionData.TxReqID.String(),
		ExternalTxID: actionData.ExternalTxID,
		ExternalFee:  actionData.ExternalFee,
		Status:       submitConfirmStatus,
	}
	portalSubmitConfirmedStatusBytes, _ := json.Marshal(portalSubmitConfirmedStatus)
	err = statedb.StorePortalSubmitConfirmedTxRequestStatus(
		stateDB,
		actionData.TxReqID.String(),
		portalSubmitConfirmedStatusBytes)
	if err != nil {
		Logger.log.Errorf("[SubmitConfirmedRequest]: error occur when storing status of submit confirm tx request"+
			": %+v\n", err)
	}

	return nil
}
