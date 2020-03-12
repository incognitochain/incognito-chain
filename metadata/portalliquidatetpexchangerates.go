package metadata

import "github.com/incognitochain/incognito-chain/database/lvdb"

type PortalLiquidateTopPercentileExchangeRates struct {
	MetadataBase
	TPValue int
	CustodianAddress string
	ExchangeRates lvdb.FinalExchangeRates
}

type PortalLiquidateTopPercentileExchangeRatesContent struct {
	CustodianAddress string
	Status string
	MetaType int
	ShardID byte
}