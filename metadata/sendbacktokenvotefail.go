package metadata

import (
	"github.com/big0t/constant-chain/database"
)

type SendBackTokenVoteFailMetadata struct {
	MetadataBase
}

func (SendBackTokenVoteFailMetadata) ValidateTxWithBlockChain(tx Transaction, bcr BlockchainRetriever, b byte, db database.DatabaseInterface) (bool, error) {
	return true, nil
}

func (SendBackTokenVoteFailMetadata) ValidateSanityData(bcr BlockchainRetriever, tx Transaction) (bool, bool, error) {
	return true, true, nil
}

func (SendBackTokenVoteFailMetadata) ValidateMetadataByItself() bool {
	return true
}

func NewSendBackTokenVoteFailMetadata() *SendBackTokenVoteFailMetadata {
	return &SendBackTokenVoteFailMetadata{
		MetadataBase: *NewMetadataBase(SendBackTokenVoteFailMeta),
	}
}
