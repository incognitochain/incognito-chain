package portalv3

import (
	"errors"
	"github.com/incognitochain/incognito-chain/common"
	portalcommonv3 "github.com/incognitochain/incognito-chain/portal/portalv3/common"
	portaltokensv3 "github.com/incognitochain/incognito-chain/portal/portalv3/portaltokens"
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

	PortalTokens                 map[string]portaltokensv3.PortalTokenProcessorV3
	PortalFeederAddress          string
	PortalETHContractAddressStr  string // smart contract of ETH for portal
	MinUnlockOverRateCollaterals uint64
}

func (p PortalParams) GetSupportedCollateralTokenIDs() []string {
	tokenIDs := []string{}
	for _, col := range p.SupportedCollateralTokens {
		tokenIDs = append(tokenIDs, col.ExternalTokenID)
	}
	return tokenIDs
}

func (p PortalParams) IsSupportedTokenCollateralV3(externalTokenID string) bool {
	isSupported, _ := common.SliceExists(p.GetSupportedCollateralTokenIDs(), externalTokenID)
	return isSupported
}

func (p PortalParams) IsPortalToken(tokenIDStr string) bool {
	isExisted, _ := common.SliceExists(portalcommonv3.PortalSupportedIncTokenIDs, tokenIDStr)
	return isExisted
}

func (p PortalParams) IsPortalExchangeRateToken(tokenIDStr string) bool {
	return p.IsPortalToken(tokenIDStr) || tokenIDStr == common.PRVIDStr || p.IsSupportedTokenCollateralV3(tokenIDStr)
}

func (p PortalParams) GetMinAmountPortalToken(tokenIDStr string) (uint64, error) {
	portalToken, ok := p.PortalTokens[tokenIDStr]
	if !ok {
		return 0, errors.New("TokenID is invalid")
	}
	return portalToken.GetMinTokenAmount(), nil
}