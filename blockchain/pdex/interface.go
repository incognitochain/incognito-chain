package pdex

type State interface {
	Version() uint
	Clone() State
	Update(StateEnvironment) error
	StoreToDB(StateEnvironment) error
	BuildInstructions(StateEnvironment) ([][]string, error)
	Upgrade(StateEnvironment) State
}

type stateProducer interface {
}

type stateProcessor interface {
}
