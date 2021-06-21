package pdex

import "github.com/incognitochain/incognito-chain/dataaccessobject/rawdbv2"

type stateBase struct {
	waitingContributions        map[string]*rawdbv2.PDEContribution
	deletedWaitingContributions map[string]*rawdbv2.PDEContribution
	poolPairs                   map[string]*rawdbv2.PDEPoolForPair
	shares                      map[string]uint64
	tradingFees                 map[string]uint64
}

//Version of state
func (s *stateBase) Version() uint {
	panic("Implement this fucntion")
}

func (s *stateBase) Clone() State {
	return nil
}

func (s *stateBase) Process(env StateEnvironment) error {
	return nil
}

func (s *stateBase) StoreToDB(env StateEnvironment) error {
	return nil
}

func (s *stateBase) BuildInstructions(env StateEnvironment) ([][]string, error) {
	return nil, nil
}

func (s *stateBase) Upgrade(StateEnvironment) State {
	panic("Implement this fucntion")
}
