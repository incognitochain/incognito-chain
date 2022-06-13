package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"
	"strings"
	"time"

	"github.com/incognitochain/incognito-chain/common"

	"github.com/incognitochain/incognito-chain/utils"
	"github.com/spf13/viper"
)

var p *param

func Param() *param {
	return p
}

//AbortParam use for unit test only
// DO NOT use this function for development process
func AbortParam() {
	p = &param{}
}

type AutoEnableFeature struct {
	MinTriggerBlockHeight int `mapstructure:"min_trigger"`
	ForceBlockHeight      int `mapstructure:"force_trigger"`
	RequiredPercentage    int `mapstructure:"require_percentage"`
}

//param for all variables in incognito node process
type param struct {
	Name                             string                       `mapstructure:"name" description:"Name defines a human-readable identifier for the network" `
	Net                              uint32                       `mapstructure:"net" description:"Net defines the magic bytes used to identify the network"`
	GenesisParam                     *genesisParam                `mapstructure:"genesis_param" description:"genesis params"`
	CommitteeSize                    committeeSize                `mapstructure:"committee_size"`
	BlockTime                        blockTime                    `mapstructure:"block_time"`
	StakingAmountShard               uint64                       `mapstructure:"staking_amount_shard"`
	ActiveShards                     int                          `mapstructure:"active_shards"`
	BasicReward                      uint64                       `mapstructure:"basic_reward"`
	EpochParam                       epochParam                   `mapstructure:"epoch_param"`
	EthContractAddressStr            string                       `mapstructure:"eth_contract_address" description:"smart contract of ETH for bridge"`
	BscContractAddressStr            string                       `mapstructure:"bsc_contract_address" description:"smart contract of BSC for bridge"`
	PlgContractAddressStr            string                       `mapstructure:"plg_contract_address" description:"smart contract of PLG for bridge"`
	FtmContractAddressStr            string                       `mapstructure:"ftm_contract_address" description:"smart contract of FTM for bridge"`
	IncognitoDAOAddress              string                       `mapstructure:"dao_address"`
	CentralizedWebsitePaymentAddress string                       `mapstructure:"centralized_website_payment_address" description:"centralized website's pubkey"`
	SwapCommitteeParam               swapCommitteeParam           `mapstructure:"swap_committee_param"`
	ConsensusParam                   consensusParam               `mapstructure:"consensus_param"`
	BeaconHeightBreakPointBurnAddr   uint64                       `mapstructure:"beacon_height_break_point_burn_addr"`
	ReplaceStakingTxHeight           uint64                       `mapstructure:"replace_staking_tx_height"`
	ETHRemoveBridgeSigEpoch          uint64                       `mapstructure:"eth_remove_bridge_sig_epoch"`
	BCHeightBreakPointNewZKP         uint64                       `mapstructure:"bc_height_break_point_new_zkp"`
	BCHeightBreakPointPrivacyV2      uint64                       `mapstructure:"bc_height_break_point_privacy_v2"`
	CoinVersion2LowestHeight         uint64                       `mapstructure:"coin_v2_lowest_height"`
	EnableFeatureFlags               map[string]uint64            `mapstructure:"enable_feature_flags" description:"featureFlag: epoch number - since that time, the feature will be enabled; 0 - disabled feature"`
	BCHeightBreakPointPortalV3       uint64                       `mapstructure:"portal_v3_height"`
	TxPoolVersion                    int                          `mapstructure:"tx_pool_version"`
	GethParam                        gethParam                    `mapstructure:"geth_param"`
	BSCParam                         bscParam                     `mapstructure:"bsc_param"`
	PLGParam                         plgParam                     `mapstructure:"plg_param"`
	FTMParam                         ftmParam                     `mapstructure:"ftm_param"`
	PDexParams                       pdexParam                    `mapstructure:"pdex_param"`
	IsEnableBPV3Stats                bool                         `mapstructure:"is_enable_bpv3_stats"`
	BridgeAggParam                   bridgeAggParam               `mapstructure:"bridge_agg_param"`
	AutoEnableFeature                map[string]AutoEnableFeature `mapstructure:"auto_enable_feature"`
	IsBackup                         bool
	PRVERC20ContractAddressStr       string `mapstructure:"prv_erc20_contract_address" description:"smart contract of prv erc20"`
	PRVBEP20ContractAddressStr       string `mapstructure:"prv_bep20_contract_address" description:"smart contract of prv bep20"`
	BCHeightBreakPointCoinOrigin     uint64             `mapstructure:"bc_height_break_point_coin_origin"`
	BatchCommitSyncModeParam         batchCommitSyncModeParam `mapstructure:"batch_commit_sync_mode_param"`
	FlatFileParam                    flatfileParam            `mapstructure:"flatfileparam"`
}

