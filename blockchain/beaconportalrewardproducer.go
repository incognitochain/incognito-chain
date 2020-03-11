package blockchain

import (
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