package metadata

import (
	"encoding/hex"
	"github.com/constant-money/constant-chain/blockchain/component"
	"github.com/constant-money/constant-chain/wallet"

	"github.com/constant-money/constant-chain/common"
	"github.com/constant-money/constant-chain/database"
	"github.com/constant-money/constant-chain/privacy"
)

//abstract class
type SealedVoteProposal struct {
	SealVoteProposalData   []byte
	LockerPaymentAddresses []privacy.PaymentAddress
}

func NewSealedVoteProposalMetadata(sealedVoteProposal []byte, lockerPubKeys []privacy.PaymentAddress) *SealedVoteProposal {
	return &SealedVoteProposal{
		SealVoteProposalData:   sealedVoteProposal,
		LockerPaymentAddresses: lockerPubKeys,
	}
}

func (sealedVoteProposal *SealedVoteProposal) ToBytes() []byte {
	record := string(sealedVoteProposal.SealVoteProposalData)
	for _, i := range sealedVoteProposal.LockerPaymentAddresses {
		record += i.String()
	}
	return []byte(record)
}

func (sealedVoteProposal *SealedVoteProposal) ValidateLockerPubKeys(bcr BlockchainRetriever, boardType common.BoardType) (bool, error) {
	//Validate these pubKeys are in board
	boardPaymentAddress := bcr.GetBoardPaymentAddress(boardType)
	for _, j := range sealedVoteProposal.LockerPaymentAddresses {
		exist := false
		for _, i := range boardPaymentAddress {
			if common.ByteEqual(i.Bytes(), j.Bytes()) {
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
	return true, true, nil
}

func (sealedVoteProposal *SealedVoteProposal) ValidateMetadataByItself() bool {
	for index1 := 0; index1 < len(sealedVoteProposal.LockerPaymentAddresses); index1++ {
		pub1 := sealedVoteProposal.LockerPaymentAddresses[index1]
		for index2 := index1 + 1; index2 < len(sealedVoteProposal.LockerPaymentAddresses); index2++ {
			pub2 := sealedVoteProposal.LockerPaymentAddresses[index2]
			if common.ByteEqual(pub1.Bytes(), pub2.Bytes()) {
				return false
			}
		}
	}
	return true
}
func NewVoteProposalData(proposalTxID common.Hash, amountOfVote int32, constitutionIndex uint32) *component.VoteProposalData {
	return &component.VoteProposalData{ProposalTxID: proposalTxID, AmountOfVote: amountOfVote, ConstitutionIndex: constitutionIndex}
}

func NewVoteProposalDataFromJson(data interface{}) *component.VoteProposalData {
	voteProposalDataData := data.(map[string]interface{})

	proposalTxIDData, _ := hex.DecodeString(voteProposalDataData["ProposalTxID"].(string))
	proposalTxID, _ := common.NewHash(proposalTxIDData)
	constitutionIndex := uint32(voteProposalDataData["ConstitutionIndex"].(float64))
	return NewVoteProposalData(
		*proposalTxID,
		int32(voteProposalDataData["AmountOfVote"].(float64)),
		constitutionIndex,
	)
}

func NewVoteProposalDataFromBytes(b []byte) *component.VoteProposalData {
	lenB := len(b)
	newHash, _ := common.NewHash(b[:lenB-8])
	return NewVoteProposalData(
		*newHash,
		common.BytesToInt32(b[lenB-8:lenB-4]),
		common.BytesToUint32(b[lenB-4:]),
	)
}

type NormalVoteProposalFromSealerMetadata struct {
	VoteProposal             component.VoteProposalData
	LockerPaymentAddress     []privacy.PaymentAddress
	PointerToLv1VoteProposal common.Hash
	PointerToLv3VoteProposal common.Hash
}

func NewNormalVoteProposalFromSealerMetadata(
	voteProposal component.VoteProposalData,
	lockerPaymentAddress []privacy.PaymentAddress,
	pointerToLv1VoteProposal common.Hash,
	pointerToLv3VoteProposal common.Hash,
) *NormalVoteProposalFromSealerMetadata {
	return &NormalVoteProposalFromSealerMetadata{
		VoteProposal:             voteProposal,
		LockerPaymentAddress:     lockerPaymentAddress,
		PointerToLv1VoteProposal: pointerToLv1VoteProposal,
		PointerToLv3VoteProposal: pointerToLv3VoteProposal,
	}
}
func (normalVoteProposalFromSealerMetadata *NormalVoteProposalFromSealerMetadata) GetBoardType() common.BoardType {
	panic("overwrite me")
}
func (normalVoteProposalFromSealerMetadata *NormalVoteProposalFromSealerMetadata) ValidateSanityData(BlockchainRetriever, Transaction) (bool, bool, error) {
	return true, true, nil
}

func (normalVoteProposalFromSealerMetadata *NormalVoteProposalFromSealerMetadata) ValidateMetadataByItself() bool {
	for index1 := 0; index1 < len(normalVoteProposalFromSealerMetadata.LockerPaymentAddress); index1++ {
		pub1 := normalVoteProposalFromSealerMetadata.LockerPaymentAddress[index1]
		for index2 := index1 + 1; index2 < len(normalVoteProposalFromSealerMetadata.LockerPaymentAddress); index2++ {
			pub2 := normalVoteProposalFromSealerMetadata.LockerPaymentAddress[index2]
			if common.ByteEqual(pub1.Bytes(), pub2.Bytes()) {
				return false
			}
		}
	}
	return true
}

func (normalVoteProposalFromSealerMetadata *NormalVoteProposalFromSealerMetadata) ToBytes() []byte {
	record := string(normalVoteProposalFromSealerMetadata.VoteProposal.ToBytes())
	for _, i := range normalVoteProposalFromSealerMetadata.LockerPaymentAddress {
		record += i.String()
	}
	record += string(normalVoteProposalFromSealerMetadata.PointerToLv1VoteProposal.GetBytes())
	record += string(normalVoteProposalFromSealerMetadata.PointerToLv3VoteProposal.GetBytes())
	return []byte(record)
}

func (normalVoteProposalFromSealerMetadata *NormalVoteProposalFromSealerMetadata) ValidateBeforeNewBlock(tx Transaction, bcr BlockchainRetriever, shardID byte) bool {
	boardType := normalVoteProposalFromSealerMetadata.GetBoardType()
	endedPivot := bcr.GetConstitutionEndHeight(boardType, shardID)
	currentBlockHeight := bcr.GetCurrentBeaconBlockHeight(shardID) + 1
	lv1Pivot := endedPivot - uint64(common.EncryptionOnePhraseDuration)
	return currentBlockHeight < endedPivot && currentBlockHeight >= lv1Pivot
}

func (normalVoteProposalFromSealerMetadata *NormalVoteProposalFromSealerMetadata) ValidateTxWithBlockChain(boardType common.BoardType,
	transaction Transaction,
	bcr BlockchainRetriever,
	shardID byte,
	db database.DatabaseInterface) (bool, error) {
	boardPubKeys := bcr.GetBoardPaymentAddress(boardType)
	for _, j := range normalVoteProposalFromSealerMetadata.LockerPaymentAddress {
		exist := false
		for _, i := range boardPubKeys {
			if common.ByteEqual(i.Bytes(), j.Bytes()) {
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

type NormalVoteProposalFromOwnerMetadata struct {
	VoteProposal             component.VoteProposalData
	LockerPaymentAddress     []privacy.PaymentAddress
	PointerToLv3VoteProposal common.Hash
}

func NewNormalVoteProposalFromOwnerMetadata(
	voteProposal component.VoteProposalData,
	lockerPaymentAddress []privacy.PaymentAddress,
	pointerToLv3VoteProposal common.Hash,
) *NormalVoteProposalFromOwnerMetadata {
	return &NormalVoteProposalFromOwnerMetadata{
		VoteProposal:             voteProposal,
		LockerPaymentAddress:     lockerPaymentAddress,
		PointerToLv3VoteProposal: pointerToLv3VoteProposal,
	}
}

func (normalVoteProposalFromOwnerMetadata *NormalVoteProposalFromOwnerMetadata) ValidateBeforeNewBlock(boardType common.BoardType, tx Transaction, bcr BlockchainRetriever, shardID byte) bool {
	endedPivot := bcr.GetConstitutionEndHeight(boardType, shardID)
	currentBlockHeight := bcr.GetCurrentBeaconBlockHeight(shardID) + 1
	lv1Pivot := endedPivot - common.EncryptionOnePhraseDuration
	return currentBlockHeight < endedPivot && currentBlockHeight >= lv1Pivot
}

func (normalVoteProposalFromOwnerMetadata *NormalVoteProposalFromOwnerMetadata) ToBytes() []byte {
	record := string(normalVoteProposalFromOwnerMetadata.VoteProposal.ToBytes())
	for _, i := range normalVoteProposalFromOwnerMetadata.LockerPaymentAddress {
		record += i.String()
	}
	record += string(normalVoteProposalFromOwnerMetadata.PointerToLv3VoteProposal.GetBytes())
	return []byte(record)
}

func (normalVoteProposalFromOwnerMetadata *NormalVoteProposalFromOwnerMetadata) ValidateSanityData(BlockchainRetriever, Transaction) (bool, bool, error) {
	return true, true, nil
}

func (normalVoteProposalFromOwnerMetadata *NormalVoteProposalFromOwnerMetadata) ValidateMetadataByItself() bool {
	for index1 := 0; index1 < len(normalVoteProposalFromOwnerMetadata.LockerPaymentAddress); index1++ {
		pub1 := normalVoteProposalFromOwnerMetadata.LockerPaymentAddress[index1]
		for index2 := index1 + 1; index2 < len(normalVoteProposalFromOwnerMetadata.LockerPaymentAddress); index2++ {
			pub2 := normalVoteProposalFromOwnerMetadata.LockerPaymentAddress[index2]
			if common.ByteEqual(pub1.Bytes(), pub2.Bytes()) {
				return false
			}
		}
	}
	return true
}

func (normalVoteProposalFromOwnerMetadata *NormalVoteProposalFromOwnerMetadata) ValidateTxWithBlockChain(
	boardType common.BoardType,
	transaction Transaction,
	bcr BlockchainRetriever,
	shardID byte,
	db database.DatabaseInterface) (bool,
	error) {
	boardPaymentAddress := bcr.GetBoardPaymentAddress(boardType)
	for _, j := range normalVoteProposalFromOwnerMetadata.LockerPaymentAddress {
		exist := false
		for _, i := range boardPaymentAddress {
			if common.ByteEqual(i.Bytes(), j.Bytes()) {
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

type PunishDecryptMetadata struct {
	PaymentAddress privacy.PaymentAddress
}

func (punishDecryptMetadata PunishDecryptMetadata) ToBytes() []byte {
	return punishDecryptMetadata.PaymentAddress.Bytes()
}

func GetPaymentAddressFromSenderKeyParams(keyParam string) (*privacy.PaymentAddress, error) {
	keyWallet, err := wallet.Base58CheckDeserialize(keyParam)
	if err != nil {
		return nil, err
	}
	return &keyWallet.KeySet.PaymentAddress, nil
}

func ListPubKeyFromListSenderKey(threePaymentAddress []string) ([][]byte, error) {
	pubKeys := make([][]byte, len(threePaymentAddress))
	for i := 0; i < len(threePaymentAddress); i++ {
		paymentAddress, err := GetPaymentAddressFromSenderKeyParams(threePaymentAddress[i])
		if err != nil {
			return nil, err
		}
		pubKeys[i] = paymentAddress.Pk
	}
	return pubKeys, nil
}

func ListPaymentAddressFromListSenderKey(listSenderKey []string) []privacy.PaymentAddress {
	paymentAddresses := make([]privacy.PaymentAddress, 0)
	for i := 0; i < 3; i++ {
		new, _ := GetPaymentAddressFromSenderKeyParams(listSenderKey[i])
		paymentAddresses = append(paymentAddresses, *new)
	}
	return paymentAddresses
}

func NewNormalVoteProposalFromOwnerMetadataFromRPC(data map[string]interface{}) (Metadata, error) {
	boardType := common.NewBoardTypeFromString(data["BoardType"].(string))
	voteProposalData := NewVoteProposalDataFromJson(data["VoteProposalData"])
	paymentAddresses := data["PaymentAddresses"].([]privacy.PaymentAddress)
	lv3TxID := data["Lv3TxID"].(common.Hash)
	var meta Metadata
	if boardType == common.DCBBoard {
		meta = NewNormalDCBVoteProposalFromOwnerMetadata(
			*voteProposalData,
			paymentAddresses,
			lv3TxID,
		)
	} else {
		meta = NewNormalGOVVoteProposalFromOwnerMetadata(
			*voteProposalData,
			paymentAddresses,
			lv3TxID,
		)
	}
	return meta, nil
}

func NewNormalVoteProposalFromSealerMetadataFromRPC(data map[string]interface{}) (Metadata, error) {
	boardType := common.NewBoardTypeFromString(data["BoardType"].(string))
	voteProposalData := NewVoteProposalDataFromJson(data["VoteProposalData"])
	paymentAddresses := data["PaymentAddresses"].([]privacy.PaymentAddress)
	lv1TxID := data["Lv1TxID"].(common.Hash)
	lv3TxID := data["Lv3TxID"].(common.Hash)
	var meta Metadata
	if boardType == common.DCBBoard {
		meta = NewNormalDCBVoteProposalFromSealerMetadata(
			*voteProposalData,
			paymentAddresses,
			lv1TxID,
			lv3TxID,
		)
	} else {
		meta = NewNormalGOVVoteProposalFromSealerMetadata(
			*voteProposalData,
			paymentAddresses,
			lv1TxID,
			lv3TxID,
		)
	}
	return meta, nil
}
