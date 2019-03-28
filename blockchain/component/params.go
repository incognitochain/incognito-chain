package component

import (
	"github.com/constant-money/constant-chain/common"
	"github.com/pkg/errors"
)

type Oracle struct {
	Bonds    map[string]uint64 // key: bondTypeID, value: price
	DCBToken uint64            // against USD
	GOVToken uint64            // against USD
	Constant uint64            // against USD
	ETH      uint64            // against USD
	BTC      uint64            // against USD
}

type LoanParams struct {
	InterestRate     uint64 `json:"InterestRate"`     // basis points, e.g. 125 represents 1.25%
	Maturity         uint64 `json:"Maturity"`         // in number of blocks
	LiquidationStart uint64 `json:"LiquidationStart"` // ratio between collateral and debt to start auto-liquidation, stored in basis points
}

func NewLoanParams(interestRate uint64, maturity uint64, liquidationStart uint64) *LoanParams {
	return &LoanParams{InterestRate: interestRate, Maturity: maturity, LiquidationStart: liquidationStart}
}

func NewLoanParamsFromJson(data interface{}) *LoanParams {
	loanParamsData := data.(map[string]interface{})
	loanParams := NewLoanParams(
		uint64(loanParamsData["InterestRate"].(float64)),
		uint64(loanParamsData["Maturity"].(float64)),
		uint64(loanParamsData["LiquidationStart"].(float64)),
	)
	return loanParams
}

func NewListLoanParamsFromJson(data interface{}) []LoanParams {
	listLoanParamsData := common.InterfaceSlice(data)
	if listLoanParamsData == nil {
		panic("listLoanParamsData")
	}
	listLoanParams := make([]LoanParams, 0)

	for _, loanParamsData := range listLoanParamsData {
		listLoanParams = append(listLoanParams, *NewLoanParamsFromJson(loanParamsData))
	}
	return listLoanParams
}

type DCBParams struct {
	ListSaleData             []SaleData
	TradeBonds               []*TradeBondWithGOV
	MinLoanResponseRequire   uint8
	MinCMBApprovalRequire    uint8
	LateWithdrawResponseFine uint64 // CST penalty for each CMB's late withdraw response
	RaiseReserveData         map[common.Hash]*RaiseReserveData
	SpendReserveData         map[common.Hash]*SpendReserveData
	DividendAmount           uint64       // maximum total Constant to pay dividend; might be less if Institution's fund ran out
	ListLoanParams           []LoanParams // component for collateralized loans of Constant
}

type RaiseReserveData struct {
	EndBlock uint64
	Amount   uint64 // # BANK tokens
}

type SpendReserveData struct {
	EndBlock        uint64
	ReserveMinPrice uint64
	Amount          uint64 // Constant to burn
}

func NewDCBParams(
	listSaleData []SaleData,
	tradeBonds []*TradeBondWithGOV,
	minLoanResponseRequire uint8,
	minCMBApprovalRequire uint8,
	lateWithdrawResponseFine uint64,
	raiseReserveData map[common.Hash]*RaiseReserveData,
	spendReserveData map[common.Hash]*SpendReserveData,
	dividendAmount uint64,
	listLoanParams []LoanParams,
) *DCBParams {
	return &DCBParams{
		ListSaleData:             listSaleData,
		TradeBonds:               tradeBonds,
		MinLoanResponseRequire:   minLoanResponseRequire,
		MinCMBApprovalRequire:    minCMBApprovalRequire,
		LateWithdrawResponseFine: lateWithdrawResponseFine,
		RaiseReserveData:         raiseReserveData,
		SpendReserveData:         spendReserveData,
		DividendAmount:           dividendAmount,
		ListLoanParams:           listLoanParams,
	}
}

func NewListSaleDataFromJson(data interface{}) []SaleData {
	listSaleDataData := common.InterfaceSlice(data)
	if listSaleDataData == nil {
		panic("list sale data")
	}
	listSaleData := make([]SaleData, 0)
	for _, i := range listSaleDataData {
		listSaleData = append(listSaleData, *NewSaleDataFromJson(i))
	}
	return listSaleData
}

func NewTradeBondsFromJson(data interface{}) ([]*TradeBondWithGOV, error) {
	tradeBondData := common.InterfaceSlice(data)
	if tradeBondData == nil {
		return nil, errors.Errorf("Invalid TradeBonds data")
	}
	tradeBonds := make([]*TradeBondWithGOV, 0)
	for _, data := range tradeBondData {
		trade, err := NewTradeBondWithGOVFromJson(data)
		if err != nil {
			return nil, err
		}
		tradeBonds = append(tradeBonds, trade)
	}
	return tradeBonds, nil
}

