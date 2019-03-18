package metadata

import (
	"github.com/constant-money/constant-chain/blockchain/component"
	"github.com/constant-money/constant-chain/common"
	"github.com/constant-money/constant-chain/database"
	"github.com/constant-money/constant-chain/metadata/fromshardins"
)

func (govVoteProposalMetadata *GOVVoteProposalMetadata) GetBoardType() common.BoardType {
	return common.GOVBoard
}

type GOVVoteProposalMetadata struct {
	VoteProposalMetadata component.VoteProposalData
	MetadataBase
}

func (govVoteProposalMetadata *GOVVoteProposalMetadata) ValidateSanityData(bcr BlockchainRetriever, tx Transaction) (bool, bool, error) {
	//return govVoteProposalMetadata.VoteProposalMetadata.ValidateSanityData(bcr, tx)
	return true, true, nil
}

func (govVoteProposalMetadata *GOVVoteProposalMetadata) ValidateMetadataByItself() bool {
	//return govVoteProposalMetadata.VoteProposalMetadata.ValidateMetadataByItself()
	return true
}

func NewGOVVoteProposalMetadata(
	voteProposal component.VoteProposalData,
) *GOVVoteProposalMetadata {
	return &GOVVoteProposalMetadata{
		VoteProposalMetadata: voteProposal,
		MetadataBase:         *NewMetadataBase(GOVVoteProposalMeta),
	}
}

func (govVoteProposalMetadata *GOVVoteProposalMetadata) Hash() *common.Hash {
	record := govVoteProposalMetadata.VoteProposalMetadata.ToBytes()

	hash := common.DoubleHashH([]byte(record))
	return &hash
}

func (govVoteProposalMetadata *GOVVoteProposalMetadata) ValidateTxWithBlockChain(tx Transaction, bcr BlockchainRetriever, shardID byte, db database.DatabaseInterface) (bool, error) {
	//Validate these pubKeys are in board
	//boardType := common.GOVBoard
	//return govVoteProposalMetadata.VoteProposalMetadata.ValidateTxWithBlockChain(
	//	boardType,
	//	tx,
	//	bcr,
	//	shardID,
	//	db,
	//)
	return true, nil
}

func (govVoteProposalMetadata *GOVVoteProposalMetadata) BuildReqActions(
	tx Transaction,
	bcr BlockchainRetriever,
	shardID byte,
) ([][]string, error) {
	voteProposal := govVoteProposalMetadata.VoteProposalMetadata
	inst := fromshardins.NewNormalVoteProposalIns(common.GOVBoard, voteProposal)

	instStr, err := inst.GetStringFormat()
	if err != nil {
		return nil, err
	}
	return [][]string{instStr}, nil
}
