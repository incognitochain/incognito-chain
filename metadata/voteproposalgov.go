package metadata

import (
	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/database"
	"github.com/ninjadotorg/constant/privacy"
)

type SealedLv1GOVVoteProposalMetadata struct {
	SealedLv1VoteProposalMetadata
}

func (sealedLv1GOVVoteProposalMetadata *SealedLv1GOVVoteProposalMetadata) GetBoardType() string {
	return "gov"
}

func NewSealedLv1GOVVoteProposalMetadata(
	sealedVoteProposal []byte,
	lockersPubKey [][]byte,
	pointerToLv2VoteProposal common.Hash,
	pointerToLv3VoteProposal common.Hash,
) *SealedLv1GOVVoteProposalMetadata {
	return &SealedLv1GOVVoteProposalMetadata{
		SealedLv1VoteProposalMetadata: *NewSealedLv1VoteProposalMetadata(
			sealedVoteProposal,
			lockersPubKey,
			pointerToLv2VoteProposal,
			pointerToLv3VoteProposal,
			*NewMetadataBase(SealedLv1GOVVoteProposalMeta),
		),
	}
}

func (sealedLv1GOVVoteProposalMetadata *SealedLv1GOVVoteProposalMetadata) Hash() *common.Hash {
	return sealedLv1GOVVoteProposalMetadata.SealedLv1VoteProposalMetadata.Hash2()
}

func (sealedLv1GOVVoteProposalMetadata *SealedLv1GOVVoteProposalMetadata) ValidateTxWithBlockChain(
	tx Transaction,
	bcr BlockchainRetriever,
	chainID byte,
	db database.DatabaseInterface,
) (bool, error) {
	//Check base seal metadata
	ok, err := sealedLv1GOVVoteProposalMetadata.SealedVoteProposal.ValidateLockerPubKeys(bcr, sealedLv1GOVVoteProposalMetadata.GetBoardType())
	if err != nil || !ok {
		return ok, err
	}

	//Check precede transaction type
	_, _, _, lv2Tx, _ := bcr.GetTransactionByHash(&sealedLv1GOVVoteProposalMetadata.PointerToLv2VoteProposal)
	if lv2Tx.GetMetadataType() != SealedLv2GOVVoteProposalMeta {
		return false, nil
	}
	_, _, _, lv3Tx, _ := bcr.GetTransactionByHash(&sealedLv1GOVVoteProposalMetadata.PointerToLv3VoteProposal)
	if lv3Tx.GetMetadataType() != SealedLv3GOVVoteProposalMeta {
		return false, nil
	}

	// check 2 array equal
	metaLv2 := lv2Tx.GetMetadata().(*SealedLv2GOVVoteProposalMetadata)
	for i := 0; i < len(sealedLv1GOVVoteProposalMetadata.SealedVoteProposal.LockerPubKeys); i++ {
		if !common.ByteEqual(sealedLv1GOVVoteProposalMetadata.SealedVoteProposal.LockerPubKeys[i], metaLv2.SealedVoteProposal.LockerPubKeys[i]) {
			return false, nil
		}
	}

	// Check encrypting
	if !common.ByteEqual(sealedLv1GOVVoteProposalMetadata.SealedVoteProposal.SealVoteProposalData,
		common.Encrypt(metaLv2.SealedVoteProposal.SealVoteProposalData, metaLv2.SealedVoteProposal.LockerPubKeys[1])) {
		return false, nil
	}
	return true, nil
}

type SealedLv2GOVVoteProposalMetadata struct {
	SealedLv2VoteProposalMetadata
}

func (sealedLv2GOVVoteProposalMetadata *SealedLv2GOVVoteProposalMetadata) GetBoardType() string {
	return "gov"
}

func NewSealedLv2GOVVoteProposalMetadata(
	sealedVoteProposal []byte,
	lockerPubKeys [][]byte,
	pointerToLv3VoteProposal common.Hash,
) *SealedLv2GOVVoteProposalMetadata {
	return &SealedLv2GOVVoteProposalMetadata{
		SealedLv2VoteProposalMetadata: *NewSealedLv2VoteProposalMetadata(
			sealedVoteProposal,
			lockerPubKeys,
			pointerToLv3VoteProposal,
			*NewMetadataBase(SealedLv2GOVVoteProposalMeta),
		),
	}
}

func (sealedLv2GOVVoteProposalMetadata *SealedLv2GOVVoteProposalMetadata) Hash() *common.Hash {
	return sealedLv2GOVVoteProposalMetadata.SealedLv2VoteProposalMetadata.Hash()
}

