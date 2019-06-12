package constantbft

const (
	ListenTimeout      = 8  //in s
	PrepareTimeout     = 8  //in s
	CommitTimeout      = 8  //in s
	DelayTime          = 50 // in ms
	MaxNormalRetryTime = 2
)

const (
	BFT_LISTEN  = "listen"
	BFT_PROPOSE = "propose"
	BFT_PREPARE = "prepare"
	BFT_COMMIT  = "commit"
)
