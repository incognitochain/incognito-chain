package statedb

import (
	"encoding/json"
	"fmt"
	"github.com/incognitochain/incognito-chain/dataaccessobject/rawdbv2"
	"sort"
	"strings"
)

func StoreWaitingPDEContributions(stateDB *StateDB, beaconHeight uint64, waitingPDEContributions map[string]*rawdbv2.PDEContribution) error {
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

func GetWaitingPDEContributions(stateDB *StateDB, beaconHeight uint64) (map[string]*rawdbv2.PDEContribution, error) {
	waitingPDEContributions := make(map[string]*rawdbv2.PDEContribution)
	waitingPDEContributionStates := stateDB.GetAllWaitingPDEContributionState(beaconHeight)
	for _, wcState := range waitingPDEContributionStates {
		key := string(GetWaitingPDEContributionKey(beaconHeight, wcState.PairID()))
		value := rawdbv2.NewPDEContribution(wcState.ContributorAddress(), wcState.TokenID(), wcState.Amount(), wcState.TxReqID())
		waitingPDEContributions[key] = value
	}
	return waitingPDEContributions, nil
}

func StorePDEPoolPairs(stateDB *StateDB, beaconHeight uint64, pdePoolPairs map[string]*rawdbv2.PDEPoolForPair) error {
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

func GetPDEPoolPair(stateDB *StateDB, beaconHeight uint64) (map[string]*rawdbv2.PDEPoolForPair, error) {
	pdePoolPairs := make(map[string]*rawdbv2.PDEPoolForPair)
	pdePoolPairStates := stateDB.GetAllPDEPoolPairState(beaconHeight)
	for _, ppState := range pdePoolPairStates {
		key := string(GetPDEPoolForPairKey(beaconHeight, ppState.Token1ID(), ppState.Token2ID()))
		value := rawdbv2.NewPDEPoolForPair(ppState.Token1ID(), ppState.Token1PoolValue(), ppState.Token2ID(), ppState.Token2PoolValue())
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
		key := string(GetPDEShareKey(beaconHeight, sState.Token1ID(), sState.Token2ID(), sState.ContributorAddress()))
		value := sState.Amount()
		pdeShares[key] = value
	}
	return pdeShares, nil
}

func GetPDEPoolForPair(stateDB *StateDB, beaconHeight uint64, tokenIDToBuy string, tokenIDToSell string) ([]byte, error) {
	tokenIDs := []string{tokenIDToBuy, tokenIDToSell}
	sort.Strings(tokenIDs)
	key := GeneratePDEPoolPairObjectKey(beaconHeight, tokenIDs[0], tokenIDs[1])
	ppState, has, err := stateDB.GetPDEPoolPairState(key)
	if err != nil {
		return []byte{}, NewStatedbError(GetPDEPoolForPairError, err)
	}
	if !has {
		return []byte{}, NewStatedbError(GetPDEPoolForPairError, fmt.Errorf("key with beacon height %+v, token1ID %+v, token2ID %+v not found", beaconHeight, tokenIDToBuy, tokenIDToSell))
	}
	res, err := json.Marshal(rawdbv2.NewPDEPoolForPair(ppState.Token1ID(), ppState.Token1PoolValue(), ppState.Token2ID(), ppState.Token2PoolValue()))
	if err != nil {
		return []byte{}, NewStatedbError(GetPDEPoolForPairError, err)
	}
	return res, nil
}

func GetLatestPDEPoolForPair(stateDB *StateDB, tokenIDToBuy string, tokenIDToSell string) ([]byte, error) {
	return []byte{}, NewStatedbError(MethodNotSupportError, fmt.Errorf("Use method GetPDEPoolForPair instead"))
}

func TrackPDEStatus(stateDB *StateDB, statusType []byte, statusSuffix []byte, statusContent byte) error {
	key := GeneratePDEStatusObjectKey(statusType, statusSuffix)
	value := NewPDEStatusStateWithValue(statusType, statusSuffix, []byte{statusContent})
	err := stateDB.SetStateObject(PDEStatusObjectType, key, value)
	if err != nil {
		return NewStatedbError(TrackPDEStatusError, err)
	}
	return nil
}

func GetPDEStatus(stateDB *StateDB, statusType []byte, statusSuffix []byte) (byte, error) {
	key := GeneratePDEStatusObjectKey(statusType, statusSuffix)
	s, has, err := stateDB.GetPDEStatusByKey(key)
	if err != nil {
		return 0, NewStatedbError(GetPDEStatusError, err)
	}
	if !has {
		return 0, NewStatedbError(GetPDEStatusError, fmt.Errorf("status %+v with prefix %+v not found", string(statusType), string(statusSuffix)))
	}
	return s.statusContent[0], nil
}

func TrackPDEContributionStatus(stateDB *StateDB, statusType []byte, statusSuffix []byte, statusContent []byte) error {
	key := GeneratePDEStatusObjectKey(statusType, statusSuffix)
	value := NewPDEStatusStateWithValue(statusType, statusSuffix, statusContent)
	err := stateDB.SetStateObject(PDEStatusObjectType, key, value)
	if err != nil {
		return NewStatedbError(TrackPDEStatusError, err)
	}
	return nil
}

func GetPDEContributionStatus(stateDB *StateDB, statusType []byte, statusSuffix []byte) ([]byte, error) {
	key := GeneratePDEStatusObjectKey(statusType, statusSuffix)
	s, has, err := stateDB.GetPDEStatusByKey(key)
	if err != nil {
		return []byte{}, NewStatedbError(GetPDEStatusError, err)
	}
	if !has {
		return []byte{}, NewStatedbError(GetPDEStatusError, fmt.Errorf("status %+v with prefix %+v not found", string(statusType), string(statusSuffix)))
	}
	return s.statusContent, nil
}
