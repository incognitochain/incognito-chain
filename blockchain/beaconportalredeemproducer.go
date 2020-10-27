package blockchain

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/incognitochain/incognito-chain/wallet"
	"sort"
	"strconv"
)

/* =======
Portal Redeem Request Processor
======= */

type portalRedeemRequestProcessor struct {
	*portalInstProcessor
}

func (p *portalRedeemRequestProcessor) getActions() map[byte][][]string {
	return p.actions
}

func (p *portalRedeemRequestProcessor) putAction(action []string, shardID byte) {
	_, found := p.actions[shardID]
	if !found {
		p.actions[shardID] = [][]string{action}
	} else {
		p.actions[shardID] = append(p.actions[shardID], action)
	}
}

func (p *portalRedeemRequestProcessor) prepareDataBeforeProcessing(stateDB *statedb.StateDB, contentStr string) (map[string]interface{}, error) {
	// parse instruction
	actionContentBytes, err := base64.StdEncoding.DecodeString(contentStr)
	if err != nil {
		Logger.log.Errorf("ERROR: an error occured while decoding content string of portal redeem request action: %+v", err)
		return nil, fmt.Errorf("ERROR: an error occured while decoding content string of portal redeem request action: %+v", err)
	}
	var actionData metadata.PortalRedeemRequestAction
	err = json.Unmarshal(actionContentBytes, &actionData)
	if err != nil {
		Logger.log.Errorf("ERROR: an error occured while unmarshal portal redeem request action: %+v", err)
		return nil, fmt.Errorf("ERROR: an error occured while unmarshal portal redeem request action: %+v", err)
	}

	optionalData := make(map[string]interface{})

	// Get redeem request with uniqueID from stateDB
	redeemRequestBytes, err := statedb.GetPortalRedeemRequestStatus(stateDB, actionData.Meta.UniqueRedeemID)
	if err != nil {
		Logger.log.Errorf("Redeem request: an error occurred while get redeem request Id from DB: %+v", err)
		return nil, fmt.Errorf("Redeem request: an error occurred while get redeem request Id from DB: %+v", err)
	}

	optionalData["isExistRedeemID"] = len(redeemRequestBytes) > 0
	return optionalData, nil
}

// beacon build new instruction from instruction received from ShardToBeaconBlock
func buildRedeemRequestInst(
	uniqueRedeemID string,
	tokenID string,
	redeemAmount uint64,
	incAddressStr string,
	remoteAddress string,
	redeemFee uint64,
	matchingCustodianDetail []*statedb.MatchingRedeemCustodianDetail,
	metaType int,
	shardID byte,
	txReqID common.Hash,
	status string,
	shardHeight uint64,
	redeemerAddressForLiquidating string,
) []string {
	redeemRequestContent := metadata.PortalRedeemRequestContent{
		UniqueRedeemID:                uniqueRedeemID,
		TokenID:                       tokenID,
		RedeemAmount:                  redeemAmount,
		RedeemerIncAddressStr:         incAddressStr,
		RemoteAddress:                 remoteAddress,
		MatchingCustodianDetail:       matchingCustodianDetail,
		RedeemFee:                     redeemFee,
		TxReqID:                       txReqID,
		ShardID:                       shardID,
		ShardHeight:                   shardHeight,
		RedeemerAddressForLiquidating: redeemerAddressForLiquidating,
	}
	redeemRequestContentBytes, _ := json.Marshal(redeemRequestContent)
	return []string{
		strconv.Itoa(metaType),
		strconv.Itoa(int(shardID)),
		status,
		string(redeemRequestContentBytes),
	}
}

