package jsonresult

import (
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
)

type GetLiquidateExchangeRates struct {
	TokenId     string                        `json:"TokenId"`
	Liquidation statedb.LiquidationPoolDetail `json:"Liquidation"`
}
