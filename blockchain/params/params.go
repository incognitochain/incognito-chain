package params

import (
	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/voting"
)

type Oracle struct {
	// TODO(@0xankylosaurus): generic prices (ETH, BTC, ...) instead of just bonds
	Bonds    map[string]uint64 // key: bondTypeID, value: price
	DCBToken uint64            // against USD
	GOVToken uint64            // against USD
	Constant uint64            // against USD
	ETH      uint64            // against USD
	BTC      uint64            // against USD
}

type LoanParams struct {
	InterestRate     uint64 `json:"InterestRate"`     // basis points, e.g. 125 represents 1.25%
	Maturity         uint32 `json:"Maturity"`         // in number of blocks
	LiquidationStart uint64 `json:"LiquidationStart"` // ratio between collateral and debt to start auto-liquidation, stored in basis points
}

func NewLoanParams(interestRate uint64, maturity uint32, liquidationStart uint64) *LoanParams {
	return &LoanParams{InterestRate: interestRate, Maturity: maturity, LiquidationStart: liquidationStart}
}

type DCBParams struct {
	SaleData               *voting.SaleData
	MinLoanResponseRequire uint8
	SaleDBCTOkensByUSDData *voting.SaleDBCTOkensByUSDData

	// TODO(@0xbunyip): read loan params from proposal instead of storing and reading separately
	LoanParams []LoanParams // params for collateralized loans of Constant
}

func NewDCBParams(saleData *voting.SaleData, minLoanResponseRequire uint8, saleDBCTOkensByUSDData *voting.SaleDBCTOkensByUSDData, loanParams []LoanParams) *DCBParams {
	return &DCBParams{
		SaleData:               saleData,
		MinLoanResponseRequire: minLoanResponseRequire,
		SaleDBCTOkensByUSDData: saleDBCTOkensByUSDData,
		LoanParams:             loanParams,
	}
}

func NewDCBParamsFromRPC(data interface{}) *DCBParams {
	arrayParams := common.InterfaceSlice(data)

	saleDataData := common.InterfaceSlice(arrayParams[0])
	saleData := voting.NewSaleData(
		common.SliceInterfaceToSliceByte(common.InterfaceSlice(saleDataData[0])),
		int32(saleDataData[1].(float64)),
		common.SliceInterfaceToSliceByte(common.InterfaceSlice(saleDataData[2])),
		uint64(saleDataData[3].(float64)),
		common.SliceInterfaceToSliceByte(common.InterfaceSlice(saleDataData[4])),
		uint64(saleDataData[5].(float64)),
	)

	minLoanResponseRequire := uint8(arrayParams[1].(float64))

	saleDBCTOkensByUSDDataData := common.InterfaceSlice(arrayParams[2])
	saleDBCTOkensByUSDData := voting.NewSaleDBCTOkensByUSDData(
		uint64(saleDBCTOkensByUSDDataData[0].(float64)),
		int32(saleDBCTOkensByUSDDataData[1].(float64)),
	)

	loanParamsData := common.InterfaceSlice(arrayParams[3])
	loanParams := make([]LoanParams, 0)
	for _, i := range loanParamsData {
		loanParamsSingleData := common.InterfaceSlice(i)
		loanParams = append(loanParams, *NewLoanParams(
			uint64(loanParamsSingleData[0].(float64)),
			uint32(loanParamsSingleData[1].(float64)),
			uint64(loanParamsSingleData[2].(float64)),
		))
	}
	return NewDCBParams(saleData, minLoanResponseRequire, saleDBCTOkensByUSDData, loanParams)
}

type GOVParams struct {
	SalaryPerTx   uint64 // salary for each tx in block(mili constant)
	BasicSalary   uint64 // basic salary per block(mili constant)
	FeePerKbTx    uint64
	SellingBonds  *voting.SellingBonds
	RefundInfo    *voting.RefundInfo
	OracleNetwork *voting.OracleNetwork
}

func NewGOVParams(salaryPerTx uint64, basicSalary uint64, feePerKbTx uint64, sellingBonds *voting.SellingBonds, refundInfo *voting.RefundInfo, oracleNetwork *voting.OracleNetwork) *GOVParams {
	return &GOVParams{SalaryPerTx: salaryPerTx, BasicSalary: basicSalary, FeePerKbTx: feePerKbTx, SellingBonds: sellingBonds, RefundInfo: refundInfo, OracleNetwork: oracleNetwork}
}

func NewGOVParamsFromRPC(data interface{}) *GOVParams {
	arrayParams := common.InterfaceSlice(data)

	salaryPerTx := uint64(arrayParams[0].(float64))

	basicSalary := uint64(arrayParams[1].(float64))

	feePerKbTx := uint64(arrayParams[2].(float64))

	sellingBondsData := common.InterfaceSlice(arrayParams[3])
	sellingBonds := voting.NewSellingBonds(
		uint64(sellingBondsData[0].(float64)),
		uint64(sellingBondsData[1].(float64)),
		uint32(sellingBondsData[2].(float64)),
		uint64(sellingBondsData[3].(float64)),
		uint32(sellingBondsData[4].(float64)),
		uint32(sellingBondsData[5].(float64)),
	)

	refundInfoData := common.InterfaceSlice(arrayParams[4])
	refundInfo := voting.NewRefundInfo(
		uint64(refundInfoData[0].(float64)),
		uint64(refundInfoData[1].(float64)),
	)

	oracleNetworkData := common.InterfaceSlice(arrayParams[5])

	oraclePubKeysInterface := common.InterfaceSlice(oracleNetworkData[0])
	oraclePubKeys := make([][]byte, 0)
	for _, i := range oraclePubKeysInterface {
		oraclePubKeys = append(oraclePubKeys, common.SliceInterfaceToSliceByte(common.InterfaceSlice(i)))
	}
	oracleNetwork := voting.NewOracleNetwork(
		oraclePubKeys,
		uint8(oracleNetworkData[1].(float64)),
		uint8(oracleNetworkData[2].(float64)),
		uint32(oracleNetworkData[3].(float64)),
		uint32(oracleNetworkData[4].(float64)),
		uint8(oracleNetworkData[5].(float64)),
	)

	return NewGOVParams(salaryPerTx, basicSalary, feePerKbTx, sellingBonds, refundInfo, oracleNetwork)
}

func (dcbParams *DCBParams) Hash() *common.Hash {
	record := string(dcbParams.SaleData.Hash().GetBytes())
	record += string(dcbParams.SaleDBCTOkensByUSDData.Hash().GetBytes())
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
	record += string(govParams.SellingBonds.Hash().GetBytes())
	record += string(govParams.RefundInfo.Hash().GetBytes())
	record += string(govParams.OracleNetwork.Hash().GetBytes())
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
