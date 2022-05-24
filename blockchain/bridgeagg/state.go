package bridgeagg

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
)

type State struct {
	unifiedTokenVaults  map[common.Hash]map[common.Hash]*statedb.BridgeAggVaultState // unifiedTokenID -> tokenID -> vault
	waitingUnshieldReqs map[common.Hash][]*statedb.BridgeAggWaitingUnshieldReq       // unifiedTokenID -> unshieldID -> waitingUnshieldReq

	deletedWaitingUnshieldReqKeyHashes []common.Hash
}

// UnifiedTokenVaults read only function do not write to result of function
func (s *State) UnifiedTokenVaults() map[common.Hash]map[common.Hash]*statedb.BridgeAggVaultState {
	return s.unifiedTokenVaults
}

// WaitingUnshieldReqs read only function do not write to result of function
func (s *State) WaitingUnshieldReqs() map[common.Hash][]*statedb.BridgeAggWaitingUnshieldReq {
	return s.waitingUnshieldReqs
}

func (s *State) DeletedWaitingUnshieldReqKeyHashes() []common.Hash {
	return s.deletedWaitingUnshieldReqKeyHashes
}

func NewState() *State {
	return &State{
		unifiedTokenVaults:                 make(map[common.Hash]map[common.Hash]*statedb.BridgeAggVaultState),
		waitingUnshieldReqs:                make(map[common.Hash][]*statedb.BridgeAggWaitingUnshieldReq),
		deletedWaitingUnshieldReqKeyHashes: []common.Hash{},
	}
}

func NewStateWithValue(
	unifiedTokenInfos map[common.Hash]map[common.Hash]*statedb.BridgeAggVaultState,
	waitingUnshieldReqs map[common.Hash][]*statedb.BridgeAggWaitingUnshieldReq,
	deletedWaitingUnshieldReqKeyHashes []common.Hash,
) *State {
	return &State{
		unifiedTokenVaults:                 unifiedTokenInfos,
		waitingUnshieldReqs:                waitingUnshieldReqs,
		deletedWaitingUnshieldReqKeyHashes: deletedWaitingUnshieldReqKeyHashes,
	}
}

func (s *State) Clone() *State {
	res := NewState()
	res.unifiedTokenVaults = s.CloneUnifiedTokenVaults()
	res.waitingUnshieldReqs = s.CloneWaitingUnshieldReqs()

	res.deletedWaitingUnshieldReqKeyHashes = []common.Hash{}
	res.deletedWaitingUnshieldReqKeyHashes = append(res.deletedWaitingUnshieldReqKeyHashes, s.deletedWaitingUnshieldReqKeyHashes...)

	return res
}

//TODO: 0xkraken
func (s *State) GetDiff(compareState *State) (*State, *StateChange, error) {
	res := NewState()
	if compareState == nil {
		return nil, nil, errors.New("compareState is nil")
	}
	stateChange := NewStateChange()

	for unifiedTokenID, vaults := range s.unifiedTokenVaults {
		if compareVaults, found := compareState.unifiedTokenVaults[unifiedTokenID]; !found {
			res.unifiedTokenVaults[unifiedTokenID] = vaults
			stateChange.unifiedTokenID[unifiedTokenID] = true
			stateChange.vaultChange[unifiedTokenID] = make(map[common.Hash]VaultChange)
			for tokenID := range vaults {
				stateChange.vaultChange[unifiedTokenID][tokenID] = VaultChange{IsChanged: true}
			}
		} else {
			for tokenID, vault := range vaults {
				if res.unifiedTokenVaults[unifiedTokenID] == nil {
					res.unifiedTokenVaults[unifiedTokenID] = make(map[common.Hash]*statedb.BridgeAggVaultState)
				}
				if compareVault, found := compareVaults[tokenID]; !found {
					res.unifiedTokenVaults[unifiedTokenID][tokenID] = vault
					stateChange.vaultChange[unifiedTokenID][tokenID] = VaultChange{IsChanged: true}
				} else {
					diff, err := s.unifiedTokenVaults[unifiedTokenID][tokenID].GetDiff(compareVault)
					if err != nil {
						return nil, nil, err
					}
					if diff != nil {
						res.unifiedTokenVaults[unifiedTokenID][tokenID] = diff
						stateChange.vaultChange[unifiedTokenID][tokenID] = VaultChange{IsChanged: true}
					}
				}
			}
		}
	}
	return res, stateChange, nil
}

