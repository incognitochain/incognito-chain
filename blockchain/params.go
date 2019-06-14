package blockchain

import (
	"github.com/incognitochain/incognito-chain/blockchain/benchmark"
)

/*
Params defines a network by its component. These component may be used by Applications
to differentiate network as well as addresses and keys for one network
from those intended for use on another network
*/
type Params struct {
	// Name defines a human-readable identifier for the network.
	Name string

	// Net defines the magic bytes used to identify the network.
	Net uint32

	// DefaultPort defines the default peer-to-peer port for the network.
	DefaultPort         string
	ShardCommitteeSize  int
	BeaconCommitteeSize int
	StakingAmountShard  uint64
	ActiveShards        int
	// GenesisBlock defines the first block of the chain.
	GenesisBeaconBlock *BeaconBlock

	// GenesisBlock defines the first block of the chain.
	GenesisShardBlock *ShardBlock
	BasicReward       uint64
	RewardHalflife    uint64
}

type GenesisParams struct {
	InitialPaymentAddress string

	InitialConstant []string

	FeePerTxKb uint64

	RandomNumber uint64

	PreSelectBeaconNodeSerializedPubkey []string
	PreSelectBeaconNode                 []string
	PreSelectShardNodeSerializedPubkey  []string
	PreSelectShardNode                  []string
}
// FOR TESTNET
var genesisParamsTestnetNew = GenesisParams{
	InitialPaymentAddress:               TestnetGenesisBlockPaymentAddress,
	RandomNumber:                        0,
	PreSelectBeaconNodeSerializedPubkey: PreSelectBeaconNodeTestnetSerializedPubkey,
	PreSelectShardNodeSerializedPubkey:  PreSelectShardNodeTestnetSerializedPubkey,

	//@Notice: InitTxsForBenchmark is for testing and benchmark only
	//InitialConstant: append(benchmark.InitTxsShard0, append(benchmark.InitTxsShard1, append(benchmark.InitTxsShard0_1, append(benchmark.InitTxsShard0_2, append(benchmark.InitTxsShard0_3, append(benchmark.InitTxsShard0_4, append(benchmark.InitTxsShard0_5, append(benchmark.InitTxsShard0_6, append(benchmark.InitTxsShard0_7, append(benchmark.InitTxsShard0_8, append(benchmark.InitTxsShard0_9, append(benchmark.InitTxsShard0_10, append(benchmark.InitTxsShard1_1, append(benchmark.InitTxsShard1_2, append(benchmark.InitTxsShard1_3, append(benchmark.InitTxsShard1_4, append(benchmark.InitTxsShard1_5, append(benchmark.InitTxsShard1_6, append(benchmark.InitTxsShard1_7, append(benchmark.InitTxsShard1_8, append(benchmark.InitTxsShard1_9, benchmark.InitTxsShard1_10...)...)...)...)...)...)...)...)...)...)...)...)...)...)...)...)...)...)...)...)...),
	InitialConstant: benchmark.GetInitTransaction(),
}

var ChainTestParam = Params{
	Name:                TestnetName,
	Net:                 Testnet,
	DefaultPort:         TestnetDefaultPort,
	ShardCommitteeSize:  TestNetShardCommitteeSize,  //TestNetShardCommitteeSize,
	BeaconCommitteeSize: TestNetBeaconCommitteeSize, //TestNetBeaconCommitteeSize,
	StakingAmountShard:  TestNetStakingAmountShard,
	ActiveShards:        TestNetActiveShards,
	// blockChain parameters
	GenesisBeaconBlock: CreateBeaconGenesisBlock(1, genesisParamsTestnetNew),
	GenesisShardBlock:  CreateShardGenesisBlock(1, genesisParamsTestnetNew),
	BasicReward:        TestnetBasicReward,
	RewardHalflife:     TestnetRewardHalflife,
}

// END TESTNET

// FOR MAINNET
var genesisParamsMainnetNew = GenesisParams{
	InitialPaymentAddress:               MainnetGenesisblockPaymentAddress,
	RandomNumber:                        0,
	PreSelectBeaconNodeSerializedPubkey: PreSelectBeaconNodeMainnetSerializedPubkey,
	PreSelectShardNodeSerializedPubkey:  PreSelectShardNodeMainnetSerializedPubkey,

	InitialConstant: MainnetInitConstant,
}

var ChainMainParam = Params{
	Name:                MainetName,
	Net:                 Mainnet,
	DefaultPort:         MainnetDefaultPort,
	ShardCommitteeSize:  MainNetShardCommitteeSize,  //MainNetShardCommitteeSize,
	BeaconCommitteeSize: MainNetBeaconCommitteeSize, //MainNetBeaconCommitteeSize,
	StakingAmountShard:  MainNetStakingAmountShard,
	ActiveShards:        MainNetActiveShards,
	// blockChain parameters
	GenesisBeaconBlock: CreateBeaconGenesisBlock(1, genesisParamsMainnetNew),
	GenesisShardBlock:  CreateShardGenesisBlock(1, genesisParamsMainnetNew),
	BasicReward:        MainnetBasicReward,
	RewardHalflife:     MainnetRewardHalflife,
}
