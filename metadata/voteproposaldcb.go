package metadata

import (
	"github.com/constant-money/constant-chain/blockchain/component"
	"github.com/constant-money/constant-chain/common"
	"github.com/constant-money/constant-chain/database"
	"github.com/constant-money/constant-chain/metadata/fromshardins"
	"github.com/constant-money/constant-chain/privacy"
)
type NormalDCBVoteProposalFromSealerMetadata struct {
	NormalVoteProposalFromSealerMetadata NormalVoteProposalFromSealerMetadata

	MetadataBase
}

func (normalDCBVoteProposalFromSealerMetadata *NormalDCBVoteProposalFromSealerMetadata) ValidateSanityData(bcr BlockchainRetriever, tx Transaction) (bool, bool, error) {
	return normalDCBVoteProposalFromSealerMetadata.NormalVoteProposalFromSealerMetadata.ValidateSanityData(bcr, tx)
}

func (normalDCBVoteProposalFromSealerMetadata *NormalDCBVoteProposalFromSealerMetadata) ValidateMetadataByItself() bool {
	return normalDCBVoteProposalFromSealerMetadata.NormalVoteProposalFromSealerMetadata.ValidateMetadataByItself()
}

func NewNormalDCBVoteProposalFromSealerMetadata(
	voteProposal component.VoteProposalData,
	lockerPaymentAddress []privacy.PaymentAddress,
	pointerToLv1VoteProposal common.Hash,
	pointerToLv3VoteProposal common.Hash,
) *NormalDCBVoteProposalFromSealerMetadata {
	return &NormalDCBVoteProposalFromSealerMetadata{
		NormalVoteProposalFromSealerMetadata: *NewNormalVoteProposalFromSealerMetadata(
			voteProposal,
			lockerPaymentAddress,
			pointerToLv1VoteProposal,
			pointerToLv3VoteProposal,
		),
		MetadataBase: *NewMetadataBase(NormalDCBVoteProposalFromSealerMeta),
	}
}

func (normalDCBVoteProposalFromSealerMetadata *NormalDCBVoteProposalFromSealerMetadata) GetBoardType() common.BoardType {
	return common.DCBBoard
}

func (normalDCBVoteProposalFromSealerMetadata *NormalDCBVoteProposalFromSealerMetadata) ValidateTxWithBlockChain(tx Transaction, bcr BlockchainRetriever, shardID byte, db database.DatabaseInterface) (bool, error) {
	//Validate these pubKeys are in board
	boardType := normalDCBVoteProposalFromSealerMetadata.GetBoardType()
	return normalDCBVoteProposalFromSealerMetadata.NormalVoteProposalFromSealerMetadata.ValidateTxWithBlockChain(
		boardType,
		tx,
		bcr,
		shardID,
		db,
	)
}

func (normalDCBVoteProposalFromSealerMetadata *NormalDCBVoteProposalFromSealerMetadata) BuildReqActions(
	tx Transaction,
	bcr BlockchainRetriever,
	shardID byte,
) ([][]string, error) {
	lv3TxID := normalDCBVoteProposalFromSealerMetadata.NormalVoteProposalFromSealerMetadata.PointerToLv3VoteProposal
	voteProposal := normalDCBVoteProposalFromSealerMetadata.NormalVoteProposalFromSealerMetadata.VoteProposal
	inst := fromshardins.NewNormalVoteProposalFromSealerIns(common.DCBBoard, lv3TxID, voteProposal)

	instStr, err := inst.GetStringFormat()
	if err != nil {
		return nil, err
	}
	return [][]string{instStr}, nil
}

type NormalDCBVoteProposalFromOwnerMetadata struct {
	NormalVoteProposalFromOwnerMetadata NormalVoteProposalFromOwnerMetadata
	MetadataBase
}

func (normalDCBVoteProposalFromOwnerMetadata *NormalDCBVoteProposalFromOwnerMetadata) ValidateSanityData(bcr BlockchainRetriever, tx Transaction) (bool, bool, error) {
	return normalDCBVoteProposalFromOwnerMetadata.NormalVoteProposalFromOwnerMetadata.ValidateSanityData(bcr, tx)
}

func (normalDCBVoteProposalFromOwnerMetadata *NormalDCBVoteProposalFromOwnerMetadata) ValidateMetadataByItself() bool {
	return normalDCBVoteProposalFromOwnerMetadata.NormalVoteProposalFromOwnerMetadata.ValidateMetadataByItself()
}

