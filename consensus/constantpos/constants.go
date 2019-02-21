package constantpos

const (
	ListenTimeout  = 18   //in s
	PrepareTimeout = 6    //in s
	CommitTimeout  = 6    //in s
	DelayTime      = 1000 // in ms
)

//PBFT
const (
	PBFT_LISTEN  = "listen"
	PBFT_PROPOSE = "propose"
	PBFT_PREPARE = "prepare"
	PBFT_COMMIT  = "commit"
)
