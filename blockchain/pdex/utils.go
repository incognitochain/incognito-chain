package pdex

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"sort"
	"strconv"
	"strings"

	v2 "github.com/incognitochain/incognito-chain/blockchain/pdex/v2utils"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/rawdbv2"
	instructionPdexv3 "github.com/incognitochain/incognito-chain/instruction/pdexv3"
	"github.com/incognitochain/incognito-chain/metadata"
	metadataPdexv3 "github.com/incognitochain/incognito-chain/metadata/pdexv3"
)

type StateChange = v2.StateChange

func NewStateChange() *StateChange {
	return v2.NewStateChange()
}

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
	contributionPairID string,
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
		PDEContributionPairID:      contributionPairID,
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

func generatePoolPairKey(token0Name, token1Name, txReqID string) string {
	if token0Name <= token1Name {
		return strings.Join([]string{token0Name, token1Name, txReqID}, "-")
	}
	return strings.Join([]string{token1Name, token0Name, txReqID}, "-")
}

//amplifier >= 10000
func calculateVirtualAmount(amount0, amount1 uint64, amplifier uint) (*big.Int, *big.Int) {
	if amplifier == metadataPdexv3.BaseAmplifier {
		return big.NewInt(0).SetUint64(amount0), big.NewInt(0).SetUint64(amount1)
	}
	vAmount0 := big.NewInt(0)
	vAmount1 := big.NewInt(0)
	vAmount0.Mul(
		new(big.Int).SetUint64(amount0),
		new(big.Int).SetUint64(uint64(amplifier)),
	)
	vAmount0.Div(
		vAmount0,
		new(big.Int).SetUint64(uint64(metadataPdexv3.BaseAmplifier)),
	)
	vAmount1.Mul(
		new(big.Int).SetUint64(amount1),
		new(big.Int).SetUint64(uint64(amplifier)),
	)
	vAmount1.Div(
		vAmount1,
		new(big.Int).SetUint64(uint64(metadataPdexv3.BaseAmplifier)),
	)

	return vAmount0, vAmount1
}

// TradePathFromState() prepares a trade path with reserves, orderbooks & directions.
// It returns cloned data only. State changes (if any) must be applied separately.
func TradePathFromState(
	sellToken common.Hash,
	tradePath []string,
	pairs map[string]*PoolPairState,
) (
	[]*rawdbv2.Pdexv3PoolPair, []map[common.Hash]*big.Int, []map[common.Hash]uint64, []map[common.Hash]uint64,
	[]v2.OrderBookIterator, []byte, common.Hash, error,
) {
	var results []*rawdbv2.Pdexv3PoolPair
	var orderbookList []v2.OrderBookIterator
	var tradeDirections []byte
	var lpFeesPerShare []map[common.Hash]*big.Int
	var protocolFees []map[common.Hash]uint64
	var stakingPoolFees []map[common.Hash]uint64

	nextTokenToSell := sellToken
	tradePathDupMap := make(map[string]bool)
	for _, pairID := range tradePath {
		// reject trade path with duplicate pools
		if tradePathDupMap[pairID] {
			return nil, nil, nil, nil, nil, nil, nextTokenToSell, fmt.Errorf("Path contains duplicate pool %s", pairID)
		} else {
			tradePathDupMap[pairID] = true
		}
		if pair, exists := pairs[pairID]; exists {
			pair = pair.Clone() // work on cloned state in case trade is rejected
			results = append(results, &pair.state)
			lpFeesPerShare = append(lpFeesPerShare, pair.lpFeesPerShare)
			protocolFees = append(protocolFees, pair.protocolFees)
			stakingPoolFees = append(stakingPoolFees, pair.stakingPoolFees)
			ob := pair.orderbook
			orderbookList = append(orderbookList, &ob)
			var td byte
			switch nextTokenToSell {
			case pair.state.Token0ID():
				td = v2.TradeDirectionSell0
				// set token to sell for next iteration. If this is the last iteration, it's THE token to buy
				nextTokenToSell = pair.state.Token1ID()
			case pair.state.Token1ID():
				td = v2.TradeDirectionSell1
				nextTokenToSell = pair.state.Token0ID()
			default:
				return nil, nil, nil, nil, nil, nil, nextTokenToSell, fmt.Errorf("Incompatible selling token %s vs next pair %s", nextTokenToSell.String(), pairID)
			}
			tradeDirections = append(tradeDirections, td)
		} else {
			return nil, nil, nil, nil, nil, nil, nextTokenToSell, fmt.Errorf("Path contains nonexistent pair %s", pairID)
		}
	}
	return results, lpFeesPerShare, protocolFees, stakingPoolFees, orderbookList, tradeDirections, nextTokenToSell, nil
}

