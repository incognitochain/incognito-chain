package metadata

import (
	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/database"
)

//abstract class
type SealedVoteProposal struct {
	VoteProposalData []byte
	LockerPubKeys    [][]byte
}

func NewSealedVoteProposalMetadata(sealedVoteProposal []byte, lockerPubKeys [][]byte) *SealedVoteProposal {
	return &SealedVoteProposal{
		VoteProposalData: sealedVoteProposal,
		LockerPubKeys:    lockerPubKeys,
	}
}

func (sealedVoteProposal *SealedVoteProposal) Hash() *common.Hash {
	record := string(sealedVoteProposal.VoteProposalData)
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
	sealedVoteProposal       SealedVoteProposal
	PointerToLv2VoteProposal common.Hash
	PointerToLv3VoteProposal common.Hash
	MetadataBase
}

type SealedLv1DCBVoteProposalMetadata struct {
	SealedLv1VoteProposalMetadata
}
type SealedLv1GOVVoteProposalMetadata struct {
	SealedLv1VoteProposalMetadata
}

func (sealedLv1VoteProposalMetadata *SealedLv1VoteProposalMetadata) GetBoardType() string {
	panic("override me")
}
func (sealedLv1DCBVoteProposalMetadata *SealedLv1DCBVoteProposalMetadata) GetBoardType() string {
	return "dcb"
}
func (sealedLv1GOVVoteProposalMetadata *SealedLv1GOVVoteProposalMetadata) GetBoardType() string {
	return "gov"
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
	_, ok, _ := sealedLv1VoteProposalMetadata.sealedVoteProposal.ValidateSanityData(bcr, tx)
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
		sealedVoteProposal:       *NewSealedVoteProposalMetadata(sealedVoteProposal, lockersPubKey),
		PointerToLv2VoteProposal: pointerToLv2VoteProposal,
		PointerToLv3VoteProposal: pointerToLv3VoteProposal,
		MetadataBase:             metadataBase,
	}
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

func (sealedLv1VoteProposalMetadata *SealedLv1VoteProposalMetadata) Hash() *common.Hash {
	record := string(sealedLv1VoteProposalMetadata.sealedVoteProposal.Hash().GetBytes())
	record += string(sealedLv1VoteProposalMetadata.PointerToLv2VoteProposal.GetBytes())
	record += string(sealedLv1VoteProposalMetadata.PointerToLv3VoteProposal.GetBytes())
	record += string(sealedLv1VoteProposalMetadata.MetadataBase.Hash().GetBytes())
	hash := common.DoubleHashH([]byte(record))
	return &hash
}

func (sealedLv1DCBVoteProposalMetadata *SealedLv1DCBVoteProposalMetadata) Hash() *common.Hash {
	return sealedLv1DCBVoteProposalMetadata.SealedLv1VoteProposalMetadata.Hash()
}
func (sealedLv1GOVVoteProposalMetadata *SealedLv1GOVVoteProposalMetadata) Hash() *common.Hash {
	return sealedLv1GOVVoteProposalMetadata.SealedLv1VoteProposalMetadata.Hash()
}

func (sealedLv1DCBVoteProposalMetadata *SealedLv1DCBVoteProposalMetadata) ValidateTxWithBlockChain(
	tx Transaction,
	bcr BlockchainRetriever,
	chainID byte,
	db database.DatabaseInterface,
) (bool, error) {
	//Check base seal metadata
	ok, err := sealedLv1DCBVoteProposalMetadata.sealedVoteProposal.ValidateLockerPubKeys(bcr, sealedLv1DCBVoteProposalMetadata.GetBoardType())
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
	for i := 0; i < len(sealedLv1DCBVoteProposalMetadata.sealedVoteProposal.LockerPubKeys); i++ {
		if !common.ByteEqual(sealedLv1DCBVoteProposalMetadata.sealedVoteProposal.LockerPubKeys[i], metaLv2.sealedVoteProposal.LockerPubKeys[i]) {
			return false, nil
		}
	}

	// Check encrypting
	if !common.ByteEqual(sealedLv1DCBVoteProposalMetadata.sealedVoteProposal.VoteProposalData,
		common.Encrypt(metaLv2.sealedVoteProposal.VoteProposalData, metaLv2.sealedVoteProposal.LockerPubKeys[1])) {
		return false, nil
	}
	return true, nil
}

