package jsonresult

type GetCurrentSellingGOVTokens struct {
	GOVTokenID     string `json:"GOVTokenID"`
	StartSellingAt uint64 `json:"StartSellingAt"`
	EndSellingAt   uint64 `json:"EndSellingAt"`
	BuyPrice       uint64 `json:"BuyPrice"`
	TotalIssue     uint64 `json:"TotalIssue"`
	Available      uint64 `json:"Available"`
}
