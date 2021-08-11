package pdex

import (
	"encoding/json"
	"errors"
	"math"
	"math/big"
	"strconv"

	"github.com/incognitochain/incognito-chain/blockchain/pdex/v2utils"
	v2 "github.com/incognitochain/incognito-chain/blockchain/pdex/v2utils"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/config"
	"github.com/incognitochain/incognito-chain/dataaccessobject/rawdbv2"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	instruction "github.com/incognitochain/incognito-chain/instruction/pdexv3"
	"github.com/incognitochain/incognito-chain/metadata"
	metadataCommon "github.com/incognitochain/incognito-chain/metadata/common"
	metadataPdexv3 "github.com/incognitochain/incognito-chain/metadata/pdexv3"
	"github.com/incognitochain/incognito-chain/utils"
	"github.com/incognitochain/incognito-chain/wallet"
)

type stateProducerV2 struct {
	stateProducerBase
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

func (sp *stateProducerV2) mintPDEXGenesis() ([][]string, error) {
	daoPaymentAddressStr := config.Param().IncognitoDAOAddress
	keyWallet, err := wallet.Base58CheckDeserialize(daoPaymentAddressStr)
	if err != nil {
		return [][]string{}, errors.New("Could not deserialize DAO payment address")
	}
	if len(keyWallet.KeySet.PaymentAddress.Pk) == 0 {
		return [][]string{}, errors.New("DAO payment address is invalid")
	}

	shardID := common.GetShardIDFromLastByte(keyWallet.KeySet.PaymentAddress.Pk[common.PublicKeySize-1])

	mintingPDEXGenesisContent := metadataPdexv3.MintPDEXGenesisContent{
		MintingPaymentAddress: daoPaymentAddressStr,
		MintingAmount:         uint64(GenesisMintingAmount * math.Pow(10, common.PDEXDenominatingDecimal)),
		ShardID:               shardID,
	}
	mintingPDEXGenesisContentBytes, _ := json.Marshal(mintingPDEXGenesisContent)

	inst := []string{
		strconv.Itoa(metadataCommon.Pdexv3MintPDEXGenesisMeta),
		strconv.Itoa(int(shardID)),
		metadataPdexv3.RequestAcceptedChainStatus,
		string(mintingPDEXGenesisContentBytes),
	}

	return [][]string{inst}, nil
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

		inst := v2utils.BuildModifyParamsInst(
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
			return result, pairs, errors.New("Can not parse add liquidity metadata")
		}

		currentAction := instruction.NewAction(
			metadataPdexv3.RefundedTrade{
				Receiver:    currentTrade.RefundReceiver,
				TokenToSell: currentTrade.TokenToSell,
				Amount:      currentTrade.SellAmount,
			},
			*tx.Hash(),
			byte(tx.GetValidationEnv().ShardID()), // sender & receiver shard must be the same
		)
		var refundInst []string = currentAction.StringSlice()

		reserves, orderbookList, tradeDirections, tokenToBuy, err :=
			tradePathFromState(currentTrade.TokenToSell, currentTrade.TradePath, pairs)
		// anytime the trade handler fails, add a refund instruction
		if err != nil {
			Logger.log.Warnf("Error preparing trade path: %v", err)
			result = append(result, refundInst)
			continue
		}

		acceptedInst, _, err :=
			v2.MaybeAcceptTrade(currentAction, currentTrade.SellAmount, currentTrade.TradingFee,
				currentTrade.TradePath, currentTrade.Receiver, reserves,
				tradeDirections, tokenToBuy, orderbookList)
		if err != nil {
			Logger.log.Warnf("Error handling trade: %v", err)
			result = append(result, refundInst)
			continue
		}

		result = append(result, acceptedInst)
	}

	return result, pairs, nil
}

func (sp *stateProducerV2) addOrder(
	txs []metadata.Transaction,
	pairs map[string]PoolPairState,
) ([][]string, map[string]PoolPairState, error) {
	result := [][]string{}
	var orderRequests []metadataPdexv3.AddOrderRequest

	for _, tx := range txs {
		item, ok := tx.GetMetadata().(*metadataPdexv3.AddOrderRequest)
		if !ok {
			return result, pairs, errors.New("Can not parse add liquidity metadata")
		}
		orderRequests = append(orderRequests, *item)
	}

	// TODO: sort
	// orderRequests := sortByFee(
	// 	orderRequests,
	// 	beaconHeight,
	// 	pairs,
	// )

	for _, currentOrderReq := range orderRequests {
		_ = currentOrderReq
		// result = append(result, currentInst)
	}

	return result, pairs, nil
}

