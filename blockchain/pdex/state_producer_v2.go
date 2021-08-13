package pdex

import (
	"encoding/json"
	"errors"
	"math/big"
	"strconv"

	v2 "github.com/incognitochain/incognito-chain/blockchain/pdex/v2utils"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/rawdbv2"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	instruction "github.com/incognitochain/incognito-chain/instruction/pdexv3"
	"github.com/incognitochain/incognito-chain/metadata"
	metadataCommon "github.com/incognitochain/incognito-chain/metadata/common"
	metadataPdexv3 "github.com/incognitochain/incognito-chain/metadata/pdexv3"
	"github.com/incognitochain/incognito-chain/utils"
)

type stateProducerV2 struct {
	stateProducerBase
}

func buildModifyParamsInst(
	params metadataPdexv3.Pdexv3Params,
	shardID byte,
	reqTxID common.Hash,
	status string,
) []string {
	modifyingParamsReqContent := metadataPdexv3.ParamsModifyingContent{
		Content: params,
		TxReqID: reqTxID,
		ShardID: shardID,
	}
	modifyingParamsReqContentBytes, _ := json.Marshal(modifyingParamsReqContent)
	return []string{
		strconv.Itoa(metadataCommon.Pdexv3ModifyParamsMeta),
		strconv.Itoa(int(shardID)),
		status,
		string(modifyingParamsReqContentBytes),
	}
}

func isValidPdexv3Params(params Params) bool {
	if params.DefaultFeeRateBPS > MaxFeeRateBPS {
		return false
	}
	for _, feeRate := range params.FeeRateBPS {
		if feeRate > MaxFeeRateBPS {
			return false
		}
	}
	if params.PRVDiscountPercent > MaxPRVDiscountPercent {
		return false
	}
	if params.TradingStakingPoolRewardPercent+params.TradingProtocolFeePercent > 100 {
		return false
	}
	if params.LimitProtocolFeePercent+params.LimitStakingPoolRewardPercent > 100 {
		return false
	}
	return true
}

