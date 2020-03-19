package jsonresult

import "github.com/incognitochain/incognito-chain/database/lvdb"

type GetLiquidateTpExchangeRates struct {
	TokenId string `json:"TokenId"`
	TopPercentile string `json:"TopPercentile"`
	Data lvdb.LiquidateTopPercentileExchangeRatesDetail `json:"Data"`
}

type GetLiquidateExchangeRates struct {
	TokenId string `json:"TokenId"`
	Liquidation lvdb.LiquidateExchangeRatesDetail `json:"Liquidation"`
}