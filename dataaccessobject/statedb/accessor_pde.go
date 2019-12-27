package statedb

import (
	"strings"

	"github.com/incognitochain/incognito-chain/dataaccessobject/rawdb"
)

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

func StorePDEShares(
	stateDB *StateDB,
	beaconHeight uint64,
	pdeShares map[string]uint64,
) error {
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
