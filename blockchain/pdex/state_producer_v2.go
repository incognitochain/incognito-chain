package pdex

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"sort"
	"strconv"

	"github.com/incognitochain/incognito-chain/blockchain/pdex/v2utils"
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
	nftIDs map[string]uint64,
) (
	[][]string, map[string]*PoolPairState, map[string]rawdbv2.Pdexv3Contribution, error,
) {
	res := [][]string{}
	for _, tx := range txs {
		shardID := byte(tx.GetValidationEnv().ShardID())
		metaData, _ := tx.GetMetadata().(*metadataPdexv3.AddLiquidityRequest)
		incomingContribution := *NewContributionWithMetaData(*metaData, *tx.Hash(), shardID)
		incomingContributionState := *statedb.NewPdexv3ContributionStateWithValue(
			incomingContribution, metaData.PairHash(),
		)
		_, found := nftIDs[metaData.NftID()]
		if metaData.NftID() == utils.EmptyString || !found {
			refundInst, err := instruction.NewRefundAddLiquidityWithValue(incomingContributionState).StringSlice()
			if err != nil {
				return res, poolPairs, waitingContributions, err
			}
			res = append(res, refundInst)
			continue
		}
		waitingContribution, found := waitingContributions[metaData.PairHash()]
		if !found {
			waitingContributions[metaData.PairHash()] = incomingContribution
			inst, err := instruction.NewWaitingAddLiquidityWithValue(incomingContributionState).StringSlice()
			if err != nil {
				return res, poolPairs, waitingContributions, err
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
			insts, err := v2utils.BuildRefundAddLiquidityInstructions(
				waitingContributionState, incomingContributionState,
			)
			if err != nil {
				return res, poolPairs, waitingContributions, err
			}
			res = append(res, insts...)
			continue
		}
		nftHash, err := common.Hash{}.NewHashFromStr(metaData.NftID())
		if err != nil {
			return res, poolPairs, waitingContributions, err
		}

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
				err = newPoolPair.addShare(
					*nftHash,
					shareAmount, beaconHeight,
					waitingContribution.TxReqID().String(),
				)
				if err != nil {
					continue
				}
				poolPairs[poolPairID] = newPoolPair
				insts, err := v2utils.BuildMatchAddLiquidityInstructions(incomingContributionState, poolPairID, *nftHash)
				if err != nil {
					return res, poolPairs, waitingContributions, err
				}
				res = append(res, insts...)
				continue
			} else {
				insts, err := v2utils.BuildRefundAddLiquidityInstructions(
					waitingContributionState, incomingContributionState,
				)
				if err != nil {
					return res, poolPairs, waitingContributions, err
				}
				res = append(res, insts...)
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
			insts, err := v2utils.BuildRefundAddLiquidityInstructions(
				token0ContributionState, token1ContributionState,
			)
			if err != nil {
				return res, poolPairs, waitingContributions, err
			}
			res = append(res, insts...)
			continue
		}
		shareAmount, err := poolPair.addReserveDataAndCalculateShare(
			token0Contribution.TokenID().String(), token1Contribution.TokenID().String(),
			actualToken0ContributionAmount, actualToken1ContributionAmount,
		)
		if err != nil {
			Logger.log.Debug("err:", err)
			continue
		}
		err = poolPair.addShare(
			*nftHash,
			shareAmount, beaconHeight,
			waitingContribution.TxReqID().String(),
		)
		if err != nil {
			Logger.log.Debug("err:", err)
			continue
		}
		insts, err := v2utils.BuildMatchAndReturnAddLiquidityInstructions(
			token0ContributionState, token1ContributionState,
			shareAmount, returnedToken0ContributionAmount,
			actualToken0ContributionAmount,
			returnedToken1ContributionAmount,
			actualToken1ContributionAmount,
			*nftHash,
		)
		if err != nil {
			return res, poolPairs, waitingContributions, err
		}
		res = append(res, insts...)
	}
	return res, poolPairs, waitingContributions, nil
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
	params Params,
) ([][]string, map[string]*PoolPairState, error) {
	result := [][]string{}
	var invalidTxs []metadataCommon.Transaction
	var fees, sellAmounts []uint64
	var feeInPRVMap map[string]bool
	var err error
	txs, feeInPRVMap, fees, sellAmounts, invalidTxs, err = getWeightedFee(txs, pairs, params)
	if err != nil {
		return result, pairs, fmt.Errorf("Error converting fee %v", err)
	}
	sort.SliceStable(txs, func(i, j int) bool {
		// compare the fee / sellAmount ratio by comparing products
		fi := big.NewInt(0).SetUint64(fees[i])
		fi.Mul(fi, big.NewInt(0).SetUint64(sellAmounts[j]))
		fj := big.NewInt(0).SetUint64(fees[j])
		fi.Mul(fj, big.NewInt(0).SetUint64(sellAmounts[i]))

		// sort descending
		return fi.Cmp(fj) == 1
	})

	for _, tx := range txs {
		currentTrade, ok := tx.GetMetadata().(*metadataPdexv3.TradeRequest)
		if !ok {
			return result, pairs, errors.New("Cannot parse trade metadata")
		}
		// sender & receiver shard must be the same
		refundInstructions, err := getRefundedTradeInstructions(currentTrade,
			feeInPRVMap[tx.Hash().String()], *tx.Hash(), byte(tx.GetValidationEnv().ShardID()))
		if err != nil {
			return result, pairs, fmt.Errorf("Error preparing trade refund %v", err)
		}

		reserves, orderbookList, tradeDirections, tokenToBuy, err :=
			tradePathFromState(currentTrade.TokenToSell, currentTrade.TradePath, pairs)
		tradeOutputReceiver, exists := currentTrade.Receiver[tokenToBuy]
		// anytime the trade handler fails, add a refund instruction
		if err != nil || !exists {
			Logger.log.Warnf("Error preparing trade path: %v", err)
			result = append(result, refundInstructions...)
			continue
		}

		acceptedTradeMd, _, err := v2.MaybeAcceptTrade(
			currentTrade.SellAmount, currentTrade.TradingFee, currentTrade.TradePath,
			tradeOutputReceiver, reserves, tradeDirections,
			tokenToBuy, currentTrade.MinAcceptableAmount, orderbookList,
		)
		if err != nil {
			Logger.log.Warnf("Error handling trade: %v", err)
			result = append(result, refundInstructions...)
			continue
		}
		action := instruction.NewAction(
			acceptedTradeMd,
			*tx.Hash(),
			byte(tx.GetValidationEnv().ShardID()), // sender & receiver shard must be the same
		)
		result = append(result, action.StringSlice())
	}

	// refund invalid-by-fee tradeRequests
	for _, tx := range invalidTxs {
		currentTrade, ok := tx.GetMetadata().(*metadataPdexv3.TradeRequest)
		if !ok {
			return result, pairs, fmt.Errorf("Cannot parse trade metadata")
		}
		refundInstructions, err := getRefundedTradeInstructions(currentTrade,
			feeInPRVMap[tx.Hash().String()], *tx.Hash(), byte(tx.GetValidationEnv().ShardID()))
		if err != nil {
			return result, pairs, fmt.Errorf("Error preparing trade refund %v", err)
		}
		result = append(result, refundInstructions...)
	}
	Logger.log.Warnf("Trade instructions: %v", result)
	return result, pairs, nil
}

