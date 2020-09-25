package blockchain

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/metadata"
	"strconv"
)

/* =======
Portal Porting Request Processor
======= */
type portalPortingRequestProcessor struct {
	*portalInstProcessor
}

func (p *portalPortingRequestProcessor) getActions() map[byte][][]string {
	return p.actions
}

func (p *portalPortingRequestProcessor) putAction(action []string, shardID byte) {
	_, found := p.actions[shardID]
	if !found {
		p.actions[shardID] = [][]string{action}
	} else {
		p.actions[shardID] = append(p.actions[shardID], action)
	}
}

func (p *portalPortingRequestProcessor) prepareDataBeforeProcessing(stateDB *statedb.StateDB, contentStr string) (map[string]interface{}, error) {
	actionContentBytes, err := base64.StdEncoding.DecodeString(contentStr)
	if err != nil {
		Logger.log.Errorf("Porting request: an error occurred while decoding content string of portal porting request action: %+v", err)
		return nil, fmt.Errorf("Porting request: an error occurred while decoding content string of portal porting request action: %+v", err)
	}

	var actionData metadata.PortalUserRegisterAction
	err = json.Unmarshal(actionContentBytes, &actionData)
	if err != nil {
		Logger.log.Errorf("Porting request: an error occurred while unmarshal portal porting request action: %+v", err)
		return nil, fmt.Errorf("Porting request: an error occurred while unmarshal portal porting request action: %+v", err)
	}

	optionalData := make(map[string]interface{})

	// Get porting request with uniqueID from stateDB
	isExistPortingID, err := statedb.IsPortingRequestIdExist(stateDB, []byte(actionData.Meta.UniqueRegisterId))
	if err != nil{
		Logger.log.Errorf("Porting request: an error occurred while get porting request Id from DB: %+v", err)
		return nil, fmt.Errorf("Porting request: an error occurred while get porting request Id from DB: %+v", err)
	}

	optionalData["isExistPortingID"] = isExistPortingID
	return optionalData, nil
}

func buildRequestPortingInst(
	metaType int,
	shardID byte,
	reqStatus string,
	uniqueRegisterId string,
	incogAddressStr string,
	pTokenId string,
	registerAmount uint64,
	portingFee uint64,
	custodian []*statedb.MatchingPortingCustodianDetail,
	txReqID common.Hash,
) []string {
	portingRequestContent := metadata.PortalPortingRequestContent{
		UniqueRegisterId: uniqueRegisterId,
		IncogAddressStr:  incogAddressStr,
		PTokenId:         pTokenId,
		RegisterAmount:   registerAmount,
		PortingFee:       portingFee,
		Custodian:        custodian,
		TxReqID:          txReqID,
		ShardID:          shardID,
	}

	portingRequestContentBytes, _ := json.Marshal(portingRequestContent)
	return []string{
		strconv.Itoa(metaType),
		strconv.Itoa(int(shardID)),
		reqStatus,
		string(portingRequestContentBytes),
	}
}

