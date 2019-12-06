package jsonresult

import (
	"github.com/incognitochain/incognito-chain/blockchain"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/common/base58"
	"github.com/incognitochain/incognito-chain/incognitokey"
)

/*
	Candidate Result From Best State
*/
type CandidateListsResult struct {
	Epoch                                  uint64                            `json:"Epoch"`
	CandidateShardWaitingForCurrentRandom  []incognitokey.CommitteePublicKey `json:"CandidateShardWaitingForCurrentRandom"`
	CandidateBeaconWaitingForCurrentRandom []incognitokey.CommitteePublicKey `json:"CandidateBeaconWaitingForCurrentRandom"`
	CandidateShardWaitingForNextRandom     []incognitokey.CommitteePublicKey `json:"CandidateShardWaitingForNextRandom"`
	CandidateBeaconWaitingForNextRandom    []incognitokey.CommitteePublicKey `json:"CandidateBeaconWaitingForNextRandom"`
}

type CommitteeListsResult struct {
	Epoch                       uint64            `json:"Epoch"`
	ShardCommittee              map[byte][]string `json:"ShardCommittee"`
	ShardRewardReceiver         map[byte][]string `json:"ShardRewardReceiver"`
	ShardPendingValidator       map[byte][]string `json:"ShardPendingValidator"`
	ShardPendingRewardReceiver  map[byte][]string `json:"ShardPendingRewardReceiver"`
	BeaconCommittee             []string          `json:"BeaconCommittee"`
	BeaconRewardReceiver        []string          `json:"BeaconRewardReceiver"`
	BeaconPendingValidator      []string          `json:"BeaconPendingValidator"`
	BeaconPendingRewardReceiver []string          `json:"BeaconPendingRewardReceiver"`
}

func NewCommitteeListsResult(epoch uint64, shardComm map[byte][]incognitokey.CommitteePublicKey, shardPendingValidator map[byte][]incognitokey.CommitteePublicKey, beaconCommittee []incognitokey.CommitteePublicKey, beaconPendingValidator []incognitokey.CommitteePublicKey) *CommitteeListsResult {
	result := &CommitteeListsResult{
		Epoch: epoch,
	}
	// ===== BEACON =====
	// pending validator
	result.BeaconPendingValidator = make([]string, 0)
	result.BeaconPendingRewardReceiver = make([]string, 0)
	for _, v := range beaconPendingValidator {
		result.BeaconPendingValidator = append(result.BeaconPendingValidator, base58.Base58Check{}.Encode(v.MiningPubKey[common.BlsConsensus], common.ZeroByte))
		result.BeaconPendingRewardReceiver = append(result.BeaconPendingRewardReceiver, base58.Base58Check{}.Encode(v.IncPubKey, common.ZeroByte))
	}
	// current committee
	result.BeaconCommittee = make([]string, 0)
	result.BeaconRewardReceiver = make([]string, 0)
	for _, v := range beaconCommittee {
		result.BeaconCommittee = append(result.BeaconCommittee, base58.Base58Check{}.Encode(v.MiningPubKey[common.BlsConsensus], common.ZeroByte))
		result.BeaconRewardReceiver = append(result.BeaconRewardReceiver, base58.Base58Check{}.Encode(v.IncPubKey, common.ZeroByte))
	}

	// ===== SHARD =====
	// pending validator
	result.ShardPendingValidator = make(map[byte][]string)
	result.ShardPendingRewardReceiver = make(map[byte][]string)
	for k, v := range shardPendingValidator {
		result.ShardPendingValidator[k] = make([]string, 0)
		for _, v1 := range v {
			result.ShardPendingValidator[k] = append(result.ShardPendingValidator[k], base58.Base58Check{}.Encode(v1.MiningPubKey[common.BlsConsensus], common.ZeroByte))
			result.ShardPendingRewardReceiver[k] = append(result.ShardPendingRewardReceiver[k], base58.Base58Check{}.Encode(v1.IncPubKey, common.ZeroByte))
		}
	}
	// current committee
	result.ShardCommittee = make(map[byte][]string)
	result.ShardRewardReceiver = make(map[byte][]string)
	for k, v := range shardComm {
		result.ShardCommittee[k] = make([]string, 0)
		for _, v1 := range v {
			result.ShardCommittee[k] = append(result.ShardCommittee[k], base58.Base58Check{}.Encode(v1.MiningPubKey[common.BlsConsensus], common.ZeroByte))
			result.ShardRewardReceiver[k] = append(result.ShardRewardReceiver[k], base58.Base58Check{}.Encode(v1.IncPubKey, common.ZeroByte))
		}
	}
	return result
}

type StakeResult struct {
	PublicKey string `json:"PublicKey"`
	CanStake  bool   `json:"CanStake"`
}

func NewStakeResult(publicKey string, canStake bool) *StakeResult {
	result := &StakeResult{
		PublicKey: publicKey,
		CanStake:  canStake,
	}
	return result
}

type TotalTransactionInShard struct {
	TotalTransactions                 uint64 `json:"TotalTransactions"`
	TotalTransactionsExcludeSystemTxs uint64 `json:"TotalTransactionsExcludeSystemTxs"`
	SalaryTransaction                 uint64 `json:"SalaryTransaction"`
}

func NewTotalTransactionInShard(shardBeststate *blockchain.ShardBestState) *TotalTransactionInShard {
	result := &TotalTransactionInShard{
		TotalTransactions:                 shardBeststate.TotalTxns,
		TotalTransactionsExcludeSystemTxs: shardBeststate.TotalTxnsExcludeSalary,
		SalaryTransaction:                 shardBeststate.TotalTxns - shardBeststate.TotalTxnsExcludeSalary,
	}
	return result
}
