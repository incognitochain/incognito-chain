package blockchain

import (
	"encoding/base64"
	"encoding/json"
	"math/big"
	"strconv"
	"strings"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/database/lvdb"
	"github.com/incognitochain/incognito-chain/metadata"
)

func (blockchain *BlockChain) buildInstructionsForPDEContribution(
	contentStr string,
	shardID byte,
	metaType int,
) ([][]string, error) {
	inst := []string{
		strconv.Itoa(metaType),
		strconv.Itoa(int(shardID)),
		"accepted",
		contentStr,
	}
	return [][]string{inst}, nil
}

func (blockchain *BlockChain) buildInstructionsForPDETrade(
	contentStr string,
	shardID byte,
	metaType int,
	currentPDEState *CurrentPDEState,
	beaconHeight uint64,
) ([][]string, error) {
	if currentPDEState == nil ||
		(currentPDEState.PDEPoolPairs == nil || len(currentPDEState.PDEPoolPairs) == 0) {
		inst := []string{
			strconv.Itoa(metaType),
			strconv.Itoa(int(shardID)),
			"refund",
			contentStr,
		}
		return [][]string{inst}, nil
	}
	contentBytes, err := base64.StdEncoding.DecodeString(contentStr)
	if err != nil {
		Logger.log.Errorf("ERROR: an error occured while decoding content string of pde trade instruction: %+v", err)
		return [][]string{}, nil
	}
	var pdeTradeReqAction metadata.PDETradeRequestAction
	err = json.Unmarshal(contentBytes, &pdeTradeReqAction)
	if err != nil {
		Logger.log.Errorf("ERROR: an error occured while unmarshaling pde trade instruction: %+v", err)
		return [][]string{}, nil
	}
	pairKey := string(lvdb.BuildPDEPoolForPairKey(beaconHeight, pdeTradeReqAction.Meta.TokenIDToBuyStr, pdeTradeReqAction.Meta.TokenIDToSellStr))

	pdePoolPair, found := currentPDEState.PDEPoolPairs[pairKey]
	if !found || (pdePoolPair.Token1PoolValue == 0 || pdePoolPair.Token2PoolValue == 0) {
		inst := []string{
			strconv.Itoa(metaType),
			strconv.Itoa(int(shardID)),
			"refund",
			contentStr,
		}
		return [][]string{inst}, nil
	}
	// trade accepted
	tokenPoolValueToBuy := pdePoolPair.Token1PoolValue
	tokenPoolValueToSell := pdePoolPair.Token2PoolValue
	if pdePoolPair.Token1IDStr == pdeTradeReqAction.Meta.TokenIDToSellStr {
		tokenPoolValueToSell = pdePoolPair.Token1PoolValue
		tokenPoolValueToBuy = pdePoolPair.Token2PoolValue
	}
	invariant := big.NewInt(0)
	invariant.Mul(big.NewInt(int64(tokenPoolValueToSell)), big.NewInt(int64(tokenPoolValueToBuy)))
	fee := pdeTradeReqAction.Meta.SellAmount / PDEDevisionAmountForFee
	newTokenPoolValueToSell := big.NewInt(0)
	newTokenPoolValueToSell.Add(big.NewInt(int64(tokenPoolValueToSell)), big.NewInt(int64(pdeTradeReqAction.Meta.SellAmount)))
	newTokenPoolValueToSellAfterFee := big.NewInt(0).Sub(newTokenPoolValueToSell, big.NewInt(int64(fee)))

	newTokenPoolValueToBuy := big.NewInt(0).Div(invariant, newTokenPoolValueToSellAfterFee).Uint64()
	modValue := big.NewInt(0).Mod(invariant, newTokenPoolValueToSellAfterFee)

	if modValue.Cmp(big.NewInt(0)) != 0 {
		newTokenPoolValueToBuy++
	}
	if tokenPoolValueToBuy <= newTokenPoolValueToBuy {
		inst := []string{
			strconv.Itoa(metaType),
			strconv.Itoa(int(shardID)),
			"refund",
			contentStr,
		}
		return [][]string{inst}, nil
	}

	receiveAmt := tokenPoolValueToBuy - newTokenPoolValueToBuy
	// update current pde state on mem
	pdePoolPair.Token1PoolValue = newTokenPoolValueToBuy
	pdePoolPair.Token2PoolValue = newTokenPoolValueToSell.Uint64()
	if pdePoolPair.Token1IDStr == pdeTradeReqAction.Meta.TokenIDToSellStr {
		pdePoolPair.Token1PoolValue = newTokenPoolValueToSell.Uint64()
		pdePoolPair.Token2PoolValue = newTokenPoolValueToBuy
	}

	pdeTradeAcceptedContent := metadata.PDETradeAcceptedContent{
		TraderAddressStr: pdeTradeReqAction.Meta.TraderAddressStr,
		TokenIDToBuyStr:  pdeTradeReqAction.Meta.TokenIDToBuyStr,
		ReceiveAmount:    receiveAmt,
		Token1IDStr:      pdePoolPair.Token1IDStr,
		Token2IDStr:      pdePoolPair.Token2IDStr,
		ShardID:          shardID,
		RequestedTxID:    pdeTradeReqAction.TxReqID,
	}
	pdeTradeAcceptedContent.Token1PoolValueOperation = metadata.TokenPoolValueOperation{
		Operator: "-",
		Value:    receiveAmt,
	}
	pdeTradeAcceptedContent.Token2PoolValueOperation = metadata.TokenPoolValueOperation{
		Operator: "+",
		Value:    pdeTradeReqAction.Meta.SellAmount,
	}
	if pdePoolPair.Token1IDStr == pdeTradeReqAction.Meta.TokenIDToSellStr {
		pdeTradeAcceptedContent.Token1PoolValueOperation = metadata.TokenPoolValueOperation{
			Operator: "+",
			Value:    pdeTradeReqAction.Meta.SellAmount,
		}
		pdeTradeAcceptedContent.Token2PoolValueOperation = metadata.TokenPoolValueOperation{
			Operator: "-",
			Value:    receiveAmt,
		}
	}
	pdeTradeAcceptedContentBytes, err := json.Marshal(pdeTradeAcceptedContent)
	if err != nil {
		Logger.log.Errorf("ERROR: an error occured while marshaling pdeTradeAcceptedContent: %+v", err)
		return [][]string{}, nil
	}
	inst := []string{
		strconv.Itoa(metaType),
		strconv.Itoa(int(shardID)),
		"accepted",
		string(pdeTradeAcceptedContentBytes),
	}
	return [][]string{inst}, nil
}

