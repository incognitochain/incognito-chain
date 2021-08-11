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

func BuildMatchAndReturnAddLiquidityInstructions(
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
