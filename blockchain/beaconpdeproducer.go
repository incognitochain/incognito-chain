package blockchain

import (
	"encoding/base64"
	"encoding/json"
	"math/big"
	"strconv"

	"github.com/incognitochain/incognito-chain/dataaccessobject/rawdbv2"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/metadata"
)

func buildWaitingContributionInst(
	pdeContributionPairID string,
	contributorAddressStr string,
	contributedAmount uint64,
	tokenIDStr string,
	metaType int,
	shardID byte,
	txReqID common.Hash,
) []string {
	waitingContribution := metadata.PDEWaitingContribution{
		PDEContributionPairID: pdeContributionPairID,
		ContributorAddressStr: contributorAddressStr,
		ContributedAmount:     contributedAmount,
		TokenIDStr:            tokenIDStr,
		TxReqID:               txReqID,
	}
	waitingContributionBytes, _ := json.Marshal(waitingContribution)
	return []string{
		strconv.Itoa(metaType),
		strconv.Itoa(int(shardID)),
		common.PDEContributionWaitingChainStatus,
		string(waitingContributionBytes),
	}
}

func buildRefundContributionInst(
	pdeContributionPairID string,
	contributorAddressStr string,
	contributedAmount uint64,
	tokenIDStr string,
	metaType int,
	shardID byte,
	txReqID common.Hash,
) []string {
	refundContribution := metadata.PDERefundContribution{
		PDEContributionPairID: pdeContributionPairID,
		ContributorAddressStr: contributorAddressStr,
		ContributedAmount:     contributedAmount,
		TokenIDStr:            tokenIDStr,
		TxReqID:               txReqID,
		ShardID:               shardID,
	}
	refundContributionBytes, _ := json.Marshal(refundContribution)
	return []string{
		strconv.Itoa(metaType),
		strconv.Itoa(int(shardID)),
		common.PDEContributionRefundChainStatus,
		string(refundContributionBytes),
	}
}

func buildMatchedContributionInst(
	pdeContributionPairID string,
	contributorAddressStr string,
	contributedAmount uint64,
	tokenIDStr string,
	metaType int,
	shardID byte,
	txReqID common.Hash,
) []string {
	matchedContribution := metadata.PDEMatchedContribution{
		PDEContributionPairID: pdeContributionPairID,
		ContributorAddressStr: contributorAddressStr,
		ContributedAmount:     contributedAmount,
		TokenIDStr:            tokenIDStr,
		TxReqID:               txReqID,
	}
	matchedContributionBytes, _ := json.Marshal(matchedContribution)
	return []string{
		strconv.Itoa(metaType),
		strconv.Itoa(int(shardID)),
		common.PDEContributionMatchedChainStatus,
		string(matchedContributionBytes),
	}
}

func buildMatchedNReturnedContributionInst(
	pdeContributionPairID string,
	contributorAddressStr string,
	actualContributedAmount uint64,
	returnedContributedAmount uint64,
	tokenIDStr string,
	metaType int,
	shardID byte,
	txReqID common.Hash,
	actualWaitingContribAmount uint64,
) []string {
	matchedNReturnedContribution := metadata.PDEMatchedNReturnedContribution{
		PDEContributionPairID:      pdeContributionPairID,
		ContributorAddressStr:      contributorAddressStr,
		ActualContributedAmount:    actualContributedAmount,
		ReturnedContributedAmount:  returnedContributedAmount,
		TokenIDStr:                 tokenIDStr,
		ShardID:                    shardID,
		TxReqID:                    txReqID,
		ActualWaitingContribAmount: actualWaitingContribAmount,
	}
	matchedNReturnedContribBytes, _ := json.Marshal(matchedNReturnedContribution)
	return []string{
		strconv.Itoa(metaType),
		strconv.Itoa(int(shardID)),
		common.PDEContributionMatchedNReturnedChainStatus,
		string(matchedNReturnedContribBytes),
	}
}

