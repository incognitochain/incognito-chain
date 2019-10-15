package blockchain

import (
	"time"

	"github.com/incognitochain/incognito-chain/common"
)

type SlashLevel struct {
	MinRange        uint8
	PunishedEpoches uint8
}

/*
Params defines a network by its component. These component may be used by Applications
to differentiate network as well as addresses and keys for one network
from those intended for use on another network
*/
type Params struct {
	Name                             string // Name defines a human-readable identifier for the network.
	Net                              uint32 // Net defines the magic bytes used to identify the network.
	DefaultPort                      string // DefaultPort defines the default peer-to-peer port for the network.
	MaxShardCommitteeSize            int
	MinShardCommitteeSize            int
	MaxBeaconCommitteeSize           int
	MinBeaconCommitteeSize           int
	MinShardBlockInterval            time.Duration
	MaxShardBlockCreation            time.Duration
	MinBeaconBlockInterval           time.Duration
	MaxBeaconBlockCreation           time.Duration
	StakingAmountShard               uint64
	ActiveShards                     int
	GenesisBeaconBlock               *BeaconBlock // GenesisBlock defines the first block of the chain.
	GenesisShardBlock                *ShardBlock  // GenesisBlock defines the first block of the chain.
	BasicReward                      uint64
	RewardHalflife                   uint64
	Epoch                            uint64
	RandomTime                       uint64
	SlashLevels                      []SlashLevel
	EthContractAddressStr            string // smart contract of ETH for bridge
	Offset                           int    // default offset for swap policy, is used for cases that good producers length is less than max committee size
	SwapOffset                       int    // is used for case that good producers length is equal to max committee size
	DevAddress                       string
	CentralizedWebsitePaymentAddress string //centralized website's pubkey
	CheckForce                       bool   // true on testnet and false on mainnet
	ChainVersion                     string
}

type GenesisParams struct {
	InitialIncognito                            []string
	FeePerTxKb                                  uint64
	RandomNumber                                uint64
	PreSelectBeaconNodeSerializedPubkey         []string
	PreSelectBeaconNodeSerializedPaymentAddress []string
	PreSelectBeaconNode                         []string
	PreSelectShardNodeSerializedPubkey          []string
	PreSelectShardNodeSerializedPaymentAddress  []string
	PreSelectShardNode                          []string
	ConsensusAlgorithm                          string
}

var ChainTestParam = Params{}
var ChainMainParam = Params{}

