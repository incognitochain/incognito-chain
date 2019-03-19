package constantbft

const (
	ListenTimeout  = 6    //in s
	PrepareTimeout = 5    //in s
	CommitTimeout  = 5    //in s
	DelayTime      = 1000 // in ms
)

const (
	PBFT_LISTEN  = "listen"
	PBFT_PROPOSE = "propose"
	PBFT_PREPARE = "prepare"
	PBFT_COMMIT  = "commit"
)
