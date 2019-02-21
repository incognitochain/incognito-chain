package constantpos

const (
	ListenTimeout  = 30 //in second
	PrepareTimeout = 6
	CommitTimeout  = 6
)

//PBFT
const (
	PBFT_LISTEN  = "listen"
	PBFT_PROPOSE = "propose"
	PBFT_PREPARE = "prepare"
	PBFT_COMMIT  = "commit"
)
