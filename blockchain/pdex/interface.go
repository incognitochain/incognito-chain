package pdex

import (
	"github.com/incognitochain/incognito-chain/dataaccessobject/rawdbv2"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
)

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
	WaitingContributionsV1() map[string]*rawdbv2.PDEContribution
	DeletedWaitingContributionsV1() map[string]*rawdbv2.PDEContribution
	PoolPairsV1() map[string]*rawdbv2.PDEPoolForPair
	WaitingContributionsV2() map[string]statedb.Pdexv3ContributionState
	DeletedWaitingContributionsV2() map[string]statedb.Pdexv3ContributionState
	PoolPairsV2() map[string]PoolPairState
	Shares() map[string]uint64
	TradingFees() map[string]uint64
}
