package constantbft

const (
	ListenTimeout      = 15  //in s
	PrepareTimeout     = 15  //in s
	CommitTimeout      = 15  //in s
	DelayTime          = 0 // in ms
	MaxNormalRetryTime = 2
)

const (
	BFT_LISTEN  = "listen"
	BFT_PROPOSE = "propose"
	BFT_PREPARE = "prepare"
	BFT_COMMIT  = "commit"
)
