package pdex

import (
	"encoding/base64"
	"encoding/json"
	"math/big"
	"sort"
	"strconv"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/rawdbv2"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/incognitochain/incognito-chain/utils"
)

type stateProducerBase struct {
}

func (sp *stateProducerBase) feeWithdrawal(
	actions [][]string,
	beaconHeight uint64,
	tradingFees map[string]uint64,
) ([][]string, map[string]uint64, error) {
	res := [][]string{}
	for _, action := range actions {
		contentStr := action[1]
		contentBytes, err := base64.StdEncoding.DecodeString(contentStr)
		if err != nil {
			Logger.log.Errorf("ERROR: an error occured while decoding content string of pde withdrawal action: %+v", err)
			return utils.EmptyStringMatrix, tradingFees, err
		}
		var feeWithdrawalRequestAction metadata.PDEFeeWithdrawalRequestAction
		err = json.Unmarshal(contentBytes, &feeWithdrawalRequestAction)
		if err != nil {
			Logger.log.Errorf("ERROR: an error occured while unmarshaling pde fee withdrawal request action: %+v", err)
			return utils.EmptyStringMatrix, tradingFees, err
		}
		wdMeta := feeWithdrawalRequestAction.Meta
		tradingFeeKeyBytes, err := rawdbv2.BuildPDETradingFeeKey(
			beaconHeight,
			wdMeta.WithdrawalToken1IDStr,
			wdMeta.WithdrawalToken2IDStr,
			wdMeta.WithdrawerAddressStr,
		)
		if err != nil {
			Logger.log.Errorf("cannot build PDETradingFeeKey for address: %v. Error: %v\n", wdMeta.WithdrawerAddressStr, err)
			return utils.EmptyStringMatrix, tradingFees, err
		}
		tradingFeeKey := string(tradingFeeKeyBytes)
		withdrawableFee, found := tradingFees[tradingFeeKey]
		if !found || withdrawableFee < wdMeta.WithdrawalFeeAmt {
			rejectedInst := []string{
				strconv.Itoa(metadata.PDEFeeWithdrawalRequestMeta),
				strconv.Itoa(int(feeWithdrawalRequestAction.ShardID)),
				common.PDEFeeWithdrawalRejectedChainStatus,
				contentStr,
			}
			res = append(res, rejectedInst)
			continue
		}
		tradingFees[tradingFeeKey] -= wdMeta.WithdrawalFeeAmt
		acceptedInst := []string{
			strconv.Itoa(metadata.PDEFeeWithdrawalRequestMeta),
			strconv.Itoa(int(feeWithdrawalRequestAction.ShardID)),
			common.PDEFeeWithdrawalAcceptedChainStatus,
			contentStr,
		}
		res = append(res, acceptedInst)
	}
	return res, tradingFees, nil
}

func (sp *stateProducerBase) sortTradeInstsByFee(
	actions [][]string,
	beaconHeight uint64,
	poolPairs map[string]*rawdbv2.PDEPoolForPair,
) []metadata.PDETradeRequestAction {
	tradesByPairs := make(map[string][]metadata.PDETradeRequestAction)

	for _, action := range actions {
		contentStr := action[1]
		contentBytes, err := base64.StdEncoding.DecodeString(contentStr)
		if err != nil {
			Logger.log.Errorf("ERROR: an error occured while decoding content string of pde trade action: %+v", err)
			continue
		}
		tradeReqAction := metadata.PDETradeRequestAction{}
		err = json.Unmarshal(contentBytes, &tradeReqAction)
		if err != nil {
			Logger.log.Errorf("ERROR: an error occured while unmarshaling pde trade action: %+v", err)
			continue
		}
		tradeMeta := tradeReqAction.Meta
		poolPairKey := string(rawdbv2.BuildPDEPoolForPairKey(beaconHeight, tradeMeta.TokenIDToBuyStr, tradeMeta.TokenIDToSellStr))
		tradesByPairs[poolPairKey] = append(tradesByPairs[poolPairKey], tradeReqAction)
	}

	notExistingPairTradeActions := []metadata.PDETradeRequestAction{}
	sortedExistingPairTradeActions := []metadata.PDETradeRequestAction{}

	var ppKeys []string
	for k := range tradesByPairs {
		ppKeys = append(ppKeys, k)
	}
	sort.Strings(ppKeys)
	for _, poolPairKey := range ppKeys {
		tradeActions := tradesByPairs[poolPairKey]
		poolPair, found := poolPairs[poolPairKey]
		if !found || poolPair == nil || poolPair.Token1PoolValue == 0 || poolPair.Token2PoolValue == 0 {
			notExistingPairTradeActions = append(notExistingPairTradeActions, tradeActions...)
			continue
		}

		// sort trade actions by trading fee
		sort.Slice(tradeActions, func(i, j int) bool {
			// comparing a/b to c/d is equivalent with comparing a*d to c*b
			firstItemProportion := big.NewInt(0)
			firstItemProportion.Mul(
				new(big.Int).SetUint64(tradeActions[i].Meta.TradingFee),
				new(big.Int).SetUint64(tradeActions[j].Meta.SellAmount),
			)
			secondItemProportion := big.NewInt(0)
			secondItemProportion.Mul(
				new(big.Int).SetUint64(tradeActions[j].Meta.TradingFee),
				new(big.Int).SetUint64(tradeActions[i].Meta.SellAmount),
			)
			return firstItemProportion.Cmp(secondItemProportion) == 1
		})
		sortedExistingPairTradeActions = append(sortedExistingPairTradeActions, tradeActions...)
	}
	return append(sortedExistingPairTradeActions, notExistingPairTradeActions...)
}

