package blockchain

import (
	"encoding/base64"
	"encoding/json"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/metadata"
	"math/big"
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

func (blockchain *BlockChain) buildInstForPortalTotalReward(rewardInfos []*statedb.RewardInfoDetail) []string {
	portalRewardContent, _ := metadata.NewPortalTotalCustodianReward(rewardInfos)
	contentStr, _ := json.Marshal(portalRewardContent)

	inst := []string{
		strconv.Itoa(metadata.PortalTotalRewardCustodianMeta),
		strconv.Itoa(-1), // no need shardID
		"portalTotalRewardInst",
		string(contentStr),
	}

	return inst
}

func updatePortalRewardInfos(
	rewardInfos []*statedb.PortalRewardInfo,
	custodianAddress string,
	tokenID string, amount uint64) ([]*statedb.PortalRewardInfo, error) {
	for i := 0; i < len(rewardInfos); i++ {
		if rewardInfos[i].GetCustodianIncAddr() == custodianAddress {
			rewardInfos[i].AddPortalRewardInfo(tokenID, amount)
			return rewardInfos, nil
		}
	}
	rewardInfos = append(rewardInfos,
		statedb.NewPortalRewardInfoWithValue(custodianAddress,
			[]*statedb.RewardInfoDetail{statedb.NewRewardInfoDetailWithValue(tokenID, amount)}))

	return rewardInfos, nil
}

func splitPortingFeeForMatchingCustodians(
	feeAmount uint64,
	portingAmount uint64,
	matchingCustodianAddresses []*statedb.MatchingPortingCustodianDetail,
	rewardInfos []*statedb.PortalRewardInfo) []*statedb.PortalRewardInfo {
	for _, matchCustodianDetail := range matchingCustodianAddresses {
		tmp := new(big.Int).Mul(new(big.Int).SetUint64(matchCustodianDetail.Amount), new(big.Int).SetUint64(feeAmount))
		splitedFee := new(big.Int).Div(tmp, new(big.Int).SetUint64(portingAmount))
		rewardInfos, _ = updatePortalRewardInfos(rewardInfos, matchCustodianDetail.IncAddress, common.PRVIDStr, splitedFee.Uint64())
	}
	return rewardInfos
}

func splitRedeemFeeForMatchingCustodians(
	feeAmount uint64,
	redeemAmount uint64,
	matchingCustodianAddresses []*statedb.MatchingRedeemCustodianDetail,
	rewardInfos []*statedb.PortalRewardInfo) []*statedb.PortalRewardInfo {
	for _, matchCustodianDetail := range matchingCustodianAddresses {
		tmp := new(big.Int).Mul(new(big.Int).SetUint64(matchCustodianDetail.GetAmount()), new(big.Int).SetUint64(feeAmount))
		splitedFee := new(big.Int).Div(tmp, new(big.Int).SetUint64(redeemAmount))
		rewardInfos, _ = updatePortalRewardInfos(rewardInfos, matchCustodianDetail.GetIncognitoAddress(), common.PRVIDStr, splitedFee.Uint64())
	}

	return rewardInfos
}

func splitRewardForCustodians(
	totalCustodianReward map[common.Hash]uint64,
	lockedCollateralState *statedb.LockedCollateralState,
	custodianState map[string]*statedb.CustodianState,
	rewardInfos []*statedb.PortalRewardInfo) []*statedb.PortalRewardInfo {
	totalLockedCollateral := lockedCollateralState.GetTotalLockedCollateralInEpoch()
	for _, custodian := range custodianState {
		lockedCollateralCustodian := lockedCollateralState.GetLockedCollateralDetail()[custodian.GetIncognitoAddress()]
		for tokenID, amount := range totalCustodianReward {
			tmp := new(big.Int).Mul(new(big.Int).SetUint64(lockedCollateralCustodian), new(big.Int).SetUint64(amount))
			splitedReward := new(big.Int).Div(tmp, new(big.Int).SetUint64(totalLockedCollateral))
			rewardInfos, _ = updatePortalRewardInfos(rewardInfos, custodian.GetIncognitoAddress(), tokenID.String(), splitedReward.Uint64())
		}
	}
	return rewardInfos
}

func (blockchain *BlockChain) buildPortalRewardsInsts(
	beaconHeight uint64, currentPortalState *CurrentPortalState, rewardForCustodianByEpoch map[common.Hash]uint64) ([][]string, error) {

	// rewardInfos are map custodians' addresses and reward amount
	rewardInfos := make([]*statedb.PortalRewardInfo, 0)

	// get porting fee from waiting porting request at beaconHeight + 1 (new waiting porting requests)
	// and split fees for matching custodians
	for _, waitingPortingReq := range currentPortalState.WaitingPortingRequests {
		if waitingPortingReq.BeaconHeight() == beaconHeight+1 {
			rewardInfos = splitPortingFeeForMatchingCustodians(
				waitingPortingReq.PortingFee(),
				waitingPortingReq.Amount(),
				waitingPortingReq.Custodians(),
				rewardInfos,
			)
		}
	}

	// get redeem fee from waiting redeem request at beaconHeight + 1 (new waiting redeem requests)
	// and split fees for matching custodians
	for _, waitingRedeemReq := range currentPortalState.WaitingRedeemRequests {
		if waitingRedeemReq.GetBeaconHeight() == beaconHeight+1 {
			rewardInfos = splitRedeemFeeForMatchingCustodians(
				waitingRedeemReq.GetRedeemFee(),
				waitingRedeemReq.GetRedeemAmount(),
				waitingRedeemReq.GetCustodians(),
				rewardInfos,
			)
		}
	}

	// if there are reward by epoch instructions (at the end of the epoch)
	// split reward for custodians
	rewardInsts := [][]string{}
	if rewardForCustodianByEpoch != nil && len(rewardForCustodianByEpoch) > 0 {
		if currentPortalState.LockedCollateralState.GetTotalLockedCollateralInEpoch() > 0 {
			// split reward for custodians
			rewardInfos = splitRewardForCustodians(
				rewardForCustodianByEpoch,
				currentPortalState.LockedCollateralState,
				currentPortalState.CustodianPoolState,
				rewardInfos)

			// create instruction for total custodian rewards
			totalCustodianRewardSlice := make([]*statedb.RewardInfoDetail, 0)
			for tokenID, amount := range rewardForCustodianByEpoch {
				totalCustodianRewardSlice = append(totalCustodianRewardSlice,
					statedb.NewRewardInfoDetailWithValue(tokenID.String(), amount))
			}
			instTotalReward := blockchain.buildInstForPortalTotalReward(totalCustodianRewardSlice)
			rewardInsts = append(rewardInsts, instTotalReward)
		}
	}

	// update reward amount for custodian
	for custodianKey, custodianState := range currentPortalState.CustodianPoolState {
		custodianAddr := custodianState.GetIncognitoAddress()
		custodianReward := custodianState.GetRewardAmount()
		if custodianReward == nil {
			custodianReward = map[string]uint64{}
		}
		for _, rewardInfo := range rewardInfos {
			if rewardInfo.GetCustodianIncAddr() == custodianAddr {
				for _, rewardDetail := range rewardInfo.GetRewards() {
					custodianReward[rewardDetail.GetTokenID()] += rewardDetail.GetAmount()
				}
				break
			}
		}
		currentPortalState.CustodianPoolState[custodianKey].SetRewardAmount(custodianReward)
	}

	inst := blockchain.buildInstForPortalReward(beaconHeight+1, rewardInfos)
	rewardInsts = append(rewardInsts, inst)

	return rewardInsts, nil
}

// beacon build new instruction from instruction received from ShardToBeaconBlock
func buildWithdrawPortalRewardInst(
	custodianAddressStr string,
	tokenID common.Hash,
	rewardAmount uint64,
	metaType int,
	shardID byte,
	txReqID common.Hash,
	status string,
) []string {
	withdrawRewardContent := metadata.PortalRequestWithdrawRewardContent{
		CustodianAddressStr: custodianAddressStr,
		TokenID:             tokenID,
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
			actionData.Meta.TokenID,
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
	keyCustodianStateStr := keyCustodianState.String()
	custodian := currentPortalState.CustodianPoolState[keyCustodianStateStr]
	if custodian == nil {
		Logger.log.Warn("WARN - [buildInstructionsForReqWithdrawPortalReward]: Not found custodian address in custodian pool.")
		inst := buildWithdrawPortalRewardInst(
			actionData.Meta.CustodianAddressStr,
			actionData.Meta.TokenID,
			0,
			actionData.Meta.Type,
			shardID,
			actionData.TxReqID,
			common.PortalReqWithdrawRewardRejectedChainStatus,
		)
		return [][]string{inst}, nil
	} else {
		rewardAmounts := custodian.GetRewardAmount()
		rewardAmount := rewardAmounts[actionData.Meta.TokenID.String()]

		if rewardAmount <= 0 {
			Logger.log.Warn("WARN - [buildInstructionsForReqWithdrawPortalReward]: Reward amount of custodian %v is zero.", meta.CustodianAddressStr)
			inst := buildWithdrawPortalRewardInst(
				actionData.Meta.CustodianAddressStr,
				actionData.Meta.TokenID,
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
			actionData.Meta.TokenID,
			rewardAmount,
			actionData.Meta.Type,
			shardID,
			actionData.TxReqID,
			common.PortalReqWithdrawRewardAcceptedChainStatus,
		)

		// update reward amount of custodian
		updatedRewardAmount := custodian.GetRewardAmount()
		updatedRewardAmount[actionData.Meta.TokenID.String()] = 0
		currentPortalState.CustodianPoolState[keyCustodianStateStr].SetRewardAmount(updatedRewardAmount)
		return [][]string{inst}, nil
	}
}