func (sealedLv1GOVVoteProposalMetadata *SealedLv1GOVVoteProposalMetadata) ValidateTxWithBlockChain(
	tx Transaction,
	bcr BlockchainRetriever,
	chainID byte,
	db database.DatabaseInterface,
) (bool, error) {
	//Check base seal metadata
	ok, err := sealedLv1GOVVoteProposalMetadata.sealedVoteProposal.ValidateLockerPubKeys(bcr, sealedLv1GOVVoteProposalMetadata.GetBoardType())
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
	for i := 0; i < len(sealedLv1GOVVoteProposalMetadata.sealedVoteProposal.LockerPubKeys); i++ {
		if !common.ByteEqual(sealedLv1GOVVoteProposalMetadata.sealedVoteProposal.LockerPubKeys[i], metaLv2.sealedVoteProposal.LockerPubKeys[i]) {
			return false, nil
		}
	}

	// Check encrypting
	if !common.ByteEqual(sealedLv1GOVVoteProposalMetadata.sealedVoteProposal.VoteProposalData,
		common.Encrypt(metaLv2.sealedVoteProposal.VoteProposalData, metaLv2.sealedVoteProposal.LockerPubKeys[1])) {
		return false, nil
	}
	return true, nil
}

type SealedLv2VoteProposalMetadata struct {
	sealedVoteProposal       SealedVoteProposal
	PointerToLv3VoteProposal common.Hash
	MetadataBase
}

type SealedLv2DCBVoteProposalMetadata struct {
	SealedLv2VoteProposalMetadata
}
type SealedLv2GOVVoteProposalMetadata struct {
	SealedLv2VoteProposalMetadata
}

func (sealedLv2VoteProposalMetadata *SealedLv2VoteProposalMetadata) GetBoardType() string {
	panic("overwrite me")
}
func (sealedLv2DCBVoteProposalMetadata *SealedLv2DCBVoteProposalMetadata) GetBoardType() string {
	return "dcb"
}
func (sealedLv2GOVVoteProposalMetadata *SealedLv2GOVVoteProposalMetadata) GetBoardType() string {
	return "gov"
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
	_, ok, _ := sealedLv2VoteProposalMetadata.sealedVoteProposal.ValidateSanityData(bcr, tx)
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
		sealedVoteProposal: *NewSealedVoteProposalMetadata(
			sealedVoteProposal,
			lockerPubKeys,
		),
		PointerToLv3VoteProposal: pointerToLv3VoteProposal,
		MetadataBase:             metadataBase,
	}

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

func NewSealedLv2GOVVoteProposalMetadata(sealedVoteProposal []byte, lockerPubKeys [][]byte, pointerToLv3VoteProposal common.Hash) *SealedLv2GOVVoteProposalMetadata {
	return &SealedLv2GOVVoteProposalMetadata{
		SealedLv2VoteProposalMetadata: *NewSealedLv2VoteProposalMetadata(
			sealedVoteProposal,
			lockerPubKeys,
			pointerToLv3VoteProposal,
			*NewMetadataBase(SealedLv2GOVVoteProposalMeta),
		),
	}
}

func (sealedLv2VoteProposalMetadata *SealedLv2VoteProposalMetadata) Hash() *common.Hash {
	record := string(sealedLv2VoteProposalMetadata.sealedVoteProposal.Hash().GetBytes())
	record += string(sealedLv2VoteProposalMetadata.PointerToLv3VoteProposal.GetBytes())
	record += string(sealedLv2VoteProposalMetadata.MetadataBase.Hash().GetBytes())
	hash := common.DoubleHashH([]byte(record))
	return &hash
}

func (sealedLv2DCBVoteProposalMetadata *SealedLv2DCBVoteProposalMetadata) Hash() *common.Hash {
	return sealedLv2DCBVoteProposalMetadata.SealedLv2VoteProposalMetadata.Hash()
}

func (sealedLv2GOVVoteProposalMetadata *SealedLv2GOVVoteProposalMetadata) Hash() *common.Hash {
	return sealedLv2GOVVoteProposalMetadata.SealedLv2VoteProposalMetadata.Hash()
}

