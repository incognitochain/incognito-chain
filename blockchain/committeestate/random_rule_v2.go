package committeestate

type RandomRuleV2 struct{}

func (randomRuleV2 *RandomRuleV2) Exec(rand int64, activeShards int, committeeChange *CommitteeChange, oldState BeaconCommitteeState) *CommitteeChange {
	return nil
}
