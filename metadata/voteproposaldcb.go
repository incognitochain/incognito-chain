package metadata

import (
	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/database"
	"github.com/ninjadotorg/constant/privacy"
)

type SealedLv1DCBVoteProposalMetadata struct {
	SealedLv1VoteProposalMetadata
}

func (sealedLv1DCBVoteProposalMetadata *SealedLv1DCBVoteProposalMetadata) GetBoardType() string {
	return "dcb"
}

func NewSealedLv1DCBVoteProposalMetadata(
	sealedVoteProposal []byte,
	lockersPubKey [][]byte,
	pointerToLv2VoteProposal common.Hash,
	pointerToLv3VoteProposal common.Hash,
) *SealedLv1DCBVoteProposalMetadata {
	return &SealedLv1DCBVoteProposalMetadata{
		SealedLv1VoteProposalMetadata: *NewSealedLv1VoteProposalMetadata(
			sealedVoteProposal,
			lockersPubKey,
			pointerToLv2VoteProposal,
			pointerToLv3VoteProposal,
			*NewMetadataBase(SealedLv1DCBVoteProposalMeta),
		),
	}
}

func (sealedLv1DCBVoteProposalMetadata *SealedLv1DCBVoteProposalMetadata) Hash() *common.Hash {
	return sealedLv1DCBVoteProposalMetadata.SealedLv1VoteProposalMetadata.Hash2()
}

func (sealedLv1DCBVoteProposalMetadata *SealedLv1DCBVoteProposalMetadata) ValidateTxWithBlockChain(
	tx Transaction,
	bcr BlockchainRetriever,
	chainID byte,
	db database.DatabaseInterface,
) (bool, error) {
	//Check base seal metadata
	ok, err := sealedLv1DCBVoteProposalMetadata.SealedVoteProposal.ValidateLockerPubKeys(bcr, sealedLv1DCBVoteProposalMetadata.GetBoardType())
	if err != nil || !ok {
		return ok, err
	}

	//Check precede transaction type
	_, _, _, lv2Tx, _ := bcr.GetTransactionByHash(&sealedLv1DCBVoteProposalMetadata.PointerToLv2VoteProposal)
	if lv2Tx.GetMetadataType() != SealedLv2DCBVoteProposalMeta {
		return false, nil
	}
	_, _, _, lv3Tx, _ := bcr.GetTransactionByHash(&sealedLv1DCBVoteProposalMetadata.PointerToLv3VoteProposal)
	if lv3Tx.GetMetadataType() != SealedLv3DCBVoteProposalMeta {
		return false, nil
	}

	// check 2 array equal
	metaLv2 := lv2Tx.GetMetadata().(*SealedLv2DCBVoteProposalMetadata)
	for i := 0; i < len(sealedLv1DCBVoteProposalMetadata.SealedVoteProposal.LockerPubKeys); i++ {
		if !common.ByteEqual(sealedLv1DCBVoteProposalMetadata.SealedVoteProposal.LockerPubKeys[i], metaLv2.SealedVoteProposal.LockerPubKeys[i]) {
			return false, nil
		}
	}

	// Check encrypting
	if !common.ByteEqual(sealedLv1DCBVoteProposalMetadata.SealedVoteProposal.SealVoteProposalData,
		common.Encrypt(metaLv2.SealedVoteProposal.SealVoteProposalData, metaLv2.SealedVoteProposal.LockerPubKeys[1])) {
		return false, nil
	}
	return true, nil
}

type SealedLv2DCBVoteProposalMetadata struct {
	SealedLv2VoteProposalMetadata
}

func (sealedLv2DCBVoteProposalMetadata *SealedLv2DCBVoteProposalMetadata) GetBoardType() string {
	return "dcb"
}

func NewSealedLv2DCBVoteProposalMetadata(
	sealedVoteProposal []byte,
	lockerPubKeys [][]byte,
	pointerToLv3VoteProposal common.Hash,
) *SealedLv2DCBVoteProposalMetadata {
	return &SealedLv2DCBVoteProposalMetadata{
		SealedLv2VoteProposalMetadata: *NewSealedLv2VoteProposalMetadata(
			sealedVoteProposal,
			lockerPubKeys,
			pointerToLv3VoteProposal,
			*NewMetadataBase(SealedLv2DCBVoteProposalMeta),
		),
	}
}

func (sealedLv2DCBVoteProposalMetadata *SealedLv2DCBVoteProposalMetadata) Hash() *common.Hash {
	return sealedLv2DCBVoteProposalMetadata.SealedLv2VoteProposalMetadata.Hash()
}

