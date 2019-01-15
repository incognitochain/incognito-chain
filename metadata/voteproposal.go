package metadata

import (
	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/database"
)

//abstract class
type SealedVoteProposal struct {
	SealVoteProposalData []byte
	LockerPubKeys        [][]byte
}

func NewSealedVoteProposalMetadata(sealedVoteProposal []byte, lockerPubKeys [][]byte) *SealedVoteProposal {
	return &SealedVoteProposal{
		SealVoteProposalData: sealedVoteProposal,
		LockerPubKeys:        lockerPubKeys,
	}
}

func (sealedVoteProposal *SealedVoteProposal) Hash2() *common.Hash {
	record := string(sealedVoteProposal.SealVoteProposalData)
	for _, i := range sealedVoteProposal.LockerPubKeys {
		record += string(i)
	}
	hash := common.DoubleHashH([]byte(record))
	return &hash
}

func (sealedVoteProposal *SealedVoteProposal) ValidateLockerPubKeys(bcr BlockchainRetriever, boardType string) (bool, error) {
	//Validate these pubKeys are in board
	boardPubKeys := bcr.GetBoardPubKeys(boardType)
	for _, j := range sealedVoteProposal.LockerPubKeys {
		exist := false
		for _, i := range boardPubKeys {
			if common.ByteEqual(i, j) {
				exist = true
				break
			}
		}
		if !exist {
			return false, nil
		}
	}
	return true, nil
}

func (sealedVoteProposal *SealedVoteProposal) ValidateSanityData(BlockchainRetriever, Transaction) (bool, bool, error) {
	for _, i := range sealedVoteProposal.LockerPubKeys {
		if len(i) != common.PubKeyLength {
			return true, false, nil
		}
	}
	return true, true, nil
}

func (sealedVoteProposal *SealedVoteProposal) ValidateMetadataByItself() bool {
	for index1 := 0; index1 < len(sealedVoteProposal.LockerPubKeys); index1++ {
		pub1 := sealedVoteProposal.LockerPubKeys[index1]
		for index2 := index1 + 1; index2 < len(sealedVoteProposal.LockerPubKeys); index2++ {
			pub2 := sealedVoteProposal.LockerPubKeys[index2]
			if !common.ByteEqual(pub1, pub2) {
				return false
			}
		}
	}
	return true
}

type SealedLv1VoteProposalMetadata struct {
	SealedVoteProposal
	PointerToLv2VoteProposal common.Hash
	PointerToLv3VoteProposal common.Hash
	MetadataBase
}

func (sealedLv1VoteProposalMetadata *SealedLv1VoteProposalMetadata) GetBoardType() string {
	panic("override me")
}
func (sealedLv1VoteProposalMetadata *SealedLv1VoteProposalMetadata) ValidataBeforeNewBlock(tx Transaction, bcr BlockchainRetriever, chainID byte) bool {
	boardType := sealedLv1VoteProposalMetadata.GetBoardType()
	endedPivot := bcr.GetConstitutionEndHeight(boardType, chainID)
	currentBlockHeight := bcr.GetCurrentBlockHeight(chainID) + 1
	lv3Pivot := endedPivot - common.EncryptionOnePhraseDuration
	lv2Pivot := lv3Pivot - common.EncryptionOnePhraseDuration
	lv1Pivot := lv2Pivot - common.EncryptionOnePhraseDuration
	if !(currentBlockHeight < lv1Pivot && currentBlockHeight >= lv2Pivot) {
		return false
	}
	return true
}

func (sealedLv1VoteProposalMetadata *SealedLv1VoteProposalMetadata) ValidateSanityData(bcr BlockchainRetriever, tx Transaction) (bool, bool, error) {
	_, ok, _ := sealedLv1VoteProposalMetadata.SealedVoteProposal.ValidateSanityData(bcr, tx)
	if !ok {
		return true, false, nil
	}
	return true, true, nil
}

func (sealedLv1VoteProposalMetadata *SealedLv1VoteProposalMetadata) ValidateMetadataByItself() bool {
	return true
}

func NewSealedLv1VoteProposalMetadata(
	sealedVoteProposal []byte,
	lockersPubKey [][]byte,
	pointerToLv2VoteProposal common.Hash,
	pointerToLv3VoteProposal common.Hash,
	metadataBase MetadataBase,
) *SealedLv1VoteProposalMetadata {
	return &SealedLv1VoteProposalMetadata{
		SealedVoteProposal:       *NewSealedVoteProposalMetadata(sealedVoteProposal, lockersPubKey),
		PointerToLv2VoteProposal: pointerToLv2VoteProposal,
		PointerToLv3VoteProposal: pointerToLv3VoteProposal,
		MetadataBase:             metadataBase,
	}
}

