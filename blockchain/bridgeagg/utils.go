package bridgeagg

import "github.com/incognitochain/incognito-chain/common"

type ModifyListTokenStatus struct {
	Status    byte `json:"Status"`
	ErrorCode uint `json:"ErrorCode,omitempty"`
}

type StateChange struct {
	unifiedTokenID map[common.Hash]bool
}

func NewStateChange() *StateChange {
	return &StateChange{}
}
