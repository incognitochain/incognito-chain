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

	InitialIncognito []string

	FeePerTxKb uint64

	RandomNumber uint64

	PreSelectBeaconNodeSerializedPubkey []string
	PreSelectBeaconNode                 []string
	PreSelectShardNodeSerializedPubkey  []string
	PreSelectShardNode                  []string
}

var ChainTestParam = Params{}
var ChainMainParam = Params{}

// FOR TESTNET
func init() {
	var genesisParamsTestnetNew = GenesisParams{
		InitialPaymentAddress:               TestnetGenesisBlockPaymentAddress,
		RandomNumber:                        0,
		PreSelectBeaconNodeSerializedPubkey: PreSelectBeaconNodeTestnetSerializedPubkey,
		PreSelectShardNodeSerializedPubkey:  PreSelectShardNodeTestnetSerializedPubkey,

		//@Notice: InitTxsForBenchmark is for testing and benchmark only
		//InitialIncognito: benchmark.GetInitTransaction(),
		InitialIncognito: TestnetInitPRV,
	}

	ChainTestParam = Params{
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

		InitialIncognito: MainnetInitPRV,
	}

	ChainMainParam = Params{
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
}