func buildAcceptedTradeInstruction(
	action metadata.PDETradeRequestAction,
	beaconHeight, receiveAmount uint64,
	poolPairs map[string]*rawdbv2.PDEPoolForPair,
) ([]string, error) {

	pairKey := string(rawdbv2.BuildPDEPoolForPairKey(beaconHeight, action.Meta.TokenIDToBuyStr, action.Meta.TokenIDToSellStr))
	poolPair := poolPairs[pairKey]

	tradeAcceptedContent := metadata.PDETradeAcceptedContent{
		TraderAddressStr: action.Meta.TraderAddressStr,
		TxRandomStr:      action.Meta.TxRandomStr,
		TokenIDToBuyStr:  action.Meta.TokenIDToBuyStr,
		ReceiveAmount:    receiveAmount,
		Token1IDStr:      poolPair.Token1IDStr,
		Token2IDStr:      poolPair.Token2IDStr,
		ShardID:          action.ShardID,
		RequestedTxID:    action.TxReqID,
	}
	tradeAcceptedContent.Token1PoolValueOperation = metadata.TokenPoolValueOperation{
		Operator: "-",
		Value:    receiveAmount,
	}
	tradeAcceptedContent.Token2PoolValueOperation = metadata.TokenPoolValueOperation{
		Operator: "+",
		Value:    action.Meta.SellAmount + action.Meta.TradingFee,
	}
	if poolPair.Token1IDStr == action.Meta.TokenIDToSellStr {
		tradeAcceptedContent.Token1PoolValueOperation = metadata.TokenPoolValueOperation{
			Operator: "+",
			Value:    action.Meta.SellAmount + action.Meta.TradingFee,
		}
		tradeAcceptedContent.Token2PoolValueOperation = metadata.TokenPoolValueOperation{
			Operator: "-",
			Value:    receiveAmount,
		}
	}
	tradeAcceptedContentBytes, err := json.Marshal(tradeAcceptedContent)
	if err != nil {
		Logger.log.Errorf("ERROR: an error occured while marshaling pdeTradeAcceptedContent: %+v", err)
		return utils.EmptyStringArray, err
	}
	return []string{
		strconv.Itoa(metadata.PDETradeRequestMeta),
		strconv.Itoa(int(action.ShardID)),
		common.PDETradeAcceptedChainStatus,
		string(tradeAcceptedContentBytes),
	}, nil
}

func shouldRefundTradeAction(
	action metadata.PDETradeRequestAction,
	beaconHeight uint64,
	poolPairs map[string]*rawdbv2.PDEPoolForPair,
	privacyV2BreakPoint, pdexv3BreakPoint uint64,
) (bool, uint64, error) {
	if beaconHeight >= privacyV2BreakPoint || beaconHeight >= pdexv3BreakPoint {
		return true, 0, nil
	}

	if len(poolPairs) == 0 || poolPairs == nil {
		return true, 0, nil
	}

	pairKey := string(rawdbv2.BuildPDEPoolForPairKey(beaconHeight, action.Meta.TokenIDToBuyStr, action.Meta.TokenIDToSellStr))
	poolPair, found := poolPairs[pairKey]
	if !found || poolPair.Token1PoolValue == 0 || poolPair.Token2PoolValue == 0 {
		return true, 0, nil
	}

	receiveAmt, newTokenPoolValueToBuy, tempNewTokenPoolValueToSell, err := calcTradeValue(poolPair, action.Meta.TokenIDToSellStr, action.Meta.SellAmount)
	if err != nil {
		return true, 0, nil
	}
	if action.Meta.MinAcceptableAmount > receiveAmt {
		return true, 0, nil
	}

	// update current pde state on mem
	newTokenPoolValueToSell := new(big.Int).SetUint64(tempNewTokenPoolValueToSell)
	fee := action.Meta.TradingFee
	newTokenPoolValueToSell.Add(newTokenPoolValueToSell, new(big.Int).SetUint64(fee))

	poolPair.Token1PoolValue = newTokenPoolValueToBuy
	poolPair.Token2PoolValue = newTokenPoolValueToSell.Uint64()
	if poolPair.Token1IDStr == action.Meta.TokenIDToSellStr {
		poolPair.Token1PoolValue = newTokenPoolValueToSell.Uint64()
		poolPair.Token2PoolValue = newTokenPoolValueToBuy
	}

	return false, receiveAmt, nil
}
