package bridgehub

import (
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
		pTokens, err := statedb.GetBridgeHubPTokenByBridgeID(sDB, info.BridgeID())
		if err != nil {
			return nil, err
		}
		pTokenMap := map[string]*statedb.BridgeHubPTokenState{}
		for _, token := range pTokens {
			pTokenMap[token.PTokenID()] = token
		}

		bridgeInfos[info.BridgeID()] = &BridgeInfo{
			Info:          info,
			PTokenAmounts: pTokenMap,
		}
	}

	// load param
	param, err := statedb.GetBridgeHubParam(sDB)
	if err != nil {
		return nil, err
	}

	// TODO: load more

	return &BridgeHubState{
		stakingInfos: nil,
		bridgeInfos:  bridgeInfos,
		params:       param,
	}, nil
}
