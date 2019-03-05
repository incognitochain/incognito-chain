package metadata

import (
	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/database"
	"github.com/ninjadotorg/constant/privacy"
)

type SealedLv1GOVVoteProposalMetadata struct {
	SealedLv1VoteProposalMetadata SealedLv1VoteProposalMetadata
	MetadataBase
}

func (sealedLv1GOVVoteProposalMetadata *SealedLv1GOVVoteProposalMetadata) ProcessWhenInsertBlockShard(tx Transaction, bcr BlockchainRetriever) error {
	boardType := common.GOVBoard
	nextConstitutionIndex := bcr.GetConstitution(boardType).GetConstitutionIndex() + 1
	err := bcr.GetDatabase().AddVoteLv1or2Proposal(boardType, nextConstitutionIndex, tx.GetMetadata().Hash())
	if err != nil {
		return err
	}
	return nil
}

func (sealedLv1GOVVoteProposalMetadata *SealedLv1GOVVoteProposalMetadata) ValidateSanityData(bcr BlockchainRetriever, tx Transaction) (bool, bool, error) {
	return sealedLv1GOVVoteProposalMetadata.SealedLv1VoteProposalMetadata.ValidateSanityData(bcr, tx)
}

func (sealedLv1GOVVoteProposalMetadata *SealedLv1GOVVoteProposalMetadata) ValidateMetadataByItself() bool {
	return sealedLv1GOVVoteProposalMetadata.SealedLv1VoteProposalMetadata.ValidateMetadataByItself()
}

func (sealedLv1GOVVoteProposalMetadata *SealedLv1GOVVoteProposalMetadata) GetBoardType() common.BoardType {
	return common.GOVBoard
}

func NewSealedLv1GOVVoteProposalMetadata(
	sealedVoteProposal []byte,
	lockersPaymentAddress []privacy.PaymentAddress,
	pointerToLv2VoteProposal common.Hash,
	pointerToLv3VoteProposal common.Hash,
) *SealedLv1GOVVoteProposalMetadata {
	return &SealedLv1GOVVoteProposalMetadata{
		SealedLv1VoteProposalMetadata: *NewSealedLv1VoteProposalMetadata(
			sealedVoteProposal,
			lockersPaymentAddress,
			pointerToLv2VoteProposal,
			pointerToLv3VoteProposal,
		),

		MetadataBase: *NewMetadataBase(SealedLv1GOVVoteProposalMeta),
	}
}

func (sealedLv1GOVVoteProposalMetadata *SealedLv1GOVVoteProposalMetadata) Hash() *common.Hash {
	record := string(sealedLv1GOVVoteProposalMetadata.SealedLv1VoteProposalMetadata.ToBytes())

	hash := common.DoubleHashH([]byte(record))
	return &hash

}

func (sealedLv1GOVVoteProposalMetadata *SealedLv1GOVVoteProposalMetadata) ValidateTxWithBlockChain(
	tx Transaction,
	bcr BlockchainRetriever,
	shardID byte,
	db database.DatabaseInterface,
) (bool, error) {
	boardType := common.GOVBoard
	return sealedLv1GOVVoteProposalMetadata.SealedLv1VoteProposalMetadata.ValidateTxWithBlockChain(
		boardType,
		tx,
		bcr,
		shardID,
		db,
	)
}

type SealedLv2GOVVoteProposalMetadata struct {
	SealedLv2VoteProposalMetadata SealedLv2VoteProposalMetadata

	MetadataBase
}

func (sealedLv2GOVVoteProposalMetadata *SealedLv2GOVVoteProposalMetadata) ProcessWhenInsertBlockShard(tx Transaction, bcr BlockchainRetriever) error {
	boardType := common.GOVBoard
	nextConstitutionIndex := bcr.GetConstitution(boardType).GetConstitutionIndex() + 1
	err := bcr.GetDatabase().AddVoteLv1or2Proposal(boardType, nextConstitutionIndex, tx.GetMetadata().Hash())
	if err != nil {
		return err
	}
	return nil
}

func (sealedLv2GOVVoteProposalMetadata *SealedLv2GOVVoteProposalMetadata) ValidateSanityData(bcr BlockchainRetriever, tx Transaction) (bool, bool, error) {
	return sealedLv2GOVVoteProposalMetadata.SealedLv2VoteProposalMetadata.ValidateSanityData(bcr, tx)
}

