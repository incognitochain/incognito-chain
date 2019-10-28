package blockchain

import (
	"encoding/base64"
	"encoding/json"
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
	invariant := tokenPoolValueToSell * tokenPoolValueToBuy
	newTokenPoolValueToSell := tokenPoolValueToSell + pdeTradeReqAction.Meta.SellAmount
	newTokenPoolValueToBuy := invariant / newTokenPoolValueToSell
	buyAmt := tokenPoolValueToBuy - newTokenPoolValueToBuy
	tradeFee := (buyAmt * PDEFree) / 1000
	receiveAmt := buyAmt - tradeFee

	// update current pde state on mem
	pdePoolPair.Token1PoolValue = newTokenPoolValueToBuy
	pdePoolPair.Token2PoolValue = newTokenPoolValueToSell
	if pdePoolPair.Token1IDStr == pdeTradeReqAction.Meta.TokenIDToSellStr {
		pdePoolPair.Token1PoolValue = newTokenPoolValueToSell
		pdePoolPair.Token2PoolValue = newTokenPoolValueToBuy
	}
	tradeFeeKey := string(lvdb.BuildPDETradeFeesKey(
		beaconHeight,
		pdeTradeReqAction.Meta.TokenIDToBuyStr,
		pdeTradeReqAction.Meta.TokenIDToSellStr,
		pdeTradeReqAction.Meta.TokenIDToBuyStr,
	))
	currentPDEState.PDEFees[tradeFeeKey] += tradeFee

	pdeTradeAcceptedContent := metadata.PDETradeAcceptedContent{
		TraderAddressStr: pdeTradeReqAction.Meta.TraderAddressStr,
		TokenIDToBuyStr:  pdeTradeReqAction.Meta.TokenIDToBuyStr,
		ReceiveAmount:    receiveAmt,
		TradeFee:         tradeFee,
		Token1IDStr:      pdePoolPair.Token1IDStr,
		Token2IDStr:      pdePoolPair.Token2IDStr,
		ShardID:          shardID,
		RequestedTxID:    pdeTradeReqAction.TxReqID,
	}
	pdeTradeAcceptedContent.Token1PoolValueOperation = metadata.TokenPoolValueOperation{
		Operator: "-",
		Value:    buyAmt,
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
			Value:    buyAmt,
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
	totalSharesForToken := uint64(0)
	for shareKey, shareAmt := range currentPDEState.PDEShares {
		totalSharesForTokenPrefix := string(lvdb.BuildPDESharesKey(
			beaconHeight,
			wdMeta.WithdrawalToken1IDStr, wdMeta.WithdrawalToken2IDStr,
			withdrawalTokenIDStr, "",
		))
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
	deductingPoolValue := uint64(0)
	if withdrawalTokenIDStr == pdePoolPair.Token2IDStr {
		deductingPoolValue = pdePoolPair.Token2PoolValue * wdSharesForToken / totalSharesForToken
		if pdePoolPair.Token2PoolValue < deductingPoolValue {
			pdePoolPair.Token2PoolValue = 0
		} else {
			pdePoolPair.Token2PoolValue -= deductingPoolValue
		}
	} else {
		deductingPoolValue = pdePoolPair.Token1PoolValue * wdSharesForToken / totalSharesForToken
		if pdePoolPair.Token1PoolValue < deductingPoolValue {
			pdePoolPair.Token1PoolValue = 0
		} else {
			pdePoolPair.Token1PoolValue -= deductingPoolValue
		}
	}
	currentPDEState.PDEShares[shareForTokenKey] -= wdSharesForToken
	deductingAmounts.PoolValue = deductingPoolValue
	deductingAmounts.Shares = wdSharesForToken

	// fee
	tradeFeeKey := string(lvdb.BuildPDETradeFeesKey(
		beaconHeight,
		wdMeta.WithdrawalToken1IDStr, wdMeta.WithdrawalToken2IDStr,
		withdrawalTokenIDStr,
	))
	totalFeesOnToken, found := currentPDEState.PDEFees[tradeFeeKey]
	if !found || totalFeesOnToken == 0 {
		return deductingAmounts
	}
	deductingFee := totalFeesOnToken * wdSharesForToken / totalSharesForToken
	deductingAmounts.TradeFees = deductingFee
	if currentPDEState.PDEFees[tradeFeeKey] < deductingFee {
		currentPDEState.PDEFees[tradeFeeKey] = 0
	} else {
		currentPDEState.PDEFees[tradeFeeKey] -= deductingFee
	}
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
		DeductingTradeFees:   deductingAmountsForToken.TradeFees,
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
	// db := blockchain.GetDatabase()
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
