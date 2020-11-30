package jsonresult

import (
	metadata2 "github.com/incognitochain/incognito-chain/portal/metadata"
)

type PortalCustodianWithdrawRequest struct {
	CustodianWithdrawRequest metadata2.CustodianWithdrawRequestStatus `json:"CustodianWithdraw"`
}