func NewNormalDCBVoteProposalFromOwnerMetadata(
	voteProposal component.VoteProposalData,
	lockerPaymentAddress []privacy.PaymentAddress,
	pointerToLv3VoteProposal common.Hash,
) *NormalDCBVoteProposalFromOwnerMetadata {
	return &NormalDCBVoteProposalFromOwnerMetadata{
		NormalVoteProposalFromOwnerMetadata: *NewNormalVoteProposalFromOwnerMetadata(
			voteProposal,
			lockerPaymentAddress,
			pointerToLv3VoteProposal,
		),
		MetadataBase: *NewMetadataBase(NormalDCBVoteProposalFromOwnerMeta),
	}
}

func (normalDCBVoteProposalFromOwnerMetadata *NormalDCBVoteProposalFromOwnerMetadata) Hash() *common.Hash {
	record := normalDCBVoteProposalFromOwnerMetadata.NormalVoteProposalFromOwnerMetadata.ToBytes()

	hash := common.DoubleHashH([]byte(record))
	return &hash
}

func (normalDCBVoteProposalFromOwnerMetadata *NormalDCBVoteProposalFromOwnerMetadata) ValidateTxWithBlockChain(tx Transaction, bcr BlockchainRetriever, shardID byte, db database.DatabaseInterface) (bool, error) {
	//Validate these pubKeys are in board
	boardType := common.DCBBoard
	return normalDCBVoteProposalFromOwnerMetadata.NormalVoteProposalFromOwnerMetadata.ValidateTxWithBlockChain(
		boardType,
		tx,
		bcr,
		shardID,
		db,
	)
}

func (normalDCBVoteProposalFromOwnerMetadata *NormalDCBVoteProposalFromOwnerMetadata) BuildReqActions(
	tx Transaction,
	bcr BlockchainRetriever,
	shardID byte,
) ([][]string, error) {
	lv3TxID := normalDCBVoteProposalFromOwnerMetadata.NormalVoteProposalFromOwnerMetadata.PointerToLv3VoteProposal
	voteProposal := normalDCBVoteProposalFromOwnerMetadata.NormalVoteProposalFromOwnerMetadata.VoteProposal
	inst := fromshardins.NewNormalVoteProposalFromOwnerIns(common.DCBBoard, lv3TxID, voteProposal)

	instStr, err := inst.GetStringFormat()
	if err != nil {
		return nil, err
	}
	return [][]string{instStr}, nil
}

type PunishDCBDecryptMetadata struct {
	PunishDecryptMetadata PunishDecryptMetadata
	MetadataBase
}

func NewPunishDCBDecryptMetadata(paymentAddress privacy.PaymentAddress) *PunishDCBDecryptMetadata {
	return &PunishDCBDecryptMetadata{
		PunishDecryptMetadata: PunishDecryptMetadata{
			PaymentAddress: paymentAddress,
		},
		MetadataBase: *NewMetadataBase(PunishDCBDecryptMeta),
	}
}

func (punishDCBDecryptMetadata *PunishDCBDecryptMetadata) Hash() *common.Hash {
	record := string(punishDCBDecryptMetadata.PunishDecryptMetadata.ToBytes())
	record += punishDCBDecryptMetadata.MetadataBase.Hash().String()

	hash := common.DoubleHashH([]byte(record))
	return &hash
}

func (punishDCBDecryptMetadata *PunishDCBDecryptMetadata) ValidateTxWithBlockChain(Transaction, BlockchainRetriever, byte, database.DatabaseInterface) (bool, error) {
	return true, nil
}

func (punishDCBDecryptMetadata *PunishDCBDecryptMetadata) ValidateSanityData(BlockchainRetriever, Transaction) (bool, bool, error) {
	return true, true, nil
}

func (punishDCBDecryptMetadata *PunishDCBDecryptMetadata) ValidateMetadataByItself() bool {
	return true
}

func (punishDCBDecryptMetadata *PunishDCBDecryptMetadata) CalculateSize() uint64 {
	return calculateSize(punishDCBDecryptMetadata)
}

func (punishDCBDecryptMetadata *PunishDCBDecryptMetadata) BuildReqActions(
	tx Transaction,
	bcr BlockchainRetriever,
	shardID byte,
) ([][]string, error) {
	inst := fromshardins.NewPunishDeryptIns(common.DCBBoard)

	instStr, err := inst.GetStringFormat()
	if err != nil {
		return nil, err
	}
	return [][]string{instStr}, nil
}
