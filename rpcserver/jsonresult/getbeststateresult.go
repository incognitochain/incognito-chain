package jsonresult

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

type StakeResult struct {
	PublicKey string `json:"PublicKey"`
	CanStake  bool   `json:"CanStake"`
}