func (sp *stateProducerV2) withdrawLPFee(
	txs []metadata.Transaction,
	beaconHeight uint64,
	pairs map[string]*PoolPairState,
) ([][]string, map[string]*PoolPairState, error) {
	instructions := [][]string{}

	for _, tx := range txs {
		shardID := byte(tx.GetValidationEnv().ShardID())
		txReqID := *tx.Hash()
		metaData, ok := tx.GetMetadata().(*metadataPdexv3.WithdrawalLPFeeRequest)
		if !ok {
			return instructions, pairs, errors.New("Can not parse withdrawal LP fee metadata")
		}

		rejectInst := v2utils.BuildWithdrawLPFeeInsts(
			metaData.PairID,
			metaData.Index,
			metaData.NfctTokenID,
			map[string]metadataPdexv3.ReceiverInfo{
				metadataPdexv3.NcftTokenType: {
					TokenID:    metaData.NfctTokenID,
					AddressStr: metaData.NfctReceiverAddress,
					Amount:     1,
				},
			},
			shardID,
			txReqID,
			metadataPdexv3.RequestRejectedChainStatus,
		)

		// check conditions
		poolPair, isExisted := pairs[metaData.PairID]
		if !isExisted {
			instructions = append(instructions, rejectInst...)
			continue
		}

		share, isExisted := poolPair.shares[metaData.NfctTokenID.String()]
		if !isExisted {
			instructions = append(instructions, rejectInst...)
			continue
		}
		_, isExisted = share[metaData.Index]
		if !isExisted {
			instructions = append(instructions, rejectInst...)
			continue
		}

		// TODO: compute amount of received LP fee
		acceptedInst := v2utils.BuildWithdrawLPFeeInsts(
			metaData.PairID,
			metaData.Index,
			metaData.NfctTokenID,
			map[string]metadataPdexv3.ReceiverInfo{
				metadataPdexv3.Token0Type: {
					TokenID:    poolPair.state.Token0ID(),
					AddressStr: metaData.FeeReceiverAddress.Token0ReceiverAddress,
					Amount:     1,
				},
				metadataPdexv3.Token1Type: {
					TokenID:    poolPair.state.Token1ID(),
					AddressStr: metaData.FeeReceiverAddress.Token1ReceiverAddress,
					Amount:     1,
				},
				metadataPdexv3.PRVType: {
					TokenID:    common.PRVCoinID,
					AddressStr: metaData.FeeReceiverAddress.PRVReceiverAddress,
					Amount:     1,
				},
				metadataPdexv3.PDEXType: {
					TokenID:    common.PDEXCoinID,
					AddressStr: metaData.FeeReceiverAddress.PDEXReceiverAddress,
					Amount:     1,
				},
				metadataPdexv3.NcftTokenType: {
					TokenID:    metaData.NfctTokenID,
					AddressStr: metaData.NfctReceiverAddress,
					Amount:     1,
				},
			},
			shardID,
			txReqID,
			metadataPdexv3.RequestAcceptedChainStatus,
		)

		// update state after fee withdrawal
		share[metaData.Index].tradingFees = map[string]uint64{}
		share[metaData.Index].lastUpdatedBeaconHeight = beaconHeight

		instructions = append(instructions, acceptedInst...)
	}

	return instructions, pairs, nil
}

func (sp *stateProducerV2) withdrawProtocolFee(
	txs []metadata.Transaction,
	pairs map[string]*PoolPairState,
) ([][]string, map[string]*PoolPairState, error) {
	instructions := [][]string{}

	for _, tx := range txs {
		shardID := byte(tx.GetValidationEnv().ShardID())
		txReqID := *tx.Hash()
		metaData, ok := tx.GetMetadata().(*metadataPdexv3.WithdrawalProtocolFeeRequest)
		if !ok {
			return instructions, pairs, errors.New("Can not parse withdrawal protocol fee metadata")
		}

		rejectInst := v2utils.BuildWithdrawProtocolFeeInsts(
			metaData.PairID,
			map[string]metadataPdexv3.ReceiverInfo{},
			shardID,
			txReqID,
			metadataPdexv3.RequestRejectedChainStatus,
		)

		// check conditions
		pair, isExisted := pairs[metaData.PairID]
		if !isExisted {
			instructions = append(instructions, rejectInst...)
			continue
		}

		protocolFeeAmount := func(poolPair *PoolPairState, tokenID common.Hash) uint64 {
			amount, isExisted := poolPair.state.ProtocolFees()[tokenID]
			if !isExisted {
				amount = 0
			}
			return amount
		}

		acceptedInst := v2utils.BuildWithdrawProtocolFeeInsts(
			metaData.PairID,
			map[string]metadataPdexv3.ReceiverInfo{
				metadataPdexv3.Token0Type: {
					TokenID:    pair.state.Token0ID(),
					AddressStr: metaData.FeeReceiverAddress.Token0ReceiverAddress,
					Amount:     protocolFeeAmount(pair, pair.state.Token0ID()),
				},
				metadataPdexv3.Token1Type: {
					TokenID:    pair.state.Token1ID(),
					AddressStr: metaData.FeeReceiverAddress.Token1ReceiverAddress,
					Amount:     protocolFeeAmount(pair, pair.state.Token1ID()),
				},
				metadataPdexv3.PRVType: {
					TokenID:    common.PRVCoinID,
					AddressStr: metaData.FeeReceiverAddress.PRVReceiverAddress,
					Amount:     protocolFeeAmount(pair, common.PRVCoinID),
				},
				metadataPdexv3.PDEXType: {
					TokenID:    common.PDEXCoinID,
					AddressStr: metaData.FeeReceiverAddress.PDEXReceiverAddress,
					Amount:     protocolFeeAmount(pair, common.PDEXCoinID),
				},
			},
			shardID,
			txReqID,
			metadataPdexv3.RequestAcceptedChainStatus,
		)

		// update state after fee withdrawal
		pair.state.SetProtocolFees(map[common.Hash]uint64{})

		instructions = append(instructions, acceptedInst...)
	}

	return instructions, pairs, nil
}
