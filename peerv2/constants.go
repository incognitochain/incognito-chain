package peerv2

import "time"

// block type
const (
	blockShard                = 0
	crossShard                = 1
	shardToBeacon             = 2
	MaxCallRecvMsgSize        = 50 << 20 // 50 MBs per gRPC response
	RequesterKeepaliveTime    = 10 * time.Minute
	RequesterKeepaliveTimeout = 20 * time.Second
)
