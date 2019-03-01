package metadata

import (
	"encoding/base64"
	"encoding/json"
	"github.com/ninjadotorg/constant/wallet"
	"strconv"

	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/database"
	"github.com/ninjadotorg/constant/privacy"
)

type VoteGOVBoardMetadata struct {
	VoteBoardMetadata VoteBoardMetadata

	MetadataBase
}

func (voteGOVBoardMetadata *VoteGOVBoardMetadata) ProcessWhenInsertBlockShard(tx Transaction, bcr BlockchainRetriever) error {
	boardType := GOVBoard
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
		voteGOVBoardMetadata.VoteBoardMetadata.CandidatePaymentAddress,
		voteAmount,
	)
	if err1 != nil {
		return err1
	}
	return nil
}

func NewVoteGOVBoardMetadata(candidatePaymentAddress privacy.PaymentAddress, BoardIndex uint32) *VoteGOVBoardMetadata {
	return &VoteGOVBoardMetadata{
		VoteBoardMetadata: *NewVoteBoardMetadata(candidatePaymentAddress, BoardIndex),
		MetadataBase:      *NewMetadataBase(VoteGOVBoardMeta),
	}
}

func NewVoteGOVBoardMetadataFromRPC(data map[string]interface{}) (Metadata, error) {
	paymentAddress := data["PaymentAddress"].(string)
	boardIndex := uint32(data["BoardIndex"].(float64))
	account, _ := wallet.Base58CheckDeserialize(paymentAddress)
	meta := NewVoteGOVBoardMetadata(account.KeySet.PaymentAddress, boardIndex)
	return meta, nil
}

func (voteGOVBoardMetadata *VoteGOVBoardMetadata) ValidateTxWithBlockChain(txr Transaction, bcr BlockchainRetriever, shardID byte, db database.DatabaseInterface) (bool, error) {
	return true, nil
}

func (voteGOVBoardMetadata *VoteGOVBoardMetadata) Hash() *common.Hash {
	record := string(voteGOVBoardMetadata.VoteBoardMetadata.GetBytes())
	record += voteGOVBoardMetadata.MetadataBase.Hash().String()
	hash := common.DoubleHashH([]byte(record))
	return &hash
}

func (voteGOVBoardMetadata *VoteGOVBoardMetadata) ValidateSanityData(bcr BlockchainRetriever, tx Transaction) (bool, bool, error) {
	return true, true, nil
}

func (voteGOVBoardMetadata *VoteGOVBoardMetadata) ValidateMetadataByItself() bool {
	return true
}

func (voteGOVBoardMetadata *VoteGOVBoardMetadata) BuildReqActions(tx Transaction, bcr BlockchainRetriever, shardID byte) ([][]string, error) {
	actionContent := map[string]interface{}{
		"reqTx": tx,
	}
	actionContentBytes, err := json.Marshal(actionContent)
	if err != nil {
		return [][]string{}, err
	}
	actionContentBase64Str := base64.StdEncoding.EncodeToString(actionContentBytes)
	action := []string{strconv.Itoa(voteGOVBoardMetadata.Type), actionContentBase64Str}
	return [][]string{action}, nil
}
