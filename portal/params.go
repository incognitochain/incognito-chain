package portal

import (
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

	PortalTokens                map[string]PortalTokenProcessor
	BNBRelayingHeaderChainID    string
	BTCRelayingHeaderChainID    string
	BTCDataFolderName           string
	BNBFullNodeProtocol         string
	BNBFullNodeHost             string
	BNBFullNodePort             string
	PortalFeederAddress         string
	PortalETHContractAddressStr string // smart contract of ETH for portal
}

