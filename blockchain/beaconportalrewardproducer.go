package blockchain

import (
	"encoding/base64"
	"encoding/json"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/metadata"
	"math"
	"sort"
	"strconv"
)

func (blockchain *BlockChain) buildInstForPortalReward(beaconHeight uint64, rewardInfos []*statedb.PortalRewardInfo) []string {
	portalRewardContent, _ := metadata.NewPortalReward(beaconHeight, rewardInfos)
	contentStr, _ := json.Marshal(portalRewardContent)

	inst := []string{
		strconv.Itoa(metadata.PortalRewardMeta),
		strconv.Itoa(-1), // no need shardID
		"portalRewardInst",
		string(contentStr),
	}

	return inst
}

func splitPortingFeeForMatchingCustodians(
	feeAmount uint64,
	portingAmount uint64,
	matchingCustodianAddresses []*lvdb.MatchingPortingCustodianDetail,
	rewardInfos map[string]uint64) {
	for _, matchCustodianDetail := range matchingCustodianAddresses {
		splitedFee := float64(matchCustodianDetail.Amount) / float64(portingAmount) * float64(feeAmount)
		rewardInfos[matchCustodianDetail.IncAddress] += uint64(math.Floor(splitedFee))
	}
}

func splitRedeemFeeForMatchingCustodians(
	feeAmount uint64,
	redeemAmount uint64,
	matchingCustodianAddresses []*statedb.MatchingRedeemCustodianDetail,
	rewardInfos map[string]uint64) {
	for _, matchCustodianDetail := range matchingCustodianAddresses {
		splitedFee := float64(matchCustodianDetail.GetAmount()) / float64(redeemAmount) * float64(feeAmount)
		rewardInfos[matchCustodianDetail.GetIncognitoAddress()] += uint64(math.Floor(splitedFee))
	}
}

func splitRewardForCustodians(
	totalReward uint64,
	totalLockedAmount uint64,
	custodianState map[string]*statedb.CustodianState,
	rewardInfos map[string]uint64) {
	for _, custodian := range custodianState {
		for _, lockedAmount := range custodian.GetLockedAmountCollateral() {
			splitedReward := float64(lockedAmount) / float64(totalLockedAmount) * float64(totalReward)
			rewardInfos[custodian.GetIncognitoAddress()] += uint64(math.Floor(splitedReward))
		}
	}
}

func (blockchain *BlockChain) buildPortalRewardsInsts(
	beaconHeight uint64, currentPortalState *CurrentPortalState) ([][]string, error) {

	// rewardInfos are map custodians' addresses and reward amount
	//rewardInfos := make([]*lvdb.PortalRewardInfo, 0)
	rewardInfos := make(map[string]uint64, 0)

	// get porting fee from waiting porting request at beaconHeight + 1 (new waiting porting requests)
	// and split fees for matching custodians
	for _, waitingPortingReq := range currentPortalState.WaitingPortingRequests {
		if waitingPortingReq.BeaconHeight == beaconHeight+1 {
			splitPortingFeeForMatchingCustodians(
				waitingPortingReq.PortingFee,
				waitingPortingReq.Amount,
				waitingPortingReq.Custodians,
				rewardInfos,
			)
		}
	}

	// get redeem fee from waiting redeem request at beaconHeight + 1 (new waiting redeem requests)
	// and split fees for matching custodians
	for _, waitingRedeemReq := range currentPortalState.WaitingRedeemRequests {
		if waitingRedeemReq.GetBeaconHeight() == beaconHeight+1 {
			splitRedeemFeeForMatchingCustodians(
				waitingRedeemReq.GetRedeemFee(),
				waitingRedeemReq.GetRedeemAmount(),
				waitingRedeemReq.GetCustodians(),
				rewardInfos,
			)
		}
	}

	// calculate rewards corresponding to locked amount collateral for each custodians
	// calculate total holding amount for each public tokens
	totalLockedCollateralAmount := uint64(0)
	for _, custodianState := range currentPortalState.CustodianPoolState {
		for _, lockedAmount := range custodianState.GetLockedAmountCollateral() {
			totalLockedCollateralAmount += lockedAmount
		}
	}

	if totalLockedCollateralAmount > 0 {
		// split reward amount
		splitRewardForCustodians(common.TotalRewardPerBlock, totalLockedCollateralAmount, currentPortalState.CustodianPoolState, rewardInfos)
	}

	// update reward amount for each custodian
	rewardInfoKeys := []string{}
	for custodianAddr, amount := range rewardInfos {
		for _, custodianState := range currentPortalState.CustodianPoolState {
			if custodianAddr == custodianState.GetIncognitoAddress() {
				custodianState.SetRewardAmount(custodianState.GetRewardAmount() + amount)
				break
			}
		}

		rewardInfoKeys = append(rewardInfoKeys, custodianAddr)
	}

	// build beacon instruction for portal reward
	// sort rewardInfos by custodian address before creating instruction
	sort.Strings(rewardInfoKeys)
	sortedRewardInfos := make([]*statedb.PortalRewardInfo, len(rewardInfos))
	for i, custodianAddr := range rewardInfoKeys {
		sortedRewardInfos[i] = statedb.NewPortalRewardInfoWithValue(custodianAddr, rewardInfos[custodianAddr])
	}
	inst := blockchain.buildInstForPortalReward(beaconHeight+1, sortedRewardInfos)

	return [][]string{inst}, nil
}

