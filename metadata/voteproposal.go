package metadata

import (
	"github.com/constant-money/constant-chain/blockchain/component"
	"github.com/constant-money/constant-chain/wallet"
	
	"github.com/constant-money/constant-chain/common"
	"github.com/constant-money/constant-chain/privacy"
)

func NewVoteProposalData(proposalTxID common.Hash, constitutionIndex uint32, voterPayment privacy.PaymentAddress) *component.VoteProposalData {
	return &component.VoteProposalData{ProposalTxID: proposalTxID, ConstitutionIndex: constitutionIndex, VoterPayment: voterPayment}
}

func NewVoteProposalDataFromJson(data interface{}) *component.VoteProposalData {
	voteProposalDataData := data.(map[string]interface{})
	proposalTxIDData := voteProposalDataData["ProposalTxID"].(string)
	proposalTxID, _ := common.NewHashFromStr(proposalTxIDData)
	constitutionIndex := uint32(voteProposalDataData["ConstitutionIndex"].(float64))
	voterPayment, err:= component.NewPaymentAddressFromString(voteProposalDataData["VoterPayment"].(string))
	if err != nil {
		panic(err)
	}
	return NewVoteProposalData(
		*proposalTxID,
		constitutionIndex,
		*voterPayment,
	)
}

func GetPaymentAddressFromSenderKeyParams(keyParam string) (*privacy.PaymentAddress, error) {
	keyWallet, err := wallet.Base58CheckDeserialize(keyParam)
	if err != nil {
		return nil, err
	}
	return &keyWallet.KeySet.PaymentAddress, nil
}

func NewVoteProposalMetadataFromRPC(data map[string]interface{}) (Metadata, error) {
	boardType := common.NewBoardTypeFromString(data["BoardType"].(string))
	voteProposalData := NewVoteProposalDataFromJson(data["VoteProposalData"])
	var meta Metadata
	if boardType == common.DCBBoard {
		meta = NewDCBVoteProposalMetadata(
			*voteProposalData,
		)
	} else {
		meta = NewGOVVoteProposalMetadata(
			*voteProposalData,
		)
	}
	return meta, nil
}
