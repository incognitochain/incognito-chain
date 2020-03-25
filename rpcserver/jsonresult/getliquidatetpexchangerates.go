package jsonresult


type GetLiquidateTpExchangeRates struct {
	TokenId string `json:"TokenId"`
	TopPercentile string `json:"TopPercentile"`
	Data lvdb.LiquidateTopPercentileExchangeRatesDetail `json:"Data"`
}

type GetLiquidateExchangeRates struct {
	TokenId string `json:"TokenId"`
	Liquidation lvdb.LiquidateExchangeRatesDetail `json:"Liquidation"`
}

type GetLiquidateAmountNeededCustodianDeposit struct {
	TokenId string `json:"TokenId"`
	IsFreeCollateralSelected bool `json:"IsFreeCollateralSelected"`
	Amount uint64 `json:"Amount"`
    FreeCollateral uint64 `json:"FreeCollateral"`
}