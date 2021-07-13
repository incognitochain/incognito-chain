package pdex

import "github.com/incognitochain/incognito-chain/dataaccessobject/statedb"

type stateProcessorV2 struct {
	stateProcessorBase
}

func (sp *stateProcessorV2) modifyParams(
	stateDB *statedb.StateDB,
	beaconHeight uint64,
	inst []string,
	params Params,
) (Params, error) {
	return params, nil
}
