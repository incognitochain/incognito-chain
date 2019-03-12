package metadata

import (
	"github.com/constant-money/constant-chain/blockchain/component"
	"github.com/constant-money/constant-chain/common"
	"github.com/constant-money/constant-chain/database"
	"github.com/constant-money/constant-chain/metadata/fromshardins"
	"github.com/constant-money/constant-chain/privacy"
)

type NormalGOVVoteProposalFromSealerMetadata struct {
	NormalVoteProposalFromSealerMetadata NormalVoteProposalFromSealerMetadata

	MetadataBase
}

func (normalGOVVoteProposalFromSealerMetadata *NormalGOVVoteProposalFromSealerMetadata) ValidateSanityData(bcr BlockchainRetriever, tx Transaction) (bool, bool, error) {
	return normalGOVVoteProposalFromSealerMetadata.NormalVoteProposalFromSealerMetadata.ValidateSanityData(bcr, tx)
}

func (normalGOVVoteProposalFromSealerMetadata *NormalGOVVoteProposalFromSealerMetadata) ValidateMetadataByItself() bool {
	return normalGOVVoteProposalFromSealerMetadata.NormalVoteProposalFromSealerMetadata.ValidateMetadataByItself()
}

func NewNormalGOVVoteProposalFromSealerMetadata(
	voteProposal component.VoteProposalData,
	lockerPaymentAddress []privacy.PaymentAddress,
	pointerToLv1VoteProposal common.Hash,
	pointerToLv3VoteProposal common.Hash,
) *NormalGOVVoteProposalFromSealerMetadata {
	return &NormalGOVVoteProposalFromSealerMetadata{
		NormalVoteProposalFromSealerMetadata: *NewNormalVoteProposalFromSealerMetadata(
			voteProposal,
			lockerPaymentAddress,
			pointerToLv1VoteProposal,
			pointerToLv3VoteProposal,
		),
		MetadataBase: *NewMetadataBase(NormalGOVVoteProposalFromSealerMeta),
	}
}

func (normalGOVVoteProposalFromSealerMetadata *NormalGOVVoteProposalFromSealerMetadata) GetBoardType() common.BoardType {
	return common.GOVBoard
}

func (normalGOVVoteProposalFromSealerMetadata *NormalGOVVoteProposalFromSealerMetadata) ValidateTxWithBlockChain(tx Transaction, bcr BlockchainRetriever, shardID byte, db database.DatabaseInterface) (bool, error) {
	//Validate these pubKeys are in board
	boardType := normalGOVVoteProposalFromSealerMetadata.GetBoardType()
	return normalGOVVoteProposalFromSealerMetadata.NormalVoteProposalFromSealerMetadata.ValidateTxWithBlockChain(
		boardType,
		tx,
		bcr,
		shardID,
		db,
	)
}

func (normalGOVVoteProposalFromSealerMetadata *NormalGOVVoteProposalFromSealerMetadata) BuildReqActions(
	tx Transaction,
	bcr BlockchainRetriever,
	shardID byte,
) ([][]string, error) {
	lv3TxID := normalGOVVoteProposalFromSealerMetadata.NormalVoteProposalFromSealerMetadata.PointerToLv3VoteProposal
	voteProposal := normalGOVVoteProposalFromSealerMetadata.NormalVoteProposalFromSealerMetadata.VoteProposal
	inst := fromshardins.NewNormalVoteProposalFromSealerIns(common.GOVBoard, lv3TxID, voteProposal)

	instStr, err := inst.GetStringFormat()
	if err != nil {
		return nil, err
	}
	return [][]string{instStr}, nil
}

type NormalGOVVoteProposalFromOwnerMetadata struct {
	NormalVoteProposalFromOwnerMetadata NormalVoteProposalFromOwnerMetadata
	MetadataBase
}

func (normalGOVVoteProposalFromOwnerMetadata *NormalGOVVoteProposalFromOwnerMetadata) ValidateSanityData(bcr BlockchainRetriever, tx Transaction) (bool, bool, error) {
	return normalGOVVoteProposalFromOwnerMetadata.NormalVoteProposalFromOwnerMetadata.ValidateSanityData(bcr, tx)
}

func (normalGOVVoteProposalFromOwnerMetadata *NormalGOVVoteProposalFromOwnerMetadata) ValidateMetadataByItself() bool {
	return normalGOVVoteProposalFromOwnerMetadata.NormalVoteProposalFromOwnerMetadata.ValidateMetadataByItself()
}

func NewNormalGOVVoteProposalFromOwnerMetadata(
	voteProposal component.VoteProposalData,
	lockerPaymentAddress []privacy.PaymentAddress,
	pointerToLv3VoteProposal common.Hash,
) *NormalGOVVoteProposalFromOwnerMetadata {
	return &NormalGOVVoteProposalFromOwnerMetadata{
		NormalVoteProposalFromOwnerMetadata: *NewNormalVoteProposalFromOwnerMetadata(
			voteProposal,
			lockerPaymentAddress,
			pointerToLv3VoteProposal,
		),
		MetadataBase: *NewMetadataBase(NormalGOVVoteProposalFromOwnerMeta),
	}
}

func (normalGOVVoteProposalFromOwnerMetadata *NormalGOVVoteProposalFromOwnerMetadata) Hash() *common.Hash {
	record := normalGOVVoteProposalFromOwnerMetadata.NormalVoteProposalFromOwnerMetadata.ToBytes()

	hash := common.DoubleHashH([]byte(record))
	return &hash
}

func (normalGOVVoteProposalFromOwnerMetadata *NormalGOVVoteProposalFromOwnerMetadata) ValidateTxWithBlockChain(tx Transaction, bcr BlockchainRetriever, shardID byte, db database.DatabaseInterface) (bool, error) {
	//Validate these pubKeys are in board
	boardType := common.GOVBoard
	return normalGOVVoteProposalFromOwnerMetadata.NormalVoteProposalFromOwnerMetadata.ValidateTxWithBlockChain(
		boardType,
		tx,
		bcr,
		shardID,
		db,
	)
}

func (normalGOVVoteProposalFromOwnerMetadata *NormalGOVVoteProposalFromOwnerMetadata) BuildReqActions(
	tx Transaction,
	bcr BlockchainRetriever,
	shardID byte,
) ([][]string, error) {
	lv3TxID := normalGOVVoteProposalFromOwnerMetadata.NormalVoteProposalFromOwnerMetadata.PointerToLv3VoteProposal
	voteProposal := normalGOVVoteProposalFromOwnerMetadata.NormalVoteProposalFromOwnerMetadata.VoteProposal
	inst := fromshardins.NewNormalVoteProposalFromOwnerIns(common.GOVBoard, lv3TxID, voteProposal)

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
