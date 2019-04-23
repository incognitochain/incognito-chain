package constantbft

const (
	ListenTimeout  = 2  //in s
	PrepareTimeout = 4  //in s
	CommitTimeout  = 2  //in s
	DelayTime      = 50 // in ms
)

const (
	PBFT_LISTEN  = "listen"
	PBFT_PROPOSE = "propose"
	PBFT_PREPARE = "prepare"
	PBFT_COMMIT  = "commit"
)
