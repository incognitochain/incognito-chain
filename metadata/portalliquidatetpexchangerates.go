package metadata

type PortalLiquidateTopPercentileExchangeRatesContent struct {
	CustodianAddress string
	Status string
	MetaType int
}

type LiquidateTopPercentileExchangeRatesDetail struct {
	TPKey int
	TPValue                  int
	HoldAmountFreeCollateral uint64
	HoldAmountPubToken       uint64
}

type LiquidateTopPercentileExchangeRatesStatus struct {
	CustodianAddress 	string
	Status				byte
	Rates        		map[string]LiquidateTopPercentileExchangeRatesDetail //ptoken | detail
}

func NewLiquidateTopPercentileExchangeRatesStatus(custodianAddress string, status byte, rates map[string]LiquidateTopPercentileExchangeRatesDetail) *LiquidateTopPercentileExchangeRatesStatus {
	return &LiquidateTopPercentileExchangeRatesStatus{CustodianAddress: custodianAddress, Status: status, Rates: rates}
}