// beacon build new instruction from instruction received from ShardToBeaconBlock
func buildWithdrawPortalRewardInst(
	custodianAddressStr string,
	rewardAmount uint64,
	metaType int,
	shardID byte,
	txReqID common.Hash,
	status string,
) []string {
	withdrawRewardContent := metadata.PortalRequestWithdrawRewardContent{
		CustodianAddressStr: custodianAddressStr,
		RewardAmount:        rewardAmount,
		TxReqID:             txReqID,
		ShardID:             shardID,
	}
	withdrawRewardContentBytes, _ := json.Marshal(withdrawRewardContent)
	return []string{
		strconv.Itoa(metaType),
		strconv.Itoa(int(shardID)),
		status,
		string(withdrawRewardContentBytes),
	}
}

// buildInstructionsForCustodianDeposit builds instruction for custodian deposit action
func (blockchain *BlockChain) buildInstructionsForReqWithdrawPortalReward(
	contentStr string,
	shardID byte,
	metaType int,
	currentPortalState *CurrentPortalState,
	beaconHeight uint64,
) ([][]string, error) {
	Logger.log.Errorf("[buildInstructionsForReqWithdrawPortalReward] Starting....")
	// parse instruction
	actionContentBytes, err := base64.StdEncoding.DecodeString(contentStr)
	if err != nil {
		Logger.log.Errorf("ERROR: an error occured while decoding content string of portal custodian deposit action: %+v", err)
		return [][]string{}, nil
	}
	var actionData metadata.PortalRequestWithdrawRewardAction
	err = json.Unmarshal(actionContentBytes, &actionData)
	if err != nil {
		Logger.log.Errorf("ERROR: an error occured while unmarshal portal custodian deposit action: %+v", err)
		return [][]string{}, nil
	}

	if currentPortalState == nil {
		Logger.log.Warn("WARN - [buildInstructionsForReqWithdrawPortalReward]: Current Portal state is null.")
		// need to refund collateral to custodian
		inst := buildWithdrawPortalRewardInst(
			actionData.Meta.CustodianAddressStr,
			0,
			actionData.Meta.Type,
			shardID,
			actionData.TxReqID,
			common.PortalReqWithdrawRewardRejectedChainStatus,
		)
		return [][]string{inst}, nil
	}
	meta := actionData.Meta

	keyCustodianState := statedb.GenerateCustodianStateObjectKey(beaconHeight, meta.CustodianAddressStr)
	keyCustodianStateStr := string(keyCustodianState[:])
	custodian := currentPortalState.CustodianPoolState[keyCustodianStateStr]
	if custodian == nil {
		Logger.log.Warn("WARN - [buildInstructionsForReqWithdrawPortalReward]: Not found custodian address in custodian pool.")
		Logger.log.Errorf("[buildInstructionsForReqWithdrawPortalReward] Rejected....")
		inst := buildWithdrawPortalRewardInst(
			actionData.Meta.CustodianAddressStr,
			0,
			actionData.Meta.Type,
			shardID,
			actionData.TxReqID,
			common.PortalReqWithdrawRewardRejectedChainStatus,
		)
		return [][]string{inst}, nil
	} else {
		rewardAmount := custodian.GetRewardAmount()
		if rewardAmount <= 0 {
			Logger.log.Warn("WARN - [buildInstructionsForReqWithdrawPortalReward]: Reward amount of custodian %v is zero.", meta.CustodianAddressStr)
			Logger.log.Errorf("[buildInstructionsForReqWithdrawPortalReward] Rejected....")
			inst := buildWithdrawPortalRewardInst(
				actionData.Meta.CustodianAddressStr,
				0,
				actionData.Meta.Type,
				shardID,
				actionData.TxReqID,
				common.PortalReqWithdrawRewardRejectedChainStatus,
			)
			return [][]string{inst}, nil
		}

		Logger.log.Errorf("[buildInstructionsForReqWithdrawPortalReward] Accepted....")
		inst := buildWithdrawPortalRewardInst(
			actionData.Meta.CustodianAddressStr,
			rewardAmount,
			actionData.Meta.Type,
			shardID,
			actionData.TxReqID,
			common.PortalReqWithdrawRewardAcceptedChainStatus,
		)

		// update reward amount of custodian
		custodian.SetRewardAmount( 0)
		return [][]string{inst}, nil
	}
}
