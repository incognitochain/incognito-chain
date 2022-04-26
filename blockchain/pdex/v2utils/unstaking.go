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
	nftHash, _ := common.Hash{}.NewHashFromStr(metaData.NftID())
	mintNftInst, err := instruction.NewMintNftWithValue(
		*nftHash, metaData.OtaReceivers()[metaData.NftID()], shardID, txReqID,
	).StringSlice(strconv.Itoa(metadataCommon.Pdexv3UnstakingRequestMeta))
	if err != nil {
		return res, err
	}
	res = append(res, mintNftInst)
	return res, nil
}

func BuildAcceptUnstakingInstructions(
	stakingPoolID, nftID common.Hash,
	unstakingAmount uint64,
	otaReceiverNft, otaReceiverUnstakingToken string,
	txReqID common.Hash, shardID byte,
) ([][]string, error) {
	res := [][]string{}
	acceptInst, err := instruction.NewAcceptUnstakingWithValue(
		stakingPoolID, nftID, unstakingAmount, otaReceiverUnstakingToken, txReqID, shardID,
	).StringSlice()
	if err != nil {
		return res, err
	}
	res = append(res, acceptInst)
	mintNftInst, err := instruction.NewMintNftWithValue(nftID, otaReceiverNft, shardID, txReqID).
		StringSlice(strconv.Itoa(metadataCommon.Pdexv3UnstakingRequestMeta))
	if err != nil {
		return res, err
	}
	res = append(res, mintNftInst)
	return res, nil
}