func (p *portalRedeemRequestProcessor) buildNewInsts(
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
		Logger.log.Errorf("ERROR: an error occured while decoding content string of portal redeem request action: %+v", err)
		return [][]string{}, nil
	}
	var actionData metadata.PortalRedeemRequestAction
	err = json.Unmarshal(actionContentBytes, &actionData)
	if err != nil {
		Logger.log.Errorf("ERROR: an error occured while unmarshal portal redeem request action: %+v", err)
		return [][]string{}, nil
	}

	meta := actionData.Meta
	rejectInst := buildRedeemRequestInst(
		meta.UniqueRedeemID,
		meta.TokenID,
		meta.RedeemAmount,
		meta.RedeemerIncAddressStr,
		meta.RemoteAddress,
		meta.RedeemFee,
		nil,
		meta.Type,
		actionData.ShardID,
		actionData.TxReqID,
		common.PortalRedeemRequestRejectedChainStatus,
		actionData.ShardHeight,
		meta.RedeemerAddressForLiquidating,
	)

	if currentPortalState == nil {
		Logger.log.Warn("WARN - [buildInstructionsForRedeemRequest]: Current Portal state is null.")
		return [][]string{rejectInst}, nil
	}

	redeemID := meta.UniqueRedeemID
	// check uniqueRedeemID is existed waitingRedeem list or not
	keyWaitingRedeemRequest := statedb.GenerateWaitingRedeemRequestObjectKey(redeemID)
	keyWaitingRedeemRequestStr := keyWaitingRedeemRequest.String()
	waitingRedeemRequest := currentPortalState.WaitingRedeemRequests[keyWaitingRedeemRequestStr]
	if waitingRedeemRequest != nil {
		Logger.log.Errorf("RedeemID is existed in waiting redeem requests list %v\n", redeemID)
		return [][]string{rejectInst}, nil
	}

	// check uniqueRedeemID is existed matched Redeem request list or not
	keyMatchedRedeemRequest := statedb.GenerateMatchedRedeemRequestObjectKey(redeemID).String()
	matchedRedeemRequest := currentPortalState.MatchedRedeemRequests[keyMatchedRedeemRequest]
	if matchedRedeemRequest != nil {
		Logger.log.Errorf("RedeemID is existed in matched redeem requests list %v\n", redeemID)
		return [][]string{rejectInst}, nil
	}

	// check uniqueRedeemID is existed in db or not
	if optionalData == nil {
		Logger.log.Errorf("Redeem request: optionalData is null")
		return [][]string{rejectInst}, nil
	}
	isExist, ok := optionalData["isExistRedeemID"].(bool)
	if !ok {
		Logger.log.Errorf("Redeem request: optionalData isExistPortingID is invalid")
		return [][]string{rejectInst}, nil
	}
	if isExist {
		Logger.log.Errorf("Redeem request: Porting request id exist in db %v", redeemID)
		return [][]string{rejectInst}, nil
	}

	// get tokenID from redeemTokenID
	tokenID := meta.TokenID

	// check redeem fee
	if currentPortalState.FinalExchangeRatesState == nil {
		Logger.log.Errorf("Redeem request: Can not get exchange rate at beaconHeight %v\n", beaconHeight)
		return [][]string{rejectInst}, nil
	}
	minRedeemFee, err := CalMinRedeemFee(meta.RedeemAmount, tokenID, currentPortalState.FinalExchangeRatesState, portalParams)
	if err != nil {
		Logger.log.Errorf("Error when calculating minimum redeem fee %v\n", err)
		return [][]string{rejectInst}, nil
	}

	if meta.RedeemFee < minRedeemFee {
		Logger.log.Errorf("Redeem fee is invalid, minRedeemFee %v, but get %v\n", minRedeemFee, meta.RedeemFee)
		return [][]string{rejectInst}, nil
	}

	// add to waiting Redeem list
	redeemRequest := statedb.NewRedeemRequestWithValue(
		meta.UniqueRedeemID,
		meta.TokenID,
		meta.RedeemerIncAddressStr,
		meta.RemoteAddress,
		meta.RedeemAmount,
		[]*statedb.MatchingRedeemCustodianDetail{},
		meta.RedeemFee,
		beaconHeight+1,
		actionData.TxReqID,
		actionData.ShardID,
		actionData.ShardHeight,
		meta.RedeemerAddressForLiquidating,
	)
	currentPortalState.WaitingRedeemRequests[keyWaitingRedeemRequestStr] = redeemRequest

	Logger.log.Infof("[Portal] Build accepted instruction for redeem request")
	inst := buildRedeemRequestInst(
		meta.UniqueRedeemID,
		meta.TokenID,
		meta.RedeemAmount,
		meta.RedeemerIncAddressStr,
		meta.RemoteAddress,
		meta.RedeemFee,
		[]*statedb.MatchingRedeemCustodianDetail{},
		meta.Type,
		actionData.ShardID,
		actionData.TxReqID,
		common.PortalRedeemRequestAcceptedChainStatus,
		actionData.ShardHeight,
		meta.RedeemerAddressForLiquidating,
	)
	return [][]string{inst}, nil
}

