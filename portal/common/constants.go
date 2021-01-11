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

	// type of action - used to determine next action
	PortalSubmitProofPorting = 0
	PortalSubmitProofRedeem  = 1

	// status of portal request - used to append to beacon instructions
	PortalRequestAcceptedChainStatus = "accepted"
	PortalRequestRejectedChainStatus = "rejected"
	PortalRequestRefundChainStatus   = "refund" // beacon reject portal request and refund PRV/PToken to requester

	PortalProducerInstSuccessChainStatus = "success"
	PortalProducerInstFailedChainStatus  = "failed"

	// cancel redeem request by liquidation
	PortalRedeemReqCancelledByLiquidationChainStatus = "cancelled"

	// Relaying header
	RelayingHeaderRejectedChainStatus    = "rejected"
	RelayingHeaderConsideringChainStatus = "considering"

	// Unlock over rate collaterals
	PortalCusUnlockOverRateCollateralsRejectedChainStatus = "rejected"
	PortalCusUnlockOverRateCollateralsAcceptedChainStatus = "accepted"

	PortalUnlockOverRateCollateralsAcceptedStatus = 1
	PortalUnlockOverRateCollateralsRejectedStatus = 1
)

const PortalBTCIDStr = "ef5947f70ead81a76a53c7c8b7317dd5245510c665d3a13921dc9a581188728b"
const PortalBNBIDStr = "6abd698ea7ddd1f98b1ecaaddab5db0453b8363ff092f0d8d7d4c6b1155fb693"

var PortalSupportedIncTokenIDs = []string{
	PortalBTCIDStr, // pBTC
	PortalBNBIDStr, // pBNB
}

var PortalTokenIDsSupportedMultiSig = []string{
	PortalBTCIDStr, // pBTC
}

const ETHChainName = "eth"
const BeaconKey1 = "023470707c011796b29b352f9a5d3bed5bd601af64aa053ddcd7e50820313fe5d5"
const BeaconKey2 = "034df839573d5b81dfd1107df5e2a5b85b0d629a8144640177b2d4284521603bf7"