func NewDCBParamsFromJson(rawData interface{}) (*DCBParams, error) {
	DCBParams := rawData.(map[string]interface{})

	listSaleData := NewListSaleDataFromJson(DCBParams["ListSaleData"])
	tradeBonds, errTrade := NewTradeBondsFromJson(DCBParams["TradeBonds"])
	if errTrade != nil {
		return nil, errTrade
	}
	minLoanResponseRequire := uint8(DCBParams["MinLoanResponseRequire"].(float64))
	minCMBApprovalRequire := uint8(DCBParams["MinCMBApprovalRequire"].(float64))
	lateWithdrawResponseFine := uint64(DCBParams["LateWithdrawResponseFine"].(float64))
	dividendAmount := uint64(DCBParams["DividendAmount"].(float64))

	raiseReserveData := NewRaiseReserveDataFromJson(DCBParams["RaiseReserveData"])
	spendReserveData := NewSpendReserveDataFromJson(DCBParams["SpendReserveData"])

	listLoanParams := NewListLoanParamsFromJson(DCBParams["ListLoanParams"])
	return NewDCBParams(
		listSaleData,
		tradeBonds,
		minLoanResponseRequire,
		minCMBApprovalRequire,
		lateWithdrawResponseFine,
		raiseReserveData,
		spendReserveData,
		dividendAmount,
		listLoanParams,
	), nil
}

type GOVParams struct {
	SalaryPerTx      uint64 // salary for each tx in block(mili constant)
	BasicSalary      uint64 // basic salary per block(mili constant)
	FeePerKbTx       uint64
	SellingBonds     *SellingBonds
	SellingGOVTokens *SellingGOVTokens
	RefundInfo       *RefundInfo
	OracleNetwork    *OracleNetwork
}

func NewGOVParams(
	salaryPerTx uint64,
	basicSalary uint64,
	feePerKbTx uint64,
	sellingBonds *SellingBonds,
	sellingGOVTokens *SellingGOVTokens,
	refundInfo *RefundInfo,
	oracleNetwork *OracleNetwork,
) *GOVParams {
	return &GOVParams{
		SalaryPerTx:      salaryPerTx,
		BasicSalary:      basicSalary,
		FeePerKbTx:       feePerKbTx,
		SellingBonds:     sellingBonds,
		SellingGOVTokens: sellingGOVTokens,
		RefundInfo:       refundInfo,
		OracleNetwork:    oracleNetwork,
	}
}

func NewGOVParamsFromJson(data interface{}) *GOVParams {
	arrayParams := data.(map[string]interface{})

	salaryPerTx := uint64(arrayParams["SalaryPerTx"].(float64))
	basicSalary := uint64(arrayParams["BasicSalary"].(float64))
	feePerKbTx := uint64(arrayParams["FeePerKbTx"].(float64))
	sellingBonds := NewSellingBondsFromJson(arrayParams["SellingBonds"])
	sellingGOVTokens := NewSellingGOVTokensFromJson(arrayParams["SellingGOVTokens"])
	refundInfo := NewRefundInfoFromJson(arrayParams["RefundInfo"])
	oracleNetwork := NewOracleNetworkFromJson(arrayParams["OracleNetwork"])

	return NewGOVParams(
		salaryPerTx,
		basicSalary,
		feePerKbTx,
		sellingBonds,
		sellingGOVTokens,
		refundInfo,
		oracleNetwork,
	)
}

func (dcbParams *DCBParams) Hash() *common.Hash {
	record := ""
	for _, saleData := range dcbParams.ListSaleData {
		record += string(saleData.Hash().GetBytes())
	}
	for _, trade := range dcbParams.TradeBonds {
		record += string(trade.Hash().GetBytes())
	}
	for key, data := range dcbParams.RaiseReserveData {
		record := string(key[:])
		record += data.Hash().String()
	}
	for key, data := range dcbParams.SpendReserveData {
		record := string(key[:])
		record += data.Hash().String()
	}
	record += string(dcbParams.MinLoanResponseRequire)
	record += string(dcbParams.MinCMBApprovalRequire)
	record += string(dcbParams.LateWithdrawResponseFine)
	record += string(dcbParams.DividendAmount)
	for _, i := range dcbParams.ListLoanParams {
		record += string(i.InterestRate)
		record += string(i.Maturity)
		record += string(i.LiquidationStart)
	}
	hash := common.HashH([]byte(record))
	return &hash
}

func (govParams *GOVParams) Hash() *common.Hash {
	record := string(govParams.SalaryPerTx)
	record += string(govParams.BasicSalary)
	record += string(govParams.FeePerKbTx)
	record += string(govParams.SellingBonds.Hash().GetBytes())
	record += string(govParams.SellingGOVTokens.Hash().GetBytes())
	record += string(govParams.RefundInfo.Hash().GetBytes())
	record += string(govParams.OracleNetwork.Hash().GetBytes())
	hash := common.HashH([]byte(record))
	return &hash
}

func (dcbParams DCBParams) ValidateSanityData() bool {
	for _, saleData := range dcbParams.ListSaleData {
		if !validAssetPair(saleData.BuyingAsset, saleData.SellingAsset) {
			return false
		}
	}
	return true
}

func (GOVParams GOVParams) ValidateSanityData() bool {
	// Todo: @0xankylosaurus
	return true
}

func validAssetPair(buyingAsset common.Hash, sellingAsset common.Hash) bool {
	// DCB Bond crowdsales
	if common.IsBondAsset(&buyingAsset) && common.IsConstantAsset(&sellingAsset) {
		return true
	} else if common.IsConstantAsset(&buyingAsset) && common.IsBondAsset(&sellingAsset) {
		return true
	}
	return false
}
