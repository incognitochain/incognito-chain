package pdex

type State interface {
	Version() uint
	Clone() State
	Process(StateEnvironment) error
	StoreToDB(StateEnvironment) error
	BuildInstructions(StateEnvironment) ([][]string, error)
	Upgrade(StateEnvironment) State
	TransformKeyWithNewBeaconHeight(beaconHeight uint64) State
}
