package netsync

import "time"

const (
	workers             = 5
	MsgLiveTime         = 40 * time.Second  // in second
	MsgsCleanupInterval = 300 * time.Second //in second
)
