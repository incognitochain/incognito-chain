package metadata

import (
	"encoding/base64"
	"encoding/json"
	"strconv"

	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/database"
	"github.com/ninjadotorg/constant/privacy"
)

type VoteDCBBoardMetadata struct {
	VoteBoardMetadata VoteBoardMetadata

	MetadataBase
}

type GovernorInterface interface {
	GetBoardIndex() uint32
}

func (voteDCBBoardMetadata *VoteDCBBoardMetadata) ProcessWhenInsertBlockShard(tx Transaction, bcr BlockchainRetriever) error {
	boardType := DCBBoard
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
		boardType.BoardTypeDB(),
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

func NewVoteDCBBoardMetadata(candidatePaymentAddress privacy.PaymentAddress) *VoteDCBBoardMetadata {
	return &VoteDCBBoardMetadata{
		VoteBoardMetadata: *NewVoteBoardMetadata(candidatePaymentAddress),
		MetadataBase:      *NewMetadataBase(VoteDCBBoardMeta),
	}
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
	actionContent := map[string]interface{}{
		"reqTx": tx,
	}
	actionContentBytes, err := json.Marshal(actionContent)
	if err != nil {
		return [][]string{}, err
	}
	actionContentBase64Str := base64.StdEncoding.EncodeToString(actionContentBytes)
	action := []string{strconv.Itoa(voteDCBBoardMetadata.Type), actionContentBase64Str}
	return [][]string{action}, nil
}
