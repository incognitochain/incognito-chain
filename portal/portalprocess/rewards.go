package portalprocess

import (
	"encoding/base64"
	"encoding/json"
	"github.com/incognitochain/incognito-chain/basemeta"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/incognitochain/incognito-chain/portal"
	metadata2 "github.com/incognitochain/incognito-chain/portal/metadata"
	"math/big"
	"sort"
	"strconv"
)

func buildInstForPortalReward(beaconHeight uint64, rewardInfos map[string]*statedb.PortalRewardInfo) []string {
	portalRewardContent := metadata.NewPortalReward(beaconHeight, rewardInfos)
	contentStr, _ := json.Marshal(portalRewardContent)

	inst := []string{
		strconv.Itoa(basemeta.PortalRewardMetaV3),
		strconv.Itoa(-1), // no need shardID
		"portalRewardInst",
		string(contentStr),
	}

	return inst
}

func buildInstForPortalTotalReward(rewardInfos map[string]uint64) []string {
	portalRewardContent := metadata.NewPortalTotalCustodianReward(rewardInfos)
	contentStr, _ := json.Marshal(portalRewardContent)

	inst := []string{
		strconv.Itoa(basemeta.PortalTotalRewardCustodianMeta),
		strconv.Itoa(-1), // no need shardID
		"portalTotalRewardInst",
		string(contentStr),
	}

	return inst
}

func updatePortalRewardInfos(
	rewardInfos map[string]*statedb.PortalRewardInfo,
	custodianAddress string,
	tokenID string, amount uint64) map[string]*statedb.PortalRewardInfo {
	if rewardInfos == nil {
		rewardInfos = make(map[string]*statedb.PortalRewardInfo)
	}
	if rewardInfos[custodianAddress] == nil {
		rewardInfos[custodianAddress] = new(statedb.PortalRewardInfo)
	}

	rewardInfos[custodianAddress].AddPortalRewardInfo(tokenID, amount)
	return rewardInfos
}

func splitPortingFeeForMatchingCustodians(
	feeAmount uint64,
	portingAmount uint64,
	matchingCustodianAddresses []*statedb.MatchingPortingCustodianDetail,
	rewardInfos map[string]*statedb.PortalRewardInfo) map[string]*statedb.PortalRewardInfo {
	totalSplitFee := uint64(0)
	for i, matchCustodianDetail := range matchingCustodianAddresses {
		tmp := new(big.Int).Mul(new(big.Int).SetUint64(matchCustodianDetail.Amount), new(big.Int).SetUint64(feeAmount))
		splitedFee := new(big.Int).Div(tmp, new(big.Int).SetUint64(portingAmount)).Uint64()
		if i == len(matchingCustodianAddresses)-1 {
			if splitedFee+totalSplitFee < feeAmount {
				splitedFee = feeAmount - totalSplitFee
			}
		}
		totalSplitFee += splitedFee
		rewardInfos = updatePortalRewardInfos(rewardInfos, matchCustodianDetail.IncAddress, common.PRVIDStr, splitedFee)
	}
	return rewardInfos
}

func splitRedeemFeeForMatchingCustodians(
	feeAmount uint64,
	redeemAmount uint64,
	matchingCustodianAddresses []*statedb.MatchingRedeemCustodianDetail,
	rewardInfos map[string]*statedb.PortalRewardInfo) map[string]*statedb.PortalRewardInfo {
	totalSplitFee := uint64(0)
	for i, matchCustodianDetail := range matchingCustodianAddresses {
		tmp := new(big.Int).Mul(new(big.Int).SetUint64(matchCustodianDetail.GetAmount()), new(big.Int).SetUint64(feeAmount))
		splitedFee := new(big.Int).Div(tmp, new(big.Int).SetUint64(redeemAmount)).Uint64()
		if i == len(matchingCustodianAddresses)-1 {
			if splitedFee+totalSplitFee < feeAmount {
				splitedFee = feeAmount - totalSplitFee
			}
		}
		totalSplitFee += splitedFee
		rewardInfos = updatePortalRewardInfos(rewardInfos, matchCustodianDetail.GetIncognitoAddress(), common.PRVIDStr, splitedFee)
	}

	return rewardInfos
}

