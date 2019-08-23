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
	CandidateShardWaitingForCurrentRandom  []incognitokey.CommitteePublicKey `json:"CandidateShardWaitingForCurrentRandom"`
	CandidateBeaconWaitingForCurrentRandom []incognitokey.CommitteePublicKey `json:"CandidateBeaconWaitingForCurrentRandom"`
	CandidateShardWaitingForNextRandom     []incognitokey.CommitteePublicKey `json:"CandidateShardWaitingForNextRandom"`
	CandidateBeaconWaitingForNextRandom    []incognitokey.CommitteePublicKey `json:"CandidateBeaconWaitingForNextRandom"`
}

type CommitteeListsResult struct {
	Epoch                  uint64                                  `json:"Epoch"`
	ShardCommittee         map[byte][]incognitokey.CommitteePublicKey `json:"ShardCommittee"`
	ShardPendingValidator  map[byte][]incognitokey.CommitteePublicKey `json:"ShardPendingValidator"`
	BeaconCommittee        []incognitokey.CommitteePublicKey          `json:"BeaconCommittee"`
	BeaconPendingValidator []incognitokey.CommitteePublicKey          `json:"BeaconPendingValidator"`
}

func NewCommitteeListsResult(epoch uint64, shardComm map[byte][]incognitokey.CommitteePublicKey, shardPendingValidator map[byte][]incognitokey.CommitteePublicKey, beaconCommittee []incognitokey.CommitteePublicKey, beaconPendingValidator []incognitokey.CommitteePublicKey) *CommitteeListsResult {
	result := &CommitteeListsResult{
		Epoch: epoch,
	}
	result.BeaconPendingValidator = make([]incognitokey.CommitteePublicKey, len(beaconPendingValidator))
	copy(result.BeaconPendingValidator, beaconPendingValidator)
	result.BeaconCommittee = make([]incognitokey.CommitteePublicKey, len(beaconCommittee))
	copy(result.BeaconCommittee, beaconCommittee)
	result.ShardCommittee = make(map[byte][]incognitokey.CommitteePublicKey)
	for k, v := range shardComm {
		result.ShardCommittee[k] = make([]incognitokey.CommitteePublicKey, len(v))
		copy(result.ShardCommittee[k], v)
	}
	result.ShardPendingValidator = make(map[byte][]incognitokey.CommitteePublicKey)
	for k, v := range shardPendingValidator {
		result.ShardPendingValidator[k] = make([]incognitokey.CommitteePublicKey, len(v))
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