type genesisParam struct {
	InitialIncognito                            []InitialIncognito
	FeePerTxKb                                  uint64 `mapstructure:"fee_per_tx_kb" description:"fee per tx calculate by kb"`
	ConsensusAlgorithm                          string `mapstructure:"consensus_algorithm"`
	BlockTimestamp                              string `mapstructure:"block_timestamp"`
	TxStake                                     string `mapstructure:"tx_stake"`
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
	MaxShardCommitteeSize            int            `mapstructure:"max_shard_committee_size"`
	MinShardCommitteeSize            int            `mapstructure:"min_shard_committee_size"`
	MaxBeaconCommitteeSize           int            `mapstructure:"max_beacon_committee_size"`
	MinBeaconCommitteeSize           int            `mapstructure:"min_beacon_committee_size"`
	InitShardCommitteeSize           int            `mapstructure:"init_shard_committee_size"`
	InitBeaconCommitteeSize          int            `mapstructure:"init_beacon_committee_size"`
	ShardCommitteeSizeKeyListV2      int            `mapstructure:"shard_committee_size_key_list_v2"`
	BeaconCommitteeSizeKeyListV2     int            `mapstructure:"beacon_committee_size_key_list_v2"`
	NumberOfFixedShardBlockValidator int            `mapstructure:"number_of_fixed_shard_block_validators"`
	IncreaseMaxShardCommitteeSize    map[uint64]int `mapstructure:"increase_max_shard_committee_size"`
}

type blockTime struct {
	MinShardBlockInterval  time.Duration `mapstructure:"min_shard_block_interval"`
	MaxShardBlockCreation  time.Duration `mapstructure:"max_shard_block_creation"`
	MinBeaconBlockInterval time.Duration `mapstructure:"min_beacon_block_interval"`
	MaxBeaconBlockCreation time.Duration `mapstructure:"max_beacon_block_creation"`
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
	AssignRuleV3Height        uint64   `mapstructure:"assign_rule_v3_height"`
	EnableSlashingHeight      uint64   `mapstructure:"enable_slashing_height"`
	EnableSlashingHeightV2    uint64   `mapstructure:"enable_slashing_height_v2"`
	StakingFlowV3Height       uint64   `mapstructure:"staking_flow_v3_height"`
	NotUseBurnedCoins         uint64   `mapstructure:"force_not_use_burned_coins"`
	Lemma2Height              uint64   `mapstructure:"lemma2_height"`
	ByzantineDetectorHeight   uint64   `mapstructure:"byzantine_detector_height"`
	BlockProducingV3Height    uint64   `mapstructure:"block_producing_v3_height"`
	Timeslot                  uint64   `mapstructure:"timeslot"`
	EpochBreakPointSwapNewKey []uint64 `mapstructure:"epoch_break_point_swap_new_key"`
}

type batchCommitSyncModeParam struct {
	TrieJournalCacheSize int                `mapstructure:"trie_journal_cache_size"`
	BlockTrieInMemory    uint64             `mapstructure:"block_trie_in_memory"`
	TrieNodeLimit        common.StorageSize `mapstructure:"trie_node_limit"`
	TrieImgsLimit        common.StorageSize `mapstructure:"trie_img_limit"`
}

type flatfileParam struct {
	MaxCacheSize uint64 `mapstructure:"maxcachesize"`
	CompLevel    int    `mapstructure:"compresslevel"`
}

func LoadParam() *param {

	network := c.Network()
	p = NewDefaultParam(network)

	//read config from file to overwrite default param
	viper.SetConfigName(utils.GetEnv(ParamFileKey, DefaultParamFile))                         // name of config file (without extension)
	viper.SetConfigType(utils.GetEnv(ConfigFileTypeKey, DefaultConfigFileType))               // REQUIRED if the config file does not have the extension in the name
	viper.AddConfigPath(filepath.Join(utils.GetEnv(ConfigDirKey, DefaultConfigDir), network)) // optionally look for config in the working directory
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			// Config file was found but another error was produced
			if err != nil {
				panic(err)
			}
		} else {
			log.Println("Using default network param for " + network)
		}
	} else {
		log.Println("Using network param file for " + network)
		//p has default param, below function only update fields which is specified in the param file
		err = viper.Unmarshal(&p)
		if err != nil {
			panic(err)
		}
	}
	p.LoadKeyByNetwork(network)
	common.TIMESLOT = p.ConsensusParam.Timeslot
	common.MaxShardNumber = p.ActiveShards

	if err := verifyParam(p); err != nil {
		panic(err)
	}

	return p
}

