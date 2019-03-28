package constantbft

const (
	ListenTimeout  = 5   //in s
	PrepareTimeout = 4   //in s
	CommitTimeout  = 4   //in s
	DelayTime      = 600 // in ms
)

const (
	PBFT_LISTEN  = "listen"
	PBFT_PROPOSE = "propose"
	PBFT_PREPARE = "prepare"
	PBFT_COMMIT  = "commit"
)
