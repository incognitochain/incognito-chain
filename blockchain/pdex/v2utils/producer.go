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
	shardID byte,
	reqTxID common.Hash,
	status string,
) []string {
	modifyingParamsReqContent := metadataPdexv3.ParamsModifyingContent{
		Content: params,
		TxReqID: reqTxID,
		ShardID: shardID,
	}
	modifyingParamsReqContentBytes, _ := json.Marshal(modifyingParamsReqContent)
	return []string{
		strconv.Itoa(metadataCommon.Pdexv3ModifyParamsMeta),
		strconv.Itoa(int(shardID)),
		status,
		string(modifyingParamsReqContentBytes),
	}
}

func BuildWithdrawLPFeeInsts(
	pairID string,
	ncftTokenID common.Hash,
	receivers map[string]metadataPdexv3.ReceiverInfo,
	shardID byte,
	reqTxID common.Hash,
	status string,
) [][]string {
	insts := [][]string{}

	for tokenType, receiver := range receivers {
		reqContent := metadataPdexv3.WithdrawalLPFeeContent{
			PairID:      pairID,
			NcftTokenID: ncftTokenID,
			TokenType:   tokenType,
			Receiver:    receiver,
			TxReqID:     reqTxID,
			ShardID:     shardID,
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
			PairID:    pairID,
			TokenType: "",
			Receiver:  metadataPdexv3.ReceiverInfo{},
			TxReqID:   reqTxID,
			ShardID:   shardID,
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
	for tokenType, receiver := range receivers {
		reqContent := metadataPdexv3.WithdrawalProtocolFeeContent{
			PairID:    pairID,
			TokenType: tokenType,
			Receiver:  receiver,
			TxReqID:   reqTxID,
			ShardID:   shardID,
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
