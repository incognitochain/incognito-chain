package manager

type BeaconCommitteeEngine interface {
	AssignCommitteeUsingRandomInstruction(rand int64) ([]string, map[byte][]string)
}
