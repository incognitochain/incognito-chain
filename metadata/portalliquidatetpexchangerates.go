package metadata

type PortalLiquidateTopPercentileExchangeRatesContent struct {
	CustodianAddress   string
	Status             string
	MetaType           int
	TP                 map[string]LiquidateTopPercentileExchangeRatesDetail
	RemainUnlockAmount map[string]uint64
}

type LiquidateTopPercentileExchangeRatesDetail struct {
	TPKey                    int
	TPValue                  uint64
	HoldAmountFreeCollateral uint64
	HoldAmountPubToken       uint64
}

type LiquidateTopPercentileExchangeRatesStatus struct {
	CustodianAddress string
	Status           byte
	Rates            map[string]LiquidateTopPercentileExchangeRatesDetail //ptoken | detail
}

func NewLiquidateTopPercentileExchangeRatesStatus(custodianAddress string, status byte, rates map[string]LiquidateTopPercentileExchangeRatesDetail) *LiquidateTopPercentileExchangeRatesStatus {
	return &LiquidateTopPercentileExchangeRatesStatus{CustodianAddress: custodianAddress, Status: status, Rates: rates}
}



/*
	 Liquidation by exchange rates v3
 */
type LiquidationByRatesDetailV3 struct {
	Ratio                            uint64 // ratio between total locked collaterals and holding public tokens when liquidation occurs
	LiquidatedPubTokenAmount         uint64
	LiquidatedCollateralAmount       uint64            // PRV collateral
	LiquidatedTokenCollateralsAmount map[string]uint64 // externalTokenID: amount
}

type RemainUnlockCollateral struct {
	PrvAmount    uint64
	TokenAmounts map[string]uint64
}

type PortalLiquidationByRatesContentV3 struct {
	CustodianIncAddress string
	Details             map[string]LiquidationByRatesDetailV3 // portalTokenID: liquidation infos
	RemainUnlockCollaterals map[string]RemainUnlockCollateral
}


type PortalLiquidationByRatesStatusV3 struct {
	CustodianIncAddress string
	Details             map[string]LiquidationByRatesDetailV3 // portalTokenID: liquidation infos
}
