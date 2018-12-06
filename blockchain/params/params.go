package params

import (
	"github.com/ninjadotorg/constant/voting"
)

type LoanParams struct {
	InterestRate     uint64 `json:"InterestRate"`     // basis points, e.g. 125 represents 1.25%
	Maturity         uint32 `json:"Maturity"`         // in number of blocks
	LiquidationStart uint64 `json:"LiquidationStart"` // ratio between collateral and debt to start auto-liquidation, stored in basis points
}

type DCBParams struct {
	SaleData               *voting.SaleData
	MinLoanResponseRequire uint8
	LoanParams             []LoanParams // params for collateralized loans of Constant
}
