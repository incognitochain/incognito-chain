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
}

type GenesisParams struct {
	InitialPaymentAddress string
	InitFundSalary        uint64
	InitialDCBToken       uint64
	InitialGOVToken       uint64
	InitialBondToken      uint64

	InitialConstant []string

	SalaryPerTx uint64
	BasicSalary uint64
	FeePerTxKb  uint64

	RandomNumber uint64

	PreSelectBeaconNodeSerializedPubkey []string
	PreSelectBeaconNode                 []string
	PreSelectShardNodeSerializedPubkey  []string
	PreSelectShardNode                  []string
}

// FOR TESTNET
var genesisParamsTestnetNew = GenesisParams{
	InitialPaymentAddress:               TestnetGenesisBlockPaymentAddress,
	InitFundSalary:                      TestnetInitFundSalary,
	InitialBondToken:                    TestnetInitBondToken,
	InitialDCBToken:                     TestnetInitDCBToken,
	InitialGOVToken:                     TestnetInitGovToken,
	BasicSalary:                         TestnetBasicSalary,
	SalaryPerTx:                         TestnetSalaryPerTx,
	FeePerTxKb:                          TestnetFeePerTxKb,
	RandomNumber:                        0,
	PreSelectBeaconNodeSerializedPubkey: PreSelectBeaconNodeTestnetSerializedPubkey,
	PreSelectShardNodeSerializedPubkey:  PreSelectShardNodeTestnetSerializedPubkey,

	//@Notice: InitTxsForBenchmark is for testing and benchmark only
	InitialConstant: TestnetInitConstant,
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
}

// END TESTNET

// FOR MAINNET
var genesisParamsMainnetNew = GenesisParams{
	InitialPaymentAddress:               MainnetGenesisblockPaymentAddress,
	InitFundSalary:                      MainnetInitFundSalary,
	InitialBondToken:                    MainnetInitBondToken,
	InitialDCBToken:                     MainnetInitDCBToken,
	InitialGOVToken:                     MainnetInitGovToken,
	BasicSalary:                         MainnetBasicSalary,
	SalaryPerTx:                         MainnetSalaryPerTx,
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
}
