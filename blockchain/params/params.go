package params

import (
	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/voting"
)

type Oracle struct {
	// TODO(@0xankylosaurus): generic prices (ETH, BTC, ...) instead of just bonds
	Bonds    map[string]uint64 // key: bondTypeID, value: price
	DCBToken uint64            // against USD
	Constant uint64            // against USD
}

type LoanParams struct {
	InterestRate     uint64 `json:"InterestRate"`     // basis points, e.g. 125 represents 1.25%
	Maturity         uint32 `json:"Maturity"`         // in number of blocks
	LiquidationStart uint64 `json:"LiquidationStart"` // ratio between collateral and debt to start auto-liquidation, stored in basis points
}

type DCBParams struct {
	SaleData               *voting.SaleData
	MinLoanResponseRequire uint8
	SaleDBCTOkensByUSDData *voting.SaleDBCTOkensByUSDData

	// TODO(@0xbunyip): read loan params from proposal instead of storing and reading separately
	LoanParams []LoanParams // params for collateralized loans of Constant
}

type GOVParams struct {
	SalaryPerTx   uint64 // salary for each tx in block(mili constant)
	BasicSalary   uint64 // basic salary per block(mili constant)
	FeePerKbTx    uint64
	SellingBonds  *voting.SellingBonds
	RefundInfo    *voting.RefundInfo
	OracleNetwork *voting.OracleNetwork
}

func (dcbParams *DCBParams) Hash() *common.Hash {
	record := string(common.ToBytes(*dcbParams.SaleData.Hash()))
	record += string(common.ToBytes(*dcbParams.SaleDBCTOkensByUSDData.Hash()))
	record += string(dcbParams.MinLoanResponseRequire)
	for _, i := range dcbParams.LoanParams {
		record += string(i.InterestRate)
		record += string(i.Maturity)
		record += string(i.LiquidationStart)
	}
	hash := common.DoubleHashH([]byte(record))
	return &hash
}

func (govParams *GOVParams) Hash() *common.Hash {
	record := string(govParams.SalaryPerTx)
	record += string(govParams.BasicSalary)
	record += string(govParams.FeePerKbTx)
	record += string(common.ToBytes(*govParams.SellingBonds.Hash()))
	record += string(common.ToBytes(*govParams.RefundInfo.Hash()))
	record += string(common.ToBytes(*govParams.OracleNetwork.Hash()))
	hash := common.DoubleHashH([]byte(record))
	return &hash
}

func (GOVParams GOVParams) Validate() bool {
	return true
}
func (DCBParams DCBParams) Validate() bool {
	return true
}

func (DCBParams DCBParams) ValidateSanityData() bool {
	// Todo: @0xbunyip
	return true
}

func (GOVParams GOVParams) ValidateSanityData() bool {
	// Todo: @0xankylosaurus
	return true
}
