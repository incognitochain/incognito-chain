package rpcservice

import (
	"encoding/json"

	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	metadataPdexv3 "github.com/incognitochain/incognito-chain/metadata/pdexv3"
)

func (blockService BlockService) GetPdexv3ParamsModifyingRequestStatus(reqTxID string) (*metadataPdexv3.ParamsModifyingRequestStatus, error) {
	stateDB := blockService.BlockChain.GetBeaconBestState().GetBeaconFeatureStateDB()
	data, err := statedb.GetPdexv3Status(
		stateDB,
		statedb.Pdexv3ParamsModifyingStatusPrefix(),
		[]byte(reqTxID),
	)
	if err != nil {
		return nil, err
	}

	var status metadataPdexv3.ParamsModifyingRequestStatus
	err = json.Unmarshal(data, &status)
	if err != nil {
		return nil, err
	}

	return &status, nil
}
