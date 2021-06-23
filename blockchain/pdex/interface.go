package pdex

import "github.com/incognitochain/incognito-chain/dataaccessobject/rawdbv2"

type State interface {
	Version() uint
	Clone() State
	Process(StateEnvironment) error
	StoreToDB(StateEnvironment) error
	BuildInstructions(StateEnvironment) ([][]string, error)
	Upgrade(StateEnvironment) State
	TransformKeyWithNewBeaconHeight(beaconHeight uint64) State
	ClearCache()
	GetDiff(State) (State, error)
	WaitingContributions() map[string]*rawdbv2.PDEContribution
	DeletedWaitingContributions() map[string]*rawdbv2.PDEContribution
	PoolPairs() map[string]*rawdbv2.PDEPoolForPair
	Shares() map[string]uint64
	TradingFees() map[string]uint64
}
