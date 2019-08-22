package jsonresult

import (
	"github.com/incognitochain/incognito-chain/blockchain"
	"github.com/incognitochain/incognito-chain/incognitokey"
)

/*
	Candidate Result From Best State
*/
type CandidateListsResult struct {
	Epoch                                  uint64                         `json:"Epoch"`
	CandidateShardWaitingForCurrentRandom  []incognitokey.CommitteePubKey `json:"CandidateShardWaitingForCurrentRandom"`
	CandidateBeaconWaitingForCurrentRandom []incognitokey.CommitteePubKey `json:"CandidateBeaconWaitingForCurrentRandom"`
	CandidateShardWaitingForNextRandom     []incognitokey.CommitteePubKey `json:"CandidateShardWaitingForNextRandom"`
	CandidateBeaconWaitingForNextRandom    []incognitokey.CommitteePubKey `json:"CandidateBeaconWaitingForNextRandom"`
}

type CommitteeListsResult struct {
	Epoch                  uint64                                  `json:"Epoch"`
	ShardCommittee         map[byte][]incognitokey.CommitteePubKey `json:"ShardCommittee"`
	ShardPendingValidator  map[byte][]incognitokey.CommitteePubKey `json:"ShardPendingValidator"`
	BeaconCommittee        []incognitokey.CommitteePubKey          `json:"BeaconCommittee"`
	BeaconPendingValidator []incognitokey.CommitteePubKey          `json:"BeaconPendingValidator"`
}

func NewCommitteeListsResult(epoch uint64, shardComm map[byte][]incognitokey.CommitteePubKey, shardPendingValidator map[byte][]incognitokey.CommitteePubKey, beaconCommittee []incognitokey.CommitteePubKey, beaconPendingValidator []incognitokey.CommitteePubKey) *CommitteeListsResult {
	result := &CommitteeListsResult{
		Epoch: epoch,
	}
	result.BeaconPendingValidator = make([]incognitokey.CommitteePubKey, len(beaconPendingValidator))
	copy(result.BeaconPendingValidator, beaconPendingValidator)
	result.BeaconCommittee = make([]incognitokey.CommitteePubKey, len(beaconCommittee))
	copy(result.BeaconCommittee, beaconCommittee)
	result.ShardCommittee = make(map[byte][]incognitokey.CommitteePubKey)
	for k, v := range shardComm {
		result.ShardCommittee[k] = make([]incognitokey.CommitteePubKey, len(v))
		copy(result.ShardCommittee[k], v)
	}
	result.ShardPendingValidator = make(map[byte][]incognitokey.CommitteePubKey)
	for k, v := range shardPendingValidator {
		result.ShardPendingValidator[k] = make([]incognitokey.CommitteePubKey, len(v))
		copy(result.ShardPendingValidator[k], v)
	}
	return result
}

type StakeResult struct {
	PublicKey string `json:"PublicKey"`
	CanStake  bool   `json:"CanStake"`
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