func splitRewardForCustodians(
	totalCustodianReward map[common.Hash]uint64,
	lockedCollateralState *statedb.LockedCollateralState,
	custodianState map[string]*statedb.CustodianState,
	rewardInfos map[string]*statedb.PortalRewardInfo) map[string]*statedb.PortalRewardInfo {
	totalLockedCollateral := lockedCollateralState.GetTotalLockedCollateralForRewards()

	totalLockCollateralSplited := uint64(0)
	totalRewardSplited := map[common.Hash]uint64{}

	sortedCustodianKeys := []string{}
	for custodianKey := range custodianState {
		sortedCustodianKeys = append(sortedCustodianKeys, custodianKey)
	}
	sort.Strings(sortedCustodianKeys)
	for _, key := range sortedCustodianKeys {
		custodian := custodianState[key]
		lockedCollateralCustodian, ok := lockedCollateralState.GetLockedCollateralDetail()[custodian.GetIncognitoAddress()]
		if !ok || lockedCollateralCustodian == 0 {
			continue
		}

		isFinalCustodian := false
		if totalLockCollateralSplited+lockedCollateralCustodian == totalLockedCollateral {
			isFinalCustodian = true
		}

		for tokenID, amount := range totalCustodianReward {
			tmp := new(big.Int).Mul(new(big.Int).SetUint64(lockedCollateralCustodian), new(big.Int).SetUint64(amount))
			splitedReward := new(big.Int).Div(tmp, new(big.Int).SetUint64(totalLockedCollateral)).Uint64()
			if isFinalCustodian {
				if splitedReward+totalRewardSplited[tokenID] < amount {
					splitedReward = amount - totalRewardSplited[tokenID]
				}
			}
			rewardInfos = updatePortalRewardInfos(rewardInfos, custodian.GetIncognitoAddress(), tokenID.String(), splitedReward)
			totalRewardSplited[tokenID] += splitedReward
		}
		totalLockCollateralSplited += lockedCollateralCustodian
	}
	return rewardInfos
}

