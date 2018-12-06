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

type GOVParams struct {
	SalaryPerTx  uint64 // salary for each tx in block(mili constant)
	BasicSalary  uint64 // basic salary per block(mili constant)
	TxFee        uint64
	SellingBonds *SellingBonds
	RefundInfo   *RefundInfo
}

type RefundInfo struct {
	ThresholdToLargeTx uint64
	RefundAmount       uint64
}

type SellingBonds struct {
	BondsToSell    uint64
	BondPrice      uint64 // in Constant unit
	Maturity       uint32
	BuyBackPrice   uint64 // in Constant unit
	StartSellingAt uint32 // start selling bonds at block height
	SellingWithin  uint32 // selling bonds within n blocks
}
