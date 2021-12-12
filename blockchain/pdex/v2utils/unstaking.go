package v2utils

import (
	"strconv"

	"github.com/incognitochain/incognito-chain/common"
	instruction "github.com/incognitochain/incognito-chain/instruction/pdexv3"
	metadataCommon "github.com/incognitochain/incognito-chain/metadata/common"
	metadataPdexv3 "github.com/incognitochain/incognito-chain/metadata/pdexv3"
)

func BuildRejectUnstakingInstructions(
	metaData metadataPdexv3.UnstakingRequest,
	txReqID common.Hash, shardID byte,
) ([][]string, error) {
	res := [][]string{}
	rejectInst, err := instruction.NewRejectUnstakingWithValue(txReqID, shardID).StringSlice()
	if err != nil {
		return res, err
	}
	res = append(res, rejectInst)
	if metaData.AccessOption.UseNft() {
		mintNftInst, err := instruction.NewMintNftWithValue(
			metaData.AccessOption.NftID,
			metaData.OtaReceivers()[metaData.AccessOption.NftID.String()], shardID, txReqID,
		).StringSlice(strconv.Itoa(metadataCommon.Pdexv3UnstakingRequestMeta))
		if err != nil {
			return res, err
		}
		res = append(res, mintNftInst)
	}

	return res, nil
}

func BuildAcceptUnstakingInstructions(
	stakingPoolID common.Hash, accessOption metadataPdexv3.AccessOption,
	unstakingAmount uint64,
	otaReceiverNft, otaReceiverUnstakingToken string,
	txReqID common.Hash, shardID byte,
	identityID common.Hash,
) ([][]string, error) {
	res := [][]string{}
	acceptInst, err := instruction.NewAcceptUnstakingWithValue(
		stakingPoolID, unstakingAmount, otaReceiverUnstakingToken, txReqID, shardID, accessOption,
	).StringSlice()
	if err != nil {
		return res, err
	}
	res = append(res, acceptInst)
	if accessOption.UseNft() {
		mintNftInst, err := instruction.NewMintNftWithValue(
			accessOption.NftID, otaReceiverNft, shardID, txReqID,
		).
			StringSlice(strconv.Itoa(metadataCommon.Pdexv3UnstakingRequestMeta))
		if err != nil {
			return res, err
		}
		res = append(res, mintNftInst)
	}
	return res, nil
}