func (sealedLv1VoteProposalMetadata *SealedLv1VoteProposalMetadata) Hash2() *common.Hash {
	record := string(sealedLv1VoteProposalMetadata.SealedVoteProposal.Hash2().GetBytes())
	record += string(sealedLv1VoteProposalMetadata.PointerToLv2VoteProposal.GetBytes())
	record += string(sealedLv1VoteProposalMetadata.PointerToLv3VoteProposal.GetBytes())
	record += string(sealedLv1VoteProposalMetadata.MetadataBase.Hash().GetBytes())
	hash := common.DoubleHashH([]byte(record))
	return &hash
}

type SealedLv2VoteProposalMetadata struct {
	SealedVoteProposal
	PointerToLv3VoteProposal common.Hash
	MetadataBase
}

func (sealedLv2VoteProposalMetadata *SealedLv2VoteProposalMetadata) Hash2() *common.Hash {
	record := string(sealedLv2VoteProposalMetadata.SealedVoteProposal.Hash2().GetBytes())
	record += string(sealedLv2VoteProposalMetadata.PointerToLv3VoteProposal.GetBytes())
	record += string(sealedLv2VoteProposalMetadata.MetadataBase.Hash().GetBytes())
	hash := common.DoubleHashH([]byte(record))
	return &hash
}

func (sealedLv2VoteProposalMetadata *SealedLv2VoteProposalMetadata) GetBoardType() string {
	panic("overwrite me")
}
func (sealedLv2VoteProposalMetadata *SealedLv2VoteProposalMetadata) ValidataBeforeNewBlock(tx Transaction, bcr BlockchainRetriever, chainID byte) bool {
	boardType := sealedLv2VoteProposalMetadata.GetBoardType()
	endedPivot := bcr.GetConstitutionEndHeight(boardType, chainID)
	currentBlockHeight := bcr.GetCurrentBlockHeight(chainID) + 1
	lv3Pivot := endedPivot - common.EncryptionOnePhraseDuration
	lv2Pivot := lv3Pivot - common.EncryptionOnePhraseDuration
	if !(currentBlockHeight < lv2Pivot && currentBlockHeight >= lv3Pivot) {
		return false
	}
	return true
}

func (sealedLv2VoteProposalMetadata *SealedLv2VoteProposalMetadata) ValidateSanityData(bcr BlockchainRetriever, tx Transaction) (bool, bool, error) {
	_, ok, _ := sealedLv2VoteProposalMetadata.SealedVoteProposal.ValidateSanityData(bcr, tx)
	if !ok {
		return true, false, nil
	}
	return true, true, nil
}

func (sealedLv2VoteProposalMetadata *SealedLv2VoteProposalMetadata) ValidateMetadataByItself() bool {
	return true
}

func NewSealedLv2VoteProposalMetadata(sealedVoteProposal []byte, lockerPubKeys [][]byte, pointerToLv3VoteProposal common.Hash, metadataBase MetadataBase) *SealedLv2VoteProposalMetadata {
	return &SealedLv2VoteProposalMetadata{
		SealedVoteProposal: *NewSealedVoteProposalMetadata(
			sealedVoteProposal,
			lockerPubKeys,
		),
		PointerToLv3VoteProposal: pointerToLv3VoteProposal,
		MetadataBase:             metadataBase,
	}

}

type SealedLv3VoteProposalMetadata struct {
	SealedVoteProposal
	MetadataBase
}

func (sealedLv3VoteProposalMetadata *SealedLv3VoteProposalMetadata) GetBoardType() string {
	panic("overwrite me")
}
func (sealedLv3VoteProposalMetadata *SealedLv3VoteProposalMetadata) ValidataBeforeNewBlock(tx Transaction, bcr BlockchainRetriever, chainID byte) bool {
	boardType := sealedLv3VoteProposalMetadata.GetBoardType()
	startedPivot := bcr.GetConstitutionStartHeight(boardType, chainID)
	endedPivot := bcr.GetConstitutionEndHeight(boardType, chainID)
	currentBlockHeight := bcr.GetCurrentBlockHeight(chainID) + 1
	lv3Pivot := endedPivot - common.EncryptionOnePhraseDuration
	if !(currentBlockHeight < lv3Pivot && currentBlockHeight >= startedPivot) {
		return false
	}
	return true
}

func (sealLv3VoteProposalMetadata *SealedLv3VoteProposalMetadata) ValidateTxWithBlockChain(tx Transaction, bcr BlockchainRetriever, b byte, db database.DatabaseInterface) (bool, error) {
	return sealLv3VoteProposalMetadata.ValidateTxWithBlockChain(tx, bcr, b, db)
}

