package blockchain

import (
	"time"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/incognitochain/incognito-chain/blockchain/signaturecounter"
	"github.com/incognitochain/incognito-chain/blockchain/types"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/portal"
	"github.com/incognitochain/incognito-chain/portal/portalrelaying"
	"github.com/incognitochain/incognito-chain/portal/portalv3"
	portalcommonv3 "github.com/incognitochain/incognito-chain/portal/portalv3/common"
	portaltokensv3 "github.com/incognitochain/incognito-chain/portal/portalv3/portaltokens"
	"github.com/incognitochain/incognito-chain/portal/portalv4"
	portalcommonv4 "github.com/incognitochain/incognito-chain/portal/portalv4/common"
	portaltokensv4 "github.com/incognitochain/incognito-chain/portal/portalv4/portaltokens"
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
	SlashLevels                      []SlashLevel
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
	BCHeightBreakPointPrivacyV2      uint64
	EnableFeatureFlags               map[int]uint64 // featureFlag: epoch number - since that time, the feature will be enabled; 0 - disabled feature
	CoinVersion2LowestHeight         uint64
	WhiteListTxs                     map[string]bool
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

func initPortalTokensV4ForTestNet() map[string]portaltokensv4.PortalTokenProcessor {
	return map[string]portaltokensv4.PortalTokenProcessor{
		portalcommonv4.PortalBTCIDStr: portaltokensv4.PortalBTCTokenProcessor{
			PortalToken: &portaltokensv4.PortalToken{
				ChainID:             TestnetBTCChainID,
				MinTokenAmount:      10,
				MultipleTokenAmount: 10,
				ExternalInputSize:   130,
				ExternalOutputSize:  43,
				ExternalTxMaxSize:   2048,
			},
			ChainParam: &chaincfg.TestNet3Params,
		},
	}
}

func initPortalTokensV4ForMainNet() map[string]portaltokensv4.PortalTokenProcessor {
	return map[string]portaltokensv4.PortalTokenProcessor{
		portalcommonv4.PortalBTCIDStr: portaltokensv4.PortalBTCTokenProcessor{
			PortalToken: &portaltokensv4.PortalToken{
				ChainID:             MainnetBTCChainID,
				MinTokenAmount:      10,
				MultipleTokenAmount: 10,
				ExternalInputSize:   192,
				ExternalOutputSize:  43,
				ExternalTxMaxSize:   2048,
			},
			ChainParam: &chaincfg.MainNetParams,
		},
	}
}

func SetupParam() {
	// FOR TESTNET
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
		SlashLevels:                      []SlashLevel{
			//SlashLevel{MinRange: 20, PunishedEpoches: 1},
			//SlashLevel{MinRange: 50, PunishedEpoches: 2},
			//SlashLevel{MinRange: 75, PunishedEpoches: 3},
		},
		CheckForce:                     false,
		ChainVersion:                   "version-chain-test.json",
		ConsensusV2Epoch:               1,
		StakingFlowV2Height:            3016278,
		EnableSlashingStakingFlowV2:    3016778,
		Timeslot:                       10,
		BeaconHeightBreakPointBurnAddr: 250000,
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
					SupportedCollateralTokens:            getSupportedPortalCollateralsTestnet(), // todo: need to be updated before deploying
					MinPortalFee:                         100,
					PortalTokens:                         initPortalTokensV3ForTestNet(),
					PortalFeederAddress:                  TestnetPortalFeeder,
					PortalETHContractAddressStr:          "0x6D53de7aFa363F779B5e125876319695dC97171E", // todo: update sc address,
					MinUnlockOverRateCollaterals:         25,
				},
			},
			RelayingParam: portalrelaying.RelayingParams{
				BNBRelayingHeaderChainID: TestnetBNBChainID,
				BTCRelayingHeaderChainID: TestnetBTCChainID,
				BTCDataFolderName:        TestnetBTCDataFolderName,
				BNBFullNodeProtocol:      TestnetBNBFullNodeProtocol,
				BNBFullNodeHost:          TestnetBNBFullNodeHost,
				BNBFullNodePort:          TestnetBNBFullNodePort,
			},
			// todo: should update before deploying
			PortalParamsV4: map[uint64]portalv4.PortalParams{
				0: {
					MasterPubKeys: map[string][][]byte{
						portalcommonv4.PortalBTCIDStr: [][]byte{
							[]byte{0x3, 0xb2, 0xd3, 0x16, 0x7d, 0x94, 0x9c, 0x25, 0x3, 0xe6, 0x9c, 0x9f, 0x29, 0x78, 0x7d, 0x9c, 0x8, 0x8d, 0x39, 0x17, 0x8d, 0xb4, 0x75, 0x40, 0x35, 0xf5, 0xae, 0x6a, 0xf0, 0x17, 0x12, 0x11, 0x0},
							[]byte{0x3, 0x98, 0x7a, 0x87, 0xd1, 0x99, 0x13, 0xbd, 0xe3, 0xef, 0xf0, 0x55, 0x79, 0x2, 0xb4, 0x90, 0x57, 0xed, 0x1c, 0x9c, 0x8b, 0x32, 0xf9, 0x2, 0xbb, 0xbb, 0x85, 0x71, 0x3a, 0x99, 0x1f, 0xdc, 0x41},
							[]byte{0x3, 0x73, 0x23, 0x5e, 0xb1, 0xc8, 0xf1, 0x84, 0xe7, 0x59, 0x17, 0x6c, 0xe3, 0x87, 0x37, 0xb7, 0x91, 0x19, 0x47, 0x1b, 0xba, 0x63, 0x56, 0xbc, 0xab, 0x8d, 0xcc, 0x14, 0x4b, 0x42, 0x99, 0x86, 0x1},
							[]byte{0x3, 0x29, 0xe7, 0x59, 0x31, 0x89, 0xca, 0x7a, 0xf6, 0x1, 0xb6, 0x35, 0x67, 0x3d, 0xb1, 0x53, 0xd4, 0x19, 0xd7, 0x6, 0x19, 0x3, 0x2a, 0x32, 0x94, 0x57, 0x76, 0xb2, 0xb3, 0x80, 0x65, 0xe1, 0x5d},
						},
					},
					NumRequiredSigs: 3,
					GeneralMultiSigAddresses: map[string]string{
						portalcommonv4.PortalBTCIDStr: "tb1qfgzhddwenekk573slpmqdutrd568ej89k37lmjr43tm9nhhulu0scjyajz",
					},
					PortalTokens: initPortalTokensV4ForTestNet(),
					DefaultFeeUnshields: map[string]uint64{
						portalcommonv4.PortalBTCIDStr: 50000, // 50000 nano pbtc = 5000 satoshi
					},
					MinUnshieldAmts: map[string]uint64{
						portalcommonv4.PortalBTCIDStr: 500000, // 500000 nano pbtc = 50000 satoshi
					},
					DustValueThreshold: map[string]uint64{
						portalcommonv4.PortalBTCIDStr: 1000000, // 1000000 nano pbtc = 100000 satoshi
					},
					BatchNumBlks:                15, // ~ 2.5 mins
					MinConfirmationIncBlockNum:  3,
					PortalReplacementAddress:    "12svfkP6w5UDJDSCwqH978PvqiqBxKmUnA9em9yAYWYJVRv7wuXY1qhhYpPAm4BDz2mLbFrRmdK3yRhnTqJCZXKHUmoi7NV83HCH2YFpctHNaDdkSiQshsjw2UFUuwdEvcidgaKmF3VJpY5f8RdN",
					MaxFeePercentageForEachStep: 20, // ~ 20% from previous fee
					TimeSpaceForFeeReplacement:  5 * time.Minute,
				},
			},
		},
		EpochBreakPointSwapNewKey:   TestnetReplaceCommitteeEpoch,
		ReplaceStakingTxHeight:      1,
		IsBackup:                    false,
		PreloadAddress:              "",
		BCHeightBreakPointNewZKP:    2300000, //TODO: change this value when deployed testnet
		ETHRemoveBridgeSigEpoch:     21920,
		EpochV2:                     TestnetEpochV2,
		EpochV2BreakPoint:           TestnetEpochV2BreakPoint,
		RandomTimeV2:                TestnetRandomTimeV2,
		PortalETHContractAddressStr: "0x6D53de7aFa363F779B5e125876319695dC97171E", // todo: update sc address
		BCHeightBreakPointPortalV3:  30158,
		BCHeightBreakPointPrivacyV2: 30, //TODO: change this value when deployed testnet
		CoinVersion2LowestHeight:    10,
		EnableFeatureFlags: map[int]uint64{
			common.PortalV3Flag:       TestnetEnablePortalV3,
			common.PortalRelayingFlag: TestnetEnablePortalRelaying,
			common.PortalV4Flag:       TestnetEnablePortalV4,
		},
		WhiteListTxs: map[string]bool{},
	}
	// END TESTNET

	// FOR TESTNET-2
	genesisParamsTestnet2New = &GenesisParams{
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
	ChainTest2Param = Params{
		Name:                   Testnet2Name,
		Net:                    Testnet2,
		DefaultPort:            Testnet2DefaultPort,
		GenesisParams:          genesisParamsTestnet2New,
		MaxShardCommitteeSize:  TestNet2ShardCommitteeSize,     //TestNetShardCommitteeSize,
		MinShardCommitteeSize:  TestNet2MinShardCommitteeSize,  //TestNetShardCommitteeSize,
		MaxBeaconCommitteeSize: TestNet2BeaconCommitteeSize,    //TestNetBeaconCommitteeSize,
		MinBeaconCommitteeSize: TestNet2MinBeaconCommitteeSize, //TestNetBeaconCommitteeSize,
		StakingAmountShard:     TestNet2StakingAmountShard,
		ActiveShards:           TestNet2ActiveShards,
		// blockChain parameters
		// GenesisBeaconBlock:               CreateGenesisBeaconBlock(1, Testnet2, Testnet2GenesisBlockTime, genesisParamsTestnet2New),
		// GenesisShardBlock:                CreateGenesisShardBlock(1, Testnet2, Testnet2GenesisBlockTime, genesisParamsTestnet2New),
		MinShardBlockInterval:            TestNet2MinShardBlkInterval,
		MaxShardBlockCreation:            TestNet2MaxShardBlkCreation,
		MinBeaconBlockInterval:           TestNet2MinBeaconBlkInterval,
		MaxBeaconBlockCreation:           TestNet2MaxBeaconBlkCreation,
		NumberOfFixedBlockValidators:     4,
		BasicReward:                      Testnet2BasicReward,
		Epoch:                            Testnet2Epoch,
		RandomTime:                       Testnet2RandomTime,
		Offset:                           Testnet2Offset,
		AssignOffset:                     Testnet2AssignOffset,
		SwapOffset:                       Testnet2SwapOffset,
		EthContractAddressStr:            Testnet2ETHContractAddressStr,
		IncognitoDAOAddress:              Testnet2IncognitoDAOAddress,
		CentralizedWebsitePaymentAddress: Testnet2CentralizedWebsitePaymentAddress,
		SlashLevels:                      []SlashLevel{
			//SlashLevel{MinRange: 20, PunishedEpoches: 1},
			//SlashLevel{MinRange: 50, PunishedEpoches: 2},
			//SlashLevel{MinRange: 75, PunishedEpoches: 3},
		},
		CheckForce:                     false,
		ChainVersion:                   "version-chain-test-2.json",
		ConsensusV2Epoch:               1e9,
		StakingFlowV2Height:            2051863,
		EnableSlashingStakingFlowV2:    2087789,
		Timeslot:                       10,
		BeaconHeightBreakPointBurnAddr: 1,
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
			// todo: should update before deploying
			PortalParamsV4: map[uint64]portalv4.PortalParams{
				0: {
					MasterPubKeys: map[string][][]byte{
						portalcommonv4.PortalBTCIDStr: [][]byte{}, // 50000 nano pbtc = 5000 satoshi
					},
					NumRequiredSigs: 3,
					GeneralMultiSigAddresses: map[string]string{
						portalcommonv4.PortalBTCIDStr: "",
					},
					PortalTokens: initPortalTokensV4ForTestNet(),
					DefaultFeeUnshields: map[string]uint64{
						portalcommonv4.PortalBTCIDStr: 100000, // 100000 nano pbtc = 10000 satoshi
					},
					MinUnshieldAmts: map[string]uint64{
						portalcommonv4.PortalBTCIDStr: 1000000, // 1000000 nano pbtc = 100000 satoshi
					},
					BatchNumBlks:               60, // ~ 10 mins
					MinConfirmationIncBlockNum: 3,
					DustValueThreshold: map[string]uint64{
						portalcommonv4.PortalBTCIDStr: 1000000, // 1000000 nano pbtc = 100000 satoshi
					},
					PortalReplacementAddress:    "12S5Lrs1XeQLbqN4ySyKtjAjd2d7sBP2tjFijzmp6avrrkQCNFMpkXm3FPzj2Wcu2ZNqJEmh9JriVuRErVwhuQnLmWSaggobEWsBEci",
					MaxFeePercentageForEachStep: 20, // ~ 20% from previous fee
					TimeSpaceForFeeReplacement:  5 * time.Minute,
				},
			},
		},
		EpochBreakPointSwapNewKey:   TestnetReplaceCommitteeEpoch,
		ReplaceStakingTxHeight:      1,
		IsBackup:                    false,
		PreloadAddress:              "",
		BCHeightBreakPointNewZKP:    1148608, //TODO: change this value when deployed testnet2
		ETHRemoveBridgeSigEpoch:     2085,
		EpochV2:                     Testnet2EpochV2,
		EpochV2BreakPoint:           Testnet2EpochV2BreakPoint,
		RandomTimeV2:                Testnet2RandomTimeV2,
		PortalETHContractAddressStr: "0xF7befD2806afD96D3aF76471cbCa1cD874AA1F46", // todo: update sc address
		BCHeightBreakPointPortalV3:  8974,                                         // todo: should update before deploying
		BCHeightBreakPointPrivacyV2: 30,                                           //TODO: change this value when deployed testnet2
		CoinVersion2LowestHeight:    10,
		EnableFeatureFlags: map[int]uint64{
			common.PortalV3Flag:       Testnet2EnablePortalV3,
			common.PortalRelayingFlag: Testnet2EnablePortalRelaying,
			common.PortalV4Flag:       Testnet2EnablePortalV4,
		},
		WhiteListTxs: Testnet2WhiteListTxs,
	}
	// END TESTNET-2

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
		SlashLevels:                      []SlashLevel{
			//SlashLevel{MinRange: 20, PunishedEpoches: 1},
			//SlashLevel{MinRange: 50, PunishedEpoches: 2},
			//SlashLevel{MinRange: 75, PunishedEpoches: 3},
		},
		CheckForce:                     false,
		ChainVersion:                   "version-chain-main.json",
		ConsensusV2Epoch:               1e9,
		StakingFlowV2Height:            1e12,
		EnableSlashingStakingFlowV2:    1e12,
		Timeslot:                       40,
		BeaconHeightBreakPointBurnAddr: 150500,
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
			// todo: should update before deploying
			PortalParamsV4: map[uint64]portalv4.PortalParams{
				0: {
					MasterPubKeys: map[string][][]byte{
						portalcommonv4.PortalBTCIDStr: [][]byte{}, // 50000 nano pbtc = 5000 satoshi
					},
					NumRequiredSigs: 3,
					GeneralMultiSigAddresses: map[string]string{
						portalcommonv4.PortalBTCIDStr: "",
					},
					PortalTokens: initPortalTokensV4ForMainNet(),
					DefaultFeeUnshields: map[string]uint64{
						portalcommonv4.PortalBTCIDStr: 100000, // 100000 nano pbtc = 10000 satoshi
					},
					MinUnshieldAmts: map[string]uint64{
						portalcommonv4.PortalBTCIDStr: 1000000, // 1000000 nano pbtc = 100000 satoshi
					},
					BatchNumBlks:               60, // ~ 10 mins
					MinConfirmationIncBlockNum: 3,
					DustValueThreshold: map[string]uint64{
						portalcommonv4.PortalBTCIDStr: 1000000, // 1000000 nano pbtc = 100000 satoshi
					},
				},
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
		PortalETHContractAddressStr: "",      // todo: update sc address
		BCHeightBreakPointPortalV3:  40,      // todo: should update before deploying
		BCHeightBreakPointPrivacyV2: 1500000, //TODO: change this value when deployed mainnet
		CoinVersion2LowestHeight:    900000,  //TODO
		EnableFeatureFlags: map[int]uint64{
			common.PortalV3Flag:       MainnetEnablePortalV3,
			common.PortalRelayingFlag: MainnetEnablePortalRelaying,
			common.PortalV4Flag:       MainnetEnablePortalV4,
		},
		WhiteListTxs: map[string]bool{},
	}
	if IsTestNet {
		if !IsTestNet2 {
			GenesisParam = genesisParamsTestnetNew
		} else {
			GenesisParam = genesisParamsTestnet2New
		}
	} else {
		GenesisParam = genesisParamsMainnetNew
	}
}

func (p *Params) CreateGenesisBlocks() {
	blockTime := ""
	switch p.Net {
	case Mainnet:
		blockTime = MainnetGenesisBlockTime
	case Testnet:
		blockTime = TestnetGenesisBlockTime
	case Testnet2:
		blockTime = Testnet2GenesisBlockTime
	}
	p.GenesisBeaconBlock = CreateGenesisBeaconBlock(1, uint16(p.Net), blockTime, p.GenesisParams)
	p.GenesisShardBlock = CreateGenesisShardBlock(1, uint16(p.Net), blockTime, p.GenesisParams)
	return
}
