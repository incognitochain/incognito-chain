package jsonresult

import "github.com/incognitochain/incognito-chain/blockchain"

/*
	Candidate Result From Best State
*/
type CandidateListsResult struct {
	Epoch                                  uint64   `json:"Epoch"`
	CandidateShardWaitingForCurrentRandom  []string `json:"CandidateShardWaitingForCurrentRandom"`
	CandidateBeaconWaitingForCurrentRandom []string `json:"CandidateBeaconWaitingForCurrentRandom"`
	CandidateShardWaitingForNextRandom     []string `json:"CandidateShardWaitingForNextRandom"`
	CandidateBeaconWaitingForNextRandom    []string `json:"CandidateBeaconWaitingForNextRandom"`
}

type CommitteeListsResult struct {
	Epoch                  uint64            `json:"Epoch"`
	ShardCommittee         map[byte][]string `json:"ShardCommittee"`
	ShardPendingValidator  map[byte][]string `json:"ShardPendingValidator"`
	BeaconCommittee        []string          `json:"BeaconCommittee"`
	BeaconPendingValidator []string          `json:"BeaconPendingValidator"`
}

func NewCommitteeListsResult(epoch uint64, shardComm map[byte][]string, shardPendingValidator map[byte][]string, beaconCommittee []string, beaconPendingValidator []string) *CommitteeListsResult {
	result := &CommitteeListsResult{
		Epoch: epoch,
	}
	result.BeaconPendingValidator = make([]string, len(beaconPendingValidator))
	copy(result.BeaconPendingValidator, beaconPendingValidator)
	result.BeaconCommittee = make([]string, len(beaconCommittee))
	copy(result.BeaconCommittee, beaconCommittee)
	result.ShardCommittee = make(map[byte][]string)
	for k, v := range shardComm {
		result.ShardCommittee[k] = make([]string, len(v))
		copy(result.ShardCommittee[k], v)
	}
	result.ShardPendingValidator = make(map[byte][]string)
	for k, v := range shardPendingValidator {
		result.ShardPendingValidator[k] = make([]string, len(v))
		copy(result.ShardPendingValidator[k], v)
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
