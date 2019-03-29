package metadata

import (
	"github.com/constant-money/constant-chain/database"
)

type SendBackTokenToOldSupporterMetadata struct {
	MetadataBase
}

func (SendBackTokenToOldSupporterMetadata) ValidateTxWithBlockChain(tx Transaction, bcr BlockchainRetriever, b byte, db database.DatabaseInterface) (bool, error) {
	return true, nil
}

func (SendBackTokenToOldSupporterMetadata) ValidateSanityData(bcr BlockchainRetriever, tx Transaction) (bool, bool, error) {
	return true, true, nil
}

func (SendBackTokenToOldSupporterMetadata) ValidateMetadataByItself() bool {
	return true
}

func NewSendBackTokenToOldSupporterMetadata() *SendBackTokenToOldSupporterMetadata {
	return &SendBackTokenToOldSupporterMetadata{
		MetadataBase: *NewMetadataBase(SendBackTokenToOldSupporterMeta),
	}
}
