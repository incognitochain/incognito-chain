package jsonresult

import "github.com/incognitochain/incognito-chain/metadata"

type PortalPortingRequest struct {
	PortingRequest metadata.PortingRequestStatus `json:"PortingRequest"`
}