func (sp *stateProducerV2) addOrder(
	txs []metadata.Transaction,
	pairs map[string]*PoolPairState,
	nftIDs map[string]uint64,
	params Params,
) ([][]string, map[string]*PoolPairState, error) {
	result := [][]string{}
	var invalidTxs []metadataCommon.Transaction
	var fees, sellAmounts []uint64
	var feeInPRVMap map[string]bool
	var err error
	txs, feeInPRVMap, fees, sellAmounts, invalidTxs, err = getWeightedFee(txs, pairs, params)
	if err != nil {
		return result, pairs, fmt.Errorf("Error converting fee %v", err)
	}
	sort.SliceStable(txs, func(i, j int) bool {
		// compare the fee / sellAmount ratio by comparing products
		fi := big.NewInt(0).SetUint64(fees[i])
		fi.Mul(fi, big.NewInt(0).SetUint64(sellAmounts[j]))
		fj := big.NewInt(0).SetUint64(fees[j])
		fi.Mul(fj, big.NewInt(0).SetUint64(sellAmounts[i]))

		// sort descending
		return fi.Cmp(fj) == 1
	})

TransactionLoop:
	for _, tx := range txs {
		currentOrderReq, ok := tx.GetMetadata().(*metadataPdexv3.AddOrderRequest)
		if !ok {
			return result, pairs, errors.New("Cannot parse AddOrder metadata")
		}
		// sender & receiver shard must be the same
		refundInstructions, err := getRefundedAddOrderInstructions(currentOrderReq,
			feeInPRVMap[tx.Hash().String()], *tx.Hash(), byte(tx.GetValidationEnv().ShardID()))
		if err != nil {
			return result, pairs, fmt.Errorf("Error preparing trade refund %v", err)
		}

		if _, exists := nftIDs[currentOrderReq.NftID.String()]; !exists {
			Logger.log.Warnf("Cannot find nftID %s for new order", currentOrderReq.NftID.String())
			result = append(result, refundInstructions...)
			continue TransactionLoop
		}

		pair, exists := pairs[currentOrderReq.PoolPairID]
		if !exists {
			Logger.log.Warnf("Cannot find pair %s for new order", currentOrderReq.PoolPairID)
			result = append(result, refundInstructions...)
			continue TransactionLoop
		}

		orderID := tx.Hash().String()
		orderbook := pair.orderbook
		for _, ord := range orderbook.orders {
			if ord.Id() == orderID {
				Logger.log.Warnf("Cannot add existing order ID %s", orderID)
				// on any error, append a refund instruction & continue to next tx
				result = append(result, refundInstructions...)
				continue TransactionLoop
			}
		}

		if currentOrderReq.TradingFee >= currentOrderReq.SellAmount {
			Logger.log.Warnf("Order %s cannot afford trading fee of %d", orderID, currentOrderReq.TradingFee)
			result = append(result, refundInstructions...)
			continue TransactionLoop
		}
		// prepare order data
		sellAmountAfterFee := currentOrderReq.SellAmount

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
			NftID:          currentOrderReq.NftID,
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

	// refund invalid-by-fee addOrder requests
	for _, tx := range invalidTxs {
		currentOrderReq, ok := tx.GetMetadata().(*metadataPdexv3.AddOrderRequest)
		if !ok {
			return result, pairs, fmt.Errorf("Cannot parse AddOrder metadata")
		}
		refundInstructions, err := getRefundedAddOrderInstructions(currentOrderReq,
			feeInPRVMap[tx.Hash().String()], *tx.Hash(), byte(tx.GetValidationEnv().ShardID()))
		if err != nil {
			return result, pairs, fmt.Errorf("Error preparing trade refund %v", err)
		}
		result = append(result, refundInstructions...)
	}

	Logger.log.Warnf("AddOrder instructions: %v", result)
	return result, pairs, nil
}

func (sp *stateProducerV2) withdrawOrder(
	txs []metadata.Transaction,
	pairs map[string]*PoolPairState,
) ([][]string, map[string]*PoolPairState, error) {
	result := [][]string{}
TransactionLoop:
	for _, tx := range txs {
		currentOrderReq, ok := tx.GetMetadata().(*metadataPdexv3.WithdrawOrderRequest)
		if !ok {
			return result, pairs, errors.New("Cannot parse AddOrder metadata")
		}

		// default to reject
		currentAction := instruction.NewAction(
			&metadataPdexv3.RejectedWithdrawOrder{
				PoolPairID: currentOrderReq.PoolPairID,
				OrderID:    currentOrderReq.OrderID,
			},
			*tx.Hash(),
			byte(tx.GetValidationEnv().ShardID()), // sender & receiver shard must be the same
		)

		pair, exists := pairs[currentOrderReq.PoolPairID]
		if !exists {
			Logger.log.Warnf("Cannot find pair %s for new order", currentOrderReq.PoolPairID)
			result = append(result, currentAction.StringSlice())
			continue TransactionLoop
		}

		orderID := currentOrderReq.OrderID
		for _, ord := range pair.orderbook.orders {
			if ord.Id() == orderID {
				if ord.NftID() == currentOrderReq.NftID {
					var currentBalance uint64
					switch currentOrderReq.TokenID {
					case pair.state.Token0ID():
						currentBalance = ord.Token0Balance()
					case pair.state.Token1ID():
						currentBalance = ord.Token1Balance()
					default:
						Logger.log.Warnf("Invalid withdraw tokenID %v for order %s",
							currentOrderReq.TokenID, orderID)
						result = append(result, currentAction.StringSlice())
						continue TransactionLoop
					}

					withdrawAmount := currentOrderReq.Amount
					if currentBalance < currentOrderReq.Amount {
						withdrawAmount = currentBalance
					}
					// accepted
					currentAction.Content = &metadataPdexv3.AcceptedWithdrawOrder{
						PoolPairID: currentOrderReq.PoolPairID,
						OrderID:    currentOrderReq.OrderID,
						Receiver:   currentOrderReq.Receiver,
						TokenID:    currentOrderReq.TokenID,
						Amount:     withdrawAmount,
					}
				} else {
					Logger.log.Warnf("Incorrect NftID %v for withdrawing order %s",
						currentOrderReq.NftID, orderID)
				}
				result = append(result, currentAction.StringSlice())
				continue TransactionLoop
			}
		}

		Logger.log.Warnf("No order with ID %s found for withdrawal", orderID)
		result = append(result, currentAction.StringSlice())
	}

	Logger.log.Warnf("WithdrawOrder instructions: %v", result)
	return result, pairs, nil
}

func (sp *stateProducerV2) withdrawLiquidity(
	txs []metadata.Transaction, poolPairs map[string]*PoolPairState, nftIDs map[string]uint64,
) (
	[][]string,
	map[string]*PoolPairState,
	error,
) {
	res := [][]string{}
	for _, tx := range txs {
		shardID := byte(tx.GetValidationEnv().ShardID())
		metaData, _ := tx.GetMetadata().(*metadataPdexv3.WithdrawLiquidityRequest)
		txReqID := *tx.Hash()
		_, found := nftIDs[metaData.NftID()]
		if metaData.NftID() == utils.EmptyString || !found {
			insts, err := v2utils.BuildRejectWithdrawLiquidityInstructions(*metaData, txReqID, shardID)
			if err != nil {
				return res, poolPairs, err
			}
			res = append(res, insts...)
			continue
		}
		poolPair, ok := poolPairs[metaData.PoolPairID()]
		if !ok || poolPair == nil {
			insts, err := v2utils.BuildRejectWithdrawLiquidityInstructions(*metaData, txReqID, shardID)
			if err != nil {
				return res, poolPairs, err
			}
			res = append(res, insts...)
			continue
		}
		shares, ok := poolPair.shares[metaData.NftID()]
		if !ok || shares == nil {
			insts, err := v2utils.BuildRejectWithdrawLiquidityInstructions(*metaData, txReqID, shardID)
			if err != nil {
				return res, poolPairs, err
			}
			res = append(res, insts...)
			continue
		}
		token0Amount, token1Amount, shareAmount, err := poolPair.deductShare(
			metaData.NftID(), metaData.ShareAmount(),
		)
		if err != nil {
			insts, err := v2utils.BuildRejectWithdrawLiquidityInstructions(*metaData, txReqID, shardID)
			if err != nil {
				return res, poolPairs, err
			}
			res = append(res, insts...)
			continue
		}

		insts, err := v2utils.BuildAcceptWithdrawLiquidityInstructions(
			*metaData,
			poolPair.state.Token0ID(), poolPair.state.Token1ID(),
			token0Amount, token1Amount, shareAmount,
			txReqID, shardID)
		if err != nil {
			return res, poolPairs, err
		}
		res = append(res, insts...)
	}
	return res, poolPairs, nil
}

func (sp *stateProducerV2) userMintNft(
	txs []metadata.Transaction,
	nftIDs map[string]uint64,
	beaconHeight, mintNftRequireAmount uint64,
) ([][]string, map[string]uint64, error) {
	res := [][]string{}
	for _, tx := range txs {
		shardID := byte(tx.GetValidationEnv().ShardID())
		metaData, _ := tx.GetMetadata().(*metadataPdexv3.UserMintNftRequest)
		txReqID := *tx.Hash()
		inst := []string{}
		var err error
		if metaData.Amount() != mintNftRequireAmount {
			inst, err = instruction.NewRejectUserMintNftWithValue(
				metaData.OtaReceive(), metaData.Amount(), shardID, txReqID,
			).StringSlice()
			if err != nil {
				Logger.log.Debugf("Can not reject mint nftID with txHash %s", txReqID.String())
				continue
			}
		} else {
			nftID := genNFT(uint64(len(nftIDs)), beaconHeight)
			nftIDs[nftID.String()] = metaData.Amount()
			inst, err = instruction.NewAcceptUserMintNftWithValue(
				metaData.OtaReceive(), metaData.Amount(), shardID, nftID, txReqID,
			).StringSlice()
			if err != nil {
				Logger.log.Debugf("Can not mint nftID with txHash %s", txReqID.String())
				continue
			}
		}
		res = append(res, inst)
	}
	return res, nftIDs, nil
}
