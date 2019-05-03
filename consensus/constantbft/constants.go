package constantbft

const (
	ListenTimeout  = 15  //in s
	PrepareTimeout = 10  //in s
	CommitTimeout  = 5  //in s
	DelayTime      = 50 // in ms
)

const (
	PBFT_LISTEN  = "listen"
	PBFT_PROPOSE = "propose"
	PBFT_PREPARE = "prepare"
	PBFT_COMMIT  = "commit"
)
