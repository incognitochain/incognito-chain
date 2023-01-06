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
	"github.com/incognitochain/incognito-chain/config"
	"github.com/incognitochain/incognito-chain/dataaccessobject/rawdbv2"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	instruction "github.com/incognitochain/incognito-chain/instruction/pdexv3"
	"github.com/incognitochain/incognito-chain/metadata"
	metadataCommon "github.com/incognitochain/incognito-chain/metadata/common"
	metadataPdexv3 "github.com/incognitochain/incognito-chain/metadata/pdexv3"
	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/incognitochain/incognito-chain/utils"
	"github.com/incognitochain/incognito-chain/wallet"
)

type stateProducerV2 struct {
	stateProducerBase
}

func (sp *stateProducerV2) addLiquidity(
	txs []metadata.Transaction,
	beaconHeight uint64,
	poolPairs map[string]*PoolPairState,
	waitingContributions map[string]rawdbv2.Pdexv3Contribution,
	nftIDs map[string]uint64,
	params *Params,
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
			Logger.log.Warnf("tx %v not found nftID", tx.Hash().String())
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
			Logger.log.Warnf("tx %v not found previous contribution", tx.Hash().String())
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
			Logger.log.Warnf("tx %v is not valid input", tx.Hash().String())
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
		rootPoolPair, found := poolPairs[poolPairID]

		if !found || rootPoolPair == nil {
			if waitingContribution.PoolPairID() == utils.EmptyString {
				newPoolPair := initPoolPairState(waitingContribution, incomingContribution)
				tempAmt := big.NewInt(0).Mul(
					big.NewInt(0).SetUint64(waitingContribution.Amount()),
					big.NewInt(0).SetUint64(incomingContribution.Amount()),
				)
				shareAmount := big.NewInt(0).Sqrt(tempAmt).Uint64()
				err = newPoolPair.addShare(
					*nftHash,
					shareAmount,
					beaconHeight,
					0,
				)
				if err != nil {
					token0ContributionState := *statedb.NewPdexv3ContributionStateWithValue(
						waitingContribution, metaData.PairHash(),
					)
					token1ContributionState := *statedb.NewPdexv3ContributionStateWithValue(
						incomingContribution, metaData.PairHash(),
					)
					insts, err := v2utils.BuildRefundAddLiquidityInstructions(
						token0ContributionState, token1ContributionState,
					)
					Logger.log.Warnf("tx %v add share err %v", tx.Hash().String(), err)
					res = append(res, insts...)
					continue
				}
				poolPairs[poolPairID] = newPoolPair
				insts, err := v2utils.BuildMatchAddLiquidityInstructions(incomingContributionState, poolPairID, *nftHash)
				if err != nil {
					return res, poolPairs, waitingContributions, err
				}
				Logger.log.Warnf("tx %v is not valid input", tx.Hash().String())
				res = append(res, insts...)
				continue
			} else {
				insts, err := v2utils.BuildRefundAddLiquidityInstructions(
					waitingContributionState, incomingContributionState,
				)
				if err != nil {
					return res, poolPairs, waitingContributions, err
				}
				Logger.log.Warnf("tx %v init a pool pair with poolPairID is not null", tx.Hash().String())
				res = append(res, insts...)
				continue
			}

		} else {
			err := validateTokenIDsByPoolPairID(
				[]common.Hash{waitingContribution.TokenID(), incomingContribution.TokenID()},
				poolPairID,
			)
			if err != nil {
				insts, err := v2utils.BuildRefundAddLiquidityInstructions(
						waitingContributionState, incomingContributionState,
				)
				if err != nil {
						return res, poolPairs, waitingContributions, err
				}
				Logger.log.Warnf("tx %v add token not in pool", tx.Hash().String())
				res = append(res, insts...)
				continue
			}
		}

		token0Contribution, token1Contribution := rootPoolPair.getContributionsByOrder(
			&waitingContribution, &incomingContribution,
		)
		token0ContributionState := *statedb.NewPdexv3ContributionStateWithValue(
			token0Contribution, metaData.PairHash(),
		)
		token1ContributionState := *statedb.NewPdexv3ContributionStateWithValue(
			token1Contribution, metaData.PairHash(),
		)
		actualToken0ContributionAmount,
			returnedToken0ContributionAmount,
			actualToken1ContributionAmount,
			returnedToken1ContributionAmount, err := rootPoolPair.
			computeActualContributedAmounts(&token0Contribution, &token1Contribution)
		if err != nil {
			insts, err := v2utils.BuildRefundAddLiquidityInstructions(
				token0ContributionState, token1ContributionState,
			)
			if err != nil {
				return res, poolPairs, waitingContributions, err
			}
			Logger.log.Warnf("tx %v compute contributed amount err %v", tx.Hash().String(), err)
			res = append(res, insts...)
			continue
		}
		if actualToken0ContributionAmount == 0 || actualToken1ContributionAmount == 0 {
			insts, err := v2utils.BuildRefundAddLiquidityInstructions(
				token0ContributionState, token1ContributionState,
			)
			if err != nil {
				return res, poolPairs, waitingContributions, err
			}
			Logger.log.Warnf("tx %v calculate contribution amount equal to 0", tx.Hash().String())
			res = append(res, insts...)
			continue
		}
		poolPair := rootPoolPair.Clone()
		shareAmount, err := poolPair.addReserveDataAndCalculateShare(
			token0Contribution.TokenID().String(), token1Contribution.TokenID().String(),
			actualToken0ContributionAmount, actualToken1ContributionAmount,
		)
		if err != nil {
			insts, err1 := v2utils.BuildRefundAddLiquidityInstructions(
				token0ContributionState, token1ContributionState,
			)
			if err1 != nil {
				return res, poolPairs, waitingContributions, err
			}
			Logger.log.Warnf("tx %v add reserve data err %v", tx.Hash().String(), err)
			res = append(res, insts...)
			continue
		}
		lmLockedBlocks := uint64(0)
		if _, exists := params.PDEXRewardPoolPairsShare[poolPairID]; exists {
			lmLockedBlocks = params.MiningRewardPendingBlocks
		}
		err = poolPair.addShare(
			*nftHash,
			shareAmount,
			beaconHeight,
			lmLockedBlocks,
		)
		if err != nil {
			insts, err1 := v2utils.BuildRefundAddLiquidityInstructions(
				token0ContributionState, token1ContributionState,
			)
			if err1 != nil {
				return res, poolPairs, waitingContributions, err
			}
			Logger.log.Warnf("tx %v add share err %v:", tx.Hash().String(), err)
			res = append(res, insts...)
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
		poolPairs[poolPairID] = poolPair
		res = append(res, insts...)
	}
	return res, poolPairs, waitingContributions, nil
}