func (s *State) CloneUnifiedTokenVaults() map[common.Hash]map[common.Hash]*statedb.BridgeAggVaultState {
	res := make(map[common.Hash]map[common.Hash]*statedb.BridgeAggVaultState)
	for unifiedTokenID, vaults := range s.unifiedTokenVaults {
		res[unifiedTokenID] = make(map[common.Hash]*statedb.BridgeAggVaultState)
		for tokenID, vault := range vaults {
			res[unifiedTokenID][tokenID] = vault.Clone()
		}
	}
	return res
}

func (s *State) CloneWaitingUnshieldReqs() map[common.Hash][]*statedb.BridgeAggWaitingUnshieldReq {
	res := make(map[common.Hash][]*statedb.BridgeAggWaitingUnshieldReq)
	for unifiedTokenID, reqs := range s.waitingUnshieldReqs {
		res[unifiedTokenID] = []*statedb.BridgeAggWaitingUnshieldReq{}
		for _, req := range reqs {
			res[unifiedTokenID] = append(res[unifiedTokenID], req.Clone())
		}
	}
	return res
}

func (s *State) CloneVaultsByUnifiedTokenID(unifiedTokenID common.Hash) (map[common.Hash]*statedb.BridgeAggVaultState, error) {
	if vaults, found := s.unifiedTokenVaults[unifiedTokenID]; found {
		res := make(map[common.Hash]*statedb.BridgeAggVaultState)
		for networkID, vault := range vaults {
			res[networkID] = vault.Clone()
		}
		return res, nil
	} else {
		return nil, fmt.Errorf("Can't find unifiedTokenID %s", unifiedTokenID.String())
	}
}

func (s *State) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(struct {
		UnifiedTokenVaults                 map[common.Hash]map[common.Hash]*statedb.BridgeAggVaultState `json:"UnifiedTokenVaults"`
		WaitingUnshieldReqs                map[common.Hash][]*statedb.BridgeAggWaitingUnshieldReq       `json:"WaitingUnshieldReqs"`
		DeletedWaitingUnshieldReqKeyHashes []common.Hash                                                `json:"DeletedWaitingUnshieldReqKeyHashes"`
	}{
		UnifiedTokenVaults:                 s.unifiedTokenVaults,
		WaitingUnshieldReqs:                s.waitingUnshieldReqs,
		DeletedWaitingUnshieldReqKeyHashes: s.deletedWaitingUnshieldReqKeyHashes,
	})
	if err != nil {
		return []byte{}, err
	}
	return data, nil
}

func (s *State) UnmarshalJSON(data []byte) error {
	temp := struct {
		UnifiedTokenVaults                 map[common.Hash]map[common.Hash]*statedb.BridgeAggVaultState `json:"UnifiedTokenVaults"`
		WaitingUnshieldReqs                map[common.Hash][]*statedb.BridgeAggWaitingUnshieldReq       `json:"WaitingUnshieldReqs"`
		DeletedWaitingUnshieldReqKeyHashes []common.Hash                                                `json:"DeletedWaitingUnshieldReqKeyHashes"`
	}{}
	err := json.Unmarshal(data, &temp)
	if err != nil {
		return err
	}
	s.unifiedTokenVaults = temp.UnifiedTokenVaults
	s.waitingUnshieldReqs = temp.WaitingUnshieldReqs
	s.deletedWaitingUnshieldReqKeyHashes = temp.DeletedWaitingUnshieldReqKeyHashes
	return nil
}
