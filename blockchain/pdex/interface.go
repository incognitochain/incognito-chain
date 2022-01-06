package pdex

type State interface {
	Version() uint
	Clone() State
	Process(StateEnvironment) error
	StoreToDB(StateEnvironment, *StateChange) error
	BuildInstructions(StateEnvironment) ([][]string, error)
	TransformKeyWithNewBeaconHeight(beaconHeight uint64)
	ClearCache()
	GetDiff(State, *StateChange) (State, *StateChange, error)
	Reader() StateReader
	Validator() StateValidator
}

type StateReader interface {
	Params() *Params
	WaitingContributions() []byte
	PoolPairs() []byte
	Shares() map[string]uint64
	TradingFees() map[string]uint64
	NftIDs() map[string]uint64
	StakingPools() map[string]*StakingPoolState
}

type StateValidator interface {
	IsValidNftID(nftID string) error
	IsValidPoolPairID(poolPairID string) error
	IsValidMintNftRequireAmount(amount uint64) error
	IsValidStakingPool(tokenID string) error
	IsValidUnstakingAmount(tokenID, nftID string, unstakingAmount uint64) error
	IsValidShareAmount(poolPairID, nftID string, shareAmount uint64) error
}
