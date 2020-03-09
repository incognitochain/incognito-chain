package jsonresult

import "github.com/incognitochain/incognito-chain/database/lvdb"

type  PortalCustodianWithdrawRequest struct {
	CustodianWithdrawRequest lvdb.CustodianWithdrawRequest `json:"CustodianWithdraw"`
}