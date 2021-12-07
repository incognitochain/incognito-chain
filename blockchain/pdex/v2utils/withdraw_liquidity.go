package v2utils

import (
	"errors"
	"strconv"

	"github.com/incognitochain/incognito-chain/common"
	instruction "github.com/incognitochain/incognito-chain/instruction/pdexv3"
	metadataCommon "github.com/incognitochain/incognito-chain/metadata/common"
	metadataPdexv3 "github.com/incognitochain/incognito-chain/metadata/pdexv3"
	"github.com/incognitochain/incognito-chain/utils"
)

func BuildRejectWithdrawLiquidityInstructions(
	metaData metadataPdexv3.WithdrawLiquidityRequest,
	txReqID common.Hash, shardID byte,
	shouldMintNft bool,
) ([][]string, error) {
	res := [][]string{}
	inst, err := instruction.NewRejectWithdrawLiquidityWithValue(txReqID, shardID).StringSlice()
	if err != nil {
		return res, err
	}
	res = append(res, inst)
	if shouldMintNft {
		nftHash, _ := common.Hash{}.NewHashFromStr(metaData.AccessOption.NftID.String())
		inst, err = instruction.NewMintNftWithValue(
			*nftHash, metaData.OtaReceivers()[metaData.AccessOption.NftID.String()], shardID, txReqID,
		).StringSlice(strconv.Itoa(metadataCommon.Pdexv3WithdrawLiquidityRequestMeta))
		if err != nil {
			return res, err
		}
		res = append(res, inst)
	}
	return res, nil
}

func BuildAcceptWithdrawLiquidityInstructions(
	metaData metadataPdexv3.WithdrawLiquidityRequest,
	token0ID, token1ID common.Hash,
	token0Amount, token1Amount, shareAmount uint64,
	txReqID common.Hash, shardID byte,
	shouldMintNft bool,
) ([][]string, error) {
	res := [][]string{}
	nftHash, err := common.Hash{}.NewHashFromStr(metaData.AccessOption.NftID.String())
	if err != nil {
		return res, err
	}
	if metaData.OtaReceivers()[token0ID.String()] == utils.EmptyString {
		return res, errors.New("invalid ota receivers")
	}
	inst0, err := instruction.NewAcceptWithdrawLiquidityWithValue(
		metaData.PoolPairID(), *nftHash,
		token0ID, token0Amount, shareAmount,
		metaData.OtaReceivers()[token0ID.String()], txReqID, shardID,
	).StringSlice()
	if err != nil {
		return res, err
	}
	res = append(res, inst0)
	if metaData.OtaReceivers()[token1ID.String()] == utils.EmptyString {
		return res, errors.New("invalid ota receivers")
	}
	inst1, err := instruction.NewAcceptWithdrawLiquidityWithValue(
		metaData.PoolPairID(), *nftHash,
		token1ID, token1Amount, shareAmount,
		metaData.OtaReceivers()[token1ID.String()], txReqID, shardID,
	).StringSlice()
	if err != nil {
		return res, err
	}
	res = append(res, inst1)

	if shouldMintNft {
		inst, err := instruction.NewMintNftWithValue(
			*nftHash, metaData.OtaReceivers()[metaData.AccessOption.NftID.String()], shardID, txReqID).
			StringSlice(strconv.Itoa(metadataCommon.Pdexv3WithdrawLiquidityRequestMeta))
		if err != nil {
			return res, err
		}
		res = append(res, inst)
	}
	return res, nil
}
