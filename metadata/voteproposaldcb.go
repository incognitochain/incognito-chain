package metadata

import (
	"github.com/constant-money/constant-chain/blockchain/component"
	"github.com/constant-money/constant-chain/common"
	"github.com/constant-money/constant-chain/database"
	"github.com/constant-money/constant-chain/metadata/fromshardins"
)

func (dcbVoteProposalMetadata *DCBVoteProposalMetadata) GetBoardType() common.BoardType {
	return common.DCBBoard
}

type DCBVoteProposalMetadata struct {
	NormalVoteProposalMetadata component.VoteProposalData
	MetadataBase
}

func (dcbVoteProposalMetadata *DCBVoteProposalMetadata) ValidateSanityData(bcr BlockchainRetriever, tx Transaction) (bool, bool, error) {
	//return dcbVoteProposalMetadata.VoteProposalMetadata.ValidateSanityData(bcr, tx)
	return true, true, nil
}

func (dcbVoteProposalMetadata *DCBVoteProposalMetadata) ValidateMetadataByItself() bool {
	//return dcbVoteProposalMetadata.VoteProposalMetadata.ValidateMetadataByItself()
	return true
}

func NewDCBVoteProposalMetadata(
	voteProposal component.VoteProposalData,
) *DCBVoteProposalMetadata {
	return &DCBVoteProposalMetadata{
		NormalVoteProposalMetadata: voteProposal,
		MetadataBase:               *NewMetadataBase(DCBVoteProposalMeta),
	}
}

func (dcbVoteProposalMetadata *DCBVoteProposalMetadata) Hash() *common.Hash {
	record := dcbVoteProposalMetadata.NormalVoteProposalMetadata.ToBytes()

	hash := common.DoubleHashH([]byte(record))
	return &hash
}

func (dcbVoteProposalMetadata *DCBVoteProposalMetadata) ValidateTxWithBlockChain(tx Transaction, bcr BlockchainRetriever, shardID byte, db database.DatabaseInterface) (bool, error) {
	//Validate these pubKeys are in board
	//boardType := common.DCBBoard
	//return dcbVoteProposalMetadata.VoteProposalMetadata.ValidateTxWithBlockChain(
	//	boardType,
	//	tx,
	//	bcr,
	//	shardID,
	//	db,
	//)
	return true, nil
}

func (dcbVoteProposalMetadata *DCBVoteProposalMetadata) BuildReqActions(
	//Hyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyy
	//Step 1 hyyyyyyyyyyyyyyyyyyyyyyyy
	tx Transaction,
	bcr BlockchainRetriever,
	shardID byte,
) ([][]string, error) {
	voteProposal := dcbVoteProposalMetadata.NormalVoteProposalMetadata
	inst := fromshardins.NewNormalVoteProposalIns(common.DCBBoard, voteProposal)

	instStr, err := inst.GetStringFormat()
	if err != nil {
		return nil, err
	}
	return [][]string{instStr}, nil
}