/* =======
Portal Request Mactching Redeem Processor
======= */

type portalRequestMatchingRedeemProcessor struct {
	*portalInstProcessor
}

func (p *portalRequestMatchingRedeemProcessor) getActions() map[byte][][]string {
	return p.actions
}

func (p *portalRequestMatchingRedeemProcessor) putAction(action []string, shardID byte) {
	_, found := p.actions[shardID]
	if !found {
		p.actions[shardID] = [][]string{action}
	} else {
		p.actions[shardID] = append(p.actions[shardID], action)
	}
}

func (p *portalRequestMatchingRedeemProcessor) prepareDataBeforeProcessing(stateDB *statedb.StateDB, contentStr string) (map[string]interface{}, error) {
	return nil, nil
}

func buildReqMatchingRedeemInst(
	redeemID string,
	incAddressStr string,
	matchingAmount uint64,
	isFullCustodian bool,
	metaType int,
	shardID byte,
	txReqID common.Hash,
	status string,
) []string {
	reqMatchingRedeemContent := metadata.PortalReqMatchingRedeemContent{
		CustodianAddressStr: incAddressStr,
		RedeemID:            redeemID,
		MatchingAmount:      matchingAmount,
		IsFullCustodian:     isFullCustodian,
		TxReqID:             txReqID,
		ShardID:             shardID,
	}
	reqMatchingRedeemContentBytes, _ := json.Marshal(reqMatchingRedeemContent)
	return []string{
		strconv.Itoa(metaType),
		strconv.Itoa(int(shardID)),
		status,
		string(reqMatchingRedeemContentBytes),
	}
}

func (p *portalRequestMatchingRedeemProcessor) buildNewInsts(
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
		Logger.log.Errorf("ERROR: an error occured while decoding content string of portal req matching redeem action: %+v", err)
		return [][]string{}, nil
	}
	var actionData metadata.PortalReqMatchingRedeemAction
	err = json.Unmarshal(actionContentBytes, &actionData)
	if err != nil {
		Logger.log.Errorf("ERROR: an error occured while unmarshal portal req matching redeem action: %+v", err)
		return [][]string{}, nil
	}

	meta := actionData.Meta
	rejectInst := buildReqMatchingRedeemInst(
		meta.RedeemID,
		meta.CustodianAddressStr,
		0,
		false,
		meta.Type,
		actionData.ShardID,
		actionData.TxReqID,
		common.PortalReqMatchingRedeemRejectedChainStatus,
	)

	if currentPortalState == nil {
		Logger.log.Warn("WARN - [buildInstructionsForRedeemRequest]: Current Portal state is null.")
		return [][]string{rejectInst}, nil
	}

	redeemID := meta.RedeemID
	amountMatching, isEnoughCustodians, err := MatchCustodianToWaitingRedeemReq(meta.CustodianAddressStr, redeemID, currentPortalState)
	if err != nil {
		Logger.log.Errorf("Error when processing matching custodian and redeemID %v\n", err)
		return [][]string{rejectInst}, nil
	}

	// if custodian is valid to matching, append custodian in matchingCustodians of redeem request
	// update custodian state
	_, err = UpdatePortalStateAfterCustodianReqMatchingRedeem(meta.CustodianAddressStr, redeemID, amountMatching, isEnoughCustodians, currentPortalState)
	if err != nil {
		Logger.log.Errorf("Error when updating portal state %v\n", err)
		return [][]string{rejectInst}, nil
	}

	inst := buildReqMatchingRedeemInst(
		meta.RedeemID,
		meta.CustodianAddressStr,
		amountMatching,
		isEnoughCustodians,
		meta.Type,
		actionData.ShardID,
		actionData.TxReqID,
		common.PortalReqMatchingRedeemAcceptedChainStatus,
	)
	return [][]string{inst}, nil
}

