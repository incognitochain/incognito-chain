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
	ActionsIndex         map[string]Action `json:"actions_index"`
	Role                 int               `json:"role"`
}

type Action struct {
	Height uint64 `json:"height"`
	TxHash string `json:"tx_hash"`
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
	BeaconLockingRole
)
