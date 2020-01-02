package blockchain

import (
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
)

func InitCurrentPDEStateFromDBV2(
	stateDB *statedb.StateDB,
	beaconHeight uint64,
) (*CurrentPDEState, error) {
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
	return &CurrentPDEState{
		WaitingPDEContributions: waitingPDEContributions,
		PDEPoolPairs:            pdePoolPairs,
		PDEShares:               pdeShares,
	}, nil
}

func storePDEStateToDBV2(
	stateDB *statedb.StateDB,
	beaconHeight uint64,
	currentPDEState *CurrentPDEState,
) error {
	err := statedb.StoreWaitingPDEContributions(stateDB, beaconHeight, currentPDEState.WaitingPDEContributions)
	if err != nil {
		return err
	}
	err = statedb.StorePDEPoolPairs(stateDB, beaconHeight, currentPDEState.PDEPoolPairs)
	if err != nil {
		return err
	}
	err = statedb.StorePDEShares(stateDB, beaconHeight, currentPDEState.PDEShares)
	if err != nil {
		return err
	}
	return nil
}
