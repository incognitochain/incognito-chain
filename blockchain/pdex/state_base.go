package pdex

import (
	"errors"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/incognitochain/incognito-chain/dataaccessobject/rawdbv2"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
)

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
	res := &stateBase{}
	res.waitingContributions = make(map[string]*rawdbv2.PDEContribution, len(s.waitingContributions))
	res.deletedWaitingContributions = make(map[string]*rawdbv2.PDEContribution, len(s.deletedWaitingContributions))
	res.poolPairs = make(map[string]*rawdbv2.PDEPoolForPair, len(s.poolPairs))
	res.shares = make(map[string]uint64, len(s.shares))
	res.tradingFees = make(map[string]uint64, len(s.tradingFees))

	for k, v := range s.waitingContributions {
		*res.waitingContributions[k] = *v
	}

	for k, v := range s.deletedWaitingContributions {
		*res.deletedWaitingContributions[k] = *v
	}

	for k, v := range s.poolPairs {
		*res.poolPairs[k] = *v
	}

	for k, v := range s.shares {
		res.shares[k] = v
	}

	for k, v := range s.tradingFees {
		res.tradingFees[k] = v
	}

	return res
}

func (s *stateBase) Process(env StateEnvironment) error {
	return nil
}

func (s *stateBase) StoreToDB(env StateEnvironment) error {
	var err error
	statedb.DeleteWaitingPDEContributions(
		env.StateDB(),
		s.deletedWaitingContributions,
	)
	err = statedb.StoreWaitingPDEContributions(
		env.StateDB(),
		s.waitingContributions,
	)
	if err != nil {
		return err
	}
	err = statedb.StorePDEPoolPairs(
		env.StateDB(),
		s.poolPairs,
	)
	if err != nil {
		return err
	}
	err = statedb.StorePDEShares(
		env.StateDB(),
		s.shares,
	)
	if err != nil {
		return err
	}
	err = statedb.StorePDETradingFees(
		env.StateDB(),
		s.tradingFees,
	)
	if err != nil {
		return err
	}
	return err
}

func (s *stateBase) BuildInstructions(env StateEnvironment) ([][]string, error) {
	panic("Implement this function")
}

func (s *stateBase) Upgrade(StateEnvironment) State {
	panic("Implement this fucntion")
}

func (s *stateBase) TransformKeyWithNewBeaconHeight(beaconHeight uint64) State {
	time1 := time.Now()
	sameHeight := false
	//transform pdex key prefix-<beaconheight>-id1-id2 (if same height, no transform)
	transformKey := func(key string, beaconHeight uint64) string {
		if sameHeight {
			return key
		}
		keySplit := strings.Split(key, "-")
		if keySplit[1] == strconv.Itoa(int(beaconHeight)) {
			sameHeight = true
		}
		keySplit[1] = strconv.Itoa(int(beaconHeight))
		return strings.Join(keySplit, "-")
	}

	newState := &stateBase{}
	newState.waitingContributions = make(map[string]*rawdbv2.PDEContribution)
	newState.deletedWaitingContributions = make(map[string]*rawdbv2.PDEContribution)
	newState.poolPairs = make(map[string]*rawdbv2.PDEPoolForPair)
	newState.shares = make(map[string]uint64)
	newState.tradingFees = make(map[string]uint64)

	for k, v := range s.waitingContributions {
		newState.waitingContributions[transformKey(k, beaconHeight)] = v
		if sameHeight {
			return s
		}
	}
	for k, v := range s.deletedWaitingContributions {
		newState.deletedWaitingContributions[transformKey(k, beaconHeight)] = v
	}
	for k, v := range s.poolPairs {
		newState.poolPairs[transformKey(k, beaconHeight)] = v
	}
	for k, v := range s.shares {
		newState.shares[transformKey(k, beaconHeight)] = v
	}
	for k, v := range s.tradingFees {
		newState.tradingFees[transformKey(k, beaconHeight)] = v
	}
	Logger.log.Infof("Time spent for transforming keys: %f", time.Since(time1).Seconds())
	return newState
}

func (s *stateBase) ClearCache() {
	s.deletedWaitingContributions = make(map[string]*rawdbv2.PDEContribution)
}

func (s *stateBase) GetDiff(compareState State) (State, error) {
	if compareState == nil {
		return nil, errors.New("compareState is nil")
	}

	res := &stateBase{}
	compareStateBase := compareState.(*stateBase)

	res.waitingContributions = make(map[string]*rawdbv2.PDEContribution)
	res.deletedWaitingContributions = make(map[string]*rawdbv2.PDEContribution)
	res.poolPairs = make(map[string]*rawdbv2.PDEPoolForPair)
	res.shares = make(map[string]uint64)
	res.tradingFees = make(map[string]uint64)

	for k, v := range s.waitingContributions {
		if m, ok := compareStateBase.waitingContributions[k]; !ok || !reflect.DeepEqual(m, v) {
			res.waitingContributions[k] = v
		}
	}
	for k, v := range s.deletedWaitingContributions {
		if m, ok := compareStateBase.deletedWaitingContributions[k]; !ok || !reflect.DeepEqual(m, v) {
			res.deletedWaitingContributions[k] = v
		}
	}
	for k, v := range s.poolPairs {
		if m, ok := compareStateBase.poolPairs[k]; !ok || !reflect.DeepEqual(m, v) {
			res.poolPairs[k] = v
		}
	}
	for k, v := range s.shares {
		if m, ok := compareStateBase.shares[k]; !ok || !reflect.DeepEqual(m, v) {
			res.shares[k] = v
		}
	}
	for k, v := range s.tradingFees {
		if m, ok := compareStateBase.tradingFees[k]; !ok || !reflect.DeepEqual(m, v) {
			res.tradingFees[k] = v
		}
	}
	return res, nil
}

func (s *stateBase) WaitingContributions() map[string]*rawdbv2.PDEContribution {
	res := make(map[string]*rawdbv2.PDEContribution, len(s.waitingContributions))
	for k, v := range s.waitingContributions {
		*res[k] = *v
	}
	return res
}

func (s *stateBase) DeletedWaitingContributions() map[string]*rawdbv2.PDEContribution {
	res := make(map[string]*rawdbv2.PDEContribution, len(s.deletedWaitingContributions))
	for k, v := range s.deletedWaitingContributions {
		*res[k] = *v
	}
	return res
}

func (s *stateBase) PoolPairs() map[string]*rawdbv2.PDEPoolForPair {
	res := make(map[string]*rawdbv2.PDEPoolForPair, len(s.poolPairs))
	for k, v := range s.poolPairs {
		*res[k] = *v
	}
	return res
}

func (s *stateBase) Shares() map[string]uint64 {
	res := make(map[string]uint64, len(s.shares))
	for k, v := range s.shares {
		res[k] = v
	}
	return res
}

func (s *stateBase) TradingFees() map[string]uint64 {
	res := make(map[string]uint64, len(s.tradingFees))
	for k, v := range s.tradingFees {
		res[k] = v
	}
	return res
}
