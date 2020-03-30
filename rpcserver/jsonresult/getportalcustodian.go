package jsonresult

import "github.com/incognitochain/incognito-chain/metadata"

type  PortalCustodianWithdrawRequest struct {
	CustodianWithdrawRequest metadata.CustodianWithdrawRequestStatus `json:"CustodianWithdraw"`
}