func deductPDEAmounts(
	withdrawalTokenIDStr string,
	wdMeta metadata.PDEWithdrawalRequest,
	currentPDEState *CurrentPDEState,
	beaconHeight uint64,
) *DeductingAmountsForTokenByWithdrawal {
	var deductingAmounts *DeductingAmountsForTokenByWithdrawal
	pairKey := string(lvdb.BuildPDEPoolForPairKey(
		beaconHeight,
		wdMeta.WithdrawalToken1IDStr, wdMeta.WithdrawalToken2IDStr,
	))
	pdePoolPair, found := currentPDEState.PDEPoolPairs[pairKey]
	if !found || pdePoolPair == nil {
		return deductingAmounts
	}
	shareForTokenKey := string(lvdb.BuildPDESharesKey(
		beaconHeight,
		wdMeta.WithdrawalToken1IDStr, wdMeta.WithdrawalToken2IDStr,
		withdrawalTokenIDStr, wdMeta.WithdrawerAddressStr,
	))
	currentSharesForToken, found := currentPDEState.PDEShares[shareForTokenKey]
	if !found || currentSharesForToken == 0 {
		return deductingAmounts
	}

	totalSharesForTokenPrefix := string(lvdb.BuildPDESharesKey(
		beaconHeight,
		wdMeta.WithdrawalToken1IDStr, wdMeta.WithdrawalToken2IDStr,
		withdrawalTokenIDStr, "",
	))
	totalSharesForToken := uint64(0)
	for shareKey, shareAmt := range currentPDEState.PDEShares {
		if strings.Contains(shareKey, totalSharesForTokenPrefix) {
			totalSharesForToken += shareAmt
		}
	}
	if totalSharesForToken == 0 {
		return deductingAmounts
	}
	wdSharesForToken := wdMeta.WithdrawalShare1Amt
	if withdrawalTokenIDStr == wdMeta.WithdrawalToken2IDStr {
		wdSharesForToken = wdMeta.WithdrawalShare2Amt
	}
	if wdSharesForToken > currentSharesForToken {
		wdSharesForToken = currentSharesForToken
	}
	if wdSharesForToken == 0 {
		return deductingAmounts
	}

	deductingAmounts = &DeductingAmountsForTokenByWithdrawal{}
	deductingPoolValue := big.NewInt(0)
	if withdrawalTokenIDStr == pdePoolPair.Token2IDStr {
		deductingPoolValue.Mul(big.NewInt(int64(pdePoolPair.Token2PoolValue)), big.NewInt(int64(wdSharesForToken)))
		deductingPoolValue.Div(deductingPoolValue, big.NewInt(int64(totalSharesForToken)))
		// deductingPoolValue = pdePoolPair.Token2PoolValue * wdSharesForToken / totalSharesForToken
		if pdePoolPair.Token2PoolValue < deductingPoolValue.Uint64() {
			pdePoolPair.Token2PoolValue = 0
		} else {
			pdePoolPair.Token2PoolValue -= deductingPoolValue.Uint64()
		}
	} else {
		// deductingPoolValue = pdePoolPair.Token1PoolValue * wdSharesForToken / totalSharesForToken
		deductingPoolValue.Mul(big.NewInt(int64(pdePoolPair.Token1PoolValue)), big.NewInt(int64(wdSharesForToken)))
		deductingPoolValue.Div(deductingPoolValue, big.NewInt(int64(totalSharesForToken)))
		if pdePoolPair.Token1PoolValue < deductingPoolValue.Uint64() {
			pdePoolPair.Token1PoolValue = 0
		} else {
			pdePoolPair.Token1PoolValue -= deductingPoolValue.Uint64()
		}
	}
	if currentPDEState.PDEShares[shareForTokenKey] < wdSharesForToken {
		currentPDEState.PDEShares[shareForTokenKey] = 0
	} else {
		currentPDEState.PDEShares[shareForTokenKey] -= wdSharesForToken
	}
	deductingAmounts.PoolValue = deductingPoolValue.Uint64()
	deductingAmounts.Shares = wdSharesForToken
	return deductingAmounts
}