func (sealLv3VoteProposalMetadata *SealedLv3VoteProposalMetadata) ValidateSanityData(bcr BlockchainRetriever, tx Transaction) (bool, bool, error) {
	return sealLv3VoteProposalMetadata.ValidateSanityData(bcr, tx)
}

func (sealLv3VoteProposalMetadata *SealedLv3VoteProposalMetadata) ValidateMetadataByItself() bool {
	return sealLv3VoteProposalMetadata.ValidateMetadataByItself()
}

func NewSealedLv3VoteProposalMetadata(
	sealedVoteProposal []byte,
	lockerPubKeys [][]byte,
	metadataBase MetadataBase,
) *SealedLv3VoteProposalMetadata {
	return &SealedLv3VoteProposalMetadata{
		SealedVoteProposal: *NewSealedVoteProposalMetadata(sealedVoteProposal, lockerPubKeys),
		MetadataBase:       metadataBase,
	}

}

type VoteProposalData struct {
	ProposalTxID common.Hash
	AmountOfVote int32
}

func NewVoteProposalData(proposalTxID common.Hash, amountOfVote int32) *VoteProposalData {
	return &VoteProposalData{ProposalTxID: proposalTxID, AmountOfVote: amountOfVote}
}

func NewVoteProposalDataFromJson(data interface{}) *VoteProposalData {
	voteProposalDataData := data.(map[string]interface{})
	return NewVoteProposalData(
		common.NewHash([]byte(voteProposalDataData["ProposalTxID"].(string))),
		int32(voteProposalDataData["AmountOfVote"].(float64)),
	)
}

func (voteProposalData VoteProposalData) ToBytes() []byte {
	b := voteProposalData.ProposalTxID.GetBytes()
	b = append(b, common.Int32ToBytes(voteProposalData.AmountOfVote)...)
	return b
}

func NewVoteProposalDataFromBytes(b []byte) *VoteProposalData {
	lenB := len(b)
	return NewVoteProposalData(
		common.NewHash(b[:lenB-4]),
		common.BytesToInt32(b[lenB-4:]),
	)
}

func (voteProposalData VoteProposalData) Hash2() *common.Hash {
	record := string(voteProposalData.ProposalTxID.GetBytes())
	record += string(voteProposalData.AmountOfVote)

	hash := common.DoubleHashH([]byte(record))
	return &hash
}

type NormalVoteProposalFromSealerMetadata struct {
	VoteProposal             VoteProposalData
	LockerPubKey             [][]byte
	PointerToLv1VoteProposal common.Hash
	PointerToLv3VoteProposal common.Hash
	MetadataBase
}

func NewNormalVoteProposalFromSealerMetadata(
	voteProposal VoteProposalData,
	lockerPubKey [][]byte,
	pointerToLv1VoteProposal common.Hash,
	pointerToLv3VoteProposal common.Hash,
	metadataBase MetadataBase,
) *NormalVoteProposalFromSealerMetadata {
	return &NormalVoteProposalFromSealerMetadata{
		VoteProposal:             voteProposal,
		LockerPubKey:             lockerPubKey,
		PointerToLv1VoteProposal: pointerToLv1VoteProposal,
		PointerToLv3VoteProposal: pointerToLv3VoteProposal,
		MetadataBase:             metadataBase,
	}
}
func (normalVoteProposalFromSealerMetadata *NormalVoteProposalFromSealerMetadata) GetBoardType() string {
	panic("overwrite me")
}
func (normalVoteProposalFromSealerMetadata *NormalVoteProposalFromSealerMetadata) ValidateSanityData(BlockchainRetriever, Transaction) (bool, bool, error) {
	for _, i := range normalVoteProposalFromSealerMetadata.LockerPubKey {
		if len(i) != common.PubKeyLength {
			return true, false, nil
		}
	}
	return true, true, nil
}

func (normalVoteProposalFromSealerMetadata *NormalVoteProposalFromSealerMetadata) ValidateMetadataByItself() bool {
	for index1 := 0; index1 < len(normalVoteProposalFromSealerMetadata.LockerPubKey); index1++ {
		pub1 := normalVoteProposalFromSealerMetadata.LockerPubKey[index1]
		for index2 := index1 + 1; index2 < len(normalVoteProposalFromSealerMetadata.LockerPubKey); index2++ {
			pub2 := normalVoteProposalFromSealerMetadata.LockerPubKey[index2]
			if !common.ByteEqual(pub1, pub2) {
				return false
			}
		}
	}
	return true
}

