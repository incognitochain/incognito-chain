package metadata

import (
	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/database"
)

type SendInitDCBVoteTokenMetadata struct {
	Amount         uint64
	ReceiverPubKey []byte

	MetadataBase
}

func NewSendInitDCBVoteTokenMetadata(data map[string]interface{}) *SendInitDCBVoteTokenMetadata {
	sendInitDCBVoteTokenMetadata := SendInitDCBVoteTokenMetadata{
		Amount:         data["Amount"].(uint64),
		ReceiverPubKey: data["ReceiverPubkey"].([]byte),
		MetadataBase: MetadataBase{
			Type: SendInitDCBVoteTokenMeta,
		},
	}
	return &sendInitDCBVoteTokenMetadata
}

func (sendInitDCBVoteTokenMetadata *SendInitDCBVoteTokenMetadata) Hash() *common.Hash {
	record := string(sendInitDCBVoteTokenMetadata.Amount)
	record += string(sendInitDCBVoteTokenMetadata.ReceiverPubKey)
	record += string(sendInitDCBVoteTokenMetadata.MetadataBase.Hash()[:])
	hash := common.DoubleHashH([]byte(record))
	return &hash
}

func (sendInitDCBVoteTokenMetadata *SendInitDCBVoteTokenMetadata) ValidateTxWithBlockChain(Transaction, BlockchainRetriever, byte, database.DatabaseInterface) (bool, error) {
	return true, nil
}

func (sendInitDCBVoteTokenMetadata *SendInitDCBVoteTokenMetadata) ValidateSanityData(BlockchainRetriever, Transaction) (bool, bool, error) {
	if len(sendInitDCBVoteTokenMetadata.ReceiverPubKey) != common.HashSize {
		return true, false, nil
	}
	return true, false, nil
}

func (sendInitDCBVoteTokenMetadata *SendInitDCBVoteTokenMetadata) ValidateMetadataByItself() bool {
	return true
}

type SendInitGOVVoteTokenMetadata struct {
	Amount         uint64
	ReceiverPubKey []byte

	MetadataBase
}

func NewSendInitGOVVoteTokenMetadata(data map[string]interface{}) *SendInitGOVVoteTokenMetadata {
	sendInitGOVVoteTokenMetadata := SendInitGOVVoteTokenMetadata{
		Amount:         data["Amount"].(uint64),
		ReceiverPubKey: data["ReceiverPubkey"].([]byte),
		MetadataBase: MetadataBase{
			Type: SendInitGOVVoteTokenMeta,
		},
	}
	return &sendInitGOVVoteTokenMetadata
}

func (sendInitGOVVoteTokenMetadata *SendInitGOVVoteTokenMetadata) Hash() *common.Hash {
	record := string(sendInitGOVVoteTokenMetadata.Amount)
	record += string(sendInitGOVVoteTokenMetadata.ReceiverPubKey)
	record += string(sendInitGOVVoteTokenMetadata.MetadataBase.Hash()[:])
	hash := common.DoubleHashH([]byte(record))
	return &hash
}

func (sendInitGOVVoteTokenMetadata *SendInitGOVVoteTokenMetadata) ValidateTxWithBlockChain(Transaction, BlockchainRetriever, byte, database.DatabaseInterface) (bool, error) {
	return true, nil
}

func (sendInitGOVVoteTokenMetadata *SendInitGOVVoteTokenMetadata) ValidateSanityData(BlockchainRetriever, Transaction) (bool, bool, error) {
	if len(sendInitGOVVoteTokenMetadata.ReceiverPubKey) != common.HashSize {
		return true, false, nil
	}
	return true, false, nil
}

func (sendInitGOVVoteTokenMetadata *SendInitGOVVoteTokenMetadata) ValidateMetadataByItself() bool {
	return true
}
