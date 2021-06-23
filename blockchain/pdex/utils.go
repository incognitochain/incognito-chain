package pdex

import (
	"encoding/json"
	"errors"
	"math/big"
	"sort"
	"strconv"
	"strings"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/config"
	"github.com/incognitochain/incognito-chain/dataaccessobject/rawdbv2"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/metadata"
)

func InitVersionByBeaconHeight(beaconHeight uint64) State {
	var state State
	return state
}

func isTradingPairContainsPRV(
	tokenIDToSellStr string,
	tokenIDToBuyStr string,
) bool {
	return tokenIDToSellStr == common.PRVCoinID.String() ||
		tokenIDToBuyStr == common.PRVCoinID.String()
}

type tradingFeeForContributorByPair struct {
	ContributorAddressStr string
	FeeAmt                uint64
	Token1IDStr           string
	Token2IDStr           string
}

type tradeInfo struct {
	tokenIDToBuyStr         string
	tokenIDToSellStr        string
	sellAmount              uint64
	newTokenPoolValueToBuy  uint64
	newTokenPoolValueToSell uint64
	receiveAmount           uint64
}

type shareInfo struct {
	shareKey string
	shareAmt uint64
}

type deductingAmountsByWithdrawal struct {
	Token1IDStr string
	PoolValue1  uint64
	Token2IDStr string
	PoolValue2  uint64
	Shares      uint64
}

func isExistedInPoolPair(
	poolPairs map[string]*rawdbv2.PDEPoolForPair,
	beaconHeight uint64,
	token1IDStr string,
	token2IDStr string,
) bool {
	poolPairKey := string(rawdbv2.BuildPDEPoolForPairKey(beaconHeight, token1IDStr, token2IDStr))
	poolPair, found := poolPairs[poolPairKey]
	if !found || poolPair == nil || poolPair.Token1PoolValue == 0 || poolPair.Token2PoolValue == 0 {
		return false
	}
	return true
}

func calcTradeValue(
	poolPair *rawdbv2.PDEPoolForPair,
	tokenIDStrToSell string,
	sellAmount uint64,
) (uint64, uint64, uint64, error) {
	tokenPoolValueToBuy := poolPair.Token1PoolValue
	tokenPoolValueToSell := poolPair.Token2PoolValue
	if poolPair.Token1IDStr == tokenIDStrToSell {
		tokenPoolValueToSell = poolPair.Token1PoolValue
		tokenPoolValueToBuy = poolPair.Token2PoolValue
	}
	invariant := big.NewInt(0)
	invariant.Mul(new(big.Int).SetUint64(tokenPoolValueToSell), new(big.Int).SetUint64(tokenPoolValueToBuy))
	newTokenPoolValueToSell := big.NewInt(0)
	newTokenPoolValueToSell.Add(new(big.Int).SetUint64(tokenPoolValueToSell), new(big.Int).SetUint64(sellAmount))

	newTokenPoolValueToBuy := big.NewInt(0).Div(invariant, newTokenPoolValueToSell).Uint64()
	modValue := big.NewInt(0).Mod(invariant, newTokenPoolValueToSell)
	if modValue.Cmp(big.NewInt(0)) != 0 {
		newTokenPoolValueToBuy++
	}
	if tokenPoolValueToBuy <= newTokenPoolValueToBuy {
		return uint64(0), uint64(0), uint64(0), errors.New("tokenPoolValueToBuy <= newTokenPoolValueToBuy")
	}
	return tokenPoolValueToBuy - newTokenPoolValueToBuy, newTokenPoolValueToBuy, newTokenPoolValueToSell.Uint64(), nil
}

