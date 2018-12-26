package metadata

import (
	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/database"
)

type RewardProposalWinnerMetadata struct {
	PubKey []byte
	Prize  uint32
	MetadataBase
}

func NewRewardProposalWinnerMetadata(pubKey []byte, prize uint32) *RewardProposalWinnerMetadata {
	return &RewardProposalWinnerMetadata{
		PubKey:       pubKey,
		Prize:        prize,
		MetadataBase: *NewMetadataBase(RewardProposalWinnerMeta),
	}
}

func (rewardProposalWinnerMetadata *RewardProposalWinnerMetadata) Hash() *common.Hash {
	record := string(rewardProposalWinnerMetadata.PubKey)
	record += common.Uint32ToString(rewardProposalWinnerMetadata.Prize)
	record += string(rewardProposalWinnerMetadata.MetadataBase.Hash().GetBytes())
	hash := common.DoubleHashH([]byte(record))
	return &hash
}

func (rewardProposalWinnerMetadata *RewardProposalWinnerMetadata) ValidateTxWithBlockChain(Transaction, BlockchainRetriever, byte, database.DatabaseInterface) (bool, error) {
	return true, nil
}

func (rewardProposalWinnerMetadata *RewardProposalWinnerMetadata) ValidateSanityData(BlockchainRetriever, Transaction) (bool, bool, error) {
	return true, true, nil
}

func (rewardProposalWinnerMetadata *RewardProposalWinnerMetadata) ValidateMetadataByItself() bool {
	return true
}
