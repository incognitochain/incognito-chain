package pdex

import (
	"errors"

	instruction "github.com/incognitochain/incognito-chain/instruction/pdexv3"
	"github.com/incognitochain/incognito-chain/metadata"
	metadataPdexV3 "github.com/incognitochain/incognito-chain/metadata/pdexv3"
	"github.com/incognitochain/incognito-chain/utils"
)

type stateProducerV2 struct {
	stateProducerBase
}

func (sp *stateProducerV2) addLiquidity(
	txs []metadata.Transaction,
	beaconHeight uint64,
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
		metaData, ok := tx.GetMetadata().(*metadataPdexV3.AddLiquidity)
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
			inst := instruction.NewWaitingAddLiquidityFromMetadata(*metaData, txReqID, shardID).StringArr()
			res = append(res, inst)
			continue
		}
		delete(waitingContributions, metaData.PairHash())
		if waitingContribution.tokenID == metaData.TokenID() ||
			waitingContribution.amplifier != metaData.Amplifier() ||
			waitingContribution.poolPairID != metaData.PoolPairID() {
			refundInst1 := instruction.NewRefundAddLiquidityWithValue(
				metaData.PairHash(), waitingContribution.refundAddress,
				waitingContribution.tokenID, waitingContribution.txReqID,
				waitingContribution.tokenAmount, waitingContribution.shardID,
			).StringArr()
			res = append(res, refundInst1)
			refundInst2 := instruction.NewRefundAddLiquidityWithValue(
				metaData.PairHash(), metaData.RefundAddress(),
				metaData.TokenID(), txReqID,
				waitingContribution.tokenAmount, shardID,
			).StringArr()
			res = append(res, refundInst2)
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
			_, err := poolPair.addContributions(waitingContribution, incomingWaitingContribution)
			if err != nil {
				return res, poolPairs, waitingContributions, err
			}
			//TODO: @tin choose right otaReceiver
			otaReceiver := waitingContribution.receiveAddress
			inst := instruction.NewMatchAddLiquidityWithValue(
				metaData.PairHash(), otaReceiver,
				metaData.TokenID(), txReqID,
				metaData.TokenAmount(), shardID,
			).StringArr()
			res = append(res, inst)
			continue
		}
		token0Contribution, token1Contribution := poolPair.getContributionsByOrder(
			waitingContribution,
			incomingWaitingContribution,
		)
		actualToken0ContributionAmount,
			returnedToken0ContributionAmount,
			actualToken1ContributionAmount,
			returnedToken1ContributionAmount := poolPair.
			computeActualContributedAmounts(token0Contribution, token1Contribution)

		if actualToken0ContributionAmount == 0 || actualToken1ContributionAmount == 0 {
			refundInst1 := instruction.NewRefundAddLiquidityWithValue(
				metaData.PairHash(), token0Contribution.refundAddress,
				token0Contribution.tokenID, token0Contribution.txReqID,
				token0Contribution.tokenAmount, token0Contribution.shardID,
			).StringArr()
			res = append(res, refundInst1)
			refundInst2 := instruction.NewRefundAddLiquidityWithValue(
				metaData.PairHash(), token1Contribution.refundAddress,
				token1Contribution.tokenID, token1Contribution.txReqID,
				token1Contribution.tokenAmount, token1Contribution.shardID,
			).StringArr()
			res = append(res, refundInst2)
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
		matchAndReturnInst0 := instruction.NewMatchAndReturnAddLiquidityWithValue(
			metaData.PairHash(),
			token0Contribution.receiveAddress, token0Contribution.refundAddress,
			token0Contribution.tokenID, nfctID, token0Contribution.txReqID,
			actualToken0ContributionAmount, returnedToken0ContributionAmount,
			token0Contribution.shardID,
		).StringArr()
		res = append(res, matchAndReturnInst0)
		matchAndReturnInst1 := instruction.NewMatchAndReturnAddLiquidityWithValue(
			metaData.PairHash(),
			token1Contribution.receiveAddress, token1Contribution.refundAddress,
			token1Contribution.tokenID, nfctID, token1Contribution.txReqID,
			actualToken1ContributionAmount, returnedToken1ContributionAmount,
			token1Contribution.shardID,
		).StringArr()
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
