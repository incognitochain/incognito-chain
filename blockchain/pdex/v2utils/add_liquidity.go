package v2utils

import (
	"strconv"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	instruction "github.com/incognitochain/incognito-chain/instruction/pdexv3"
	metadataCommon "github.com/incognitochain/incognito-chain/metadata/common"
)

func BuildRefundAddLiquidityInstructions(
	waitingContributionState, incomingContributionState statedb.Pdexv3ContributionState,
) ([][]string, error) {
	res := [][]string{}
	refundInst0, err := instruction.NewRefundAddLiquidityWithValue(waitingContributionState).StringSlice()
	if err != nil {
		return res, err
	}
	res = append(res, refundInst0)
	refundInst1, err := instruction.NewRefundAddLiquidityWithValue(incomingContributionState).StringSlice()
	if err != nil {
		return res, err
	}
	res = append(res, refundInst1)
	return res, nil
}

func BuildMatchAddLiquidityInstructions(
	waitingContributionState statedb.Pdexv3ContributionState,
	poolPairID string, txReqID common.Hash, shardID byte,
	shouldMintAccessCoin bool, accessID common.Hash, accessOTA []byte,
) ([][]string, error) {
	res := [][]string{}
	contributionState := waitingContributionState.Value()
	contributionState.SetNftID(accessID)
	contributionState.SetAccessOTA(accessOTA)
	waitingContributionState.SetValue(contributionState)
	inst0, err := instruction.NewMatchAddLiquidityWithValue(
		waitingContributionState, poolPairID,
	).StringSlice()
	if err != nil {
		return res, err
	}
	res = append(res, inst0)
	if shouldMintAccessCoin {
		inst1, err := instruction.NewMintAccessTokenWithValue(
			contributionState.OtaReceiver(), shardID, txReqID,
		).StringSlice(strconv.Itoa(metadataCommon.Pdexv3AddLiquidityRequestMeta))
		if err != nil {
			return res, err
		}
		res = append(res, inst1)
	}
	return res, nil
}

func BuildMatchAndReturnAddLiquidityInstructions(
	token0ContributionState, token1ContributionState statedb.Pdexv3ContributionState,
	shareAmount, returnedToken0ContributionAmount, actualToken0ContributionAmount,
	returnedToken1ContributionAmount, actualToken1ContributionAmount uint64,
	txReqID common.Hash, shardID byte,
	accessOTA []byte, shouldMintAccessCoin bool, accessID common.Hash,
) ([][]string, error) {
	res := [][]string{}
	token0Contribution := token0ContributionState.Value()
	token1Contribution := token1ContributionState.Value()
	token0Contribution.SetNftID(accessID)
	token1Contribution.SetNftID(accessID)
	token0ContributionState.SetValue(token0Contribution)
	token1ContributionState.SetValue(token1Contribution)
	matchAndReturnInst0, err := instruction.NewMatchAndReturnAddLiquidityWithValue(
		token0ContributionState, shareAmount, returnedToken0ContributionAmount,
		actualToken1ContributionAmount, returnedToken1ContributionAmount,
		token1Contribution.TokenID(), accessOTA,
	).StringSlice()
	if err != nil {
		return res, err
	}
	res = append(res, matchAndReturnInst0)
	matchAndReturnInst1, err := instruction.NewMatchAndReturnAddLiquidityWithValue(
		token1ContributionState, shareAmount, returnedToken1ContributionAmount,
		actualToken0ContributionAmount, returnedToken0ContributionAmount,
		token0Contribution.TokenID(), accessOTA,
	).StringSlice()
	if err != nil {
		return res, err
	}
	res = append(res, matchAndReturnInst1)
	if shouldMintAccessCoin {
		mintAccessTokenInst, err := instruction.NewMintAccessTokenWithValue(
			token0Contribution.OtaReceiver(),
			shardID, txReqID,
		).StringSlice(strconv.Itoa(metadataCommon.Pdexv3AddLiquidityRequestMeta))
		if err != nil {
			return res, err
		}
		res = append(res, mintAccessTokenInst)
	}
	return res, nil
}
