package constantbft

const (
	ListenTimeout  = 20  //in s
	PrepareTimeout = 20  //in s
	CommitTimeout  = 20  //in s
	DelayTime      = 50 // in ms
)

const (
	BFT_LISTEN  = "listen"
	BFT_PROPOSE = "propose"
	BFT_PREPARE = "prepare"
	BFT_COMMIT  = "commit"
)