func genNFT(index, beaconHeight uint64) common.Hash {
	hash := append(common.Uint64ToBytes(index), common.Uint64ToBytes(beaconHeight)...)
	return common.HashH(append(hashPrefix, hash...))
}

func executeOperationUint64(amount0, amount1 uint64, operator byte) (uint64, error) {
	var res, temp uint64
	switch operator {
	case addOperator:
		temp = amount0 + amount1
		if temp < amount0 {
			return res, errors.New("operation is out of range")
		}
	case subOperator:
		temp = amount0 - amount1
		if temp > amount0 {
			return res, errors.New("operation is out of range")
		}
	default:
		return res, errors.New("Not recognize operator")
	}
	res = temp
	return res, nil
}

func executeOperationBigInt(amount0, amount1 *big.Int, operator byte) (*big.Int, error) {
	res := big.NewInt(0)
	switch operator {
	case addOperator:
		res = res.Add(amount0, amount1)
	case subOperator:
		res = res.Sub(amount0, amount1)
	default:
		return res, errors.New("Not recognize operator")
	}
	return res, nil
}

func CalculateShareAmount(token0Amount, token1Amount, amount0, amount1, poolPairShareAmount uint64) uint64 {
	liquidityToken0 := big.NewInt(0).Mul(
		big.NewInt(0).SetUint64(amount0),
		big.NewInt(0).SetUint64(poolPairShareAmount),
	)
	liquidityToken0 = liquidityToken0.Div(
		liquidityToken0,
		big.NewInt(0).SetUint64(token0Amount),
	)
	liquidityToken1 := big.NewInt(0).Mul(
		big.NewInt(0).SetUint64(amount1),
		big.NewInt(0).SetUint64(poolPairShareAmount),
	)
	liquidityToken1 = liquidityToken1.Div(
		liquidityToken1,
		big.NewInt(0).SetUint64(token1Amount),
	)
	if liquidityToken0.Uint64() < liquidityToken1.Uint64() {
		return liquidityToken0.Uint64()
	}
	return liquidityToken1.Uint64()
}

// getTokenPricesAgainstPRV() returns the price of all available tokens in pools compared to PRV
// (price is represented in 2 big.Int). Each token's price is computed using
// its largest-normalized-liquidity pool with PRV
func getTokenPricesAgainstPRV(pairs map[string]*PoolPairState, minPRVReserve uint64) map[common.Hash][3]*big.Int {
	resultMap := make(map[common.Hash][3]*big.Int)
	for _, pair := range pairs {
		var tokenID common.Hash
		virtualTokenReserve := big.NewInt(0)
		virtualPRVReserve := big.NewInt(0)
		if pair.state.Token0ID() == common.PRVCoinID {
			tokenID = pair.state.Token1ID()
			virtualTokenReserve.Set(pair.state.Token1VirtualAmount())
			virtualPRVReserve.Set(pair.state.Token0VirtualAmount())
		} else if pair.state.Token1ID() == common.PRVCoinID {
			tokenID = pair.state.Token0ID()
			virtualTokenReserve.Set(pair.state.Token0VirtualAmount())
			virtualPRVReserve.Set(pair.state.Token1VirtualAmount())
		}

		// compare normalized PRV reserve against minPRVReserve -> compare PRV vReserve * baseAmplifer against minPRVReserve * amplifier rate
		temp1 := big.NewInt(0).Mul(virtualPRVReserve, big.NewInt(metadataPdexv3.BaseAmplifier))
		temp2 := big.NewInt(0).Mul(big.NewInt(0).SetUint64(minPRVReserve), big.NewInt(0).SetUint64(uint64(pair.state.Amplifier())))
		if temp1.Cmp(temp2) < 0 {
			continue
		}

		normalizedLiquidity := big.NewInt(0).Mul(virtualTokenReserve, virtualPRVReserve)
		normalizedLiquidity.Mul(normalizedLiquidity, big.NewInt(metadataPdexv3.BaseAmplifier))
		normalizedLiquidity.Div(normalizedLiquidity, big.NewInt(0).SetUint64(uint64(pair.state.Amplifier())))
		normalizedLiquidity.Mul(normalizedLiquidity, big.NewInt(metadataPdexv3.BaseAmplifier))
		normalizedLiquidity.Div(normalizedLiquidity, big.NewInt(0).SetUint64(uint64(pair.state.Amplifier())))

		isChosenPool := false
		if item, exists := resultMap[tokenID]; !exists {
			isChosenPool = true
		} else {
			liqCmp := normalizedLiquidity.Cmp(item[2])
			if liqCmp == 1 {
				// for each pair of token/PRV, choose pool to maximize normalized liquidity
				isChosenPool = true
			} else if liqCmp == 0 {
				// handle equalities explicitly to keep result deterministic regardless of map traversing order
				// break equality with direct rate comparison (token / PRV)
				temp := resultMap[tokenID]
				theirVirtualTokenReserve := temp[0]
				theirVirtualPRVReserve := temp[1]
				rateCmp := big.NewInt(0).Mul(virtualTokenReserve, theirVirtualPRVReserve).
					Cmp(big.NewInt(0).Mul(theirVirtualTokenReserve, virtualPRVReserve))
				if rateCmp == 1 {
					// when token/PRV pools tie in liquidity, maximize token/PRV rate to the benefit of current user
					isChosenPool = true
				}
			}
		}

		if isChosenPool {
			resultMap[tokenID] = [3]*big.Int{virtualTokenReserve, virtualPRVReserve, normalizedLiquidity}
		}
	}
	return resultMap
}

