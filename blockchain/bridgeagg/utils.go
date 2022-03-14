package bridgeagg

import "github.com/incognitochain/incognito-chain/common"

type ModifyListTokenStatus struct {
	Status    byte `json:"Status"`
	ErrorCode uint `json:"ErrorCode,omitempty"`
}

type ConvertStatus struct {
	Status    byte `json:"Status"`
	ErrorCode uint `json:"ErrorCode,omitempty"`
}

type VaultChange struct {
	IsChanged        bool
	IsReserveChanged bool
}

func NewVaultChange() *VaultChange {
	return &VaultChange{}
}

type StateChange struct {
	unifiedTokenID map[common.Hash]bool
	vaultChange    map[common.Hash]map[uint]VaultChange
}

func NewStateChange() *StateChange {
	return &StateChange{
		unifiedTokenID: make(map[common.Hash]bool),
		vaultChange:    make(map[common.Hash]map[uint]VaultChange),
	}
}