func buildPortalRewardsInsts(
	beaconHeight uint64,
	currentPortalState *CurrentPortalState,
	rewardForCustodianByEpoch map[common.Hash]uint64,
	newMatchedRedeemReqIDs []string) ([][]string, error) {

	// rewardInfos are map custodians' addresses and reward amount
	rewardInfos := make(map[string]*statedb.PortalRewardInfo, 0)

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

	// get new matched redeem requests at beaconHeight + 1
	// and split fees for matching custodians
	for _, newMatchedRedeemID := range newMatchedRedeemReqIDs {
		matchedRedeemKey := statedb.GenerateMatchedRedeemRequestObjectKey(newMatchedRedeemID).String()
		matchedRedeemReq, ok := currentPortalState.MatchedRedeemRequests[matchedRedeemKey]
		if !ok || matchedRedeemReq == nil {
			continue
		}
		rewardInfos = splitRedeemFeeForMatchingCustodians(
			matchedRedeemReq.GetRedeemFee(),
			matchedRedeemReq.GetRedeemAmount(),
			matchedRedeemReq.GetCustodians(),
			rewardInfos,
		)
	}

	// if there are reward by epoch instructions (at the end of the epoch)
	// split reward for custodians
	rewardInsts := [][]string{}
	if rewardForCustodianByEpoch != nil && len(rewardForCustodianByEpoch) > 0 {
		if currentPortalState.LockedCollateralForRewards.GetTotalLockedCollateralForRewards() > 0 {
			Logger.log.Infof("buildPortalRewardsInsts rewardForCustodianByEpoch %v", rewardForCustodianByEpoch)
			// split reward for custodians
			rewardInfos = splitRewardForCustodians(
				rewardForCustodianByEpoch,
				currentPortalState.LockedCollateralForRewards,
				currentPortalState.CustodianPoolState,
				rewardInfos)

			// create instruction for total custodian rewards
			totalRewardForCustodians := make(map[string]uint64, 0)

			// sort totalCustodianReward before processing
			sortedCustodianRewardKeys := make([]common.Hash, 0)
			for key := range rewardForCustodianByEpoch {
				sortedCustodianRewardKeys = append(sortedCustodianRewardKeys, key)
			}
			sort.Slice(sortedCustodianRewardKeys, func(i, j int) bool {
				return sortedCustodianRewardKeys[i].String() < sortedCustodianRewardKeys[j].String()
			})
			for _, key := range sortedCustodianRewardKeys {
				tokenID := key
				amount := rewardForCustodianByEpoch[key]
				totalRewardForCustodians[tokenID.String()] = amount
			}
			instTotalReward := buildInstForPortalTotalReward(totalRewardForCustodians)
			rewardInsts = append(rewardInsts, instTotalReward)
		}
	}

	// update reward amount for custodian
	UpdateCustodianRewards(currentPortalState, rewardInfos)

	sortedRewardInfos := make(map[string]*statedb.PortalRewardInfo, 0)
	custodianAddrKeys := make([]string, 0)

	for custodianAddr := range rewardInfos {
		custodianAddrKeys = append(custodianAddrKeys, custodianAddr)
	}
	sort.Strings(custodianAddrKeys)
	for _, key := range custodianAddrKeys {
		rewardInfo := rewardInfos[key].GetRewards()
		tokenIDs := make([]string, 0)
		for tokenID := range rewardInfo {
			tokenIDs = append(tokenIDs, tokenID)
		}
		sort.Strings(tokenIDs)
		rewardInfoTmp := map[string]uint64{}
		for _, tokenID := range tokenIDs {
			rewardInfoTmp[tokenID] = rewardInfo[tokenID]
		}

		sortedRewardInfos[key] = new(statedb.PortalRewardInfo)
		sortedRewardInfos[key].SetRewards(rewardInfoTmp)
	}
	inst := buildInstForPortalReward(beaconHeight+1, sortedRewardInfos)
	rewardInsts = append(rewardInsts, inst)

	return rewardInsts, nil
}


/* =======
Portal Custodian Request Withdraw Rewards Processor
======= */

type portalReqWithdrawRewardProcessor struct {
	*portalInstProcessor
}

func (p *portalReqWithdrawRewardProcessor) GetActions() map[byte][][]string {
	return p.actions
}

func (p *portalReqWithdrawRewardProcessor) PutAction(action []string, shardID byte) {
	_, found := p.actions[shardID]
	if !found {
		p.actions[shardID] = [][]string{action}
	} else {
		p.actions[shardID] = append(p.actions[shardID], action)
	}
}

