package blockchain

import "github.com/incognitochain/incognito-chain/database"

type CustodianState struct {
	//IncognitoAddress string
	TotalCollateral  uint64
	FreeCollateral   uint64
	HoldingPubTokens map[string]uint64
}

type PortingRequest struct {
	UniquePortingID string
	TxReqID         string
	TokenID         string
	PorterAddress   string
	Amount          uint64
	Custodians      map[string]uint64
}

type RedeemRequest struct {
	UniqueRedeemID        string
	TxReqID               string
	TokenID               string
	RedeemerAddress       string
	RedeemerRemoteAddress string
	Amount                uint64
	Custodians            map[string]uint64
}

type CurrentPortalState struct {
	CustodianPoolState map[string]*CustodianState // custodian_address: CustodianState
	PortingRequests    []*PortingRequest
	RedeemRequests     []*RedeemRequest
}

func NewCustodianState(totalColl uint64, freeColl uint64, holdingPubTokens map[string]uint64) (*CustodianState, error) {
	return &CustodianState{
		TotalCollateral:  totalColl,
		FreeCollateral:   freeColl,
		HoldingPubTokens: holdingPubTokens,
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
