package metadata

import (
	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/database"
)

type RewardShareOldBoardMetadata struct {
	candidatePubKey []byte
	voterPubKey     []byte

	MetadataBase
}

func NewRewardShareOldBoardMetadata(candidatePubKey []byte, voterPubKey []byte, boardType string) *RewardShareOldBoardMetadata {
	metadataType := 0
	if boardType == "dcb" {
		metadataType = RewardShareOldDCBBoardMeta
	} else {
		metadataType = RewardShareOldGOVBoardMeta
	}

	return &RewardShareOldBoardMetadata{
		candidatePubKey: candidatePubKey,
		voterPubKey:     voterPubKey,
		MetadataBase: MetadataBase{
			Type: metadataType,
		},
	}
}

func (rewardShareOldBoardMetadata *RewardShareOldBoardMetadata) Hash() *common.Hash {
	record := string(rewardShareOldBoardMetadata.voterPubKey)
	record += string(rewardShareOldBoardMetadata.candidatePubKey)
	record += string(rewardShareOldBoardMetadata.MetadataBase.Hash().GetBytes())
	hash := common.DoubleHashH([]byte(record))
	return &hash
}

func (rewardShareOldBoardMetadata *RewardShareOldBoardMetadata) ValidateTxWithBlockChain(tx Transaction, bcr BlockchainRetriever, b byte, db database.DatabaseInterface) (bool, error) {
	return true, nil
}

func (rewardShareOldBoardMetadata *RewardShareOldBoardMetadata) ValidateSanityData(bcr BlockchainRetriever, tx Transaction) (bool, bool, error) {
	return true, true, nil
}

func (rewardShareOldBoardMetadata *RewardShareOldBoardMetadata) ValidateMetadataByItself() bool {
	return true
}
