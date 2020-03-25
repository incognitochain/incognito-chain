package statedb

import "github.com/incognitochain/incognito-chain/common"

type LiquidateExchangeRatesDetail struct {
	HoldAmountFreeCollateral uint64
	HoldAmountPubToken       uint64
}

type LiquidateExchangeRates struct {
	Rates map[string]LiquidateExchangeRatesDetail //ptoken | detail
}

func GenerateLiquidateExchangeRatesObjectKey(portingRequestId string) common.Hash {
	return common.Hash{}
}
