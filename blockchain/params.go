package blockchain

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

	BasicReward    uint64
	RewardHalflife uint64
	FeePerTxKb     uint64

	RandomNumber uint64

	PreSelectBeaconNodeSerializedPubkey []string
	PreSelectBeaconNode                 []string
	PreSelectShardNodeSerializedPubkey  []string
	PreSelectShardNode                  []string
}

// FOR TESTNET
var genesisParamsTestnetNew = GenesisParams{
	InitialPaymentAddress:               TestnetGenesisBlockPaymentAddress,
	BasicReward:                         TestnetBasicReward,
	RewardHalflife:                      TestnetRewardHalflife,
	FeePerTxKb:                          TestnetFeePerTxKb,
	RandomNumber:                        0,
	PreSelectBeaconNodeSerializedPubkey: PreSelectBeaconNodeTestnetSerializedPubkey,
	PreSelectShardNodeSerializedPubkey:  PreSelectShardNodeTestnetSerializedPubkey,

	//@Notice: InitTxsForBenchmark is for testing and benchmark only
	InitialConstant: append(TestnetInitConstant, append(InitTxsShard0, append(InitTxsShard1, InitTxsForBenchmark...)...)...),
	InitialConstant: append(TestnetInitConstant, append(InitTxsShard0_1, InitTxsShard1_1...)...),
	//InitialConstant: append(TestnetInitConstant, InitTxsShard0...),
	//InitialConstant: TestnetInitConstant,
}

var ChainTestParam = Params{
	Name:                TestnetName,
	Net:                 Testnet,
	DefaultPort:         TestnetDefaultPort,
	ShardCommitteeSize:  TestNetShardCommitteeSize,  //TestNetShardCommitteeSize,
	BeaconCommitteeSize: TestNetBeaconCommitteeSize, //TestNetBeaconCommitteeSize,
	ActiveShards:        TestNetActiveShards,
	// blockChain parameters
	GenesisBeaconBlock: CreateBeaconGenesisBlock(1, genesisParamsTestnetNew),
	GenesisShardBlock:  CreateShardGenesisBlock(1, genesisParamsTestnetNew),
	BasicReward:        genesisParamsTestnetNew.BasicReward,
	RewardHalflife:     genesisParamsTestnetNew.RewardHalflife,
}

// END TESTNET

// FOR MAINNET
var genesisParamsMainnetNew = GenesisParams{
	InitialPaymentAddress:               MainnetGenesisblockPaymentAddress,
	BasicReward:                         MainnetBasicReward,
	RewardHalflife:                      MainnetRewardHalflife,
	FeePerTxKb:                          MainnetFeePerTxKb,
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
	ActiveShards:        MainNetActiveShards,
	// blockChain parameters
	GenesisBeaconBlock: CreateBeaconGenesisBlock(1, genesisParamsMainnetNew),
	GenesisShardBlock:  CreateShardGenesisBlock(1, genesisParamsMainnetNew),
	BasicReward:        genesisParamsMainnetNew.BasicReward,
	RewardHalflife:     genesisParamsMainnetNew.RewardHalflife,
}
