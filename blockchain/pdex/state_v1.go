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

type stateV1 struct {
	waitingContributions        map[string]*rawdbv2.PDEContribution
	deletedWaitingContributions map[string]*rawdbv2.PDEContribution
	poolPairs                   map[string]*rawdbv2.PDEPoolForPair
	shares                      map[string]uint64
	tradingFees                 map[string]uint64
}

func (s *stateV1) Version() uint {
	return BasicVersion
}

func (s *stateV1) Clone() State {
	var state State
	return state
}

func (s *stateV1) Update(env StateEnvironment) ([][]string, error) {
	instructions := [][]string{}

	// handle fee withdrawal
	tempInstructions, err := s.buildInstructionsForFeeWithdrawal(
		env.FeeWithdrawalActions(),
		env.BeaconHeight(),
	)
	if err != nil {
		return instructions, err
	}
	instructions = append(instructions, tempInstructions...)

	// handle trade
	tempInstructions, err = s.buildInstructionsForTrade(
		env.TradeActions(),
		env.BeaconHeight(),
	)
	if err != nil {
		return instructions, err
	}
	instructions = append(instructions, tempInstructions...)

	// handle cross pool trade

	return instructions, nil
}

func (s *stateV1) buildInstructionsForCrossPoolTrade(
	actions [][]string,
	beaconHeight uint64,
) ([][]string, error) {
	res := [][]string{}
	return res, nil
}

func (s *stateV1) buildInstructionsForFeeWithdrawal(
	actions [][]string,
	beaconHeight uint64,
) ([][]string, error) {
	res := [][]string{}
	for _, action := range actions {
		contentStr := action[1]
		contentBytes, err := base64.StdEncoding.DecodeString(contentStr)
		if err != nil {
			Logger.Errorf("ERROR: an error occured while decoding content string of pde withdrawal action: %+v", err)
			return utils.EmptyStringMatrix, nil
		}
		var feeWithdrawalRequestAction metadata.PDEFeeWithdrawalRequestAction
		err = json.Unmarshal(contentBytes, &feeWithdrawalRequestAction)
		if err != nil {
			Logger.Errorf("ERROR: an error occured while unmarshaling pde fee withdrawal request action: %+v", err)
			return utils.EmptyStringMatrix, nil
		}
		wdMeta := feeWithdrawalRequestAction.Meta
		tradingFeeKeyBytes, err := rawdbv2.BuildPDETradingFeeKey(
			beaconHeight,
			wdMeta.WithdrawalToken1IDStr,
			wdMeta.WithdrawalToken2IDStr,
			wdMeta.WithdrawerAddressStr,
		)
		if err != nil {
			Logger.Errorf("cannot build PDETradingFeeKey for address: %v. Error: %v\n", wdMeta.WithdrawerAddressStr, err)
			return utils.EmptyStringMatrix, err
		}
		tradingFeeKey := string(tradingFeeKeyBytes)
		withdrawableFee, found := s.tradingFees[tradingFeeKey]
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
		s.tradingFees[tradingFeeKey] -= wdMeta.WithdrawalFeeAmt
		acceptedInst := []string{
			strconv.Itoa(metadata.PDEFeeWithdrawalRequestMeta),
			strconv.Itoa(int(feeWithdrawalRequestAction.ShardID)),
			common.PDEFeeWithdrawalAcceptedChainStatus,
			contentStr,
		}
		res = append(res, acceptedInst)
	}
	return res, nil
}

func (s *stateV1) buildInstructionsForTrade(
	actions [][]string,
	beaconHeight uint64,
) ([][]string, error) {
	res := [][]string{}

	// handle trade
	sortedTradesActions := s.sortPDETradeInstsByFee(
		actions,
		beaconHeight,
	)
	for _, tradeAction := range sortedTradesActions {
		should, receiveAmount, err := s.shouldRefundTradeAction(tradeAction, beaconHeight)
		if err != nil {
			continue
		}
		if should {
			actionContentBytes, err := json.Marshal(tradeAction)
			if err != nil {
				return utils.EmptyStringMatrix, err
			}
			actionStr := base64.StdEncoding.EncodeToString(actionContentBytes)
			inst := []string{
				strconv.Itoa(metadata.PDETradeRequestMeta),
				strconv.Itoa(int(tradeAction.ShardID)),
				common.PDETradeRefundChainStatus,
				actionStr,
			}
			res = append(res, inst)
			continue
		}
		inst, err := s.buildAcceptedTradeInstruction(tradeAction, beaconHeight, receiveAmount)
		if err != nil {
			continue
		}
		res = append(res, inst)
	}

	return res, nil
}

