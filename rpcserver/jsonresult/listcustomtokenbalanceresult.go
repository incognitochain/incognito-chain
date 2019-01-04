package jsonresult

type CustomTokenBalance struct {
	Name       string `json:"Name"`
	Symbol     string `json:"Symbol"`
	Amount     uint64 `json:"Amount"`
	TokenID    string `json:"TokenID"`
	TokenImage string `json:"TokenImage"`
}

type ListCustomTokenBalance struct {
	PaymentAddress         string               `json:"PaymentAddress"`
	ListCustomTokenBalance []CustomTokenBalance `json:"ListCustomTokenBalance"`
}

type ListLoanResponseApproved struct {
	Approvers map[string][]string `json:"Approvers"`
	Approved  map[string]bool     `json:"Approved"`
}
