package metadata

import (
	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/database"
	"github.com/ninjadotorg/constant/metadata/fromshardins"
	"github.com/ninjadotorg/constant/privacy"
	"github.com/ninjadotorg/constant/wallet"
)

type VoteDCBBoardMetadata struct {
	VoteBoardMetadata VoteBoardMetadata

	MetadataBase
}

type GovernorInterface interface {
	GetBoardIndex() uint32
}

func (voteDCBBoardMetadata *VoteDCBBoardMetadata) ProcessWhenInsertBlockShard(tx Transaction, bcr BlockchainRetriever) error {
	boardType := common.DCBBoard
	voteAmount, err := tx.GetAmountOfVote()
	if err != nil {
		return err
	}
	payment, err := tx.GetVoterPaymentAddress()
	if err != nil {
		return err
	}
	governor := bcr.GetGovernor(boardType)
	boardIndex := governor.GetBoardIndex() + 1
	err1 := bcr.GetDatabase().AddVoteBoard(
		boardType,
		boardIndex,
		*payment,
		voteDCBBoardMetadata.VoteBoardMetadata.CandidatePaymentAddress,
		voteAmount,
	)
	if err1 != nil {
		return err1
	}
	return nil
}

func NewVoteDCBBoardMetadata(candidatePaymentAddress privacy.PaymentAddress, boardIndex uint32) *VoteDCBBoardMetadata {
	return &VoteDCBBoardMetadata{
		VoteBoardMetadata: *NewVoteBoardMetadata(candidatePaymentAddress, boardIndex),
		MetadataBase:      *NewMetadataBase(VoteDCBBoardMeta),
	}
}

func NewVoteDCBBoardMetadataFromRPC(data map[string]interface{}) (Metadata, error) {
	paymentAddress := data["PaymentAddress"].(string)
	boardIndex := uint32(data["BoardIndex"].(float64))
	account, _ := wallet.Base58CheckDeserialize(paymentAddress)
	meta := NewVoteDCBBoardMetadata(account.KeySet.PaymentAddress, boardIndex)
	return meta, nil
}

func (voteDCBBoardMetadata *VoteDCBBoardMetadata) ValidateTxWithBlockChain(txr Transaction, bcr BlockchainRetriever, shardID byte, db database.DatabaseInterface) (bool, error) {
	return true, nil
}

func (voteDCBBoardMetadata *VoteDCBBoardMetadata) Hash() *common.Hash {
	record := string(voteDCBBoardMetadata.VoteBoardMetadata.GetBytes())
	record += voteDCBBoardMetadata.MetadataBase.Hash().String()
	hash := common.DoubleHashH([]byte(record))
	return &hash
}

func (voteDCBBoardMetadata *VoteDCBBoardMetadata) ValidateSanityData(bcr BlockchainRetriever, tx Transaction) (bool, bool, error) {
	return true, true, nil
}

func (voteDCBBoardMetadata *VoteDCBBoardMetadata) ValidateMetadataByItself() bool {
	return true
}

func (voteDCBBoardMetadata *VoteDCBBoardMetadata) BuildReqActions(tx Transaction, bcr BlockchainRetriever, shardID byte) ([][]string, error) {
	inst := fromshardins.NewVoteBoardIns(
		common.DCBBoard,
		voteDCBBoardMetadata.VoteBoardMetadata.CandidatePaymentAddress,
		voteDCBBoardMetadata.VoteBoardMetadata.BoardIndex,
	)
	instStr, err := inst.GetStringFormat()
	if err != nil {
		return nil, err
	}
	return [][]string{instStr}, nil
}
