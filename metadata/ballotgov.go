package metadata

import (
	"github.com/ninjadotorg/constant/common"
)

//abstract class
type SealedGOVBallotMetadata struct {
	SealedBallot []byte
	LockerPubKey [][]byte

	MetadataBase
}

func (sealGOVBallotMetadata *SealedGOVBallotMetadata) Hash() *common.Hash {
	record := string(sealGOVBallotMetadata.SealedBallot)
	for _, i := range sealGOVBallotMetadata.LockerPubKey {
		record += string(i)
	}
	hash := common.DoubleHashH([]byte(record))
	return &hash
}

func (sealGOVBallotMetadata *SealedGOVBallotMetadata) ValidateTxWithBlockChain(tx Transaction, bcr BlockchainRetriever, chainID byte) (bool, error) {
	//Validate these pubKeys are in board
	govBoardPubKeys := bcr.GetGOVBoardPubKeys()
	for _, j := range sealGOVBallotMetadata.LockerPubKey {
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
	return true, nil
}

func (sealGOVBallotMetadata *SealedGOVBallotMetadata) ValidateSanityData(BlockchainRetriever, Transaction) (bool, bool, error) {
	for _, i := range sealGOVBallotMetadata.LockerPubKey {
		if len(i) != common.HashSize {
			return true, false, nil
		}
	}
	return true, true, nil
}

func (sealGOVBallotMetadata *SealedGOVBallotMetadata) ValidateMetadataByItself() bool {
	for index1 := 0; index1 < len(sealGOVBallotMetadata.LockerPubKey); index1++ {
		pub1 := sealGOVBallotMetadata.LockerPubKey[index1]
		for index2 := index1 + 1; index2 < len(sealGOVBallotMetadata.LockerPubKey); index2++ {
			pub2 := sealGOVBallotMetadata.LockerPubKey[index2]
			if !common.ByteEqual(pub1, pub2) {
				return false
			}
		}
	}
	return true
}

type SealedLv1GOVBallotMetadata struct {
	SealedGOVBallotMetadata
	PointerToLv2Ballot *common.Hash
	PointerToLv3Ballot *common.Hash
}

func NewSealedLv1GOVBallotMetadata(data map[string]interface{}) *SealedLv1GOVBallotMetadata {
	return &SealedLv1GOVBallotMetadata{
		SealedGOVBallotMetadata: SealedGOVBallotMetadata{
			SealedBallot: data["SealedBallot"].([]byte),
			LockerPubKey: data["LockerPubKey"].([][]byte),
			MetadataBase: MetadataBase{
				Type: SealedLv1GOVBallotMeta,
			},
		},
		PointerToLv2Ballot: data["PointerToLv2Ballot"].(*common.Hash),
		PointerToLv3Ballot: data["PointerToLv3Ballot"].(*common.Hash),
	}
}

func (sealedLv1GOVBallotMetadata *SealedLv1GOVBallotMetadata) Hash() *common.Hash {
	record := string(common.ToBytes(sealedLv1GOVBallotMetadata.SealedGOVBallotMetadata.Hash()))
	record += string(common.ToBytes(sealedLv1GOVBallotMetadata.PointerToLv2Ballot))
	record += string(common.ToBytes(sealedLv1GOVBallotMetadata.PointerToLv3Ballot))
	hash := common.DoubleHashH([]byte(record))
	return &hash
}

func (sealedLv1GOVBallotMetadata *SealedLv1GOVBallotMetadata) ValidateTxWithBlockChain(tx Transaction, bcr BlockchainRetriever, chainID byte) (bool, error) {
	//Check base seal metadata
	ok, err := sealedLv1GOVBallotMetadata.SealedGOVBallotMetadata.ValidateTxWithBlockChain(tx, bcr, chainID)
	if err != nil || !ok {
		return ok, err
	}

	//Check precede transaction type
	_, _, _, lv2Tx, _ := bcr.GetTransactionByHash(sealedLv1GOVBallotMetadata.PointerToLv2Ballot)
	if lv2Tx.GetMetadataType() != SealedLv2GOVBallotMeta {
		return false, nil
	}
	_, _, _, lv3Tx, _ := bcr.GetTransactionByHash(sealedLv1GOVBallotMetadata.PointerToLv3Ballot)
	if lv3Tx.GetMetadataType() != SealedLv3GOVBallotMeta {
		return false, nil
	}

	// check 2 array equal
	metaLv2 := lv2Tx.GetMetadata().(*SealedLv2GOVBallotMetadata)
	for i := 0; i < len(sealedLv1GOVBallotMetadata.LockerPubKey); i++ {
		if !common.ByteEqual(sealedLv1GOVBallotMetadata.LockerPubKey[i], metaLv2.LockerPubKey[i]) {
			return false, nil
		}
	}

	// Check encrypting
	if !common.ByteEqual(sealedLv1GOVBallotMetadata.SealedBallot, common.Encrypt(metaLv2.SealedBallot, metaLv2.LockerPubKey[1]).([]byte)) {
		return false, nil
	}
	return true, nil
}

type SealedLv2GOVBallotMetadata struct {
	SealedGOVBallotMetadata
	PointerToLv3Ballot *common.Hash
}

func NewSealedLv2GOVBallotMetadata(data map[string]interface{}) *SealedLv2GOVBallotMetadata {
	return &SealedLv2GOVBallotMetadata{
		SealedGOVBallotMetadata: SealedGOVBallotMetadata{
			SealedBallot: data["SealedBallot"].([]byte),
			LockerPubKey: data["LockerPubKey"].([][]byte),
			MetadataBase: MetadataBase{
				Type: SealedLv2GOVBallotMeta,
			},
		},
		PointerToLv3Ballot: data["PointerToLv3Ballot"].(*common.Hash),
	}
}

func (sealedLv2GOVBallotMetadata *SealedLv2GOVBallotMetadata) Hash() *common.Hash {
	record := string(common.ToBytes(sealedLv2GOVBallotMetadata.SealedGOVBallotMetadata.Hash()))
	record += string(common.ToBytes(sealedLv2GOVBallotMetadata.PointerToLv3Ballot))
	hash := common.DoubleHashH([]byte(record))
	return &hash
}

func (sealedLv2GOVBallotMetadata *SealedLv2GOVBallotMetadata) ValidateTxWithBlockChain(tx Transaction, bcr BlockchainRetriever, chainID byte) (bool, error) {
	//Check base seal metadata
	ok, err := sealedLv2GOVBallotMetadata.SealedGOVBallotMetadata.ValidateTxWithBlockChain(tx, bcr, chainID)
	if err != nil || !ok {
		return ok, err
	}

	//Check precede transaction type
	_, _, _, lv3Tx, _ := bcr.GetTransactionByHash(sealedLv2GOVBallotMetadata.PointerToLv3Ballot)
	if lv3Tx.GetMetadataType() != SealedLv3GOVBallotMeta {
		return false, nil
	}

	// check 2 array equal
	metaLv3 := lv3Tx.GetMetadata().(*SealedLv3GOVBallotMetadata)
	for i := 0; i < len(sealedLv2GOVBallotMetadata.LockerPubKey); i++ {
		if !common.ByteEqual(sealedLv2GOVBallotMetadata.LockerPubKey[i], metaLv3.LockerPubKey[i]) {
			return false, nil
		}
	}

	// Check encrypting
	if !common.ByteEqual(sealedLv2GOVBallotMetadata.SealedBallot, common.Encrypt(metaLv3.SealedBallot, metaLv3.LockerPubKey[2]).([]byte)) {
		return false, nil
	}
	return true, nil
}

type SealedLv3GOVBallotMetadata struct {
	SealedGOVBallotMetadata
}

func NewSealedLv3GOVBallotMetadata(data map[string]interface{}) *SealedLv3GOVBallotMetadata {
	return &SealedLv3GOVBallotMetadata{
		SealedGOVBallotMetadata: SealedGOVBallotMetadata{
			SealedBallot: data["SealedBallot"].([]byte),
			LockerPubKey: data["LockerPubKey"].([][]byte),
			MetadataBase: MetadataBase{
				Type: SealedLv3GOVBallotMeta,
			},
		},
	}
}

type NormalGOVBallotFromSealerMetadata struct {
	Ballot             []byte
	LockerPubKey       [][]byte
	PointerToLv1Ballot *common.Hash
	PointerToLv3Ballot *common.Hash
	MetadataBase
}

func (normalGOVBallotFromSealerMetadata *NormalGOVBallotFromSealerMetadata) ValidateSanityData(BlockchainRetriever, Transaction) (bool, bool, error) {
	for _, i := range normalGOVBallotFromSealerMetadata.LockerPubKey {
		if len(i) != common.HashSize {
			return true, false, nil
		}
	}
	return true, true, nil
}

func (normalGOVBallotFromSealerMetadata *NormalGOVBallotFromSealerMetadata) ValidateMetadataByItself() bool {
	for index1 := 0; index1 < len(normalGOVBallotFromSealerMetadata.LockerPubKey); index1++ {
		pub1 := normalGOVBallotFromSealerMetadata.LockerPubKey[index1]
		for index2 := index1 + 1; index2 < len(normalGOVBallotFromSealerMetadata.LockerPubKey); index2++ {
			pub2 := normalGOVBallotFromSealerMetadata.LockerPubKey[index2]
			if !common.ByteEqual(pub1, pub2) {
				return false
			}
		}
	}
	return true
}

func NewNormalGOVBallotFromSealerMetadata(data map[string]interface{}) *NormalGOVBallotFromSealerMetadata {
	return &NormalGOVBallotFromSealerMetadata{
		Ballot:             data["Ballot"].([]byte),
		LockerPubKey:       data["LockerPubKey"].([][]byte),
		PointerToLv1Ballot: data["PointerToLv1Ballot"].(*common.Hash),
		PointerToLv3Ballot: data["PointerToLv3Ballot"].(*common.Hash),
		MetadataBase: MetadataBase{
			Type: NormalGOVBallotMetaFromSealer,
		},
	}
}

func (normalGOVBallotFromSealerMetadata *NormalGOVBallotFromSealerMetadata) Hash() *common.Hash {
	record := string(normalGOVBallotFromSealerMetadata.Ballot)
	for _, i := range normalGOVBallotFromSealerMetadata.LockerPubKey {
		record += string(i)
	}
	record += string(common.ToBytes(normalGOVBallotFromSealerMetadata.PointerToLv1Ballot))
	record += string(common.ToBytes(normalGOVBallotFromSealerMetadata.PointerToLv3Ballot))
	hash := common.DoubleHashH([]byte(record))
	return &hash
}

func (normalGOVBallotFromSealerMetadata *NormalGOVBallotFromSealerMetadata) ValidateTxWithBlockChain(tx Transaction, bcr BlockchainRetriever, chainID byte) (bool, error) {
	//Validate these pubKeys are in board
	govBoardPubKeys := bcr.GetGOVBoardPubKeys()
	for _, j := range normalGOVBallotFromSealerMetadata.LockerPubKey {
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
	_, _, _, lv1Tx, _ := bcr.GetTransactionByHash(normalGOVBallotFromSealerMetadata.PointerToLv1Ballot)
	if lv1Tx.GetMetadataType() != SealedLv1GOVBallotMeta {
		return false, nil
	}
	_, _, _, lv3Tx, _ := bcr.GetTransactionByHash(normalGOVBallotFromSealerMetadata.PointerToLv3Ballot)
	if lv3Tx.GetMetadataType() != SealedLv3GOVBallotMeta {
		return false, nil
	}

	// check 2 array equal
	metaLv1 := lv1Tx.GetMetadata().(*SealedLv1GOVBallotMetadata)
	for i := 0; i < len(normalGOVBallotFromSealerMetadata.LockerPubKey); i++ {
		if !common.ByteEqual(normalGOVBallotFromSealerMetadata.LockerPubKey[i], metaLv1.LockerPubKey[i]) {
			return false, nil
		}
	}

	// Check encrypting
	if !common.ByteEqual(normalGOVBallotFromSealerMetadata.Ballot, common.Encrypt(metaLv1.SealedBallot, metaLv1.LockerPubKey[0]).([]byte)) {
		return false, nil
	}
	return true, nil
}

type NormalGOVBallotFromOwnerMetadata struct {
	Ballot             []byte
	LockerPubKey       [][]byte
	PointerToLv3Ballot *common.Hash
	MetadataBase
}

func NewNormalGOVBallotFromOwnerMetadata(data map[string]interface{}) *NormalGOVBallotFromOwnerMetadata {
	return &NormalGOVBallotFromOwnerMetadata{
		Ballot:             data["Ballot"].([]byte),
		LockerPubKey:       data["LockerPubKey"].([][]byte),
		PointerToLv3Ballot: data["PointerToLv3Ballot"].(*common.Hash),
		MetadataBase: MetadataBase{
			Type: NormalGOVBallotMetaFromOwner,
		},
	}
}

func (normalGOVBallotFromOwnerMetadata *NormalGOVBallotFromOwnerMetadata) Hash() *common.Hash {
	record := string(normalGOVBallotFromOwnerMetadata.Ballot)
	for _, i := range normalGOVBallotFromOwnerMetadata.LockerPubKey {
		record += string(i)
	}
	record += string(common.ToBytes(normalGOVBallotFromOwnerMetadata.PointerToLv3Ballot))
	hash := common.DoubleHashH([]byte(record))
	return &hash
}

func (normalGOVBallotFromOwnerMetadata *NormalGOVBallotFromOwnerMetadata) ValidateTxWithBlockChain(tx Transaction, bcr BlockchainRetriever, chainID byte) (bool, error) {
	//Validate these pubKeys are in board
	govBoardPubKeys := bcr.GetGOVBoardPubKeys()
	for _, j := range normalGOVBallotFromOwnerMetadata.LockerPubKey {
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
	_, _, _, lv3Tx, _ := bcr.GetTransactionByHash(normalGOVBallotFromOwnerMetadata.PointerToLv3Ballot)
	if lv3Tx.GetMetadataType() != SealedLv3GOVBallotMeta {
		return false, nil
	}

	// check 2 array equal
	metaLv3 := lv3Tx.GetMetadata().(*SealedLv3GOVBallotMetadata)
	for i := 0; i < len(normalGOVBallotFromOwnerMetadata.LockerPubKey); i++ {
		if !common.ByteEqual(normalGOVBallotFromOwnerMetadata.LockerPubKey[i], metaLv3.LockerPubKey[i]) {
			return false, nil
		}
	}

	// Check encrypting
	if !common.ByteEqual(
		metaLv3.SealedBallot,
		common.Encrypt(
			common.Encrypt(
				common.Encrypt(
					normalGOVBallotFromOwnerMetadata.Ballot,
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

func (normalGOVBallotFromOwnerMetadata *NormalGOVBallotFromOwnerMetadata) ValidateSanityData(BlockchainRetriever, Transaction) (bool, bool, error) {
	for _, i := range normalGOVBallotFromOwnerMetadata.LockerPubKey {
		if len(i) != common.HashSize {
			return true, false, nil
		}
	}
	return true, true, nil
}

func (normalGOVBallotFromOwnerMetadata *NormalGOVBallotFromOwnerMetadata) ValidateMetadataByItself() bool {
	for index1 := 0; index1 < len(normalGOVBallotFromOwnerMetadata.LockerPubKey); index1++ {
		pub1 := normalGOVBallotFromOwnerMetadata.LockerPubKey[index1]
		for index2 := index1 + 1; index2 < len(normalGOVBallotFromOwnerMetadata.LockerPubKey); index2++ {
			pub2 := normalGOVBallotFromOwnerMetadata.LockerPubKey[index2]
			if !common.ByteEqual(pub1, pub2) {
				return false
			}
		}
	}
	return true
}
