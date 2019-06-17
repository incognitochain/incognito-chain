package constantbft

import "time"

const (
	ListenTimeout       = 5 * time.Second        //in s
	AgreeTimeout        = 3 * time.Second        //in s
	CommitTimeout       = 3 * time.Second        //in s
	MaxNetworkDelayTime = 150 * time.Millisecond // in ms
	MaxNormalRetryTime  = 2
)

const (
	BFT_LISTEN  = "listen"
	BFT_PROPOSE = "propose"
	BFT_AGREE   = "agree"
	BFT_COMMIT  = "commit"
)