func (sealedLv2GOVVoteProposalMetadata *SealedLv2GOVVoteProposalMetadata) ValidateTxWithBlockChain(tx Transaction, bcr BlockchainRetriever, chainID byte, db database.DatabaseInterface) (bool, error) {
	boardType := sealedLv2GOVVoteProposalMetadata.GetBoardType()
	//Check base seal metadata
	ok, err := sealedLv2GOVVoteProposalMetadata.SealedVoteProposal.ValidateLockerPubKeys(bcr, boardType)
	if err != nil || !ok {
		return ok, err
	}

	//Check precede transaction type
	_, _, _, lv3Tx, _ := bcr.GetTransactionByHash(&sealedLv2GOVVoteProposalMetadata.PointerToLv3VoteProposal)
	if lv3Tx.GetMetadataType() != SealedLv3GOVVoteProposalMeta {
		return false, nil
	}

	// check 2 array equal
	metaLv3 := lv3Tx.GetMetadata().(*SealedLv3GOVVoteProposalMetadata)
	for i := 0; i < len(sealedLv2GOVVoteProposalMetadata.SealedVoteProposal.LockerPubKeys); i++ {
		if !common.ByteEqual(sealedLv2GOVVoteProposalMetadata.SealedVoteProposal.LockerPubKeys[i], metaLv3.SealedVoteProposal.LockerPubKeys[i]) {
			return false, nil
		}
	}

	// Check encrypting
	if !common.ByteEqual(
		sealedLv2GOVVoteProposalMetadata.SealedVoteProposal.SealVoteProposalData,
		common.Encrypt(metaLv3.SealedVoteProposal.SealVoteProposalData, metaLv3.SealedVoteProposal.LockerPubKeys[2]),
	) {
		return false, nil
	}
	return true, nil
}

type SealedLv3GOVVoteProposalMetadata struct {
	SealedLv3VoteProposalMetadata
}

func (sealedLv3GOVVoteProposalMetadata *SealedLv3GOVVoteProposalMetadata) GetBoardType() string {
	return "gov"
}

func NewSealedLv3GOVVoteProposalMetadata(
	sealedVoteProposal []byte,
	lockerPubKeys [][]byte,
) *SealedLv3GOVVoteProposalMetadata {
	return &SealedLv3GOVVoteProposalMetadata{
		SealedLv3VoteProposalMetadata: *NewSealedLv3VoteProposalMetadata(
			sealedVoteProposal, lockerPubKeys,
			*NewMetadataBase(SealedLv3GOVVoteProposalMeta),
		),
	}
}

func NewSealedLv3GOVVoteProposalMetadataFromJson(data interface{}) *SealedLv3GOVVoteProposalMetadata {
	dataSealedLv3GOVVoteProposal := data.(map[string]interface{})
	return NewSealedLv3GOVVoteProposalMetadata(
		[]byte(dataSealedLv3GOVVoteProposal["SealedVoteProposal"].(string)),
		common.SliceInterfaceToSliceSliceByte(dataSealedLv3GOVVoteProposal["LockerPubKeys"].([]interface{})),
	)
}

type NormalGOVVoteProposalFromSealerMetadata struct {
	NormalVoteProposalFromSealerMetadata
}

func NewNormalGOVVoteProposalFromSealerMetadata(
	voteProposal VoteProposalData,
	lockerPubKey [][]byte,
	pointerToLv1VoteProposal common.Hash,
	pointerToLv3VoteProposal common.Hash,
) *NormalGOVVoteProposalFromSealerMetadata {
	return &NormalGOVVoteProposalFromSealerMetadata{
		NormalVoteProposalFromSealerMetadata: *NewNormalVoteProposalFromSealerMetadata(
			voteProposal,
			lockerPubKey,
			pointerToLv1VoteProposal,
			pointerToLv3VoteProposal,
			*NewMetadataBase(NormalGOVVoteProposalFromSealerMeta),
		),
	}
}

func (normalGOVVoteProposalFromSealerMetadata *NormalGOVVoteProposalFromSealerMetadata) GetBoardType() string {
	return "gov"
}

