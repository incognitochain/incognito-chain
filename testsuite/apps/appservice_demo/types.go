package main

type Key struct {
	PrivateKey         string `json:"private_key"`
	PaymentAddress     string `json:"payment_address"`
	OTAPrivateKey      string `json:"ota_private_key"`
	MiningKey          string `json:"mining_key"`
	MiningPublicKey    string `json:"mining_public_key"`
	ValidatorPublicKey string `json:"validator_public_key"`
}

type Validator struct {
	Key
	HasStakedShard       bool              `json:"has_staked_shard"`
	HasStakedBeacon      bool              `json:"has_staked_beacon"`
	StakeShardFromHeight uint64            `json:"stake_shard_from_height"`
	ExpectActionsIndex   map[string]uint64 `json:"expect_actions_index"`
	ActualActionsIndex   map[string]uint64 `json:"actual_actions_index"`
	Role                 int               `json:"role"`
}

const (
	NormalRole = iota
	ShardCandidateRole
	ShardSyncingRole
	ShardPendingRole
	ShardCommitteeRole
	BeaconWaitingRole
	BeaconPendingRole
	BeaconCommitteeRole
)