func (p *portalPortingRequestProcessor) buildNewInsts(
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
		Logger.log.Errorf("Porting request: an error occurred while decoding content string of portal porting request action: %+v", err)
		return [][]string{}, nil
	}

	var actionData metadata.PortalUserRegisterAction
	err = json.Unmarshal(actionContentBytes, &actionData)
	if err != nil {
		Logger.log.Errorf("Porting request: an error occurred while unmarshal portal porting request action: %+v", err)
		return [][]string{}, nil
	}

	rejectInst := buildRequestPortingInst(
		actionData.Meta.Type,
		shardID,
		common.PortalPortingRequestRejectedChainStatus,
		actionData.Meta.UniqueRegisterId,
		actionData.Meta.IncogAddressStr,
		actionData.Meta.PTokenId,
		actionData.Meta.RegisterAmount,
		actionData.Meta.PortingFee,
		nil,
		actionData.TxReqID,
	)

	if currentPortalState == nil {
		Logger.log.Errorf("Porting request: Current Portal state is null")
		return [][]string{rejectInst}, nil
	}

	// check unique id from optionalData which get from statedb
	if optionalData == nil {
		Logger.log.Errorf("Porting request: optionalData is null")
		return [][]string{rejectInst}, nil
	}
	isExist, ok := optionalData["isExistPortingID"].(bool)
	if !ok {
		Logger.log.Errorf("Porting request: optionalData isExistPortingID is invalid")
		return [][]string{rejectInst}, nil
	}
	if isExist {
		Logger.log.Errorf("Porting request: Porting request id exist in db %v", actionData.Meta.UniqueRegisterId)
		return [][]string{rejectInst}, nil
	}

	waitingPortingRequestKey := statedb.GeneratePortalWaitingPortingRequestObjectKey(actionData.Meta.UniqueRegisterId)
	if _, ok := currentPortalState.WaitingPortingRequests[waitingPortingRequestKey.String()]; ok {
		Logger.log.Errorf("Porting request: Waiting porting request exist, key %v", waitingPortingRequestKey)
		return [][]string{rejectInst}, nil
	}

	// get exchange rates
	exchangeRatesState := currentPortalState.FinalExchangeRatesState
	if exchangeRatesState == nil {
		Logger.log.Errorf("Porting request, exchange rates not found")
		return [][]string{rejectInst}, nil
	}

	// validate porting fees
	exchangePortingFees, err := CalMinPortingFee(actionData.Meta.RegisterAmount, actionData.Meta.PTokenId, exchangeRatesState, portalParams.MinPercentPortingFee)
	if err != nil {
		Logger.log.Errorf("Porting request: Calculate Porting fee is error %v", err)
		return [][]string{rejectInst}, nil
	}
	if actionData.Meta.PortingFee < exchangePortingFees {
		Logger.log.Errorf("Porting request: Porting fees is invalid: expected %v - get %v", exchangePortingFees, actionData.Meta.PortingFee)
		return [][]string{rejectInst}, nil
	}

	// pick-up custodians
	if len(currentPortalState.CustodianPoolState) <= 0 {
		Logger.log.Errorf("Porting request: Custodian not found")
		return [][]string{rejectInst}, nil
	}

	var sortCustodianStateByFreeCollateral []CustodianStateSlice
	sortCustodianByAmountAscent(actionData.Meta, currentPortalState.CustodianPoolState, &sortCustodianStateByFreeCollateral)

	if len(sortCustodianStateByFreeCollateral) <= 0 {
		Logger.log.Errorf("Porting request, custodian not found")
		return [][]string{rejectInst}, nil
	}
	pickedCustodians, err := pickUpCustodians(actionData.Meta, exchangeRatesState, sortCustodianStateByFreeCollateral, currentPortalState, portalParams)
	if err != nil || len(pickedCustodians) == 0 {
		Logger.log.Errorf("Porting request: an error occurred while picking up custodians for the porting request: %+v", err)
		return [][]string{rejectInst}, nil
	}

	// Update custodian state after finishing choosing enough custodians for the porting request
	for _, cus := range pickedCustodians {
		cusKey := statedb.GenerateCustodianStateObjectKey(cus.IncAddress).String()
		// update custodian state
		err := UpdateCustodianStateAfterMatchingPortingRequest(currentPortalState, cusKey, actionData.Meta.PTokenId, cus.LockedAmountCollateral)
		if err != nil {
			Logger.log.Errorf("Porting request: an error occurred while updating custodian state: %+v", err)
			return [][]string{rejectInst}, nil
		}
	}

	acceptInst := buildRequestPortingInst(
		actionData.Meta.Type,
		shardID,
		common.PortalPortingRequestAcceptedChainStatus,
		actionData.Meta.UniqueRegisterId,
		actionData.Meta.IncogAddressStr,
		actionData.Meta.PTokenId,
		actionData.Meta.RegisterAmount,
		actionData.Meta.PortingFee,
		pickedCustodians,
		actionData.TxReqID,
	)

	newPortingRequestStateWaiting := statedb.NewWaitingPortingRequestWithValue(
		actionData.Meta.UniqueRegisterId,
		actionData.TxReqID,
		actionData.Meta.PTokenId,
		actionData.Meta.IncogAddressStr,
		actionData.Meta.RegisterAmount,
		pickedCustodians,
		actionData.Meta.PortingFee,
		beaconHeight+1,
	)

	keyWaitingPortingRequest := statedb.GeneratePortalWaitingPortingRequestObjectKey(actionData.Meta.UniqueRegisterId)
	currentPortalState.WaitingPortingRequests[keyWaitingPortingRequest.String()] = newPortingRequestStateWaiting

	return [][]string{acceptInst}, nil
}

/* =======
Portal Request Ptoken Processor
======= */

type portalRequestPTokenProcessor struct {
	*portalInstProcessor
}

func (p *portalRequestPTokenProcessor) getActions() map[byte][][]string {
	return p.actions
}

func (p *portalRequestPTokenProcessor) putAction(action []string, shardID byte) {
	_, found := p.actions[shardID]
	if !found {
		p.actions[shardID] = [][]string{action}
	} else {
		p.actions[shardID] = append(p.actions[shardID], action)
	}
}

