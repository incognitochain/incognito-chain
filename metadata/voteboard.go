package metadata

import (
	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/database"
)

type VoteDCBBoardMetadata struct {
	CandidatePubKey []byte

	MetadataBase
}

func NewVoteDCBBoardMetadata(voteDCBBoardMetadata map[string]interface{}) *VoteDCBBoardMetadata {
	return &VoteDCBBoardMetadata{
		CandidatePubKey: voteDCBBoardMetadata["candidatePubKey"].([]byte),
	}
}

func (voteDCBBoardMetadata *VoteDCBBoardMetadata) ValidateTxWithBlockChain(txr Transaction, bcr BlockchainRetriever, chainID byte, db database.DatabaseInterface) (bool, error) {
	return true, nil
}

func (voteDCBBoardMetadata *VoteDCBBoardMetadata) GetType() int {
	return VoteDCBBoardMeta
}

func (voteDCBBoardMetadata *VoteDCBBoardMetadata) Hash() *common.Hash {
	record := string(voteDCBBoardMetadata.CandidatePubKey)
	record += string(voteDCBBoardMetadata.MetadataBase.Hash()[:])
	hash := common.DoubleHashH([]byte(record))
	return &hash
}

func (voteDCBBoardMetadata *VoteDCBBoardMetadata) ValidateSanityData(bcr BlockchainRetriever, tx Transaction) (bool, bool, error) {
	if len(voteDCBBoardMetadata.CandidatePubKey) != common.HashSize {
		return true, false, nil
	}
	return true, true, nil
}

func (voteDCBBoardMetadata *VoteDCBBoardMetadata) ValidateMetadataByItself() bool {
	return true
}

type VoteGOVBoardMetadata struct {
	CandidatePubKey []byte

	MetadataBase
}

func NewVoteGOVBoardMetadata(voteGOVBoardMetadata map[string]interface{}) *VoteGOVBoardMetadata {
	return &VoteGOVBoardMetadata{
		CandidatePubKey: voteGOVBoardMetadata["candidatePubKey"].([]byte),
	}
}

func (voteGOVBoardMetadata *VoteGOVBoardMetadata) ValidateTxWithBlockChain(txr Transaction, bcr BlockchainRetriever, chainID byte, db database.DatabaseInterface) (bool, error) {
	return true, nil
}

func (voteGOVBoardMetadata *VoteGOVBoardMetadata) GetType() int {
	return VoteGOVBoardMeta
}
func (voteGOVBoardMetadata *VoteGOVBoardMetadata) Hash() *common.Hash {
	record := string(voteGOVBoardMetadata.CandidatePubKey)
	record += string(voteGOVBoardMetadata.MetadataBase.Hash()[:])
	hash := common.DoubleHashH([]byte(record))
	return &hash
}

func (voteGOVBoardMetadata *VoteGOVBoardMetadata) ValidateSanityData(bcr BlockchainRetriever, tx Transaction) (bool, bool, error) {
	if len(voteGOVBoardMetadata.CandidatePubKey) != common.HashSize {
		return true, false, nil
	}
	return true, true, nil
}

func (voteGOVBoardMetadata *VoteGOVBoardMetadata) ValidateMetadataByItself() bool {
	return true
}
