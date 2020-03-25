package statedb

import "github.com/incognitochain/incognito-chain/common"

type LiquidateExchangeRatesDetail struct {
	HoldAmountFreeCollateral uint64
	HoldAmountPubToken       uint64
}

type LiquidateExchangeRatesPool struct {
	rates map[string]LiquidateExchangeRatesDetail //ptoken | detail
}

func (l *LiquidateExchangeRatesPool) Rates() map[string]LiquidateExchangeRatesDetail {
	return l.rates
}

func (l *LiquidateExchangeRatesPool) SetRates(rates map[string]LiquidateExchangeRatesDetail) {
	l.rates = rates
}

func NewLiquidateExchangeRatesPool() *LiquidateExchangeRatesPool {
	return &LiquidateExchangeRatesPool{}
}

func NewLiquidateExchangeRatesPoolWithValue(rates map[string]LiquidateExchangeRatesDetail) *LiquidateExchangeRatesPool {
	return &LiquidateExchangeRatesPool{rates: rates}
}

func GeneratePortalLiquidateExchangeRatesPoolObjectKey(beaconHeight uint64) common.Hash {
	return common.Hash{}
}
