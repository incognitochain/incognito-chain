package portal

import (
	"sort"
	"time"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/incognitochain/incognito-chain/config"
	"github.com/incognitochain/incognito-chain/portal/portalrelaying"
	"github.com/incognitochain/incognito-chain/portal/portalv3"
	portalcommonv3 "github.com/incognitochain/incognito-chain/portal/portalv3/common"
	portaltokensv3 "github.com/incognitochain/incognito-chain/portal/portalv3/portaltokens"
	"github.com/incognitochain/incognito-chain/portal/portalv4"
	portaltokensv4 "github.com/incognitochain/incognito-chain/portal/portalv4/portaltokens"
)

type PortalParams struct {
	RelayingParam  portalrelaying.RelayingParams
	PortalParamsV3 map[uint64]portalv3.PortalParams
	PortalParamsV4 map[uint64]portalv4.PortalParams
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

func (p PortalParams) GetPortalParamsV4(beaconHeight uint64) portalv4.PortalParams {
	portalParamMap := p.PortalParamsV4
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

// SetupPortalParam Do not use this function in development or production process
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
		switch config.Config().TestNetVersion {
		case config.TestNetVersion1Number:
			*p = testnet1PortalParams
		case config.TestNetVersion2Number:
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
	PortalParamsV4: map[uint64]portalv4.PortalParams{
		0: {
			MasterPubKeys: map[string][][]byte{
				LocalPortalV4BTCID: [][]byte{
					[]byte{0x3, 0xb2, 0xd3, 0x16, 0x7d, 0x94, 0x9c, 0x25, 0x3, 0xe6, 0x9c, 0x9f, 0x29, 0x78, 0x7d, 0x9c, 0x8, 0x8d, 0x39, 0x17, 0x8d, 0xb4, 0x75, 0x40, 0x35, 0xf5, 0xae, 0x6a, 0xf0, 0x17, 0x12, 0x11, 0x0},
					[]byte{0x3, 0x98, 0x7a, 0x87, 0xd1, 0x99, 0x13, 0xbd, 0xe3, 0xef, 0xf0, 0x55, 0x79, 0x2, 0xb4, 0x90, 0x57, 0xed, 0x1c, 0x9c, 0x8b, 0x32, 0xf9, 0x2, 0xbb, 0xbb, 0x85, 0x71, 0x3a, 0x99, 0x1f, 0xdc, 0x41},
					[]byte{0x3, 0x73, 0x23, 0x5e, 0xb1, 0xc8, 0xf1, 0x84, 0xe7, 0x59, 0x17, 0x6c, 0xe3, 0x87, 0x37, 0xb7, 0x91, 0x19, 0x47, 0x1b, 0xba, 0x63, 0x56, 0xbc, 0xab, 0x8d, 0xcc, 0x14, 0x4b, 0x42, 0x99, 0x86, 0x1},
					[]byte{0x3, 0x29, 0xe7, 0x59, 0x31, 0x89, 0xca, 0x7a, 0xf6, 0x1, 0xb6, 0x35, 0x67, 0x3d, 0xb1, 0x53, 0xd4, 0x19, 0xd7, 0x6, 0x19, 0x3, 0x2a, 0x32, 0x94, 0x57, 0x76, 0xb2, 0xb3, 0x80, 0x65, 0xe1, 0x5d},
				},
			},
			NumRequiredSigs: 3,
			GeneralMultiSigAddresses: map[string]string{
				LocalPortalV4BTCID: "tb1qfgzhddwenekk573slpmqdutrd568ej89k37lmjr43tm9nhhulu0scjyajz",
			},
			PortalTokens: initPortalTokensV4ForLocal(),
			DefaultFeeUnshields: map[string]uint64{
				LocalPortalV4BTCID: 50000, // 50000 nano pbtc = 5000 satoshi
			},
			MinUnshieldAmts: map[string]uint64{
				LocalPortalV4BTCID: 500000, // 500000 nano pbtc = 50000 satoshi
			},
			DustValueThreshold: map[string]uint64{
				LocalPortalV4BTCID: 1000000, // 1000000 nano pbtc = 100000 satoshi
			},
			BatchNumBlks:                15, // ~ 2.5 mins
			PortalReplacementAddress:    "12svfkP6w5UDJDSCwqH978PvqiqBxKmUnA9em9yAYWYJVRv7wuXY1qhhYpPAm4BDz2mLbFrRmdK3yRhnTqJCZXKHUmoi7NV83HCH2YFpctHNaDdkSiQshsjw2UFUuwdEvcidgaKmF3VJpY5f8RdN",
			MaxFeePercentageForEachStep: 20, // ~ 20% from previous fee
			TimeSpaceForFeeReplacement:  5 * time.Minute,
			PortalV4TokenIDs: []string{
				LocalPortalV4BTCID,
			},
		},
	},
}

var testnet1PortalParams = PortalParams{
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
	PortalParamsV4: map[uint64]portalv4.PortalParams{
		0: {
			MasterPubKeys: map[string][][]byte{
				TestnetPortalV4BTCID: [][]byte{
					[]byte{0x3, 0xb2, 0xd3, 0x16, 0x7d, 0x94, 0x9c, 0x25, 0x3, 0xe6, 0x9c, 0x9f, 0x29, 0x78, 0x7d, 0x9c, 0x8, 0x8d, 0x39, 0x17, 0x8d, 0xb4, 0x75, 0x40, 0x35, 0xf5, 0xae, 0x6a, 0xf0, 0x17, 0x12, 0x11, 0x0},
					[]byte{0x3, 0x98, 0x7a, 0x87, 0xd1, 0x99, 0x13, 0xbd, 0xe3, 0xef, 0xf0, 0x55, 0x79, 0x2, 0xb4, 0x90, 0x57, 0xed, 0x1c, 0x9c, 0x8b, 0x32, 0xf9, 0x2, 0xbb, 0xbb, 0x85, 0x71, 0x3a, 0x99, 0x1f, 0xdc, 0x41},
					[]byte{0x3, 0x73, 0x23, 0x5e, 0xb1, 0xc8, 0xf1, 0x84, 0xe7, 0x59, 0x17, 0x6c, 0xe3, 0x87, 0x37, 0xb7, 0x91, 0x19, 0x47, 0x1b, 0xba, 0x63, 0x56, 0xbc, 0xab, 0x8d, 0xcc, 0x14, 0x4b, 0x42, 0x99, 0x86, 0x1},
					[]byte{0x3, 0x29, 0xe7, 0x59, 0x31, 0x89, 0xca, 0x7a, 0xf6, 0x1, 0xb6, 0x35, 0x67, 0x3d, 0xb1, 0x53, 0xd4, 0x19, 0xd7, 0x6, 0x19, 0x3, 0x2a, 0x32, 0x94, 0x57, 0x76, 0xb2, 0xb3, 0x80, 0x65, 0xe1, 0x5d},
				},
			},
			NumRequiredSigs: 3,
			GeneralMultiSigAddresses: map[string]string{
				TestnetPortalV4BTCID: "tb1qfgzhddwenekk573slpmqdutrd568ej89k37lmjr43tm9nhhulu0scjyajz",
			},
			PortalTokens: initPortalTokensV4ForTestNet(),
			DefaultFeeUnshields: map[string]uint64{
				TestnetPortalV4BTCID: 50000, // 50000 nano pbtc = 5000 satoshi
			},
			MinUnshieldAmts: map[string]uint64{
				TestnetPortalV4BTCID: 500000, // 500000 nano pbtc = 50000 satoshi
			},
			DustValueThreshold: map[string]uint64{
				TestnetPortalV4BTCID: 1000000, // 1000000 nano pbtc = 100000 satoshi
			},
			BatchNumBlks:                15, // ~ 2.5 mins
			PortalReplacementAddress:    "12svfkP6w5UDJDSCwqH978PvqiqBxKmUnA9em9yAYWYJVRv7wuXY1qhhYpPAm4BDz2mLbFrRmdK3yRhnTqJCZXKHUmoi7NV83HCH2YFpctHNaDdkSiQshsjw2UFUuwdEvcidgaKmF3VJpY5f8RdN",
			MaxFeePercentageForEachStep: 20, // ~ 20% from previous fee
			TimeSpaceForFeeReplacement:  5 * time.Minute,
			PortalV4TokenIDs: []string{
				TestnetPortalV4BTCID,
			},
		},
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
			PortalTokens:                         initPortalTokensV3ForTestNet2(),
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
	PortalParamsV4: map[uint64]portalv4.PortalParams{
		0: {
			MasterPubKeys: map[string][][]byte{
				Testnet2PortalV4BTCID: [][]byte{
					[]byte{0x3, 0xb2, 0xd3, 0x16, 0x7d, 0x94, 0x9c, 0x25, 0x3, 0xe6, 0x9c, 0x9f, 0x29, 0x78, 0x7d, 0x9c, 0x8, 0x8d, 0x39, 0x17, 0x8d, 0xb4, 0x75, 0x40, 0x35, 0xf5, 0xae, 0x6a, 0xf0, 0x17, 0x12, 0x11, 0x0},
					[]byte{0x3, 0x98, 0x7a, 0x87, 0xd1, 0x99, 0x13, 0xbd, 0xe3, 0xef, 0xf0, 0x55, 0x79, 0x2, 0xb4, 0x90, 0x57, 0xed, 0x1c, 0x9c, 0x8b, 0x32, 0xf9, 0x2, 0xbb, 0xbb, 0x85, 0x71, 0x3a, 0x99, 0x1f, 0xdc, 0x41},
					[]byte{0x3, 0x73, 0x23, 0x5e, 0xb1, 0xc8, 0xf1, 0x84, 0xe7, 0x59, 0x17, 0x6c, 0xe3, 0x87, 0x37, 0xb7, 0x91, 0x19, 0x47, 0x1b, 0xba, 0x63, 0x56, 0xbc, 0xab, 0x8d, 0xcc, 0x14, 0x4b, 0x42, 0x99, 0x86, 0x1},
					[]byte{0x3, 0x29, 0xe7, 0x59, 0x31, 0x89, 0xca, 0x7a, 0xf6, 0x1, 0xb6, 0x35, 0x67, 0x3d, 0xb1, 0x53, 0xd4, 0x19, 0xd7, 0x6, 0x19, 0x3, 0x2a, 0x32, 0x94, 0x57, 0x76, 0xb2, 0xb3, 0x80, 0x65, 0xe1, 0x5d},
				},
			},
			NumRequiredSigs: 3,
			GeneralMultiSigAddresses: map[string]string{
				Testnet2PortalV4BTCID: "tb1qfgzhddwenekk573slpmqdutrd568ej89k37lmjr43tm9nhhulu0scjyajz",
			},
			PortalTokens: initPortalTokensV4ForTestNet2(),
			DefaultFeeUnshields: map[string]uint64{
				Testnet2PortalV4BTCID: 50000, // 50000 nano pbtc = 5000 satoshi
			},
			MinUnshieldAmts: map[string]uint64{
				Testnet2PortalV4BTCID: 500000, // 500000 nano pbtc = 50000 satoshi
			},
			DustValueThreshold: map[string]uint64{
				Testnet2PortalV4BTCID: 1000000, // 1000000 nano pbtc = 100000 satoshi
			},
			BatchNumBlks:                15, // ~ 2.5 mins
			PortalReplacementAddress:    "12svfkP6w5UDJDSCwqH978PvqiqBxKmUnA9em9yAYWYJVRv7wuXY1qhhYpPAm4BDz2mLbFrRmdK3yRhnTqJCZXKHUmoi7NV83HCH2YFpctHNaDdkSiQshsjw2UFUuwdEvcidgaKmF3VJpY5f8RdN",
			MaxFeePercentageForEachStep: 20, // ~ 20% from previous fee
			TimeSpaceForFeeReplacement:  5 * time.Minute,
			PortalV4TokenIDs: []string{
				Testnet2PortalV4BTCID,
			},
		},
	},
}

// should update param before deploying production
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
	PortalParamsV4: map[uint64]portalv4.PortalParams{
		0: {
			MasterPubKeys:   map[string][][]byte{},
			NumRequiredSigs: 3,
			GeneralMultiSigAddresses: map[string]string{
				MainnetPortalV4BTCID: "tb1qfgzhddwenekk573slpmqdutrd568ej89k37lmjr43tm9nhhulu0scjyajz",
			},
			PortalTokens: initPortalTokensV4ForMainNet(),
			DefaultFeeUnshields: map[string]uint64{
				MainnetPortalV4BTCID: 50000, // 50000 nano pbtc = 5000 satoshi
			},
			MinUnshieldAmts: map[string]uint64{
				MainnetPortalV4BTCID: 500000, // 500000 nano pbtc = 50000 satoshi
			},
			DustValueThreshold: map[string]uint64{
				MainnetPortalV4BTCID: 1000000, // 1000000 nano pbtc = 100000 satoshi
			},
			BatchNumBlks:                15, // ~ 2.5 mins
			PortalReplacementAddress:    "12svfkP6w5UDJDSCwqH978PvqiqBxKmUnA9em9yAYWYJVRv7wuXY1qhhYpPAm4BDz2mLbFrRmdK3yRhnTqJCZXKHUmoi7NV83HCH2YFpctHNaDdkSiQshsjw2UFUuwdEvcidgaKmF3VJpY5f8RdN",
			MaxFeePercentageForEachStep: 20, // ~ 20% from previous fee
			TimeSpaceForFeeReplacement:  5 * time.Minute,
			PortalV4TokenIDs: []string{
				MainnetPortalV4BTCID,
			},
		},
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

func initPortalTokensV4ForLocal() map[string]portaltokensv4.PortalTokenProcessor {
	return map[string]portaltokensv4.PortalTokenProcessor{
		LocalPortalV4BTCID: portaltokensv4.PortalBTCTokenProcessor{
			PortalToken: &portaltokensv4.PortalToken{
				ChainID:             TestnetBTCChainID,
				MinTokenAmount:      10,
				MultipleTokenAmount: 10,
				ExternalInputSize:   130,
				ExternalOutputSize:  43,
				ExternalTxMaxSize:   2048,
			},
			ChainParam:    &chaincfg.TestNet3Params,
			PortalTokenID: LocalPortalV4BTCID,
		},
	}
}

func initPortalTokensV4ForTestNet() map[string]portaltokensv4.PortalTokenProcessor {
	return map[string]portaltokensv4.PortalTokenProcessor{
		TestnetPortalV4BTCID: portaltokensv4.PortalBTCTokenProcessor{
			PortalToken: &portaltokensv4.PortalToken{
				ChainID:             TestnetBTCChainID,
				MinTokenAmount:      10,
				MultipleTokenAmount: 10,
				ExternalInputSize:   130,
				ExternalOutputSize:  43,
				ExternalTxMaxSize:   2048,
			},
			ChainParam:    &chaincfg.TestNet3Params,
			PortalTokenID: TestnetPortalV4BTCID,
		},
	}
}

func initPortalTokensV4ForTestNet2() map[string]portaltokensv4.PortalTokenProcessor {
	return map[string]portaltokensv4.PortalTokenProcessor{
		Testnet2PortalV4BTCID: portaltokensv4.PortalBTCTokenProcessor{
			PortalToken: &portaltokensv4.PortalToken{
				ChainID:             Testnet2BTCChainID,
				MinTokenAmount:      10,
				MultipleTokenAmount: 10,
				ExternalInputSize:   130,
				ExternalOutputSize:  43,
				ExternalTxMaxSize:   2048,
			},
			ChainParam:    &chaincfg.TestNet3Params,
			PortalTokenID: Testnet2PortalV4BTCID,
		},
	}
}

func initPortalTokensV4ForMainNet() map[string]portaltokensv4.PortalTokenProcessor {
	return map[string]portaltokensv4.PortalTokenProcessor{
		MainnetPortalV4BTCID: portaltokensv4.PortalBTCTokenProcessor{
			PortalToken: &portaltokensv4.PortalToken{
				ChainID:             MainnetBTCChainID,
				MinTokenAmount:      10,
				MultipleTokenAmount: 10,
				ExternalInputSize:   192,
				ExternalOutputSize:  43,
				ExternalTxMaxSize:   2048,
			},
			ChainParam:    &chaincfg.MainNetParams,
			PortalTokenID: MainnetPortalV4BTCID,
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

	// @@NOTE: should update tokenID before deploying
	// portal token v4
	LocalPortalV4BTCID    = "ef5947f70ead81a76a53c7c8b7317dd5245510c665d3a13921dc9a581188728b"
	TestnetPortalV4BTCID  = "4584d5e9b2fc0337dfb17f4b5bb025e5b82c38cfa4f54e8a3d4fcdd03954ff82"
	Testnet2PortalV4BTCID = "4584d5e9b2fc0337dfb17f4b5bb025e5b82c38cfa4f54e8a3d4fcdd03954ff82"
	MainnetPortalV4BTCID  = "b832e5d3b1f01a4f0623f7fe91d6673461e1f5d37d91fe78c5c2e6183ff39696"
)
