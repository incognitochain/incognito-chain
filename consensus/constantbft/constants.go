package constantbft

import "time"

const (
	ListenTimeout       = 5 * time.Second        //in s
	PrepareTimeout      = 3 * time.Second        //in s
	CommitTimeout       = 5 * time.Second        //in s
	MaxNetworkDelayTime = 150 * time.Millisecond // in ms
	MaxNormalRetryTime  = 2
)

const (
	BFT_LISTEN  = "listen"
	BFT_PROPOSE = "propose"
	BFT_PREPARE = "prepare"
	BFT_COMMIT  = "commit"
)
