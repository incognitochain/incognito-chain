package pdex

type State interface {
	Version() uint
	Clone() State
	Process(StateEnvironment) error
	StoreToDB(StateEnvironment, *StateChange) error
	BuildInstructions(StateEnvironment) ([][]string, error)
	Upgrade(StateEnvironment) State
	TransformKeyWithNewBeaconHeight(beaconHeight uint64)
	ClearCache()
	GetDiff(State, *StateChange) (State, *StateChange, error)
	Reader() StateReader
}

type StateReader interface {
	Params() Params
	WaitingContributions() []byte
	PoolPairs() []byte
	Shares() map[string]uint64
	TradingFees() map[string]uint64
	NftIDs() map[string]uint64
	StakingPools() map[string]*StakingPoolState
}