// getWeightedFee() converts the fee paid in PRV to the equivalent token amount,
// then applies the system's discount rate. Fee paid in non-PRV tokens remain the same.
func getWeightedFee(txs []metadata.Transaction, pairs map[string]*PoolPairState, params *Params,
) ([]metadata.Transaction, map[string]bool, map[string]uint64, map[string]uint64, []metadata.Transaction, error) {
	temp := uint64(100) - uint64(params.PRVDiscountPercent)
	if temp > 100 {
		return nil, nil, nil, nil, nil, fmt.Errorf("PRV Discount percent invalid")
	}
	discountPercent := big.NewInt(0).SetUint64(temp)
	rateMap := getTokenPricesAgainstPRV(pairs, params.MinPRVReserveTradingRate)
	var resultTransactions, invalidTransactions []metadata.Transaction
	fees := make(map[string]uint64)
	sellAmounts := make(map[string]uint64)
	// true indicates fee paid in PRV while selling other token
	feeInPRVMap := make(map[string]bool)
	for _, tx := range txs {
		md := tx.GetMetadata()
		var sellingTokenID common.Hash
		var fee, amount uint64
		switch v := md.(type) {
		case *metadataPdexv3.TradeRequest:
			sellingTokenID = v.TokenToSell
			fee = v.TradingFee
			amount = v.SellAmount
		default:
			Logger.log.Warnf("Cannot get trading fee of metadata type %v", md.GetType())
			continue
		}

		if sellingTokenID == common.PRVCoinID {
			feeInPRVMap[tx.Hash().String()] = false // only true when trading another token with PRV as fee
			// sell & fee in PRV
			temp := big.NewInt(0).SetUint64(fee)
			// convert the fee from PRV to equivalent token by applying discount percent
			temp.Mul(temp, big.NewInt(100))
			temp.Div(temp, discountPercent)
			if !temp.IsUint64() {
				Logger.log.Warnf("Equivalent fee out of uint64 range")
				invalidTransactions = append(invalidTransactions, tx)
				continue
			}
			fee = temp.Uint64()
		} else {
			// error was handled by tx validation
			_, burnedPRVCoin, _, _, _ := tx.GetTxFullBurnData()
			feeInPRV := burnedPRVCoin != nil
			feeInPRVMap[tx.Hash().String()] = feeInPRV
			if feeInPRV {
				// sell other token & fee in PRV
				rates, exists := rateMap[sellingTokenID]
				if !exists {
					Logger.log.Warnf("Cannot get price of token %s against PRV", sellingTokenID.String())
					invalidTransactions = append(invalidTransactions, tx)
					continue
				}
				temp := big.NewInt(0).SetUint64(fee)
				temp.Mul(temp, rates[0])
				// convert the fee from PRV to equivalent token by applying discount percent
				temp.Mul(temp, big.NewInt(100))
				temp.Div(temp, discountPercent)
				// divisions moved to last to improve precision
				temp.Div(temp, rates[1])
				if !temp.IsUint64() {
					Logger.log.Warnf("Equivalent fee out of uint64 range")
					invalidTransactions = append(invalidTransactions, tx)
					continue
				}
				// mark the fee as paid in PRV, and return the equivalent token amount
				fee = temp.Uint64()
			}
		}
		resultTransactions = append(resultTransactions, tx)
		fees[tx.Hash().String()] = fee
		sellAmounts[tx.Hash().String()] = amount
	}
	return resultTransactions, feeInPRVMap, fees, sellAmounts, invalidTransactions, nil
}

