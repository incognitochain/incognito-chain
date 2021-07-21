package rpcservice

import (
	"encoding/json"

	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	metadataPDexV3 "github.com/incognitochain/incognito-chain/metadata/pdexv3"
)

func (blockService BlockService) GetPDexV3ParamsModifyingRequestStatus(reqTxID string) (*metadataPDexV3.ParamsModifyingRequestStatus, error) {
	stateDB := blockService.BlockChain.GetBeaconBestState().GetBeaconFeatureStateDB()
	data, err := statedb.GetPDexV3Status(
		stateDB,
		statedb.PDexV3ParamsModifyingStatusPrefix(),
		[]byte(reqTxID),
	)
	if err != nil {
		return nil, err
	}

	var status metadataPDexV3.ParamsModifyingRequestStatus
	err = json.Unmarshal(data, &status)
	if err != nil {
		return nil, err
	}

	return &status, nil
}
