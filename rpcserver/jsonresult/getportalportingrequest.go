package jsonresult

import (
	metadata2 "github.com/incognitochain/incognito-chain/portal/metadata"
)

type PortalPortingRequest struct {
	PortingRequest metadata2.PortingRequestStatus `json:"PortingRequest"`
}
