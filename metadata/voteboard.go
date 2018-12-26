package metadata

import (
	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/database"
)

type VoteDCBBoardMetadata struct {
	CandidatePubKey []byte

	MetadataBase
}

func NewVoteDCBBoardMetadata(candidatePubKey []byte) *VoteDCBBoardMetadata {
	return &VoteDCBBoardMetadata{
		CandidatePubKey: candidatePubKey,
		MetadataBase:    *NewMetadataBase(VoteDCBBoardMeta),
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
	record += string(voteDCBBoardMetadata.MetadataBase.Hash().GetBytes())
	hash := common.DoubleHashH([]byte(record))
	return &hash
}

func (voteDCBBoardMetadata *VoteDCBBoardMetadata) ValidateSanityData(bcr BlockchainRetriever, tx Transaction) (bool, bool, error) {
	if len(voteDCBBoardMetadata.CandidatePubKey) != common.PubKeyLength {
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

func NewVoteGOVBoardMetadata(candidatePubKey []byte) *VoteGOVBoardMetadata {
	return &VoteGOVBoardMetadata{
		CandidatePubKey: candidatePubKey,
		MetadataBase:    *NewMetadataBase(VoteGOVBoardMeta),
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
	record += string(voteGOVBoardMetadata.MetadataBase.Hash().GetBytes())
	hash := common.DoubleHashH([]byte(record))
	return &hash
}

func (voteGOVBoardMetadata *VoteGOVBoardMetadata) ValidateSanityData(bcr BlockchainRetriever, tx Transaction) (bool, bool, error) {
	if len(voteGOVBoardMetadata.CandidatePubKey) != common.PubKeyLength {
		return true, false, nil
	}
	return true, true, nil
}

func (voteGOVBoardMetadata *VoteGOVBoardMetadata) ValidateMetadataByItself() bool {
	return true
}
