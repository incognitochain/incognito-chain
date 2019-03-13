package metadata

import (
	"github.com/constant-money/constant-chain/blockchain/component"
	"github.com/constant-money/constant-chain/common"
	"github.com/constant-money/constant-chain/database"
	"github.com/constant-money/constant-chain/metadata/fromshardins"
	"github.com/constant-money/constant-chain/privacy"
)


func (normalGOVVoteProposalMetadata *NormalGOVVoteProposalMetadata) GetBoardType() common.BoardType {
	return common.GOVBoard
}
type NormalGOVVoteProposalMetadata struct {
	NormalVoteProposalMetadata component.VoteProposalData
	MetadataBase
}

func (normalGOVVoteProposalMetadata *NormalGOVVoteProposalMetadata) ValidateSanityData(bcr BlockchainRetriever, tx Transaction) (bool, bool, error) {
	//return normalGOVVoteProposalMetadata.NormalVoteProposalMetadata.ValidateSanityData(bcr, tx)
	return true, true, nil
}

func (normalGOVVoteProposalMetadata *NormalGOVVoteProposalMetadata) ValidateMetadataByItself() bool {
	//return normalGOVVoteProposalMetadata.NormalVoteProposalMetadata.ValidateMetadataByItself()
	return true
}

func NewNormalGOVVoteProposalMetadata(
	voteProposal component.VoteProposalData,
	lockerPaymentAddress []privacy.PaymentAddress,
	pointerToLv3VoteProposal common.Hash,
) *NormalGOVVoteProposalMetadata {
	return &NormalGOVVoteProposalMetadata{
		NormalVoteProposalMetadata: voteProposal,
		MetadataBase: *NewMetadataBase(NormalGOVVoteProposalMeta),
	}
}

func (normalGOVVoteProposalMetadata *NormalGOVVoteProposalMetadata) Hash() *common.Hash {
	record := normalGOVVoteProposalMetadata.NormalVoteProposalMetadata.ToBytes()

	hash := common.DoubleHashH([]byte(record))
	return &hash
}

func (normalGOVVoteProposalMetadata *NormalGOVVoteProposalMetadata) ValidateTxWithBlockChain(tx Transaction, bcr BlockchainRetriever, shardID byte, db database.DatabaseInterface) (bool, error) {
	//Validate these pubKeys are in board
	//boardType := common.GOVBoard
	//return normalGOVVoteProposalMetadata.NormalVoteProposalMetadata.ValidateTxWithBlockChain(
	//	boardType,
	//	tx,
	//	bcr,
	//	shardID,
	//	db,
	//)
	return true, nil
}

func (normalGOVVoteProposalMetadata *NormalGOVVoteProposalMetadata) BuildReqActions(
	tx Transaction,
	bcr BlockchainRetriever,
	shardID byte,
) ([][]string, error) {
	voteProposal := normalGOVVoteProposalMetadata.NormalVoteProposalMetadata
	inst := fromshardins.NewNormalVoteProposalIns(common.GOVBoard,  voteProposal)

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