func verifyParam(p *param) error {

	if p.CommitteeSize.MaxShardCommitteeSize < p.CommitteeSize.MinShardCommitteeSize {
		return fmt.Errorf("MaxCommitteeSize %+v < MinCommitteeSize %+v",
			p.CommitteeSize.MaxShardCommitteeSize, p.CommitteeSize.MinShardCommitteeSize)
	}

	if p.CommitteeSize.MaxBeaconCommitteeSize < p.CommitteeSize.MinBeaconCommitteeSize {
		return fmt.Errorf("MaxCommitteeSize %+v < MinCommitteeSize %+v",
			p.CommitteeSize.MaxBeaconCommitteeSize, p.CommitteeSize.MinBeaconCommitteeSize)
	}

	if p.CommitteeSize.MinShardCommitteeSize < p.CommitteeSize.NumberOfFixedShardBlockValidator {
		return fmt.Errorf("MinShardCommitteeSize %+v < NumberOfFixedShardBlockValidator %+v",
			p.CommitteeSize.MinShardCommitteeSize, p.CommitteeSize.NumberOfFixedShardBlockValidator)
	}

	if p.CommitteeSize.MinShardCommitteeSize < 4 {
		return fmt.Errorf("MinShardCommitteeSize %+v < %+v",
			p.CommitteeSize.MinShardCommitteeSize, 4)
	}

	if p.CommitteeSize.MinBeaconCommitteeSize < 4 {
		return fmt.Errorf("MinBeaconCommitteeSize %+v < %+v",
			p.CommitteeSize.MinBeaconCommitteeSize, 4)
	}

	if p.CommitteeSize.InitShardCommitteeSize != p.CommitteeSize.ShardCommitteeSizeKeyListV2 {
		return fmt.Errorf("InitShardCommitteeSize %+v < ShardCommitteeSizeKeyListV2 %+v",
			p.CommitteeSize.InitShardCommitteeSize, p.CommitteeSize.ShardCommitteeSizeKeyListV2)
	}

	if p.CommitteeSize.InitBeaconCommitteeSize != p.CommitteeSize.BeaconCommitteeSizeKeyListV2 {
		return fmt.Errorf("InitBeaconCommitteeSize %+v < BeaconCommitteeSizeKeyListV2 %+v",
			p.CommitteeSize.InitBeaconCommitteeSize, p.CommitteeSize.BeaconCommitteeSizeKeyListV2)
	}

	if p.EpochParam.RandomTime >= p.EpochParam.NumberOfBlockInEpoch {
		return fmt.Errorf("RandomTime %+v >= NumberOfBlockInEpoch %+v",
			p.EpochParam.RandomTime, p.EpochParam.NumberOfBlockInEpoch)
	}

	return nil
}