// build refund instructions for an unsuccessful trade
// note: only refund if amount > 0
func refundForCrossPoolTrade(
	sequentialTrades []*tradeInfo,
	action metadata.PDECrossPoolTradeRequestAction,
) ([][]string, error) {

	refundTradingFeeInst := buildCrossPoolTradeRefundInst(
		action.Meta.SubTraderAddressStr,
		action.Meta.SubTxRandomStr,
		common.PRVIDStr,
		action.Meta.TradingFee,
		common.PDECrossPoolTradeFeeRefundChainStatus,
		action.ShardID,
		action.TxReqID,
	)
	refundSellingTokenInst := buildCrossPoolTradeRefundInst(
		action.Meta.TraderAddressStr,
		action.Meta.TxRandomStr,
		sequentialTrades[0].tokenIDToSellStr,
		sequentialTrades[0].sellAmount,
		common.PDECrossPoolTradeSellingTokenRefundChainStatus,
		action.ShardID,
		action.TxReqID,
	)

	refundInsts := [][]string{}
	if len(refundTradingFeeInst) > 0 {
		refundInsts = append(refundInsts, refundTradingFeeInst)
	}
	if len(refundSellingTokenInst) > 0 {
		refundInsts = append(refundInsts, refundSellingTokenInst)
	}
	return refundInsts, nil
}

func buildInstsForUntradableActions(
	actions []metadata.PDECrossPoolTradeRequestAction,
) [][]string {
	untradableInsts := [][]string{}
	for _, tradeAction := range actions {
		//Work-around solution for the sake of syncing mainnet
		//Todo: find better solution
		if _, err := metadata.AssertPaymentAddressAndTxVersion(tradeAction.Meta.TraderAddressStr, 1); err == nil {
			if len(tradeAction.Meta.SubTraderAddressStr) == 0 {
				tradeAction.Meta.SubTraderAddressStr = tradeAction.Meta.TraderAddressStr
			}
		}

		refundTradingFeeInst := buildCrossPoolTradeRefundInst(
			tradeAction.Meta.SubTraderAddressStr,
			tradeAction.Meta.SubTxRandomStr,
			common.PRVCoinID.String(),
			tradeAction.Meta.TradingFee,
			common.PDECrossPoolTradeFeeRefundChainStatus,
			tradeAction.ShardID,
			tradeAction.TxReqID,
		)
		if len(refundTradingFeeInst) > 0 {
			untradableInsts = append(untradableInsts, refundTradingFeeInst)
		}

		refundSellingTokenInst := buildCrossPoolTradeRefundInst(
			tradeAction.Meta.TraderAddressStr,
			tradeAction.Meta.TxRandomStr,
			tradeAction.Meta.TokenIDToSellStr,
			tradeAction.Meta.SellAmount,
			common.PDECrossPoolTradeSellingTokenRefundChainStatus,
			tradeAction.ShardID,
			tradeAction.TxReqID,
		)
		if len(refundSellingTokenInst) > 0 {
			untradableInsts = append(untradableInsts, refundSellingTokenInst)
		}
	}
	return untradableInsts
}

func buildCrossPoolTradeRefundInst(
	traderAddressStr string,
	txRandomStr string,
	tokenIDStr string,
	amount uint64,
	status string,
	shardID byte,
	txReqID common.Hash,
) []string {
	if amount == 0 {
		return []string{}
	}
	refundCrossPoolTrade := metadata.PDERefundCrossPoolTrade{
		TraderAddressStr: traderAddressStr,
		TxRandomStr:      txRandomStr,
		TokenIDStr:       tokenIDStr,
		Amount:           amount,
		ShardID:          shardID,
		TxReqID:          txReqID,
	}
	refundCrossPoolTradeBytes, _ := json.Marshal(refundCrossPoolTrade)
	return []string{
		strconv.Itoa(metadata.PDECrossPoolTradeRequestMeta),
		strconv.Itoa(int(shardID)),
		status,
		string(refundCrossPoolTradeBytes),
	}
}

