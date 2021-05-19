package rpcservice

import (
	"encoding/json"

	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/metadata"
)

// ============================= Portal v4 ===============================
func (blockService BlockService) GetPortalShieldingRequestStatus(reqTxID string) (*metadata.PortalShieldingRequestStatus, error) {
	stateDB := blockService.BlockChain.GetBeaconBestState().GetBeaconFeatureStateDB()
	data, err := statedb.GetShieldingRequestStatus(stateDB, reqTxID)
	if err != nil {
		return nil, err
	}

	var status metadata.PortalShieldingRequestStatus
	err = json.Unmarshal(data, &status)
	if err != nil {
		return nil, err
	}

	return &status, nil
}

func (blockService BlockService) GetPortalUnshieldingRequestStatus(unshieldID string) (*metadata.PortalUnshieldRequestStatus, error) {
	stateDB := blockService.BlockChain.GetBeaconBestState().GetBeaconFeatureStateDB()
	data, err := statedb.GetPortalUnshieldRequestStatus(stateDB, unshieldID)
	if err != nil {
		return nil, err
	}

	var status metadata.PortalUnshieldRequestStatus
	err = json.Unmarshal(data, &status)
	if err != nil {
		return nil, err
	}

	return &status, nil
}

func (blockService BlockService) GetPortalBatchUnshieldingRequestStatus(batchID string) (*metadata.PortalUnshieldRequestBatchStatus, error) {
	stateDB := blockService.BlockChain.GetBeaconBestState().GetBeaconFeatureStateDB()
	data, err := statedb.GetPortalBatchUnshieldRequestStatus(stateDB, batchID)
	if err != nil {
		return nil, err
	}

	var status metadata.PortalUnshieldRequestBatchStatus
	err = json.Unmarshal(data, &status)
	if err != nil {
		return nil, err
	}

	return &status, nil
}

func (blockService BlockService) GetPortalReqReplacementFeeStatus(reqTxID string) (*metadata.PortalReplacementFeeRequestStatus, error) {
	stateDB := blockService.BlockChain.GetBeaconBestState().GetBeaconFeatureStateDB()
	data, err := statedb.GetPortalUnshieldBatchReplacementRequestStatus(stateDB, reqTxID)
	if err != nil {
		return nil, err
	}

	var status metadata.PortalReplacementFeeRequestStatus
	err = json.Unmarshal(data, &status)
	if err != nil {
		return nil, err
	}

	return &status, nil
}

func (blockService BlockService) GetPortalSubmitConfirmedTxStatus(reqTxID string) (*metadata.PortalSubmitConfirmedTxStatus, error) {
	stateDB := blockService.BlockChain.GetBeaconBestState().GetBeaconFeatureStateDB()
	data, err := statedb.GetPortalSubmitConfirmedTxRequestStatus(stateDB, reqTxID)
	if err != nil {
		return nil, err
	}

	var status metadata.PortalSubmitConfirmedTxStatus
	err = json.Unmarshal(data, &status)
	if err != nil {
		return nil, err
	}

	return &status, nil
}