func buildPDEWithdrawalAcceptedInst(
	wdMeta metadata.PDEWithdrawalRequest,
	shardID byte,
	metaType int,
	withdrawalTokenIDStr string,
	deductingAmountsForToken *DeductingAmountsForTokenByWithdrawal,
	txReqID common.Hash,
) ([]string, error) {
	wdAcceptedContent := metadata.PDEWithdrawalAcceptedContent{
		WithdrawalTokenIDStr: withdrawalTokenIDStr,
		WithdrawerAddressStr: wdMeta.WithdrawerAddressStr,
		DeductingPoolValue:   deductingAmountsForToken.PoolValue,
		DeductingShares:      deductingAmountsForToken.Shares,
		PairToken1IDStr:      wdMeta.WithdrawalToken1IDStr,
		PairToken2IDStr:      wdMeta.WithdrawalToken2IDStr,
		TxReqID:              txReqID,
		ShardID:              shardID,
	}
	wdAcceptedContentBytes, err := json.Marshal(wdAcceptedContent)
	if err != nil {
		Logger.log.Errorf("ERROR: an error occured while marshaling PDEWithdrawalAcceptedContent: %+v", err)
		return []string{}, nil
	}
	return []string{
		strconv.Itoa(metaType),
		strconv.Itoa(int(shardID)),
		"accepted",
		string(wdAcceptedContentBytes),
	}, nil
}

func (blockchain *BlockChain) buildInstructionsForPDEWithdrawal(
	contentStr string,
	shardID byte,
	metaType int,
	currentPDEState *CurrentPDEState,
	beaconHeight uint64,
) ([][]string, error) {
	contentBytes, err := base64.StdEncoding.DecodeString(contentStr)
	if err != nil {
		Logger.log.Errorf("ERROR: an error occured while decoding content string of pde withdrawal action: %+v", err)
		return [][]string{}, nil
	}
	var pdeWithdrawalRequestAction metadata.PDEWithdrawalRequestAction
	err = json.Unmarshal(contentBytes, &pdeWithdrawalRequestAction)
	if err != nil {
		Logger.log.Errorf("ERROR: an error occured while unmarshaling pde withdrawal request action: %+v", err)
		return [][]string{}, nil
	}
	wdMeta := pdeWithdrawalRequestAction.Meta
	deductingAmountsForToken1 := deductPDEAmounts(
		wdMeta.WithdrawalToken1IDStr,
		wdMeta,
		currentPDEState,
		beaconHeight,
	)
	deductingAmountsForToken2 := deductPDEAmounts(
		wdMeta.WithdrawalToken2IDStr,
		wdMeta,
		currentPDEState,
		beaconHeight,
	)
	insts := [][]string{}
	if deductingAmountsForToken1 != nil {
		inst, err := buildPDEWithdrawalAcceptedInst(
			wdMeta,
			shardID,
			metaType,
			wdMeta.WithdrawalToken1IDStr,
			deductingAmountsForToken1,
			pdeWithdrawalRequestAction.TxReqID,
		)
		if err != nil {
			return [][]string{}, nil
		}
		insts = append(insts, inst)
	}
	if deductingAmountsForToken2 != nil {
		inst, err := buildPDEWithdrawalAcceptedInst(
			wdMeta,
			shardID,
			metaType,
			wdMeta.WithdrawalToken2IDStr,
			deductingAmountsForToken2,
			pdeWithdrawalRequestAction.TxReqID,
		)
		if err != nil {
			return [][]string{}, nil
		}
		insts = append(insts, inst)
	}
	return insts, nil
}
