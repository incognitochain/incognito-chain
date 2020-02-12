package blockchain

import (
	"encoding/json"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/database"
	"github.com/incognitochain/incognito-chain/database/lvdb"
)

type CurrentPortalState struct {
	CustodianPoolState map[string]*lvdb.CustodianState // key : beaconHeight || custodian_address
	PortingRequests    map[string]*lvdb.PortingRequest // key : beaconHeight || UniquePortingID
	RedeemRequests     map[string]*lvdb.RedeemRequest  // key : beaconHeight || UniqueRedeemID
}

func NewCustodianState(
	incognitoAddress string,
	totalColl uint64,
	freeColl uint64,
	holdingPubTokens map[string]uint64,
	remoteAddresses map[string]string,
) (*lvdb.CustodianState, error) {
	return &lvdb.CustodianState{
		IncognitoAddress: incognitoAddress,
		TotalCollateral:  totalColl,
		FreeCollateral:   freeColl,
		HoldingPubTokens: holdingPubTokens,
		RemoteAddresses:  remoteAddresses,
	}, nil
}

func NewPortingRequestState(
	uniquePortingID string,
	txReqID common.Hash,
	tokenID string,
	porterAddress string,
	amount uint64,
	custodians map[string]uint64,
	portingFee uint64,
) (*lvdb.PortingRequest, error) {
	return &lvdb.PortingRequest{
		UniquePortingID: uniquePortingID,
		TxReqID:         txReqID,
		TokenID:         tokenID,
		PorterAddress:   porterAddress,
		Amount:          amount,
		Custodians:      custodians,
		PortingFee:      portingFee,
	}, nil
}

// todo
func InitCurrentPortalStateFromDB(
	db database.DatabaseInterface,
	beaconHeight uint64,
) (*CurrentPortalState, error) {
	custodianPoolState, err := getCustodianPoolState(db, beaconHeight)
	if err != nil {
		return nil, err
	}
	portingRequestsState, err := getPortingRequestsState(db, beaconHeight)
	if err != nil {
		return nil, err
	}
	redeemRequestsState, err := getRedeemRequestsState(db, beaconHeight)
	if err != nil {
		return nil, err
	}

	return &CurrentPortalState{
		CustodianPoolState: custodianPoolState,
		PortingRequests:    portingRequestsState,
		RedeemRequests:     redeemRequestsState,
	}, nil
}

// todo
func storePortalStateToDB(
	db database.DatabaseInterface,
	beaconHeight uint64,
	currentPortalState *CurrentPortalState,
) error {
	return nil
}


func getCustodianPoolState(
	db database.DatabaseInterface,
	beaconHeight uint64,
) (map[string]*lvdb.CustodianState, error) {
		custodianPoolState := make(map[string]*lvdb.CustodianState)
		custodianPoolStateKeysBytes, custodianPoolStateValuesBytes, err := db.GetAllRecordsPortalByPrefix(beaconHeight, lvdb.CustodianStatePrefix)
		if err != nil {
			return nil, err
		}
		for idx, custodianPoolStateKeyBytes := range custodianPoolStateKeysBytes {
			var custodianState lvdb.CustodianState
			err = json.Unmarshal(custodianPoolStateValuesBytes[idx], &custodianState)
			if err != nil {
				return nil, err
			}
			custodianPoolState[string(custodianPoolStateKeyBytes)] = &custodianState
		}
		return custodianPoolState, nil
	}

func getPortingRequestsState(
	db database.DatabaseInterface,
	beaconHeight uint64,
) (map[string]*lvdb.PortingRequest, error) {
	portingRequestState := make(map[string]*lvdb.PortingRequest)
	portingRequestStateKeysBytes, portingRequestStateValuesBytes, err := db.GetAllRecordsPortalByPrefix(beaconHeight, lvdb.PortalPortingRequestsPrefix)
	if err != nil {
		return nil, err
	}
	for idx, portingRequestStateKeyBytes := range portingRequestStateKeysBytes {
		var portingRequest lvdb.PortingRequest
		err = json.Unmarshal(portingRequestStateValuesBytes[idx], &portingRequest)
		if err != nil {
			return nil, err
		}

		portingRequestState[string(portingRequestStateKeyBytes)] = &portingRequest
	}
	return portingRequestState, nil
}

func getRedeemRequestsState(
	db database.DatabaseInterface,
	beaconHeight uint64,
) (map[string]*lvdb.RedeemRequest, error) {
	redeemRequestState := make(map[string]*lvdb.RedeemRequest)
	redeemRequestStateKeysBytes, redeemRequestStateValuesBytes, err := db.GetAllRecordsPortalByPrefix(beaconHeight, lvdb.PortalRedeemRequestsPrefix)
	if err != nil {
		return nil, err
	}
	for idx, portingRequestStateKeyBytes := range redeemRequestStateKeysBytes {
		var redeemRequest lvdb.RedeemRequest
		err = json.Unmarshal(redeemRequestStateValuesBytes[idx], &redeemRequest)
		if err != nil {
			return nil, err
		}

		redeemRequestState[string(portingRequestStateKeyBytes)] = &redeemRequest
	}
	return redeemRequestState, nil
}

