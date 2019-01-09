package params

import (
	"github.com/ninjadotorg/constant/common"
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
	SaleData                 *SaleData
	MinLoanResponseRequire   uint8
	MinCMBApprovalRequire    uint8
	LateWithdrawResponseFine uint64 // CST penalty for each CMB's late withdraw response
	SaleDBCTOkensByUSDData   *SaleDBCTOkensByUSDData

	// TODO(@0xbunyip): read loan params from proposal instead of storing and reading separately
	LoanParams []LoanParams // params for collateralized loans of Constant
}

func NewDCBParams(
	saleData *SaleData,
 minLoanResponseRequire uint8,
 minCMBApprovalRequire uint8,
 lateWithdrawResponseFine uint64,
		 saleDBCTOkensByUSDData *SaleDBCTOkensByUSDData,
 loanParams []LoanParams,
 ) *DCBParams {
	return &DCBParams{SaleData: saleData, MinLoanResponseRequire: minLoanResponseRequire, MinCMBApprovalRequire: minCMBApprovalRequire, LateWithdrawResponseFine: lateWithdrawResponseFine, SaleDBCTOkensByUSDData: saleDBCTOkensByUSDData, LoanParams: loanParams}
}

func NewDCBParamsFromJson(rawData interface{}) *DCBParams {

	data := rawData.(map[string]interface{})

	saleDataData := data["saleData"].(map[string]interface{})
	saleData := NewSaleData(
		common.SliceInterfaceToSliceByte(common.InterfaceSlice(saleDataData["saleID"])),
		int32(saleDataData["endBlock"].(float64)),
		common.SliceInterfaceToSliceByte(common.InterfaceSlice(saleDataData["buyingAsset"])),
		uint64(saleDataData["buyingAmount"].(float64)),
		common.SliceInterfaceToSliceByte(common.InterfaceSlice(saleDataData["sellingAsset"])),
		uint64(saleDataData["sellingAmount"].(float64)),
	)
	
	
	minLoanResponseRequire := uint8(data["minLoanResponseRequire"].(float64))
	minCMBApprovalRequire  := uint8(data["minCMBApprovalRequire"].(float64))
	lateWithdrawResponseFine := uint64(data["lateWithdrawResponseFine"].(float64))

	saleDBCTOkensByUSDDataData := data["saleDBCTOkensByUSDDataData"].(map[string]interface{})
	saleDBCTOkensByUSDData := NewSaleDBCTOkensByUSDData(
		uint64(saleDBCTOkensByUSDDataData["amount"].(float64)),
		int32(saleDBCTOkensByUSDDataData["endBlock"].(float64)),
	)

	loanParamsData := common.InterfaceSlice(data["loanParams"])
	loanParams := make([]LoanParams, 0)
	for _, i := range loanParamsData {
		loanParamsSingleData := i.(map[string]interface{})
		loanParams = append(loanParams, *NewLoanParams(
			uint64(loanParamsSingleData["interestRate"].(float64)),
			uint32(loanParamsSingleData["maturity"].(float64)),
			uint64(loanParamsSingleData["liquidationStart"].(float64)),
		))
	}
	return NewDCBParams(saleData, minLoanResponseRequire, minCMBApprovalRequire, lateWithdrawResponseFine, saleDBCTOkensByUSDData, loanParams)
}

type GOVParams struct {
	SalaryPerTx   uint64 // salary for each tx in block(mili constant)
	BasicSalary   uint64 // basic salary per block(mili constant)
	FeePerKbTx    uint64
	SellingBonds  *SellingBonds
	RefundInfo    *RefundInfo
	OracleNetwork *OracleNetwork
}

func NewGOVParams(salaryPerTx uint64, basicSalary uint64, feePerKbTx uint64, sellingBonds *SellingBonds, refundInfo *RefundInfo, oracleNetwork *OracleNetwork) *GOVParams {
	return &GOVParams{SalaryPerTx: salaryPerTx, BasicSalary: basicSalary, FeePerKbTx: feePerKbTx, SellingBonds: sellingBonds, RefundInfo: refundInfo, OracleNetwork: oracleNetwork}
}

func NewGOVParamsFromRPC(data interface{}) *GOVParams {
	arrayParams := common.InterfaceSlice(data)

	salaryPerTx := uint64(arrayParams[0].(float64))

	basicSalary := uint64(arrayParams[1].(float64))

	feePerKbTx := uint64(arrayParams[2].(float64))

	sellingBondsData := common.InterfaceSlice(arrayParams[3])
	sellingBonds := NewSellingBonds(
		uint64(sellingBondsData[0].(float64)),
		uint64(sellingBondsData[1].(float64)),
		uint32(sellingBondsData[2].(float64)),
		uint64(sellingBondsData[3].(float64)),
		uint32(sellingBondsData[4].(float64)),
		uint32(sellingBondsData[5].(float64)),
	)

	refundInfoData := common.InterfaceSlice(arrayParams[4])
	refundInfo := NewRefundInfo(
		uint64(refundInfoData[0].(float64)),
		uint64(refundInfoData[1].(float64)),
	)

	oracleNetworkData := common.InterfaceSlice(arrayParams[5])

	oraclePubKeysInterface := common.InterfaceSlice(oracleNetworkData[0])
	oraclePubKeys := make([][]byte, 0)
	for _, i := range oraclePubKeysInterface {
		oraclePubKeys = append(oraclePubKeys, common.SliceInterfaceToSliceByte(common.InterfaceSlice(i)))
	}
	oracleNetwork := NewOracleNetwork(
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
	record += string(dcbParams.MinCMBApprovalRequire)
	record += string(dcbParams.LateWithdrawResponseFine)
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
