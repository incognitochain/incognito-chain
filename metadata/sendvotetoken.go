package metadata

import (
	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/database"
)

type SendInitDCBVoteTokenMetadata struct {
	Amount         uint32
	ReceiverPubKey []byte

	MetadataBase
}

func NewSendInitDCBVoteTokenMetadata(amount uint32, receiverPubKey []byte) *SendInitDCBVoteTokenMetadata {
	return &SendInitDCBVoteTokenMetadata{
		Amount:         amount,
		ReceiverPubKey: receiverPubKey,
		MetadataBase:   *NewMetadataBase(SendInitDCBVoteTokenMeta),
	}

}

func (sendInitDCBVoteTokenMetadata *SendInitDCBVoteTokenMetadata) Hash() *common.Hash {
	record := string(sendInitDCBVoteTokenMetadata.Amount)
	record += string(sendInitDCBVoteTokenMetadata.ReceiverPubKey)
	record += string(sendInitDCBVoteTokenMetadata.MetadataBase.Hash().GetBytes())
	hash := common.DoubleHashH([]byte(record))
	return &hash
}

func (sendInitDCBVoteTokenMetadata *SendInitDCBVoteTokenMetadata) ValidateTxWithBlockChain(Transaction, BlockchainRetriever, byte, database.DatabaseInterface) (bool, error) {
	return true, nil
}

func (sendInitDCBVoteTokenMetadata *SendInitDCBVoteTokenMetadata) ValidateSanityData(BlockchainRetriever, Transaction) (bool, bool, error) {
	if len(sendInitDCBVoteTokenMetadata.ReceiverPubKey) != common.PubKeyLength {
		return true, false, nil
	}
	return true, false, nil
}

func (sendInitDCBVoteTokenMetadata *SendInitDCBVoteTokenMetadata) ValidateMetadataByItself() bool {
	return true
}

type SendInitGOVVoteTokenMetadata struct {
	Amount         uint32
	ReceiverPubKey []byte

	MetadataBase
}

func NewSendInitGOVVoteTokenMetadata(amount uint32, receiverPubKey []byte) *SendInitGOVVoteTokenMetadata {
	return &SendInitGOVVoteTokenMetadata{
		Amount:         amount,
		ReceiverPubKey: receiverPubKey,
		MetadataBase:   *NewMetadataBase(SendInitGOVVoteTokenMeta),
	}

}

func (sendInitGOVVoteTokenMetadata *SendInitGOVVoteTokenMetadata) Hash() *common.Hash {
	record := string(sendInitGOVVoteTokenMetadata.Amount)
	record += string(sendInitGOVVoteTokenMetadata.ReceiverPubKey)
	record += string(sendInitGOVVoteTokenMetadata.MetadataBase.Hash().GetBytes())
	hash := common.DoubleHashH([]byte(record))
	return &hash
}

func (sendInitGOVVoteTokenMetadata *SendInitGOVVoteTokenMetadata) ValidateTxWithBlockChain(Transaction, BlockchainRetriever, byte, database.DatabaseInterface) (bool, error) {
	return true, nil
}

func (sendInitGOVVoteTokenMetadata *SendInitGOVVoteTokenMetadata) ValidateSanityData(BlockchainRetriever, Transaction) (bool, bool, error) {
	if len(sendInitGOVVoteTokenMetadata.ReceiverPubKey) != common.PubKeyLength {
		return true, false, nil
	}
	return true, false, nil
}

func (sendInitGOVVoteTokenMetadata *SendInitGOVVoteTokenMetadata) ValidateMetadataByItself() bool {
	return true
}