func (normalVoteProposalFromSealerMetadata *NormalVoteProposalFromSealerMetadata) Hash2() *common.Hash {
	record := string(normalVoteProposalFromSealerMetadata.VoteProposal.Hash2().GetBytes())
	for _, i := range normalVoteProposalFromSealerMetadata.LockerPubKey {
		record += string(i)
	}
	record += string(normalVoteProposalFromSealerMetadata.PointerToLv1VoteProposal.GetBytes())
	record += string(normalVoteProposalFromSealerMetadata.PointerToLv3VoteProposal.GetBytes())
	record += string(normalVoteProposalFromSealerMetadata.MetadataBase.Hash().GetBytes())
	hash := common.DoubleHashH([]byte(record))
	return &hash
}

func (normalVoteProposalFromSealerMetadata *NormalVoteProposalFromSealerMetadata) ValidataBeforeNewBlock(tx Transaction, bcr BlockchainRetriever, chainID byte) bool {
	boardType := normalVoteProposalFromSealerMetadata.GetBoardType()
	endedPivot := bcr.GetConstitutionEndHeight(boardType, chainID)
	currentBlockHeight := bcr.GetCurrentBlockHeight(chainID) + 1
	lv3Pivot := endedPivot - common.EncryptionOnePhraseDuration
	lv2Pivot := lv3Pivot - common.EncryptionOnePhraseDuration
	lv1Pivot := lv2Pivot - common.EncryptionOnePhraseDuration
	if !(currentBlockHeight < endedPivot && currentBlockHeight >= lv1Pivot) {
		return false
	}
	return true
}

type NormalVoteProposalFromOwnerMetadata struct {
	VoteProposal             VoteProposalData
	LockerPubKey             [][]byte
	PointerToLv3VoteProposal common.Hash
	MetadataBase
}

func NewNormalVoteProposalFromOwnerMetadata(
	voteProposal VoteProposalData,
	lockerPubKey [][]byte,
	pointerToLv3VoteProposal common.Hash,
	metadataBase MetadataBase,
) *NormalVoteProposalFromOwnerMetadata {
	return &NormalVoteProposalFromOwnerMetadata{
		VoteProposal:             voteProposal,
		LockerPubKey:             lockerPubKey,
		PointerToLv3VoteProposal: pointerToLv3VoteProposal,
		MetadataBase:             metadataBase,
	}
}

func (normalVoteProposalFromOwnerMetadata *NormalVoteProposalFromOwnerMetadata) ValidataBeforeNewBlock(tx Transaction, bcr BlockchainRetriever, chainID byte) bool {
	endedPivot := bcr.GetConstitutionEndHeight("dcb", chainID)
	currentBlockHeight := bcr.GetCurrentBlockHeight(chainID) + 1
	lv3Pivot := endedPivot - common.EncryptionOnePhraseDuration
	lv2Pivot := lv3Pivot - common.EncryptionOnePhraseDuration
	lv1Pivot := lv2Pivot - common.EncryptionOnePhraseDuration
	if !(currentBlockHeight < endedPivot && currentBlockHeight >= lv1Pivot) {
		return false
	}
	return true
}

func (normalVoteProposalFromOwnerMetadata *NormalVoteProposalFromOwnerMetadata) Hash2() *common.Hash {
	record := string(normalVoteProposalFromOwnerMetadata.VoteProposal.Hash2().GetBytes())
	for _, i := range normalVoteProposalFromOwnerMetadata.LockerPubKey {
		record += string(i)
	}
	record += string(normalVoteProposalFromOwnerMetadata.PointerToLv3VoteProposal.GetBytes())
	record += string(normalVoteProposalFromOwnerMetadata.MetadataBase.Hash().GetBytes())
	hash := common.DoubleHashH([]byte(record))
	return &hash
}

func (normalVoteProposalFromOwnerMetadata *NormalVoteProposalFromOwnerMetadata) ValidateSanityData(BlockchainRetriever, Transaction) (bool, bool, error) {
	for _, i := range normalVoteProposalFromOwnerMetadata.LockerPubKey {
		if len(i) != common.PubKeyLength {
			return true, false, nil
		}
	}
	return true, true, nil
}

func (normalVoteProposalFromOwnerMetadata *NormalVoteProposalFromOwnerMetadata) ValidateMetadataByItself() bool {
	for index1 := 0; index1 < len(normalVoteProposalFromOwnerMetadata.LockerPubKey); index1++ {
		pub1 := normalVoteProposalFromOwnerMetadata.LockerPubKey[index1]
		for index2 := index1 + 1; index2 < len(normalVoteProposalFromOwnerMetadata.LockerPubKey); index2++ {
			pub2 := normalVoteProposalFromOwnerMetadata.LockerPubKey[index2]
			if !common.ByteEqual(pub1, pub2) {
				return false
			}
		}
	}
	return true
}
