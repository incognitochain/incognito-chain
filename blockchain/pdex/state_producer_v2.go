package pdex

import (
	"encoding/json"
	"errors"
	"fmt"
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

func (sp *stateProducerV2) addLiquidity(
	txs []metadata.Transaction,
	beaconHeight uint64,
	poolPairs map[string]*PoolPairState,
	waitingContributions map[string]rawdbv2.Pdexv3Contribution,
	nftIDs map[string]bool,
	stateDB *statedb.StateDB,
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
			insts, err := v2utils.BuildRefundAddLiquidityInstructions(
				waitingContributionState, incomingContributionState,
			)
			if err != nil {
				return res, poolPairs, waitingContributions, nftIDs, err
			}
			res = append(res, insts...)
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
				nftID, nftIDs, err = newPoolPair.addShare(
					poolPairID, nftID, nftIDs,
					shareAmount, beaconHeight,
					waitingContribution.TxReqID().String(),
					stateDB,
				)
				if err != nil {
					continue
				}
				poolPairs[poolPairID] = newPoolPair
				insts, err := v2utils.BuildMatchAddLiquidityInstructions(incomingContributionState, poolPairID, nftID)
				if err != nil {
					return res, poolPairs, waitingContributions, nftIDs, err
				}
				res = append(res, insts...)
				continue
			} else {
				insts, err := v2utils.BuildRefundAddLiquidityInstructions(
					waitingContributionState, incomingContributionState,
				)
				if err != nil {
					return res, poolPairs, waitingContributions, nftIDs, err
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
				return res, poolPairs, waitingContributions, nftIDs, err
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
		nftID, nftIDs, err = poolPair.addShare(
			poolPairID, nftID, nftIDs, shareAmount, beaconHeight,
			waitingContribution.TxReqID().String(), stateDB,
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
			nftID,
		)
		if err != nil {
			return res, poolPairs, waitingContributions, nftIDs, err
		}
		res = append(res, insts...)
	}
	return res, poolPairs, waitingContributions, nftIDs, nil
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
	params *Params,
	pairs map[string]*PoolPairState,
	stakingPools map[string]*StakingPoolState,
) ([][]string, *Params, error) {
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
		isValidParams, errorMsg := isValidPdexv3Params(&newParams, pairs, stakingPools)

		status := ""
		if isValidParams {
			status = metadataPdexv3.RequestAcceptedChainStatus
			params = &newParams
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

	return instructions, params, nil
}

func (sp *stateProducerV2) mintPDEX(
	mintingAmount uint64,
	params *Params,
	pairs map[string]*PoolPairState,
) ([][]string, map[string]*PoolPairState, error) {
	instructions := [][]string{}

	totalRewardShare := uint64(0)
	for _, shareAmount := range params.PDEXRewardPoolPairsShare {
		totalRewardShare += uint64(shareAmount)
	}

	for pairID, shareRewardAmount := range params.PDEXRewardPoolPairsShare {
		pair, isExisted := pairs[pairID]
		if !isExisted {
			return instructions, pairs, fmt.Errorf("Could not find pair %v for distributing PDEX reward", pairID)
		}

		// pairReward = mintingAmount * shareRewardAmount / totalRewardShare
		pairReward := new(big.Int).Mul(new(big.Int).SetUint64(mintingAmount), new(big.Int).SetUint64(uint64(shareRewardAmount)))
		pairReward = new(big.Int).Div(pairReward, new(big.Int).SetUint64(totalRewardShare))

		// update state of PDEX token in pool pair state
		oldLPFeesPerShare, isExisted := pair.state.LPFeesPerShare()[common.PDEXCoinID]
		if !isExisted {
			oldLPFeesPerShare = big.NewInt(0)
		}

		// delta (fee / LP share) = pairReward * BASE / totalLPShare
		deltaLPFeesPerShare := new(big.Int).Mul(pairReward, BaseLPFeesPerShare)
		deltaLPFeesPerShare = new(big.Int).Div(deltaLPFeesPerShare, new(big.Int).SetUint64(pair.state.ShareAmount()))

		// update accumulated sum of (fee / LP share)
		newLPFeesPerShare := new(big.Int).Add(oldLPFeesPerShare, deltaLPFeesPerShare)
		tempLPFeesPerShare := pair.state.LPFeesPerShare()
		tempLPFeesPerShare[common.PDEXCoinID] = newLPFeesPerShare

		pair.state.SetLPFeesPerShare(tempLPFeesPerShare)

		instructions = append(instructions, v2utils.BuildMintPDEXInst(pairID, uint(pairReward.Int64()))...)
	}

	return instructions, pairs, nil
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
	stateDB *statedb.StateDB,
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
			metaData.PoolPairID,
			metaData.NftID,
			map[string]metadataPdexv3.ReceiverInfo{
				metadataPdexv3.NftTokenType: {
					TokenID:    metaData.NftID,
					AddressStr: metaData.NftReceiverAddress,
					Amount:     1,
				},
			},
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

		share, isExisted := poolPair.shares[metaData.NftID.String()]
		if !isExisted {
			instructions = append(instructions, rejectInst...)
			continue
		}

		// compute amount of received LP fee
		reward, err := poolPair.RecomputeLPFee(metaData.PoolPairID, metaData.NftID, stateDB)
		if err != nil {
			return instructions, pairs, errors.New("Can not track LP reward")
		}

		acceptedInst := v2utils.BuildWithdrawLPFeeInsts(
			metaData.PoolPairID,
			metaData.NftID,
			map[string]metadataPdexv3.ReceiverInfo{
				metadataPdexv3.Token0Type: {
					TokenID:    poolPair.state.Token0ID(),
					AddressStr: metaData.FeeReceiverAddress.Token0ReceiverAddress,
					Amount:     reward[poolPair.state.Token0ID().String()],
				},
				metadataPdexv3.Token1Type: {
					TokenID:    poolPair.state.Token1ID(),
					AddressStr: metaData.FeeReceiverAddress.Token1ReceiverAddress,
					Amount:     reward[poolPair.state.Token1ID().String()],
				},
				metadataPdexv3.PRVType: {
					TokenID:    common.PRVCoinID,
					AddressStr: metaData.FeeReceiverAddress.PRVReceiverAddress,
					Amount:     reward[common.PRVIDStr],
				},
				metadataPdexv3.PDEXType: {
					TokenID:    common.PDEXCoinID,
					AddressStr: metaData.FeeReceiverAddress.PDEXReceiverAddress,
					Amount:     reward[common.PDEXIDStr],
				},
				metadataPdexv3.NftTokenType: {
					TokenID:    metaData.NftID,
					AddressStr: metaData.NftReceiverAddress,
					Amount:     1,
				},
			},
			shardID,
			txReqID,
			metadataPdexv3.RequestAcceptedChainStatus,
		)

		// update state after fee withdrawal
		share.tradingFees = map[string]uint64{}
		share.lastUpdatedBeaconHeight = beaconHeight

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
			metaData.PoolPairID,
			map[string]metadataPdexv3.ReceiverInfo{},
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

		protocolFeeAmount := func(poolPair *PoolPairState, tokenID common.Hash) uint64 {
			amount, isExisted := poolPair.state.ProtocolFees()[tokenID]
			if !isExisted {
				amount = 0
			}
			return amount
		}

		acceptedInst := v2utils.BuildWithdrawProtocolFeeInsts(
			metaData.PoolPairID,
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

func (sp *stateProducerV2) withdrawLiquidity(
	txs []metadata.Transaction,
	poolPairs map[string]*PoolPairState,
	beaconHeight uint64,
	stateDB *statedb.StateDB,
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
			metaData.PoolPairID(), metaData.NftID(), metaData.ShareAmount(), beaconHeight, stateDB,
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
