package metadata

import (
	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/database"
	"github.com/ninjadotorg/constant/privacy"
)

type RewardProposalWinnerMetadata struct {
	PaymentAddress privacy.PaymentAddress
	Prize          uint32
	MetadataBase
}

func NewRewardProposalWinnerMetadata(paymentAddress privacy.PaymentAddress, prize uint32) *RewardProposalWinnerMetadata {
	return &RewardProposalWinnerMetadata{
		PaymentAddress: paymentAddress,
		Prize:          prize,
		MetadataBase:   *NewMetadataBase(RewardProposalWinnerMeta),
	}
}

func (rewardProposalWinnerMetadata *RewardProposalWinnerMetadata) Hash() *common.Hash {
	record := string(rewardProposalWinnerMetadata.PaymentAddress.Bytes())
	record += common.Uint32ToString(rewardProposalWinnerMetadata.Prize)
	record += string(rewardProposalWinnerMetadata.MetadataBase.Hash().GetBytes())
	hash := common.DoubleHashH([]byte(record))
	return &hash
}

func (rewardProposalWinnerMetadata *RewardProposalWinnerMetadata) ValidateTxWithBlockChain(Transaction, BlockchainRetriever, byte, database.DatabaseInterface) (bool, error) {
	return common.TrueValue, nil
}

func (rewardProposalWinnerMetadata *RewardProposalWinnerMetadata) ValidateSanityData(BlockchainRetriever, Transaction) (bool, bool, error) {
	return common.TrueValue, common.TrueValue, nil
}

func (rewardProposalWinnerMetadata *RewardProposalWinnerMetadata) ValidateMetadataByItself() bool {
	return common.TrueValue
}
