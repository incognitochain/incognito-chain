package metadata

import (
	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/database"
	"github.com/ninjadotorg/constant/privacy"
)

type SendInitDCBVoteTokenMetadata struct {
	Amount                 uint32
	ReceiverPaymentAddress privacy.PaymentAddress

	MetadataBase
}

func NewSendInitDCBVoteTokenMetadata(amount uint32, receiverPaymentAddress privacy.PaymentAddress) *SendInitDCBVoteTokenMetadata {
	return &SendInitDCBVoteTokenMetadata{
		Amount:                 amount,
		ReceiverPaymentAddress: receiverPaymentAddress,
		MetadataBase:           *NewMetadataBase(SendInitDCBVoteTokenMeta),
	}

}

func (sendInitDCBVoteTokenMetadata *SendInitDCBVoteTokenMetadata) Hash() *common.Hash {
	record := string(sendInitDCBVoteTokenMetadata.Amount)
	record += sendInitDCBVoteTokenMetadata.ReceiverPaymentAddress.String()
	record += sendInitDCBVoteTokenMetadata.MetadataBase.Hash().String()
	hash := common.DoubleHashH([]byte(record))
	return &hash
}

func (sendInitDCBVoteTokenMetadata *SendInitDCBVoteTokenMetadata) ValidateTxWithBlockChain(Transaction, BlockchainRetriever, byte, database.DatabaseInterface) (bool, error) {
	return true, nil
}

func (sendInitDCBVoteTokenMetadata *SendInitDCBVoteTokenMetadata) ValidateSanityData(BlockchainRetriever, Transaction) (bool, bool, error) {
	return true, false, nil
}

func (sendInitDCBVoteTokenMetadata *SendInitDCBVoteTokenMetadata) ValidateMetadataByItself() bool {
	return true
}

type SendInitGOVVoteTokenMetadata struct {
	Amount                 uint32
	ReceiverPaymentAddress privacy.PaymentAddress

	MetadataBase
}

func NewSendInitGOVVoteTokenMetadata(amount uint32, receiverPaymentAddress privacy.PaymentAddress) *SendInitGOVVoteTokenMetadata {
	return &SendInitGOVVoteTokenMetadata{
		Amount:                 amount,
		ReceiverPaymentAddress: receiverPaymentAddress,
		MetadataBase:           *NewMetadataBase(SendInitGOVVoteTokenMeta),
	}

}

func (sendInitGOVVoteTokenMetadata *SendInitGOVVoteTokenMetadata) Hash() *common.Hash {
	record := string(sendInitGOVVoteTokenMetadata.Amount)
	record += sendInitGOVVoteTokenMetadata.ReceiverPaymentAddress.String()
	record += sendInitGOVVoteTokenMetadata.MetadataBase.Hash().String()
	hash := common.DoubleHashH([]byte(record))
	return &hash
}

func (sendInitGOVVoteTokenMetadata *SendInitGOVVoteTokenMetadata) ValidateTxWithBlockChain(Transaction, BlockchainRetriever, byte, database.DatabaseInterface) (bool, error) {
	return true, nil
}

func (sendInitGOVVoteTokenMetadata *SendInitGOVVoteTokenMetadata) ValidateSanityData(BlockchainRetriever, Transaction) (bool, bool, error) {
	return true, false, nil
}

func (sendInitGOVVoteTokenMetadata *SendInitGOVVoteTokenMetadata) ValidateMetadataByItself() bool {
	return true
}

func (sendInitGOVVoteTokenMetadata *SendInitGOVVoteTokenMetadata) CalculateSize() uint64 {
	return calculateSize(sendInitGOVVoteTokenMetadata)
}