func (p *portalRequestPTokenProcessor) prepareDataBeforeProcessing(stateDB *statedb.StateDB, contentStr string) (map[string]interface{}, error) {
	return nil, nil
}

// beacon build new instruction from instruction received from ShardToBeaconBlock
func buildReqPTokensInst(
	uniquePortingID string,
	tokenID string,
	incogAddressStr string,
	portingAmount uint64,
	portingProof string,
	metaType int,
	shardID byte,
	txReqID common.Hash,
	status string,
) []string {
	reqPTokenContent := metadata.PortalRequestPTokensContent{
		UniquePortingID: uniquePortingID,
		TokenID:         tokenID,
		IncogAddressStr: incogAddressStr,
		PortingAmount:   portingAmount,
		PortingProof:    portingProof,
		TxReqID:         txReqID,
		ShardID:         shardID,
	}
	reqPTokenContentBytes, _ := json.Marshal(reqPTokenContent)
	return []string{
		strconv.Itoa(metaType),
		strconv.Itoa(int(shardID)),
		status,
		string(reqPTokenContentBytes),
	}
}

func (p *portalRequestPTokenProcessor) buildNewInsts(
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
		Logger.log.Errorf("ERROR: an error occured while decoding content string of portal request ptoken action: %+v", err)
		return [][]string{}, nil
	}
	var actionData metadata.PortalRequestPTokensAction
	err = json.Unmarshal(actionContentBytes, &actionData)
	if err != nil {
		Logger.log.Errorf("ERROR: an error occured while unmarshal portal request ptoken action: %+v", err)
		return [][]string{}, nil
	}
	meta := actionData.Meta

	rejectInst := buildReqPTokensInst(
		meta.UniquePortingID,
		meta.TokenID,
		meta.IncogAddressStr,
		meta.PortingAmount,
		meta.PortingProof,
		meta.Type,
		shardID,
		actionData.TxReqID,
		common.PortalReqPTokensRejectedChainStatus,
	)

	if currentPortalState == nil {
		Logger.log.Warn("Request PTokens: Current Portal state is null.")
		return [][]string{rejectInst}, nil
	}

	// check meta.UniquePortingID is in waiting PortingRequests list in portal state or not
	portingID := meta.UniquePortingID
	keyWaitingPortingRequestStr := statedb.GeneratePortalWaitingPortingRequestObjectKey(portingID).String()
	waitingPortingRequest := currentPortalState.WaitingPortingRequests[keyWaitingPortingRequestStr]
	if waitingPortingRequest == nil {
		Logger.log.Errorf("PortingID is not existed in waiting porting requests list")
		return [][]string{rejectInst}, nil
	}

	// check tokenID
	if meta.TokenID != waitingPortingRequest.TokenID() {
		Logger.log.Errorf("TokenID is not correct in portingID req")
		return [][]string{rejectInst}, nil
	}

	// check porting amount
	if meta.PortingAmount != waitingPortingRequest.Amount() {
		Logger.log.Errorf("PortingAmount is not correct in portingID req")
		return [][]string{rejectInst}, nil
	}

	portalTokenProcessor := bc.config.ChainParams.PortalTokens[meta.TokenID]
	if portalTokenProcessor == nil {
		Logger.log.Errorf("TokenID is not supported currently on Portal")
		return [][]string{rejectInst}, nil
	}

	isValid, err := portalTokenProcessor.ParseAndVerifyProofForPorting(meta.PortingProof, waitingPortingRequest, bc)
	if !isValid || err != nil {
		Logger.log.Error("Parse proof and verify proof failed: %v", err)
		return [][]string{rejectInst}, nil
	}

	// update holding public token for custodians
	for _, cusDetail := range waitingPortingRequest.Custodians() {
		custodianKey := statedb.GenerateCustodianStateObjectKey(cusDetail.IncAddress)
		UpdateCustodianStateAfterUserRequestPToken(currentPortalState, custodianKey.String(), waitingPortingRequest.TokenID(), cusDetail.Amount)
	}

	inst := buildReqPTokensInst(
		actionData.Meta.UniquePortingID,
		actionData.Meta.TokenID,
		actionData.Meta.IncogAddressStr,
		actionData.Meta.PortingAmount,
		actionData.Meta.PortingProof,
		actionData.Meta.Type,
		shardID,
		actionData.TxReqID,
		common.PortalReqPTokensAcceptedChainStatus,
	)

	// remove waiting porting request from currentPortalState
	deleteWaitingPortingRequest(currentPortalState, keyWaitingPortingRequestStr)
	return [][]string{inst}, nil
}
