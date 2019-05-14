package component

import (
	"github.com/constant-money/constant-chain/common"
)

type Oracle struct {
	Bonds    map[string]uint64 // key: bondTypeID, value: price
	DCBToken uint64            // against USD
	GOVToken uint64            // against USD
	Constant uint64            // against USD
	ETH      uint64            // against USD
	BTC      uint64            // against USD
}

type DCBParams struct {
}

func NewDCBParams() *DCBParams {
	return &DCBParams{}
}

func NewDCBParamsFromJson(rawData interface{}) (*DCBParams, error) {
	return NewDCBParams(), nil
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
	hash := common.HashH([]byte(record))
	return &hash
}

func (govParams *GOVParams) Hash() *common.Hash {
	record := string(govParams.SalaryPerTx)
	record += string(govParams.BasicSalary)
	record += string(govParams.FeePerKbTx)
	if govParams.SellingBonds != nil {
		record += string(govParams.SellingBonds.Hash().GetBytes())
	}
	if govParams.SellingGOVTokens != nil {
		record += string(govParams.SellingGOVTokens.Hash().GetBytes())
	}
	if govParams.RefundInfo != nil {
		record += string(govParams.RefundInfo.Hash().GetBytes())
	}
	if govParams.OracleNetwork != nil {
		record += string(govParams.OracleNetwork.Hash().GetBytes())
	}
	hash := common.HashH([]byte(record))
	return &hash
}

func (dcbParams DCBParams) ValidateSanityData() bool {
	return true
}

func (govParams GOVParams) ValidateSanityData() bool {
	// validation for selling bonds params
	sellingBonds := govParams.SellingBonds
	if sellingBonds != nil {
		if sellingBonds.TotalIssue == 0 || sellingBonds.BondsToSell == 0 ||
			sellingBonds.BondPrice == 0 || sellingBonds.Maturity == 0 ||
			sellingBonds.BuyBackPrice == 0 || sellingBonds.SellingWithin == 0 {
			return false
		}
		if sellingBonds.TotalIssue != sellingBonds.BondsToSell {
			return false
		}
	}

	// validation for selling gov tokens params
	sellingGOVTokens := govParams.SellingGOVTokens
	if sellingGOVTokens != nil {
		if sellingGOVTokens.TotalIssue == 0 || sellingGOVTokens.GOVTokensToSell == 0 ||
			sellingGOVTokens.GOVTokenPrice == 0 || sellingGOVTokens.SellingWithin == 0 {
			return false
		}

		if sellingGOVTokens.TotalIssue != sellingGOVTokens.GOVTokensToSell {
			return false
		}
	}

	// validation for oracle network
	oracleNetwork := govParams.OracleNetwork
	if oracleNetwork != nil {
		if oracleNetwork.WrongTimesAllowed == 0 || oracleNetwork.Quorum == 0 ||
			oracleNetwork.AcceptableErrorMargin == 0 || oracleNetwork.UpdateFrequency == 0 {
			return false
		}
	}
	return true
}
