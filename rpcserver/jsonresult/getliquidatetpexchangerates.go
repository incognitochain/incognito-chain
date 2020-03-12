package jsonresult

import "github.com/incognitochain/incognito-chain/database/lvdb"

type GetLiquidateTpExchangeRates struct {
	TokenSymbol string `json:"TokenSymbol"`
	TopPercentile string `json:"TopPercentile"`
	Data lvdb.LiquidateTopPercentileExchangeRatesDetail `json:"data"`
}

type GetLiquidateExchangeRates struct {
	TokenSymbol string `json:"TokenSymbol"`
	Liquidation lvdb.LiquidateExchangeRatesDetail `json:"Liquidation"`
}