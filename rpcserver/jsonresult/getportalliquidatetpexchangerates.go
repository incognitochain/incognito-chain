package jsonresult

import (
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/metadata"
)

type GetLiquidateTpExchangeRates struct {
	TokenId       string                                             `json:"TokenId"`
	TopPercentile string                                             `json:"TopPercentile"`
	Data          metadata.LiquidateTopPercentileExchangeRatesDetail `json:"Data"`
}

type GetLiquidateExchangeRates struct {
	TokenId     string                        `json:"TokenId"`
	Liquidation statedb.LiquidationPoolDetail `json:"Liquidation"`
}

type GetLiquidateAmountNeededCustodianDeposit struct {
	TokenId string `json:"TokenId"`
	Amount  uint64 `json:"Amount"`
}
