package blockchain

import (
	"time"

	"github.com/incognitochain/incognito-chain/blockchain/signaturecounter"

	"github.com/incognitochain/incognito-chain/portal"
	"github.com/incognitochain/incognito-chain/portal/portalrelaying"
	"github.com/incognitochain/incognito-chain/portal/portalv3"
	portalcommonv3 "github.com/incognitochain/incognito-chain/portal/portalv3/common"
	portaltokensv3 "github.com/incognitochain/incognito-chain/portal/portalv3/portaltokens"

	"github.com/incognitochain/incognito-chain/blockchain/types"

	"github.com/incognitochain/incognito-chain/common"
)

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
	NumberOfFixedBlockValidators     int
	StakingAmountShard               uint64
	ActiveShards                     int
	GenesisBeaconBlock               *types.BeaconBlock // GenesisBlock defines the first block of the chain.
	GenesisShardBlock                *types.ShardBlock  // GenesisBlock defines the first block of the chain.
	BasicReward                      uint64
	Epoch                            uint64
	EpochV2                          uint64
	EpochV2BreakPoint                uint64
	RandomTime                       uint64
	RandomTimeV2                     uint64
	EthContractAddressStr            string // smart contract of ETH for bridge
	Offset                           int    // default offset for swap policy, is used for cases that good producers length is less than max committee size
	SwapOffset                       int    // is used for case that good producers length is equal to max committee size
	MaxSwapOrAssign                  int
	IncognitoDAOAddress              string
	CentralizedWebsitePaymentAddress string //centralized website's pubkey
	CheckForce                       bool   // true on testnet and false on mainnet
	ChainVersion                     string
	AssignOffset                     int
	ConsensusV2Epoch                 uint64
	StakingFlowV2Height              uint64
	EnableSlashingStakingFlowV2      uint64
	Timeslot                         uint64
	BeaconHeightBreakPointBurnAddr   uint64
	PortalParams                     portal.PortalParams
	EpochBreakPointSwapNewKey        []uint64
	IsBackup                         bool
	PreloadAddress                   string
	ReplaceStakingTxHeight           uint64
	ETHRemoveBridgeSigEpoch          uint64
	BCHeightBreakPointNewZKP         uint64
	MissingSignaturePenalty          []signaturecounter.Penalty
	PortalETHContractAddressStr      string // smart contract of ETH for portal
	BCHeightBreakPointPortalV3       uint64
	EnableFeatureFlags               map[int]uint64 // featureFlag: epoch number - since that time, the feature will be enabled; 0 - disabled feature
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
var ChainTest2Param = Params{}
var ChainMainParam = Params{}

var genesisParamsTestnetNew *GenesisParams
var genesisParamsTestnet2New *GenesisParams
var genesisParamsMainnetNew *GenesisParams
var GenesisParam *GenesisParams

func initPortalTokensV3ForTestNet() map[string]portaltokensv3.PortalTokenProcessorV3 {
	return map[string]portaltokensv3.PortalTokenProcessorV3{
		portalcommonv3.PortalBTCIDStr: &portaltokensv3.PortalBTCTokenProcessor{
			&portaltokensv3.PortalTokenV3{
				ChainID: TestnetBTCChainID,
			},
		},
		portalcommonv3.PortalBNBIDStr: &portaltokensv3.PortalBNBTokenProcessor{
			&portaltokensv3.PortalTokenV3{
				ChainID: TestnetBNBChainID,
			},
		},
	}
}

func initPortalTokensV3ForMainNet() map[string]portaltokensv3.PortalTokenProcessorV3 {
	return map[string]portaltokensv3.PortalTokenProcessorV3{
		portalcommonv3.PortalBTCIDStr: &portaltokensv3.PortalBTCTokenProcessor{
			&portaltokensv3.PortalTokenV3{
				ChainID: MainnetBTCChainID,
			},
		},
		portalcommonv3.PortalBNBIDStr: &portaltokensv3.PortalBNBTokenProcessor{
			&portaltokensv3.PortalTokenV3{
				ChainID: MainnetBNBChainID,
			},
		},
	}
}

// external tokenID there is no 0x prefix, in lower case
// @@Note: need to update before deploying
func getSupportedPortalCollateralsMainnet() []portalv3.PortalCollateral {
	return []portalv3.PortalCollateral{
		{"0000000000000000000000000000000000000000", 9}, // eth
		{"dac17f958d2ee523a2206206994597c13d831ec7", 6}, // usdt
		{"a0b86991c6218b36c1d19d4a2e9eb0ce3606eb48", 6}, // usdc
	}
}

// external tokenID there is no 0x prefix, in lower case
// @@Note: need to update before deploying
func getSupportedPortalCollateralsTestnet() []portalv3.PortalCollateral {
	return []portalv3.PortalCollateral{
		{"0000000000000000000000000000000000000000", 9}, // eth
		{"3a829f4b97660d970428cd370c4e41cbad62092b", 6}, // usdt, kovan testnet
		{"75b0622cec14130172eae9cf166b92e5c112faff", 6}, // usdc, kovan testnet
	}
}

