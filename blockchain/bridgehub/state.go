package bridgehub

import (
	"errors"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
)

type BridgeHubState struct {
	// TODO: staking asset is PRV or others?
	stakingInfos map[string]*StakerInfo // bridgePubKey : Staker info

	// bridgePubKey only belongs one Bridge
	bridgeInfos map[string]*BridgeInfo // BridgePoolPubKey : BridgeInfo

	tokenPrices map[string]uint64 // pTokenID: price * 1e6

	params *statedb.BridgeHubParamState
}

type StakerInfo struct {
	StakeAmount      uint64      `json:"StakeAmount"`
	TokenID          common.Hash `json:"TokenID"`
	TxReqID          string      `json:"TxReqID"`
	BridgePubKey     string      `json:"BridgePubKey"`
	BridgePoolPubKey string      `json:"BridgePoolPubKey"`
}

type BridgeInfo struct {
	Info          *statedb.BridgeInfoState
	PTokenAmounts map[common.Hash]*statedb.BridgeHubPTokenState // key: pToken
}

// read only function
func (s *BridgeHubState) StakingInfos() map[string]*StakerInfo {
	return s.stakingInfos
}

func (s *BridgeHubState) BridgeInfos() map[string]*BridgeInfo {
	return s.bridgeInfos
}

func (s *BridgeHubState) TokenPrices() map[string]uint64 {
	return s.tokenPrices
}
func (s *BridgeHubState) Params() *statedb.BridgeHubParamState {
	return s.params
}

func NewBridgeHubState() *BridgeHubState {
	return &BridgeHubState{
		stakingInfos: make(map[string]*StakerInfo),
		bridgeInfos:  make(map[string]*BridgeInfo),
		tokenPrices:  make(map[string]uint64),
		params:       nil,
	}
}

func (s *BridgeHubState) Clone() *BridgeHubState {
	res := NewBridgeHubState()

	if s.params != nil {
		res.params = s.params.Clone()
	}

	// clone bridgeInfos
	bridgeInfos := map[string]*BridgeInfo{}
	for bridgeID, info := range s.bridgeInfos {
		infoTmp := &BridgeInfo{}
		infoTmp.Info = info.Info.Clone()

		infoTmp.PTokenAmounts = map[common.Hash]*statedb.BridgeHubPTokenState{}
		for ptokenID, pTokenState := range info.PTokenAmounts {
			infoTmp.PTokenAmounts[ptokenID] = pTokenState.Clone()
		}
		bridgeInfos[bridgeID] = infoTmp
	}
	res.bridgeInfos = bridgeInfos

	// TODO: coding for stakingInfo, tokenPrices

	return res
}

func (s *BridgeHubState) GetDiff(preState *BridgeHubState) (*BridgeHubState, error) {
	if preState == nil {
		return nil, errors.New("preState is nil")
	}

	diffState := NewBridgeHubState()

	// get diff bridgeInfos
	newBridgeInfos := map[string]*BridgeInfo{}
	for bridgeID, bridgeInfo := range s.bridgeInfos {
		isUpdateBridgeInfo := false
		if preBridge, found := preState.bridgeInfos[bridgeID]; found {
			// check info
			isUpdateBridgeInfo = preBridge.Info.IsDiff(bridgeInfo.Info)

			// check list ptoken
			for pTokenID, pTokenInfo := range bridgeInfo.PTokenAmounts {
				isUpdate := true
				if prePTokenInfo, found := preBridge.PTokenAmounts[pTokenID]; found && !prePTokenInfo.IsDiff(pTokenInfo) {
					isUpdate = false
				}
				if isUpdate {
					if newBridgeInfos[bridgeID] == nil {
						newBridgeInfos[bridgeID] = &BridgeInfo{}
					}
					if newBridgeInfos[bridgeID].PTokenAmounts == nil {
						newBridgeInfos[bridgeID].PTokenAmounts = map[common.Hash]*statedb.BridgeHubPTokenState{}
					}
					newBridgeInfos[bridgeID].PTokenAmounts[pTokenID] = pTokenInfo
				}
			}

		} else {
			isUpdateBridgeInfo = true
		}

		if isUpdateBridgeInfo {
			if newBridgeInfos[bridgeID] == nil {
				newBridgeInfos[bridgeID] = &BridgeInfo{}
			}
			newBridgeInfos[bridgeID].Info = bridgeInfo.Info
		}
	}
	diffState.bridgeInfos = newBridgeInfos

	// get diff param
	if s.params != nil && s.params.IsDiff(preState.params) {
		diffState.params = s.params
	} else {
		diffState.params = nil
	}

	// get diff staking info
	for k, v := range s.stakingInfos {
		if preState.stakingInfos[k] != v {
			diffState.stakingInfos[k] = v
		}
	}

	// TODO: coding for tokenPrices

	return diffState, nil
}
