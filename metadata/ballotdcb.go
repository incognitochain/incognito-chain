package metadata

import (
	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/database"
)

//abstract class
type SealedDCBBallotMetadata struct {
	SealedBallot []byte
	LockerPubKey [][]byte

	MetadataBase
}

func (sealDCBBallotMetadata *SealedDCBBallotMetadata) Hash() *common.Hash {
	record := string(sealDCBBallotMetadata.SealedBallot)
	for _, i := range sealDCBBallotMetadata.LockerPubKey {
		record += string(i)
	}
	hash := common.DoubleHashH([]byte(record))
	return &hash
}

func (sealDCBBallotMetadata *SealedDCBBallotMetadata) ValidateTxWithBlockChain(tx Transaction, bcr BlockchainRetriever, chainID byte, db database.DatabaseInterface) (bool, error) {
	//Validate these pubKeys are in board
	dcbBoardPubKeys := bcr.GetDCBBoardPubKeys()
	for _, j := range sealDCBBallotMetadata.LockerPubKey {
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

func (sealDCBBallotMetadata *SealedDCBBallotMetadata) ValidateSanityData(BlockchainRetriever, Transaction) (bool, bool, error) {
	for _, i := range sealDCBBallotMetadata.LockerPubKey {
		if len(i) != common.HashSize {
			return true, false, nil
		}
	}
	return true, true, nil
}

func (sealDCBBallotMetadata *SealedDCBBallotMetadata) ValidateMetadataByItself() bool {
	for index1 := 0; index1 < len(sealDCBBallotMetadata.LockerPubKey); index1++ {
		pub1 := sealDCBBallotMetadata.LockerPubKey[index1]
		for index2 := index1 + 1; index2 < len(sealDCBBallotMetadata.LockerPubKey); index2++ {
			pub2 := sealDCBBallotMetadata.LockerPubKey[index2]
			if !common.ByteEqual(pub1, pub2) {
				return false
			}
		}
	}
	return true
}

type SealedLv1DCBBallotMetadata struct {
	SealedDCBBallotMetadata
	PointerToLv2Ballot common.Hash
	PointerToLv3Ballot common.Hash
}

func NewSealedLv1DCBBallotMetadata(data map[string]interface{}) *SealedLv1DCBBallotMetadata {
	return &SealedLv1DCBBallotMetadata{
		SealedDCBBallotMetadata: SealedDCBBallotMetadata{
			SealedBallot: data["SealedBallot"].([]byte),
			LockerPubKey: data["LockerPubKey"].([][]byte),
			MetadataBase: MetadataBase{
				Type: SealedLv1DCBBallotMeta,
			},
		},
		PointerToLv2Ballot: data["PointerToLv2Ballot"].(common.Hash),
		PointerToLv3Ballot: data["PointerToLv3Ballot"].(common.Hash),
	}
}

func (sealedLv1DCBBallotMetadata *SealedLv1DCBBallotMetadata) Hash() *common.Hash {
	record := string(common.ToBytes(*sealedLv1DCBBallotMetadata.SealedDCBBallotMetadata.Hash()))
	record += string(common.ToBytes(sealedLv1DCBBallotMetadata.PointerToLv2Ballot))
	record += string(common.ToBytes(sealedLv1DCBBallotMetadata.PointerToLv3Ballot))
	hash := common.DoubleHashH([]byte(record))
	return &hash
}

func (sealedLv1DCBBallotMetadata *SealedLv1DCBBallotMetadata) ValidateTxWithBlockChain(tx Transaction, bcr BlockchainRetriever, chainID byte, db database.DatabaseInterface) (bool, error) {
	//Check base seal metadata
	ok, err := sealedLv1DCBBallotMetadata.SealedDCBBallotMetadata.ValidateTxWithBlockChain(tx, bcr, chainID, db)
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
	for i := 0; i < len(sealedLv1DCBBallotMetadata.LockerPubKey); i++ {
		if !common.ByteEqual(sealedLv1DCBBallotMetadata.LockerPubKey[i], metaLv2.LockerPubKey[i]) {
			return false, nil
		}
	}

	// Check encrypting
	if !common.ByteEqual(sealedLv1DCBBallotMetadata.SealedBallot, common.Encrypt(metaLv2.SealedBallot, metaLv2.LockerPubKey[1]).([]byte)) {
		return false, nil
	}
	return true, nil
}

type SealedLv2DCBBallotMetadata struct {
	SealedDCBBallotMetadata
	PointerToLv3Ballot common.Hash
}

func NewSealedLv2DCBBallotMetadata(data map[string]interface{}) *SealedLv2DCBBallotMetadata {
	return &SealedLv2DCBBallotMetadata{
		SealedDCBBallotMetadata: SealedDCBBallotMetadata{
			SealedBallot: data["SealedBallot"].([]byte),
			LockerPubKey: data["LockerPubKey"].([][]byte),
			MetadataBase: MetadataBase{
				Type: SealedLv2DCBBallotMeta,
			},
		},
		PointerToLv3Ballot: data["PointerToLv3Ballot"].(common.Hash),
	}
}

func (sealedLv2DCBBallotMetadata *SealedLv2DCBBallotMetadata) Hash() *common.Hash {
	record := string(common.ToBytes(*sealedLv2DCBBallotMetadata.SealedDCBBallotMetadata.Hash()))
	record += string(common.ToBytes(sealedLv2DCBBallotMetadata.PointerToLv3Ballot))
	hash := common.DoubleHashH([]byte(record))
	return &hash
}

func (sealedLv2DCBBallotMetadata *SealedLv2DCBBallotMetadata) ValidateTxWithBlockChain(tx Transaction, bcr BlockchainRetriever, chainID byte, db database.DatabaseInterface) (bool, error) {
	//Check base seal metadata
	ok, err := sealedLv2DCBBallotMetadata.SealedDCBBallotMetadata.ValidateTxWithBlockChain(tx, bcr, chainID, db)
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
	for i := 0; i < len(sealedLv2DCBBallotMetadata.LockerPubKey); i++ {
		if !common.ByteEqual(sealedLv2DCBBallotMetadata.LockerPubKey[i], metaLv3.LockerPubKey[i]) {
			return false, nil
		}
	}

	// Check encrypting
	if !common.ByteEqual(sealedLv2DCBBallotMetadata.SealedBallot, common.Encrypt(metaLv3.SealedBallot, metaLv3.LockerPubKey[2]).([]byte)) {
		return false, nil
	}
	return true, nil
}

type SealedLv3DCBBallotMetadata struct {
	SealedDCBBallotMetadata
}

func NewSealedLv3DCBBallotMetadata(data map[string]interface{}) *SealedLv3DCBBallotMetadata {
	return &SealedLv3DCBBallotMetadata{
		SealedDCBBallotMetadata: SealedDCBBallotMetadata{
			SealedBallot: data["SealedBallot"].([]byte),
			LockerPubKey: data["LockerPubKey"].([][]byte),
			MetadataBase: MetadataBase{
				Type: SealedLv3DCBBallotMeta,
			},
		},
	}
}

type NormalDCBBallotFromSealerMetadata struct {
	Ballot             []byte
	LockerPubKey       [][]byte
	PointerToLv1Ballot common.Hash
	PointerToLv3Ballot common.Hash
	MetadataBase
}

func (normalDCBBallotFromSealerMetadata *NormalDCBBallotFromSealerMetadata) ValidateSanityData(BlockchainRetriever, Transaction) (bool, bool, error) {
	for _, i := range normalDCBBallotFromSealerMetadata.LockerPubKey {
		if len(i) != common.HashSize {
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

func NewNormalDCBBallotFromSealerMetadata(data map[string]interface{}) *NormalDCBBallotFromSealerMetadata {
	return &NormalDCBBallotFromSealerMetadata{
		Ballot:             data["Ballot"].([]byte),
		LockerPubKey:       data["LockerPubKey"].([][]byte),
		PointerToLv1Ballot: data["PointerToLv1Ballot"].(common.Hash),
		PointerToLv3Ballot: data["PointerToLv3Ballot"].(common.Hash),
		MetadataBase: MetadataBase{
			Type: NormalDCBBallotMetaFromSealer,
		},
	}
}

func (normalDCBBallotFromSealerMetadata *NormalDCBBallotFromSealerMetadata) Hash() *common.Hash {
	record := string(normalDCBBallotFromSealerMetadata.Ballot)
	for _, i := range normalDCBBallotFromSealerMetadata.LockerPubKey {
		record += string(i)
	}
	record += string(common.ToBytes(normalDCBBallotFromSealerMetadata.PointerToLv1Ballot))
	record += string(common.ToBytes(normalDCBBallotFromSealerMetadata.PointerToLv3Ballot))
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
		if !common.ByteEqual(normalDCBBallotFromSealerMetadata.LockerPubKey[i], metaLv1.LockerPubKey[i]) {
			return false, nil
		}
	}

	// Check encrypting
	if !common.ByteEqual(normalDCBBallotFromSealerMetadata.Ballot, common.Encrypt(metaLv1.SealedBallot, metaLv1.LockerPubKey[0]).([]byte)) {
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

func NewNormalDCBBallotFromOwnerMetadata(data map[string]interface{}) *NormalDCBBallotFromOwnerMetadata {
	return &NormalDCBBallotFromOwnerMetadata{
		Ballot:             data["Ballot"].([]byte),
		LockerPubKey:       data["LockerPubKey"].([][]byte),
		PointerToLv3Ballot: data["PointerToLv3Ballot"].(common.Hash),
		MetadataBase: MetadataBase{
			Type: NormalDCBBallotMetaFromOwner,
		},
	}
}

func (normalDCBBallotFromOwnerMetadata *NormalDCBBallotFromOwnerMetadata) Hash() *common.Hash {
	record := string(normalDCBBallotFromOwnerMetadata.Ballot)
	for _, i := range normalDCBBallotFromOwnerMetadata.LockerPubKey {
		record += string(i)
	}
	record += string(common.ToBytes(normalDCBBallotFromOwnerMetadata.PointerToLv3Ballot))
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
		if !common.ByteEqual(normalDCBBallotFromOwnerMetadata.LockerPubKey[i], metaLv3.LockerPubKey[i]) {
			return false, nil
		}
	}

	// Check encrypting
	if !common.ByteEqual(
		metaLv3.SealedBallot,
		common.Encrypt(
			common.Encrypt(
				common.Encrypt(
					normalDCBBallotFromOwnerMetadata.Ballot,
					metaLv3.LockerPubKey[2],
				),
				metaLv3.LockerPubKey[1],
			),
			metaLv3.LockerPubKey[0],
		).([]byte)) {
		return false, nil
	}
	return true, nil
}

func (normalDCBBallotFromOwnerMetadata *NormalDCBBallotFromOwnerMetadata) ValidateSanityData(BlockchainRetriever, Transaction) (bool, bool, error) {
	for _, i := range normalDCBBallotFromOwnerMetadata.LockerPubKey {
		if len(i) != common.HashSize {
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
