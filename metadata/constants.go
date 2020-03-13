package metadata

import "github.com/incognitochain/incognito-chain/common"

const (
	InvalidMeta = 1

	IssuingRequestMeta     = 24
	IssuingResponseMeta    = 25
	ContractingRequestMeta = 26
	BurningRequestMeta     = 27
	IssuingETHRequestMeta  = 80
	IssuingETHResponseMeta = 81

	ShardBlockReward             = 36
	AcceptedBlockRewardInfoMeta  = 37
	ShardBlockSalaryResponseMeta = 38
	BeaconRewardRequestMeta      = 39
	BeaconSalaryResponseMeta     = 40
	ReturnStakingMeta            = 41
	IncDAORewardRequestMeta      = 42
	ShardBlockRewardRequestMeta  = 43
	WithDrawRewardRequestMeta    = 44
	WithDrawRewardResponseMeta   = 45

	//statking
	ShardStakingMeta    = 63
	StopAutoStakingMeta = 127
	BeaconStakingMeta   = 64

	// Incognito -> Ethereum bridge
	BeaconSwapConfirmMeta = 70
	BridgeSwapConfirmMeta = 71
	BurningConfirmMeta    = 72

	// pde
	PDEContributionMeta         = 90
	PDETradeRequestMeta         = 91
	PDETradeResponseMeta        = 92
	PDEWithdrawalRequestMeta    = 93
	PDEWithdrawalResponseMeta   = 94
	PDEContributionResponseMeta = 95

	// portal
	PortalCustodianDepositMeta                 = 100
	PortalUserRegisterMeta                     = 101
	PortalUserRequestPTokenMeta                = 102
	PortalCustodianDepositResponseMeta         = 103
	PortalUserRequestPTokenResponseMeta        = 104
	PortalExchangeRatesMeta                    = 105
	PortalRedeemRequestMeta                    = 106
	PortalRedeemRequestResponseMeta            = 107
	PortalRequestUnlockCollateralMeta          = 108
	PortalRequestUnlockCollateralResponseMeta  = 109
	PortalCustodianWithdrawRequestMeta         = 110
	PortalCustodianWithdrawResponseMeta        = 111
	PortalLiquidateCustodianMeta               = 112
	PortalLiquidateCustodianResponseMeta       = 113
	PortalLiquidateTPExchangeRatesMeta         = 114
	PortalLiquidateTPExchangeRatesResponseMeta = 115

	PortalRewardMeta = 116
	PortalRequestWithdrawRewardMeta = 117
	PortalRequestWithdrawRewardResponseMeta = 118
	PortalRedeemLiquidateExchangeRatesMeta = 119

	// relaying
	RelayingBNBHeaderMeta = 200
	RelayingBTCHeaderMeta = 201
)

var minerCreatedMetaTypes = []int{
	ShardBlockReward,
	BeaconSalaryResponseMeta,
	IssuingResponseMeta,
	IssuingETHResponseMeta,
	ReturnStakingMeta,
	WithDrawRewardResponseMeta,
	PDETradeResponseMeta,
	PDEWithdrawalResponseMeta,
	PDEContributionResponseMeta,
	PortalUserRequestPTokenResponseMeta,
	PortalCustodianDepositResponseMeta,
	PortalRedeemRequestResponseMeta,
	PortalCustodianWithdrawResponseMeta,
	PortalLiquidateCustodianResponseMeta,
	PortalRequestWithdrawRewardResponseMeta,
}

// Special rules for shardID: stored as 2nd param of instruction of BeaconBlock
const (
	AllShards  = -1
	BeaconOnly = -2
)

var (
	// if the blockchain is running in Docker container
	// then using GETH_NAME env's value (aka geth container name)
	// otherwise using localhost
	EthereumLightNodeHost     = common.GetENV("GETH_NAME", "127.0.0.1")
	EthereumLightNodeProtocol = common.GetENV("GETH_PROTOCOL", "http")
	EthereumLightNodePort     = common.GetENV("GETH_PORT", "8545")
)

//const (
//	EthereumLightNodeProtocol = "http"
//	EthereumLightNodePort     = "8545"
//)
const (
	StopAutoStakingAmount = 0
)

var AcceptedWithdrawRewardRequestVersion = []int{0, 1}

// portal

const (
	PortalTokenSymbolBTC = "BTC"
	PortalTokenSymbolBNB = "BNB"
	PortalTokenSymbolPRV = "PRV"
)

var PortalSupportedTokenSymbols = []string{
	"BTC", // pBTC
	"BNB", // pBNB
}

var PortalSupportedIncTokenIDs = []string{
	"b832e5d3b1f01a4f0623f7fe91d6673461e1f5d37d91fe78c5c2e6183ff39696", // pBTC
	"b2655152784e8639fa19521a7035f331eea1f1e911b2f3200a507ebb4554387b", // pBNB
}

var PortalSupportedExchangeRatesSymbols = []string{
	"BTC",
	"BNB",
	"PRV",
}

var PortalSupportedTokenMap = map[string]string{
	"BTC": "b832e5d3b1f01a4f0623f7fe91d6673461e1f5d37d91fe78c5c2e6183ff39696",
	"BNB": "b2655152784e8639fa19521a7035f331eea1f1e911b2f3200a507ebb4554387b",
}
