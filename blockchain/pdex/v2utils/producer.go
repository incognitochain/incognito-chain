package v2utils

import (
	"encoding/json"
	"strconv"

	"github.com/incognitochain/incognito-chain/common"
	metadataCommon "github.com/incognitochain/incognito-chain/metadata/common"
	metadataPdexv3 "github.com/incognitochain/incognito-chain/metadata/pdexv3"
)

func BuildModifyParamsInst(
	params metadataPdexv3.Pdexv3Params,
	errorMsg string,
	shardID byte,
	reqTxID common.Hash,
	status string,
) []string {
	modifyingParamsReqContent := metadataPdexv3.ParamsModifyingContent{
		Content:  params,
		ErrorMsg: errorMsg,
		TxReqID:  reqTxID,
		ShardID:  shardID,
	}
	modifyingParamsReqContentBytes, _ := json.Marshal(modifyingParamsReqContent)
	return []string{
		strconv.Itoa(metadataCommon.Pdexv3ModifyParamsMeta),
		strconv.Itoa(int(shardID)),
		status,
		string(modifyingParamsReqContentBytes),
	}
}

func BuildMintPDEXInst(
	pairID string,
	mintingAmount uint,
) [][]string {
	reqContent := metadataPdexv3.MintPDEXBlockRewardContent{
		PoolPairID: pairID,
		Amount:     mintingAmount,
	}
	reqContentBytes, _ := json.Marshal(reqContent)

	return [][]string{
		{
			strconv.Itoa(metadataCommon.Pdexv3MintPDEXBlockRewardMeta),
			strconv.Itoa(-1),
			metadataPdexv3.RequestAcceptedChainStatus,
			string(reqContentBytes),
		},
	}
}

func BuildWithdrawLPFeeInsts(
	pairID string,
	nftID common.Hash,
	receivers map[string]metadataPdexv3.ReceiverInfo,
	shardID byte,
	reqTxID common.Hash,
	status string,
) [][]string {
	insts := [][]string{}

	for tokenType := range receivers {
		reqContent := metadataPdexv3.WithdrawalLPFeeContent{
			PoolPairID: pairID,
			NftID:      nftID,
			TokenType:  tokenType,
			Receivers:  receivers,
			TxReqID:    reqTxID,
			ShardID:    shardID,
		}
		reqContentBytes, _ := json.Marshal(reqContent)
		insts = append(insts, []string{
			strconv.Itoa(metadataCommon.Pdexv3WithdrawLPFeeRequestMeta),
			strconv.Itoa(int(shardID)),
			status,
			string(reqContentBytes),
		})
	}

	return insts
}

func BuildWithdrawProtocolFeeInsts(
	pairID string,
	receivers map[string]metadataPdexv3.ReceiverInfo,
	shardID byte,
	reqTxID common.Hash,
	status string,
) [][]string {
	if status == metadataPdexv3.RequestRejectedChainStatus {
		reqContent := metadataPdexv3.WithdrawalProtocolFeeContent{
			PoolPairID: pairID,
			TokenType:  "",
			Receivers:  map[string]metadataPdexv3.ReceiverInfo{},
			TxReqID:    reqTxID,
			ShardID:    shardID,
		}
		reqContentBytes, _ := json.Marshal(reqContent)
		inst := []string{
			strconv.Itoa(metadataCommon.Pdexv3WithdrawProtocolFeeRequestMeta),
			strconv.Itoa(int(shardID)),
			status,
			string(reqContentBytes),
		}
		return [][]string{inst}
	}

	insts := [][]string{}
	for tokenType := range receivers {
		reqContent := metadataPdexv3.WithdrawalProtocolFeeContent{
			PoolPairID: pairID,
			TokenType:  tokenType,
			Receivers:  receivers,
			TxReqID:    reqTxID,
			ShardID:    shardID,
		}
		reqContentBytes, _ := json.Marshal(reqContent)
		insts = append(insts, []string{
			strconv.Itoa(metadataCommon.Pdexv3WithdrawProtocolFeeRequestMeta),
			strconv.Itoa(int(shardID)),
			status,
			string(reqContentBytes),
		})
	}

	return insts
}