// external tokenID there is no 0x prefix, in lower case
// @@Note: need to update before deploying
func getSupportedPortalCollateralsTestnet2() []portalv3.PortalCollateral {
	return []portalv3.PortalCollateral{
		{"0000000000000000000000000000000000000000", 9}, // eth
		{"3a829f4b97660d970428cd370c4e41cbad62092b", 6}, // usdt, kovan testnet
		{"75b0622cec14130172eae9cf166b92e5c112faff", 6}, // usdc, kovan testnet
	}
}

func SetupParam() {
	// FOR TESTNET
	genesisParamsTestnetNew = &GenesisParams{
		//@Notice: InitTxsForBenchmark is for testing and testparams only
		//InitialIncognito: IntegrationTestInitPRV,
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
		// GenesisBeaconBlock:               CreateGenesisBeaconBlock(1, Testnet, TestnetGenesisBlockTime, genesisParamsTestnetNew),
		// GenesisShardBlock:                CreateGenesisShardBlock(1, Testnet, TestnetGenesisBlockTime, genesisParamsTestnetNew),
		MinShardBlockInterval:            TestNetMinShardBlkInterval,
		MaxShardBlockCreation:            TestNetMaxShardBlkCreation,
		MinBeaconBlockInterval:           TestNetMinBeaconBlkInterval,
		MaxBeaconBlockCreation:           TestNetMaxBeaconBlkCreation,
		NumberOfFixedBlockValidators:     4,
		BasicReward:                      TestnetBasicReward,
		Epoch:                            TestnetEpoch,
		RandomTime:                       TestnetRandomTime,
		Offset:                           TestnetOffset,
		AssignOffset:                     TestnetAssignOffset,
		SwapOffset:                       TestnetSwapOffset,
		EthContractAddressStr:            TestnetETHContractAddressStr,
		IncognitoDAOAddress:              TestnetIncognitoDAOAddress,
		CentralizedWebsitePaymentAddress: TestnetCentralizedWebsitePaymentAddress,
		CheckForce:                       false,
		ChainVersion:                     "version-chain-test-2.json",
		ConsensusV2Epoch:                 1e9,
		StakingFlowV2Height:              2051863,
		EnableSlashingStakingFlowV2:      2087789,
		Timeslot:                         10,
		BeaconHeightBreakPointBurnAddr:   1,
		PortalParams: portal.PortalParams{
			PortalParamsV3: map[uint64]portalv3.PortalParams{
				0: {
					TimeOutCustodianReturnPubToken:       15 * time.Minute,
					TimeOutWaitingPortingRequest:         15 * time.Minute,
					TimeOutWaitingRedeemRequest:          10 * time.Minute,
					MaxPercentLiquidatedCollateralAmount: 105,
					MaxPercentCustodianRewards:           10, // todo: need to be updated before deploying
					MinPercentCustodianRewards:           1,
					MinLockCollateralAmountInEpoch:       10000 * 1e9, // 10000 usd
					MinPercentLockedCollateral:           150,
					TP120:                                120,
					TP130:                                130,
					MinPercentPortingFee:                 0.01,
					MinPercentRedeemFee:                  0.01,
					SupportedCollateralTokens:            getSupportedPortalCollateralsTestnet2(), // todo: need to be updated before deploying
					MinPortalFee:                         100,
					PortalTokens:                         initPortalTokensV3ForTestNet(),
					PortalFeederAddress:                  Testnet2PortalFeeder,
					PortalETHContractAddressStr:          "0xF7befD2806afD96D3aF76471cbCa1cD874AA1F46", // todo: update sc address,
					MinUnlockOverRateCollaterals:         25,
				},
			},
			RelayingParam: portalrelaying.RelayingParams{
				BNBRelayingHeaderChainID: Testnet2BNBChainID,
				BTCRelayingHeaderChainID: Testnet2BTCChainID,
				BTCDataFolderName:        Testnet2BTCDataFolderName,
				BNBFullNodeProtocol:      Testnet2BNBFullNodeProtocol,
				BNBFullNodeHost:          Testnet2BNBFullNodeHost,
				BNBFullNodePort:          Testnet2BNBFullNodePort,
			},
		},
		ReplaceStakingTxHeight:      1,
		IsBackup:                    false,
		PreloadAddress:              "",
		BCHeightBreakPointNewZKP:    1148608, //TODO: change this value when deployed testnet2
		ETHRemoveBridgeSigEpoch:     2085,
		EpochV2:                     Testnet2EpochV2,
		EpochV2BreakPoint:           Testnet2EpochV2BreakPoint,
		RandomTimeV2:                Testnet2RandomTimeV2,
		PortalETHContractAddressStr: "0xF7befD2806afD96D3aF76471cbCa1cD874AA1F46", // todo: update sc address
		BCHeightBreakPointPortalV3:  1328816,
		EnableFeatureFlags: map[int]uint64{
			common.PortalV3Flag:       Testnet2EnablePortalV3,
			common.PortalRelayingFlag: Testnet2EnablePortalRelaying,
		},
	}
	// END TESTNET-2

	// FOR MAINNET
	genesisParamsMainnetNew = &GenesisParams{
		InitialIncognito:   MainnetInitPRV,
		ConsensusAlgorithm: common.BlsConsensus,
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
		// GenesisBeaconBlock:               CreateGenesisBeaconBlock(1, Mainnet, MainnetGenesisBlockTime, genesisParamsMainnetNew),
		// GenesisShardBlock:                CreateGenesisShardBlock(1, Mainnet, MainnetGenesisBlockTime, genesisParamsMainnetNew),
		MinShardBlockInterval:            MainnetMinShardBlkInterval,
		MaxShardBlockCreation:            MainnetMaxShardBlkCreation,
		MinBeaconBlockInterval:           MainnetMinBeaconBlkInterval,
		MaxBeaconBlockCreation:           MainnetMaxBeaconBlkCreation,
		NumberOfFixedBlockValidators:     22,
		BasicReward:                      MainnetBasicReward,
		Epoch:                            MainnetEpoch,
		RandomTime:                       MainnetRandomTime,
		Offset:                           MainnetOffset,
		SwapOffset:                       MainnetSwapOffset,
		AssignOffset:                     MainnetAssignOffset,
		EthContractAddressStr:            MainETHContractAddressStr,
		IncognitoDAOAddress:              MainnetIncognitoDAOAddress,
		CentralizedWebsitePaymentAddress: MainnetCentralizedWebsitePaymentAddress,
		CheckForce:                       false,
		ChainVersion:                     "version-chain-main.json",
		ConsensusV2Epoch:                 1e9,
		StakingFlowV2Height:              1e12,
		EnableSlashingStakingFlowV2:      1e12,
		Timeslot:                         40,
		BeaconHeightBreakPointBurnAddr:   150500,
		PortalParams: portal.PortalParams{
			PortalParamsV3: map[uint64]portalv3.PortalParams{
				0: {
					TimeOutCustodianReturnPubToken:       24 * time.Hour,
					TimeOutWaitingPortingRequest:         24 * time.Hour,
					TimeOutWaitingRedeemRequest:          15 * time.Minute,
					MaxPercentLiquidatedCollateralAmount: 120,
					MaxPercentCustodianRewards:           20, // todo: need to be updated before deploying
					MinPercentCustodianRewards:           1,
					MinLockCollateralAmountInEpoch:       35000 * 1e9, // 35000 usd = 350 * 100
					MinPercentLockedCollateral:           200,
					TP120:                                120,
					TP130:                                130,
					MinPercentPortingFee:                 0.01,
					MinPercentRedeemFee:                  0.01,
					SupportedCollateralTokens:            getSupportedPortalCollateralsMainnet(), // todo: need to be updated before deploying
					MinPortalFee:                         100,
					PortalTokens:                         initPortalTokensV3ForMainNet(),
					PortalFeederAddress:                  MainnetPortalFeeder,
					PortalETHContractAddressStr:          "", // todo: update sc address,
					MinUnlockOverRateCollaterals:         25,
				},
			},
			RelayingParam: portalrelaying.RelayingParams{
				BNBRelayingHeaderChainID: MainnetBNBChainID,
				BTCRelayingHeaderChainID: MainnetBTCChainID,
				BTCDataFolderName:        MainnetBTCDataFolderName,
				BNBFullNodeProtocol:      MainnetBNBFullNodeProtocol,
				BNBFullNodeHost:          MainnetBNBFullNodeHost,
				BNBFullNodePort:          MainnetBNBFullNodePort,
			},
		},
		EpochBreakPointSwapNewKey:   MainnetReplaceCommitteeEpoch,
		ReplaceStakingTxHeight:      559380,
		IsBackup:                    false,
		PreloadAddress:              "",
		BCHeightBreakPointNewZKP:    934858,
		ETHRemoveBridgeSigEpoch:     1973,
		EpochV2:                     MainnetEpochV2,
		EpochV2BreakPoint:           MainnetEpochV2BreakPoint,
		RandomTimeV2:                MainnetRandomTimeV2,
		PortalETHContractAddressStr: "", // todo: update sc address
		BCHeightBreakPointPortalV3:  40, // todo: should update before deploying
		EnableFeatureFlags: map[int]uint64{
			common.PortalV3Flag:       MainnetEnablePortalV3,
			common.PortalRelayingFlag: MainnetEnablePortalRelaying,
		},
	}
}
