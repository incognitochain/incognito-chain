package metadata

import (
	"encoding/hex"
	"github.com/constant-money/constant-chain/blockchain/component"
	"github.com/constant-money/constant-chain/wallet"

	"github.com/constant-money/constant-chain/common"
	"github.com/constant-money/constant-chain/database"
	"github.com/constant-money/constant-chain/privacy"
)


func NewVoteProposalData(proposalTxID common.Hash, constitutionIndex uint32) *component.VoteProposalData {
	return &component.VoteProposalData{ProposalTxID: proposalTxID, ConstitutionIndex: constitutionIndex}
}

func NewVoteProposalDataFromJson(data interface{}) *component.VoteProposalData {
	voteProposalDataData := data.(map[string]interface{})

	proposalTxIDData, _ := hex.DecodeString(voteProposalDataData["ProposalTxID"].(string))
	proposalTxID, _ := common.NewHash(proposalTxIDData)
	constitutionIndex := uint32(voteProposalDataData["ConstitutionIndex"].(float64))
	return NewVoteProposalData(
		*proposalTxID,
		constitutionIndex,
	)
}

type NormalVoteProposalMetadata struct {
	VoteProposal             component.VoteProposalData
	LockerPaymentAddress     []privacy.PaymentAddress
	PointerToLv3VoteProposal common.Hash
}

func NewNormalVoteProposalMetadata(
	voteProposal component.VoteProposalData,
	lockerPaymentAddress []privacy.PaymentAddress,
	pointerToLv3VoteProposal common.Hash,
) *NormalVoteProposalMetadata {
	return &NormalVoteProposalMetadata{
		VoteProposal:             voteProposal,
		LockerPaymentAddress:     lockerPaymentAddress,
		PointerToLv3VoteProposal: pointerToLv3VoteProposal,
	}
}

func (normalVoteProposalMetadata *NormalVoteProposalMetadata) ValidateBeforeNewBlock(boardType common.BoardType, tx Transaction, bcr BlockchainRetriever, shardID byte) bool {
	endedPivot := bcr.GetConstitutionEndHeight(boardType, shardID)
	currentBlockHeight := bcr.GetCurrentBeaconBlockHeight(shardID) + 1
	lv1Pivot := endedPivot - common.EncryptionOnePhraseDuration
	return currentBlockHeight < endedPivot && currentBlockHeight >= lv1Pivot
}

func (normalVoteProposalMetadata *NormalVoteProposalMetadata) ToBytes() []byte {
	record := string(normalVoteProposalMetadata.VoteProposal.ToBytes())
	for _, i := range normalVoteProposalMetadata.LockerPaymentAddress {
		record += i.String()
	}
	record += string(normalVoteProposalMetadata.PointerToLv3VoteProposal.GetBytes())
	return []byte(record)
}

func (normalVoteProposalMetadata *NormalVoteProposalMetadata) ValidateSanityData(BlockchainRetriever, Transaction) (bool, bool, error) {
	return true, true, nil
}

func (normalVoteProposalMetadata *NormalVoteProposalMetadata) ValidateMetadataByItself() bool {
	for index1 := 0; index1 < len(normalVoteProposalMetadata.LockerPaymentAddress); index1++ {
		pub1 := normalVoteProposalMetadata.LockerPaymentAddress[index1]
		for index2 := index1 + 1; index2 < len(normalVoteProposalMetadata.LockerPaymentAddress); index2++ {
			pub2 := normalVoteProposalMetadata.LockerPaymentAddress[index2]
			if common.ByteEqual(pub1.Bytes(), pub2.Bytes()) {
				return false
			}
		}
	}
	return true
}

func (normalVoteProposalMetadata *NormalVoteProposalMetadata) ValidateTxWithBlockChain(
	boardType common.BoardType,
	transaction Transaction,
	bcr BlockchainRetriever,
	shardID byte,
	db database.DatabaseInterface) (bool,
	error) {
	boardPaymentAddress := bcr.GetBoardPaymentAddress(boardType)
	for _, j := range normalVoteProposalMetadata.LockerPaymentAddress {
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


func NewNormalVoteProposalMetadataFromRPC(data map[string]interface{}) (Metadata, error) {
	boardType := common.NewBoardTypeFromString(data["BoardType"].(string))
	voteProposalData := NewVoteProposalDataFromJson(data["VoteProposalData"])
	paymentAddresses := data["PaymentAddresses"].([]privacy.PaymentAddress)
	lv1TxID := data["Lv1TxID"].(common.Hash)
	var meta Metadata
	if boardType == common.DCBBoard {
		meta = NewNormalDCBVoteProposalMetadata(
			*voteProposalData,
			paymentAddresses,
			lv1TxID,
		)
	} else {
		meta = NewNormalGOVVoteProposalMetadata(
			*voteProposalData,
			paymentAddresses,
			lv1TxID,
		)
	}
	return meta, nil
}
