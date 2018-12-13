package metadata

import (
	"github.com/ninjadotorg/constant/blockchain/params"
	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/voting"
)

type SubmitDCBProposalMetadata struct {
	DCBParams       params.DCBParams
	ExecuteDuration int32
	Explanation     string

	MetadataBase
}

//calling from rpc function
func NewSubmitDCBProposalMetadataFromJson(jsonData map[string]interface{}) *SubmitDCBProposalMetadata {
	LoanParamList := (jsonData["LoanParams"].([]interface{}))
	loanParams := make([]params.LoanParams, 0)
	for _, param := range LoanParamList {
		j := param.(map[string]interface{})
		loanParams = append(loanParams, params.LoanParams{
			InterestRate:     uint64(j["InterestRate"].(float64)),
			Maturity:         uint32(j["Maturity"].(float64)),
			LiquidationStart: uint64(j["LiquidationStart"].(float64)),
		})
	}
	submitDCBProposalMetadata := SubmitDCBProposalMetadata{
		DCBParams: params.DCBParams{
			SaleData: &voting.SaleData{
				SaleID:       []byte(jsonData["SaleID"].(string)),
				BuyingAsset:  []byte(jsonData["BuyingAsset"].(string)),
				SellingAsset: []byte(jsonData["SellingAsset"].(string)),
				EndBlock:     int32(jsonData["EndBlock"].(float64)),
			},
			MinLoanResponseRequire: uint8(jsonData["MinLoanResponseRequire"].(float64)),
			LoanParams:             loanParams,
		},
		ExecuteDuration: int32(jsonData["ExecuteDuration"].(float64)),
		Explanation:     jsonData["Explanation"].(string),
		MetadataBase: MetadataBase{
			Type: SubmitDCBProposalMeta,
		},
	}
	return &submitDCBProposalMetadata
}

func (submitDCBProposalMetadata *SubmitDCBProposalMetadata) Hash() *common.Hash {
	record := string(common.ToBytes(submitDCBProposalMetadata.DCBParams.Hash()))
	record += string(submitDCBProposalMetadata.ExecuteDuration)
	record += submitDCBProposalMetadata.Explanation
	hash := common.DoubleHashH([]byte(record))
	return &hash
}

func (submitDCBProposalMetadata *SubmitDCBProposalMetadata) ValidateTxWithBlockChain(Transaction, BlockchainRetriever, byte) (bool, error) {
	return true, nil
}

func (submitDCBProposalMetadata *SubmitDCBProposalMetadata) ValidateSanityData(BlockchainRetriever, Transaction) (bool, bool, error) {
	if !submitDCBProposalMetadata.DCBParams.ValidateSanityData() {
		return true, false, nil
	}
	if submitDCBProposalMetadata.ExecuteDuration < common.MinimumBlockOfProposalDuration ||
		submitDCBProposalMetadata.ExecuteDuration > common.MaximumBlockOfProposalDuration {
		return true, false, nil
	}
	if len(submitDCBProposalMetadata.Explanation) > common.MaximumProposalExplainationLength {
		return true, false, nil
	}
	return true, true, nil
}

func (submitDCBProposalMetadata *SubmitDCBProposalMetadata) ValidateMetadataByItself() bool {
	return true
}

type SubmitGOVProposalMetadata struct {
	GOVParams       params.GOVParams
	ExecuteDuration int32
	Explaination    string

	MetadataBase
}

//calling from rpc function
func NewSubmitGOVProposalMetadataFromJson(jsonData map[string]interface{}) *SubmitGOVProposalMetadata {
	submitGOVProposalMetadata := SubmitGOVProposalMetadata{
		GOVParams: params.GOVParams{
			SalaryPerTx: uint64(jsonData["SalaryPerTx"].(float64)),
			BasicSalary: uint64(jsonData["BasicSalary"].(float64)),
			TxFee:       uint64(jsonData["TxFee"].(float64)),
			SellingBonds: &voting.SellingBonds{
				BondsToSell:    uint64(jsonData["BondsToSell"].(float64)),
				BondPrice:      uint64(jsonData["BondPrice"].(float64)),
				Maturity:       uint32(jsonData["Maturity"].(float64)),
				BuyBackPrice:   uint64(jsonData["BuyBackPrice"].(float64)),
				StartSellingAt: uint32(jsonData["StartSellingAt"].(float64)),
				SellingWithin:  uint32(jsonData["SellingWithin"].(float64)),
			},
			RefundInfo: &voting.RefundInfo{
				ThresholdToLargeTx: uint64(jsonData["ThresholdToLargeTx"].(float64)),
				RefundAmount:       uint64(jsonData["RefundAmount"].(float64)),
			},
		},
		ExecuteDuration: int32(jsonData["ExecuteDuration"].(float64)),
		Explaination:    string(jsonData["Explaination"].(string)),
		MetadataBase: MetadataBase{
			Type: SubmitGOVProposalMeta,
		},
	}
	return &submitGOVProposalMetadata
}

func (submitGOVProposalMetadata *SubmitGOVProposalMetadata) Hash() *common.Hash {
	record := string(common.ToBytes(submitGOVProposalMetadata.GOVParams.Hash()))
	record += string(submitGOVProposalMetadata.ExecuteDuration)
	record += submitGOVProposalMetadata.Explaination
	hash := common.DoubleHashH([]byte(record))
	return &hash
}

func (submitGOVProposalMetadata *SubmitGOVProposalMetadata) ValidateTxWithBlockChain(Transaction, BlockchainRetriever, byte) (bool, error) {
	return true, nil
}

func (submitGOVProposalMetadata *SubmitGOVProposalMetadata) ValidateSanityData(BlockchainRetriever, Transaction) (bool, bool, error) {
	if !submitGOVProposalMetadata.GOVParams.ValidateSanityData() {
		return true, false, nil
	}
	if submitGOVProposalMetadata.ExecuteDuration < common.MinimumBlockOfProposalDuration ||
		submitGOVProposalMetadata.ExecuteDuration > common.MaximumBlockOfProposalDuration {
		return true, false, nil
	}
	if len(submitGOVProposalMetadata.Explaination) > common.MaximumProposalExplainationLength {
		return true, false, nil
	}
	return true, true, nil
}

func (submitGOVProposalMetadata *SubmitGOVProposalMetadata) ValidateMetadataByItself() bool {
	return true
}
