package blockchain

import (
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

// todo
func InitCurrentPortalStateFromDB(
	db database.DatabaseInterface,
	beaconHeight uint64,
) (*CurrentPortalState, error) {

	return &CurrentPortalState{}, nil
}

// todo
func storePortalStateToDB(
	db database.DatabaseInterface,
	beaconHeight uint64,
	currentPortalState *CurrentPortalState,
) error {
	return nil
}
