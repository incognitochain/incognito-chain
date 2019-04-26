package constantbft

const (
	ListenTimeout  = 3  //in s
	PrepareTimeout = 2  //in s
	CommitTimeout  = 2  //in s
	DelayTime      = 50 // in ms
)

const (
	BFT_LISTEN  = "listen"
	BFT_PROPOSE = "propose"
	BFT_PREPARE = "prepare"
	BFT_COMMIT  = "commit"
)