func (sp *stateProducerV2) addLiquidity(
	txs []metadata.Transaction,
	beaconHeight uint64,
	poolPairs map[string]*PoolPairState,
	waitingContributions map[string]rawdbv2.Pdexv3Contribution,
	nftIDs map[string]bool,
) (
	[][]string,
	map[string]*PoolPairState,
	map[string]rawdbv2.Pdexv3Contribution,
	map[string]bool, error,
) {
	res := [][]string{}
	for _, tx := range txs {
		shardID := byte(tx.GetValidationEnv().ShardID())
		metaData, _ := tx.GetMetadata().(*metadataPdexv3.AddLiquidityRequest)
		incomingContribution := *NewContributionWithMetaData(*metaData, *tx.Hash(), shardID)

		incomingContributionState := *statedb.NewPdexv3ContributionStateWithValue(
			incomingContribution, metaData.PairHash(),
		)
		if metaData.NftID() != utils.EmptyString && !nftIDs[metaData.NftID()] {
			refundInst, err := instruction.NewRefundAddLiquidityWithValue(incomingContributionState).StringSlice()
			if err != nil {
				return res, poolPairs, waitingContributions, nftIDs, err
			}
			res = append(res, refundInst)
			continue
		}
		waitingContribution, found := waitingContributions[metaData.PairHash()]
		if !found {
			waitingContributions[metaData.PairHash()] = incomingContribution
			inst, err := instruction.NewWaitingAddLiquidityWithValue(incomingContributionState).StringSlice()
			if err != nil {
				return res, poolPairs, waitingContributions, nftIDs, err
			}
			res = append(res, inst)
			continue
		}
		delete(waitingContributions, metaData.PairHash())
		waitingContributionState := *statedb.NewPdexv3ContributionStateWithValue(
			waitingContribution, metaData.PairHash(),
		)
		if waitingContribution.TokenID().String() == incomingContribution.TokenID().String() ||
			waitingContribution.Amplifier() != incomingContribution.Amplifier() ||
			waitingContribution.PoolPairID() != incomingContribution.PoolPairID() ||
			waitingContribution.NftID().String() != incomingContribution.NftID().String() {
			refundInst0, err := instruction.NewRefundAddLiquidityWithValue(waitingContributionState).StringSlice()
			if err != nil {
				return res, poolPairs, waitingContributions, nftIDs, err
			}
			res = append(res, refundInst0)
			refundInst1, err := instruction.NewRefundAddLiquidityWithValue(incomingContributionState).StringSlice()
			if err != nil {
				return res, poolPairs, waitingContributions, nftIDs, err
			}
			res = append(res, refundInst1)
			continue
		}
		nftHash, err := common.Hash{}.NewHashFromStr(metaData.NftID())
		if err != nil {
			return res, poolPairs, waitingContributions, nftIDs, err
		}
		nftID := *nftHash

		poolPairID := utils.EmptyString
		if waitingContribution.PoolPairID() == utils.EmptyString {
			poolPairID = generatePoolPairKey(waitingContribution.TokenID().String(), metaData.TokenID(), waitingContribution.TxReqID().String())
		} else {
			poolPairID = waitingContribution.PoolPairID()
		}
		poolPair, found := poolPairs[poolPairID]
		if !found {
			if waitingContribution.PoolPairID() == utils.EmptyString {
				newPoolPair := initPoolPairState(waitingContribution, incomingContribution)
				tempAmt := big.NewInt(0).Mul(
					big.NewInt(0).SetUint64(waitingContribution.Amount()),
					big.NewInt(0).SetUint64(incomingContribution.Amount()),
				)
				shareAmount := big.NewInt(0).Sqrt(tempAmt).Uint64()
				nftID, nftIDs, err = newPoolPair.addShare(nftID, nftIDs, shareAmount, beaconHeight)
				if err != nil {
					return res, poolPairs, waitingContributions, nftIDs, err
				}
				poolPairs[poolPairID] = newPoolPair
				insts, err := sp.matchAddLiquidity(incomingContributionState, poolPairID, nftID)
				if err != nil {
					return res, poolPairs, waitingContributions, nftIDs, err
				}
				res = append(res, insts...)
				continue
			} else {
				refundInst0, err := instruction.NewRefundAddLiquidityWithValue(waitingContributionState).StringSlice()
				if err != nil {
					return res, poolPairs, waitingContributions, nftIDs, err
				}
				res = append(res, refundInst0)
				refundInst1, err := instruction.NewRefundAddLiquidityWithValue(incomingContributionState).StringSlice()
				if err != nil {
					return res, poolPairs, waitingContributions, nftIDs, err
				}
				res = append(res, refundInst1)
				continue
			}
		}
		token0Contribution, token1Contribution := poolPair.getContributionsByOrder(
			&waitingContribution, &incomingContribution,
		)
		actualToken0ContributionAmount,
			returnedToken0ContributionAmount,
			actualToken1ContributionAmount,
			returnedToken1ContributionAmount := poolPair.
			computeActualContributedAmounts(&token0Contribution, &token1Contribution)

		token0ContributionState := *statedb.NewPdexv3ContributionStateWithValue(
			token0Contribution, metaData.PairHash(),
		)
		token1ContributionState := *statedb.NewPdexv3ContributionStateWithValue(
			token1Contribution, metaData.PairHash(),
		)
		if actualToken0ContributionAmount == 0 || actualToken1ContributionAmount == 0 {
			refundInst0, err := instruction.NewRefundAddLiquidityWithValue(
				token0ContributionState,
			).StringSlice()
			if err != nil {
				return res, poolPairs, waitingContributions, nftIDs, err
			}
			res = append(res, refundInst0)
			refundInst1, err := instruction.NewRefundAddLiquidityWithValue(
				token1ContributionState,
			).StringSlice()
			if err != nil {
				return res, poolPairs, waitingContributions, nftIDs, err
			}
			res = append(res, refundInst1)
			continue
		}
		shareAmount := poolPair.updateReserveAndCalculateShare(
			token0Contribution.TokenID().String(), token1Contribution.TokenID().String(),
			actualToken0ContributionAmount, actualToken1ContributionAmount,
		)
		nftID, nftIDs, err = poolPair.addShare(nftID, nftIDs, shareAmount, beaconHeight)
		if err != nil {
			return res, poolPairs, waitingContributions, nftIDs, err
		}
		insts, err := sp.matchAndReturnAddLiquidity(
			token0ContributionState, token1ContributionState,
			shareAmount, returnedToken0ContributionAmount,
			actualToken0ContributionAmount,
			returnedToken1ContributionAmount,
			actualToken1ContributionAmount,
			nftID,
		)
		res = append(res, insts...)
	}
	return res, poolPairs, waitingContributions, nftIDs, nil
}

