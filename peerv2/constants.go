package peerv2

// block type
const (
	blockShard         = 0
	crossShard         = 1
	shardToBeacon      = 2
	MaxCallRecvMsgSize = 50 << 20 // 50 MBs per gRPC response
	MaxConnectionRetry = 6        // connect to new highway after 6 failed retries
)