func (s *stateV1) buildAcceptedTradeInstruction(
	action metadata.PDETradeRequestAction,
	beaconHeight, receiveAmount uint64,
) ([]string, error) {

	pairKey := string(rawdbv2.BuildPDEPoolForPairKey(beaconHeight, action.Meta.TokenIDToBuyStr, action.Meta.TokenIDToSellStr))
	poolPair := s.poolPairs[pairKey]

	pdeTradeAcceptedContent := metadata.PDETradeAcceptedContent{
		TraderAddressStr: action.Meta.TraderAddressStr,
		TxRandomStr:      action.Meta.TxRandomStr,
		TokenIDToBuyStr:  action.Meta.TokenIDToBuyStr,
		ReceiveAmount:    receiveAmount,
		Token1IDStr:      poolPair.Token1IDStr,
		Token2IDStr:      poolPair.Token2IDStr,
		ShardID:          action.ShardID,
		RequestedTxID:    action.TxReqID,
	}
	pdeTradeAcceptedContent.Token1PoolValueOperation = metadata.TokenPoolValueOperation{
		Operator: "-",
		Value:    receiveAmount,
	}
	pdeTradeAcceptedContent.Token2PoolValueOperation = metadata.TokenPoolValueOperation{
		Operator: "+",
		Value:    action.Meta.SellAmount + action.Meta.TradingFee,
	}
	if poolPair.Token1IDStr == action.Meta.TokenIDToSellStr {
		pdeTradeAcceptedContent.Token1PoolValueOperation = metadata.TokenPoolValueOperation{
			Operator: "+",
			Value:    action.Meta.SellAmount + action.Meta.TradingFee,
		}
		pdeTradeAcceptedContent.Token2PoolValueOperation = metadata.TokenPoolValueOperation{
			Operator: "-",
			Value:    receiveAmount,
		}
	}
	pdeTradeAcceptedContentBytes, err := json.Marshal(pdeTradeAcceptedContent)
	if err != nil {
		Logger.Errorf("ERROR: an error occured while marshaling pdeTradeAcceptedContent: %+v", err)
		return utils.EmptyStringArray, err
	}
	return []string{
		strconv.Itoa(metadata.PDETradeRequestMeta),
		strconv.Itoa(int(action.ShardID)),
		common.PDETradeAcceptedChainStatus,
		string(pdeTradeAcceptedContentBytes),
	}, nil
}

func (s *stateV1) sortPDETradeInstsByFee(
	actions [][]string,
	beaconHeight uint64,
) []metadata.PDETradeRequestAction {
	// TODO: @tin improve here for v2 by sorting only with fee not necessary with poolPairs sort
	tradesByPairs := make(map[string][]metadata.PDETradeRequestAction)

	for _, action := range actions {
		contentStr := action[1]
		contentBytes, err := base64.StdEncoding.DecodeString(contentStr)
		if err != nil {
			Logger.Errorf("ERROR: an error occured while decoding content string of pde trade action: %+v", err)
			continue
		}
		pdeTradeReqAction := metadata.PDETradeRequestAction{}
		err = json.Unmarshal(contentBytes, &pdeTradeReqAction)
		if err != nil {
			Logger.Errorf("ERROR: an error occured while unmarshaling pde trade action: %+v", err)
			continue
		}
		tradeMeta := pdeTradeReqAction.Meta
		poolPairKey := string(rawdbv2.BuildPDEPoolForPairKey(beaconHeight, tradeMeta.TokenIDToBuyStr, tradeMeta.TokenIDToSellStr))
		tradesByPairs[poolPairKey] = append(tradesByPairs[poolPairKey], pdeTradeReqAction)
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
		poolPair, found := s.poolPairs[poolPairKey]
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

func (s *stateV1) shouldRefundTradeAction(
	action metadata.PDETradeRequestAction,
	beaconHeight uint64,
) (bool, uint64, error) {

	if len(s.poolPairs) == 0 || s.poolPairs == nil {
		return true, 0, nil
	}

	pairKey := string(rawdbv2.BuildPDEPoolForPairKey(beaconHeight, action.Meta.TokenIDToBuyStr, action.Meta.TokenIDToSellStr))
	poolPair, found := s.poolPairs[pairKey]
	if !found || poolPair.Token1PoolValue == 0 || poolPair.Token2PoolValue == 0 {
		return true, 0, nil
	}

	tokenPoolValueToBuy := poolPair.Token1PoolValue
	tokenPoolValueToSell := poolPair.Token2PoolValue
	if poolPair.Token1IDStr == action.Meta.TokenIDToSellStr {
		tokenPoolValueToSell = poolPair.Token1PoolValue
		tokenPoolValueToBuy = poolPair.Token2PoolValue
	}
	invariant := big.NewInt(0)
	invariant.Mul(new(big.Int).SetUint64(tokenPoolValueToSell), new(big.Int).SetUint64(tokenPoolValueToBuy))
	newTokenPoolValueToSell := big.NewInt(0)
	newTokenPoolValueToSell.Add(new(big.Int).SetUint64(tokenPoolValueToSell), new(big.Int).SetUint64(action.Meta.SellAmount))

	newTokenPoolValueToBuy := big.NewInt(0).Div(invariant, newTokenPoolValueToSell).Uint64()
	modValue := big.NewInt(0).Mod(invariant, newTokenPoolValueToSell)
	if modValue.Cmp(big.NewInt(0)) != 0 {
		newTokenPoolValueToBuy++
	}
	if tokenPoolValueToBuy <= newTokenPoolValueToBuy {
		return true, 0, nil
	}

	receiveAmt := tokenPoolValueToBuy - newTokenPoolValueToBuy
	if action.Meta.MinAcceptableAmount > receiveAmt {
		return true, 0, nil
	}

	// update current pde state on mem
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

func (s *stateV1) buildInstsForCrossPoolTrade(actions [][]string) ([][]string, error) {
	res := [][]string{}
	return res, nil
}

func (s *stateV1) buildInstructionsForWithdrawal(actions [][]string) ([][]string, error) {
	res := [][]string{}
	return res, nil
}

func (s *stateV1) buildInstructionsForContribution(actions [][]string) ([][]string, error) {
	res := [][]string{}
	return res, nil
}

func (s *stateV1) Upgrade(env StateEnvironment) State {
	var state State
	return state
}