func (sealedLv2DCBVoteProposalMetadata *SealedLv2DCBVoteProposalMetadata) ValidateTxWithBlockChain(tx Transaction, bcr BlockchainRetriever, chainID byte, db database.DatabaseInterface) (bool, error) {
	boardType := sealedLv2DCBVoteProposalMetadata.GetBoardType()
	//Check base seal metadata
	ok, err := sealedLv2DCBVoteProposalMetadata.SealedVoteProposal.ValidateLockerPubKeys(bcr, boardType)
	if err != nil || !ok {
		return ok, err
	}

	//Check precede transaction type
	_, _, _, lv3Tx, _ := bcr.GetTransactionByHash(&sealedLv2DCBVoteProposalMetadata.PointerToLv3VoteProposal)
	if lv3Tx.GetMetadataType() != SealedLv3DCBVoteProposalMeta {
		return false, nil
	}

	// check 2 array equal
	metaLv3 := lv3Tx.GetMetadata().(*SealedLv3DCBVoteProposalMetadata)
	for i := 0; i < len(sealedLv2DCBVoteProposalMetadata.SealedVoteProposal.LockerPubKeys); i++ {
		if !common.ByteEqual(sealedLv2DCBVoteProposalMetadata.SealedVoteProposal.LockerPubKeys[i], metaLv3.SealedVoteProposal.LockerPubKeys[i]) {
			return false, nil
		}
	}

	// Check encrypting
	if !common.ByteEqual(
		sealedLv2DCBVoteProposalMetadata.SealedVoteProposal.SealVoteProposalData,
		common.Encrypt(metaLv3.SealedVoteProposal.SealVoteProposalData, metaLv3.SealedVoteProposal.LockerPubKeys[2]),
	) {
		return false, nil
	}
	return true, nil
}

type SealedLv3DCBVoteProposalMetadata struct {
	SealedLv3VoteProposalMetadata
}

func (sealedLv3DCBVoteProposalMetadata *SealedLv3DCBVoteProposalMetadata) GetBoardType() string {
	return "dcb"
}

func NewSealedLv3DCBVoteProposalMetadata(
	sealedVoteProposal []byte,
	lockerPubKeys [][]byte,
) *SealedLv3DCBVoteProposalMetadata {
	return &SealedLv3DCBVoteProposalMetadata{
		SealedLv3VoteProposalMetadata: *NewSealedLv3VoteProposalMetadata(
			sealedVoteProposal, lockerPubKeys,
			*NewMetadataBase(SealedLv3DCBVoteProposalMeta),
		),
	}
}

func NewSealedLv3DCBVoteProposalMetadataFromJson(data interface{}) *SealedLv3DCBVoteProposalMetadata {
	dataSealedLv3DCBVoteProposal := data.(map[string]interface{})
	return NewSealedLv3DCBVoteProposalMetadata(
		[]byte(dataSealedLv3DCBVoteProposal["SealedVoteProposal"].(string)),
		common.SliceInterfaceToSliceSliceByte(dataSealedLv3DCBVoteProposal["LockerPubKeys"].([]interface{})),
	)
}

type NormalDCBVoteProposalFromSealerMetadata struct {
	NormalVoteProposalFromSealerMetadata
}

func NewNormalDCBVoteProposalFromSealerMetadata(
	voteProposal VoteProposalData,
	lockerPubKey [][]byte,
	pointerToLv1VoteProposal common.Hash,
	pointerToLv3VoteProposal common.Hash,
) *NormalDCBVoteProposalFromSealerMetadata {
	return &NormalDCBVoteProposalFromSealerMetadata{
		NormalVoteProposalFromSealerMetadata: *NewNormalVoteProposalFromSealerMetadata(
			voteProposal,
			lockerPubKey,
			pointerToLv1VoteProposal,
			pointerToLv3VoteProposal,
			*NewMetadataBase(NormalDCBVoteProposalFromSealerMeta),
		),
	}
}

func (normalDCBVoteProposalFromSealerMetadata *NormalDCBVoteProposalFromSealerMetadata) GetBoardType() string {
	return "dcb"
}

