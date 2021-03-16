package common

const PortalBTCIDStr = "ef5947f70ead81a76a53c7c8b7317dd5245510c665d3a13921dc9a581188728b"

var PortalV4SupportedIncTokenIDs = []string{
	PortalBTCIDStr, // pBTC
}

const (
	// status of portal v4 request - used to append to beacon instructions
	PortalV4RequestAcceptedChainStatus = "accepted"
	PortalV4RequestRejectedChainStatus = "rejected"
	PortalV4RequestRefundedChainStatus = "refunded"

	// status of portal v4 request - used to store db
	PortalV4RequestRejectedStatus = 0
	PortalV4RequestAcceptedStatus = 1

	/// status of unshield processing - used to store db
	PortalUnshieldReqWaitingStatus   = 0
	PortalUnshieldReqProcessedStatus = 1
	PortalUnshieldReqCompletedStatus = 2
	PortalUnshieldReqRefundedStatus  = 3

	// status of batching unshield processing by batchID - used to store db
	PortalBatchUnshieldReqProcessedStatus = 0
	PortalBatchUnshieldReqCompletedStatus = 1
)
