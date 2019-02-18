package metadata

import (
	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/database"
	"github.com/ninjadotorg/constant/privacy"
)

type SendInitVoteTokenMetadata struct {
	Amount                 uint32
	ReceiverPaymentAddress privacy.PaymentAddress
}

func NewSendInitVoteTokenMetadata(amount uint32, receiverPaymentAddress privacy.PaymentAddress) *SendInitVoteTokenMetadata {
	return &SendInitVoteTokenMetadata{Amount: amount, ReceiverPaymentAddress: receiverPaymentAddress}
}

func (sendInitVoteTokenMetadata SendInitVoteTokenMetadata) ToBytes() []byte {
	record := string(common.Uint32ToBytes(sendInitVoteTokenMetadata.Amount))
	record += sendInitVoteTokenMetadata.ReceiverPaymentAddress.String()
	return []byte(record)
}

type SendInitDCBVoteTokenMetadata struct {
	SendInitVoteTokenMetadata SendInitVoteTokenMetadata
	MetadataBase
}

func NewSendInitDCBVoteTokenMetadata(amount uint32, receiverPaymentAddress privacy.PaymentAddress) *SendInitDCBVoteTokenMetadata {
	return &SendInitDCBVoteTokenMetadata{
		SendInitVoteTokenMetadata: *NewSendInitVoteTokenMetadata(
			amount,
			receiverPaymentAddress,
		),
		MetadataBase: *NewMetadataBase(SendInitDCBVoteTokenMeta),
	}

}

func (sendInitDCBVoteTokenMetadata *SendInitDCBVoteTokenMetadata) Hash() *common.Hash {
	record := string(sendInitDCBVoteTokenMetadata.SendInitVoteTokenMetadata.ToBytes())
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
	SendInitVoteTokenMetadata SendInitVoteTokenMetadata

	MetadataBase
}

func NewSendInitGOVVoteTokenMetadata(amount uint32, receiverPaymentAddress privacy.PaymentAddress) *SendInitGOVVoteTokenMetadata {
	return &SendInitGOVVoteTokenMetadata{
		SendInitVoteTokenMetadata: *NewSendInitVoteTokenMetadata(
			amount,
			receiverPaymentAddress,
		),
		MetadataBase: *NewMetadataBase(SendInitGOVVoteTokenMeta),
	}

}

func (sendInitGOVVoteTokenMetadata *SendInitGOVVoteTokenMetadata) Hash() *common.Hash {
	record := string(sendInitGOVVoteTokenMetadata.SendInitVoteTokenMetadata.ToBytes())
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