func buildWithdrawalAcceptedInst(
	action metadata.PDEWithdrawalRequestAction,
	withdrawalTokenIDStr string,
	deductingPoolValue uint64,
	deductingShares uint64,
) ([]string, error) {
	wdAcceptedContent := metadata.PDEWithdrawalAcceptedContent{
		WithdrawalTokenIDStr: withdrawalTokenIDStr,
		WithdrawerAddressStr: action.Meta.WithdrawerAddressStr,
		DeductingPoolValue:   deductingPoolValue,
		DeductingShares:      deductingShares,
		PairToken1IDStr:      action.Meta.WithdrawalToken1IDStr,
		PairToken2IDStr:      action.Meta.WithdrawalToken2IDStr,
		TxReqID:              action.TxReqID,
		ShardID:              action.ShardID,
	}
	wdAcceptedContentBytes, err := json.Marshal(wdAcceptedContent)
	if err != nil {
		Logger.log.Errorf("ERROR: an error occured while marshaling PDEWithdrawalAcceptedContent: %+v", err)
		return []string{}, nil
	}
	return []string{
		strconv.Itoa(metadata.PDEWithdrawalRequestMeta),
		strconv.Itoa(int(action.ShardID)),
		common.PDEWithdrawalAcceptedChainStatus,
		string(wdAcceptedContentBytes),
	}, nil
}

func buildWaitingContributionInst(
	action metadata.PDEContributionAction,
	metaType int,
) []string {
	waitingContribution := metadata.PDEWaitingContribution{
		PDEContributionPairID: action.Meta.PDEContributionPairID,
		ContributorAddressStr: action.Meta.ContributorAddressStr,
		ContributedAmount:     action.Meta.ContributedAmount,
		TokenIDStr:            action.Meta.TokenIDStr,
		TxReqID:               action.TxReqID,
	}
	waitingContributionBytes, _ := json.Marshal(waitingContribution)
	return []string{
		strconv.Itoa(metaType),
		strconv.Itoa(int(action.ShardID)),
		common.PDEContributionWaitingChainStatus,
		string(waitingContributionBytes),
	}
}

func buildRefundContributionInst(
	action metadata.PDEContributionAction,
	metaType int,
) []string {
	refundContribution := metadata.PDERefundContribution{
		PDEContributionPairID: action.Meta.PDEContributionPairID,
		ContributorAddressStr: action.Meta.ContributorAddressStr,
		ContributedAmount:     action.Meta.ContributedAmount,
		TokenIDStr:            action.Meta.TokenIDStr,
		TxReqID:               action.TxReqID,
		ShardID:               action.ShardID,
	}
	refundContributionBytes, _ := json.Marshal(refundContribution)
	return []string{
		strconv.Itoa(metaType),
		strconv.Itoa(int(action.ShardID)),
		common.PDEContributionRefundChainStatus,
		string(refundContributionBytes),
	}
}

func buildMatchedContributionInst(
	action metadata.PDEContributionAction,
	metaType int,
) []string {
	matchedContribution := metadata.PDEMatchedContribution{
		PDEContributionPairID: action.Meta.PDEContributionPairID,
		ContributorAddressStr: action.Meta.ContributorAddressStr,
		ContributedAmount:     action.Meta.ContributedAmount,
		TokenIDStr:            action.Meta.TokenIDStr,
		TxReqID:               action.TxReqID,
	}
	matchedContributionBytes, _ := json.Marshal(matchedContribution)
	return []string{
		strconv.Itoa(metaType),
		strconv.Itoa(int(action.ShardID)),
		common.PDEContributionMatchedChainStatus,
		string(matchedContributionBytes),
	}
}

