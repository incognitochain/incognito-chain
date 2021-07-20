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
				metaData.PoolPairID(), metaData.ReceiverAddress(),
				metaData.RefundAddress(), metaData.TokenID(), txReqID,
				metaData.TokenAmount(), metaData.Amplifier(), shardID,
			)
			inst := instruction.NewWaitingAddLiquidityFromMetadata(*metaData, txReqID, shardID).StringArr()
			res = append(res, inst)
			continue
		}
		if waitingContribution.tokenID == metaData.TokenID() ||
			waitingContribution.amplifier != metaData.Amplifier() ||
			waitingContribution.poolPairID != metaData.PoolPairID() {
			delete(waitingContributions, metaData.PairHash())
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
		poolPair, found := poolPairs[poolPairID]
		incomingWaitingContribution := *NewContributionWithValue(
			poolPairID, metaData.ReceiverAddress(), metaData.RefundAddress(),
			metaData.TokenID(), txReqID, metaData.TokenAmount(),
			metaData.Amplifier(), shardID,
		)

		if !found {
			delete(waitingContributions, metaData.PairHash())

			addLiquidityToPoolPair()
			err := updateWaitingContributionPairToPool(
				beaconHeight,
				waitingContribution,
				incomingWaitingContribution,
				poolPairs,
				shares,
			)
			if err != nil {
				return res, poolPairs, waitingContributions, err
			}
			matchedInst := buildMatchedContributionInst(
				contributionAction,
				metaType,
			)
			res = append(res, matchedInst)
			continue
		}
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
