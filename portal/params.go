package portal

import (
	"sort"
	"time"

	"github.com/incognitochain/incognito-chain/config"
	"github.com/incognitochain/incognito-chain/portal/portalrelaying"
	"github.com/incognitochain/incognito-chain/portal/portalv3"
	portalcommonv3 "github.com/incognitochain/incognito-chain/portal/portalv3/common"
	portaltokensv3 "github.com/incognitochain/incognito-chain/portal/portalv3/portaltokens"
)

type PortalParams struct {
	PortalParamsV3 map[uint64]portalv3.PortalParams
	RelayingParam  portalrelaying.RelayingParams
}

func (p PortalParams) GetPortalParamsV3(beaconHeight uint64) portalv3.PortalParams {
	portalParamMap := p.PortalParamsV3
	// only has one value - default value
	if len(portalParamMap) == 1 {
		return portalParamMap[0]
	}

	bchs := []uint64{}
	for bch := range portalParamMap {
		bchs = append(bchs, bch)
	}
	sort.Slice(bchs, func(i, j int) bool {
		return bchs[i] < bchs[j]
	})

	bchKey := bchs[len(bchs)-1]
	for i := len(bchs) - 1; i >= 0; i-- {
		if beaconHeight < bchs[i] {
			continue
		}
		bchKey = bchs[i]
		break
	}

	return portalParamMap[bchKey]
}

var p *PortalParams

func GetPortalParams() *PortalParams {
	return p
}

//SetupPortalParam Do not use this function in development or production process
// Only use for unit test
func SetupPortalParam(newPortalParam *PortalParams) {
	p = &PortalParams{}
	*p = *newPortalParam
}

func SetupParam() {
	p = new(PortalParams)
	if config.Config().IsLocal {
		*p = localPortalParam
	}
	if config.Config().IsTestNet {
		if config.Config().TestNetVersion == config.TestNetVersion2Number {
			*p = testnet2PortalParams
		}
	}
	if config.Config().IsMainNet {
		*p = mainnetPortalParam
	}
}

var localPortalParam = PortalParams{
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
}

var testnet2PortalParams = PortalParams{
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
}

var mainnetPortalParam = PortalParams{
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
}

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

func initPortalTokensV3ForTestNet2() map[string]portaltokensv3.PortalTokenProcessorV3 {
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

const (
	// relaying header chain
	TestnetBNBChainID        = "Binance-Chain-Ganges"
	TestnetBTCChainID        = "Bitcoin-Testnet"
	TestnetBTCDataFolderName = "btcrelayingv14"

	// BNB fullnode
	TestnetBNBFullNodeHost     = "data-seed-pre-0-s3.binance.org"
	TestnetBNBFullNodeProtocol = "https"
	TestnetBNBFullNodePort     = "443"
	TestnetPortalFeeder        = "12S2ciPBja9XCnEVEcsPvmCLeQH44vF8DMwSqgkH7wFETem5FiqiEpFfimETcNqDkARfht1Zpph9u5eQkjEnWsmZ5GB5vhc928EoNYH"

	// relaying header chain
	Testnet2BNBChainID        = "Binance-Chain-Ganges"
	Testnet2BTCChainID        = "Bitcoin-Testnet-2"
	Testnet2BTCDataFolderName = "btcrelayingv11"

	// BNB fullnode
	Testnet2BNBFullNodeHost     = "data-seed-pre-0-s3.binance.org"
	Testnet2BNBFullNodeProtocol = "https"
	Testnet2BNBFullNodePort     = "443"
	Testnet2PortalFeeder        = "12S2ciPBja9XCnEVEcsPvmCLeQH44vF8DMwSqgkH7wFETem5FiqiEpFfimETcNqDkARfht1Zpph9u5eQkjEnWsmZ5GB5vhc928EoNYH"

	// relaying header chain
	MainnetBNBChainID        = "Binance-Chain-Tigris"
	MainnetBTCChainID        = "Bitcoin-Mainnet"
	MainnetBTCDataFolderName = "btcrelayingv7"

	// BNB fullnode
	MainnetBNBFullNodeHost     = "dataseed1.ninicoin.io"
	MainnetBNBFullNodeProtocol = "https"
	MainnetBNBFullNodePort     = "443"
	MainnetPortalFeeder        = "12RwJVcDx4SM4PvjwwPrCRPZMMRT9g6QrnQUHD54EbtDb6AQbe26ciV6JXKyt4WRuFQVqLKqUUbb7VbWxR5V6KaG9HyFbKf6CrRxhSm"
)

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