// getRefundInstructions() creates up to 2 intructions for refunding rejected trade requests
// in both token & PRV
func getRefundedTradeInstructions(md *metadataPdexv3.TradeRequest, feeInPRV bool,
	txID common.Hash, shardID byte) ([][]string, error) {
	refundReceiver, exists := md.Receiver[md.TokenToSell]
	if !exists {
		return nil, fmt.Errorf("Refund receiver not found in Trade Request")
	}

	// prepare refund instructions
	tokenRefundAmount := md.SellAmount + md.TradingFee
	if feeInPRV {
		tokenRefundAmount = md.SellAmount
	}
	sellingTokenRefundAction := instructionPdexv3.NewAction(
		&metadataPdexv3.RefundedTrade{
			Receiver: refundReceiver,
			TokenID:  md.TokenToSell,
			Amount:   tokenRefundAmount,
		},
		txID,
		shardID,
	)
	var refundInstructions [][]string = [][]string{sellingTokenRefundAction.StringSlice()}
	if feeInPRV {
		// prepare PRV refund if trading fee was paid in PRV; not applicable to requests that sell PRV.
		prvReceiver, exists := md.Receiver[common.PRVCoinID]
		if !exists {
			return nil, fmt.Errorf("Fee (PRV) Refund receiver not found in Trade Request")
		}
		feeRefundAction := instructionPdexv3.NewAction(
			&metadataPdexv3.RefundedTrade{
				Receiver: prvReceiver,
				TokenID:  common.PRVCoinID,
				Amount:   md.TradingFee,
			},
			txID,
			shardID,
		)
		refundInstructions = append(refundInstructions, feeRefundAction.StringSlice())
	}
	return refundInstructions, nil
}

// getRefundInstructions() creates up to 2 intructions for refunding rejected addOrder requests
// in both token & PRV
func getRefundedAddOrderInstructions(md *metadataPdexv3.AddOrderRequest,
	txID common.Hash, shardID byte) ([][]string, error) {
	refundReceiver, exists := md.Receiver[md.TokenToSell]
	if !exists {
		return nil, fmt.Errorf("Refund receiver not found in Trade Request")
	}

	// prepare refund instructions
	tokenRefundAmount := md.SellAmount
	sellingTokenRefundAction := instructionPdexv3.NewAction(
		&metadataPdexv3.RefundedAddOrder{
			Receiver: refundReceiver,
			TokenID:  md.TokenToSell,
			Amount:   tokenRefundAmount,
		},
		txID,
		shardID,
	)
	var refundInstructions [][]string = [][]string{sellingTokenRefundAction.StringSlice()}
	return refundInstructions, nil
}

func resetKeyValueToZero(m map[common.Hash]uint64) map[common.Hash]uint64 {
	for key := range m {
		m[key] = 0
	}
	return m
}

func getMapWithoutZeroValue(m map[common.Hash]uint64) map[common.Hash]uint64 {
	result := map[common.Hash]uint64{}
	for key, value := range m {
		if value != 0 {
			result[key] = value
		}
	}
	return result
}

func CombineReward(
	reward1 map[common.Hash]uint64,
	reward2 map[common.Hash]uint64,
) map[common.Hash]uint64 {
	result := map[common.Hash]uint64{}
	for key, value := range reward1 {
		result[key] = value
	}
	for key, value := range reward2 {
		if _, exists := result[key]; exists {
			// the reward is not greater than the total supply
			result[key] = value + result[key]
		} else {
			result[key] = value
		}
	}
	return getMapWithoutZeroValue(result)
}

func addOrderReward(
	base map[string]*OrderReward, additional map[string]map[common.Hash]uint64,
) map[string]*OrderReward {
	for nftID, reward := range additional {
		for tokenID, amt := range reward {
			if _, ok := base[nftID]; !ok {
				base[nftID] = NewOrderReward()
			}

			base[nftID].AddReward(tokenID, amt)
		}
	}
	return base
}

func addMakingVolume(
	base map[common.Hash]*MakingVolume, additional map[common.Hash]map[string]*big.Int,
) map[common.Hash]*MakingVolume {
	for tokenID, volume := range additional {
		for nftID, amt := range volume {
			if _, ok := base[tokenID]; !ok {
				base[tokenID] = NewMakingVolume()
			}

			base[tokenID].AddVolume(nftID, amt)
		}
	}
	return base
}

func getSortedPoolPairIDs(poolPairs map[string]*PoolPairState) []string {
	// To store the keys in slice in sorted order
	keys := make([]string, len(poolPairs))
	i := 0
	for poolPairID := range poolPairs {
		keys[i] = poolPairID
		i++
	}
	sort.SliceStable(keys, func(i, j int) bool {
		return keys[i] < keys[j]
	})
	return keys
}
