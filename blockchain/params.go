package blockchain

/*
Params defines a network by its params. These params may be used by Applications
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

type IcoParams struct {
	InitialPaymentAddress string
	InitFundSalary        uint64
	InitialDCBToken       uint64
	InitialCMBToken       uint64
	InitialGOVToken       uint64
	InitialBondToken      uint64
	InitialVoteDCBToken   uint64
	InitialVoteGOVToken   uint64
}

// FOR TESTNET
var icoParamsTestnetNew = IcoParams{
	InitialPaymentAddress: TestnetGenesisBlockPaymentAddress,
	InitFundSalary:        TestnetInitFundSalary,
	InitialBondToken:      TestnetInitBondToken,
	InitialCMBToken:       TestnetInitCmBToken,
	InitialDCBToken:       TestnetInitDCBToken,
	InitialGOVToken:       TestnetInitGovToken,
}

var ChainTestParam = Params{
	Name:                TestnetName,
	Net:                 Testnet,
	DefaultPort:         TestnetDefaultPort,
	ShardCommitteeSize:  TestNetShardCommitteeSize,  //TestNetShardCommitteeSize,
	BeaconCommitteeSize: TestNetBeaconCommitteeSize, //TestNetBeaconCommitteeSize,
	ActiveShards:        TestNetActiveShards,
	// blockChain parameters
	GenesisBeaconBlock: CreateBeaconGenesisBlock(1, icoParamsTestnetNew, 1000, 1000, 0),
	GenesisShardBlock:  CreateShardGenesisBlock(1, icoParamsTestnetNew),
}
// END TESTNET

// FOR MAINNET
var icoParamsMainnetNew = IcoParams{
	InitialPaymentAddress: MainnetGenesisblockPaymentAddress,
	InitFundSalary:        MainnetInitFundSalary,
	InitialBondToken:      MainnetInitBondToken,
	InitialCMBToken:       MainnetInitCmBToken,
	InitialDCBToken:       MainnetInitDCBToken,
	InitialGOVToken:       MainnetInitGovToken,
}

var ChainMainParam = Params{
	Name:                MainetName,
	Net:                 Mainnet,
	DefaultPort:         MainnetDefaultPort,
	ShardCommitteeSize:  MainNetShardCommitteeSize,  //MainNetShardCommitteeSize,
	BeaconCommitteeSize: MainNetBeaconCommitteeSize, //MainNetBeaconCommitteeSize,
	ActiveShards:        MainNetActiveShards,
	// blockChain parameters
	GenesisBeaconBlock: CreateBeaconGenesisBlock(1, icoParamsMainnetNew, 1000, 1000, 0),
	GenesisShardBlock:  CreateShardGenesisBlock(1, icoParamsMainnetNew),
}
