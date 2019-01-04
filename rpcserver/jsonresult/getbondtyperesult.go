package jsonresult

type GetBondTypeResult struct {
	BondTypes map[string]GetBondTypeResultItem // key is bond id
}

type GetBondTypeResultItem struct {
	BondName       string `json:"BondName"`
	BondSymbol     string `json:"BondSymbol"`
	StartSellingAt uint32 `json:"StartSellingAt"`
	EndSellingAt   uint32 `json:"EndSellingAt"`
	Maturity       uint32 `json:"Maturity"`
	BuyBackPrice   uint64 `json:"BuyBackPrice"`
	BuyPrice       uint64 `json:"BuyPrice"`
	TotalIssue     uint64 `json:"TotalIssue"`
	Available      uint64 `json:"Available"`
}
