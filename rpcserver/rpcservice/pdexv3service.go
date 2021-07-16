package rpcservice

import (
	"encoding/json"

	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/metadata"
)

func (blockService BlockService) GetPDexV3ParamsModifyingRequestStatus(reqTxID string) (*metadata.PDexV3ParamsModifyingRequestStatus, error) {
	stateDB := blockService.BlockChain.GetBeaconBestState().GetBeaconFeatureStateDB()
	data, err := statedb.GetPDexV3Status(
		stateDB,
		statedb.PDexV3ParamsModifyingStatusPrefix(),
		[]byte(reqTxID),
	)
	if err != nil {
		return nil, err
	}

	var status metadata.PDexV3ParamsModifyingRequestStatus
	err = json.Unmarshal(data, &status)
	if err != nil {
		return nil, err
	}

	return &status, nil
}
