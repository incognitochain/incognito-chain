package metadata

import (
	"github.com/constant-money/constant-chain/database"
)

type SendBackTokenVoteBoardFailMetadata struct {
	MetadataBase
}

func (SendBackTokenVoteBoardFailMetadata) ValidateTxWithBlockChain(tx Transaction, bcr BlockchainRetriever, b byte, db database.DatabaseInterface) (bool, error) {
	return true, nil
}

func (SendBackTokenVoteBoardFailMetadata) ValidateSanityData(bcr BlockchainRetriever, tx Transaction) (bool, bool, error) {
	return true, true, nil
}

func (SendBackTokenVoteBoardFailMetadata) ValidateMetadataByItself() bool {
	return true
}

func NewSendBackTokenVoteFailMetadata() *SendBackTokenVoteBoardFailMetadata {
	return &SendBackTokenVoteBoardFailMetadata{
		MetadataBase: *NewMetadataBase(SendBackTokenVoteBoardFailMeta),
	}
}
