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

// this function should be in privacy package
func NewPaymentAddress(pk privacy.PublicKey, tk privacy.TransmissionKey) *privacy.PaymentAddress {
	payment := privacy.PaymentAddress{}
	payment.Pk = pk
	payment.Tk = tk
	return &payment
}

func NewPaymentAddressFromJson(data interface{}) *privacy.PaymentAddress {
	paymentAddressData := data.(map[string]interface{})
	x := paymentAddressData["Pk"].(string)
	_ = x
	paymentAddress := NewPaymentAddress(
		[]byte(paymentAddressData["Pk"].(string)),
		[]byte(paymentAddressData["Tk"].(string)),
	)
	return paymentAddress
}

func NewSubmitDCBProposalMetadataFromJson(data interface{}) *SubmitDCBProposalMetadata {
	SubmitDCBProposalData := data.(map[string]interface{})
	meta := NewSubmitDCBProposalMetadata(
		*params.NewDCBParamsFromJson(SubmitDCBProposalData["DCBParams"]),
		uint32(SubmitDCBProposalData["ExecuteDuration"].(float64)),
		SubmitDCBProposalData["Explanation"].(string),
		NewPaymentAddressFromJson(SubmitDCBProposalData["PaymentAddress"]),
	)
	return meta
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

func NewSubmitGOVProposalMetadata(
	govParams params.GOVParams,
	executeDuration uint32,
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

	paymentAddressData := submitGOVProposalData["PaymentAddress"].(map[string]interface{})
	paymentAddress := privacy.PaymentAddress{
		Pk: []byte(paymentAddressData["Pk"].(string)),
		Tk: []byte(paymentAddressData["Tk"].(string)),
	}

	return NewSubmitGOVProposalMetadata(
		*params.NewGOVParamsFromJson(submitGOVProposalData["GOVParams"]),
		uint32(submitGOVProposalData["ExecuteDuration"].(float64)),
		submitGOVProposalData["Explanation"].(string),
		&paymentAddress,
	)
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