func (sealedLv2DCBVoteProposalMetadata *SealedLv2DCBVoteProposalMetadata) ValidateTxWithBlockChain(tx Transaction, bcr BlockchainRetriever, chainID byte, db database.DatabaseInterface) (bool, error) {
	boardType := sealedLv2DCBVoteProposalMetadata.GetBoardType()
	//Check base seal metadata
	ok, err := sealedLv2DCBVoteProposalMetadata.sealedVoteProposal.ValidateLockerPubKeys(bcr, boardType)
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
	for i := 0; i < len(sealedLv2DCBVoteProposalMetadata.sealedVoteProposal.LockerPubKeys); i++ {
		if !common.ByteEqual(sealedLv2DCBVoteProposalMetadata.sealedVoteProposal.LockerPubKeys[i], metaLv3.SealedVoteProposal.LockerPubKeys[i]) {
			return false, nil
		}
	}

	// Check encrypting
	if !common.ByteEqual(
		sealedLv2DCBVoteProposalMetadata.sealedVoteProposal.VoteProposalData,
		common.Encrypt(metaLv3.SealedVoteProposal.VoteProposalData, metaLv3.SealedVoteProposal.LockerPubKeys[2]),
	) {
		return false, nil
	}
	return true, nil
}

func (sealedLv2GOVVoteProposalMetadata *SealedLv2GOVVoteProposalMetadata) ValidateTxWithBlockChain(tx Transaction, bcr BlockchainRetriever, chainID byte, db database.DatabaseInterface) (bool, error) {
	boardType := sealedLv2GOVVoteProposalMetadata.GetBoardType()
	//Check base seal metadata
	ok, err := sealedLv2GOVVoteProposalMetadata.sealedVoteProposal.ValidateLockerPubKeys(bcr, boardType)
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
	for i := 0; i < len(sealedLv2GOVVoteProposalMetadata.sealedVoteProposal.LockerPubKeys); i++ {
		if !common.ByteEqual(sealedLv2GOVVoteProposalMetadata.sealedVoteProposal.LockerPubKeys[i], metaLv3.SealedVoteProposal.LockerPubKeys[i]) {
			return false, nil
		}
	}

	// Check encrypting
	if !common.ByteEqual(
		sealedLv2GOVVoteProposalMetadata.sealedVoteProposal.VoteProposalData,
		common.Encrypt(metaLv3.SealedVoteProposal.VoteProposalData, metaLv3.SealedVoteProposal.LockerPubKeys[2]),
	) {
		return false, nil
	}
	return true, nil
}

type SealedLv3VoteProposalMetadata struct {
	SealedVoteProposal SealedVoteProposal
	MetadataBase
}

type SealedLv3DCBVoteProposalMetadata struct {
	SealedLv3VoteProposalMetadata
}
type SealedLv3GOVVoteProposalMetadata struct {
	SealedLv3VoteProposalMetadata
}