func (normalDCBVoteProposalFromSealerMetadata *NormalDCBVoteProposalFromSealerMetadata) ValidateTxWithBlockChain(tx Transaction, bcr BlockchainRetriever, chainID byte, db database.DatabaseInterface) (bool, error) {
	//Validate these pubKeys are in board
	boardType := normalDCBVoteProposalFromSealerMetadata.GetBoardType()
	dcbBoardPubKeys := bcr.GetBoardPubKeys(boardType)
	for _, j := range normalDCBVoteProposalFromSealerMetadata.LockerPubKey {
		exist := false
		for _, i := range dcbBoardPubKeys {
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
	_, _, _, lv1Tx, _ := bcr.GetTransactionByHash(&normalDCBVoteProposalFromSealerMetadata.PointerToLv1VoteProposal)
	if lv1Tx.GetMetadataType() != SealedLv1DCBVoteProposalMeta {
		return false, nil
	}
	_, _, _, lv3Tx, _ := bcr.GetTransactionByHash(&normalDCBVoteProposalFromSealerMetadata.PointerToLv3VoteProposal)
	if lv3Tx.GetMetadataType() != SealedLv3DCBVoteProposalMeta {
		return false, nil
	}

	// check 2 array equal
	metaLv1 := lv1Tx.GetMetadata().(*SealedLv1DCBVoteProposalMetadata)
	for i := 0; i < len(normalDCBVoteProposalFromSealerMetadata.LockerPubKey); i++ {
		if !common.ByteEqual(normalDCBVoteProposalFromSealerMetadata.LockerPubKey[i], metaLv1.SealedVoteProposal.LockerPubKeys[i]) {
			return false, nil
		}
	}

	// Check encrypting
	if !common.ByteEqual(normalDCBVoteProposalFromSealerMetadata.VoteProposal.ToBytes(), common.Encrypt(metaLv1.SealedVoteProposal.SealVoteProposalData, metaLv1.SealedVoteProposal.LockerPubKeys[0])) {
		return false, nil
	}
	return true, nil
}

type NormalDCBVoteProposalFromOwnerMetadata struct {
	NormalVoteProposalFromOwnerMetadata
}

func NewNormalDCBVoteProposalFromOwnerMetadata(
	voteProposal VoteProposalData,
	lockerPubKey [][]byte,
	pointerToLv3VoteProposal common.Hash,
) *NormalDCBVoteProposalFromOwnerMetadata {
	return &NormalDCBVoteProposalFromOwnerMetadata{
		NormalVoteProposalFromOwnerMetadata: *NewNormalVoteProposalFromOwnerMetadata(
			voteProposal,
			lockerPubKey,
			pointerToLv3VoteProposal,
			*NewMetadataBase(NormalDCBVoteProposalFromOwnerMeta),
		),
	}
}

func (normalDCBVoteProposalFromOwnerMetadata *NormalDCBVoteProposalFromOwnerMetadata) Hash() *common.Hash {
	return normalDCBVoteProposalFromOwnerMetadata.NormalVoteProposalFromOwnerMetadata.Hash()
}

func (normalDCBVoteProposalFromOwnerMetadata *NormalDCBVoteProposalFromOwnerMetadata) ValidateTxWithBlockChain(tx Transaction, bcr BlockchainRetriever, chainID byte, db database.DatabaseInterface) (bool, error) {
	//Validate these pubKeys are in board
	dcbBoardPubKeys := bcr.GetBoardPubKeys("dcb")
	for _, j := range normalDCBVoteProposalFromOwnerMetadata.LockerPubKey {
		exist := false
		for _, i := range dcbBoardPubKeys {
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
	_, _, _, lv3Tx, _ := bcr.GetTransactionByHash(&normalDCBVoteProposalFromOwnerMetadata.PointerToLv3VoteProposal)
	if lv3Tx.GetMetadataType() != SealedLv3DCBVoteProposalMeta {
		return false, nil
	}

	// check 2 array equal
	metaLv3 := lv3Tx.GetMetadata().(*SealedLv3DCBVoteProposalMetadata)
	for i := 0; i < len(normalDCBVoteProposalFromOwnerMetadata.LockerPubKey); i++ {
		if !common.ByteEqual(normalDCBVoteProposalFromOwnerMetadata.LockerPubKey[i], metaLv3.SealedVoteProposal.LockerPubKeys[i]) {
			return false, nil
		}
	}

	// Check encrypting
	if !common.ByteEqual(
		metaLv3.SealedVoteProposal.SealVoteProposalData,
		common.Encrypt(
			common.Encrypt(
				common.Encrypt(
					normalDCBVoteProposalFromOwnerMetadata.VoteProposal.ToBytes(),
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
	record += string(punishDCBDecryptMetadata.MetadataBase.Hash().GetBytes())
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