func buildMatchedAndReturnedContributionInst(
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

func updateWaitingContributionPairToPool(
	beaconHeight uint64,
	waitingContribution1 *rawdbv2.PDEContribution,
	waitingContribution2 *rawdbv2.PDEContribution,
	poolPairs map[string]*rawdbv2.PDEPoolForPair,
	shares map[string]uint64,
) error {
	err := addShareAmountUp(
		beaconHeight,
		waitingContribution1.TokenIDStr,
		waitingContribution2.TokenIDStr,
		waitingContribution1.TokenIDStr,
		waitingContribution1.ContributorAddressStr,
		waitingContribution1.Amount,
		poolPairs,
		shares,
	)
	if err != nil {
		return err
	}

	waitingContributions := []*rawdbv2.PDEContribution{waitingContribution1, waitingContribution2}
	sort.Slice(waitingContributions, func(i, j int) bool {
		return waitingContributions[i].TokenIDStr < waitingContributions[j].TokenIDStr
	})
	poolForPairKey := string(rawdbv2.BuildPDEPoolForPairKey(beaconHeight, waitingContributions[0].TokenIDStr, waitingContributions[1].TokenIDStr))
	poolForPair, found := poolPairs[poolForPairKey]
	var amountToken1, amountToken2 uint64

	if !found || poolForPair == nil {
		amountToken1 = waitingContributions[0].Amount
		amountToken2 = waitingContributions[1].Amount
	} else {
		amountToken1 = poolForPair.Token1PoolValue + waitingContributions[0].Amount
		amountToken2 = poolForPair.Token2PoolValue + waitingContributions[1].Amount
	}
	storePoolForPair(
		poolPairs,
		poolForPairKey,
		waitingContributions[0].TokenIDStr,
		amountToken1,
		waitingContributions[1].TokenIDStr,
		amountToken2,
	)
	return nil
}

func addShareAmountUp(
	beaconHeight uint64,
	token1IDStr string,
	token2IDStr string,
	contributedTokenIDStr string,
	contributorAddrStr string,
	amt uint64,
	poolPairs map[string]*rawdbv2.PDEPoolForPair,
	shares map[string]uint64,
) error {
	shareOnTokenPrefixBytes, err := rawdbv2.BuildPDESharesKeyV2(beaconHeight, token1IDStr, token2IDStr, "")
	if err != nil {
		Logger.log.Errorf("cannot build PDESharesKeyV2. Error: %v\n", err)
		return err
	}

	shareOnTokenPrefix := string(shareOnTokenPrefixBytes)
	totalSharesOnToken := uint64(0)
	for key, value := range shares {
		if strings.Contains(key, shareOnTokenPrefix) {
			totalSharesOnToken += value
		}
	}
	shareKeyBytes, err := rawdbv2.BuildPDESharesKeyV2(beaconHeight, token1IDStr, token2IDStr, contributorAddrStr)
	if err != nil {
		Logger.log.Errorf("cannot find pdeShareKey for address: %v. Error: %v\n", contributorAddrStr, err)
		return err
	}

	shareKey := string(shareKeyBytes)
	if totalSharesOnToken == 0 {
		shares[shareKey] = amt
		return nil
	}
	poolPairKey := string(rawdbv2.BuildPDEPoolForPairKey(beaconHeight, token1IDStr, token2IDStr))
	poolPair, found := poolPairs[poolPairKey]
	if !found || poolPair == nil {
		shares[shareKey] = amt
		return errors.New("poolPair is null")
	}
	poolValue := poolPair.Token1PoolValue
	if poolPair.Token2IDStr == contributedTokenIDStr {
		poolValue = poolPair.Token2PoolValue
	}
	if poolValue == 0 {
		shares[shareKey] = amt
		return nil
	}
	increasingAmt := big.NewInt(0)

	increasingAmt.Mul(new(big.Int).SetUint64(totalSharesOnToken), new(big.Int).SetUint64(amt))
	increasingAmt.Div(increasingAmt, new(big.Int).SetUint64(poolValue))

	currentShare, found := shares[shareKey]
	addedUpAmt := increasingAmt.Uint64()
	if found {
		addedUpAmt += currentShare
	}
	shares[shareKey] = addedUpAmt
	return nil
}

func storePoolForPair(
	poolPairs map[string]*rawdbv2.PDEPoolForPair,
	pdePoolForPairKey string,
	token1IDStr string,
	token1PoolValue uint64,
	token2IDStr string,
	token2PoolValue uint64,
) {
	poolForPair := &rawdbv2.PDEPoolForPair{
		Token1IDStr:     token1IDStr,
		Token1PoolValue: token1PoolValue,
		Token2IDStr:     token2IDStr,
		Token2PoolValue: token2PoolValue,
	}
	poolPairs[pdePoolForPairKey] = poolForPair
}

func InitPDEStateFromDB(
	stateDB *statedb.StateDB,
	beaconHeight uint64,
) (State, error) {

	var state State

	if beaconHeight < config.Param().PDEV3Height {
		if beaconHeight == 1 {
			return &stateV1{}, nil
		}
		return InitStateV1(stateDB, beaconHeight)
	}

	return state, nil
}
