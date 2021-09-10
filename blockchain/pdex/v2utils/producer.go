package v2utils

import (
	"encoding/json"
	"sort"
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

func BuildMintBlockRewardInst(
	pairID string,
	mintingAmount uint64,
	mintingTokenID common.Hash,
) [][]string {
	reqContent := metadataPdexv3.MintBlockRewardContent{
		PoolPairID: pairID,
		Amount:     mintingAmount,
		TokenID:    mintingTokenID,
	}
	reqContentBytes, _ := json.Marshal(reqContent)

	return [][]string{
		{
			strconv.Itoa(metadataCommon.Pdexv3MintBlockRewardMeta),
			strconv.Itoa(-1),
			metadataPdexv3.RequestAcceptedChainStatus,
			string(reqContentBytes),
		},
	}
}

func BuildWithdrawLPFeeInsts(
	pairID string,
	nftID common.Hash,
	receivers map[common.Hash]metadataPdexv3.ReceiverInfo,
	shardID byte,
	reqTxID common.Hash,
	status string,
) [][]string {
	if status == metadataPdexv3.RequestRejectedChainStatus {
		reqContent := metadataPdexv3.WithdrawalLPFeeContent{
			PoolPairID: pairID,
			NftID:      nftID,
			TokenID:    common.Hash{},
			Receiver:   metadataPdexv3.ReceiverInfo{},
			TxReqID:    reqTxID,
			ShardID:    shardID,
		}
		reqContentBytes, _ := json.Marshal(reqContent)
		inst := []string{
			strconv.Itoa(metadataCommon.Pdexv3WithdrawLPFeeRequestMeta),
			strconv.Itoa(int(shardID)),
			status,
			string(reqContentBytes),
		}
		return [][]string{inst}
	}

	// To store the keys in slice in sorted order
	keys := make([]common.Hash, len(receivers))
	i := 0
	for k := range receivers {
		keys[i] = k
		i++
	}
	sort.SliceStable(keys, func(i, j int) bool {
		return keys[i].String() < keys[j].String()
	})

	insts := [][]string{}
	for i, tokenID := range keys {
		isLastInstOfReqTx := i == len(keys)-1
		reqContent := metadataPdexv3.WithdrawalLPFeeContent{
			PoolPairID: pairID,
			NftID:      nftID,
			TokenID:    tokenID,
			Receiver:   receivers[tokenID],
			IsLastInst: isLastInstOfReqTx,
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
	address string,
	amounts map[common.Hash]uint64,
	shardID byte,
	reqTxID common.Hash,
	status string,
) [][]string {
	if status == metadataPdexv3.RequestRejectedChainStatus {
		reqContent := metadataPdexv3.WithdrawalProtocolFeeContent{
			PoolPairID: pairID,
			Address:    address,
			TokenID:    common.Hash{},
			Amount:     0,
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

	// To store the keys in slice in sorted order
	keys := make([]common.Hash, len(amounts))
	i := 0
	for k := range amounts {
		keys[i] = k
		i++
	}
	sort.SliceStable(keys, func(i, j int) bool {
		return keys[i].String() < keys[j].String()
	})

	insts := [][]string{}
	for i, tokenID := range keys {
		isLastInstOfReqTx := i == len(keys)-1
		reqContent := metadataPdexv3.WithdrawalProtocolFeeContent{
			PoolPairID: pairID,
			Address:    address,
			TokenID:    tokenID,
			Amount:     amounts[tokenID],
			IsLastInst: isLastInstOfReqTx,
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

func BuildDistributeStakingRewardInst(
	stakingToken string,
	rewards map[common.Hash]uint64,
) [][]string {
	reqContent := metadataPdexv3.DistributeStakingRewardContent{
		StakingPoolID: stakingToken,
		Rewards:       rewards,
	}
	reqContentBytes, _ := json.Marshal(reqContent)

	return [][]string{
		{
			strconv.Itoa(metadataCommon.Pdexv3DistributeStakingRewardMeta),
			strconv.Itoa(-1),
			metadataPdexv3.RequestAcceptedChainStatus,
			string(reqContentBytes),
		},
	}
}

func BuildWithdrawStakingRewardInsts(
	stakingPoolID string,
	nftID common.Hash,
	receivers map[common.Hash]metadataPdexv3.ReceiverInfo,
	shardID byte,
	reqTxID common.Hash,
	status string,
) [][]string {
	if status == metadataPdexv3.RequestRejectedChainStatus {
		reqContent := metadataPdexv3.WithdrawalStakingRewardContent{
			StakingPoolID: stakingPoolID,
			NftID:         nftID,
			TokenID:       common.Hash{},
			Receiver:      metadataPdexv3.ReceiverInfo{},
			TxReqID:       reqTxID,
			ShardID:       shardID,
		}
		reqContentBytes, _ := json.Marshal(reqContent)
		inst := []string{
			strconv.Itoa(metadataCommon.Pdexv3WithdrawStakingRewardRequestMeta),
			strconv.Itoa(int(shardID)),
			status,
			string(reqContentBytes),
		}
		return [][]string{inst}
	}

	// To store the keys in slice in sorted order
	keys := make([]common.Hash, len(receivers))
	i := 0
	for k := range receivers {
		keys[i] = k
		i++
	}
	sort.SliceStable(keys, func(i, j int) bool {
		return keys[i].String() < keys[j].String()
	})

	insts := [][]string{}
	for i, tokenID := range keys {
		isLastInstOfReqTx := i == len(keys)-1
		reqContent := metadataPdexv3.WithdrawalStakingRewardContent{
			StakingPoolID: stakingPoolID,
			NftID:         nftID,
			TokenID:       tokenID,
			Receiver:      receivers[tokenID],
			IsLastInst:    isLastInstOfReqTx,
			TxReqID:       reqTxID,
			ShardID:       shardID,
		}
		reqContentBytes, _ := json.Marshal(reqContent)
		insts = append(insts, []string{
			strconv.Itoa(metadataCommon.Pdexv3WithdrawStakingRewardRequestMeta),
			strconv.Itoa(int(shardID)),
			status,
			string(reqContentBytes),
		})
	}

	return insts
}
