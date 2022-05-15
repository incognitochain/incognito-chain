package bridgeagg

import (
	"errors"
	"fmt"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
)

type State struct {
	unifiedTokenInfos   map[common.Hash]map[common.Hash]*statedb.BridgeAggVaultState         // unifiedTokenID -> tokenID -> vault
	waitingUnshieldReqs map[common.Hash]map[common.Hash]*statedb.BridgeAggWaitingUnshieldReq // unifiedTokenID -> unshieldID -> waitingUnshieldReq

	deletedWaitingUnshieldReqKeyHashes []common.Hash
}

func (s *State) UnifiedTokenInfos() map[common.Hash]map[common.Hash]*statedb.BridgeAggVaultState {
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
		unifiedTokenInfos:                  make(map[common.Hash]map[common.Hash]*statedb.BridgeAggVaultState),
		waitingUnshieldReqs:                make(map[common.Hash]map[common.Hash]*statedb.BridgeAggWaitingUnshieldReq),
		deletedWaitingUnshieldReqKeyHashes: []common.Hash{},
	}
}

func NewStateWithValue(
	unifiedTokenInfos map[common.Hash]map[common.Hash]*statedb.BridgeAggVaultState,
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
		res.unifiedTokenInfos[unifiedTokenID] = make(map[common.Hash]*statedb.BridgeAggVaultState)
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
	if compareState == nil {
		return nil, nil, errors.New("compareState is nil")
	}
	stateChange := NewStateChange()

	for unifiedTokenID, vaults := range s.unifiedTokenInfos {
		if compareVaults, found := compareState.unifiedTokenInfos[unifiedTokenID]; !found {
			res.unifiedTokenInfos[unifiedTokenID] = vaults
			stateChange.unifiedTokenID[unifiedTokenID] = true
			stateChange.vaultChange[unifiedTokenID] = make(map[common.Hash]VaultChange)
			for tokenID := range vaults {
				stateChange.vaultChange[unifiedTokenID][tokenID] = VaultChange{IsChanged: true}
			}
		} else {
			for tokenID, vault := range vaults {
				if res.unifiedTokenInfos[unifiedTokenID] == nil {
					res.unifiedTokenInfos[unifiedTokenID] = make(map[common.Hash]*statedb.BridgeAggVaultState)
				}
				if compareVault, found := compareVaults[tokenID]; !found {
					res.unifiedTokenInfos[unifiedTokenID][tokenID] = vault
					stateChange.vaultChange[unifiedTokenID][tokenID] = VaultChange{IsChanged: true}
				} else {
					diff, err := s.unifiedTokenInfos[unifiedTokenID][tokenID].GetDiff(compareVault)
					if err != nil {
						return nil, nil, err
					}
					if diff != nil {
						res.unifiedTokenInfos[unifiedTokenID][tokenID] = diff
						stateChange.vaultChange[unifiedTokenID][tokenID] = VaultChange{IsChanged: true}
					}
				}
			}
		}
	}
	return res, stateChange, nil
}

func (s *State) CloneUnifiedTokenInfos() map[common.Hash]map[common.Hash]*statedb.BridgeAggVaultState {
	res := make(map[common.Hash]map[common.Hash]*statedb.BridgeAggVaultState)
	for unifiedTokenID, vaults := range s.unifiedTokenInfos {
		res[unifiedTokenID] = make(map[common.Hash]*statedb.BridgeAggVaultState)
		for tokenID, vault := range vaults {
			res[unifiedTokenID][tokenID] = vault.Clone()
		}
	}
	return res
}

func (s *State) CloneVaultsByUnifiedTokenID(unifiedTokenID common.Hash) (map[common.Hash]*statedb.BridgeAggVaultState, error) {
	if vaults, found := s.unifiedTokenInfos[unifiedTokenID]; found {
		res := make(map[common.Hash]*statedb.BridgeAggVaultState)
		for networkID, vault := range vaults {
			res[networkID] = vault.Clone()
		}
		return res, nil
	} else {
		return nil, fmt.Errorf("Can't find unifiedTokenID %s", unifiedTokenID.String())
	}
}
