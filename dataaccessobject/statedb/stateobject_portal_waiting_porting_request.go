package statedb

import "github.com/incognitochain/incognito-chain/common"


type MatchingPortingCustodianDetail struct {
	IncAddress             string
	RemoteAddress          string
	Amount                 uint64
	LockedAmountCollateral uint64
	RemainCollateral       uint64
}

type WaitingPortingRequest struct {
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

func (w *WaitingPortingRequest) BeaconHeight() uint64 {
	return w.beaconHeight
}

func (w *WaitingPortingRequest) SetBeaconHeight(beaconHeight uint64) {
	w.beaconHeight = beaconHeight
}

func (w *WaitingPortingRequest) Status() int {
	return w.status
}

func (w *WaitingPortingRequest) SetStatus(status int) {
	w.status = status
}

func (w *WaitingPortingRequest) PortingFee() uint64 {
	return w.portingFee
}

func (w *WaitingPortingRequest) SetPortingFee(portingFee uint64) {
	w.portingFee = portingFee
}

func (w *WaitingPortingRequest) Custodians() []*MatchingPortingCustodianDetail {
	return w.custodians
}

func (w *WaitingPortingRequest) SetCustodians(custodians []*MatchingPortingCustodianDetail) {
	w.custodians = custodians
}

func (w *WaitingPortingRequest) Amount() uint64 {
	return w.amount
}

func (w *WaitingPortingRequest) SetAmount(amount uint64) {
	w.amount = amount
}

func (w *WaitingPortingRequest) PorterAddress() string {
	return w.porterAddress
}

func (w *WaitingPortingRequest) SetPorterAddress(porterAddress string) {
	w.porterAddress = porterAddress
}

func (w *WaitingPortingRequest) TokenID() string {
	return w.tokenID
}

func (w *WaitingPortingRequest) SetTokenID(tokenID string) {
	w.tokenID = tokenID
}

func (w *WaitingPortingRequest) TxReqID() common.Hash {
	return w.txReqID
}

func (w *WaitingPortingRequest) SetTxReqID(txReqID common.Hash) {
	w.txReqID = txReqID
}

func (w *WaitingPortingRequest) UniquePortingID() string {
	return w.uniquePortingID
}

func (w *WaitingPortingRequest) SetUniquePortingID(uniquePortingID string) {
	w.uniquePortingID = uniquePortingID
}

func NewWaitingPortingRequest() *WaitingPortingRequest {
	return &WaitingPortingRequest{}
}

func NewWaitingPortingRequestWithValue(uniquePortingID string, txReqID common.Hash, tokenID string, porterAddress string, amount uint64, custodians []*MatchingPortingCustodianDetail, portingFee uint64, status int, beaconHeight uint64) *WaitingPortingRequest {
	return &WaitingPortingRequest{uniquePortingID: uniquePortingID, txReqID: txReqID, tokenID: tokenID, porterAddress: porterAddress, amount: amount, custodians: custodians, portingFee: portingFee, status: status, beaconHeight: beaconHeight}
}

func GeneratePortalWaitingPortingRequestObjectKey(portingRequestId string) common.Hash {
	//	key := append(PortalPortingRequestsPrefix, []byte(uniquePortingID)...)
	//	return string(key) //prefix + uniqueId
	return common.Hash{}
}



