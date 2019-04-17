package constantbft

const (
	ListenTimeout  = 3   //in s
	PrepareTimeout = 2   //in s
	CommitTimeout  = 2   //in s
	DelayTime      = 100 // in ms
)

const (
	PBFT_LISTEN  = "listen"
	PBFT_PROPOSE = "propose"
	PBFT_PREPARE = "prepare"
	PBFT_COMMIT  = "commit"
)
