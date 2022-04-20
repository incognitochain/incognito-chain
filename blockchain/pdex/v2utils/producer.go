package v2utils

import (
	"encoding/json"
	"math/big"
	"sort"
	"strconv"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/config"
	metadataCommon "github.com/incognitochain/incognito-chain/metadata/common"
	metadataPdexv3 "github.com/incognitochain/incognito-chain/metadata/pdexv3"
)

func GetPDEXRewardsForBlock(
	beaconHeight uint64,
	mintingBlocks int, decayIntervals int, pdexRewardFirstInterval uint64,
	decayRateBPS int, bps int,
) uint64 {
	if beaconHeight <= config.Param().PDexParams.Pdexv3BreakPointHeight {
		return 0
	}
	// mint PDEX reward at the end of each epoch
	if beaconHeight%config.Param().EpochParam.NumberOfBlockInEpoch != 0 {
		return 0
	}

	epochSize := config.Param().EpochParam.NumberOfBlockInEpoch
	mintingEpochs := mintingBlocks / int(epochSize)

	pdexBlockRewards := uint64(0)
	intervalLength := uint64(mintingEpochs / decayIntervals)
	decayIntevalIdx := (beaconHeight - config.Param().PDexParams.Pdexv3BreakPointHeight) / intervalLength / epochSize
	if decayIntevalIdx < uint64(decayIntervals) {
		curIntervalReward := pdexRewardFirstInterval
		for i := uint64(0); i < decayIntevalIdx; i++ {
			decayAmount := new(big.Int).Mul(new(big.Int).SetUint64(curIntervalReward), big.NewInt(int64(decayRateBPS)))
			decayAmount = new(big.Int).Div(decayAmount, big.NewInt(int64(bps)))
			curIntervalReward -= decayAmount.Uint64()
		}
		pdexBlockRewards = curIntervalReward / intervalLength
	}
	return pdexBlockRewards
}

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
	accessOption metadataPdexv3.AccessOption,
	receivers map[common.Hash]metadataPdexv3.ReceiverInfo,
	shardID byte,
	reqTxID common.Hash,
	status string, accessOTA []byte,
) [][]string {
	if status == metadataPdexv3.RequestRejectedChainStatus {
		reqContent := metadataPdexv3.WithdrawalLPFeeContent{
			PoolPairID:   pairID,
			AccessOption: accessOption,
			TokenID:      common.Hash{},
			Receiver:     metadataPdexv3.ReceiverInfo{},
			TxReqID:      reqTxID,
			ShardID:      shardID,
			AccessOTA:    accessOTA,
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
			PoolPairID:   pairID,
			AccessOption: accessOption,
			TokenID:      tokenID,
			Receiver:     receivers[tokenID],
			IsLastInst:   isLastInstOfReqTx,
			TxReqID:      reqTxID,
			ShardID:      shardID,
			AccessOTA:    accessOTA,
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
	accessOption metadataPdexv3.AccessOption,
	receivers map[common.Hash]metadataPdexv3.ReceiverInfo,
	shardID byte,
	reqTxID common.Hash,
	status string,
	nextAccessOTA []byte,
) [][]string {
	if status == metadataPdexv3.RequestRejectedChainStatus {
		reqContent := metadataPdexv3.WithdrawalStakingRewardContent{
			StakingPoolID: stakingPoolID,
			AccessOption:  accessOption,
			TokenID:       common.Hash{},
			Receiver:      metadataPdexv3.ReceiverInfo{},
			TxReqID:       reqTxID,
			ShardID:       shardID,
			AccessOTA:     nextAccessOTA,
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
			AccessOption:  accessOption,
			TokenID:       tokenID,
			Receiver:      receivers[tokenID],
			IsLastInst:    isLastInstOfReqTx,
			TxReqID:       reqTxID,
			ShardID:       shardID,
			AccessOTA:     nextAccessOTA,
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

func BuildDistributeMiningOrderRewardInsts(
	pairID string,
	makingTokenID common.Hash,
	mintingAmount uint64,
	mintingTokenID common.Hash,
) [][]string {
	reqContent := metadataPdexv3.DistributeMiningOrderRewardContent{
		PoolPairID:    pairID,
		MakingTokenID: makingTokenID,
		Amount:        mintingAmount,
		TokenID:       mintingTokenID,
	}
	reqContentBytes, _ := json.Marshal(reqContent)

	return [][]string{
		{
			strconv.Itoa(metadataCommon.Pdexv3DistributeMiningOrderRewardMeta),
			strconv.Itoa(-1),
			metadataPdexv3.RequestAcceptedChainStatus,
			string(reqContentBytes),
		},
	}
}
