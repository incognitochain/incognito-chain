package bridgeagg

import (
	"bytes"
	"errors"
	"sort"
	"strconv"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/metadata"
)

type State struct {
	unifiedTokenInfos   map[common.Hash]map[common.Hash]*Vault                               // unifiedTokenID -> tokenID -> vault
	waitingUnshieldReqs map[common.Hash]map[common.Hash]*statedb.BridgeAggWaitingUnshieldReq // unifiedTokenID -> unshieldID -> waitingUnshieldReq

	deletedWaitingUnshieldReqKeyHashes []common.Hash
}

func (s *State) UnifiedTokenInfos() map[common.Hash]map[common.Hash]*Vault {
	return s.unifiedTokenInfos
}

func (s *State) WaitingUnshieldReqs() map[common.Hash]map[common.Hash]*statedb.BridgeAggWaitingUnshieldReq {
	return s.waitingUnshieldReqs
}

func (s *State) DeletedWaitingUnshieldReqKeyHashes() []common.Hash {
	return s.deletedWaitingUnshieldReqKeyHashes
}

func NewState() *State {
	return &State{
		unifiedTokenInfos:                  make(map[common.Hash]map[common.Hash]*Vault),
		waitingUnshieldReqs:                make(map[common.Hash]map[common.Hash]*statedb.BridgeAggWaitingUnshieldReq),
		deletedWaitingUnshieldReqKeyHashes: []common.Hash{},
	}
}

func NewStateWithValue(
	unifiedTokenInfos map[common.Hash]map[common.Hash]*Vault,
	waitingUnshieldReqs map[common.Hash]map[common.Hash]*statedb.BridgeAggWaitingUnshieldReq,
	deletedWaitingUnshieldReqKeyHashes []common.Hash,
) *State {
	return &State{
		unifiedTokenInfos:                  unifiedTokenInfos,
		waitingUnshieldReqs:                waitingUnshieldReqs,
		deletedWaitingUnshieldReqKeyHashes: deletedWaitingUnshieldReqKeyHashes,
	}
}

func (s *State) Clone() *State {
	res := NewState()
	for unifiedTokenID, vaults := range s.unifiedTokenInfos {
		res.unifiedTokenInfos[unifiedTokenID] = make(map[common.Hash]*Vault)
		for networkID, vault := range vaults {
			res.unifiedTokenInfos[unifiedTokenID][networkID] = vault.Clone()
		}
	}

	for unifiedTokenID, reqs := range s.waitingUnshieldReqs {
		res.waitingUnshieldReqs[unifiedTokenID] = make(map[common.Hash]*statedb.BridgeAggWaitingUnshieldReq)
		for unshieldID, req := range reqs {
			res.waitingUnshieldReqs[unifiedTokenID][unshieldID] = req.Clone()
		}
	}

	res.deletedWaitingUnshieldReqKeyHashes = []common.Hash{}
	res.deletedWaitingUnshieldReqKeyHashes = append(res.deletedWaitingUnshieldReqKeyHashes, s.deletedWaitingUnshieldReqKeyHashes...)

	return res
}

func (s *State) GetDiff(compareState *State) (*State, *StateChange, error) {
	res := NewState()
	stateChange := NewStateChange()
	if compareState == nil {
		return nil, nil, errors.New("compareState is nil")
	}

	for unifiedTokenID, vaults := range s.unifiedTokenInfos {
		if compareVaults, found := compareState.unifiedTokenInfos[unifiedTokenID]; !found {
			res.unifiedTokenInfos[unifiedTokenID] = vaults
			stateChange.unifiedTokenID[unifiedTokenID] = true
		} else {
			for tokenID, vault := range vaults {
				if res.unifiedTokenInfos[unifiedTokenID] == nil {
					res.unifiedTokenInfos[unifiedTokenID] = make(map[common.Hash]*Vault)
				}
				if compareVault, ok := compareVaults[tokenID]; !ok {
					res.unifiedTokenInfos[unifiedTokenID][tokenID] = vault
					if stateChange.vaultChange[unifiedTokenID] == nil {
						stateChange.vaultChange[unifiedTokenID] = make(map[common.Hash]VaultChange)
					}
					stateChange.vaultChange[unifiedTokenID][tokenID] = VaultChange{
						IsChanged: true,
					}
				} else {
					temp, vaultChange, err := s.unifiedTokenInfos[unifiedTokenID][tokenID].GetDiff(compareVault)
					if err != nil {
						return nil, nil, err
					}
					if temp != nil {
						res.unifiedTokenInfos[unifiedTokenID][tokenID] = temp
						if stateChange.vaultChange[unifiedTokenID] == nil {
							stateChange.vaultChange[unifiedTokenID] = make(map[common.Hash]VaultChange)
						}
						stateChange.vaultChange[unifiedTokenID][tokenID] = *vaultChange
					}
				}
			}
		}
	}
	return res, stateChange, nil
}

// func (s *State) ClearCache() {
// 	s.processor.clearCache()
// }

// func (s *State) UnifiedTokenIDCached(txReqID common.Hash) (common.Hash, error) {
// 	if res, found := s.processor.UnshieldTxsCache[txReqID]; found {
// 		return res, nil
// 	} else {
// 		return common.Hash{}, fmt.Errorf("txID %s not found in cache", txReqID.String())
// 	}
// }

func (s *State) BuildAddTokenInstruction(beaconHeight uint64, sDBs map[int]*statedb.StateDB, ac *metadata.AccumulatedValues, triggeredFeature map[string]uint64) ([][]string, *metadata.AccumulatedValues, error) {
	res := [][]string{}
	var err error

	checkpoints := []uint64{}
	for k := range triggeredFeature {
		if len(k) > len(unifiedTokenTriggerPrefix) {
			if bytes.Equal([]byte(k[:len(unifiedTokenTriggerPrefix)]), []byte(unifiedTokenTriggerPrefix)) {
				checkpoint, err := strconv.ParseUint(k[len(unifiedTokenTriggerPrefix):], 10, 64)
				if err != nil {
					return [][]string{}, ac, err
				}
				checkpoints = append(checkpoints, checkpoint)
			}
		}
	}
	if len(checkpoints) == 0 {
		return [][]string{}, ac, nil
	}
	sort.Slice(checkpoints, func(i, j int) bool {
		return checkpoints[i] < checkpoints[j]
	})

	// after beacon generate autoenablefeature instruction, TriggerFeature will mark the height of the trigger time.
	// we only need to process once at the mark time (block after trigger)
	//ex: trigger at 8, block 9 will process punified config
	checkpointIndex := -1
	for index, checkpoint := range checkpoints {
		if beaconHeight == triggeredFeature[unifiedTokenTriggerPrefix+strconv.FormatUint(checkpoint, 10)]+1 {
			checkpointIndex = index
			break
		}
	}
	if checkpointIndex == -1 {
		return [][]string{}, ac, nil
	}
	res, s.unifiedTokenInfos, ac, err = addTokenProducer(s.unifiedTokenInfos, beaconHeight, sDBs, ac, checkpoints[checkpointIndex])
	return res, ac, err
}
