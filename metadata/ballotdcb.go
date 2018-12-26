package metadata

import (
	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/database"
)

//abstract class
type SealedDCBBallot struct {
	BallotData    []byte
	LockerPubKeys [][]byte
}

func NewSealedDCBBallotMetadata(sealedBallot []byte, lockerPubKeys [][]byte) *SealedDCBBallot {
	return &SealedDCBBallot{
		BallotData:    sealedBallot,
		LockerPubKeys: lockerPubKeys,
	}
}

func (sealDCBBallotMetadata *SealedDCBBallot) Hash() *common.Hash {
	record := string(sealDCBBallotMetadata.BallotData)
	for _, i := range sealDCBBallotMetadata.LockerPubKeys {
		record += string(i)
	}
	hash := common.DoubleHashH([]byte(record))
	return &hash
}

func (sealDCBBallotMetadata *SealedDCBBallot) ValidateTxWithBlockChain(tx Transaction, bcr BlockchainRetriever, chainID byte, db database.DatabaseInterface) (bool, error) {
	//Validate these pubKeys are in board
	dcbBoardPubKeys := bcr.GetDCBBoardPubKeys()
	for _, j := range sealDCBBallotMetadata.LockerPubKeys {
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
	return true, nil
}

func (sealDCBBallotMetadata *SealedDCBBallot) ValidateSanityData(BlockchainRetriever, Transaction) (bool, bool, error) {
	for _, i := range sealDCBBallotMetadata.LockerPubKeys {
		if len(i) != common.PubKeyLength {
			return true, false, nil
		}
	}
	return true, true, nil
}

func (sealDCBBallotMetadata *SealedDCBBallot) ValidateMetadataByItself() bool {
	for index1 := 0; index1 < len(sealDCBBallotMetadata.LockerPubKeys); index1++ {
		pub1 := sealDCBBallotMetadata.LockerPubKeys[index1]
		for index2 := index1 + 1; index2 < len(sealDCBBallotMetadata.LockerPubKeys); index2++ {
			pub2 := sealDCBBallotMetadata.LockerPubKeys[index2]
			if !common.ByteEqual(pub1, pub2) {
				return false
			}
		}
	}
	return true
}

type SealedLv1DCBBallotMetadata struct {
	sealedDCBBallot    SealedDCBBallot
	PointerToLv2Ballot common.Hash
	PointerToLv3Ballot common.Hash
	MetadataBase
}

func (sealedLv1DCBBallotMetadata *SealedLv1DCBBallotMetadata) ValidateSanityData(bcr BlockchainRetriever, tx Transaction) (bool, bool, error) {
	_, ok, _ := sealedLv1DCBBallotMetadata.sealedDCBBallot.ValidateSanityData(bcr, tx)
	if !ok {
		return true, false, nil
	}
	return true, true, nil
}

func (sealedLv1DCBBallotMetadata *SealedLv1DCBBallotMetadata) ValidateMetadataByItself() bool {
	panic("implement me")
}

func NewSealedLv1DCBBallotMetadata(sealedBallot []byte, lockersPubKey [][]byte, pointerToLv2Ballot common.Hash, pointerToLv3Ballot common.Hash) *SealedLv1DCBBallotMetadata {
	return &SealedLv1DCBBallotMetadata{
		sealedDCBBallot:    *NewSealedDCBBallotMetadata(sealedBallot, lockersPubKey),
		PointerToLv2Ballot: pointerToLv2Ballot,
		PointerToLv3Ballot: pointerToLv3Ballot,
		MetadataBase:       *NewMetadataBase(SealedLv1DCBBallotMeta),
	}
}

func (sealedLv1DCBBallotMetadata *SealedLv1DCBBallotMetadata) Hash() *common.Hash {
	record := string(sealedLv1DCBBallotMetadata.sealedDCBBallot.Hash().GetBytes())
	record += string(sealedLv1DCBBallotMetadata.PointerToLv2Ballot.GetBytes())
	record += string(sealedLv1DCBBallotMetadata.PointerToLv3Ballot.GetBytes())
	record += string(sealedLv1DCBBallotMetadata.MetadataBase.Hash().GetBytes())
	hash := common.DoubleHashH([]byte(record))
	return &hash
}

func (sealedLv1DCBBallotMetadata *SealedLv1DCBBallotMetadata) ValidateTxWithBlockChain(tx Transaction, bcr BlockchainRetriever, chainID byte, db database.DatabaseInterface) (bool, error) {
	//Check base seal metadata
	ok, err := sealedLv1DCBBallotMetadata.sealedDCBBallot.ValidateTxWithBlockChain(tx, bcr, chainID, db)
	if err != nil || !ok {
		return ok, err
	}

	//Check precede transaction type
	_, _, _, lv2Tx, _ := bcr.GetTransactionByHash(&sealedLv1DCBBallotMetadata.PointerToLv2Ballot)
	if lv2Tx.GetMetadataType() != SealedLv2DCBBallotMeta {
		return false, nil
	}
	_, _, _, lv3Tx, _ := bcr.GetTransactionByHash(&sealedLv1DCBBallotMetadata.PointerToLv3Ballot)
	if lv3Tx.GetMetadataType() != SealedLv3DCBBallotMeta {
		return false, nil
	}

	// check 2 array equal
	metaLv2 := lv2Tx.GetMetadata().(*SealedLv2DCBBallotMetadata)
	for i := 0; i < len(sealedLv1DCBBallotMetadata.sealedDCBBallot.LockerPubKeys); i++ {
		if !common.ByteEqual(sealedLv1DCBBallotMetadata.sealedDCBBallot.LockerPubKeys[i], metaLv2.sealedDCBBallot.LockerPubKeys[i]) {
			return false, nil
		}
	}

	// Check encrypting
	if !common.ByteEqual(sealedLv1DCBBallotMetadata.sealedDCBBallot.BallotData, common.Encrypt(metaLv2.sealedDCBBallot.BallotData, metaLv2.sealedDCBBallot.LockerPubKeys[1]).([]byte)) {
		return false, nil
	}
	return true, nil
}

type SealedLv2DCBBallotMetadata struct {
	sealedDCBBallot    SealedDCBBallot
	PointerToLv3Ballot common.Hash
	MetadataBase
}

func (sealedLv2DCBBallotMetadata *SealedLv2DCBBallotMetadata) ValidateSanityData(bcr BlockchainRetriever, tx Transaction) (bool, bool, error) {
	_, ok, _ := sealedLv2DCBBallotMetadata.sealedDCBBallot.ValidateSanityData(bcr, tx)
	if !ok {
		return true, false, nil
	}
	return true, true, nil
}

func (sealedLv2DCBBallotMetadata *SealedLv2DCBBallotMetadata) ValidateMetadataByItself() bool {
	return true
}

func NewSealedLv2DCBBallotMetadata(sealedBallot []byte, lockerPubKeys [][]byte, pointerToLv3Ballot common.Hash) *SealedLv2DCBBallotMetadata {
	return &SealedLv2DCBBallotMetadata{
		sealedDCBBallot:    *NewSealedDCBBallotMetadata(sealedBallot, lockerPubKeys),
		PointerToLv3Ballot: pointerToLv3Ballot,
		MetadataBase:       *NewMetadataBase(SealedLv2DCBBallotMeta),
	}
}

func (sealedLv2DCBBallotMetadata *SealedLv2DCBBallotMetadata) Hash() *common.Hash {
	record := string(sealedLv2DCBBallotMetadata.sealedDCBBallot.Hash().GetBytes())
	record += string(sealedLv2DCBBallotMetadata.PointerToLv3Ballot.GetBytes())
	record += string(sealedLv2DCBBallotMetadata.MetadataBase.Hash().GetBytes())
	hash := common.DoubleHashH([]byte(record))
	return &hash
}

func (sealedLv2DCBBallotMetadata *SealedLv2DCBBallotMetadata) ValidateTxWithBlockChain(tx Transaction, bcr BlockchainRetriever, chainID byte, db database.DatabaseInterface) (bool, error) {
	//Check base seal metadata
	ok, err := sealedLv2DCBBallotMetadata.sealedDCBBallot.ValidateTxWithBlockChain(tx, bcr, chainID, db)
	if err != nil || !ok {
		return ok, err
	}

	//Check precede transaction type
	_, _, _, lv3Tx, _ := bcr.GetTransactionByHash(&sealedLv2DCBBallotMetadata.PointerToLv3Ballot)
	if lv3Tx.GetMetadataType() != SealedLv3DCBBallotMeta {
		return false, nil
	}

	// check 2 array equal
	metaLv3 := lv3Tx.GetMetadata().(*SealedLv3DCBBallotMetadata)
	for i := 0; i < len(sealedLv2DCBBallotMetadata.sealedDCBBallot.LockerPubKeys); i++ {
		if !common.ByteEqual(sealedLv2DCBBallotMetadata.sealedDCBBallot.LockerPubKeys[i], metaLv3.SealedDCBBallot.LockerPubKeys[i]) {
			return false, nil
		}
	}

	// Check encrypting
	if !common.ByteEqual(sealedLv2DCBBallotMetadata.sealedDCBBallot.BallotData, common.Encrypt(metaLv3.SealedDCBBallot.BallotData, metaLv3.SealedDCBBallot.LockerPubKeys[2]).([]byte)) {
		return false, nil
	}
	return true, nil
}

type SealedLv3DCBBallotMetadata struct {
	SealedDCBBallot SealedDCBBallot
	MetadataBase
}

func (sealLv3DCBBallotMetadata *SealedLv3DCBBallotMetadata) ValidateTxWithBlockChain(tx Transaction, bcr BlockchainRetriever, b byte, db database.DatabaseInterface) (bool, error) {
	return sealLv3DCBBallotMetadata.ValidateTxWithBlockChain(tx, bcr, b, db)
}

func (sealLv3DCBBallotMetadata *SealedLv3DCBBallotMetadata) ValidateSanityData(bcr BlockchainRetriever, tx Transaction) (bool, bool, error) {
	return sealLv3DCBBallotMetadata.ValidateSanityData(bcr, tx)
}

func (sealLv3DCBBallotMetadata *SealedLv3DCBBallotMetadata) ValidateMetadataByItself() bool {
	return sealLv3DCBBallotMetadata.ValidateMetadataByItself()
}

func NewSealedLv3DCBBallotMetadata(sealedBallot []byte, lockerPubKeys [][]byte) *SealedLv3DCBBallotMetadata {
	return &SealedLv3DCBBallotMetadata{
		SealedDCBBallot: *NewSealedDCBBallotMetadata(sealedBallot, lockerPubKeys),
		MetadataBase:    *NewMetadataBase(SealedLv3DCBBallotMeta),
	}
}

type NormalDCBBallotFromSealerMetadata struct {
	Ballot             []byte
	LockerPubKey       [][]byte
	PointerToLv1Ballot common.Hash
	PointerToLv3Ballot common.Hash
	MetadataBase
}

func NewNormalDCBBallotFromSealerMetadata(ballot []byte, lockerPubKey [][]byte, pointerToLv1Ballot common.Hash, pointerToLv3Ballot common.Hash) *NormalDCBBallotFromSealerMetadata {
	return &NormalDCBBallotFromSealerMetadata{
		Ballot:             ballot,
		LockerPubKey:       lockerPubKey,
		PointerToLv1Ballot: pointerToLv1Ballot,
		PointerToLv3Ballot: pointerToLv3Ballot,
		MetadataBase:       *NewMetadataBase(NormalDCBBallotMetaFromSealerMeta),
	}
}

func (normalDCBBallotFromSealerMetadata *NormalDCBBallotFromSealerMetadata) ValidateSanityData(BlockchainRetriever, Transaction) (bool, bool, error) {
	for _, i := range normalDCBBallotFromSealerMetadata.LockerPubKey {
		if len(i) != common.PubKeyLength {
			return true, false, nil
		}
	}
	return true, true, nil
}

func (normalDCBBallotFromSealerMetadata *NormalDCBBallotFromSealerMetadata) ValidateMetadataByItself() bool {
	for index1 := 0; index1 < len(normalDCBBallotFromSealerMetadata.LockerPubKey); index1++ {
		pub1 := normalDCBBallotFromSealerMetadata.LockerPubKey[index1]
		for index2 := index1 + 1; index2 < len(normalDCBBallotFromSealerMetadata.LockerPubKey); index2++ {
			pub2 := normalDCBBallotFromSealerMetadata.LockerPubKey[index2]
			if !common.ByteEqual(pub1, pub2) {
				return false
			}
		}
	}
	return true
}

func (normalDCBBallotFromSealerMetadata *NormalDCBBallotFromSealerMetadata) Hash() *common.Hash {
	record := string(normalDCBBallotFromSealerMetadata.Ballot)
	for _, i := range normalDCBBallotFromSealerMetadata.LockerPubKey {
		record += string(i)
	}
	record += string(normalDCBBallotFromSealerMetadata.PointerToLv1Ballot.GetBytes())
	record += string(normalDCBBallotFromSealerMetadata.PointerToLv3Ballot.GetBytes())
	record += string(normalDCBBallotFromSealerMetadata.MetadataBase.Hash().GetBytes())
	hash := common.DoubleHashH([]byte(record))
	return &hash
}

func (normalDCBBallotFromSealerMetadata *NormalDCBBallotFromSealerMetadata) ValidateTxWithBlockChain(tx Transaction, bcr BlockchainRetriever, chainID byte, db database.DatabaseInterface) (bool, error) {
	//Validate these pubKeys are in board
	dcbBoardPubKeys := bcr.GetDCBBoardPubKeys()
	for _, j := range normalDCBBallotFromSealerMetadata.LockerPubKey {
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
	_, _, _, lv1Tx, _ := bcr.GetTransactionByHash(&normalDCBBallotFromSealerMetadata.PointerToLv1Ballot)
	if lv1Tx.GetMetadataType() != SealedLv1DCBBallotMeta {
		return false, nil
	}
	_, _, _, lv3Tx, _ := bcr.GetTransactionByHash(&normalDCBBallotFromSealerMetadata.PointerToLv3Ballot)
	if lv3Tx.GetMetadataType() != SealedLv3DCBBallotMeta {
		return false, nil
	}

	// check 2 array equal
	metaLv1 := lv1Tx.GetMetadata().(*SealedLv1DCBBallotMetadata)
	for i := 0; i < len(normalDCBBallotFromSealerMetadata.LockerPubKey); i++ {
		if !common.ByteEqual(normalDCBBallotFromSealerMetadata.LockerPubKey[i], metaLv1.sealedDCBBallot.LockerPubKeys[i]) {
			return false, nil
		}
	}

	// Check encrypting
	if !common.ByteEqual(normalDCBBallotFromSealerMetadata.Ballot, common.Encrypt(metaLv1.sealedDCBBallot.BallotData, metaLv1.sealedDCBBallot.LockerPubKeys[0]).([]byte)) {
		return false, nil
	}
	return true, nil
}

type NormalDCBBallotFromOwnerMetadata struct {
	Ballot             []byte
	LockerPubKey       [][]byte
	PointerToLv3Ballot common.Hash
	MetadataBase
}

func NewNormalDCBBallotFromOwnerMetadata(ballot []byte, lockerPubKey [][]byte, pointerToLv3Ballot common.Hash) *NormalDCBBallotFromOwnerMetadata {
	return &NormalDCBBallotFromOwnerMetadata{
		Ballot:             ballot,
		LockerPubKey:       lockerPubKey,
		PointerToLv3Ballot: pointerToLv3Ballot,
		MetadataBase:       *NewMetadataBase(NormalDCBBallotMetaFromOwnerMeta),
	}
}

func (normalDCBBallotFromOwnerMetadata *NormalDCBBallotFromOwnerMetadata) Hash() *common.Hash {
	record := string(normalDCBBallotFromOwnerMetadata.Ballot)
	for _, i := range normalDCBBallotFromOwnerMetadata.LockerPubKey {
		record += string(i)
	}
	record += string(normalDCBBallotFromOwnerMetadata.PointerToLv3Ballot.GetBytes())
	record += string(normalDCBBallotFromOwnerMetadata.MetadataBase.Hash().GetBytes())
	hash := common.DoubleHashH([]byte(record))
	return &hash
}

func (normalDCBBallotFromOwnerMetadata *NormalDCBBallotFromOwnerMetadata) ValidateTxWithBlockChain(tx Transaction, bcr BlockchainRetriever, chainID byte, db database.DatabaseInterface) (bool, error) {
	//Validate these pubKeys are in board
	dcbBoardPubKeys := bcr.GetDCBBoardPubKeys()
	for _, j := range normalDCBBallotFromOwnerMetadata.LockerPubKey {
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
	_, _, _, lv3Tx, _ := bcr.GetTransactionByHash(&normalDCBBallotFromOwnerMetadata.PointerToLv3Ballot)
	if lv3Tx.GetMetadataType() != SealedLv3DCBBallotMeta {
		return false, nil
	}

	// check 2 array equal
	metaLv3 := lv3Tx.GetMetadata().(*SealedLv3DCBBallotMetadata)
	for i := 0; i < len(normalDCBBallotFromOwnerMetadata.LockerPubKey); i++ {
		if !common.ByteEqual(normalDCBBallotFromOwnerMetadata.LockerPubKey[i], metaLv3.SealedDCBBallot.LockerPubKeys[i]) {
			return false, nil
		}
	}

	// Check encrypting
	if !common.ByteEqual(
		metaLv3.SealedDCBBallot.BallotData,
		common.Encrypt(
			common.Encrypt(
				common.Encrypt(
					normalDCBBallotFromOwnerMetadata.Ballot,
					metaLv3.SealedDCBBallot.LockerPubKeys[2],
				),
				metaLv3.SealedDCBBallot.LockerPubKeys[1],
			),
			metaLv3.SealedDCBBallot.LockerPubKeys[0],
		).([]byte)) {
		return false, nil
	}
	return true, nil
}

func (normalDCBBallotFromOwnerMetadata *NormalDCBBallotFromOwnerMetadata) ValidateSanityData(BlockchainRetriever, Transaction) (bool, bool, error) {
	for _, i := range normalDCBBallotFromOwnerMetadata.LockerPubKey {
		if len(i) != common.PubKeyLength {
			return true, false, nil
		}
	}
	return true, true, nil
}

func (normalDCBBallotFromOwnerMetadata *NormalDCBBallotFromOwnerMetadata) ValidateMetadataByItself() bool {
	for index1 := 0; index1 < len(normalDCBBallotFromOwnerMetadata.LockerPubKey); index1++ {
		pub1 := normalDCBBallotFromOwnerMetadata.LockerPubKey[index1]
		for index2 := index1 + 1; index2 < len(normalDCBBallotFromOwnerMetadata.LockerPubKey); index2++ {
			pub2 := normalDCBBallotFromOwnerMetadata.LockerPubKey[index2]
			if !common.ByteEqual(pub1, pub2) {
				return false
			}
		}
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