func isRightRatio(
	waitingContribution1 *rawdbv2.PDEContribution,
	waitingContribution2 *rawdbv2.PDEContribution,
	poolPair *rawdbv2.PDEPoolForPair,
) bool {
	if poolPair == nil {
		return true
	}
	if poolPair.Token1PoolValue == 0 || poolPair.Token2PoolValue == 0 {
		return true
	}
	if waitingContribution1.TokenIDStr == poolPair.Token1IDStr {
		expectedContribAmt := big.NewInt(0)
		expectedContribAmt.Mul(
			new(big.Int).SetUint64(waitingContribution1.Amount),
			new(big.Int).SetUint64(poolPair.Token2PoolValue),
		)
		expectedContribAmt.Div(
			expectedContribAmt,
			new(big.Int).SetUint64(poolPair.Token1PoolValue),
		)
		return expectedContribAmt.Uint64() == waitingContribution2.Amount
	}
	if waitingContribution1.TokenIDStr == poolPair.Token2IDStr {
		expectedContribAmt := big.NewInt(0)
		expectedContribAmt.Mul(
			new(big.Int).SetUint64(waitingContribution1.Amount),
			new(big.Int).SetUint64(poolPair.Token1PoolValue),
		)
		expectedContribAmt.Div(
			expectedContribAmt,
			new(big.Int).SetUint64(poolPair.Token2PoolValue),
		)
		return expectedContribAmt.Uint64() == waitingContribution2.Amount
	}
	return false
}

func computeActualContributedAmounts(
	waitingContribution1 *rawdbv2.PDEContribution,
	waitingContribution2 *rawdbv2.PDEContribution,
	poolPair *rawdbv2.PDEPoolForPair,
) (uint64, uint64, uint64, uint64) {
	if poolPair.Token1PoolValue == 0 || poolPair.Token2PoolValue == 0 {
		return waitingContribution1.Amount, 0, waitingContribution2.Amount, 0
	}
	if poolPair.Token1IDStr == waitingContribution1.TokenIDStr {
		// waitingAmtTemp = waitingContribution2.Amount * poolPair.Token1PoolValue / poolPair.Token2PoolValue
		contribution1Amt := big.NewInt(0)
		tempAmt := big.NewInt(0)
		tempAmt.Mul(
			new(big.Int).SetUint64(waitingContribution2.Amount),
			new(big.Int).SetUint64(poolPair.Token1PoolValue),
		)
		tempAmt.Div(
			tempAmt,
			new(big.Int).SetUint64(poolPair.Token2PoolValue),
		)
		if tempAmt.Uint64() > waitingContribution1.Amount {
			contribution1Amt = new(big.Int).SetUint64(waitingContribution1.Amount)
		} else {
			contribution1Amt = tempAmt
		}
		contribution2Amt := big.NewInt(0)
		contribution2Amt.Mul(
			contribution1Amt,
			new(big.Int).SetUint64(poolPair.Token2PoolValue),
		)
		contribution2Amt.Div(
			contribution2Amt,
			new(big.Int).SetUint64(poolPair.Token1PoolValue),
		)
		actualContribution1Amt := contribution1Amt.Uint64()
		actualContribution2Amt := contribution2Amt.Uint64()
		return actualContribution1Amt, waitingContribution1.Amount - actualContribution1Amt, actualContribution2Amt, waitingContribution2.Amount - actualContribution2Amt
	}
	if poolPair.Token1IDStr == waitingContribution2.TokenIDStr {
		// tempAmt = waitingContribution2.Amount * poolPair.Token1PoolValue / poolPair.Token2PoolValue
		contribution2Amt := big.NewInt(0)
		tempAmt := big.NewInt(0)
		tempAmt.Mul(
			new(big.Int).SetUint64(waitingContribution1.Amount),
			new(big.Int).SetUint64(poolPair.Token1PoolValue),
		)
		tempAmt.Div(
			tempAmt,
			new(big.Int).SetUint64(poolPair.Token2PoolValue),
		)
		if tempAmt.Uint64() > waitingContribution2.Amount {
			contribution2Amt = new(big.Int).SetUint64(waitingContribution2.Amount)
		} else {
			contribution2Amt = tempAmt
		}
		contribution1Amt := big.NewInt(0)
		contribution1Amt.Mul(
			contribution2Amt,
			new(big.Int).SetUint64(poolPair.Token2PoolValue),
		)
		contribution1Amt.Div(
			contribution1Amt,
			new(big.Int).SetUint64(poolPair.Token1PoolValue),
		)
		actualContribution1Amt := contribution1Amt.Uint64()
		actualContribution2Amt := contribution2Amt.Uint64()
		return actualContribution1Amt, waitingContribution1.Amount - actualContribution1Amt, actualContribution2Amt, waitingContribution2.Amount - actualContribution2Amt
	}
	return 0, 0, 0, 0
}

