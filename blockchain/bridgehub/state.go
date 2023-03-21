package bridgehub

import (
	"errors"

	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
)

type BridgeHubState struct {
	// TODO: staking asset is PRV or others?
	stakingInfos map[string]*statedb.BridgeStakingInfoState // bridgePubKey : Staker info

	// bridgePubKey only belongs one Bridge
	bridgeInfos map[string]*BridgeInfo // BridgePoolPubKey : BridgeInfo

	tokenPrices map[string]uint64 // pTokenID: price * 1e6

	params *statedb.BridgeHubParamState
}

type BridgeInfo struct {
	Info        *statedb.BridgeInfoState
	NetworkInfo map[int]*statedb.BridgeHubNetworkState // key: networkId
}

// read only function
func (s *BridgeHubState) StakingInfos() map[string]*statedb.BridgeStakingInfoState {
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
		stakingInfos: make(map[string]*statedb.BridgeStakingInfoState),
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

		infoTmp.NetworkInfo = map[int]*statedb.BridgeHubNetworkState{}
		for networkId, networkInfo := range info.NetworkInfo {
			infoTmp.NetworkInfo[networkId] = networkInfo.Clone()
		}
		bridgeInfos[bridgeID] = infoTmp
	}
	res.bridgeInfos = bridgeInfos

	// clone stakingInfo
	stakingInfos := map[string]*statedb.BridgeStakingInfoState{}
	for bridgeHubKey, info := range s.stakingInfos {
		stakingInfos[bridgeHubKey] = statedb.NewBridgeStakingInfoStateWithValue(
			info.StakingAmount(),
			info.TokenID(),
			info.BridgePubKey(),
			info.BridgePoolPubKey(),
		)
	}
	res.stakingInfos = stakingInfos

	// TODO: coding for tokenPrices

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
			for networkId, networkInfo := range bridgeInfo.NetworkInfo {
				isUpdate := true
				if prePTokenInfo, found := preBridge.NetworkInfo[networkId]; found && !prePTokenInfo.IsDiff(networkInfo) {
					isUpdate = false
				}
				if isUpdate {
					if newBridgeInfos[bridgeID] == nil {
						newBridgeInfos[bridgeID] = &BridgeInfo{}
					}
					if newBridgeInfos[bridgeID].NetworkInfo == nil {
						newBridgeInfos[bridgeID].NetworkInfo = map[int]*statedb.BridgeHubNetworkState{}
					}
					newBridgeInfos[bridgeID].NetworkInfo[networkId] = networkInfo
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
