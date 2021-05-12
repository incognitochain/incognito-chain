package config

import "time"

type PortalParam struct {
	PortalParamV3              PortalParamV3 `yaml:"portal_param_v3"`
	RelayingParam              RelayingParam `yaml:"relaying_param"`
	BCHeightBreakPointPortalV3 uint64        `yaml:"bc_height_break_point_portal_v3"`
}

type PortalParamV3 struct {
	TimeOutCustodianReturnPubToken       time.Duration      `yaml:"time_out_custodian_return_pub_token"`
	TimeOutWaitingPortingRequest         time.Duration      `yaml:"time_out_waiting_porting_request"`
	TimeOutWaitingRedeemRequest          time.Duration      `yaml:"time_out_waiting_redeem_request"`
	MaxPercentLiquidatedCollateralAmount uint64             `yaml:"max_percent_liquidated_collateral_amount"`
	MaxPercentCustodianRewards           uint64             `yaml:"max_percent_custodian_rewards"`
	MinPercentCustodianRewards           uint64             `yaml:"min_percent_custodian_rewards"`
	MinLockCollateralAmountInEpoch       uint64             `yaml:"min_lock_collateral_amount_in_epoch"`
	MinPercentLockedCollateral           uint64             `yaml:"min_percent_locked_collateral"`
	TP120                                uint64             `yaml:"tp120"`
	TP130                                uint64             `yaml:"tp130"`
	MinPercentPortingFee                 float64            `yaml:"min_percent_porting_fee"`
	MinPercentRedeemFee                  float64            `yaml:"min_percent_redeem_fee"`
	SupportedCollateralTokens            []PortalCollateral `yaml:"supported_collateral_tokens"`
	MinPortalFee                         uint64             `yaml:"min_portal_fee" description:"nano PRV"`

	PortalTokens                 map[string]PortalTokenV3 `yaml:"portal_tokens"`
	PortalFeederAddress          string                   `yaml:"portal_feeder_address"`
	PortalETHContractAddressStr  string                   `yaml:"portal_eth_contract_address"` // smart contract of ETH for portal
	MinUnlockOverRateCollaterals uint64                   `yaml:"min_unlock_over_rate_collaterals"`
}

type RelayingParam struct {
	BNBRelayingHeaderChainID string `yaml:"bnb_relaying_header_chain_id"`
	BTCRelayingHeaderChainID string `yaml:"btc_relaying_header_chain_id"`
	BTCDataFolderName        string `yaml:"btc_data_folder_name"`
	BNBFullNodeProtocol      string `yaml:"bnb_full_node_protocol"`
	BNBFullNodeHost          string `yaml:"bnb_full_node_host"`
	BNBFullNodePort          string `yaml:"bnb_full_node_port"`
}

type PortalCollateral struct {
	ExternalTokenID string `yaml:"external_token_id"`
	Decimal         uint8  `yaml:"decimal"`
}

type PortalTokenV3 struct {
	ChainID string `yaml:"chain_id"`
	//MinTokenAmount uint64 // minimum amount for porting/redeem
}
