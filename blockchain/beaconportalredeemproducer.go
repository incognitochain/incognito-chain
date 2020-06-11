package blockchain

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/binance-chain/go-sdk/types/msg"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/incognitochain/incognito-chain/relaying/bnb"
	btcrelaying "github.com/incognitochain/incognito-chain/relaying/btc"
	"github.com/incognitochain/incognito-chain/wallet"
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

//func (blockchain * BlockChain) buildInstCancel

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
	for _, waitingRedeemKey := range waitingRedeemKeys {
		waitingRedeem := currentPortalState.WaitingRedeemRequests[waitingRedeemKey]
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
			)
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

// buildInstructionsForReqUnlockCollateral builds instruction for custodian deposit action
func (blockchain *BlockChain) buildInstructionsForReqUnlockCollateral(
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

	if currentPortalState == nil {
		Logger.log.Warn("WARN - [buildInstructionsForReqUnlockCollateral]: Current Portal state is null.")
		inst := buildReqUnlockCollateralInst(
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
		return [][]string{inst}, nil
	}

	// check meta.UniqueRedeemID is in matched RedeemRequests list in portal state or not
	redeemID := meta.UniqueRedeemID
	keyMatchedRedeemRequestStr := statedb.GenerateMatchedRedeemRequestObjectKey(redeemID).String()
	matchedRedeemRequest := currentPortalState.MatchedRedeemRequests[keyMatchedRedeemRequestStr]
	if matchedRedeemRequest == nil {
		Logger.log.Errorf("redeemID is not existed in matched redeem requests list")
		inst := buildReqUnlockCollateralInst(
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
		return [][]string{inst}, nil
	}

	// check status of request unlock collateral by redeemID
	redeemReqStatusBytes, err := statedb.GetPortalRedeemRequestStatus(stateDB, redeemID)
	if err != nil {
		Logger.log.Errorf("Can not get redeem request by redeemID from db %v\n", err)
		inst := buildReqUnlockCollateralInst(
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
		return [][]string{inst}, nil
	}
	var redeemRequest metadata.PortalRedeemRequestStatus
	err = json.Unmarshal(redeemReqStatusBytes, &redeemRequest)
	if err != nil {
		Logger.log.Errorf("Can not unmarshal redeem request %v\n", err)
		inst := buildReqUnlockCollateralInst(
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
		return [][]string{inst}, nil
	}

	if redeemRequest.Status != common.PortalRedeemReqMatchedStatus {
		Logger.log.Errorf("Redeem request %v has invalid status %v\n", redeemID, redeemRequest.Status)
		inst := buildReqUnlockCollateralInst(
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
		return [][]string{inst}, nil
	}

	// check tokenID
	if meta.TokenID != matchedRedeemRequest.GetTokenID() {
		Logger.log.Errorf("TokenID is not correct in redeemID req")
		inst := buildReqUnlockCollateralInst(
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
		return [][]string{inst}, nil
	}

	// check redeem amount of matching custodian
	amountMatchingCustodian := uint64(0)
	isFoundCustodian := false
	for _, cus := range matchedRedeemRequest.GetCustodians() {
		if cus.GetIncognitoAddress() == meta.CustodianAddressStr {
			amountMatchingCustodian = cus.GetAmount()
			isFoundCustodian = true
			break
		}
	}

	if !isFoundCustodian {
		Logger.log.Errorf("Custodian address %v is not in redeemID req %v", meta.CustodianAddressStr, meta.UniqueRedeemID)
		inst := buildReqUnlockCollateralInst(
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
		return [][]string{inst}, nil
	}

	if meta.RedeemAmount != amountMatchingCustodian {
		Logger.log.Errorf("RedeemAmount is not correct in redeemID req")
		inst := buildReqUnlockCollateralInst(
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
		return [][]string{inst}, nil
	}

	// validate proof and memo in tx
	if meta.TokenID == common.PortalBTCIDStr {
		btcChain := blockchain.config.BTCChain
		if btcChain == nil {
			Logger.log.Error("BTC relaying chain should not be null")
			inst := buildReqUnlockCollateralInst(
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
			return [][]string{inst}, nil
		}
		// parse PortingProof in meta
		btcTxProof, err := btcrelaying.ParseBTCProofFromB64EncodeStr(meta.RedeemProof)
		if err != nil {
			Logger.log.Errorf("PortingProof is invalid %v\n", err)
			inst := buildReqUnlockCollateralInst(
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
			return [][]string{inst}, nil
		}

		isValid, err := btcChain.VerifyTxWithMerkleProofs(btcTxProof)
		if !isValid || err != nil {
			Logger.log.Errorf("Verify btcTxProof failed %v", err)
			inst := buildReqUnlockCollateralInst(
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
			return [][]string{inst}, nil
		}

		// extract attached message from txOut's OP_RETURN
		btcAttachedMsg, err := btcrelaying.ExtractAttachedMsgFromTx(btcTxProof.BTCTx)
		if err != nil {
			Logger.log.Errorf("Could not extract message from btc proof with error: ", err)
			inst := buildReqUnlockCollateralInst(
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
			return [][]string{inst}, nil
		}

		rawMsg := fmt.Sprintf("%s%s", meta.UniqueRedeemID, meta.CustodianAddressStr)
		encodedMsg := btcrelaying.HashAndEncodeBase58(rawMsg)
		if btcAttachedMsg != encodedMsg {
			Logger.log.Errorf("The hash of combination of UniqueRedeemID(%s) and CustodianAddressStr(%s) is not matched to tx's attached message", meta.UniqueRedeemID, meta.CustodianAddressStr)
			inst := buildReqUnlockCollateralInst(
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
			return [][]string{inst}, nil
		}

		// check whether amount transfer in txBNB is equal redeem amount or not
		// check receiver and amount in tx
		// get list matching custodians in matchedRedeemRequest

		outputs := btcTxProof.BTCTx.TxOut
		remoteAddressNeedToBeTransfer := matchedRedeemRequest.GetRedeemerRemoteAddress()
		amountNeedToBeTransfer := meta.RedeemAmount
		amountNeedToBeTransferInBTC := btcrelaying.ConvertIncPBTCAmountToExternalBTCAmount(int64(amountNeedToBeTransfer))

		isChecked := false
		for _, out := range outputs {
			addrStr, err := btcChain.ExtractPaymentAddrStrFromPkScript(out.PkScript)
			if err != nil {
				Logger.log.Warnf("[portal] ExtractPaymentAddrStrFromPkScript: could not extract payment address string from pkscript with err: %v\n", err)
				continue
			}
			if addrStr != remoteAddressNeedToBeTransfer {
				continue
			}
			if out.Value < amountNeedToBeTransferInBTC {
				Logger.log.Errorf("BTC-TxProof is invalid - the transferred amount to %s must be equal to or greater than %d, but got %d", addrStr, amountNeedToBeTransferInBTC, out.Value)
				inst := buildReqUnlockCollateralInst(
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
				return [][]string{inst}, nil
			} else {
				isChecked = true
				break
			}
		}

		if !isChecked {
			Logger.log.Error("BTC-TxProof is invalid")
			inst := buildReqUnlockCollateralInst(
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
			return [][]string{inst}, nil
		}

		// calculate unlock amount
		custodianStateKey := statedb.GenerateCustodianStateObjectKey(meta.CustodianAddressStr)
		custodianStateKeyStr := custodianStateKey.String()
		unlockAmount, err := CalUnlockCollateralAmount(currentPortalState, custodianStateKeyStr, meta.RedeemAmount, meta.TokenID)
		if err != nil {
			Logger.log.Errorf("Error calculating unlock amount for custodian %v", err)
			inst := buildReqUnlockCollateralInst(
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
			return [][]string{inst}, nil
		}

		// update custodian state (FreeCollateral, LockedAmountCollateral)
		err = updateCustodianStateAfterReqUnlockCollateral(
			currentPortalState.CustodianPoolState[custodianStateKeyStr],
			unlockAmount, meta.TokenID)
		if err != nil {
			Logger.log.Errorf("Error when updating custodian state after unlocking collateral %v", err)
			inst := buildReqUnlockCollateralInst(
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
			return [][]string{inst}, nil
		}

		// update redeem request state in WaitingRedeemRequest (remove custodian from matchingCustodianDetail)
		updatedCustodians, err := removeCustodianFromMatchingRedeemCustodians(
			currentPortalState.MatchedRedeemRequests[keyMatchedRedeemRequestStr].GetCustodians(), meta.CustodianAddressStr)
		if err != nil {
			Logger.log.Errorf("ERROR: an error occurred while removing custodian %v from matching custodians", meta.CustodianAddressStr)
			inst := buildReqUnlockCollateralInst(
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
			return [][]string{inst}, nil
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

	} else if meta.TokenID == common.PortalBNBIDStr {
		// parse PortingProof in meta
		txProofBNB, err := bnb.ParseBNBProofFromB64EncodeStr(meta.RedeemProof)
		if err != nil {
			Logger.log.Errorf("RedeemProof is invalid %v\n", err)
			inst := buildReqUnlockCollateralInst(
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
			return [][]string{inst}, nil
		}

		// check minimum confirmations block of bnb proof
		latestBNBBlockHeight, err2 := blockchain.GetLatestBNBBlkHeight()
		if err2 != nil {
			Logger.log.Errorf("Can not get latest relaying bnb block height %v\n", err)
			inst := buildReqUnlockCollateralInst(
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
			return [][]string{inst}, nil
		}

		if latestBNBBlockHeight < txProofBNB.BlockHeight+bnb.MinConfirmationsBlock {
			Logger.log.Errorf("Not enough min bnb confirmations block %v, latestBNBBlockHeight %v - txProofBNB.BlockHeight %v\n",
				bnb.MinConfirmationsBlock, latestBNBBlockHeight, txProofBNB.BlockHeight)
			inst := buildReqUnlockCollateralInst(
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
			return [][]string{inst}, nil
		}

		dataHash, err2 := blockchain.GetBNBDataHash(txProofBNB.BlockHeight)
		if err2 != nil {
			Logger.log.Errorf("Error when get data hash in blockHeight %v - %v\n",
				txProofBNB.BlockHeight, err2)
			inst := buildReqUnlockCollateralInst(
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
			return [][]string{inst}, nil
		}

		isValid, err := txProofBNB.Verify(dataHash)
		if !isValid || err != nil {
			Logger.log.Errorf("Verify txProofBNB failed %v", err)
			inst := buildReqUnlockCollateralInst(
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
			return [][]string{inst}, nil
		}

		// parse Tx from Data in txProofBNB
		txBNB, err := bnb.ParseTxFromData(txProofBNB.Proof.Data)
		if err != nil {
			Logger.log.Errorf("Data in RedeemProof is invalid %v", err)
			inst := buildReqUnlockCollateralInst(
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
			return [][]string{inst}, nil
		}

		// check memo attach redeemID req (compare hash memo)
		memo := txBNB.Memo
		memoHashBytes, err2 := base64.StdEncoding.DecodeString(memo)
		if err2 != nil {
			Logger.log.Errorf("Can not decode memo in tx bnb proof", err2)
			inst := buildReqUnlockCollateralInst(
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
			return [][]string{inst}, nil
		}

		expectedRedeemMemo := RedeemMemoBNB{
			RedeemID:                  redeemID,
			CustodianIncognitoAddress: meta.CustodianAddressStr}
		expectedRedeemMemoBytes, _ := json.Marshal(expectedRedeemMemo)
		expectedRedeemMemoHashBytes := common.HashB(expectedRedeemMemoBytes)

		if !bytes.Equal(memoHashBytes, expectedRedeemMemoHashBytes) {
			Logger.log.Errorf("Memo redeem is invalid")
			inst := buildReqUnlockCollateralInst(
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
			return [][]string{inst}, nil
		}

		// check whether amount transfer in txBNB is equal redeem amount or not
		// check receiver and amount in tx
		// get list matching custodians in matchedRedeemRequest

		outputs := txBNB.Msgs[0].(msg.SendMsg).Outputs

		remoteAddressNeedToBeTransfer := matchedRedeemRequest.GetRedeemerRemoteAddress()
		amountNeedToBeTransfer := meta.RedeemAmount
		amountNeedToBeTransferInBNB := convertIncPBNBAmountToExternalBNBAmount(int64(amountNeedToBeTransfer))

		isChecked := false
		for _, out := range outputs {
			addr, _ := bnb.GetAccAddressString(&out.Address, blockchain.config.ChainParams.BNBRelayingHeaderChainID)
			if addr != remoteAddressNeedToBeTransfer {
				continue
			}

			// calculate amount that was transferred to custodian's remote address
			amountTransfer := int64(0)
			for _, coin := range out.Coins {
				if coin.Denom == bnb.DenomBNB {
					amountTransfer += coin.Amount
					// note: log error for debug
					Logger.log.Infof("TxProof-BNB coin.Amount %d",
						coin.Amount)
				}
			}
			if amountTransfer < amountNeedToBeTransferInBNB {
				Logger.log.Errorf("TxProof-BNB is invalid - Amount transfer to %s must be equal to or greater than %d, but got %d",
					addr, amountNeedToBeTransferInBNB, amountTransfer)
				inst := buildReqUnlockCollateralInst(
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
				return [][]string{inst}, nil
			} else {
				isChecked = true
				break
			}
		}

		if !isChecked {
			Logger.log.Errorf("TxProof-BNB is invalid - Receiver address is invalid, expected %v",
				remoteAddressNeedToBeTransfer)
			inst := buildReqUnlockCollateralInst(
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
			return [][]string{inst}, nil
		}

		// calculate unlock amount
		custodianStateKey := statedb.GenerateCustodianStateObjectKey(meta.CustodianAddressStr)
		custodianStateKeyStr := custodianStateKey.String()
		unlockAmount, err2 := CalUnlockCollateralAmount(currentPortalState, custodianStateKeyStr, meta.RedeemAmount, meta.TokenID)
		if err2 != nil {
			Logger.log.Errorf("Error calculating unlock amount for custodian %v", err2)
			inst := buildReqUnlockCollateralInst(
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
			return [][]string{inst}, nil
		}

		// update custodian state (FreeCollateral, LockedAmountCollateral)
		err2 = updateCustodianStateAfterReqUnlockCollateral(
			currentPortalState.CustodianPoolState[custodianStateKeyStr],
			unlockAmount, meta.TokenID)
		if err2 != nil {
			Logger.log.Errorf("Error when updating custodian state after unlocking collateral %v", err2)
			inst := buildReqUnlockCollateralInst(
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
			return [][]string{inst}, nil
		}

		// update redeem request state in WaitingRedeemRequest (remove custodian from matchingCustodianDetail)
		updatedCustodians, err2 := removeCustodianFromMatchingRedeemCustodians(
			currentPortalState.MatchedRedeemRequests[keyMatchedRedeemRequestStr].GetCustodians(), meta.CustodianAddressStr)
		if err2 != nil {
			inst := buildReqUnlockCollateralInst(
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
			return [][]string{inst}, nil
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
	} else {
		Logger.log.Errorf("TokenID is not supported currently on Portal")
		inst := buildReqUnlockCollateralInst(
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
		return [][]string{inst}, nil
	}
}
