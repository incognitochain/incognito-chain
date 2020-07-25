package blockchain

import (
	"time"

	"github.com/incognitochain/incognito-chain/common"
)

type SlashLevel struct {
	MinRange        uint8
	PunishedEpoches uint8
}
type PortalParams struct {
	TimeOutCustodianReturnPubToken       time.Duration
	TimeOutWaitingPortingRequest         time.Duration
	TimeOutWaitingRedeemRequest          time.Duration
	MaxPercentLiquidatedCollateralAmount uint64
	MaxPercentCustodianRewards           uint64
	MinPercentCustodianRewards           uint64
	MinLockCollateralAmountInEpoch       uint64
	MinPercentLockedCollateral           uint64
	TP120                                uint64
	TP130                                uint64
	MinPercentPortingFee                 float64
	MinPercentRedeemFee                  float64
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
	GenesisParams                    *GenesisParams
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
	Epoch                            uint64
	RandomTime                       uint64
	SlashLevels                      []SlashLevel
	EthContractAddressStr            string // smart contract of ETH for bridge
	Offset                           int    // default offset for swap policy, is used for cases that good producers length is less than max committee size
	SwapOffset                       int    // is used for case that good producers length is equal to max committee size
	IncognitoDAOAddress              string
	CentralizedWebsitePaymentAddress string //centralized website's pubkey
	CheckForce                       bool   // true on testnet and false on mainnet
	ChainVersion                     string
	AssignOffset                     int
	ConsensusV2Epoch                 uint64
	BeaconHeightBreakPointBurnAddr   uint64
	BNBRelayingHeaderChainID         string
	BTCRelayingHeaderChainID         string
	BNBFullNodeProtocol              string
	BNBFullNodeHost                  string
	BNBFullNodePort                  string
	PortalParams                     map[uint64]PortalParams
	PortalFeederAddress              string
	EpochBreakPointSwapNewKey        []uint64
	IsBackup                         bool
}

type GenesisParams struct {
	InitialIncognito                            []string // init tx for genesis block
	FeePerTxKb                                  uint64
	PreSelectBeaconNodeSerializedPubkey         []string
	SelectBeaconNodeSerializedPubkeyV2          map[uint64][]string
	PreSelectBeaconNodeSerializedPaymentAddress []string
	SelectBeaconNodeSerializedPaymentAddressV2  map[uint64][]string
	PreSelectBeaconNode                         []string
	PreSelectShardNodeSerializedPubkey          []string
	SelectShardNodeSerializedPubkeyV2           map[uint64][]string
	PreSelectShardNodeSerializedPaymentAddress  []string
	SelectShardNodeSerializedPaymentAddressV2   map[uint64][]string
	PreSelectShardNode                          []string
	ConsensusAlgorithm                          string
}

var ChainTestParam = Params{}
var ChainMainParam = Params{}

var genesisParamsTestnetNew *GenesisParams
var genesisParamsMainnetNew *GenesisParams
var GenesisParam *GenesisParams