func (sealedLv2GOVVoteProposalMetadata *SealedLv2GOVVoteProposalMetadata) ValidateMetadataByItself() bool {
	return sealedLv2GOVVoteProposalMetadata.SealedLv2VoteProposalMetadata.ValidateMetadataByItself()
}

func (sealedLv2GOVVoteProposalMetadata *SealedLv2GOVVoteProposalMetadata) GetBoardType() common.BoardType {
	return common.GOVBoard
}

func NewSealedLv2GOVVoteProposalMetadata(
	sealedVoteProposal []byte,
	lockerPaymentAddress []privacy.PaymentAddress,
	pointerToLv3VoteProposal common.Hash,
) *SealedLv2GOVVoteProposalMetadata {
	return &SealedLv2GOVVoteProposalMetadata{
		SealedLv2VoteProposalMetadata: *NewSealedLv2VoteProposalMetadata(
			sealedVoteProposal,
			lockerPaymentAddress,
			pointerToLv3VoteProposal,
		),

		MetadataBase: *NewMetadataBase(SealedLv2GOVVoteProposalMeta),
	}
}

func (sealedLv2GOVVoteProposalMetadata *SealedLv2GOVVoteProposalMetadata) Hash() *common.Hash {
	record := sealedLv2GOVVoteProposalMetadata.SealedLv2VoteProposalMetadata.ToBytes()
	hash := common.DoubleHashH([]byte(record))
	return &hash
}

func (sealedLv2GOVVoteProposalMetadata *SealedLv2GOVVoteProposalMetadata) ValidateTxWithBlockChain(tx Transaction, bcr BlockchainRetriever, shardID byte, db database.DatabaseInterface) (bool, error) {
	boardType := sealedLv2GOVVoteProposalMetadata.GetBoardType()
	return sealedLv2GOVVoteProposalMetadata.SealedLv2VoteProposalMetadata.ValidateTxWithBlockChain(
		boardType,
		tx,
		bcr,
		shardID,
		db,
	)

}

type SealedLv3GOVVoteProposalMetadata struct {
	SealedLv3VoteProposalMetadata SealedLv3VoteProposalMetadata

	MetadataBase
}

func (sealedLv3GOVVoteProposalMetadata *SealedLv3GOVVoteProposalMetadata) ProcessWhenInsertBlockShard(tx Transaction, bcr BlockchainRetriever) error {
	boardType := common.GOVBoard
	nextConstitutionIndex := bcr.GetConstitution(boardType).GetConstitutionIndex() + 1
	err := bcr.GetDatabase().AddVoteLv3Proposal(boardType, nextConstitutionIndex, tx.GetMetadata().Hash())
	if err != nil {
		return err
	}
	return nil
}

func (sealedLv3GOVVoteProposalMetadata *SealedLv3GOVVoteProposalMetadata) ValidateTxWithBlockChain(tx Transaction, bcr BlockchainRetriever, b byte, db database.DatabaseInterface) (bool, error) {
	return sealedLv3GOVVoteProposalMetadata.SealedLv3VoteProposalMetadata.ValidateTxWithBlockChain(tx, bcr, b, db)
}

func (sealedLv3GOVVoteProposalMetadata *SealedLv3GOVVoteProposalMetadata) ValidateSanityData(bcr BlockchainRetriever, tx Transaction) (bool, bool, error) {
	return sealedLv3GOVVoteProposalMetadata.SealedLv3VoteProposalMetadata.ValidateSanityData(bcr, tx)
}

func (sealedLv3GOVVoteProposalMetadata *SealedLv3GOVVoteProposalMetadata) ValidateMetadataByItself() bool {
	return sealedLv3GOVVoteProposalMetadata.SealedLv3VoteProposalMetadata.ValidateMetadataByItself()
}

func (sealedLv3GOVVoteProposalMetadata *SealedLv3GOVVoteProposalMetadata) GetBoardType() common.BoardType {
	return common.GOVBoard
}

