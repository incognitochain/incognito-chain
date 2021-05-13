package config

import (
	"path/filepath"
	"strconv"
	"time"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/spf13/viper"
)

var p *param

func Param() *param {
	return p
}

//param for all variables in incognito node process
type param struct {
	Name                             string             `mapstructure:"name" description:"Name defines a human-readable identifier for the network" `
	Net                              uint32             `description:"Net defines the magic bytes used to identify the network"`
	GenesisParam                     *genesisParam      `mapstructure:"genesis_param" description:"genesis params"`
	CommitteeSize                    committeeSize      `mapstructure:"committee_size"`
	BlockTime                        blockTime          `mapstructure:"block_time"`
	StakingAmountShard               uint64             `mapstructure:"staking_amount_shard"`
	ActiveShards                     int                `mapstructure:"active_shards"`
	BasicReward                      uint64             `mapstructure:"basic_reward"`
	EpochParam                       epochParam         `mapstructure:"epoch_param"`
	EthContractAddressStr            string             `mapstructure:"eth_contract_address" description:"smart contract of ETH for bridge"`
	IncognitoDAOAddress              string             `mapstructure:"dao_address"`
	CentralizedWebsitePaymentAddress string             `mapstructure:"centralized_website_payment_address" description:"centralized website's pubkey"`
	SwapCommitteeParam               swapCommitteeParam `mapstructure:"swap_committee_param"`
	ConsensusParam                   consensusParam     `mapstructure:"consensus_param"`
	BeaconHeightBreakPointBurnAddr   uint64             `mapstructure:"beacon_height_break_point_burn_addr"`
	ReplaceStakingTxHeight           uint64             `mapstructure:"replace_staking_tx_height"`
	ETHRemoveBridgeSigEpoch          uint64             `mapstructure:"eth_remove_bridge_sig_epoch"`
	BCHeightBreakPointNewZKP         uint64             `mapstructure:"bc_height_break_point_new_zkp"`
	EnableFeatureFlags               map[int]uint64     `mapstructure:"enable_feature_flags" description:"featureFlag: epoch number - since that time, the feature will be enabled; 0 - disabled feature"`
	PortalParam                      PortalParam        `mapstructure:"portal_param"`
	IsBackup                         bool
}

type genesisParam struct {
	InitialIncognito                            []initialIncognito `mapstructure:"initial_incognito" description:"init tx for genesis block"`
	FeePerTxKb                                  uint64             `mapstructure:"fee_per_tx_kb" description:"fee per tx calculate by kb"`
	ConsensusAlgorithm                          string             `mapstructure:"consensus_algorithm"`
	BlockTimestamp                              string             `mapstructure:"block_timestamp"`
	TxStake                                     string             `mapstructure:"tx_stake"`
	PreSelectBeaconNodeSerializedPubkey         []string
	SelectBeaconNodeSerializedPubkeyV2          map[uint64][]string
	PreSelectBeaconNodeSerializedPaymentAddress []string
	SelectBeaconNodeSerializedPaymentAddressV2  map[uint64][]string
	PreSelectShardNodeSerializedPubkey          []string
	SelectShardNodeSerializedPubkeyV2           map[uint64][]string
	PreSelectShardNodeSerializedPaymentAddress  []string
	SelectShardNodeSerializedPaymentAddressV2   map[uint64][]string
}

type committeeSize struct {
	MaxShardCommitteeSize   int `mapstructure:"max_shard_committee_size"`
	MinShardCommitteeSize   int `mapstructure:"min_shard_committee_size"`
	MaxBeaconCommitteeSize  int `mapstructure:"max_beacon_committee_size"`
	MinBeaconCommitteeSize  int `mapstructure:"min_beacon_committee_size"`
	InitShardCommitteeSize  int `mapstructure:"init_shard_committee_size"`
	InitBeaconCommitteeSize int `mapstructure:"init_beacon_committee_size"`
}

type blockTime struct {
	MinShardBlockInterval        time.Duration `mapstructure:"min_shard_block_interval"`
	MaxShardBlockCreation        time.Duration `mapstructure:"max_shard_block_creation"`
	MinBeaconBlockInterval       time.Duration `mapstructure:"min_beacon_block_interval"`
	MaxBeaconBlockCreation       time.Duration `mapstructure:"min_beacon_block_creation"`
	NumberOfFixedBlockValidators int           `mapstructure:"number_of_fixed_shard_block_validators"`
}

type epochParam struct {
	NumberOfBlockInEpoch   uint64 `mapstructure:"number_of_block_in_epoch"`
	NumberOfBlockInEpochV2 uint64 `mapstructure:"number_of_block_in_epoch_v2"`
	EpochV2BreakPoint      uint64 `mapstructure:"epoch_v2_break_point"`
	RandomTime             uint64 `mapstructure:"random_time"`
	RandomTimeV2           uint64 `mapstructure:"random_time_v2"`
}

type swapCommitteeParam struct {
	Offset       int `mapstructure:"offset" description:"default offset for swap policy, is used for cases that good producers length is less than max committee size"`
	SwapOffset   int `mapstructure:"swap_offset" description:"is used for case that good producers length is equal to max committee size"`
	AssignOffset int `mapstructure:"assign_offset"`
}

type consensusParam struct {
	ConsensusV2Epoch          uint64   `mapstructure:"consensus_v2_epoch"`
	StakingFlowV2Height       uint64   `mapstructure:"staking_flow_v2_height"`
	EnableSlashingHeight      uint64   `mapstructure:"enable_slashing_height"`
	Timeslot                  uint64   `mapstructure:"timeslot"`
	EpochBreakPointSwapNewKey []uint64 `mapstructure:"epoch_break_point_swap_new_key"`
}

func LoadParam() *param {
	network := ""
	switch common.GetEnv(NetworkKey, LocalNetwork) {
	case LocalNetwork:
		network = LocalNetwork
		c.IsLocal = true
	case TestNetNetwork:
		network = TestNetNetwork
		c.IsTestNet = true
		testnetVersion := common.GetEnv(NetworkVersionKey, TestNetVersion1)
		version, err := strconv.Atoi(testnetVersion)
		if err != nil {
			panic(err)
		}
		network += testnetVersion
		c.TestNetVersion = version
	case MainnetNetwork:
		network = MainnetNetwork
		c.IsMainNet = true
	}

	//read config from file
	viper.SetConfigName(common.GetEnv(ParamFileKey, DefaultParamFile))                         // name of config file (without extension)
	viper.SetConfigType(common.GetEnv(ConfigFileTypeKey, DefaultConfigFileType))               // REQUIRED if the config file does not have the extension in the name
	viper.AddConfigPath(filepath.Join(common.GetEnv(ConfigDirKey, DefaultConfigDir), network)) // optionally look for config in the working directory
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			// Config file was found but another error was produced
			if err != nil {
				panic(err)
			}
		} else {
			panic(err)
		}
	} else {
		err = viper.Unmarshal(&p)
		if err != nil {
			panic(err)
		}
	}

	return p
}

type initialIncognito struct {
	Version              int    `mapstructure:"Version"`
	Type                 string `mapstructure:"Type"`
	LockTime             uint64 `mapstructure:"LockTime"`
	Fee                  int    `mapstructure:"Fee"`
	Info                 string `mapstructure:"Info"`
	SigPubKey            string `mapstructure:"SigPubKey"`
	Sig                  string `mapstructure:"Sig"`
	Proof                string `mapstructure:"Proof"`
	PubKeyLastByteSender int    `mapstructure:"PubKeyLastByteSender"`
	Metadata             string `mapstructure:"Metadata"`
}