func (normalGOVVoteProposalFromSealerMetadata *NormalGOVVoteProposalFromSealerMetadata) ValidateTxWithBlockChain(tx Transaction, bcr BlockchainRetriever, chainID byte, db database.DatabaseInterface) (bool, error) {
	//Validate these pubKeys are in board
	boardType := normalGOVVoteProposalFromSealerMetadata.GetBoardType()
	govBoardPubKeys := bcr.GetBoardPubKeys(boardType)
	for _, j := range normalGOVVoteProposalFromSealerMetadata.LockerPubKey {
		exist := false
		for _, i := range govBoardPubKeys {
			if common.ByteEqual(i, j) {
				exist = true
				break
			}
		}
		if !exist {
			return false, nil
		}
	}

	//Check precede transaction type
	_, _, _, lv1Tx, _ := bcr.GetTransactionByHash(&normalGOVVoteProposalFromSealerMetadata.PointerToLv1VoteProposal)
	if lv1Tx.GetMetadataType() != SealedLv1GOVVoteProposalMeta {
		return false, nil
	}
	_, _, _, lv3Tx, _ := bcr.GetTransactionByHash(&normalGOVVoteProposalFromSealerMetadata.PointerToLv3VoteProposal)
	if lv3Tx.GetMetadataType() != SealedLv3GOVVoteProposalMeta {
		return false, nil
	}

	// check 2 array equal
	metaLv1 := lv1Tx.GetMetadata().(*SealedLv1GOVVoteProposalMetadata)
	for i := 0; i < len(normalGOVVoteProposalFromSealerMetadata.LockerPubKey); i++ {
		if !common.ByteEqual(normalGOVVoteProposalFromSealerMetadata.LockerPubKey[i], metaLv1.SealedVoteProposal.LockerPubKeys[i]) {
			return false, nil
		}
	}

	// Check encrypting
	if !common.ByteEqual(normalGOVVoteProposalFromSealerMetadata.VoteProposal.ToBytes(), common.Encrypt(metaLv1.SealedVoteProposal.SealVoteProposalData, metaLv1.SealedVoteProposal.LockerPubKeys[0])) {
		return false, nil
	}
	return true, nil
}

type NormalGOVVoteProposalFromOwnerMetadata struct {
	NormalVoteProposalFromOwnerMetadata
}

func NewNormalGOVVoteProposalFromOwnerMetadata(
	voteProposal VoteProposalData,
	lockerPubKey [][]byte,
	pointerToLv3VoteProposal common.Hash,
) *NormalGOVVoteProposalFromOwnerMetadata {
	return &NormalGOVVoteProposalFromOwnerMetadata{
		NormalVoteProposalFromOwnerMetadata: *NewNormalVoteProposalFromOwnerMetadata(
			voteProposal,
			lockerPubKey,
			pointerToLv3VoteProposal,
			*NewMetadataBase(NormalGOVVoteProposalFromOwnerMeta),
		),
	}
}

func (normalGOVVoteProposalFromOwnerMetadata *NormalGOVVoteProposalFromOwnerMetadata) Hash() *common.Hash {
	return normalGOVVoteProposalFromOwnerMetadata.NormalVoteProposalFromOwnerMetadata.Hash()
}

func (normalGOVVoteProposalFromOwnerMetadata *NormalGOVVoteProposalFromOwnerMetadata) ValidateTxWithBlockChain(tx Transaction, bcr BlockchainRetriever, chainID byte, db database.DatabaseInterface) (bool, error) {
	//Validate these pubKeys are in board
	govBoardPubKeys := bcr.GetBoardPubKeys("gov")
	for _, j := range normalGOVVoteProposalFromOwnerMetadata.LockerPubKey {
		exist := false
		for _, i := range govBoardPubKeys {
			if common.ByteEqual(i, j) {
				exist = true
				break
			}
		}
		if !exist {
			return false, nil
		}
	}

	//Check precede transaction type
	_, _, _, lv3Tx, _ := bcr.GetTransactionByHash(&normalGOVVoteProposalFromOwnerMetadata.PointerToLv3VoteProposal)
	if lv3Tx.GetMetadataType() != SealedLv3GOVVoteProposalMeta {
		return false, nil
	}

	// check 2 array equal
	metaLv3 := lv3Tx.GetMetadata().(*SealedLv3GOVVoteProposalMetadata)
	for i := 0; i < len(normalGOVVoteProposalFromOwnerMetadata.LockerPubKey); i++ {
		if !common.ByteEqual(normalGOVVoteProposalFromOwnerMetadata.LockerPubKey[i], metaLv3.SealedVoteProposal.LockerPubKeys[i]) {
			return false, nil
		}
	}

	// Check encrypting
	if !common.ByteEqual(
		metaLv3.SealedVoteProposal.SealVoteProposalData,
		common.Encrypt(
			common.Encrypt(
				common.Encrypt(
					normalGOVVoteProposalFromOwnerMetadata.VoteProposal.ToBytes(),
					metaLv3.SealedVoteProposal.LockerPubKeys[2],
				),
				metaLv3.SealedVoteProposal.LockerPubKeys[1],
			),
			metaLv3.SealedVoteProposal.LockerPubKeys[0],
		)) {
		return false, nil
	}
	return true, nil
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
	record += string(punishGOVDecryptMetadata.MetadataBase.Hash().GetBytes())
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