func NewSealedLv3GOVVoteProposalMetadata(
	sealedVoteProposal []byte,
	lockerPaymentAddress []privacy.PaymentAddress,
) *SealedLv3GOVVoteProposalMetadata {
	return &SealedLv3GOVVoteProposalMetadata{
		SealedLv3VoteProposalMetadata: *NewSealedLv3VoteProposalMetadata(
			sealedVoteProposal, lockerPaymentAddress,
		),
		MetadataBase: *NewMetadataBase(SealedLv3GOVVoteProposalMeta),
	}
}

type NormalGOVVoteProposalFromSealerMetadata struct {
	NormalVoteProposalFromSealerMetadata NormalVoteProposalFromSealerMetadata

	MetadataBase
}

func (normalGOVVoteProposalFromSealerMetadata *NormalGOVVoteProposalFromSealerMetadata) ProcessWhenInsertBlockShard(tx Transaction, bcr BlockchainRetriever) error {
	boardType := common.GOVBoard
	nextConstitutionIndex := bcr.GetConstitution(boardType).GetConstitutionIndex() + 1
	err := bcr.GetDatabase().AddVoteNormalProposalFromOwner(
		boardType,
		nextConstitutionIndex,
		&normalGOVVoteProposalFromSealerMetadata.NormalVoteProposalFromSealerMetadata.PointerToLv3VoteProposal,
		normalGOVVoteProposalFromSealerMetadata.NormalVoteProposalFromSealerMetadata.VoteProposal.ToBytes(),
	)
	if err != nil {
		return err
	}
	return nil
}

func (normalGOVVoteProposalFromSealerMetadata *NormalGOVVoteProposalFromSealerMetadata) ValidateSanityData(bcr BlockchainRetriever, tx Transaction) (bool, bool, error) {
	return normalGOVVoteProposalFromSealerMetadata.NormalVoteProposalFromSealerMetadata.ValidateSanityData(bcr, tx)
}

func (normalGOVVoteProposalFromSealerMetadata *NormalGOVVoteProposalFromSealerMetadata) ValidateMetadataByItself() bool {
	return normalGOVVoteProposalFromSealerMetadata.NormalVoteProposalFromSealerMetadata.ValidateMetadataByItself()
}

func NewNormalGOVVoteProposalFromSealerMetadata(
	voteProposal VoteProposalData,
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

type NormalGOVVoteProposalFromOwnerMetadata struct {
	NormalVoteProposalFromOwnerMetadata NormalVoteProposalFromOwnerMetadata
	MetadataBase
}

func (normalGOVVoteProposalFromOwnerMetadata *NormalGOVVoteProposalFromOwnerMetadata) ProcessWhenInsertBlockShard(tx Transaction, bcr BlockchainRetriever) error {
	boardType := common.GOVBoard
	nextConstitutionIndex := bcr.GetConstitution(boardType).GetConstitutionIndex() + 1
	err := bcr.GetDatabase().AddVoteNormalProposalFromOwner(
		boardType,
		nextConstitutionIndex,
		&normalGOVVoteProposalFromOwnerMetadata.NormalVoteProposalFromOwnerMetadata.PointerToLv3VoteProposal,
		normalGOVVoteProposalFromOwnerMetadata.NormalVoteProposalFromOwnerMetadata.VoteProposal.ToBytes(),
	)
	if err != nil {
		return err
	}
	return nil

}

func (normalGOVVoteProposalFromOwnerMetadata *NormalGOVVoteProposalFromOwnerMetadata) ValidateSanityData(bcr BlockchainRetriever, tx Transaction) (bool, bool, error) {
	return normalGOVVoteProposalFromOwnerMetadata.NormalVoteProposalFromOwnerMetadata.ValidateSanityData(bcr, tx)
}

func (normalGOVVoteProposalFromOwnerMetadata *NormalGOVVoteProposalFromOwnerMetadata) ValidateMetadataByItself() bool {
	return normalGOVVoteProposalFromOwnerMetadata.NormalVoteProposalFromOwnerMetadata.ValidateMetadataByItself()
}

func NewNormalGOVVoteProposalFromOwnerMetadata(
	voteProposal VoteProposalData,
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

type PunishGOVDecryptMetadata struct {
	PunishDecryptMetadata PunishDecryptMetadata
	MetadataBase
}

func (punishGOVDecryptMetadata *PunishGOVDecryptMetadata) ProcessWhenInsertBlockShard(tx Transaction, bcr BlockchainRetriever) error {
	// todo @0xjackalope
	return nil
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
