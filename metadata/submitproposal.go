package metadata

import (
	"github.com/ninjadotorg/constant/blockchain/params"
	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/database"
	"github.com/ninjadotorg/constant/privacy"
)

type SubmitDCBProposalMetadata struct {
	DCBParams       params.DCBParams
	ExecuteDuration uint64
	Explanation     string
	PaymentAddress  privacy.PaymentAddress

	MetadataBase
}

func NewSubmitDCBProposalMetadata(DCBParams params.DCBParams, executeDuration uint64, explanation string, address *privacy.PaymentAddress) *SubmitDCBProposalMetadata {
	return &SubmitDCBProposalMetadata{
		DCBParams:       DCBParams,
		ExecuteDuration: executeDuration,
		Explanation:     explanation,
		PaymentAddress:  *address,
		MetadataBase:    *NewMetadataBase(SubmitDCBProposalMeta),
	}
}

func NewSubmitDCBProposalMetadataFromJson(data interface{}) *SubmitDCBProposalMetadata {
	SubmitDCBProposalData := data.(map[string]interface{})
	meta := NewSubmitDCBProposalMetadata(
		*params.NewDCBParamsFromJson(SubmitDCBProposalData["DCBParams"]),
		uint64(SubmitDCBProposalData["ExecuteDuration"].(float64)),
		SubmitDCBProposalData["Explanation"].(string),
		SubmitDCBProposalData["PaymentAddress"].(*privacy.PaymentAddress),
	)
	return meta
}

func (submitDCBProposalMetadata *SubmitDCBProposalMetadata) Hash() *common.Hash {
	record := submitDCBProposalMetadata.DCBParams.Hash().String()
	record += string(submitDCBProposalMetadata.ExecuteDuration)
	record += submitDCBProposalMetadata.Explanation
	record += submitDCBProposalMetadata.PaymentAddress.String()
	record += submitDCBProposalMetadata.MetadataBase.Hash().String()
	hash := common.DoubleHashH([]byte(record))
	return &hash
}

func (submitDCBProposalMetadata *SubmitDCBProposalMetadata) ValidateTxWithBlockChain(Transaction, BlockchainRetriever, byte, database.DatabaseInterface) (bool, error) {
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
	ExecuteDuration uint64
	Explanation     string
	PaymentAddress  privacy.PaymentAddress

	MetadataBase
}

func NewSubmitGOVProposalMetadata(
	govParams params.GOVParams,
	executeDuration uint64,
	explanation string,
	address *privacy.PaymentAddress,
) *SubmitGOVProposalMetadata {
	return &SubmitGOVProposalMetadata{
		GOVParams:       govParams,
		ExecuteDuration: executeDuration,
		Explanation:     explanation,
		PaymentAddress:  *address,
		MetadataBase:    *NewMetadataBase(SubmitGOVProposalMeta),
	}
}

func NewSubmitGOVProposalMetadataFromJson(data interface{}) *SubmitGOVProposalMetadata {
	submitGOVProposalData := data.(map[string]interface{})

	return NewSubmitGOVProposalMetadata(
		*params.NewGOVParamsFromJson(submitGOVProposalData["GOVParams"]),
		uint64(submitGOVProposalData["ExecuteDuration"].(float64)),
		submitGOVProposalData["Explanation"].(string),
		submitGOVProposalData["PaymentAddress"].(*privacy.PaymentAddress),
	)
}

func (submitGOVProposalMetadata *SubmitGOVProposalMetadata) Hash() *common.Hash {
	record := submitGOVProposalMetadata.GOVParams.Hash().String()
	record += string(submitGOVProposalMetadata.ExecuteDuration)
	record += submitGOVProposalMetadata.Explanation
	record += submitGOVProposalMetadata.PaymentAddress.String()
	record += submitGOVProposalMetadata.MetadataBase.Hash().String()
	hash := common.DoubleHashH([]byte(record))
	return &hash
}

func (submitGOVProposalMetadata *SubmitGOVProposalMetadata) ValidateTxWithBlockChain(Transaction, BlockchainRetriever, byte, database.DatabaseInterface) (bool, error) {
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
	if len(submitGOVProposalMetadata.Explanation) > common.MaximumProposalExplainationLength {
		return true, false, nil
	}
	return true, true, nil
}

func (submitGOVProposalMetadata *SubmitGOVProposalMetadata) ValidateMetadataByItself() bool {
	return true
}