func (sp *stateProducerV2) matchAddLiquidity(
	incomingContributionState statedb.Pdexv3ContributionState,
	poolPairID string, nftID common.Hash,
) ([][]string, error) {
	res := [][]string{}
	inst0, err := instruction.NewMatchAddLiquidityWithValue(incomingContributionState, poolPairID, nftID).StringSlice()
	if err != nil {
		return res, err
	}
	res = append(res, inst0)
	if !nftID.IsZeroValue() {
		value := incomingContributionState.Value()
		inst1, err := instruction.NewMintNftWithValue(
			nftID,
			value.ReceiveAddress(),
			value.ShardID(),
		).StringSlice(strconv.Itoa(metadataCommon.Pdexv3AddLiquidityRequestMeta))
		if err != nil {
			return res, err
		}
		res = append(res, inst1)
	}
	return res, nil
}

func (sp *stateProducerV2) matchAndReturnAddLiquidity(
	token0ContributionState, token1ContributionState statedb.Pdexv3ContributionState,
	shareAmount, returnedToken0ContributionAmount, actualToken0ContributionAmount,
	returnedToken1ContributionAmount, actualToken1ContributionAmount uint64,
	nftID common.Hash,
) ([][]string, error) {
	res := [][]string{}
	token0Contribution := token0ContributionState.Value()
	token1Contribution := token1ContributionState.Value()
	matchAndReturnInst0, err := instruction.NewMatchAndReturnAddLiquidityWithValue(
		token0ContributionState, shareAmount, returnedToken0ContributionAmount,
		actualToken1ContributionAmount, returnedToken1ContributionAmount,
		token1Contribution.TokenID(), nftID,
	).StringSlice()
	if err != nil {
		return res, err
	}
	res = append(res, matchAndReturnInst0)
	matchAndReturnInst1, err := instruction.NewMatchAndReturnAddLiquidityWithValue(
		token1ContributionState, shareAmount, returnedToken1ContributionAmount,
		actualToken0ContributionAmount, returnedToken0ContributionAmount,
		token0Contribution.TokenID(), nftID,
	).StringSlice()
	if err != nil {
		return res, err
	}
	res = append(res, matchAndReturnInst1)
	if !nftID.IsZeroValue() {
		inst, err := instruction.NewMintNftWithValue(
			nftID,
			token0Contribution.ReceiveAddress(),
			token0Contribution.ShardID(),
		).StringSlice(strconv.Itoa(metadataCommon.Pdexv3AddLiquidityRequestMeta))
		if err != nil {
			return res, err
		}
		res = append(res, inst)
	}
	return res, nil
}

func (sp *stateProducerV2) modifyParams(
	txs []metadata.Transaction,
	beaconHeight uint64,
	params Params,
) ([][]string, Params, error) {
	instructions := [][]string{}

	for _, tx := range txs {
		shardID := byte(tx.GetValidationEnv().ShardID())
		txReqID := *tx.Hash()
		metaData, ok := tx.GetMetadata().(*metadataPdexv3.ParamsModifyingRequest)
		if !ok {
			return instructions, params, errors.New("Can not parse params modifying metadata")
		}

		// check conditions
		metadataParams := metaData.Pdexv3Params
		newParams := Params(metadataParams)
		isValidParams := isValidPdexv3Params(newParams)

		status := ""
		if isValidParams {
			status = metadataPdexv3.RequestAcceptedChainStatus
			params = newParams
		} else {
			status = metadataPdexv3.RequestRejectedChainStatus
		}

		inst := buildModifyParamsInst(
			metadataParams,
			shardID,
			txReqID,
			status,
		)
		instructions = append(instructions, inst)
	}

	return instructions, params, nil
}

func (sp *stateProducerV2) trade(
	txs []metadata.Transaction,
	pairs map[string]*PoolPairState,
) ([][]string, map[string]*PoolPairState, error) {
	result := [][]string{}
	// TODO: sort
	// tradeRequests := sortByFee(
	// 	tradeRequests,
	// 	beaconHeight,
	// 	pairs,
	// )

	for _, tx := range txs {
		currentTrade, ok := tx.GetMetadata().(*metadataPdexv3.TradeRequest)
		if !ok {
			return result, pairs, errors.New("Cannot parse trade metadata")
		}
		refundReceiver, exists := currentTrade.Receiver[currentTrade.TokenToSell]
		if !exists {
			return result, pairs, errors.New("Refund receiver not found in Trade Request")
		}

		currentAction := instruction.NewAction(
			metadataPdexv3.RefundedTrade{
				Receiver:    refundReceiver,
				TokenToSell: currentTrade.TokenToSell,
				Amount:      currentTrade.SellAmount,
			},
			*tx.Hash(),
			byte(tx.GetValidationEnv().ShardID()), // sender & receiver shard must be the same
		)
		var refundInst []string = currentAction.StringSlice()

		reserves, orderbookList, tradeDirections, tokenToBuy, err :=
			tradePathFromState(currentTrade.TokenToSell, currentTrade.TradePath, pairs)
		tradeOutputReceiver, exists := currentTrade.Receiver[tokenToBuy]
		// anytime the trade handler fails, add a refund instruction
		if err != nil || !exists {
			Logger.log.Warnf("Error preparing trade path: %v", err)
			result = append(result, refundInst)
			continue
		}

		acceptedInst, _, err :=
			v2.MaybeAcceptTrade(currentAction, currentTrade.SellAmount, currentTrade.TradingFee,
				currentTrade.TradePath, tradeOutputReceiver, reserves,
				tradeDirections, tokenToBuy, currentTrade.MinAcceptableAmount, orderbookList)
		if err != nil {
			Logger.log.Warnf("Error handling trade: %v", err)
			result = append(result, refundInst)
			continue
		}
		result = append(result, acceptedInst)
	}
	Logger.log.Warnf("Trade instructions: %v", result)
	return result, pairs, nil
}

