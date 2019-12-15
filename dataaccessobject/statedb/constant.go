package statedb

// Object type
const (
	TestObjectType = iota
	SerialNumberObjectType
	AllShardCommitteeObjectType
	CommitteeObjectType
	RewardReceiverObjectType
	AutoStakingObjectType
)

// Prefix length
const (
	prefixHashKeyLength = 12
	prefixKeyLength     = 20
)
