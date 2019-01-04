package jsonresult

type GetBondTypeResult struct {
	BondID         []byte `json:"bondId"`
	StartSellingAt uint32 `json:"startSellingAt"`
	ExpiredDate    uint32 `json:"expiredDate"`
	BuyBackPrice   uint64 `json:"buyBackPrice"`
	BuyPrice       uint64 `json:"buyPrice"`
	TotalIssue     uint64 `json:"totalIssue"`
	Available      uint64 `json:"available"`
}
