package metadata

import (
	"github.com/constant-money/constant-chain/metadata/fromshardins"
	"github.com/constant-money/constant-chain/wallet"

	"github.com/constant-money/constant-chain/common"
	"github.com/constant-money/constant-chain/database"
	"github.com/constant-money/constant-chain/privacy"
)

type VoteGOVBoardMetadata struct {
	VoteBoardMetadata VoteBoardMetadata

	MetadataBase
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
	hash := common.HashH([]byte(record))
	return &hash
}

func (voteGOVBoardMetadata *VoteGOVBoardMetadata) ValidateSanityData(bcr BlockchainRetriever, tx Transaction) (bool, bool, error) {
	return true, true, nil
}

func (voteGOVBoardMetadata *VoteGOVBoardMetadata) ValidateMetadataByItself() bool {
	return true
}

func (voteGOVBoardMetadata *VoteGOVBoardMetadata) BuildReqActions(tx Transaction, bcr BlockchainRetriever, shardID byte) ([][]string, error) {
	voterPaymentAddress, err := tx.GetVoterPaymentAddress()
	if err != nil {
		return nil, err
	}
	amountOfVote, err := tx.GetAmountOfVote()
	if err != nil {
		return nil, err
	}
	inst := fromshardins.NewVoteBoardIns(
		common.GOVBoard,
		voteGOVBoardMetadata.VoteBoardMetadata.CandidatePaymentAddress,
		*voterPaymentAddress,
		voteGOVBoardMetadata.VoteBoardMetadata.BoardIndex,
		amountOfVote,
	)
	instStr, err := inst.GetStringFormat()
	if err != nil {
		return nil, err
	}
	return [][]string{instStr}, nil
}
