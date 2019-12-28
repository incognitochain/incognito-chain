package statedb

import (
	"fmt"
	"strings"

	"github.com/incognitochain/incognito-chain/dataaccessobject/rawdb"
)

// tempkey: WaitingPDEContributionPrefix - beacon height - pairid
func StoreWaitingPDEContributions(stateDB *StateDB, beaconHeight uint64, waitingPDEContributions map[string]*rawdb.PDEContribution) error {
	for tempKey, contribution := range waitingPDEContributions {
		strs := strings.Split(tempKey, "-")
		pairID := strs[2]
		key := GenerateWaitingPDEContributionObjectKey(beaconHeight, pairID)
		value := NewWaitingPDEContributionStateWithValue(beaconHeight, pairID, contribution.ContributorAddressStr, contribution.TokenIDStr, contribution.Amount, contribution.TxReqID)
		err := stateDB.SetStateObject(WaitingPDEContributionObjectType, key, value)
		if err != nil {
			return NewStatedbError(StoreWaitingPDEContributionError, err)
		}
	}
	return nil
}

func GetWaitingPDEContributions(stateDB *StateDB, beaconHeight uint64) (map[string]*rawdb.PDEContribution, error) {
	waitingPDEContributions := make(map[string]*rawdb.PDEContribution)
	waitingPDEContributionStates := stateDB.GetAllWaitingPDEContributionState(beaconHeight)
	for _, wcState := range waitingPDEContributionStates {
		key := string(WaitingPDEContributionPrefix()) + fmt.Sprintf("%d-", beaconHeight) + wcState.PairID()
		value := rawdb.NewPDEContribution(wcState.ContributorAddress(), wcState.TokenID(), wcState.Amount(), wcState.TxReqID())
		waitingPDEContributions[key] = value
	}
	return waitingPDEContributions, nil
}

// tempkey: PDEPoolPrefix - beacon height - token1ID - token2ID
func StorePDEPoolPairs(
	stateDB *StateDB,
	beaconHeight uint64,
	pdePoolPairs map[string]*rawdb.PDEPoolForPair,
) error {
	for _, pdePoolPair := range pdePoolPairs {
		key := GeneratePDEPoolPairObjectKey(beaconHeight, pdePoolPair.Token1IDStr, pdePoolPair.Token2IDStr)
		value := NewPDEPoolPairStateWithValue(beaconHeight, pdePoolPair.Token1IDStr, pdePoolPair.Token1PoolValue, pdePoolPair.Token2IDStr, pdePoolPair.Token2PoolValue)
		err := stateDB.SetStateObject(PDEPoolPairObjectType, key, value)
		if err != nil {
			return NewStatedbError(StorePDEPoolPairError, err)
		}
	}
	return nil
}

func GetPDEPoolPair(stateDB *StateDB, beaconHeight uint64) (map[string]*rawdb.PDEPoolForPair, error) {
	pdePoolPairs := make(map[string]*rawdb.PDEPoolForPair)
	pdePoolPairStates := stateDB.GetAllPDEPoolPairState(beaconHeight)
	for _, ppState := range pdePoolPairStates {
		key := string(PDEPoolPrefix()) + fmt.Sprintf("%d-", beaconHeight) + ppState.Token1ID() + "-" + ppState.Token2ID()
		value := rawdb.NewPDEPoolForPair(ppState.Token1ID(), ppState.Token1PoolValue(), ppState.Token2ID(), ppState.Token2PoolValue())
		pdePoolPairs[key] = value
	}
	return pdePoolPairs, nil
}

func StorePDEShares(stateDB *StateDB, beaconHeight uint64, pdeShares map[string]uint64) error {
	for tempKey, shareAmount := range pdeShares {
		strs := strings.Split(tempKey, "-")
		token1ID := strs[2]
		token2ID := strs[3]
		contributorAddress := strs[4]
		key := GeneratePDEShareObjectKey(beaconHeight, token1ID, token2ID, contributorAddress)
		value := NewPDEShareStateWithValue(beaconHeight, token1ID, token2ID, contributorAddress, shareAmount)
		err := stateDB.SetStateObject(PDEShareObjectType, key, value)
		if err != nil {
			return NewStatedbError(StorePDEShareError, err)
		}
	}
	return nil
}

func GetPDEShares(stateDB *StateDB, beaconHeight uint64) (map[string]uint64, error) {
	pdeShares := make(map[string]uint64)
	pdeShareStates := stateDB.GetAllPDEShareState(beaconHeight)
	for _, sState := range pdeShareStates {
		key := string(PDESharePrefix()) + fmt.Sprintf("%d-", beaconHeight) + sState.Token1ID() + "-" + sState.Token2ID() + "-" + sState.ContributorAddress()
		value := sState.Amount()
		pdeShares[key] = value
	}
	return pdeShares, nil
}