func (p *portalReqWithdrawRewardProcessor) PrepareDataForBlockProducer(stateDB *statedb.StateDB, contentStr string) (map[string]interface{}, error) {
	return nil, nil
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
	withdrawRewardContent := metadata2.PortalRequestWithdrawRewardContent{
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

func (p *portalReqWithdrawRewardProcessor) BuildNewInsts(
	bc basemeta.ChainRetriever,
	contentStr string,
	shardID byte,
	currentPortalState *CurrentPortalState,
	beaconHeight uint64,
	shardHeights map[byte]uint64,
	portalParams portal.PortalParams,
	optionalData map[string]interface{},
) ([][]string, error) {
	// parse instruction
	actionContentBytes, err := base64.StdEncoding.DecodeString(contentStr)
	if err != nil {
		Logger.log.Errorf("ERROR: an error occured while decoding content string of portal custodian deposit action: %+v", err)
		return [][]string{}, nil
	}
	var actionData metadata2.PortalRequestWithdrawRewardAction
	err = json.Unmarshal(actionContentBytes, &actionData)
	if err != nil {
		Logger.log.Errorf("ERROR: an error occured while unmarshal portal custodian deposit action: %+v", err)
		return [][]string{}, nil
	}

	rejectInst := buildWithdrawPortalRewardInst(
		actionData.Meta.CustodianAddressStr,
		actionData.Meta.TokenID,
		0,
		actionData.Meta.Type,
		shardID,
		actionData.TxReqID,
		common.PortalReqWithdrawRewardRejectedChainStatus,
	)

	if currentPortalState == nil {
		Logger.log.Warn("WARN - [buildInstructionsForReqWithdrawPortalReward]: Current Portal state is null.")
		return [][]string{rejectInst}, nil
	}
	meta := actionData.Meta

	keyCustodianState := statedb.GenerateCustodianStateObjectKey(meta.CustodianAddressStr)
	keyCustodianStateStr := keyCustodianState.String()
	custodian := currentPortalState.CustodianPoolState[keyCustodianStateStr]
	if custodian == nil {
		Logger.log.Warn("WARN - [buildInstructionsForReqWithdrawPortalReward]: Not found custodian address in custodian pool.")
		return [][]string{rejectInst}, nil
	} else {
		rewardAmounts := custodian.GetRewardAmount()
		rewardAmount := rewardAmounts[actionData.Meta.TokenID.String()]

		if rewardAmount <= 0 {
			Logger.log.Warn("WARN - [buildInstructionsForReqWithdrawPortalReward]: Reward amount of custodian %v is zero.", meta.CustodianAddressStr)
			return [][]string{rejectInst}, nil
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

func (p *portalReqWithdrawRewardProcessor) ProcessInsts(
	stateDB *statedb.StateDB,
	beaconHeight uint64,
	instructions []string,
	currentPortalState *CurrentPortalState,
	portalParams portal.PortalParams,
	updatingInfoByTokenID map[common.Hash]basemeta.UpdatingInfo,
) error {
	// unmarshal instructions content
	var actionData metadata2.PortalRequestWithdrawRewardContent
	err := json.Unmarshal([]byte(instructions[3]), &actionData)
	if err != nil {
		Logger.log.Errorf("Can not unmarshal instruction content %v - Error %v\n", instructions[3], err)
		return nil
	}

	reqStatus := instructions[2]
	if reqStatus == common.PortalReqWithdrawRewardAcceptedChainStatus {
		// update reward amount of custodian
		cusStateKey := statedb.GenerateCustodianStateObjectKey(actionData.CustodianAddressStr)
		cusStateKeyStr := cusStateKey.String()
		custodianState := currentPortalState.CustodianPoolState[cusStateKeyStr]
		if custodianState == nil {
			Logger.log.Errorf("[processPortalWithdrawReward] Can not get custodian state with key %v", cusStateKey)
			return nil
		}
		updatedRewardAmount := custodianState.GetRewardAmount()
		updatedRewardAmount[actionData.TokenID.String()] = 0
		currentPortalState.CustodianPoolState[cusStateKeyStr].SetRewardAmount(updatedRewardAmount)

		// track request withdraw portal reward
		portalReqRewardStatus := metadata2.PortalRequestWithdrawRewardStatus{
			Status:              common.PortalReqWithdrawRewardAcceptedStatus,
			CustodianAddressStr: actionData.CustodianAddressStr,
			TokenID:             actionData.TokenID,
			RewardAmount:        actionData.RewardAmount,
			TxReqID:             actionData.TxReqID,
		}
		portalReqRewardStatusBytes, _ := json.Marshal(portalReqRewardStatus)
		err = statedb.StorePortalRequestWithdrawRewardStatus(
			stateDB,
			actionData.TxReqID.String(),
			portalReqRewardStatusBytes,
		)
		if err != nil {
			Logger.log.Errorf("ERROR: an error occured while tracking liquidation custodian: %+v", err)
			return nil
		}

	} else if reqStatus == common.PortalReqUnlockCollateralRejectedChainStatus {
		// track request withdraw portal reward
		portalReqRewardStatus := metadata2.PortalRequestWithdrawRewardStatus{
			Status:              common.PortalReqWithdrawRewardRejectedStatus,
			CustodianAddressStr: actionData.CustodianAddressStr,
			TokenID:             actionData.TokenID,
			RewardAmount:        actionData.RewardAmount,
			TxReqID:             actionData.TxReqID,
		}
		portalReqRewardStatusBytes, _ := json.Marshal(portalReqRewardStatus)
		err = statedb.StorePortalRequestWithdrawRewardStatus(
			stateDB,
			actionData.TxReqID.String(),
			portalReqRewardStatusBytes,
		)
		if err != nil {
			Logger.log.Errorf("ERROR: an error occured while tracking liquidation custodian: %+v", err)
			return nil
		}
	}

	return nil
}

/*
Portal reward process
 */

func ProcessPortalReward(
	stateDB *statedb.StateDB,
	beaconHeight uint64, instructions []string,
	currentPortalState *CurrentPortalState,
	portalParams portal.PortalParams, epoch uint64) error {

	// unmarshal instructions content
	var actionData metadata.PortalRewardContent
	err := json.Unmarshal([]byte(instructions[3]), &actionData)
	if err != nil {
		Logger.log.Errorf("Can not unmarshal instruction content %v - Error %v\n", instructions[3], err)
		return nil
	}

	reqStatus := instructions[2]
	metaType, _ := strconv.Atoi(instructions[0])
	if reqStatus == "portalRewardInst" {
		// update reward amount for custodian
		UpdateCustodianRewards(currentPortalState, actionData.Rewards)

		// at the end of epoch
		if (beaconHeight+1)%epoch == 1 {
			currentPortalState.LockedCollateralForRewards.Reset()
		}

		// update locked collateral for rewards base on holding public tokens
		if metaType == basemeta.PortalRewardMetaV3 {
			UpdateLockedCollateralForRewardsV3(currentPortalState, portalParams)
		} else if metaType == basemeta.PortalRewardMeta {
			UpdateLockedCollateralForRewards(currentPortalState, portalParams)
		}

		// store reward at beacon height into db
		err = statedb.StorePortalRewards(
			stateDB,
			beaconHeight+1,
			actionData.Rewards,
		)
		if err != nil {
			Logger.log.Errorf("ERROR: an error occured while tracking liquidation custodian: %+v", err)
			return nil
		}
	} else {
		Logger.log.Errorf("ERROR: Invalid status of instruction: %+v", reqStatus)
		return nil
	}

	return nil
}

func ProcessPortalTotalCustodianReward(
	stateDB *statedb.StateDB,
	beaconHeight uint64, instructions []string,
	currentPortalState *CurrentPortalState,
	portalParams portal.PortalParams,
	epoch uint64) error {

	// unmarshal instructions content
	var actionData metadata.PortalTotalCustodianReward
	err := json.Unmarshal([]byte(instructions[3]), &actionData)
	if err != nil {
		Logger.log.Errorf("Can not unmarshal instruction content %v - Error %v\n", instructions[3], err)
		return nil
	}

	reqStatus := instructions[2]
	if reqStatus == "portalTotalRewardInst" {
		epoch := beaconHeight / epoch
		// store total custodian reward into db
		err = statedb.StoreRewardFeatureState(
			stateDB,
			statedb.PortalRewardName,
			actionData.Rewards,
			epoch,
		)
		if err != nil {
			Logger.log.Errorf("ERROR: an error occured while storing total custodian reward: %+v", err)
			return nil
		}
	} else {
		Logger.log.Errorf("ERROR: Invalid status of instruction: %+v", reqStatus)
		return nil
	}

	return nil
}




