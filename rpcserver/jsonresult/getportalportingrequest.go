package jsonresult

import "github.com/incognitochain/incognito-chain/database/lvdb"

type PortalPortingRequest struct {
	PortingRequest map[string]*lvdb.PortingRequest `json:"PortingRequest"`
}
