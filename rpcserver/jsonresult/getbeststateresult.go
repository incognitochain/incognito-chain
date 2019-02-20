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