// FOR TESTNET
func init() {
	var genesisParamsTestnetNew = GenesisParams{
		RandomNumber:                                0,
		PreSelectBeaconNodeSerializedPubkey:         PreSelectBeaconNodeTestnetSerializedPubkey,
		PreSelectBeaconNodeSerializedPaymentAddress: PreSelectBeaconNodeTestnetSerializedPaymentAddress,
		PreSelectShardNodeSerializedPubkey:          PreSelectShardNodeTestnetSerializedPubkey,
		PreSelectShardNodeSerializedPaymentAddress:  PreSelectShardNodeTestnetSerializedPaymentAddress,
		//@Notice: InitTxsForBenchmark is for testing and testparams only
		//InitialIncognito: IntegrationTestInitPRV,
		InitialIncognito:   TestnetInitPRV,
		ConsensusAlgorithm: common.BlsConsensus,
	}
	ChainTestParam = Params{
		Name:                   TestnetName,
		Net:                    Testnet,
		DefaultPort:            TestnetDefaultPort,
		MaxShardCommitteeSize:  TestNetShardCommitteeSize,     //TestNetShardCommitteeSize,
		MinShardCommitteeSize:  TestNetMinShardCommitteeSize,  //TestNetShardCommitteeSize,
		MaxBeaconCommitteeSize: TestNetBeaconCommitteeSize,    //TestNetBeaconCommitteeSize,
		MinBeaconCommitteeSize: TestNetMinBeaconCommitteeSize, //TestNetBeaconCommitteeSize,
		StakingAmountShard:     TestNetStakingAmountShard,
		ActiveShards:           TestNetActiveShards,
		// blockChain parameters
		GenesisBeaconBlock:               CreateBeaconGenesisBlock(1, genesisParamsTestnetNew),
		GenesisShardBlock:                CreateShardGenesisBlock(1, genesisParamsTestnetNew),
		MinShardBlockInterval:            TestNetMinShardBlkInterval,
		MaxShardBlockCreation:            TestNetMaxShardBlkCreation,
		MinBeaconBlockInterval:           TestNetMinBeaconBlkInterval,
		MaxBeaconBlockCreation:           TestNetMaxBeaconBlkCreation,
		BasicReward:                      TestnetBasicReward,
		RewardHalflife:                   TestnetRewardHalflife,
		Epoch:                            TestnetEpoch,
		RandomTime:                       TestnetRandomTime,
		Offset:                           TestnetOffset,
		SwapOffset:                       TestnetSwapOffset,
		EthContractAddressStr:            TestnetETHContractAddressStr,
		DevAddress:                       TestnetDevAddress,
		CentralizedWebsitePaymentAddress: TestnetCentralizedWebsitePaymentAddress,
		SlashLevels: []SlashLevel{
			//SlashLevel{MinRange: 20, PunishedEpoches: 1},
			SlashLevel{MinRange: 50, PunishedEpoches: 2},
			SlashLevel{MinRange: 75, PunishedEpoches: 3},
		},
		CheckForce:   false,
		ChainVersion: "version-chain.json",
	}
	// END TESTNET
	// FOR MAINNET
	var genesisParamsMainnetNew = GenesisParams{
		RandomNumber:                        0,
		PreSelectBeaconNodeSerializedPubkey: PreSelectBeaconNodeMainnetSerializedPubkey,
		PreSelectShardNodeSerializedPubkey:  PreSelectShardNodeMainnetSerializedPubkey,
		InitialIncognito:                    MainnetInitPRV,
		ConsensusAlgorithm:                  common.BlsConsensus,
	}
	ChainMainParam = Params{
		Name:                   MainetName,
		Net:                    Mainnet,
		DefaultPort:            MainnetDefaultPort,
		MaxShardCommitteeSize:  MainNetShardCommitteeSize,  //MainNetShardCommitteeSize,
		MaxBeaconCommitteeSize: MainNetBeaconCommitteeSize, //MainNetBeaconCommitteeSize,
		StakingAmountShard:     MainNetStakingAmountShard,
		ActiveShards:           MainNetActiveShards,
		// blockChain parameters
		GenesisBeaconBlock:               CreateBeaconGenesisBlock(1, genesisParamsMainnetNew),
		GenesisShardBlock:                CreateShardGenesisBlock(1, genesisParamsMainnetNew),
		MinShardBlockInterval:            MainnetMinShardBlkInterval,
		MaxShardBlockCreation:            MainnetMaxShardBlkCreation,
		MinBeaconBlockInterval:           MainnetMinBeaconBlkInterval,
		MaxBeaconBlockCreation:           MainnetMaxBeaconBlkCreation,
		BasicReward:                      MainnetBasicReward,
		RewardHalflife:                   MainnetRewardHalflife,
		Epoch:                            MainnetEpoch,
		RandomTime:                       MainnetRandomTime,
		Offset:                           MainnetOffset,
		SwapOffset:                       MainnetSwapOffset,
		EthContractAddressStr:            MainETHContractAddressStr,
		DevAddress:                       MainnetDevAddress,
		CentralizedWebsitePaymentAddress: MainnetCentralizedWebsitePaymentAddress,
		SlashLevels: []SlashLevel{
			//SlashLevel{MinRange: 20, PunishedEpoches: 1},
			SlashLevel{MinRange: 50, PunishedEpoches: 2},
			SlashLevel{MinRange: 75, PunishedEpoches: 3},
		},
		CheckForce:   false,
		ChainVersion: "version-chain-main.json",
	}
}
