package metadata

import (
	"encoding/base64"
	"encoding/json"
	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/database"
	"github.com/ninjadotorg/constant/privacy"
	"strconv"
)

type VoteGOVBoardMetadata struct {
	VoteBoardMetadata VoteBoardMetadata

	MetadataBase
}

func NewVoteGOVBoardMetadata(candidatePaymentAddress privacy.PaymentAddress) *VoteGOVBoardMetadata {
	return &VoteGOVBoardMetadata{
		VoteBoardMetadata: *NewVoteBoardMetadata(candidatePaymentAddress),
		MetadataBase:      *NewMetadataBase(VoteGOVBoardMeta),
	}
}

func (voteGOVBoardMetadata *VoteGOVBoardMetadata) ValidateTxWithBlockChain(txr Transaction, bcr BlockchainRetriever, shardID byte, db database.DatabaseInterface) (bool, error) {
	return true, nil
}

func (voteGOVBoardMetadata *VoteGOVBoardMetadata) GetType() int {
	return VoteGOVBoardMeta
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
