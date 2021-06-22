package pdex

import (
	"strconv"
	"strings"
	"time"

	"github.com/incognitochain/incognito-chain/dataaccessobject/rawdbv2"
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
	return nil
}

func (s *stateBase) BuildInstructions(env StateEnvironment) ([][]string, error) {
	return nil, nil
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
