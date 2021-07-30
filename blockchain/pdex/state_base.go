package pdex

import (
	"github.com/incognitochain/incognito-chain/dataaccessobject/rawdbv2"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
)

type stateBase struct {
}

func newStateBase() *stateBase {
	return &stateBase{}
}

func newStateBaseWithValue() *stateBase {
	return &stateBase{}
}

//Version of state
func (s *stateBase) Version() uint {
	panic("Implement this fucntion")
}

func (s *stateBase) Clone() State {
	res := newStateBase()

	return res
}

func (s *stateBase) Process(env StateEnvironment) error {
	return nil
}

func (s *stateBase) StoreToDB(env StateEnvironment, stateChagne *StateChange) error {
	var err error
	return err
}

func (s *stateBase) BuildInstructions(env StateEnvironment) ([][]string, error) {
	panic("Implement this function")
}

func (s *stateBase) Upgrade(StateEnvironment) State {
	panic("Implement this fucntion")
}

func (s *stateBase) TransformKeyWithNewBeaconHeight(beaconHeight uint64) {
	panic("Implement this fucntion")
}

func (s *stateBase) ClearCache() {
	panic("Implement this fucntion")
}

func (s *stateBase) GetDiff(compareState State, stateChange *StateChange) (State, *StateChange, error) {
	panic("Implement this fucntion")
}

func (s *stateBase) WaitingContributionsV1() map[string]*rawdbv2.PDEContribution {
	panic("Implement this fucntion")
}

func (s *stateBase) DeletedWaitingContributionsV1() map[string]*rawdbv2.PDEContribution {
	panic("Implement this fucntion")
}

func (s *stateBase) Params() Params {
	panic("Implement this fucntion")
}

func (s *stateBase) PoolPairsV1() map[string]*rawdbv2.PDEPoolForPair {
	panic("Implement this fucntion")
}

func (s *stateBase) WaitingContributionsV2() map[string]statedb.Pdexv3ContributionState {
	panic("Implement this fucntion")
}

func (s *stateBase) DeletedWaitingContributionsV2() map[string]statedb.Pdexv3ContributionState {
	panic("Implement this fucntion")
}

func (s *stateBase) PoolPairsV2() map[string]PoolPairState {
	panic("Implement this fucntion")
}

func (s *stateBase) Shares() map[string]uint64 {
	panic("Implement this fucntion")
}

func (s *stateBase) TradingFees() map[string]uint64 {
	panic("Implement this fucntion")
}

func (s *stateBase) Reader() StateReader {
	panic("Implement this fucntion")
}