func (sealedLv3VoteProposalMetadata *SealedLv3VoteProposalMetadata) GetBoardType() string {
	panic("overwrite me")
}
func (sealedLv3DCBVoteProposalMetadata *SealedLv3DCBVoteProposalMetadata) GetBoardType() string {
	return "dcb"
}
func (sealedLv3GOVVoteProposalMetadata *SealedLv3GOVVoteProposalMetadata) GetBoardType() string {
	return "gov"
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

func NewSealedLv3DCBVoteProposalMetadataFromJson(data interface{}) *SealedLv3DCBVoteProposalMetadata {
	dataSealedLv3DCBVoteProposal := data.(map[string]interface{})
	return NewSealedLv3DCBVoteProposalMetadata(
		[]byte(dataSealedLv3DCBVoteProposal["SealedVoteProposal"].(string)),
		common.SliceInterfaceToSliceSliceByte(dataSealedLv3DCBVoteProposal["LockerPubKeys"].([]interface{})),
	)
}
func NewSealedLv3GOVVoteProposalMetadataFromJson(data interface{}) *SealedLv3GOVVoteProposalMetadata {
	dataSealedLv3GOVVoteProposal := data.(map[string]interface{})
	return NewSealedLv3GOVVoteProposalMetadata(
		[]byte(dataSealedLv3GOVVoteProposal["SealedVoteProposal"].(string)),
		common.SliceInterfaceToSliceSliceByte(dataSealedLv3GOVVoteProposal["LockerPubKeys"].([]interface{})),
	)
}

type NormalVoteProposalFromSealerMetadata struct {
	VoteProposal             []byte
	LockerPubKey             [][]byte
	PointerToLv1VoteProposal common.Hash
	PointerToLv3VoteProposal common.Hash
	MetadataBase
}

type NormalDCBVoteProposalFromSealerMetadata struct {
	NormalVoteProposalFromSealerMetadata
}
type NormalGOVVoteProposalFromSealerMetadata struct {
	NormalVoteProposalFromSealerMetadata
}

func NewNormalVoteProposalFromSealerMetadata(
	voteProposal []byte,
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
func NewNormalDCBVoteProposalFromSealerMetadata(
	voteProposal []byte,
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
func NewNormalGOVVoteProposalFromSealerMetadata(
	voteProposal []byte,
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

func (normalVoteProposalFromSealerMetadata *NormalVoteProposalFromSealerMetadata) GetBoardType() string {
	panic("overwrite me")
}
func (normalDCBVoteProposalFromSealerMetadata *NormalDCBVoteProposalFromSealerMetadata) GetBoardType() string {
	return "dcb"
}
func (normalGOVVoteProposalFromSealerMetadata *NormalGOVVoteProposalFromSealerMetadata) GetBoardType() string {
	return "gov"
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

func (normalVoteProposalFromSealerMetadata *NormalVoteProposalFromSealerMetadata) Hash() *common.Hash {
	record := string(normalVoteProposalFromSealerMetadata.VoteProposal)
	for _, i := range normalVoteProposalFromSealerMetadata.LockerPubKey {
		record += string(i)
	}
	record += string(normalVoteProposalFromSealerMetadata.PointerToLv1VoteProposal.GetBytes())
	record += string(normalVoteProposalFromSealerMetadata.PointerToLv3VoteProposal.GetBytes())
	record += string(normalVoteProposalFromSealerMetadata.MetadataBase.Hash().GetBytes())
	hash := common.DoubleHashH([]byte(record))
	return &hash
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
		if !common.ByteEqual(normalDCBVoteProposalFromSealerMetadata.LockerPubKey[i], metaLv1.sealedVoteProposal.LockerPubKeys[i]) {
			return false, nil
		}
	}

	// Check encrypting
	if !common.ByteEqual(normalDCBVoteProposalFromSealerMetadata.VoteProposal, common.Encrypt(metaLv1.sealedVoteProposal.VoteProposalData, metaLv1.sealedVoteProposal.LockerPubKeys[0])) {
		return false, nil
	}
	return true, nil
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
		if !common.ByteEqual(normalGOVVoteProposalFromSealerMetadata.LockerPubKey[i], metaLv1.sealedVoteProposal.LockerPubKeys[i]) {
			return false, nil
		}
	}

	// Check encrypting
	if !common.ByteEqual(normalGOVVoteProposalFromSealerMetadata.VoteProposal, common.Encrypt(metaLv1.sealedVoteProposal.VoteProposalData, metaLv1.sealedVoteProposal.LockerPubKeys[0])) {
		return false, nil
	}
	return true, nil
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
	VoteProposal             []byte
	LockerPubKey             [][]byte
	PointerToLv3VoteProposal common.Hash
	MetadataBase
}

type NormalDCBVoteProposalFromOwnerMetadata struct {
	NormalVoteProposalFromOwnerMetadata
}
type NormalGOVVoteProposalFromOwnerMetadata struct {
	NormalVoteProposalFromOwnerMetadata
}

func NewNormalVoteProposalFromOwnerMetadata(
	voteProposal []byte,
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

func NewNormalDCBVoteProposalFromOwnerMetadata(
	voteProposal []byte,
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

func NewNormalGOVVoteProposalFromOwnerMetadata(
	voteProposal []byte,
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

func (normalVoteProposalFromOwnerMetadata *NormalVoteProposalFromOwnerMetadata) Hash() *common.Hash {
	record := string(normalVoteProposalFromOwnerMetadata.VoteProposal)
	for _, i := range normalVoteProposalFromOwnerMetadata.LockerPubKey {
		record += string(i)
	}
	record += string(normalVoteProposalFromOwnerMetadata.PointerToLv3VoteProposal.GetBytes())
	record += string(normalVoteProposalFromOwnerMetadata.MetadataBase.Hash().GetBytes())
	hash := common.DoubleHashH([]byte(record))
	return &hash
}

func (normalDCBVoteProposalFromOwnerMetadata *NormalDCBVoteProposalFromOwnerMetadata) Hash() *common.Hash {
	return normalDCBVoteProposalFromOwnerMetadata.NormalVoteProposalFromOwnerMetadata.Hash()
}
func (normalGOVVoteProposalFromOwnerMetadata *NormalGOVVoteProposalFromOwnerMetadata) Hash() *common.Hash {
	return normalGOVVoteProposalFromOwnerMetadata.NormalVoteProposalFromOwnerMetadata.Hash()
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
		metaLv3.SealedVoteProposal.VoteProposalData,
		common.Encrypt(
			common.Encrypt(
				common.Encrypt(
					normalDCBVoteProposalFromOwnerMetadata.VoteProposal,
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
		metaLv3.SealedVoteProposal.VoteProposalData,
		common.Encrypt(
			common.Encrypt(
				common.Encrypt(
					normalGOVVoteProposalFromOwnerMetadata.VoteProposal,
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

type PunishDCBDecryptMetadata struct {
	pubKey []byte
	MetadataBase
}

func NewPunishDCBDecryptMetadata(pubKey []byte) *PunishDCBDecryptMetadata {
	return &PunishDCBDecryptMetadata{
		pubKey:       pubKey,
		MetadataBase: *NewMetadataBase(PunishDCBDecryptMeta),
	}
}

func (punishDCBDecryptMetadata *PunishDCBDecryptMetadata) Hash() *common.Hash {
	record := string(punishDCBDecryptMetadata.pubKey)
	record += string(punishDCBDecryptMetadata.MetadataBase.Hash()[:])
	hash := common.DoubleHashH([]byte(record))
	return &hash
}

//todo @0xjackalope validate within block chain and current block

func (punishDCBDecryptMetadata *PunishDCBDecryptMetadata) ValidateTxWithBlockChain(Transaction, BlockchainRetriever, byte, database.DatabaseInterface) (bool, error) {
	return true, nil
}

func (punishDCBDecryptMetadata *PunishDCBDecryptMetadata) ValidateSanityData(BlockchainRetriever, Transaction) (bool, bool, error) {
	if len(punishDCBDecryptMetadata.pubKey) != common.PubKeyLength {
		return true, false, nil
	}
	return true, true, nil
}

func (punishDCBDecryptMetadata *PunishDCBDecryptMetadata) ValidateMetadataByItself() bool {
	return true
}

type PunishGOVDecryptMetadata struct {
	pubKey []byte
	MetadataBase
}

func NewPunishGOVDecryptMetadata(pubKey []byte) *PunishGOVDecryptMetadata {
	return &PunishGOVDecryptMetadata{
		pubKey:       pubKey,
		MetadataBase: *NewMetadataBase(PunishGOVDecryptMeta),
	}
}

func (punishGOVDecryptMetadata *PunishGOVDecryptMetadata) Hash() *common.Hash {
	record := string(punishGOVDecryptMetadata.pubKey)
	record += string(punishGOVDecryptMetadata.MetadataBase.Hash()[:])
	hash := common.DoubleHashH([]byte(record))
	return &hash
}

//todo @0xjackalope validate within block chain and current block

func (punishGOVDecryptMetadata *PunishGOVDecryptMetadata) ValidateTxWithBlockChain(Transaction, BlockchainRetriever, byte, database.DatabaseInterface) (bool, error) {
	return true, nil
}

func (punishGOVDecryptMetadata *PunishGOVDecryptMetadata) ValidateSanityData(BlockchainRetriever, Transaction) (bool, bool, error) {
	if len(punishGOVDecryptMetadata.pubKey) != common.PubKeyLength {
		return true, false, nil
	}
	return true, true, nil
}

func (punishGOVDecryptMetadata *PunishGOVDecryptMetadata) ValidateMetadataByItself() bool {
	return true
}