// PortalPickMoreCustodiansForRedeemReqContent - Beacon builds a new instruction with this content after timeout of redeem request
// It will be appended to beaconBlock
type PortalPickMoreCustodiansForRedeemReqContent struct {
	RedeemID   string
	Custodians []*statedb.MatchingRedeemCustodianDetail
}

func buildInstPickingCustodiansForTimeoutWaitingRedeem(
	redeemID string,
	custodians []*statedb.MatchingRedeemCustodianDetail,
	metaType int,
	status string) []string {
	pickMoreCustodians := PortalPickMoreCustodiansForRedeemReqContent{
		RedeemID:   redeemID,
		Custodians: custodians,
	}
	pickMoreCustodiansBytes, _ := json.Marshal(pickMoreCustodians)
	return []string{
		strconv.Itoa(metaType),
		strconv.Itoa(-1),
		status,
		string(pickMoreCustodiansBytes),
	}
}

// checkAndPickMoreCustodianForWaitingRedeemRequest check waiting redeem requests get timeout or not
// if the waiting redeem request gets timeout, but not enough matching custodians, auto pick up more custodians
func (blockchain *BlockChain) checkAndPickMoreCustodianForWaitingRedeemRequest(
	beaconHeight uint64,
	currentPortalState *CurrentPortalState) ([][]string, error) {
	insts := [][]string{}
	Logger.log.Errorf("=========== Check and pick more custodian for waiting redeem request ===========")
	waitingRedeemKeys := []string{}
	for key := range currentPortalState.WaitingRedeemRequests {
		waitingRedeemKeys = append(waitingRedeemKeys, key)
	}
	sort.Strings(waitingRedeemKeys)
	for _, waitingRedeemKey := range waitingRedeemKeys {
		waitingRedeem := currentPortalState.WaitingRedeemRequests[waitingRedeemKey]
		Logger.log.Errorf("=========== Check and pick more custodian for waiting redeem request %+v", waitingRedeem)
		if !blockchain.checkBlockTimeIsReached(beaconHeight, waitingRedeem.GetBeaconHeight(), blockchain.ShardChain[waitingRedeem.ShardID()].multiView.GetBestView().GetHeight(), waitingRedeem.ShardHeight(), blockchain.GetPortalParams(beaconHeight)) {
			continue
		}
		Logger.log.Errorf("=========== Check and pick more custodian for waiting redeem request =========== %v", waitingRedeem.GetUniqueRedeemID())

		// calculate amount need to be matched
		totalMatchedAmount := uint64(0)
		for _, cus := range waitingRedeem.GetCustodians() {
			totalMatchedAmount += cus.GetAmount()
		}
		neededMatchingAmountInPToken := waitingRedeem.GetRedeemAmount() - totalMatchedAmount
		if neededMatchingAmountInPToken > waitingRedeem.GetRedeemAmount() || neededMatchingAmountInPToken <= 0 {
			Logger.log.Errorf("Amount need to matching in redeem request %v is less than zero", neededMatchingAmountInPToken)
			continue
		}

		// pick up more custodians
		moreCustodians, err := pickupCustodianForRedeem(neededMatchingAmountInPToken, waitingRedeem.GetTokenID(), currentPortalState)
		if err != nil {
			Logger.log.Errorf("Error when pick up more custodians for timeout redeem request %v\n", err)
			inst := buildInstPickingCustodiansForTimeoutWaitingRedeem(
				waitingRedeem.GetUniqueRedeemID(),
				moreCustodians,
				metadata.PortalPickMoreCustodianForRedeemMeta,
				common.PortalPickMoreCustodianRedeemFailedChainStatus,
			)
			insts = append(insts, inst)

			// build instruction reject redeem request
			err := UpdateCustodianStateAfterRejectRedeemRequestByLiquidation(currentPortalState, waitingRedeem, beaconHeight)
			if err != nil {
				Logger.log.Errorf("[checkAndBuildInstRejectRedeemRequestByLiquidationExchangeRate] Error when updating custodian state %v - RedeemID %v\n: ",
					err, waitingRedeem.GetUniqueRedeemID())
				break
			}

			// remove redeem request from waiting redeem requests list
			deleteWaitingRedeemRequest(currentPortalState, waitingRedeemKey)

			// get shardId of redeemer
			redeemerKey, err := wallet.Base58CheckDeserialize(waitingRedeem.GetRedeemerAddress())
			if err != nil {
				Logger.log.Errorf("[checkAndBuildInstRejectRedeemRequestByLiquidationExchangeRate] Error when deserializing redeemer address string in redeemID %v - %v\n: ",
					waitingRedeem.GetUniqueRedeemID(), err)
				break
			}
			shardID := common.GetShardIDFromLastByte(redeemerKey.KeySet.PaymentAddress.Pk[len(redeemerKey.KeySet.PaymentAddress.Pk)-1])

			// build instruction
			inst2 := buildRedeemRequestInst(
				waitingRedeem.GetUniqueRedeemID(),
				waitingRedeem.GetTokenID(),
				waitingRedeem.GetRedeemAmount(),
				waitingRedeem.GetRedeemerAddress(),
				waitingRedeem.GetRedeemerRemoteAddress(),
				waitingRedeem.GetRedeemFee(),
				waitingRedeem.GetCustodians(),
				metadata.PortalRedeemRequestMeta,
				shardID,
				common.Hash{},
				common.PortalRedeemReqCancelledByLiquidationChainStatus,
				waitingRedeem.ShardHeight(),
				waitingRedeem.GetRedeemAddressForLiquidating(),
			)
			Logger.log.Errorf("Instruction reject waiting redeem request because of not enough custodians %+v", inst2)
			insts = append(insts, inst2)
			continue
		}

		// update custodian state (holding public tokens)
		_, err = UpdatePortalStateAfterPickMoreCustodiansForWaitingRedeemReq(
			moreCustodians, waitingRedeem, currentPortalState)
		if err != nil {
			Logger.log.Errorf("Error when updating portal state %v\n", err)
			inst := buildInstPickingCustodiansForTimeoutWaitingRedeem(
				waitingRedeem.GetUniqueRedeemID(),
				moreCustodians,
				metadata.PortalPickMoreCustodianForRedeemMeta,
				common.PortalPickMoreCustodianRedeemFailedChainStatus,
			)
			insts = append(insts, inst)
			continue
		}

		inst := buildInstPickingCustodiansForTimeoutWaitingRedeem(
			waitingRedeem.GetUniqueRedeemID(),
			moreCustodians,
			metadata.PortalPickMoreCustodianForRedeemMeta,
			common.PortalPickMoreCustodianRedeemSuccessChainStatus,
		)
		insts = append(insts, inst)
	}

	return insts, nil
}

