package jsonresult

type CustomTokenBalance struct {
	Name    string `json:"Name"`
	Symbol  string `json:"Symbol"`
	Amount  uint64 `json:"Amount"`
	TokenID string `json:"TokenID"`
}

type ListCustomTokenBalance struct {
	PaymentAddress         string               `json:"PaymentAddress"`
	ListCustomTokenBalance []CustomTokenBalance `json:"ListCustomTokenBalance"`
}
