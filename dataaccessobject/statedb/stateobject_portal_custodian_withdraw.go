package statedb

import "github.com/incognitochain/incognito-chain/common"

type CustodianWithdrawRequest struct {
	paymentAddress                string
	amount                        uint64
	status                        int
	remainCustodianFreeCollateral uint64
}

func NewCustodianWithdrawRequestWithKey(paymentAddress string, amount uint64, status int, remainCustodianFreeCollateral uint64) *CustodianWithdrawRequest {
	return &CustodianWithdrawRequest{paymentAddress: paymentAddress, amount: amount, status: status, remainCustodianFreeCollateral: remainCustodianFreeCollateral}
}


func NewCustodianWithdrawRequest() *CustodianWithdrawRequest {
	return &CustodianWithdrawRequest{}
}


func GenerateCustodianWithdrawObjectKey(key string) common.Hash {
	//	key := append(PortalPortingRequestsPrefix, []byte(uniquePortingID)...)
	//	return string(key) //prefix + uniqueId
	return common.Hash{}
}