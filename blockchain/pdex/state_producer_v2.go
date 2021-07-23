package pdex

import (
	"errors"

	instruction "github.com/incognitochain/incognito-chain/instruction/pdexv3"
	"github.com/incognitochain/incognito-chain/metadata"
	metadataPdexv3 "github.com/incognitochain/incognito-chain/metadata/pdexv3"
	"github.com/incognitochain/incognito-chain/utils"
)

type stateProducerV2 struct {
	stateProducerBase
}

func (sp *stateProducerV2) addLiquidity(
	txs []metadata.Transaction,
	poolPairs map[string]PoolPairState,
	waitingContributions map[string]Contribution,
) (
	[][]string,
	map[string]PoolPairState,
	map[string]Contribution,
	error,
) {
	res := [][]string{}
	for _, tx := range txs {
		shardID := byte(tx.GetValidationEnv().ShardID())
		txReqID := tx.Hash().String()
		metaData, ok := tx.GetMetadata().(*metadataPdexv3.AddLiquidity)
		if !ok {
			return res, poolPairs, waitingContributions, errors.New("Can not parse add liquidity metadata")
		}
		waitingContribution, found := waitingContributions[metaData.PairHash()]
		if !found {
			waitingContributions[metaData.PairHash()] = *NewContributionWithValue(
				metaData.PoolPairID(), metaData.ReceiveAddress(),
				metaData.RefundAddress(), metaData.TokenID(), txReqID,
				metaData.TokenAmount(), metaData.Amplifier(), shardID,
			)
			inst := instruction.NewWaitingAddLiquidityFromMetadata(*metaData, txReqID, shardID).StringSlice()
			res = append(res, inst)
			continue
		}
		delete(waitingContributions, metaData.PairHash())
		waitingContributionMetaData := metadataPdexv3.NewAddLiquidityWithValue(
			waitingContribution.poolPairID, metaData.PairHash(),
			waitingContribution.receiveAddress, waitingContribution.refundAddress,
			waitingContribution.tokenID, waitingContribution.tokenAmount,
			waitingContribution.amplifier,
		)
		if waitingContribution.tokenID == metaData.TokenID() ||
			waitingContribution.amplifier != metaData.Amplifier() ||
			waitingContribution.poolPairID != metaData.PoolPairID() {
			refundInst0 := instruction.NewRefundAddLiquidityFromMetadata(
				*waitingContributionMetaData, waitingContribution.txReqID, waitingContribution.shardID,
			).StringSlice()
			res = append(res, refundInst0)
			refundInst1 := instruction.NewRefundAddLiquidityFromMetadata(
				*metaData, txReqID, shardID,
			).StringSlice()
			res = append(res, refundInst1)
			continue
		}

		poolPairID := utils.EmptyString
		if waitingContribution.poolPairID == utils.EmptyString {
			poolPairID = generatePoolPairKey(waitingContribution.tokenID, metaData.TokenID(), waitingContribution.txReqID)
		} else {
			poolPairID = waitingContribution.poolPairID
		}
		incomingWaitingContribution := *NewContributionWithValue(
			poolPairID, metaData.ReceiveAddress(), metaData.RefundAddress(),
			metaData.TokenID(), txReqID, metaData.TokenAmount(),
			metaData.Amplifier(), shardID,
		)
		poolPair, found := poolPairs[poolPairID]
		if !found {
			poolPairs[poolPairID] = *initPoolPairState(waitingContribution, incomingWaitingContribution)
			poolPair := poolPairs[poolPairID]
			nfctID, err := poolPair.addShare(poolPair.token0RealAmount)
			if err != nil {
				return res, poolPairs, waitingContributions, err
			}
			inst := instruction.NewMatchAddLiquidityFromMetadata(
				*metaData, txReqID, shardID, poolPairID, nfctID,
			).StringSlice()
			res = append(res, inst)
			continue
		}
		token0Contribution, token1Contribution, token0Metadata, token1Metadata := poolPair.getContributionsByOrder(
			&waitingContribution,
			&incomingWaitingContribution,
			waitingContributionMetaData,
			metaData,
		)
		actualToken0ContributionAmount,
			returnedToken0ContributionAmount,
			actualToken1ContributionAmount,
			returnedToken1ContributionAmount := poolPair.
			computeActualContributedAmounts(token0Contribution, token1Contribution)

		if actualToken0ContributionAmount == 0 || actualToken1ContributionAmount == 0 {
			refundInst0 := instruction.NewRefundAddLiquidityFromMetadata(
				token0Metadata, token0Contribution.txReqID, token0Contribution.shardID,
			).StringSlice()
			res = append(res, refundInst0)
			refundInst1 := instruction.NewRefundAddLiquidityFromMetadata(
				token1Metadata, token1Contribution.txReqID, token1Contribution.shardID,
			).StringSlice()
			res = append(res, refundInst1)
			continue
		}

		// change token amount
		token0Contribution.tokenAmount = actualToken0ContributionAmount
		token1Contribution.tokenAmount = actualToken1ContributionAmount
		//
		nfctID, err := poolPair.addContributions(token0Contribution, token1Contribution)
		if err != nil {
			return res, poolPairs, waitingContributions, err
		}
		matchAndReturnInst0 := instruction.NewMatchAndReturnAddLiquidityFromMetadata(
			token0Metadata, token0Contribution.txReqID, token0Contribution.shardID,
			returnedToken0ContributionAmount, nfctID,
		).StringSlice()
		res = append(res, matchAndReturnInst0)
		matchAndReturnInst1 := instruction.NewMatchAndReturnAddLiquidityFromMetadata(
			token1Metadata, token1Contribution.txReqID, token1Contribution.shardID,
			returnedToken1ContributionAmount, nfctID,
		).StringSlice()
		res = append(res, matchAndReturnInst1)
	}

	return res, poolPairs, waitingContributions, nil
}

func (sp *stateProducerV2) modifyParams(
	actions [][]string,
	beaconHeight uint64,
	params Params,
) ([][]string, error) {
	return [][]string{}, nil
}
