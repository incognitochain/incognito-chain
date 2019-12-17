package statedb

// version
const (
	defaultVersion = 0
)

// Object type
const (
	TestObjectType = iota
	SerialNumberObjectType
	AllShardCommitteeObjectType
	CommitteeObjectType
	CommitteeRewardObjectType
	RewardRequestObjectType
	BlackListProducerObjectType
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
