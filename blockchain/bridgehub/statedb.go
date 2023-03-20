package bridgehub

import (
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
)

func InitManager(sDB *statedb.StateDB) (*Manager, error) {
	state, err := InitStateFromDB(sDB)
	if err != nil {
		return nil, err
	}
	return NewManagerWithValue(state), nil
}

func InitStateFromDB(sDB *statedb.StateDB) (*BridgeHubState, error) {
	// load list brigde infos
	listBridgeInfos, err := statedb.GetBridgeHubBridgeInfo(sDB)
	if err != nil {
		return nil, err
	}
	bridgeInfos := map[string]*BridgeInfo{}
	for _, info := range listBridgeInfos {
		pTokens, err := statedb.GetBridgeHubPTokenByBridgeID(sDB, info.BriPubKey())
		if err != nil {
			return nil, err
		}
		pTokenMap := map[common.Hash]*statedb.BridgeHubPTokenState{}
		for _, token := range pTokens {
			pTokenMap[token.PTokenID()] = token
		}

		bridgeInfos[info.BriPubKey()] = &BridgeInfo{
			Info:          info,
			PTokenAmounts: pTokenMap,
		}
	}

	// load param
	param, err := statedb.GetBridgeHubParam(sDB)
	if err != nil {
		return nil, err
	}

	// load staking info
	stakingInfo := map[string]uint64{}
	stakingInfos, err := statedb.GetBridgeStakingInfo(sDB)
	if err != nil {
		return nil, err
	}
	for _, v := range stakingInfos {
		stakingInfo[v.Validator()] = v.StakingAmount()
	}

	// TODO: load more

	return &BridgeHubState{
		stakingInfos: nil,
		bridgeInfos:  bridgeInfos,
		params:       param,
	}, nil
}
