package constantbft

const (
	ListenTimeout  = 15  //in s
	PrepareTimeout = 5  //in s
	CommitTimeout  = 10  //in s
	DelayTime      = 50 // in ms
)

const (
	BFT_LISTEN  = "listen"
	BFT_PROPOSE = "propose"
	BFT_PREPARE = "prepare"
	BFT_COMMIT  = "commit"
)