func (sp *stateProducerV2) addOrder(
	txs []metadata.Transaction,
	pairs map[string]*PoolPairState,
) ([][]string, map[string]*PoolPairState, error) {
	result := [][]string{}
TransactionLoop:
	for _, tx := range txs {
		currentOrderReq, ok := tx.GetMetadata().(*metadataPdexv3.AddOrderRequest)
		if !ok {
			return result, pairs, errors.New("Cannot parse AddOrder metadata")
		}

		// TODO : PRV-based fee
		refundReceiver, exists := currentOrderReq.Receiver[currentOrderReq.TokenToSell]
		if !exists {
			return result, pairs, errors.New("Refund receiver not found in AddOrder Request")
		}

		currentAction := instruction.NewAction(
			metadataPdexv3.RefundedAddOrder{
				Receiver: refundReceiver,
				TokenID:  currentOrderReq.TokenToSell,
				Amount:   currentOrderReq.SellAmount,
			},
			*tx.Hash(),
			byte(tx.GetValidationEnv().ShardID()), // sender & receiver shard must be the same
		)
		var refundInst []string = currentAction.StringSlice()

		pair, exists := pairs[currentOrderReq.PoolPairID]
		if !exists {
			Logger.log.Warnf("Cannot find pair %s for new order", currentOrderReq.PoolPairID)
			result = append(result, refundInst)
			continue TransactionLoop
		}

		orderID := tx.Hash().String()
		orderbook := pair.orderbook
		for _, ord := range orderbook.orders {
			if ord.Id() == orderID {
				Logger.log.Warnf("Cannot add existing order ID %s", orderID)
				// on any error, append a refund instruction & continue to next tx
				result = append(result, refundInst)
				continue TransactionLoop
			}
		}

		if currentOrderReq.TradingFee >= currentOrderReq.SellAmount {
			Logger.log.Warnf("Order %s cannot afford trading fee of %d", orderID, currentOrderReq.TradingFee)
			result = append(result, refundInst)
			continue TransactionLoop
		}
		// prepare order data
		sellAmountAfterFee := currentOrderReq.SellAmount - currentOrderReq.TradingFee

		var tradeDirection byte
		var token0Rate, token1Rate uint64
		var token0Balance, token1Balance uint64
		if currentOrderReq.TokenToSell == pair.state.Token0ID() {
			tradeDirection = v2.TradeDirectionSell0
			// set order's rates according to request, then set selling token's balance to sellAmount
			// and buying token to 0
			token0Rate = sellAmountAfterFee
			token1Rate = currentOrderReq.MinAcceptableAmount
			token0Balance = sellAmountAfterFee
			token1Balance = 0
		} else {
			tradeDirection = v2.TradeDirectionSell1
			token1Rate = sellAmountAfterFee
			token0Rate = currentOrderReq.MinAcceptableAmount
			token1Balance = sellAmountAfterFee
			token0Balance = 0
		}

		acceptedMd := metadataPdexv3.AcceptedAddOrder{
			PoolPairID:     currentOrderReq.PoolPairID,
			OrderID:        orderID,
			Token0Rate:     token0Rate,
			Token1Rate:     token1Rate,
			Token0Balance:  token0Balance,
			Token1Balance:  token1Balance,
			TradeDirection: tradeDirection,
		}

		acceptedAction := instruction.NewAction(
			&acceptedMd,
			*tx.Hash(),
			byte(tx.GetValidationEnv().ShardID()), // sender & receiver shard must be the same
		)
		result = append(result, acceptedAction.StringSlice())
	}

	Logger.log.Warnf("AddOrder instructions: %v", result)
	return result, pairs, nil
}
