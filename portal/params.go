package portal

import (
	"github.com/incognitochain/incognito-chain/portal/portaltokens"
	"sort"
	"time"
)

type PortalCollateral struct {
	ExternalTokenID string
	Decimal         uint8
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
	SupportedCollateralTokens            []PortalCollateral
	MinPortalFee                         uint64 // nano PRV

	PortalTokens                map[string]portaltokens.PortalTokenProcessor
	PortalFeederAddress         string
	PortalETHContractAddressStr string // smart contract of ETH for portal

	RelayingParams RelayingParams
}

type RelayingParams struct {
	BNBRelayingHeaderChainID string
	BTCRelayingHeaderChainID string
	BTCDataFolderName        string
	BNBFullNodeProtocol      string
	BNBFullNodeHost          string
	BNBFullNodePort          string
}


func GetLatestPortalParams(params map[uint64]PortalParams) PortalParams {
	if len(params) == 1 {
		return params[0]
	}

	bchs := []uint64{}
	for bch := range params {
		bchs = append(bchs, bch)
	}
	sort.Slice(bchs, func(i, j int) bool {
		return bchs[i] > bchs[j]
	})

	bchKey := bchs[0]
	return params[bchKey]
}
