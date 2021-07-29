package pdex

import (
	"encoding/json"
	"math/big"
	"strconv"

	"errors"

	"github.com/incognitochain/incognito-chain/common"
	metadataCommon "github.com/incognitochain/incognito-chain/metadata/common"

	"github.com/incognitochain/incognito-chain/dataaccessobject/rawdbv2"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	instruction "github.com/incognitochain/incognito-chain/instruction/pdexv3"
	"github.com/incognitochain/incognito-chain/metadata"
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
	poolPairs map[string]PoolPairState,
	waitingContributions map[string]rawdbv2.Pdexv3Contribution,
) (
	[][]string,
	map[string]PoolPairState,
	map[string]rawdbv2.Pdexv3Contribution,
	error,
) {
	res := [][]string{}
	for _, tx := range txs {
		shardID := byte(tx.GetValidationEnv().ShardID())
		metaData, ok := tx.GetMetadata().(*metadataPdexv3.AddLiquidity)
		if !ok {
			return res, poolPairs, waitingContributions, errors.New("Can not parse add liquidity metadata")
		}
		incomingContribution := *NewContributionWithMetaData(*metaData, *tx.Hash(), shardID)
		incomingContributionState := *statedb.NewPdexv3ContributionStateWithValue(
			incomingContribution, metaData.PairHash(),
		)
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
			waitingContribution.PoolPairID() != incomingContribution.PoolPairID() {
			refundInst0, err := instruction.NewRefundAddLiquidityWithValue(waitingContributionState).StringSlice()
			if err != nil {
				return res, poolPairs, waitingContributions, err
			}
			res = append(res, refundInst0)
			refundInst1, err := instruction.NewRefundAddLiquidityWithValue(incomingContributionState).StringSlice()
			if err != nil {
				return res, poolPairs, waitingContributions, err
			}
			res = append(res, refundInst1)
			continue
		}

		poolPairID := utils.EmptyString
		if waitingContribution.PoolPairID() == utils.EmptyString {
			poolPairID = generatePoolPairKey(waitingContribution.TokenID().String(), metaData.TokenID(), waitingContribution.TxReqID().String())
		} else {
			poolPairID = waitingContribution.PoolPairID()
		}
		poolPair, found := poolPairs[poolPairID]
		if !found {
			newPoolPair := *initPoolPairState(waitingContribution, incomingContribution)
			tempAmt := big.NewInt(0).Mul(
				big.NewInt(0).SetUint64(waitingContribution.Amount()),
				big.NewInt(0).SetUint64(incomingContribution.Amount()),
			)
			shareAmount := big.NewInt(0).Sqrt(tempAmt).Uint64()
			nfctID := poolPair.addShare(poolPairID, shareAmount, beaconHeight)
			inst, err := instruction.NewMatchAddLiquidityWithValue(
				incomingContributionState, poolPairID, nfctID,
			).StringSlice()
			if err != nil {
				return res, poolPairs, waitingContributions, err
			}
			res = append(res, inst)
			poolPairs[poolPairID] = newPoolPair
			continue
		}
		token0Contribution, token1Contribution := poolPair.getContributionsByOrder(
			&waitingContribution,
			&incomingContribution,
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
				return res, poolPairs, waitingContributions, err
			}
			res = append(res, refundInst0)
			refundInst1, err := instruction.NewRefundAddLiquidityWithValue(
				token1ContributionState,
			).StringSlice()
			if err != nil {
				return res, poolPairs, waitingContributions, err
			}
			res = append(res, refundInst1)
			continue
		}

		shareAmount := poolPair.updateReserveAndShares(
			token0Contribution.TokenID().String(), token1Contribution.TokenID().String(),
			actualToken0ContributionAmount, actualToken1ContributionAmount,
		)
		nfctID := poolPair.addShare(poolPairID, shareAmount, beaconHeight)
		matchAndReturnInst0, err := instruction.NewMatchAndReturnAddLiquidityWithValue(
			token0ContributionState, shareAmount, returnedToken0ContributionAmount,
			actualToken1ContributionAmount, returnedToken1ContributionAmount,
			token1Contribution.TokenID(), nfctID,
		).StringSlice()
		if err != nil {
			return res, poolPairs, waitingContributions, err
		}
		res = append(res, matchAndReturnInst0)
		matchAndReturnInst1, err := instruction.NewMatchAndReturnAddLiquidityWithValue(
			token1ContributionState, shareAmount, returnedToken1ContributionAmount,
			actualToken0ContributionAmount, returnedToken0ContributionAmount,
			token0Contribution.TokenID(), nfctID,
		).StringSlice()
		if err != nil {
			return res, poolPairs, waitingContributions, err
		}
		res = append(res, matchAndReturnInst1)
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
