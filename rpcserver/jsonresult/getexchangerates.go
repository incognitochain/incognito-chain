package jsonresult

type FinalExchangeRatesDetailResult struct {
	Value uint64 `json:"Value"`
}

type FinalExchangeRatesResult struct {
	Rates map[string]FinalExchangeRatesDetailResult `json:"Rates"`
}