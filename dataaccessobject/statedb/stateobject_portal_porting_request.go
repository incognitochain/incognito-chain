package statedb

import "github.com/incognitochain/incognito-chain/common"


type MatchingPortingCustodianDetail struct {
	IncAddress             string
	RemoteAddress          string
	Amount                 uint64
	LockedAmountCollateral uint64
	RemainCollateral       uint64
}

type PortingRequest struct {
	uniquePortingID string
	txReqID         common.Hash
	tokenID         string
	porterAddress   string
	amount          uint64
	custodians      []*MatchingPortingCustodianDetail
	portingFee      uint64
	status          int
	beaconHeight    uint64
}

func NewPortingRequest(uniquePortingID string, txReqID common.Hash, tokenID string, porterAddress string, amount uint64, custodians []*MatchingPortingCustodianDetail, portingFee uint64, status int, beaconHeight uint64) *PortingRequest {
	return &PortingRequest{uniquePortingID: uniquePortingID, txReqID: txReqID, tokenID: tokenID, porterAddress: porterAddress, amount: amount, custodians: custodians, portingFee: portingFee, status: status, beaconHeight: beaconHeight}
}

func GeneratePortingRequestObjectKey(portingRequestId string) common.Hash {
	//	key := append(PortalPortingRequestsPrefix, []byte(uniquePortingID)...)
	//	return string(key) //prefix + uniqueId
	return common.Hash{}
}