func (sp *stateProducerV2) mintPDEXGenesis() ([][]string, error) {
	receivingAddressStr := config.Param().PDexParams.ProtocolFundAddress
	keyWallet, err := wallet.Base58CheckDeserialize(receivingAddressStr)
	if err != nil {
		return [][]string{}, fmt.Errorf("Can not parse protocol fund address: %v", err)
	}
	if len(keyWallet.KeySet.PaymentAddress.Pk) == 0 {
		return [][]string{}, fmt.Errorf("Protocol fund address is invalid")
	}

	shardID := common.GetShardIDFromLastByte(keyWallet.KeySet.PaymentAddress.Pk[common.PublicKeySize-1])

	mintingPDEXGenesisContent := metadataPdexv3.MintPDEXGenesisContent{
		MintingPaymentAddress: receivingAddressStr,
		MintingAmount:         1,
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
	params *Params,
	pairs map[string]*PoolPairState,
	stakingPools map[string]*StakingPoolState,
) ([][]string, *Params, map[string]*StakingPoolState, error) {
	instructions := [][]string{}

	for _, tx := range txs {
		shardID := byte(tx.GetValidationEnv().ShardID())
		txReqID := *tx.Hash()
		metaData, ok := tx.GetMetadata().(*metadataPdexv3.ParamsModifyingRequest)
		if !ok {
			return instructions, params, stakingPools, errors.New("Can not parse params modifying metadata")
		}

		// check conditions
		metadataParams := metaData.Pdexv3Params
		newParams := Params(metadataParams)
		isValidParams, errorMsg := isValidPdexv3Params(&newParams, pairs)

		status := ""
		if isValidParams {
			status = metadataPdexv3.RequestAcceptedChainStatus
			params = &newParams
			stakingPools = addStakingPoolState(stakingPools, params.StakingPoolsShare)
		} else {
			status = metadataPdexv3.RequestRejectedChainStatus
		}

		inst := v2utils.BuildModifyParamsInst(
			metadataParams,
			errorMsg,
			shardID,
			txReqID,
			status,
		)
		instructions = append(instructions, inst)
	}

	return instructions, params, stakingPools, nil
}

func (sp *stateProducerV2) mintReward(
	tokenID common.Hash,
	mintingAmount uint64,
	params *Params,
	pairs map[string]*PoolPairState,
	isLiquidityMining bool,
) ([][]string, map[string]*PoolPairState, error) {
	instructions := [][]string{}

	totalRewardShare := uint64(0)
	for _, shareAmount := range params.PDEXRewardPoolPairsShare {
		totalRewardShare += uint64(shareAmount)
	}

	// To store the keys in slice in sorted order
	keys := make([]string, len(params.PDEXRewardPoolPairsShare))
	i := 0
	for k := range params.PDEXRewardPoolPairsShare {
		keys[i] = k
		i++
	}
	sort.SliceStable(keys, func(i, j int) bool {
		return keys[i] < keys[j]
	})

	for _, pairID := range keys {
		shareRewardAmount := params.PDEXRewardPoolPairsShare[pairID]
		pair, isExisted := pairs[pairID]
		if !isExisted {
			return instructions, pairs, fmt.Errorf("Could not find pair %v for distributing PDEX reward", pairID)
		}

		// pairReward = mintingAmount * shareRewardAmount / totalRewardShare
		pairReward := new(big.Int).Mul(new(big.Int).SetUint64(mintingAmount), new(big.Int).SetUint64(uint64(shareRewardAmount)))
		pairReward = new(big.Int).Div(pairReward, new(big.Int).SetUint64(totalRewardShare))

		if !pairReward.IsUint64() {
			return instructions, pairs, fmt.Errorf("Could not calculate PDEX reward for pair %v", pairID)
		}

		if pairReward.Uint64() == 0 {
			continue
		}

		orderRewardBPS := params.OrderLiquidityMiningBPS[pairID]
		lpRewardAmt := new(big.Int).Set(pairReward)

		if isLiquidityMining && orderRewardBPS > 0 && pair.makingVolume != nil {
			orderRewardAmt := new(big.Int).Mul(pairReward, new(big.Int).SetUint64(uint64(orderRewardBPS)))
			orderRewardAmt.Div(orderRewardAmt, new(big.Int).SetUint64(uint64(BPS)))

			makingVolumeToken0 := pair.makingVolume[pair.state.Token0ID()]
			if makingVolumeToken0 != nil && makingVolumeToken0.volume != nil && len(makingVolumeToken0.volume) != 0 {
				orderRewards := v2.SplitOrderRewardLiquidityMining(
					makingVolumeToken0.volume,
					orderRewardAmt, tokenID,
				)

				for nftID, reward := range orderRewards {
					if _, ok := pair.orderRewards[nftID]; !ok {
						pair.orderRewards[nftID] = NewOrderReward()
					}
					pair.orderRewards[nftID].AddReward(tokenID, reward)
				}
				lpRewardAmt.Sub(lpRewardAmt, orderRewardAmt)

				delete(pair.makingVolume, pair.state.Token0ID())

				instructions = append(instructions, v2utils.BuildDistributeMiningOrderRewardInsts(
					pairID, pair.state.Token0ID(), orderRewardAmt.Uint64(), tokenID,
				)...)
			}
			makingVolumeToken1 := pair.makingVolume[pair.state.Token1ID()]
			if makingVolumeToken1 != nil && makingVolumeToken1.volume != nil && len(makingVolumeToken1.volume) != 0 {
				orderRewards := v2.SplitOrderRewardLiquidityMining(
					makingVolumeToken1.volume,
					orderRewardAmt, tokenID,
				)

				for nftID, reward := range orderRewards {
					if _, ok := pair.orderRewards[nftID]; !ok {
						pair.orderRewards[nftID] = NewOrderReward()
					}
					pair.orderRewards[nftID].AddReward(tokenID, reward)
				}
				lpRewardAmt.Sub(lpRewardAmt, orderRewardAmt)

				delete(pair.makingVolume, pair.state.Token1ID())

				instructions = append(instructions, v2utils.BuildDistributeMiningOrderRewardInsts(
					pairID, pair.state.Token1ID(), orderRewardAmt.Uint64(), tokenID,
				)...)
			}
		}

		pair.lmRewardsPerShare = v2utils.NewTradingPairWithValue(
			&pair.state,
		).AddLMRewards(
			tokenID, lpRewardAmt, BaseLPFeesPerShare,
			pair.lmRewardsPerShare)

		instructions = append(instructions, v2utils.BuildMintBlockRewardInst(pairID, lpRewardAmt.Uint64(), tokenID)...)
	}

	return instructions, pairs, nil
}

func (sp *stateProducerV2) trade(
	txs []metadata.Transaction,
	pairs map[string]*PoolPairState,
	params *Params,
) ([][]string, map[string]*PoolPairState, error) {
	result := [][]string{}
	var invalidTxs []metadataCommon.Transaction
	var fees, sellAmounts map[string]uint64
	var feeInPRVMap map[string]bool
	var err error
	txs, feeInPRVMap, fees, sellAmounts, invalidTxs, err = getWeightedFee(txs, pairs, params)
	if err != nil {
		return result, pairs, fmt.Errorf("Error converting fee %v", err)
	}
	sort.SliceStable(txs, func(i, j int) bool {
		// compare the fee / sellAmount ratio by comparing products
		fi := big.NewInt(0).SetUint64(fees[txs[i].Hash().String()])
		fi.Mul(fi, big.NewInt(0).SetUint64(sellAmounts[txs[j].Hash().String()]))
		fj := big.NewInt(0).SetUint64(fees[txs[j].Hash().String()])
		fj.Mul(fj, big.NewInt(0).SetUint64(sellAmounts[txs[i].Hash().String()]))

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

		// trading fee must be not less than the minimum trading fee
		if len(currentTrade.TradePath) == 0 {
			Logger.log.Infof("Trade path is empty")
			result = append(result, refundInstructions...)
			continue
		}

		poolFees := []uint{}
		feeRateBPS := uint(0)
		for _, pair := range currentTrade.TradePath {
			poolFee := params.DefaultFeeRateBPS
			if customizedFee, ok := params.FeeRateBPS[pair]; ok {
				poolFee = customizedFee
			}
			poolFees = append(poolFees, poolFee)
			feeRateBPS += poolFee
		}

		// compare the fee / sellAmount ratio with feeRateBPS by comparing products
		feeAmountCompare := new(big.Int).Mul(new(big.Int).SetUint64(fees[tx.Hash().String()]), new(big.Int).SetUint64(BPS))
		sellAmountCompare := new(big.Int).Mul(new(big.Int).SetUint64(sellAmounts[tx.Hash().String()]), new(big.Int).SetUint64(uint64(feeRateBPS)))
		if feeAmountCompare.Cmp(sellAmountCompare) == -1 {
			Logger.log.Infof("Trade fee is not enough")
			result = append(result, refundInstructions...)
			continue
		}

		// get relevant, cloned data from state for the trade path
		reserves, lpFeesPerShares, protocolFees, stakingPoolFees, orderbookList, tradeDirections, tokenToBuy, err :=
			TradePathFromState(currentTrade.TokenToSell, currentTrade.TradePath, pairs)
		tradeOutputReceiver, exists := currentTrade.Receiver[tokenToBuy]
		// anytime the trade handler fails, add a refund instruction
		if err != nil || !exists {
			Logger.log.Warnf("Error preparing trade path: %v", err)
			result = append(result, refundInstructions...)
			continue
		}

		acceptedTradeMd, _, err := v2.MaybeAcceptTrade(
			currentTrade.SellAmount, 0, currentTrade.TradePath,
			tradeOutputReceiver, reserves,
			lpFeesPerShares, protocolFees, stakingPoolFees,
			tradeDirections,
			tokenToBuy, currentTrade.MinAcceptableAmount, orderbookList,
		)
		if err != nil {
			Logger.log.Warnf("Error handling trade: %v", err)
			result = append(result, refundInstructions...)
			continue
		}

		orderRewardsChanges := []map[string]map[common.Hash]uint64{}
		orderMakingChanges := []map[common.Hash]map[string]*big.Int{}
		acceptedTradeMd, orderRewardsChanges, orderMakingChanges, err = v2.TrackFee(
			currentTrade.TradingFee, feeInPRVMap[tx.Hash().String()], currentTrade.TokenToSell, BaseLPFeesPerShare, BPS,
			currentTrade.TradePath, reserves, lpFeesPerShares, protocolFees, stakingPoolFees,
			tradeDirections, orderbookList,
			poolFees, feeRateBPS,
			acceptedTradeMd,
			params.TradingProtocolFeePercent, params.TradingStakingPoolRewardPercent, params.StakingRewardTokens,
			params.DefaultOrderTradingRewardRatioBPS, params.OrderTradingRewardRatioBPS,
		)
		if err != nil {
			Logger.log.Warnf("Error handling fee distribution: %v", err)
			result = append(result, refundInstructions...)
			continue
		}

		// apply state changes for trade consistency in the same block
		for index, pairID := range currentTrade.TradePath {
			changedPair := pairs[pairID]
			changedPair.state = *reserves[index]
			addOrderReward(changedPair.orderRewards, orderRewardsChanges[index])
			if _, ok := params.PDEXRewardPoolPairsShare[pairID]; ok && params.DAOContributingPercent > 0 {
				addMakingVolume(changedPair.makingVolume, orderMakingChanges[index])
			}
			changedPair.lpFeesPerShare = lpFeesPerShares[index]
			changedPair.protocolFees = protocolFees[index]
			changedPair.stakingPoolFees = stakingPoolFees[index]
			orderbook, _ := orderbookList[index].(*Orderbook) // type is determined; see TradePathFromState()
			changedPair.orderbook = *orderbook
			pairs[pairID] = changedPair
		}
		// "accept" instruction
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
	params *Params,
	orderCountByNftID map[string]uint,
) ([][]string, map[string]*PoolPairState, error) {
	result := [][]string{}

TransactionLoop:
	for _, tx := range txs {
		currentOrderReq, ok := tx.GetMetadata().(*metadataPdexv3.AddOrderRequest)
		if !ok {
			return result, pairs, errors.New("Cannot parse AddOrder metadata")
		}
		// sender & receiver shard must be the same
		refundInstructions, err := getRefundedAddOrderInstructions(currentOrderReq,
			*tx.Hash(), byte(tx.GetValidationEnv().ShardID()))
		if err != nil {
			return result, pairs, fmt.Errorf("Error preparing trade refund %v", err)
		}

		// check that the nftID exists
		if _, exists := nftIDs[currentOrderReq.NftID.String()]; !exists {
			Logger.log.Warnf("Cannot find nftID %s for new order", currentOrderReq.NftID.String())
			result = append(result, refundInstructions...)
			continue TransactionLoop
		}
		// check that the nftID has not exceeded its order count limit
		if orderCountByNftID[currentOrderReq.NftID.String()] >= params.MaxOrdersPerNft {
			Logger.log.Warnf("AddOrder: NftID %s has reached order count limit of %d",
				currentOrderReq.NftID.String(), params.MaxOrdersPerNft)
			result = append(result, refundInstructions...)
			continue TransactionLoop
		}

		pair, exists := pairs[currentOrderReq.PoolPairID]
		if !exists {
			Logger.log.Warnf("Cannot find pair %s for new order", currentOrderReq.PoolPairID)
			result = append(result, refundInstructions...)
			continue TransactionLoop
		}
		if v2.HasInsufficientLiquidity(pair.state) {
			Logger.log.Warnf("No liquidity in pair %s", currentOrderReq.PoolPairID)
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

		// prepare order data
		sellAmountAfterFee := currentOrderReq.SellAmount

		var tradeDirection byte
		var token0Rate, token1Rate uint64
		var token0Balance, token1Balance uint64
		var tokenToBuy common.Hash
		switch currentOrderReq.TokenToSell {
		case pair.state.Token0ID():
			tradeDirection = v2.TradeDirectionSell0
			// set order's rates according to request, then set selling token's balance to sellAmount
			// and buying token to 0
			token0Rate = sellAmountAfterFee
			token1Rate = currentOrderReq.MinAcceptableAmount
			token0Balance = sellAmountAfterFee
			token1Balance = 0
			tokenToBuy = pair.state.Token1ID()
		case pair.state.Token1ID():
			tradeDirection = v2.TradeDirectionSell1
			token1Rate = sellAmountAfterFee
			token0Rate = currentOrderReq.MinAcceptableAmount
			token1Balance = sellAmountAfterFee
			token0Balance = 0
			tokenToBuy = pair.state.Token0ID()
		default:
			Logger.log.Warnf("Cannot find pair %s for new order", currentOrderReq.PoolPairID)
			result = append(result, refundInstructions...)
			continue TransactionLoop
		}

		// receivers on order to withdraw to after fully matched
		_, exists = currentOrderReq.Receiver[tokenToBuy]
		if !exists {
			Logger.log.Warnf("Receiver for buying token %v not found for new order", tokenToBuy)
			result = append(result, refundInstructions...)
			continue TransactionLoop
		}
		token0RecvStr, _ := currentOrderReq.Receiver[pair.state.Token0ID()].String()
		token1RecvStr, _ := currentOrderReq.Receiver[pair.state.Token1ID()].String()

		// increment order count to keep same-block requests from exceeding limit
		orderCountByNftID[currentOrderReq.NftID.String()] = orderCountByNftID[currentOrderReq.NftID.String()] + 1

		acceptedMd := metadataPdexv3.AcceptedAddOrder{
			PoolPairID:     currentOrderReq.PoolPairID,
			OrderID:        orderID,
			NftID:          currentOrderReq.NftID,
			Token0Rate:     token0Rate,
			Token1Rate:     token1Rate,
			Token0Balance:  token0Balance,
			Token1Balance:  token1Balance,
			TradeDirection: tradeDirection,
			Receiver:       [2]string{token0RecvStr, token1RecvStr},
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

		// always return NFT in response
		nftReceiver, exists := currentOrderReq.Receiver[currentOrderReq.NftID]
		if !exists {
			return result, pairs, fmt.Errorf("NFT receiver not found in WithdrawOrder Request")
		}
		recvStr, _ := nftReceiver.String() // error handled in tx validation

		mintInstruction, err := instruction.NewMintNftWithValue(
			currentOrderReq.NftID, recvStr, byte(tx.GetValidationEnv().ShardID()), *tx.Hash(),
		).StringSlice(strconv.Itoa(metadataCommon.Pdexv3WithdrawOrderRequestMeta))
		result = append(result, mintInstruction)
		if err != nil {
			return result, pairs, err
		}

		// default to reject
		refundAction := instruction.NewAction(
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
			result = append(result, refundAction.StringSlice())
			continue TransactionLoop
		}

		orderID := currentOrderReq.OrderID
		for _, ord := range pair.orderbook.orders {
			if ord.Id() == orderID {
				if ord.NftID() == currentOrderReq.NftID {
					withdrawResults := make(map[common.Hash]uint64)
					_, withdrawToken0 := currentOrderReq.Receiver[pair.state.Token0ID()]
					_, withdrawToken1 := currentOrderReq.Receiver[pair.state.Token1ID()]
					accepted := false

					if withdrawToken0 && withdrawToken1 {
						if currentOrderReq.Amount != 0 {
							Logger.log.Warnf("Invalid amount %v withdrawing both tokens from order %s (expect %d)",
								currentOrderReq.Amount, orderID, 0)
							result = append(result, refundAction.StringSlice())
							continue TransactionLoop
						}
					}

					// for each token in pool that will be withdrawn, cap withdrawAmount & set new balance in state
					if withdrawToken0 {
						currentBalance := ord.Token0Balance()
						amt := currentOrderReq.Amount
						if currentBalance < amt || amt == 0 {
							amt = currentBalance
						}
						if amt > 0 {
							ord.SetToken0Balance(currentBalance - amt)
							withdrawResults[pair.state.Token0ID()] = amt
							accepted = true
						}
					}
					if withdrawToken1 {
						currentBalance := ord.Token1Balance()
						amt := currentOrderReq.Amount
						if currentBalance < amt || amt == 0 {
							amt = currentBalance
						}
						if amt > 0 {
							ord.SetToken1Balance(currentBalance - amt)
							withdrawResults[pair.state.Token1ID()] = amt
							accepted = true
						}
					}

					if !accepted {
						Logger.log.Warnf("Invalid withdraw tokenID %v for order %s",
							currentOrderReq.Receiver, orderID)
						result = append(result, refundAction.StringSlice())
						continue TransactionLoop
					}
					// apply orderbook changes for withdraw consistency in the same block
					pairs[currentOrderReq.PoolPairID] = pair

					// To store the keys in slice in sorted order
					keys := make([]common.Hash, len(withdrawResults))
					i := 0
					for key := range withdrawResults {
						keys[i] = key
						i++
					}
					sort.SliceStable(keys, func(i, j int) bool {
						return keys[i].String() < keys[j].String()
					})

					// "accepted" metadata
					for _, key := range keys {
						acceptedAction := instruction.NewAction(
							&metadataPdexv3.AcceptedWithdrawOrder{
								PoolPairID: currentOrderReq.PoolPairID,
								OrderID:    currentOrderReq.OrderID,
								Receiver:   currentOrderReq.Receiver[key],
								TokenID:    key,
								Amount:     withdrawResults[key],
							},
							*tx.Hash(),
							byte(tx.GetValidationEnv().ShardID()),
						)
						result = append(result, acceptedAction.StringSlice())
					}
				} else {
					Logger.log.Warnf("Incorrect NftID %v for withdrawing order %s",
						currentOrderReq.NftID, orderID)
					result = append(result, refundAction.StringSlice())
				}
				continue TransactionLoop
			}
		}

		Logger.log.Warnf("No order with ID %s found for withdrawal", orderID)
		result = append(result, refundAction.StringSlice())
	}

	Logger.log.Warnf("WithdrawOrder instructions: %v", result)
	return result, pairs, nil
}

func (sp *stateProducerV2) withdrawAllMatchedOrders(
	pairs map[string]*PoolPairState, limitTxsPerShard uint,
) ([][]string, map[string]*PoolPairState, error) {
	result := [][]string{}
	numberTxsPerShard := make(map[byte]uint)
	pairIDs := getSortedPoolPairIDs(pairs)
	for _, pairID := range pairIDs {
		pair := pairs[pairID] // no need to check found sorted from poolPairs list before
		for _, ord := range pair.orderbook.orders {
			temp := &v2utils.MatchingOrder{ord}
			// check if this order can be matched any further
			if canMatch, err := temp.CanMatch(1 - ord.TradeDirection()); canMatch || err != nil {
				continue
			}

			// an order that isn't further matchable is eligible for automatic withdrawal
			token0Recv := privacy.OTAReceiver{}
			token0Recv.FromString(ord.Token0Receiver()) // error ignored (handled when adding this order)
			token1Recv := privacy.OTAReceiver{}
			token1Recv.FromString(ord.Token1Receiver()) // error ignored (handled when adding this order)

			txHash, _ := common.Hash{}.NewHashFromStr(ord.Id()) // order ID is a valid hash
			shardID := token0Recv.GetShardID()                  // receivers for tokens must belong to the same shard as sender

			var outputInstructions [][]string
			// will withdraw any outstanding balance
			if ord.Token0Balance() > 0 {
				currentBalance := ord.Token0Balance()
				acceptedAction := instruction.NewAction(
					&metadataPdexv3.AcceptedWithdrawOrder{
						PoolPairID: pairID,
						OrderID:    ord.Id(),
						Receiver:   token0Recv,
						TokenID:    pair.state.Token0ID(),
						Amount:     currentBalance,
					},
					*txHash,
					shardID,
				)
				outputInstructions = append(outputInstructions, acceptedAction.StringSlice())
			}
			if ord.Token1Balance() > 0 {
				currentBalance := ord.Token1Balance()
				acceptedAction := instruction.NewAction(
					&metadataPdexv3.AcceptedWithdrawOrder{
						PoolPairID: pairID,
						OrderID:    ord.Id(),
						Receiver:   token1Recv,
						TokenID:    pair.state.Token1ID(),
						Amount:     currentBalance,
					},
					*txHash,
					shardID,
				)
				outputInstructions = append(outputInstructions, acceptedAction.StringSlice())
			}

			if numberTxsPerShard[shardID]+uint(len(outputInstructions)) > limitTxsPerShard {
				continue
			}
			numberTxsPerShard[shardID] += uint(len(outputInstructions))
			// apply orderbook changes & accept withdrawal(s)
			ord.SetToken0Balance(0)
			ord.SetToken1Balance(0)
			pairs[pairID] = pair
			result = append(result, outputInstructions...)
		}
	}

	Logger.log.Warnf("WithdrawAllMatchedOrder instructions: %v", result)
	return result, pairs, nil
}

func (sp *stateProducerV2) withdrawLPFee(
	txs []metadata.Transaction,
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

		_, isExisted := metaData.Receivers[metaData.NftID]
		if !isExisted {
			return instructions, pairs, fmt.Errorf("NFT receiver not found in WithdrawalLPFeeRequest")
		}
		addressStr, err := metaData.Receivers[metaData.NftID].String()
		if err != nil {
			return instructions, pairs, fmt.Errorf("NFT receiver invalid in WithdrawalLPFeeRequest")
		}
		mintNftInst := instruction.NewMintNftWithValue(metaData.NftID, addressStr, shardID, txReqID)
		mintNftInstStr, err := mintNftInst.StringSlice(strconv.Itoa(metadataCommon.Pdexv3WithdrawLPFeeRequestMeta))
		if err != nil {
			return instructions, pairs, fmt.Errorf("Can not parse mint NFT instruction")
		}

		instructions = append(instructions, mintNftInstStr)

		rejectInst := v2utils.BuildWithdrawLPFeeInsts(
			metaData.PoolPairID,
			metaData.NftID,
			map[common.Hash]metadataPdexv3.ReceiverInfo{},
			shardID,
			txReqID,
			metadataPdexv3.RequestRejectedChainStatus,
		)

		// check conditions
		poolPair, isExisted := pairs[metaData.PoolPairID]
		if !isExisted {
			instructions = append(instructions, rejectInst...)
			continue
		}

		lpReward := map[common.Hash]uint64{}
		share, isExistedShare := poolPair.shares[metaData.NftID.String()]
		if isExistedShare {
			// compute amount of received LP reward
			lpReward, err = poolPair.RecomputeLPRewards(metaData.NftID)
			if err != nil {
				return instructions, pairs, fmt.Errorf("Could not track LP reward: %v\n", err)
			}
		}

		orderReward := map[common.Hash]uint64{}
		order, isExistedOrderReward := poolPair.orderRewards[metaData.NftID.String()]
		if isExistedOrderReward {
			// compute amount of received LOP reward
			orderReward = order.uncollectedRewards
		}

		reward := CombineReward(lpReward, orderReward)

		if reward == nil || len(reward) == 0 {
			Logger.log.Infof("No reward to withdraw")
			instructions = append(instructions, rejectInst...)
			continue
		}

		receiversInfo := map[common.Hash]metadataPdexv3.ReceiverInfo{}
		notEnoughOTA := false
		for tokenID := range reward {
			if _, isExisted := metaData.Receivers[tokenID]; !isExisted {
				notEnoughOTA = true
				break
			}
			receiversInfo[tokenID] = metadataPdexv3.ReceiverInfo{
				Address: metaData.Receivers[tokenID],
				Amount:  reward[tokenID],
			}
		}
		if notEnoughOTA {
			Logger.log.Warnf("Not enough OTA in withdraw LP fee request")
			instructions = append(instructions, rejectInst...)
			continue
		}

		acceptedInst := v2utils.BuildWithdrawLPFeeInsts(
			metaData.PoolPairID,
			metaData.NftID,
			receiversInfo,
			shardID,
			txReqID,
			metadataPdexv3.RequestAcceptedChainStatus,
		)

		// update state after fee withdrawal
		if isExistedShare {
			share.tradingFees = resetKeyValueToZero(share.tradingFees)
			share.lastLPFeesPerShare = poolPair.LpFeesPerShare()
			share.lastLmRewardsPerShare = poolPair.LmRewardsPerShare()
		}

		if isExistedOrderReward {
			delete(poolPair.orderRewards, metaData.NftID.String())
		}

		instructions = append(instructions, acceptedInst...)
	}

	return instructions, pairs, nil
}

func (sp *stateProducerV2) withdrawProtocolFee(
	txs []metadata.Transaction,
	pairs map[string]*PoolPairState,
) ([][]string, map[string]*PoolPairState, error) {
	instructions := [][]string{}

	receivingAddressStr := config.Param().PDexParams.ProtocolFundAddress
	keyWallet, err := wallet.Base58CheckDeserialize(receivingAddressStr)
	if err != nil {
		return instructions, pairs, fmt.Errorf("Can not parse protocol fund address: %v", err)
	}
	if len(keyWallet.KeySet.PaymentAddress.Pk) == 0 {
		return instructions, pairs, fmt.Errorf("Protocol fund address is invalid")
	}

	shardID := common.GetShardIDFromLastByte(keyWallet.KeySet.PaymentAddress.Pk[common.PublicKeySize-1])

	for _, tx := range txs {
		txReqID := *tx.Hash()
		metaData, ok := tx.GetMetadata().(*metadataPdexv3.WithdrawalProtocolFeeRequest)
		if !ok {
			return instructions, pairs, errors.New("Can not parse withdrawal protocol fee metadata")
		}

		rejectInst := v2utils.BuildWithdrawProtocolFeeInsts(
			metaData.PoolPairID,
			receivingAddressStr,
			map[common.Hash]uint64{},
			shardID,
			txReqID,
			metadataPdexv3.RequestRejectedChainStatus,
		)

		// check conditions
		pair, isExisted := pairs[metaData.PoolPairID]
		if !isExisted {
			instructions = append(instructions, rejectInst...)
			continue
		}

		rewardAmount := getMapWithoutZeroValue(pair.protocolFees)

		if rewardAmount == nil || len(rewardAmount) == 0 {
			instructions = append(instructions, rejectInst...)
			continue
		}

		acceptedInst := v2utils.BuildWithdrawProtocolFeeInsts(
			metaData.PoolPairID,
			receivingAddressStr,
			rewardAmount,
			shardID,
			txReqID,
			metadataPdexv3.RequestAcceptedChainStatus,
		)

		// update state after fee withdrawal
		pair.protocolFees = resetKeyValueToZero(pair.protocolFees)

		instructions = append(instructions, acceptedInst...)
	}

	return instructions, pairs, nil
}

func (sp *stateProducerV2) withdrawLiquidity(
	txs []metadata.Transaction, poolPairs map[string]*PoolPairState, nftIDs map[string]uint64,
	lmLockedBlocks uint64,
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

		rejectInsts, err := v2utils.BuildRejectWithdrawLiquidityInstructions(*metaData, txReqID, shardID)
		if err != nil {
			return res, poolPairs, err
		}

		_, found := nftIDs[metaData.NftID()]
		if metaData.NftID() == utils.EmptyString || !found {
			Logger.log.Warnf("tx %v not found nftID", tx.Hash().String())
			res = append(res, rejectInsts...)
			continue
		}
		rootPoolPair, ok := poolPairs[metaData.PoolPairID()]
		if !ok || rootPoolPair == nil {
			Logger.log.Warnf("tx %v PoolPairID is not found", tx.Hash().String())
			res = append(res, rejectInsts...)
			continue
		}
		if rootPoolPair.isEmpty() {
			Logger.log.Warnf("tx %v poolPair is empty", tx.Hash().String())
			res = append(res, rejectInsts...)
			continue
		}
		shares, ok := rootPoolPair.shares[metaData.NftID()]
		if !ok || shares == nil {
			Logger.log.Warnf("tx %v not found staker", tx.Hash().String())
			res = append(res, rejectInsts...)
			continue
		}
		if shares.amount == 0 || metaData.ShareAmount() == 0 {
			Logger.log.Warnf("tx %v share amount is invalid", tx.Hash().String())
			res = append(res, rejectInsts...)
			continue
		}
		poolPair := rootPoolPair.Clone()
		token0Amount, token1Amount, shareAmount, err := poolPair.deductShare(
			metaData.NftID(), metaData.ShareAmount(),
		)
		if err != nil {
			Logger.log.Warnf("tx %v deductShare err %v", tx.Hash().String(), err)
			res = append(res, rejectInsts...)
			continue
		}

		insts, err := v2utils.BuildAcceptWithdrawLiquidityInstructions(
			*metaData,
			poolPair.state.Token0ID(), poolPair.state.Token1ID(),
			token0Amount, token1Amount, shareAmount,
			txReqID, shardID)
		if err != nil {
			Logger.log.Warnf("tx %v fail to build accept instruction %v", tx.Hash().String(), err)
			res = append(res, rejectInsts...)
			continue
		}
		res = append(res, insts...)
		poolPairs[metaData.PoolPairID()] = poolPair
	}
	return res, poolPairs, nil
}

func (sp *stateProducerV2) userMintNft(
	txs []metadata.Transaction,
	nftIDs map[string]uint64,
	beaconHeight, mintNftRequireAmount uint64,
) ([][]string, map[string]uint64, uint64, error) {
	res := [][]string{}
	burningPRVAmount := uint64(0)
	for _, tx := range txs {
		shardID := byte(tx.GetValidationEnv().ShardID())
		metaData, _ := tx.GetMetadata().(*metadataPdexv3.UserMintNftRequest)
		txReqID := *tx.Hash()
		inst := []string{}
		var err error
		if metaData.Amount() != mintNftRequireAmount {
			inst, err = instruction.NewRejectUserMintNftWithValue(
				metaData.OtaReceiver(), metaData.Amount(), shardID, txReqID,
			).StringSlice()
			if err != nil {
				return res, nftIDs, burningPRVAmount, err
			}
		} else {
			nftID := genNFT(uint64(len(nftIDs)), beaconHeight)
			nftIDs[nftID.String()] = metaData.Amount()
			inst, err = instruction.NewAcceptUserMintNftWithValue(
				metaData.OtaReceiver(), metaData.Amount(), shardID, nftID, txReqID,
			).StringSlice()
			if err != nil {
				return res, nftIDs, burningPRVAmount, err
			}
			burningPRVAmount += metaData.Amount()
		}
		res = append(res, inst)
	}
	return res, nftIDs, burningPRVAmount, nil
}

func (sp *stateProducerV2) staking(
	txs []metadata.Transaction,
	nftIDs map[string]uint64,
	stakingPoolStates map[string]*StakingPoolState,
	beaconHeight uint64,
) ([][]string, map[string]*StakingPoolState, error) {
	res := [][]string{}
	for _, tx := range txs {
		shardID := byte(tx.GetValidationEnv().ShardID())
		metaData, _ := tx.GetMetadata().(*metadataPdexv3.StakingRequest)
		txReqID := *tx.Hash()
		stakingTokenHash, err := common.Hash{}.NewHashFromStr(metaData.TokenID())
		if err != nil {
			return res, stakingPoolStates, err
		}
		rejectInst, err := instruction.NewRejectStakingWithValue(
			metaData.OtaReceiver(), *stakingTokenHash, txReqID, shardID, metaData.TokenAmount(),
		).StringSlice()
		if err != nil {
			Logger.log.Infof("tx hash %s error %v", txReqID, err)
			return res, stakingPoolStates, err
		}
		rootStakingPoolState, found := stakingPoolStates[metaData.TokenID()]
		if !found || rootStakingPoolState == nil {
			Logger.log.Warnf("tx %v not found poolPair", tx.Hash().String())
			res = append(res, rejectInst)
			continue
		}
		_, found = nftIDs[metaData.NftID()]
		if metaData.NftID() == utils.EmptyString || !found {
			Logger.log.Warnf("tx %v not found nftID ", tx.Hash().String())
			res = append(res, rejectInst)
			continue
		}
		stakingPoolState := rootStakingPoolState.Clone()
		err = stakingPoolState.updateLiquidity(metaData.NftID(), metaData.TokenAmount(), beaconHeight, addOperator)
		if err != nil {
			Logger.log.Warnf("tx %v update liquidity err %v ", tx.Hash().String(), err)
			res = append(res, rejectInst)
			continue
		}
		nftHash, err := common.Hash{}.NewHashFromStr(metaData.NftID())
		if err != nil {
			return res, stakingPoolStates, err
		}
		inst, err := instruction.NewAcceptStakingWtihValue(
			*nftHash, *stakingTokenHash, txReqID, shardID, metaData.TokenAmount(),
		).StringSlice()
		if err != nil {
			return res, stakingPoolStates, err
		}
		res = append(res, inst)
		stakingPoolStates[metaData.TokenID()] = stakingPoolState
	}
	return res, stakingPoolStates, nil
}

func (sp *stateProducerV2) unstaking(
	txs []metadata.Transaction,
	nftIDs map[string]uint64,
	stakingPoolStates map[string]*StakingPoolState,
	beaconHeight uint64,
) ([][]string, map[string]*StakingPoolState, error) {
	res := [][]string{}
	for _, tx := range txs {
		shardID := byte(tx.GetValidationEnv().ShardID())
		metaData, _ := tx.GetMetadata().(*metadataPdexv3.UnstakingRequest)
		txReqID := *tx.Hash()
		stakingPoolID, _ := common.Hash{}.NewHashFromStr(metaData.StakingPoolID())
		rootStakingPoolState, found := stakingPoolStates[metaData.StakingPoolID()]
		rejectInsts, err := v2.BuildRejectUnstakingInstructions(*metaData, txReqID, shardID)
		if err != nil {
			return res, stakingPoolStates, err
		}
		if !found || rootStakingPoolState == nil {
			Logger.log.Warnf("tx %v not found poolPair", tx.Hash().String())
			res = append(res, rejectInsts...)
			continue
		}
		_, found = nftIDs[metaData.NftID()]
		if metaData.NftID() == utils.EmptyString || !found {
			Logger.log.Warnf("tx %v not found nftID", tx.Hash().String())
			res = append(res, rejectInsts...)
			continue
		}
		staker, found := rootStakingPoolState.stakers[metaData.NftID()]
		if !found || staker == nil {
			Logger.log.Warnf("tx %v not found staker", tx.Hash().String())
			res = append(res, rejectInsts...)
			continue
		}
		if staker.liquidity == 0 || metaData.UnstakingAmount() == 0 || rootStakingPoolState.liquidity == 0 {
			Logger.log.Warnf("tx %v unstaking amount is 0", tx.Hash().String())
			res = append(res, rejectInsts...)
			continue
		}
		stakingPoolState := rootStakingPoolState.Clone()
		err = stakingPoolState.updateLiquidity(metaData.NftID(), metaData.UnstakingAmount(), beaconHeight, subOperator)
		if err != nil {
			Logger.log.Warnf("tx %v updateLiquidity err %v", tx.Hash().String(), err)
			res = append(res, rejectInsts...)
			continue
		}
		if metaData.OtaReceivers()[metaData.StakingPoolID()] == utils.EmptyString {
			Logger.log.Warnf("tx %v ota receiver is invalid", tx.Hash().String())
			res = append(res, rejectInsts...)
			continue
		}
		nftHash, _ := common.Hash{}.NewHashFromStr(metaData.NftID())
		insts, err := v2.BuildAcceptUnstakingInstructions(
			*stakingPoolID, *nftHash, metaData.UnstakingAmount(),
			metaData.OtaReceivers()[metaData.NftID()],
			metaData.OtaReceivers()[metaData.StakingPoolID()],
			txReqID, shardID,
		)
		if err != nil {
			return res, stakingPoolStates, err
		}
		res = append(res, insts...)
		stakingPoolStates[metaData.StakingPoolID()] = stakingPoolState
	}
	return res, stakingPoolStates, nil
}

func (sp *stateProducerV2) distributeStakingReward(
	poolPairs map[string]*PoolPairState,
	params *Params,
	stakingPools map[string]*StakingPoolState,
) ([][]string, map[string]*StakingPoolState, error) {
	// Prepare staking reward for distributing
	rewards := map[common.Hash]uint64{}
	for _, poolPair := range poolPairs {
		for tokenID, reward := range poolPair.stakingPoolFees {
			_, ok := rewards[tokenID]
			if !ok {
				rewards[tokenID] = 0
			}
			rewards[tokenID] += reward
		}
	}

	instructions := [][]string{}

	totalRewardShare := uint64(0)
	for _, shareAmount := range params.StakingPoolsShare {
		totalRewardShare += uint64(shareAmount)
	}

	if totalRewardShare == 0 {
		Logger.log.Warnf("Total staking reward share is 0")
		return instructions, stakingPools, nil
	}

	// To store the keys in slice in sorted order
	keys := make([]string, len(params.StakingPoolsShare))
	i := 0
	for k := range params.StakingPoolsShare {
		keys[i] = k
		i++
	}
	sort.SliceStable(keys, func(i, j int) bool {
		return keys[i] < keys[j]
	})

	for _, stakingToken := range keys {
		shareRewardAmount := params.StakingPoolsShare[stakingToken]
		pool, isExisted := stakingPools[stakingToken]
		if !isExisted {
			return instructions, stakingPools, fmt.Errorf("Could not find pool %v for distributing staking reward", stakingToken)
		}

		poolReward := map[common.Hash]uint64{}
		for rewardToken, rewardAmount := range rewards {
			// pairReward = mintingAmount * shareRewardAmount / totalRewardShare
			pairReward := new(big.Int).Mul(new(big.Int).SetUint64(rewardAmount), new(big.Int).SetUint64(uint64(shareRewardAmount)))
			pairReward = new(big.Int).Div(pairReward, new(big.Int).SetUint64(totalRewardShare))

			if !pairReward.IsUint64() {
				return instructions, stakingPools, fmt.Errorf("Could not calculate staking reward for pool %v", stakingToken)
			}

			if pairReward.Uint64() == 0 {
				continue
			}

			poolReward[rewardToken] = pairReward.Uint64()
			pool.AddReward(rewardToken, pairReward.Uint64())
		}
		if len(poolReward) > 0 {
			instructions = append(instructions, v2utils.BuildDistributeStakingRewardInst(stakingToken, poolReward)...)
		}
	}

	return instructions, stakingPools, nil
}

func (sp *stateProducerV2) withdrawStakingReward(
	txs []metadata.Transaction,
	pools map[string]*StakingPoolState,
) ([][]string, map[string]*StakingPoolState, error) {
	instructions := [][]string{}

	for _, tx := range txs {
		shardID := byte(tx.GetValidationEnv().ShardID())
		txReqID := *tx.Hash()
		metaData, ok := tx.GetMetadata().(*metadataPdexv3.WithdrawalStakingRewardRequest)
		if !ok {
			return instructions, pools, errors.New("Can not parse withdrawal staking reward metadata")
		}

		_, isExisted := metaData.Receivers[metaData.NftID]
		if !isExisted {
			return instructions, pools, fmt.Errorf("NFT receiver not found in WithdrawalStakingRewardRequest")
		}
		addressStr, err := metaData.Receivers[metaData.NftID].String()
		if err != nil {
			return instructions, pools, fmt.Errorf("NFT receiver invalid in WithdrawalStakingRewardRequest")
		}
		mintNftInst := instruction.NewMintNftWithValue(metaData.NftID, addressStr, shardID, txReqID)
		mintNftInstStr, err := mintNftInst.StringSlice(strconv.Itoa(metadataCommon.Pdexv3WithdrawStakingRewardRequestMeta))
		if err != nil {
			return instructions, pools, fmt.Errorf("Can not parse mint NFT instruction")
		}

		instructions = append(instructions, mintNftInstStr)

		rejectInst := v2utils.BuildWithdrawStakingRewardInsts(
			metaData.StakingPoolID,
			metaData.NftID,
			map[common.Hash]metadataPdexv3.ReceiverInfo{},
			shardID,
			txReqID,
			metadataPdexv3.RequestRejectedChainStatus,
		)

		// check conditions
		pool, isExisted := pools[metaData.StakingPoolID]
		if !isExisted {
			instructions = append(instructions, rejectInst...)
			continue
		}

		share, isExisted := pool.stakers[metaData.NftID.String()]
		if !isExisted {
			instructions = append(instructions, rejectInst...)
			continue
		}

		// compute amount of received staking reward
		reward, err := pool.RecomputeStakingRewards(metaData.NftID)
		if err != nil {
			return instructions, pools, fmt.Errorf("Could not track staking reward: %v\n", err)
		}

		if reward == nil || len(reward) == 0 {
			Logger.log.Infof("No staking reward to withdraw")
			instructions = append(instructions, rejectInst...)
			continue
		}

		receiversInfo := map[common.Hash]metadataPdexv3.ReceiverInfo{}
		notEnoughOTA := false
		for tokenID := range reward {
			if _, isExisted := metaData.Receivers[tokenID]; !isExisted {
				notEnoughOTA = true
				break
			}
			receiversInfo[tokenID] = metadataPdexv3.ReceiverInfo{
				Address: metaData.Receivers[tokenID],
				Amount:  reward[tokenID],
			}
		}
		if notEnoughOTA {
			Logger.log.Warnf("Not enough OTA in withdrawal staking reward request")
			instructions = append(instructions, rejectInst...)
			continue
		}

		acceptedInst := v2utils.BuildWithdrawStakingRewardInsts(
			metaData.StakingPoolID,
			metaData.NftID,
			receiversInfo,
			shardID,
			txReqID,
			metadataPdexv3.RequestAcceptedChainStatus,
		)

		// update state after fee withdrawal
		share.rewards = resetKeyValueToZero(share.rewards)
		share.lastRewardsPerShare = pool.RewardsPerShare()

		instructions = append(instructions, acceptedInst...)
	}

	return instructions, pools, nil
}
