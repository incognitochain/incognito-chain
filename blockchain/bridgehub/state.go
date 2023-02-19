package bridgehub

import (
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
)

type BridgeHubState struct {
	// TODO: staking asset is PRV or others?
	stakingInfos map[string]uint64 // bridgePubKey : amount PRV stake

	// bridgePubKey only belongs one Bridge
	bridgeInfos map[string]*BridgeInfo // BridgeID : BridgeInfo

	tokenPrices map[string]uint64 // pTokenID: price * 1e6

	params *statedb.BridgeHubParamState
}

type BridgeInfo struct {
	Info          *statedb.BridgeInfoState
	PTokenAmounts map[string]*statedb.BridgeHubPTokenState // key: pToken
}

func (s *BridgeHubState) StakingInfos() map[string]uint64 {
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
	return &BridgeHubState{}
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

		infoTmp.PTokenAmounts = map[string]*statedb.BridgeHubPTokenState{}
		for ptokenID, pTokenState := range info.PTokenAmounts {
			infoTmp.PTokenAmounts[ptokenID] = pTokenState.Clone()
		}
		bridgeInfos[bridgeID] = infoTmp
	}

	// TODO 0xkraken: code more

	return res
}