//key1,key2 : default key of the network
func (p *param) LoadKey(key1 []byte, key2 []byte) {
	network := c.Network()
	configPath := filepath.Join(utils.GetEnv(ConfigDirKey, DefaultConfigDir), network)

	//load from file, otherwise use default setup
	keyData, err := ioutil.ReadFile(filepath.Join(configPath, KeyListFileName))
	if err != nil {
		if strings.Index(err.Error(), "no such file or directory") > -1 && key1 != nil {
			keyData = key1
			log.Println("Using default keylist 1 for " + network)
		} else {
			panic(err)
		}
	}

	//load from file, otherwise use default setup
	keyDataV2, err := ioutil.ReadFile(filepath.Join(configPath, KeyListV2FileName))
	if err != nil {
		if strings.Index(err.Error(), "no such file or directory") > -1 && key2 != nil {
			keyDataV2 = key2
			log.Println("Using default keylist 2 for " + network)
		} else {
			panic(err)
		}
	}

	type AccountKey struct {
		PaymentAddress     string
		CommitteePublicKey string
	}

	type KeyList struct {
		Shard  map[int][]AccountKey
		Beacon []AccountKey
	}

	keylist := KeyList{}
	keylistV2 := []KeyList{}

	err = json.Unmarshal(keyData, &keylist)
	if err != nil {
		panic(err)
	}

	err = json.Unmarshal(keyDataV2, &keylistV2)
	if err != nil {
		panic(err)
	}

	for i := 0; i < p.CommitteeSize.InitBeaconCommitteeSize; i++ {
		p.GenesisParam.PreSelectBeaconNodeSerializedPubkey =
			append(p.GenesisParam.PreSelectBeaconNodeSerializedPubkey, keylist.Beacon[i].CommitteePublicKey)
		p.GenesisParam.PreSelectBeaconNodeSerializedPaymentAddress =
			append(p.GenesisParam.PreSelectBeaconNodeSerializedPaymentAddress, keylist.Beacon[i].PaymentAddress)
	}

	for i := 0; i < p.ActiveShards; i++ {
		for j := 0; j < p.CommitteeSize.InitShardCommitteeSize; j++ {
			p.GenesisParam.PreSelectShardNodeSerializedPubkey =
				append(p.GenesisParam.PreSelectShardNodeSerializedPubkey, keylist.Shard[i][j].CommitteePublicKey)
			p.GenesisParam.PreSelectShardNodeSerializedPaymentAddress =
				append(p.GenesisParam.PreSelectShardNodeSerializedPaymentAddress, keylist.Shard[i][j].PaymentAddress)
		}
	}
	for _, v := range keylistV2 {
		for i := 0; i < p.CommitteeSize.BeaconCommitteeSizeKeyListV2; i++ {
			p.GenesisParam.SelectBeaconNodeSerializedPubkeyV2[p.ConsensusParam.EpochBreakPointSwapNewKey[0]] =
				append(p.GenesisParam.SelectBeaconNodeSerializedPubkeyV2[p.ConsensusParam.EpochBreakPointSwapNewKey[0]], v.Beacon[i].CommitteePublicKey)
			p.GenesisParam.SelectBeaconNodeSerializedPaymentAddressV2[p.ConsensusParam.EpochBreakPointSwapNewKey[0]] =
				append(p.GenesisParam.SelectBeaconNodeSerializedPaymentAddressV2[p.ConsensusParam.EpochBreakPointSwapNewKey[0]], v.Beacon[i].PaymentAddress)
		}
		for i := 0; i < p.ActiveShards; i++ {
			for j := 0; j < p.CommitteeSize.ShardCommitteeSizeKeyListV2; j++ {
				p.GenesisParam.SelectShardNodeSerializedPubkeyV2[p.ConsensusParam.EpochBreakPointSwapNewKey[0]] =
					append(p.GenesisParam.SelectShardNodeSerializedPubkeyV2[p.ConsensusParam.EpochBreakPointSwapNewKey[0]], v.Shard[i][j].CommitteePublicKey)
				p.GenesisParam.SelectShardNodeSerializedPaymentAddressV2[p.ConsensusParam.EpochBreakPointSwapNewKey[0]] =
					append(p.GenesisParam.SelectShardNodeSerializedPaymentAddressV2[p.ConsensusParam.EpochBreakPointSwapNewKey[0]], v.Shard[i][j].PaymentAddress)
			}
		}
	}

}

type bscParam struct {
	Host []string `mapstructure:"host"`
}

type pdexParam struct {
	Pdexv3BreakPointHeight uint64 `mapstructure:"pdex_v3_break_point_height"`
	ProtocolFundAddress    string `mapstructure:"protocol_fund_address"`
	AdminAddress           string `mapstructure:"admin_address"`
	Params                 struct {
		DefaultFeeRateBPS               uint            `mapstructure:"default_fee_rate_bps"`
		PRVDiscountPercent              uint            `mapstructure:"prv_discount_percent"`
		TradingProtocolFeePercent       uint            `mapstructure:"trading_protocol_fee_percent"`
		TradingStakingPoolRewardPercent uint            `mapstructure:"trading_staking_pool_reward_percent"`
		StakingPoolsShare               map[string]uint `mapstructure:"staking_pool_share"`
		MintNftRequireAmount            uint64          `mapstructure:"mint_nft_require_amount"`
		MaxOrdersPerNft                 uint            `mapstructure:"max_orders_per_nft"`
		AutoWithdrawOrderLimitAmount    uint            `mapstructure:"auto_withdraw_order_limit_amount"`
		MinPRVReserveTradingRate        uint64          `mapstructure:"min_prv_reserve_trading_rate"`
	} `mapstructure:"params"`
}

func (bschParam *bscParam) GetFromEnv() {
	if utils.GetEnv(BSCHostKey, utils.EmptyString) != utils.EmptyString {
		bschParam.Host = []string{utils.GetEnv(BSCHostKey, utils.EmptyString)}
	}
}

type plgParam struct {
	Host []string `mapstructure:"host"`
}

func (plgParam *plgParam) GetFromEnv() {
	if utils.GetEnv(PLGHostKey, utils.EmptyString) != utils.EmptyString {
		plgParam.Host = []string{utils.GetEnv(PLGHostKey, utils.EmptyString)}
	}
}

type ftmParam struct {
	Host []string `mapstructure:"host"`
}

func (ftmParam *ftmParam) GetFromEnv() {
	if utils.GetEnv(FTMHostKey, utils.EmptyString) != utils.EmptyString {
		ftmParam.Host = []string{utils.GetEnv(FTMHostKey, utils.EmptyString)}
	}
}

type bridgeAggParam struct {
	AdminAddress string `mapstructure:"admin_address"`
	BaseDecimal  uint   `mapstructure:"base_decimal"`
	MaxLenOfPath int    `mapstructure:"max_len_of_path"`
}
