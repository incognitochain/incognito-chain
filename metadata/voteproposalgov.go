package metadata

import (
	"github.com/constant-money/constant-chain/blockchain/component"
	"github.com/constant-money/constant-chain/common"
	"github.com/constant-money/constant-chain/database"
	"github.com/constant-money/constant-chain/metadata/fromshardins"
	"github.com/constant-money/constant-chain/privacy"
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

type PunishGOVDecryptMetadata struct {
	PunishDecryptMetadata PunishDecryptMetadata
	MetadataBase
}

func NewPunishGOVDecryptMetadata(paymentAddress privacy.PaymentAddress) *PunishGOVDecryptMetadata {
	return &PunishGOVDecryptMetadata{
		PunishDecryptMetadata: PunishDecryptMetadata{
			PaymentAddress: paymentAddress,
		},
		MetadataBase: *NewMetadataBase(PunishGOVDecryptMeta),
	}
}

func (punishGOVDecryptMetadata *PunishGOVDecryptMetadata) Hash() *common.Hash {
	record := string(punishGOVDecryptMetadata.PunishDecryptMetadata.ToBytes())
	record += punishGOVDecryptMetadata.MetadataBase.Hash().String()

	hash := common.DoubleHashH([]byte(record))
	return &hash
}

func (punishGOVDecryptMetadata *PunishGOVDecryptMetadata) ValidateTxWithBlockChain(Transaction, BlockchainRetriever, byte, database.DatabaseInterface) (bool, error) {
	return true, nil
}

func (punishGOVDecryptMetadata *PunishGOVDecryptMetadata) ValidateSanityData(BlockchainRetriever, Transaction) (bool, bool, error) {
	return true, true, nil
}

func (punishGOVDecryptMetadata *PunishGOVDecryptMetadata) ValidateMetadataByItself() bool {
	return true
}

func (punishGOVDecryptMetadata *PunishGOVDecryptMetadata) CalculateSize() uint64 {
	return calculateSize(punishGOVDecryptMetadata)
}

func (punishGOVDecryptMetadata *PunishGOVDecryptMetadata) BuildReqActions(
	tx Transaction,
	bcr BlockchainRetriever,
	shardID byte,
) ([][]string, error) {
	inst := fromshardins.NewPunishDeryptIns(common.GOVBoard)

	instStr, err := inst.GetStringFormat()
	if err != nil {
		return nil, err
	}
	return [][]string{instStr}, nil
}