func (blockchain *BlockChain) buildInstructionsForPDEContribution(
	contentStr string,
	shardID byte,
	metaType int,
	currentPDEState *CurrentPDEState,
	beaconHeight uint64,
	isPRVRequired bool,
) ([][]string, error) {
	if currentPDEState == nil {
		Logger.log.Warn("WARN - [buildInstructionsForPDEContribution]: Current PDE state is null.")
		inst := []string{
			strconv.Itoa(metaType),
			strconv.Itoa(int(shardID)),
			common.PDEContributionRefundChainStatus,
			contentStr,
		}
		return [][]string{inst}, nil
	}
	contentBytes, err := base64.StdEncoding.DecodeString(contentStr)
	if err != nil {
		Logger.log.Errorf("ERROR: an error occured while decoding content string of pde withdrawal action: %+v", err)
		return [][]string{}, nil
	}
	var pdeContributionAction metadata.PDEContributionAction
	err = json.Unmarshal(contentBytes, &pdeContributionAction)
	if err != nil {
		Logger.log.Errorf("ERROR: an error occured while unmarshaling pde contribution action: %+v", err)
		return [][]string{}, nil
	}
	meta := pdeContributionAction.Meta
	waitingContribPairKey := string(rawdbv2.BuildWaitingPDEContributionKey(beaconHeight, meta.PDEContributionPairID))
	waitingContribution, found := currentPDEState.WaitingPDEContributions[waitingContribPairKey]
	if !found || waitingContribution == nil {
		currentPDEState.WaitingPDEContributions[waitingContribPairKey] = &rawdbv2.PDEContribution{
			ContributorAddressStr: meta.ContributorAddressStr,
			TokenIDStr:            meta.TokenIDStr,
			Amount:                meta.ContributedAmount,
			TxReqID:               pdeContributionAction.TxReqID,
		}
		inst := buildWaitingContributionInst(
			meta.PDEContributionPairID,
			meta.ContributorAddressStr,
			meta.ContributedAmount,
			meta.TokenIDStr,
			metaType,
			shardID,
			pdeContributionAction.TxReqID,
		)
		return [][]string{inst}, nil
	}
	if waitingContribution.TokenIDStr == meta.TokenIDStr ||
		waitingContribution.ContributorAddressStr != meta.ContributorAddressStr ||
		(isPRVRequired && waitingContribution.TokenIDStr != common.PRVIDStr && meta.TokenIDStr != common.PRVIDStr) {
		delete(currentPDEState.WaitingPDEContributions, waitingContribPairKey)
		refundInst1 := buildRefundContributionInst(
			meta.PDEContributionPairID,
			meta.ContributorAddressStr,
			meta.ContributedAmount,
			meta.TokenIDStr,
			metaType,
			shardID,
			pdeContributionAction.TxReqID,
		)
		refundInst2 := buildRefundContributionInst(
			meta.PDEContributionPairID,
			waitingContribution.ContributorAddressStr,
			waitingContribution.Amount,
			waitingContribution.TokenIDStr,
			metaType,
			shardID,
			waitingContribution.TxReqID,
		)
		return [][]string{refundInst1, refundInst2}, nil
	}
	// contributed to 2 diff sides of a pair and its a first contribution of this pair
	poolPairs := currentPDEState.PDEPoolPairs
	poolPairKey := string(rawdbv2.BuildPDEPoolForPairKey(beaconHeight, waitingContribution.TokenIDStr, meta.TokenIDStr))
	poolPair, found := poolPairs[poolPairKey]
	incomingWaitingContribution := &rawdbv2.PDEContribution{
		ContributorAddressStr: meta.ContributorAddressStr,
		TokenIDStr:            meta.TokenIDStr,
		Amount:                meta.ContributedAmount,
		TxReqID:               pdeContributionAction.TxReqID,
	}

	if !found || poolPair == nil {
		delete(currentPDEState.WaitingPDEContributions, waitingContribPairKey)
		updateWaitingContributionPairToPoolV2(
			beaconHeight,
			waitingContribution,
			incomingWaitingContribution,
			currentPDEState,
		)
		matchedInst := buildMatchedContributionInst(
			meta.PDEContributionPairID,
			meta.ContributorAddressStr,
			meta.ContributedAmount,
			meta.TokenIDStr,
			metaType,
			shardID,
			pdeContributionAction.TxReqID,
		)
		return [][]string{matchedInst}, nil
	}

	// isRightRatio(waitingContribution, incomingWaitingContribution, poolPair)
	actualWaitingContribAmt, returnedWaitingContribAmt, actualIncomingWaitingContribAmt, returnedIncomingWaitingContribAmt := computeActualContributedAmounts(
		waitingContribution,
		incomingWaitingContribution,
		poolPair,
	)
	if actualWaitingContribAmt == 0 || actualIncomingWaitingContribAmt == 0 {
		delete(currentPDEState.WaitingPDEContributions, waitingContribPairKey)
		refundInst1 := buildRefundContributionInst(
			meta.PDEContributionPairID,
			meta.ContributorAddressStr,
			meta.ContributedAmount,
			meta.TokenIDStr,
			metaType,
			shardID,
			pdeContributionAction.TxReqID,
		)
		refundInst2 := buildRefundContributionInst(
			meta.PDEContributionPairID,
			waitingContribution.ContributorAddressStr,
			waitingContribution.Amount,
			waitingContribution.TokenIDStr,
			metaType,
			shardID,
			waitingContribution.TxReqID,
		)
		return [][]string{refundInst1, refundInst2}, nil
	}

	delete(currentPDEState.WaitingPDEContributions, waitingContribPairKey)
	actualWaitingContrib := &rawdbv2.PDEContribution{
		ContributorAddressStr: waitingContribution.ContributorAddressStr,
		TokenIDStr:            waitingContribution.TokenIDStr,
		Amount:                actualWaitingContribAmt,
		TxReqID:               waitingContribution.TxReqID,
	}
	actualIncomingWaitingContrib := &rawdbv2.PDEContribution{
		ContributorAddressStr: meta.ContributorAddressStr,
		TokenIDStr:            meta.TokenIDStr,
		Amount:                actualIncomingWaitingContribAmt,
		TxReqID:               pdeContributionAction.TxReqID,
	}
	updateWaitingContributionPairToPoolV2(
		beaconHeight,
		actualWaitingContrib,
		actualIncomingWaitingContrib,
		currentPDEState,
	)
	matchedNReturnedInst1 := buildMatchedNReturnedContributionInst(
		meta.PDEContributionPairID,
		meta.ContributorAddressStr,
		actualIncomingWaitingContribAmt,
		returnedIncomingWaitingContribAmt,
		meta.TokenIDStr,
		metaType,
		shardID,
		pdeContributionAction.TxReqID,
		actualWaitingContribAmt,
	)
	matchedNReturnedInst2 := buildMatchedNReturnedContributionInst(
		meta.PDEContributionPairID,
		waitingContribution.ContributorAddressStr,
		actualWaitingContribAmt,
		returnedWaitingContribAmt,
		waitingContribution.TokenIDStr,
		metaType,
		shardID,
		waitingContribution.TxReqID,
		0,
	)
	return [][]string{matchedNReturnedInst1, matchedNReturnedInst2}, nil
}

type tradingFeeForContributorByPair struct {
	ContributorAddressStr string
	FeeAmt                uint64
	Token1IDStr           string
	Token2IDStr           string
}
