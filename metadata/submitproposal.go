package metadata

import (
	"github.com/ninjadotorg/constant/blockchain/params"
	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/database"
	"github.com/ninjadotorg/constant/privacy"
)

type SubmitDCBProposalMetadata struct {
	DCBParams       params.DCBParams
	ExecuteDuration uint32
	Explanation     string
	PaymentAddress  privacy.PaymentAddress

	MetadataBase
}

func NewSubmitDCBProposalMetadata(DCBParams params.DCBParams, executeDuration uint32, explanation string, address *privacy.PaymentAddress) *SubmitDCBProposalMetadata {
	return &SubmitDCBProposalMetadata{
		DCBParams:       DCBParams,
		ExecuteDuration: executeDuration,
		Explanation:     explanation,
		PaymentAddress:  *address,
		MetadataBase:    *NewMetadataBase(SubmitDCBProposalMeta),
	}
}

func (submitDCBProposalMetadata *SubmitDCBProposalMetadata) Hash() *common.Hash {
	record := string(submitDCBProposalMetadata.DCBParams.Hash().GetBytes())
	record += string(submitDCBProposalMetadata.ExecuteDuration)
	record += submitDCBProposalMetadata.Explanation
	record += string(submitDCBProposalMetadata.PaymentAddress.Bytes())
	record += string(submitDCBProposalMetadata.MetadataBase.Hash().GetBytes())
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
	ExecuteDuration uint32
	Explanation     string
	PaymentAddress  privacy.PaymentAddress

	MetadataBase
}

func NewSubmitGOVProposalMetadata(GOVParams params.GOVParams, executeDuration uint32, explaination string, address *privacy.PaymentAddress) *SubmitGOVProposalMetadata {
	return &SubmitGOVProposalMetadata{
		GOVParams:       GOVParams,
		ExecuteDuration: executeDuration,
		Explanation:     explaination,
		PaymentAddress:  *address,
		MetadataBase:    *NewMetadataBase(SubmitGOVProposalMeta),
	}
}

func (submitGOVProposalMetadata *SubmitGOVProposalMetadata) Hash() *common.Hash {
	record := string(submitGOVProposalMetadata.GOVParams.Hash().GetBytes())
	record += string(submitGOVProposalMetadata.ExecuteDuration)
	record += submitGOVProposalMetadata.Explanation
	record += string(submitGOVProposalMetadata.PaymentAddress.Bytes())
	record += string(submitGOVProposalMetadata.MetadataBase.Hash().GetBytes())
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
