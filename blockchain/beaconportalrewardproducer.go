package blockchain

import (
	"encoding/base64"
	"encoding/json"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/database/lvdb"
	"github.com/incognitochain/incognito-chain/metadata"
	"strconv"
)

func (blockchain *BlockChain) buildInstForPortalReward(beaconHeight uint64, receivers map[string]uint64) []string {
	portalRewardContent, _ := metadata.NewPortalReward(beaconHeight, receivers)
	contentStr, _ := json.Marshal(portalRewardContent)

	inst := []string{
		strconv.Itoa(metadata.PortalRewardMeta),
		strconv.Itoa(-1),		// no need shardID
		"portalRewardInst",
		string(contentStr),
	}

	return inst
}

func splitPortingFeeForMatchingCustodians(
	feeAmount uint64,
	portingAmount uint64,
	matchingCustodianAddresses map[string]lvdb.MatchingPortingCustodianDetail,
	rewardInfos map[string]uint64) {
	for incAddr, matchCustodianDetail := range matchingCustodianAddresses {
		splitedFee := feeAmount * matchCustodianDetail.Amount / portingAmount
		rewardInfos[incAddr] += splitedFee
	}
}

func splitRedeemFeeForMatchingCustodians(
	feeAmount uint64,
	redeemAmount uint64,
	matchingCustodianAddresses map[string]*lvdb.MatchingRedeemCustodianDetail,
	rewardInfos map[string]uint64) {
	for incAddr, matchCustodianDetail := range matchingCustodianAddresses {
		splitedFee := feeAmount * matchCustodianDetail.Amount / redeemAmount
		rewardInfos[incAddr] += splitedFee
	}
}

func splitRewardForCustodians(
	totalReward uint64,
	totalLockedAmount uint64,
	custodianState map[string]*lvdb.CustodianState,
	rewardInfos map[string]uint64)  {
	for _, custodian := range custodianState {
		for _, lockedAmount := range custodian.LockedAmountCollateral {
			splitedReward := totalReward * lockedAmount / totalLockedAmount
			rewardInfos[custodian.IncognitoAddress] += splitedReward
		}
	}
}

func (blockchain *BlockChain) buildPortalRewardsInsts(
	beaconHeight uint64, currentPortalState *CurrentPortalState) ([][]string, error) {

	// receivers are map custodians' addresses and reward amount
	receivers := make(map[string]uint64, 0)

	// get porting fee from waiting porting request at beaconHeight + 1 (new waiting porting requests)
	// and split fees for matching custodians
	for _, waitingPortingReq := range currentPortalState.WaitingPortingRequests {
		if waitingPortingReq.BeaconHeight == beaconHeight + 1 {
			splitPortingFeeForMatchingCustodians(
				waitingPortingReq.PortingFee,
				waitingPortingReq.Amount,
				waitingPortingReq.Custodians,
				receivers,
			)
		}
	}

	// get redeem fee from waiting redeem request at beaconHeight + 1 (new waiting redeem requests)
	// and split fees for matching custodians
	for _, waitingRedeemReq := range currentPortalState.WaitingRedeemRequests {
		if waitingRedeemReq.BeaconHeight == beaconHeight + 1 {
			splitRedeemFeeForMatchingCustodians(
				waitingRedeemReq.RedeemFee,
				waitingRedeemReq.RedeemAmount,
				waitingRedeemReq.Custodians,
				receivers,
			)
		}
	}

	// calculate rewards corresponding to locked amount collateral for each custodians
	// calculate total holding amount for each public tokens
	totalLockedCollateralAmount := uint64(0)
	for _, custodianState := range currentPortalState.CustodianPoolState {
		for _, lockedAmount := range custodianState.LockedAmountCollateral {
			totalLockedCollateralAmount += lockedAmount
		}
	}

	if totalLockedCollateralAmount > 0 {
		// split reward amount
		splitRewardForCustodians(common.TotalRewardPerBlock, totalLockedCollateralAmount, currentPortalState.CustodianPoolState, receivers)
	}

	// update reward amount for each custodian
	for _, custodianState := range currentPortalState.CustodianPoolState {
		custodianState.RewardAmount += receivers[custodianState.IncognitoAddress]
	}

	// build beacon instruction for portal reward
	inst := blockchain.buildInstForPortalReward(beaconHeight + 1, receivers)

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
		RewardAmount: rewardAmount,
		TxReqID:         txReqID,
		ShardID:         shardID,
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

	keyCustodianState := lvdb.NewCustodianStateKey(beaconHeight, meta.CustodianAddressStr)
	custodian := currentPortalState.CustodianPoolState[keyCustodianState]
	if custodian == nil {
		Logger.log.Warn("WARN - [buildInstructionsForReqWithdrawPortalReward]: Not found custodian address in custodian pool.")
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
		rewardAmount := custodian.RewardAmount
		if rewardAmount <= 0 {
			Logger.log.Warn("WARN - [buildInstructionsForReqWithdrawPortalReward]: Reward amount of custodian %v is zero.", meta.CustodianAddressStr)
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

		inst := buildWithdrawPortalRewardInst(
			actionData.Meta.CustodianAddressStr,
			rewardAmount,
			actionData.Meta.Type,
			shardID,
			actionData.TxReqID,
			common.PortalReqWithdrawRewardAcceptedChainStatus,
		)

		// update reward amount of custodian
		custodian.RewardAmount = 0
		return [][]string{inst}, nil
	}
}