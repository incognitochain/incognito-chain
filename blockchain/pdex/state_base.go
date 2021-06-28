package pdex

import (
	"github.com/incognitochain/incognito-chain/dataaccessobject/rawdbv2"
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

func (s *stateBase) StoreToDB(env StateEnvironment) error {
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

func (s *stateBase) GetDiff(compareState State) (State, error) {
	panic("Implement this fucntion")
}

func (s *stateBase) WaitingContributions() map[string]*rawdbv2.PDEContribution {
	panic("Implement this fucntion")
}

func (s *stateBase) DeletedWaitingContributions() map[string]*rawdbv2.PDEContribution {
	panic("Implement this fucntion")
}

func (s *stateBase) PoolPairs() map[string]*rawdbv2.PDEPoolForPair {
	panic("Implement this fucntion")
}

func (s *stateBase) Shares() map[string]uint64 {
	panic("Implement this fucntion")
}

func (s *stateBase) TradingFees() map[string]uint64 {
	panic("Implement this fucntion")
}
