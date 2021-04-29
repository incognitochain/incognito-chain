package common

const (
	// status of porting processing - used to store db
	PortalPortingReqSuccessStatus    = 1
	PortalPortingReqWaitingStatus    = 2
	PortalPortingReqExpiredStatus    = 3
	PortalPortingReqLiquidatedStatus = 4

	// status of redeem processing  - used to store db
	PortalRedeemReqSuccessStatus                = 1
	PortalRedeemReqWaitingStatus                = 2
	PortalRedeemReqMatchedStatus                = 3
	PortalRedeemReqLiquidatedStatus             = 4
	PortalRedeemReqCancelledByLiquidationStatus = 5

	// status of portal request - used to store db
	PortalRequestRejectedStatus = 0
	PortalRequestAcceptedStatus = 1

	// status of portal request - used to append to beacon instructions
	PortalRequestAcceptedChainStatus = "accepted"
	PortalRequestRejectedChainStatus = "rejected"
	PortalRequestRefundChainStatus   = "refund" // beacon reject portal request and refund PRV/PToken to requester

	PortalProducerInstSuccessChainStatus = "success"
	PortalProducerInstFailedChainStatus  = "failed"

	// cancel redeem request by liquidation
	PortalRedeemReqCancelledByLiquidationChainStatus = "cancelled"
)

const PortalBTCIDStr = "ef5947f70ead81a76a53c7c8b7317dd5245510c665d3a13921dc9a581188728b"
const PortalBNBIDStr = "6abd698ea7ddd1f98b1ecaaddab5db0453b8363ff092f0d8d7d4c6b1155fb693"

var PortalSupportedIncTokenIDs = []string{
	PortalBTCIDStr, // pBTC
	PortalBNBIDStr, // pBNB
}

const ETHChainName = "eth"
