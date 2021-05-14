package config

import (
	"sort"
	"time"
)

type PortalParam struct {
	PortalParamV3              map[uint64]PortalParamV3 `mapstructure:"portal_param_v3"`
	RelayingParam              RelayingParam            `mapstructure:"relaying_param"`
	BCHeightBreakPointPortalV3 uint64                   `mapstructure:"bc_height_break_point_portal_v3"`
}

type PortalParamV3 struct {
	TimeOutCustodianReturnPubToken       time.Duration            `mapstructure:"time_out_custodian_return_pub_token"`
	TimeOutWaitingPortingRequest         time.Duration            `mapstructure:"time_out_waiting_porting_request"`
	TimeOutWaitingRedeemRequest          time.Duration            `mapstructure:"time_out_waiting_redeem_request"`
	MaxPercentLiquidatedCollateralAmount uint64                   `mapstructure:"max_percent_liquidated_collateral_amount"`
	MaxPercentCustodianRewards           uint64                   `mapstructure:"max_percent_custodian_rewards"`
	MinPercentCustodianRewards           uint64                   `mapstructure:"min_percent_custodian_rewards"`
	MinLockCollateralAmountInEpoch       uint64                   `mapstructure:"min_lock_collateral_amount_in_epoch"`
	MinPercentLockedCollateral           uint64                   `mapstructure:"min_percent_locked_collateral"`
	TP120                                uint64                   `mapstructure:"tp120"`
	TP130                                uint64                   `mapstructure:"tp130"`
	MinPercentPortingFee                 float64                  `mapstructure:"min_percent_porting_fee"`
	MinPercentRedeemFee                  float64                  `mapstructure:"min_percent_redeem_fee"`
	SupportedCollateralTokens            []PortalCollateral       `mapstructure:"supported_collateral_tokens"`
	MinPortalFee                         uint64                   `mapstructure:"min_portal_fee" description:"nano PRV"`
	PortalTokens                         map[string]PortalTokenV3 `mapstructure:"portal_tokens"`
	PortalFeederAddress                  string                   `mapstructure:"portal_feeder_address"`
	PortalETHContractAddressStr          string                   `mapstructure:"portal_eth_contract_address"` // smart contract of ETH for portal
	MinUnlockOverRateCollaterals         uint64                   `mapstructure:"min_unlock_over_rate_collaterals"`
}

type RelayingParam struct {
	BNBRelayingHeaderChainID string `mapstructure:"bnb_relaying_header_chain_id"`
	BTCRelayingHeaderChainID string `mapstructure:"btc_relaying_header_chain_id"`
	BTCDataFolderName        string `mapstructure:"btc_data_folder_name"`
	BNBFullNodeProtocol      string `mapstructure:"bnb_full_node_protocol"`
	BNBFullNodeHost          string `mapstructure:"bnb_full_node_host"`
	BNBFullNodePort          string `mapstructure:"bnb_full_node_port"`
}

type PortalCollateral struct {
	ExternalTokenID string `mapstructure:"external_token_id"`
	Decimal         uint8  `mapstructure:"decimal"`
}

type PortalTokenV3 struct {
	ChainID string `mapstructure:"chain_id"`
	//MinTokenAmount uint64 // minimum amount for porting/redeem
}

func (p PortalParam) GetPortalParamsV3(beaconHeight uint64) PortalParamV3 {
	portalParamMap := p.PortalParamV3
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
