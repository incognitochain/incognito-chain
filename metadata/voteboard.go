package metadata

import (
	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/database"
	"github.com/ninjadotorg/constant/privacy"
)

type VoteDCBBoardMetadata struct {
	CandidatePaymentAddress privacy.PaymentAddress

	MetadataBase
}

func NewVoteDCBBoardMetadata(candidatePaymentAddress privacy.PaymentAddress) *VoteDCBBoardMetadata {
	return &VoteDCBBoardMetadata{
		CandidatePaymentAddress: candidatePaymentAddress,
		MetadataBase:            *NewMetadataBase(VoteDCBBoardMeta),
	}
}

func (voteDCBBoardMetadata *VoteDCBBoardMetadata) ValidateTxWithBlockChain(txr Transaction, bcr BlockchainRetriever, shardID byte, db database.DatabaseInterface) (bool, error) {
	return true, nil
}

func (voteDCBBoardMetadata *VoteDCBBoardMetadata) GetType() int {
	return VoteDCBBoardMeta
}

func (voteDCBBoardMetadata *VoteDCBBoardMetadata) Hash() *common.Hash {
	record := voteDCBBoardMetadata.CandidatePaymentAddress.String()
	record += voteDCBBoardMetadata.MetadataBase.Hash().String()
	hash := common.DoubleHashH([]byte(record))
	return &hash
}

func (voteDCBBoardMetadata *VoteDCBBoardMetadata) ValidateSanityData(bcr BlockchainRetriever, tx Transaction) (bool, bool, error) {
	return true, true, nil
}

func (voteDCBBoardMetadata *VoteDCBBoardMetadata) ValidateMetadataByItself() bool {
	return true
}

type VoteGOVBoardMetadata struct {
	CandidatePaymentAddress privacy.PaymentAddress

	MetadataBase
}

func NewVoteGOVBoardMetadata(candidatePaymentAddress privacy.PaymentAddress) *VoteGOVBoardMetadata {
	return &VoteGOVBoardMetadata{
		CandidatePaymentAddress: candidatePaymentAddress,
		MetadataBase:            *NewMetadataBase(VoteGOVBoardMeta),
	}
}

func (voteGOVBoardMetadata *VoteGOVBoardMetadata) ValidateTxWithBlockChain(txr Transaction, bcr BlockchainRetriever, shardID byte, db database.DatabaseInterface) (bool, error) {
	return true, nil
}

func (voteGOVBoardMetadata *VoteGOVBoardMetadata) GetType() int {
	return VoteGOVBoardMeta
}
func (voteGOVBoardMetadata *VoteGOVBoardMetadata) Hash() *common.Hash {
	record := voteGOVBoardMetadata.CandidatePaymentAddress.String()
	record += voteGOVBoardMetadata.MetadataBase.Hash().String()
	hash := common.DoubleHashH([]byte(record))
	return &hash
}

func (voteGOVBoardMetadata *VoteGOVBoardMetadata) ValidateSanityData(bcr BlockchainRetriever, tx Transaction) (bool, bool, error) {
	return true, true, nil
}

func (voteGOVBoardMetadata *VoteGOVBoardMetadata) ValidateMetadataByItself() bool {
	return true
}

func (voteGOVBoardMetadata *VoteGOVBoardMetadata) CalculateSize() uint64 {
	return calculateSize(voteGOVBoardMetadata)
}
