package config

import (
	"time"
)

//Param for all variables in incognito node process
type Param struct {
	Name                             string             `yaml:"name" description:"Name defines a human-readable identifier for the network" `
	Net                              uint32             `description:"Net defines the magic bytes used to identify the network"`
	GenesisParam                     *genesisParam      `yaml:"genesis_param" description:"genesis params"`
	CommitteeSize                    committeeSize      `yaml:"committee_size"`
	BlockTime                        blockTime          `yaml:"block_time"`
	StakingAmountShard               uint64             `yaml:"staking_amount_shard"`
	ActiveShards                     int                `yaml:"active_shards"`
	BasicReward                      uint64             `yaml:"basic_reward"`
	EpochParam                       epochParam         `yaml:"epoch_param"`
	EthContractAddressStr            string             `yaml:"eth_contract_address" description:"smart contract of ETH for bridge"`
	IncognitoDAOAddress              string             `yaml:"dao_address"`
	CentralizedWebsitePaymentAddress string             `yaml:"centralized_website_payment_address" description:"centralized website's pubkey"`
	SwapCommitteeParam               swapCommitteeParam `yaml:"swap_committee_param"`
	ConsensusParam                   consensusParam     `yaml:"consensus_param"`
	BeaconHeightBreakPointBurnAddr   uint64             `yaml:"beacon_height_break_point_burn_addr"`
	ReplaceStakingTxHeight           uint64             `yaml:"replace_staking_tx_height"`
	ETHRemoveBridgeSigEpoch          uint64             `yaml:"eth_remove_bridge_sig_epoch"`
	BCHeightBreakPointNewZKP         uint64             `yaml:"bc_height_break_point_new_zkp"`
	EnableFeatureFlags               map[int]uint64     `yaml:"enable_feature_flags" description:"featureFlag: epoch number - since that time, the feature will be enabled; 0 - disabled feature"`
	PortalParam                      PortalParam        `yaml:"portal_param"`
	IsBackup                         bool
}

type genesisParam struct {
	InitialIncognito                            []string `yaml:"initial_incognito" description:"init tx for genesis block"`
	FeePerTxKb                                  uint64   `yaml:"fee_per_tx_kb" description:"fee per tx calculate by kb"`
	ConsensusAlgorithm                          string   `yaml:"consensus_algorithm"`
	PreSelectBeaconNodeSerializedPubkey         []string
	SelectBeaconNodeSerializedPubkeyV2          map[uint64][]string
	PreSelectBeaconNodeSerializedPaymentAddress []string
	SelectBeaconNodeSerializedPaymentAddressV2  map[uint64][]string
	PreSelectShardNodeSerializedPubkey          []string
	SelectShardNodeSerializedPubkeyV2           map[uint64][]string
	PreSelectShardNodeSerializedPaymentAddress  []string
	SelectShardNodeSerializedPaymentAddressV2   map[uint64][]string
	BeaconBlock                                 string `yaml:"beacon_block" description:"GenesisBlock defines the first block of the chain"`
	ShardBlock                                  string `yaml:"shard_block" description:"GenesisBlock defines the first block of the chain"`
}

type committeeSize struct {
	MaxShardCommitteeSize   int `yaml:"max_shard_committee_size"`
	MinShardCommitteeSize   int `yaml:"min_shard_committee_size"`
	MaxBeaconCommitteeSize  int `yaml:"max_beacon_committee_size"`
	MinBeaconCommitteeSize  int `yaml:"min_beacon_committee_size"`
	InitShardCommitteeSize  int `yaml:"init_shard_committee_size"`
	InitBeaconCommitteeSize int `yaml:"init_beacon_committee_size"`
}

type blockTime struct {
	MinShardBlockInterval        time.Duration `yaml:"min_shard_block_interval"`
	MaxShardBlockCreation        time.Duration `yaml:"max_shard_block_creation"`
	MinBeaconBlockInterval       time.Duration `yaml:"min_beacon_block_interval"`
	MaxBeaconBlockCreation       time.Duration `yaml:"min_beacon_block_creation"`
	NumberOfFixedBlockValidators int           `yaml:"number_of_fixed_shard_block_validators"`
}

type epochParam struct {
	NumberOfBlockInEpoch   uint64 `yaml:"number_of_block_in_epoch"`
	NumberOfBlockInEpochV2 uint64 `yaml:"number_of_block_in_epoch_v2"`
	EpochV2BreakPoint      uint64 `yaml:"epoch_v2_break_point"`
	RandomTime             uint64 `yaml:"random_time"`
	RandomTimeV2           uint64 `yaml:"random_time_v2"`
}

type swapCommitteeParam struct {
	Offset       int `yaml:"offset" description:"default offset for swap policy, is used for cases that good producers length is less than max committee size"`
	SwapOffset   int `yaml:"swap_offset" description:"is used for case that good producers length is equal to max committee size"`
	AssignOffset int `yaml:"assign_offset"`
}

type consensusParam struct {
	ConsensusV2Epoch          uint64   `yaml:"consensus_v2_epoch"`
	StakingFlowV2Height       uint64   `yaml:"staking_flow_v2_height"`
	EnableSlashingHeight      uint64   `yaml:"enable_slashing_height"`
	Timeslot                  uint64   `yaml:"timeslot"`
	EpochBreakPointSwapNewKey []uint64 `yaml:"epoch_break_point_swap_new_key"`
}
