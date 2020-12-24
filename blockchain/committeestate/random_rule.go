package committeestate

type RandomRule interface {
	Exec(int64, int, *CommitteeChange, BeaconCommitteeState) (*CommitteeChange, int)
}
