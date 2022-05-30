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
			MinShieldAmts: map[string]uint64{
				LocalPortalV4BTCID: 5000, // 5000 nano pbtc = 500 satoshi
			},
			MinUnshieldAmts: map[string]uint64{
				LocalPortalV4BTCID: 500000, // 500000 nano pbtc = 50000 satoshi
			},
			DustValueThreshold: map[string]uint64{
				LocalPortalV4BTCID: 10000000, // 1000000 nano pbtc = 0.01 BTC
			},
			MinUTXOsInVault: map[string]uint64{
				LocalPortalV4BTCID: 50,
			},
			BatchNumBlks:                15, // ~ 2.5 mins
			PortalReplacementAddress:    "12svfkP6w5UDJDSCwqH978PvqiqBxKmUnA9em9yAYWYJVRv7wuXY1qhhYpPAm4BDz2mLbFrRmdK3yRhnTqJCZXKHUmoi7NV83HCH2YFpctHNaDdkSiQshsjw2UFUuwdEvcidgaKmF3VJpY5f8RdN",
			MaxFeePercentageForEachStep: 20, // ~ 20% from previous fee
			TimeSpaceForFeeReplacement:  5 * time.Minute,
			MaxUnshieldFees: map[string]uint64{
				LocalPortalV4BTCID: 1000000, // 1000000 nano pbtc = 100000 satoshi
			},
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
					[]byte{0x2, 0x30, 0x34, 0xcb, 0x1a, 0x50, 0xf6, 0x7f, 0x5e, 0xb2, 0x53, 0x9e, 0x68, 0x3b, 0xd4,
						0x80, 0x73, 0x71, 0x2a, 0xdf, 0xf3, 0x25, 0x94, 0x34, 0x72, 0x6d, 0x62, 0x80, 0x83, 0xd2, 0x6f, 0x4c, 0xdd},
					[]byte{0x2, 0x74, 0x61, 0x32, 0x93, 0xe7, 0x93, 0x85, 0x94, 0xd2, 0x58, 0xfb, 0xcf, 0xc5, 0x33,
						0x78, 0xdc, 0x82, 0xcd, 0x64, 0xd1, 0xc0, 0x33, 0x1, 0x71, 0x2f, 0x90, 0x85, 0x72, 0xb9, 0x17, 0xab, 0xc7},
					[]byte{0x3, 0x67, 0x7a, 0x81, 0xfc, 0x9c, 0x4c, 0x9c, 0x6, 0x28, 0xd2, 0xf6, 0xd0, 0x1e, 0x27,
						0x15, 0xbb, 0x54, 0x11, 0x75, 0xe9, 0x62, 0xae, 0x78, 0x8f, 0xff, 0x26, 0x75, 0x1e, 0xb5, 0x24, 0xe0, 0xeb},
					[]byte{0x3, 0x2, 0xdb, 0xd4, 0xd4, 0x6b, 0x4e, 0xef, 0xe9, 0xa6, 0xe8, 0x64, 0xce, 0xeb, 0xb5,
						0x11, 0x25, 0x71, 0x28, 0x8a, 0xc4, 0xce, 0xca, 0xf4, 0x10, 0xd4, 0x16, 0x5f, 0x4c, 0x4c, 0xeb, 0x27, 0xe3},
				},
			},
			NumRequiredSigs: 3,
			GeneralMultiSigAddresses: map[string]string{
				TestnetPortalV4BTCID: "tb1qjjy5aqpf86979y6jdkvy8nwh2z3r3qtt7tr9ux0wj4lk8vydffnq7azu84",
			},
			PortalTokens: initPortalTokensV4ForTestNet(),
			DefaultFeeUnshields: map[string]uint64{
				TestnetPortalV4BTCID: 50000, // nano pbtc
			},
			MinShieldAmts: map[string]uint64{
				TestnetPortalV4BTCID: 100000, // nano pbtc
			},
			MinUnshieldAmts: map[string]uint64{
				TestnetPortalV4BTCID: 100000, // nano pbtc
			},
			DustValueThreshold: map[string]uint64{
				TestnetPortalV4BTCID: 10000000, // nano pbtc
			},
			MinUTXOsInVault: map[string]uint64{
				TestnetPortalV4BTCID: 50,
			},
			BatchNumBlks:                20,
			PortalReplacementAddress:    "12sv8WUvkvFfD5SW3aaXDSPs8yx2SxPdbv6a2LAU6FJb2kBKqmLcCuQ6ZQst4fg7THBTBtERaqMpJ7KBgsnRYobmysFEM2pbMwLE2kGzwyxgSijnZT7VQGeuUxBryC1Z6ebd8EWqDUkxwpW7Gqt8",
			MaxFeePercentageForEachStep: 10, // ~ 10% from previous fee
			TimeSpaceForFeeReplacement:  5 * time.Minute,
			MaxUnshieldFees: map[string]uint64{
				TestnetPortalV4BTCID: 100000, // pbtc
			},
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
					[]byte{0x2, 0x30, 0x34, 0xcb, 0x1a, 0x50, 0xf6, 0x7f, 0x5e, 0xb2, 0x53, 0x9e, 0x68, 0x3b, 0xd4,
						0x80, 0x73, 0x71, 0x2a, 0xdf, 0xf3, 0x25, 0x94, 0x34, 0x72, 0x6d, 0x62, 0x80, 0x83, 0xd2, 0x6f, 0x4c, 0xdd},
					[]byte{0x2, 0x74, 0x61, 0x32, 0x93, 0xe7, 0x93, 0x85, 0x94, 0xd2, 0x58, 0xfb, 0xcf, 0xc5, 0x33,
						0x78, 0xdc, 0x82, 0xcd, 0x64, 0xd1, 0xc0, 0x33, 0x1, 0x71, 0x2f, 0x90, 0x85, 0x72, 0xb9, 0x17, 0xab, 0xc7},
					[]byte{0x3, 0x67, 0x7a, 0x81, 0xfc, 0x9c, 0x4c, 0x9c, 0x6, 0x28, 0xd2, 0xf6, 0xd0, 0x1e, 0x27,
						0x15, 0xbb, 0x54, 0x11, 0x75, 0xe9, 0x62, 0xae, 0x78, 0x8f, 0xff, 0x26, 0x75, 0x1e, 0xb5, 0x24, 0xe0, 0xeb},
					[]byte{0x3, 0x2, 0xdb, 0xd4, 0xd4, 0x6b, 0x4e, 0xef, 0xe9, 0xa6, 0xe8, 0x64, 0xce, 0xeb, 0xb5,
						0x11, 0x25, 0x71, 0x28, 0x8a, 0xc4, 0xce, 0xca, 0xf4, 0x10, 0xd4, 0x16, 0x5f, 0x4c, 0x4c, 0xeb, 0x27, 0xe3},
				},
			},
			NumRequiredSigs: 3,
			GeneralMultiSigAddresses: map[string]string{
				Testnet2PortalV4BTCID: "tb1qjjy5aqpf86979y6jdkvy8nwh2z3r3qtt7tr9ux0wj4lk8vydffnq7azu84",
			},
			PortalTokens: initPortalTokensV4ForTestNet2(),
			DefaultFeeUnshields: map[string]uint64{
				Testnet2PortalV4BTCID: 50000, // nano pbtc
			},
			MinShieldAmts: map[string]uint64{
				Testnet2PortalV4BTCID: 100000, // nano pbtc
			},
			MinUnshieldAmts: map[string]uint64{
				Testnet2PortalV4BTCID: 100000, // nano pbtc
			},
			DustValueThreshold: map[string]uint64{
				Testnet2PortalV4BTCID: 10000000, // nano pbtc
			},
			MinUTXOsInVault: map[string]uint64{
				Testnet2PortalV4BTCID: 50,
			},
			BatchNumBlks:                20, //
			PortalReplacementAddress:    "12sv8WUvkvFfD5SW3aaXDSPs8yx2SxPdbv6a2LAU6FJb2kBKqmLcCuQ6ZQst4fg7THBTBtERaqMpJ7KBgsnRYobmysFEM2pbMwLE2kGzwyxgSijnZT7VQGeuUxBryC1Z6ebd8EWqDUkxwpW7Gqt8",
			MaxFeePercentageForEachStep: 10, // ~ 10% from previous fee
			TimeSpaceForFeeReplacement:  5 * time.Minute,
			MaxUnshieldFees: map[string]uint64{
				Testnet2PortalV4BTCID: 100000, // pbtc
			},
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
			MasterPubKeys: map[string][][]byte{
				MainnetPortalV4BTCID: [][]byte{
					[]byte{0x2, 0x39, 0x42, 0x3d, 0xad, 0x93, 0x8f, 0xcb, 0xe5, 0xb5, 0xef, 0x7b, 0x7b, 0x9a, 0xf, 0x28,
						0x4, 0x19, 0x53, 0x66, 0x7f, 0xee, 0x72, 0xe4, 0x81, 0xf9, 0xe6, 0xb, 0x81, 0x41, 0xd7, 0x3a, 0x36},
					[]byte{0x2, 0x8d, 0xc, 0xd7, 0x83, 0x9d, 0x5e, 0xc5, 0x7b, 0x77, 0x1a, 0xf1, 0x2, 0xb8, 0x72, 0xd0,
						0x4f, 0x34, 0xb4, 0xeb, 0x17, 0xac, 0xa1, 0x9f, 0xdf, 0xa, 0x64, 0xbf, 0xd, 0x36, 0x76, 0x66, 0x87},
					[]byte{0x3, 0x78, 0x52, 0x33, 0xe3, 0x8, 0x3a, 0xd8, 0x58, 0x77, 0x76, 0x29, 0xa0, 0x17, 0xb6, 0xdd,
						0x16, 0x43, 0x18, 0x8b, 0xb4, 0xa3, 0xaf, 0x45, 0xf0, 0xb5, 0x91, 0x8c, 0x84, 0xf2, 0x73, 0x56, 0x44},
					[]byte{0x3, 0x61, 0x9d, 0xc9, 0xfb, 0x6d, 0x8, 0x2a, 0x5c, 0x98, 0x45, 0xbc, 0xbf, 0x86, 0xfb, 0x47,
						0x4, 0xbe, 0x67, 0x46, 0xa, 0x59, 0xc4, 0xbc, 0x1d, 0xec, 0xc0, 0xe8, 0xe4, 0x3e, 0x1d, 0x6d, 0x0},
					[]byte{0x2, 0xe4, 0x1d, 0x40, 0xe6, 0xf3, 0x80, 0xad, 0x51, 0xca, 0x17, 0x87, 0xfe, 0xc8, 0x23, 0x8d,
						0xa4, 0xc2, 0x88, 0xfc, 0xfb, 0x6f, 0x2b, 0xcc, 0xd9, 0xa6, 0x1c, 0x2, 0xe5, 0x4a, 0x31, 0x34, 0x39},
					[]byte{0x2, 0xf0, 0xc, 0xe3, 0xec, 0x4, 0xdb, 0x75, 0x59, 0x99, 0x70, 0xc6, 0xfd, 0xc5, 0x2, 0x2f,
						0xad, 0x6b, 0x8d, 0x18, 0x86, 0x71, 0x44, 0xcf, 0xe6, 0x93, 0x92, 0xbb, 0xd1, 0x60, 0xc1, 0x1b, 0x5c},
					[]byte{0x2, 0x65, 0x96, 0x49, 0xab, 0xd4, 0xe5, 0x97, 0x7d, 0x5b, 0x67, 0x4c, 0x6d, 0xa1, 0xf, 0x9,
						0x28, 0xa0, 0x8c, 0x67, 0x8d, 0x7f, 0x50, 0xcc, 0x10, 0xf0, 0xfe, 0xe5, 0x68, 0xa8, 0x57, 0x63, 0xd8},
				},
			},
			NumRequiredSigs: 5,
			GeneralMultiSigAddresses: map[string]string{
				MainnetPortalV4BTCID: "bc1qmx3s84mu3wuv69dlmrtlqpuduaejxqchd6rcfm9nhhujm5hhe7hqar8l93",
			},
			PortalTokens: initPortalTokensV4ForMainNet(),
			DefaultFeeUnshields: map[string]uint64{
				MainnetPortalV4BTCID: 30000, // nano pbtc
			},
			MinShieldAmts: map[string]uint64{
				MainnetPortalV4BTCID: 100000, // nano pbtc
			},
			MinUnshieldAmts: map[string]uint64{
				MainnetPortalV4BTCID: 1000000, // nano pbtc
			},
			DustValueThreshold: map[string]uint64{
				MainnetPortalV4BTCID: 10000000, // nano pbtc ~ 0.01 BTC
			},
			MinUTXOsInVault: map[string]uint64{
				MainnetPortalV4BTCID: 500,
			},
			BatchNumBlks:                45, // ~ 30 mins
			PortalReplacementAddress:    "12sgiLdxrrmWx1qyoqxemoKdjAvUko8txG8isq3woUK73ocB4dtjaFzZVmCYQYcchzNEkptAzCK3tZF55xQvx4gcT82KzXCkMXFMbdP1A3kkhQ3NhxKpqufayLbBJ2v7MCdfkS8wvfrLXdhAXAMG",
			MaxFeePercentageForEachStep: 25, // ~ 25% from previous fee
			TimeSpaceForFeeReplacement:  5 * time.Minute,
			MaxUnshieldFees: map[string]uint64{
				MainnetPortalV4BTCID: 5000000, // pbtc
			},
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
				ExternalTxMaxSize:   5120,
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
				ExternalTxMaxSize:   5120,
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
				ExternalTxMaxSize:   51200,
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
				ExternalTxMaxSize:   51200,
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
	TestnetBTCDataFolderName = "btcrelayingv15"

	// BNB fullnode
	TestnetBNBFullNodeHost     = "data-seed-pre-0-s3.binance.org"
	TestnetBNBFullNodeProtocol = "https"
	TestnetBNBFullNodePort     = "443"
	TestnetPortalFeeder        = "12S2ciPBja9XCnEVEcsPvmCLeQH44vF8DMwSqgkH7wFETem5FiqiEpFfimETcNqDkARfht1Zpph9u5eQkjEnWsmZ5GB5vhc928EoNYH"

	// relaying header chain
	Testnet2BNBChainID        = "Binance-Chain-Ganges"
	Testnet2BTCChainID        = "Bitcoin-Testnet-2"
	Testnet2BTCDataFolderName = "btcrelayingv12"

	// BNB fullnode
	Testnet2BNBFullNodeHost     = "data-seed-pre-0-s3.binance.org"
	Testnet2BNBFullNodeProtocol = "https"
	Testnet2BNBFullNodePort     = "443"
	Testnet2PortalFeeder        = "12S2ciPBja9XCnEVEcsPvmCLeQH44vF8DMwSqgkH7wFETem5FiqiEpFfimETcNqDkARfht1Zpph9u5eQkjEnWsmZ5GB5vhc928EoNYH"

	// relaying header chain
	MainnetBNBChainID        = "Binance-Chain-Tigris"
	MainnetBTCChainID        = "Bitcoin-Mainnet"
	MainnetBTCDataFolderName = "btcrelayingv8"

	// BNB fullnode
	MainnetBNBFullNodeHost     = "dataseed1.ninicoin.io"
	MainnetBNBFullNodeProtocol = "https"
	MainnetBNBFullNodePort     = "443"
	MainnetPortalFeeder        = "12RwJVcDx4SM4PvjwwPrCRPZMMRT9g6QrnQUHD54EbtDb6AQbe26ciV6JXKyt4WRuFQVqLKqUUbb7VbWxR5V6KaG9HyFbKf6CrRxhSm"

	// portal token v4
	LocalPortalV4BTCID    = "ef5947f70ead81a76a53c7c8b7317dd5245510c665d3a13921dc9a581188728b"
	TestnetPortalV4BTCID  = "4584d5e9b2fc0337dfb17f4b5bb025e5b82c38cfa4f54e8a3d4fcdd03954ff82"
	Testnet2PortalV4BTCID = "4584d5e9b2fc0337dfb17f4b5bb025e5b82c38cfa4f54e8a3d4fcdd03954ff82"
	MainnetPortalV4BTCID  = "b832e5d3b1f01a4f0623f7fe91d6673461e1f5d37d91fe78c5c2e6183ff39696"
)
