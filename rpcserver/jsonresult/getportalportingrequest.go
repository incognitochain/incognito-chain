package jsonresult

import "github.com/incognitochain/incognito-chain/database/lvdb"

type PortalPortingRequest struct {
	PortingRequest lvdb.PortingRequest `json:"PortingRequest"`
}