// FOR TESTNET
func init() {
	genesisParamsTestnetNew = &GenesisParams{
		PreSelectBeaconNodeSerializedPubkey:         PreSelectBeaconNodeTestnetSerializedPubkey,
		PreSelectBeaconNodeSerializedPaymentAddress: PreSelectBeaconNodeTestnetSerializedPaymentAddress,
		PreSelectShardNodeSerializedPubkey:          PreSelectShardNodeTestnetSerializedPubkey,
		PreSelectShardNodeSerializedPaymentAddress:  PreSelectShardNodeTestnetSerializedPaymentAddress,
		SelectBeaconNodeSerializedPubkeyV2:          SelectBeaconNodeTestnetSerializedPubkeyV2,
		SelectBeaconNodeSerializedPaymentAddressV2:  SelectBeaconNodeTestnetSerializedPaymentAddressV2,
		SelectShardNodeSerializedPubkeyV2:           SelectShardNodeTestnetSerializedPubkeyV2,
		SelectShardNodeSerializedPaymentAddressV2:   SelectShardNodeTestnetSerializedPaymentAddressV2,
		//@Notice: InitTxsForBenchmark is for testing and testparams only
		//InitialIncognito: IntegrationTestInitPRV,
		InitialIncognito:   TestnetInitPRV,
		ConsensusAlgorithm: common.BlsConsensus,
	}
	ChainTestParam = Params{
		Name:                   TestnetName,
		Net:                    Testnet,
		DefaultPort:            TestnetDefaultPort,
		GenesisParams:          genesisParamsTestnetNew,
		MaxShardCommitteeSize:  TestNetShardCommitteeSize,     //TestNetShardCommitteeSize,
		MinShardCommitteeSize:  TestNetMinShardCommitteeSize,  //TestNetShardCommitteeSize,
		MaxBeaconCommitteeSize: TestNetBeaconCommitteeSize,    //TestNetBeaconCommitteeSize,
		MinBeaconCommitteeSize: TestNetMinBeaconCommitteeSize, //TestNetBeaconCommitteeSize,
		StakingAmountShard:     TestNetStakingAmountShard,
		ActiveShards:           TestNetActiveShards,
		// blockChain parameters
		GenesisBeaconBlock:               CreateBeaconGenesisBlock(1, Testnet, TestnetGenesisBlockTime, genesisParamsTestnetNew),
		GenesisShardBlock:                CreateShardGenesisBlock(1, Testnet, TestnetGenesisBlockTime, genesisParamsTestnetNew),
		MinShardBlockInterval:            TestNetMinShardBlkInterval,
		MaxShardBlockCreation:            TestNetMaxShardBlkCreation,
		MinBeaconBlockInterval:           TestNetMinBeaconBlkInterval,
		MaxBeaconBlockCreation:           TestNetMaxBeaconBlkCreation,
		BasicReward:                      TestnetBasicReward,
		Epoch:                            TestnetEpoch,
		RandomTime:                       TestnetRandomTime,
		Offset:                           TestnetOffset,
		AssignOffset:                     TestnetAssignOffset,
		SwapOffset:                       TestnetSwapOffset,
		EthContractAddressStr:            TestnetETHContractAddressStr,
		IncognitoDAOAddress:              TestnetIncognitoDAOAddress,
		CentralizedWebsitePaymentAddress: TestnetCentralizedWebsitePaymentAddress,
		SlashLevels:                      []SlashLevel{
			//SlashLevel{MinRange: 20, PunishedEpoches: 1},
			//SlashLevel{MinRange: 50, PunishedEpoches: 2},
			//SlashLevel{MinRange: 75, PunishedEpoches: 3},
		},
		CheckForce:                     false,
		ChainVersion:                   "version-chain-test.json",
		ConsensusV2Epoch:               16930,
		BeaconHeightBreakPointBurnAddr: 250000,
		BNBRelayingHeaderChainID:       TestnetBNBChainID,
		BTCRelayingHeaderChainID:       TestnetBTCChainID,
		BNBFullNodeProtocol:            TestnetBNBFullNodeProtocol,
		BNBFullNodeHost:                TestnetBNBFullNodeHost,
		BNBFullNodePort:                TestnetBNBFullNodePort,
		PortalFeederAddress:            TestnetPortalFeeder,
		PortalParams: map[uint64]PortalParams{
			0: {
				TimeOutCustodianReturnPubToken:       1 * time.Hour,
				TimeOutWaitingPortingRequest:         1 * time.Hour,
				TimeOutWaitingRedeemRequest:          10 * time.Minute,
				MaxPercentLiquidatedCollateralAmount: 105,
				MaxPercentCustodianRewards:           10, // todo: need to be updated before deploying
				MinPercentCustodianRewards:           1,
				MinLockCollateralAmountInEpoch:       5000 * 1e9, // 5000 prv
				MinPercentLockedCollateral:           150,
				TP120:                                120,
				TP130:                                130,
				MinPercentPortingFee:                 0.01,
				MinPercentRedeemFee:                  0.01,
			},
		},
		EpochBreakPointSwapNewKey: TestnetReplaceCommitteeEpoch,
		IsBackup:                  false,
	}
	// END TESTNET
	// FOR MAINNET
	genesisParamsMainnetNew = &GenesisParams{
		PreSelectBeaconNodeSerializedPubkey:         PreSelectBeaconNodeMainnetSerializedPubkey,
		PreSelectBeaconNodeSerializedPaymentAddress: PreSelectBeaconNodeMainnetSerializedPaymentAddress,
		PreSelectShardNodeSerializedPubkey:          PreSelectShardNodeMainnetSerializedPubkey,
		PreSelectShardNodeSerializedPaymentAddress:  PreSelectShardNodeMainnetSerializedPaymentAddress,
		SelectBeaconNodeSerializedPubkeyV2:          SelectBeaconNodeMainnetSerializedPubkeyV2,
		SelectBeaconNodeSerializedPaymentAddressV2:  SelectBeaconNodeMainnetSerializedPaymentAddressV2,
		SelectShardNodeSerializedPubkeyV2:           SelectShardNodeMainnetSerializedPubkeyV2,
		SelectShardNodeSerializedPaymentAddressV2:   SelectShardNodeMainnetSerializedPaymentAddressV2,
		InitialIncognito:                            MainnetInitPRV,
		ConsensusAlgorithm:                          common.BlsConsensus,
	}
	ChainMainParam = Params{
		Name:                   MainetName,
		Net:                    Mainnet,
		DefaultPort:            MainnetDefaultPort,
		GenesisParams:          genesisParamsMainnetNew,
		MaxShardCommitteeSize:  MainNetShardCommitteeSize, //MainNetShardCommitteeSize,
		MinShardCommitteeSize:  MainNetMinShardCommitteeSize,
		MaxBeaconCommitteeSize: MainNetBeaconCommitteeSize, //MainNetBeaconCommitteeSize,
		MinBeaconCommitteeSize: MainNetMinBeaconCommitteeSize,
		StakingAmountShard:     MainNetStakingAmountShard,
		ActiveShards:           MainNetActiveShards,
		// blockChain parameters
		GenesisBeaconBlock:               CreateBeaconGenesisBlock(1, Mainnet, MainnetGenesisBlockTime, genesisParamsMainnetNew),
		GenesisShardBlock:                CreateShardGenesisBlock(1, Mainnet, MainnetGenesisBlockTime, genesisParamsMainnetNew),
		MinShardBlockInterval:            MainnetMinShardBlkInterval,
		MaxShardBlockCreation:            MainnetMaxShardBlkCreation,
		MinBeaconBlockInterval:           MainnetMinBeaconBlkInterval,
		MaxBeaconBlockCreation:           MainnetMaxBeaconBlkCreation,
		BasicReward:                      MainnetBasicReward,
		Epoch:                            MainnetEpoch,
		RandomTime:                       MainnetRandomTime,
		Offset:                           MainnetOffset,
		SwapOffset:                       MainnetSwapOffset,
		AssignOffset:                     MainnetAssignOffset,
		EthContractAddressStr:            MainETHContractAddressStr,
		IncognitoDAOAddress:              MainnetIncognitoDAOAddress,
		CentralizedWebsitePaymentAddress: MainnetCentralizedWebsitePaymentAddress,
		SlashLevels:                      []SlashLevel{
			//SlashLevel{MinRange: 20, PunishedEpoches: 1},
			//SlashLevel{MinRange: 50, PunishedEpoches: 2},
			//SlashLevel{MinRange: 75, PunishedEpoches: 3},
		},
		CheckForce:                     false,
		ChainVersion:                   "version-chain-main.json",
		ConsensusV2Epoch:               1e9,
		BeaconHeightBreakPointBurnAddr: 150500,
		BNBRelayingHeaderChainID:       MainnetBNBChainID,
		BTCRelayingHeaderChainID:       MainnetBTCChainID,
		BNBFullNodeProtocol:            MainnetBNBFullNodeProtocol,
		BNBFullNodeHost:                MainnetBNBFullNodeHost,
		BNBFullNodePort:                MainnetBNBFullNodePort,
		PortalFeederAddress:            MainnetPortalFeeder,
		PortalParams: map[uint64]PortalParams{
			0: {
				TimeOutCustodianReturnPubToken:       24 * time.Hour,
				TimeOutWaitingPortingRequest:         24 * time.Hour,
				TimeOutWaitingRedeemRequest:          15 * time.Minute,
				MaxPercentLiquidatedCollateralAmount: 120,
				MaxPercentCustodianRewards:           20,
				MinPercentCustodianRewards:           1,
				MinPercentLockedCollateral:           200,
				MinLockCollateralAmountInEpoch:       17500 * 1e9, // 17500 prv
				TP120:                                120,
				TP130:                                130,
				MinPercentPortingFee:                 0.01,
				MinPercentRedeemFee:                  0.01,
			},
		},
		EpochBreakPointSwapNewKey: MainnetReplaceCommitteeEpoch,
		IsBackup:                  false,
	}
	if IsTestNet {
		GenesisParam = genesisParamsTestnetNew
	} else {
		GenesisParam = genesisParamsMainnetNew
	}
}
