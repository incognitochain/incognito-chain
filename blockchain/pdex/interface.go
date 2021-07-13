package pdex

type State interface {
	Version() uint
	Clone() State
	Process(StateEnvironment) error
	StoreToDB(StateEnvironment) error
	BuildInstructions(StateEnvironment) ([][]string, error)
	Upgrade(StateEnvironment) State
	TransformKeyWithNewBeaconHeight(beaconHeight uint64)
	ClearCache()
	GetDiff(State) (State, error)
	Reader() StateReader
}

type StateReader interface {
	WaitingContributions() map[string]interface{}
	DeletedWaitingContributions() map[string]interface{}
	PoolPairs() map[string]interface{}
	Shares() map[string]uint64
	TradingFees() map[string]uint64
}
