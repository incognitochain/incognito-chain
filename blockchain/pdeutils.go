package blockchain

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/incognitochain/incognito-chain/dataaccessobject/rawdbv2"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
)

type CurrentPDEState struct {
	WaitingPDEContributions        map[string]*rawdbv2.PDEContribution
	DeletedWaitingPDEContributions map[string]*rawdbv2.PDEContribution
	PDEPoolPairs                   map[string]*rawdbv2.PDEPoolForPair
	PDEShares                      map[string]uint64
	PDETradingFees                 map[string]uint64
}

func (s *CurrentPDEState) Copy() *CurrentPDEState {
	v := new(CurrentPDEState)
	b := new(bytes.Buffer)
	err := gob.NewEncoder(b).Encode(s)
	if err != nil {
		return nil
	}
	err = gob.NewDecoder(bytes.NewBuffer(b.Bytes())).Decode(v)
	if err != nil {
		return nil
	}
	return v
}

type DeductingAmountsByWithdrawal struct {
	Token1IDStr string
	PoolValue1  uint64
	Token2IDStr string
	PoolValue2  uint64
	Shares      uint64
}

type DeductingAmountsByWithdrawalWithPRVFee struct {
	Token1IDStr   string
	PoolValue1    uint64
	Token2IDStr   string
	PoolValue2    uint64
	Shares        uint64
	FeeTokenIDStr string
	FeeAmount     uint64
}

func (lastState *CurrentPDEState) transformKeyWithNewBeaconHeight(beaconHeight uint64) *CurrentPDEState {
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

	newState := &CurrentPDEState{}
	newState.WaitingPDEContributions = make(map[string]*rawdbv2.PDEContribution)
	newState.DeletedWaitingPDEContributions = make(map[string]*rawdbv2.PDEContribution)
	newState.PDEPoolPairs = make(map[string]*rawdbv2.PDEPoolForPair)
	newState.PDEShares = make(map[string]uint64)
	newState.PDETradingFees = make(map[string]uint64)

	for k, v := range lastState.WaitingPDEContributions {
		newState.WaitingPDEContributions[transformKey(k, beaconHeight)] = v
		if sameHeight {
			return lastState
		}
	}
	for k, v := range lastState.DeletedWaitingPDEContributions {
		newState.DeletedWaitingPDEContributions[transformKey(k, beaconHeight)] = v
	}
	for k, v := range lastState.PDEPoolPairs {
		newState.PDEPoolPairs[transformKey(k, beaconHeight)] = v
	}
	for k, v := range lastState.PDEShares {
		newState.PDEShares[transformKey(k, beaconHeight)] = v
	}
	for k, v := range lastState.PDETradingFees {
		newState.PDETradingFees[transformKey(k, beaconHeight)] = v
	}
	Logger.log.Infof("Time spent for transforming keys: %f", time.Since(time1).Seconds())
	return newState
}

func InitCurrentPDEStateFromDB(
	stateDB *statedb.StateDB,
	beaconHeight uint64,
) (*CurrentPDEState, error) {
	// TODO: @tin change here with pdexState in beaconBestState
	/*if lastState != nil {*/
	//newState := lastState.transformKeyWithNewBeaconHeight(beaconHeight)
	//return newState, nil
	/*}*/
	waitingPDEContributions, err := statedb.GetWaitingPDEContributions(stateDB, beaconHeight)
	if err != nil {
		return nil, err
	}
	pdePoolPairs, err := statedb.GetPDEPoolPair(stateDB, beaconHeight)
	if err != nil {
		return nil, err
	}
	pdeShares, err := statedb.GetPDEShares(stateDB, beaconHeight)
	if err != nil {
		return nil, err
	}
	pdeTradingFees, err := statedb.GetPDETradingFees(stateDB, beaconHeight)
	if err != nil {
		return nil, err
	}

	return &CurrentPDEState{
		WaitingPDEContributions:        waitingPDEContributions,
		PDEPoolPairs:                   pdePoolPairs,
		PDEShares:                      pdeShares,
		PDETradingFees:                 pdeTradingFees,
		DeletedWaitingPDEContributions: make(map[string]*rawdbv2.PDEContribution),
	}, nil
}

func storePDEStateToDB(
	stateDB *statedb.StateDB,
	currentPDEState *CurrentPDEState,
) error {
	var err error
	statedb.DeleteWaitingPDEContributions(stateDB, currentPDEState.DeletedWaitingPDEContributions)
	err = statedb.StoreWaitingPDEContributions(stateDB, currentPDEState.WaitingPDEContributions)
	if err != nil {
		return err
	}
	err = statedb.StorePDEPoolPairs(stateDB, currentPDEState.PDEPoolPairs)
	if err != nil {
		return err
	}
	err = statedb.StorePDEShares(stateDB, currentPDEState.PDEShares)
	if err != nil {
		return err
	}
	err = statedb.StorePDETradingFees(stateDB, currentPDEState.PDETradingFees)
	if err != nil {
		return err
	}
	return nil
}

func replaceNewBCHeightInKeyStr(key string, newBeaconHeight uint64) string {
	parts := strings.Split(key, "-")
	if len(parts) <= 1 {
		return key
	}
	parts[1] = fmt.Sprintf("%d", newBeaconHeight)
	newKey := ""
	for idx, part := range parts {
		if idx == len(parts)-1 {
			newKey += part
			continue
		}
		newKey += (part + "-")
	}
	return newKey
}
