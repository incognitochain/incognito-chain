package jsonresult

type FinalExchangeRatesDetailResult struct {
	Value uint64 `json:"Value"`
}

type FinalExchangeRatesResult struct {
	BeaconHeight uint64 `json:"BeaconHeight"`
	Rates map[string]FinalExchangeRatesDetailResult `json:"Rates"`
}

type ExchangeRatesResult struct {
	Rates map[string]uint64 `json:"Rates"`
}