/* =======
Portal Request Unlock Collateral Processor
======= */

type portalRequestUnlockCollateralProcessor struct {
	*portalInstProcessor
}

func (p *portalRequestUnlockCollateralProcessor) getActions() map[byte][][]string {
	return p.actions
}

func (p *portalRequestUnlockCollateralProcessor) putAction(action []string, shardID byte) {
	_, found := p.actions[shardID]
	if !found {
		p.actions[shardID] = [][]string{action}
	} else {
		p.actions[shardID] = append(p.actions[shardID], action)
	}
}

func (p *portalRequestUnlockCollateralProcessor) prepareDataBeforeProcessing(stateDB *statedb.StateDB, contentStr string) (map[string]interface{}, error) {
	return nil, nil
}

// beacon build new instruction from instruction received from ShardToBeaconBlock
func buildReqUnlockCollateralInst(
	uniqueRedeemID string,
	tokenID string,
	custodianAddressStr string,
	redeemAmount uint64,
	unlockAmount uint64,
	redeemProof string,
	metaType int,
	shardID byte,
	txReqID common.Hash,
	status string,
) []string {
	reqUnlockCollateralContent := metadata.PortalRequestUnlockCollateralContent{
		UniqueRedeemID:      uniqueRedeemID,
		TokenID:             tokenID,
		CustodianAddressStr: custodianAddressStr,
		RedeemAmount:        redeemAmount,
		UnlockAmount:        unlockAmount,
		RedeemProof:         redeemProof,
		TxReqID:             txReqID,
		ShardID:             shardID,
	}
	reqUnlockCollateralContentBytes, _ := json.Marshal(reqUnlockCollateralContent)
	return []string{
		strconv.Itoa(metaType),
		strconv.Itoa(int(shardID)),
		status,
		string(reqUnlockCollateralContentBytes),
	}
}

