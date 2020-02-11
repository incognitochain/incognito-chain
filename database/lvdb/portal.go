package lvdb

import "fmt"

type CustodianState struct {
	IncognitoAddress string
	TotalCollateral  uint64
	FreeCollateral   uint64
	HoldingPubTokens map[string]uint64
	RemoteAddresses  map[string]string
}

type PortingRequest struct {
	UniquePortingID string
	TxReqID         string
	TokenID         string
	PorterAddress   string
	Amount          uint64
	Custodians      map[string]uint64
	PortingFee      uint64
}

type RedeemRequest struct {
	UniqueRedeemID        string
	TxReqID               string
	TokenID               string
	RedeemerAddress       string
	RedeemerRemoteAddress string
	Amount                uint64
	Custodians            map[string]uint64
	RedeemFee             uint64
}

func NewCustodianStateKey (beaconHeight uint64, custodianAddress string) string {
	beaconHeightBytes := []byte(fmt.Sprintf("%d-", beaconHeight))
	key := append(CustodianStatePrefix, beaconHeightBytes...)
	key = append(key, []byte(custodianAddress)...)
	return string(key)
}