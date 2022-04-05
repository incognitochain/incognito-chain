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
	txReqID common.Hash, shardID byte, accessOTA []byte,
) ([][]string, error) {
	res := [][]string{}
	rejectInst, err := instruction.NewRejectUnstakingWithValue(
		txReqID, shardID, metaData.StakingPoolID(), metaData.AccessID, accessOTA,
	).StringSlice()
	if err != nil {
		return res, err
	}
	res = append(res, rejectInst)
	if metaData.AccessOption.UseNft() {
		mintNftInst, err := instruction.NewMintNftWithValue(
			*metaData.AccessOption.NftID,
			metaData.OtaReceivers()[metaData.AccessOption.NftID.String()], // otaReceivers map has been check in shard metaData Tx verifier
			shardID, txReqID,
		).StringSlice(strconv.Itoa(metadataCommon.Pdexv3UnstakingRequestMeta))
		if err != nil {
			return res, err
		}
		res = append(res, mintNftInst)
	}
	return res, nil
}

func BuildAcceptUnstakingInstructions(
	stakingPoolID common.Hash, metaData metadataPdexv3.UnstakingRequest,
	txReqID common.Hash, shardID byte, accessOTA []byte,
) ([][]string, error) {
	res := [][]string{}
	acceptInst, err := instruction.NewAcceptUnstakingWithValue(
		stakingPoolID, metaData.UnstakingAmount(), metaData.OtaReceivers()[stakingPoolID.String()],
		txReqID, shardID, metaData.AccessOption, accessOTA,
	).StringSlice()
	if err != nil {
		return res, err
	}
	res = append(res, acceptInst)
	if metaData.AccessOption.UseNft() {
		mintNftInst, err := instruction.NewMintNftWithValue(
			*metaData.AccessOption.NftID,
			metaData.OtaReceivers()[metaData.NftID.String()],
			shardID, txReqID,
		).StringSlice(strconv.Itoa(metadataCommon.Pdexv3UnstakingRequestMeta))
		if err != nil {
			return res, err
		}
		res = append(res, mintNftInst)
	}
	return res, nil
}