func (p *portalRequestUnlockCollateralProcessor) buildNewInsts(
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
		Logger.log.Errorf("ERROR: an error occured while decoding content string of portal request unlock collateral action: %+v", err)
		return [][]string{}, nil
	}
	var actionData metadata.PortalRequestUnlockCollateralAction
	err = json.Unmarshal(actionContentBytes, &actionData)
	if err != nil {
		Logger.log.Errorf("ERROR: an error occured while unmarshal portal request unlock collateral action: %+v", err)
		return [][]string{}, nil
	}
	meta := actionData.Meta

	rejectInst := buildReqUnlockCollateralInst(
		meta.UniqueRedeemID,
		meta.TokenID,
		meta.CustodianAddressStr,
		meta.RedeemAmount,
		0,
		meta.RedeemProof,
		meta.Type,
		shardID,
		actionData.TxReqID,
		common.PortalReqUnlockCollateralRejectedChainStatus,
	)

	if currentPortalState == nil {
		Logger.log.Warn("WARN - [buildInstructionsForReqUnlockCollateral]: Current Portal state is null.")
		return [][]string{rejectInst}, nil
	}

	// check meta.UniqueRedeemID is in matched RedeemRequests list in portal state or not
	redeemID := meta.UniqueRedeemID
	keyMatchedRedeemRequestStr := statedb.GenerateMatchedRedeemRequestObjectKey(redeemID).String()
	matchedRedeemRequest := currentPortalState.MatchedRedeemRequests[keyMatchedRedeemRequestStr]
	if matchedRedeemRequest == nil {
		Logger.log.Errorf("redeemID is not existed in matched redeem requests list")
		return [][]string{rejectInst}, nil
	}

	// check tokenID
	if meta.TokenID != matchedRedeemRequest.GetTokenID() {
		Logger.log.Errorf("TokenID is not correct in redeemID req")
		return [][]string{rejectInst}, nil
	}

	// check redeem amount of matching custodian
	matchedCustodian := new(statedb.MatchingRedeemCustodianDetail)
	for _, cus := range matchedRedeemRequest.GetCustodians() {
		if cus.GetIncognitoAddress() == meta.CustodianAddressStr {
			matchedCustodian = cus
			break
		}
	}
	if matchedCustodian.GetIncognitoAddress() == "" {
		Logger.log.Errorf("Custodian address %v is not in redeemID req %v", meta.CustodianAddressStr, meta.UniqueRedeemID)
		return [][]string{rejectInst}, nil
	}

	if meta.RedeemAmount != matchedCustodian.GetAmount() {
		Logger.log.Errorf("RedeemAmount is not correct in redeemID req")
		return [][]string{rejectInst}, nil
	}

	portalTokenProcessor := bc.config.ChainParams.PortalTokens[meta.TokenID]
	if portalTokenProcessor == nil {
		Logger.log.Errorf("TokenID %v is not supported currently on Portal", meta.TokenID)
		return [][]string{rejectInst}, nil
	}

	isValid, err := portalTokenProcessor.ParseAndVerifyProofForRedeem(meta.RedeemProof, matchedRedeemRequest, bc, matchedCustodian)
	if !isValid || err != nil {
		Logger.log.Error("Parse and verify redeem proof failed: %v", err)
		return [][]string{rejectInst}, nil
	}

	// calculate unlock amount
	custodianStateKey := statedb.GenerateCustodianStateObjectKey(meta.CustodianAddressStr)
	var unlockAmount uint64
	custodianStateKeyStr := custodianStateKey.String()
	if meta.Type == metadata.PortalRequestUnlockCollateralMeta {
		unlockAmount, err = CalUnlockCollateralAmount(currentPortalState, custodianStateKeyStr, meta.RedeemAmount, meta.TokenID)
		if err != nil {
			Logger.log.Errorf("Error calculating unlock amount for custodian %v", err)
			return [][]string{rejectInst}, nil
		}

		// update custodian state (FreeCollateral, LockedAmountCollateral)
		// unlock amount in prv
		err = updateCustodianStateAfterReqUnlockCollateral(
			currentPortalState.CustodianPoolState[custodianStateKeyStr],
			unlockAmount, meta.TokenID)
		if err != nil {
			Logger.log.Errorf("Error when updating custodian state after unlocking collateral %v", err)
			return [][]string{rejectInst}, nil
		}
	} else {
		unlockAmount, _, err = CalUnlockCollateralAmountV3(currentPortalState, custodianStateKeyStr, meta.RedeemAmount, meta.TokenID, portalParams)
		if err != nil {
			Logger.log.Errorf("Error calculating unlock amount for custodian V3 %v", err)
			return [][]string{rejectInst}, nil
		}

		// update custodian state (FreeCollateral, LockedAmountCollateral)
		// unlock amount in usdt
		err = updateCustodianStateAfterReqUnlockCollateralV3(
			currentPortalState.CustodianPoolState[custodianStateKeyStr],
			unlockAmount, meta.TokenID, portalParams, currentPortalState)
		if err != nil {
			Logger.log.Errorf("Error when updating custodian state after unlocking collateral %v", err)
			return [][]string{rejectInst}, nil
		}
	}

	// update redeem request state in WaitingRedeemRequest (remove custodian from matchingCustodianDetail)
	updatedCustodians, err := removeCustodianFromMatchingRedeemCustodians(
		currentPortalState.MatchedRedeemRequests[keyMatchedRedeemRequestStr].GetCustodians(), meta.CustodianAddressStr)
	if err != nil {
		Logger.log.Errorf("ERROR: an error occurred while removing custodian %v from matching custodians", meta.CustodianAddressStr)
		return [][]string{rejectInst}, nil
	}
	currentPortalState.MatchedRedeemRequests[keyMatchedRedeemRequestStr].SetCustodians(updatedCustodians)

	// remove redeem request from WaitingRedeemRequest list when all matching custodians return public token to user
	// when list matchingCustodianDetail is empty
	if len(currentPortalState.MatchedRedeemRequests[keyMatchedRedeemRequestStr].GetCustodians()) == 0 {
		deleteMatchedRedeemRequest(currentPortalState, keyMatchedRedeemRequestStr)
	}

	inst := buildReqUnlockCollateralInst(
		meta.UniqueRedeemID,
		meta.TokenID,
		meta.CustodianAddressStr,
		meta.RedeemAmount,
		unlockAmount,
		meta.RedeemProof,
		meta.Type,
		shardID,
		actionData.TxReqID,
		common.PortalReqUnlockCollateralAcceptedChainStatus,
	)

	return [][]string{inst}, nil
}
