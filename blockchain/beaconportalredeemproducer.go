package blockchain

import (
	"encoding/base64"
	"encoding/json"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/metadata"
	"sort"
	"strconv"
)

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
) []string {
	redeemRequestContent := metadata.PortalRedeemRequestContent{
		UniqueRedeemID:          uniqueRedeemID,
		TokenID:                 tokenID,
		RedeemAmount:            redeemAmount,
		RedeemerIncAddressStr:   incAddressStr,
		RemoteAddress:           remoteAddress,
		MatchingCustodianDetail: matchingCustodianDetail,
		RedeemFee:               redeemFee,
		TxReqID:                 txReqID,
		ShardID:                 shardID,
	}
	redeemRequestContentBytes, _ := json.Marshal(redeemRequestContent)
	return []string{
		strconv.Itoa(metaType),
		strconv.Itoa(int(shardID)),
		status,
		string(redeemRequestContentBytes),
	}
}

// buildInstructionsForRedeemRequest builds instruction for redeem request action
func (blockchain *BlockChain) buildInstructionsForRedeemRequest(
	stateDB *statedb.StateDB,
	contentStr string,
	shardID byte,
	metaType int,
	currentPortalState *CurrentPortalState,
	beaconHeight uint64,
	portalParams PortalParams,
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
	if currentPortalState == nil {
		Logger.log.Warn("WARN - [buildInstructionsForRedeemRequest]: Current Portal state is null.")
		// need to mint ptoken to user
		inst := buildRedeemRequestInst(
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
		)
		return [][]string{inst}, nil
	}

	redeemID := meta.UniqueRedeemID

	// check uniqueRedeemID is existed waitingRedeem list or not
	keyWaitingRedeemRequest := statedb.GenerateWaitingRedeemRequestObjectKey(redeemID)
	keyWaitingRedeemRequestStr := keyWaitingRedeemRequest.String()
	waitingRedeemRequest := currentPortalState.WaitingRedeemRequests[keyWaitingRedeemRequestStr]
	if waitingRedeemRequest != nil {
		Logger.log.Errorf("RedeemID is existed in waiting redeem requests list %v\n", redeemID)
		inst := buildRedeemRequestInst(
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
		)
		return [][]string{inst}, nil
	}

	// check uniqueRedeemID is existed matched Redeem request list or not
	keyMatchedRedeemRequest := statedb.GenerateMatchedRedeemRequestObjectKey(redeemID).String()
	matchedRedeemRequest := currentPortalState.MatchedRedeemRequests[keyMatchedRedeemRequest]
	if matchedRedeemRequest != nil {
		Logger.log.Errorf("RedeemID is existed in matched redeem requests list %v\n", redeemID)
		inst := buildRedeemRequestInst(
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
		)
		return [][]string{inst}, nil
	}

	// check uniqueRedeemID is existed in db or not
	redeemRequestBytes, err := statedb.GetPortalRedeemRequestStatus(stateDB, meta.UniqueRedeemID)
	if err != nil {
		Logger.log.Errorf("Can not get redeem req status for redeemID %v, %v\n", meta.UniqueRedeemID, err)
		inst := buildRedeemRequestInst(
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
		)
		return [][]string{inst}, nil
	} else if len(redeemRequestBytes) > 0 {
		Logger.log.Errorf("RedeemID is existed in redeem requests list in db %v\n", redeemID)
		inst := buildRedeemRequestInst(
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
		)
		return [][]string{inst}, nil
	}

	// get tokenID from redeemTokenID
	tokenID := meta.TokenID

	// check redeem fee
	if currentPortalState.FinalExchangeRatesState == nil {
		Logger.log.Errorf("Can not get exchange rate at beaconHeight %v\n", beaconHeight)
		inst := buildRedeemRequestInst(
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
		)
		return [][]string{inst}, nil
	}
	minRedeemFee, err := CalMinRedeemFee(meta.RedeemAmount, tokenID, currentPortalState.FinalExchangeRatesState, portalParams.MinPercentRedeemFee)
	if err != nil {
		Logger.log.Errorf("Error when calculating minimum redeem fee %v\n", err)
		inst := buildRedeemRequestInst(
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
		)
		return [][]string{inst}, nil
	}

	if meta.RedeemFee < minRedeemFee {
		Logger.log.Errorf("Redeem fee is invalid, minRedeemFee %v, but get %v\n", minRedeemFee, meta.RedeemFee)
		inst := buildRedeemRequestInst(
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
		)
		return [][]string{inst}, nil
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
	)
	return [][]string{inst}, nil
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

func (blockchain *BlockChain) buildInstructionsForReqMatchingRedeem(
	stateDB *statedb.StateDB,
	contentStr string,
	shardID byte,
	metaType int,
	currentPortalState *CurrentPortalState,
	beaconHeight uint64,
	portalParams PortalParams,
	newMatchedRedeemReqIDs []string,
) ([][]string, []string, error) {
	// parse instruction
	actionContentBytes, err := base64.StdEncoding.DecodeString(contentStr)
	if err != nil {
		Logger.log.Errorf("ERROR: an error occured while decoding content string of portal req matching redeem action: %+v", err)
		return [][]string{}, newMatchedRedeemReqIDs, nil
	}
	var actionData metadata.PortalReqMatchingRedeemAction
	err = json.Unmarshal(actionContentBytes, &actionData)
	if err != nil {
		Logger.log.Errorf("ERROR: an error occured while unmarshal portal req matching redeem action: %+v", err)
		return [][]string{}, newMatchedRedeemReqIDs, nil
	}

	meta := actionData.Meta
	if currentPortalState == nil {
		Logger.log.Warn("WARN - [buildInstructionsForRedeemRequest]: Current Portal state is null.")
		inst := buildReqMatchingRedeemInst(
			meta.RedeemID,
			meta.CustodianAddressStr,
			0,
			false,
			meta.Type,
			actionData.ShardID,
			actionData.TxReqID,
			common.PortalReqMatchingRedeemRejectedChainStatus,
		)
		return [][]string{inst}, newMatchedRedeemReqIDs, nil
	}

	redeemID := meta.RedeemID
	amountMatching, isEnoughCustodians, err := MatchCustodianToWaitingRedeemReq(meta.CustodianAddressStr, redeemID, currentPortalState)
	if err != nil {
		Logger.log.Errorf("Error when processing matching custodian and redeemID %v\n", err)
		inst := buildReqMatchingRedeemInst(
			meta.RedeemID,
			meta.CustodianAddressStr,
			0,
			false,
			meta.Type,
			actionData.ShardID,
			actionData.TxReqID,
			common.PortalReqMatchingRedeemRejectedChainStatus,
		)
		return [][]string{inst}, newMatchedRedeemReqIDs, nil
	}

	// if custodian is valid to matching, append custodian in matchingCustodians of redeem request
	// update custodian state
	_, err = UpdatePortalStateAfterCustodianReqMatchingRedeem(meta.CustodianAddressStr, redeemID, amountMatching, isEnoughCustodians, currentPortalState)
	if err != nil {
		Logger.log.Errorf("Error when updating portal state %v\n", err)
		inst := buildReqMatchingRedeemInst(
			meta.RedeemID,
			meta.CustodianAddressStr,
			0,
			false,
			meta.Type,
			actionData.ShardID,
			actionData.TxReqID,
			common.PortalReqMatchingRedeemRejectedChainStatus,
		)
		return [][]string{inst}, newMatchedRedeemReqIDs, nil
	}

	if isEnoughCustodians {
		newMatchedRedeemReqIDs = append(newMatchedRedeemReqIDs, redeemID)
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
	return [][]string{inst}, newMatchedRedeemReqIDs, nil
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
	currentPortalState *CurrentPortalState,
	newMatchedRedeemReqIDs []string) ([][]string, []string, error) {
	insts := [][]string{}
	waitingRedeemKeys := []string{}
	for key := range currentPortalState.WaitingRedeemRequests {
		waitingRedeemKeys = append(waitingRedeemKeys, key)
	}
	sort.Strings(waitingRedeemKeys)
	for _, key := range waitingRedeemKeys {
		waitingRedeem := currentPortalState.WaitingRedeemRequests[key]
		redeemID := waitingRedeem.GetUniqueRedeemID()
		if (beaconHeight+1)-waitingRedeem.GetBeaconHeight() < blockchain.convertDurationTimeToBeaconBlocks(blockchain.GetPortalParams(beaconHeight).TimeOutWaitingRedeemRequest) {
			continue
		}

		// calculate amount need to be matched
		totalMatchedAmount := uint64(0)
		for _, cus := range waitingRedeem.GetCustodians() {
			totalMatchedAmount += cus.GetAmount()
		}
		neededMatchingAmountInPToken := waitingRedeem.GetRedeemAmount() - totalMatchedAmount
		if neededMatchingAmountInPToken <= 0 {
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

		newMatchedRedeemReqIDs = append(newMatchedRedeemReqIDs, redeemID)
		inst := buildInstPickingCustodiansForTimeoutWaitingRedeem(
			waitingRedeem.GetUniqueRedeemID(),
			moreCustodians,
			metadata.PortalPickMoreCustodianForRedeemMeta,
			common.PortalPickMoreCustodianRedeemSuccessChainStatus,
		)
		insts = append(insts, inst)
	}

	return insts, newMatchedRedeemReqIDs, nil
}
