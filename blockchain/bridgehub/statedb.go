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
		vaultAddrs, err := statedb.GetBridgeHubNetworkInfoByBridgeID(sDB, info.BriPubKey())
		if err != nil {
			return nil, err
		}
		for _, networkInfo := range vaultAddrs {
			pTokens, err := statedb.GetBridgeHubPToken(sDB, info.BriPubKey(), networkInfo.NetworkId())
			if err != nil {
				return nil, err
			}
			bridgeInfos[info.BriPubKey()].NetworkInfo[networkInfo.NetworkId()] = &BridgeNetwork{
				networkId:    networkInfo.NetworkId(),
				vaultAddress: networkInfo.VaultAddress(),
				pTokens:      pTokens,
			}
		}
	}

	// load param
	param, err := statedb.GetBridgeHubParam(sDB)
	if err != nil {
		return nil, err
	}

	// load staking info
	stakingInfo := map[string]*statedb.BridgeStakingInfoState{}
	stakingInfos, err := statedb.GetBridgeStakingInfo(sDB)
	if err != nil {
		return nil, err
	}
	for _, v := range stakingInfos {
		stakingInfo[v.BridgePubKey()] = v
	}

	// TODO: load more

	return &BridgeHubState{
		stakingInfos: nil,
		bridgeInfos:  bridgeInfos,
		params:       param,
	}, nil
}
