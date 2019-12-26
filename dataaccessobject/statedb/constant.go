package statedb

// version
const (
	defaultVersion = 0
)

// Object type
const (
	TestObjectType = iota
	CommitteeObjectType
	CommitteeRewardObjectType
	RewardRequestObjectType
	BlackListProducerObjectType
	SerialNumberObjectType
	CommitmentObjectType
	CommitmentIndexObjectType
	CommitmentLengthObjectType
	SNDerivatorObjectType
	OutputCoinObjectType
	TokenObjectType
	WaitingPDEContributionObjectType
	PDEPoolPairObjectType
	PDEShareObjectType
	PDEStatusObjectType
)

// Prefix length
const (
	prefixHashKeyLength = 12
	prefixKeyLength     = 20
)

// Committee Role
const (
	NextEpochCandidate = iota
	CurrentEpochCandidate
	SubstituteValidator
	CurrentValidator
)
const (
	BeaconShardID    = -1
	CandidateShardID = -2
)

// PDE Track Status type
const (
	WaitingContributionStatus = iota
	TradeStatus
	WithdrawStatus
)
