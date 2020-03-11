package metadata

import "github.com/incognitochain/incognito-chain/database/lvdb"

type TopPercentileExchangeRatesLiquidation struct {
	MetadataBase
	TPValue int
	CustodianAddress string
	ExchangeRates lvdb.FinalExchangeRates
}

type TopPercentileExchangeRatesLiquidationContent struct {
	TPValue map[string]int
	CustodianAddress string
	ExchangeRates lvdb.FinalExchangeRates
	Status string
	MetaType int
	ShardID byte
}