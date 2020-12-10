package metadata

import (
	"github.com/incognitochain/incognito-chain/basemeta"
)

// don't have PortalExpiredWaitingPortingReq because this action only occurs on beacon automatically.
//type PortalExpiredWaitingPortingReq struct {
//	basemeta.MetadataBase
//	UniquePortingID      string
//	ExpiredByLiquidation bool
//}

// PortalExpiredWaitingPortingReqContent - Beacon builds a new instruction with this content after detecting user haven't sent public token to custodian
// It will be appended to beaconBlock
type PortalExpiredWaitingPortingReqContent struct {
	basemeta.MetadataBase
	UniquePortingID      string
	ExpiredByLiquidation bool
	ShardID              byte
}

// PortalExpiredWaitingPortingReqStatus - Beacon tracks status of custodian liquidation into db
type PortalExpiredWaitingPortingReqStatus struct {
	//Status               byte		// dont need to store this status
	UniquePortingID      string
	ShardID              byte
	ExpiredByLiquidation bool
	ExpiredBeaconHeight  uint64
}