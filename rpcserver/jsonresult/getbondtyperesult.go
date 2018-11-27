package jsonresult

type GetBondTypeResult struct {
	BondID         string `json:"bondId"`
	StartSellingAt uint32 `json:"startSellingAt"`
	Maturity       uint32 `json:"maturity"`
	BuyBackPrice   uint64 `json:"buyBackPrice"`
}
