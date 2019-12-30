package peerv2

import "time"

// block type
const (
	blockShard         = 0
	crossShard         = 1
	shardToBeacon      = 2
	MaxCallRecvMsgSize = 50 << 20 // 50 MBs per gRPC response
	MaxConnectionRetry = 6        // connect to new highway after 6 failed retries

	RegisterTimestep          = 1 * time.Second  // Re-register to highway
	ReconnectHighwayTimestep  = 10 * time.Second // Check libp2p connection
	UpdateHighwayListTimestep = 30 * time.Minute // RPC to update list of highways every
	RequesterDialTimestep     = 10 * time.Second // Check gRPC connection
	DialTimeout               = 2 * time.Second  // Timeout for dialing's context
)
