package committeestate

type RandomRuleV1 struct {
}

func (randomRuleV1 *RandomRuleV1) Exec(rand int64, activeShards int, committeeChange *CommitteeChange, oldState BeaconCommitteeState) *CommitteeChange {
	// b == newstate -> only write
	// oldstate -> only read
	//newCommitteeChange := committeeChange
	//candidateStructs := oldState.ShardCommonPool()[:b.numberOfAssignedCandidates]
	//candidates, _ := incognitokey.CommitteeKeyListToString(candidateStructs)
	//newCommitteeChange = b.assign(candidates, rand, activeShards, newCommitteeChange, oldState)
	//newCommitteeChange.NextEpochShardCandidateRemoved = append(newCommitteeChange.NextEpochShardCandidateRemoved, candidateStructs...)
	//return newCommitteeChange,
	return nil
}
