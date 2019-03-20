package metadata

import (
	"github.com/constant-money/constant-chain/common"
	"github.com/constant-money/constant-chain/database"
	"github.com/constant-money/constant-chain/privacy"
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
	record := rewardProposalWinnerMetadata.PaymentAddress.String()
	record += common.Uint32ToString(rewardProposalWinnerMetadata.Prize)
	record += rewardProposalWinnerMetadata.MetadataBase.Hash().String()
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

func (rewardProposalWinnerMetadata *RewardProposalWinnerMetadata) CalculateSize() uint64 {
	return calculateSize(rewardProposalWinnerMetadata)
}
