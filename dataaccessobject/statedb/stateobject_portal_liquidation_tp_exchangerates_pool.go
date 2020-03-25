package statedb

import "github.com/incognitochain/incognito-chain/common"

type LiquidateExchangeRatesDetail struct {
	HoldAmountFreeCollateral uint64
	HoldAmountPubToken       uint64
}

type LiquidateExchangeRates struct {
	rates map[string]LiquidateExchangeRatesDetail //ptoken | detail
}

func (l *LiquidateExchangeRates) Rates() map[string]LiquidateExchangeRatesDetail {
	return l.rates
}

func (l *LiquidateExchangeRates) SetRates(rates map[string]LiquidateExchangeRatesDetail) {
	l.rates = rates
}

func NewLiquidateExchangeRates() *LiquidateExchangeRates {
	return &LiquidateExchangeRates{}
}

func NewLiquidateExchangeRatesWithValue(rates map[string]LiquidateExchangeRatesDetail) *LiquidateExchangeRates {
	return &LiquidateExchangeRates{rates: rates}
}

func GeneratePortalLiquidateExchangeRatesObjectKey(beaconHeight uint64) common.Hash {
	return common.Hash{}
}
