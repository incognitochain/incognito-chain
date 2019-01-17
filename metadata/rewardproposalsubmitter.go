package metadata

import (
	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/database"
)

//validate by checking vout address of this tx equal to vin address of winning proposal
type RewardDCBProposalSubmitterMetadata struct {
	MetadataBase
}

func NewRewardDCBProposalSubmitterMetadata() *RewardDCBProposalSubmitterMetadata {
	return &RewardDCBProposalSubmitterMetadata{
		MetadataBase: *NewMetadataBase(RewardDCBProposalSubmitterMeta),
	}
}

func (rewardDCBProposalSubmitterMetadata *RewardDCBProposalSubmitterMetadata) Hash() *common.Hash {
	record := rewardDCBProposalSubmitterMetadata.MetadataBase.Hash().String()
	hash := common.DoubleHashH([]byte(record))
	return &hash
}

func (rewardDCBProposalSubmitterMetadata *RewardDCBProposalSubmitterMetadata) ValidateTxWithBlockChain(tx Transaction, bcr BlockchainRetriever, b byte, db database.DatabaseInterface) (bool, error) {
	return common.TrueValue, nil
}

func (rewardDCBProposalSubmitterMetadata *RewardDCBProposalSubmitterMetadata) ValidateSanityData(bcr BlockchainRetriever, tx Transaction) (bool, bool, error) {
	return common.TrueValue, common.TrueValue, nil
}

func (rewardDCBProposalSubmitterMetadata *RewardDCBProposalSubmitterMetadata) ValidateMetadataByItself() bool {
	return common.TrueValue
}

type RewardGOVProposalSubmitterMetadata struct {
	MetadataBase
}

func NewRewardGOVProposalSubmitterMetadata() *RewardGOVProposalSubmitterMetadata {
	return &RewardGOVProposalSubmitterMetadata{
		MetadataBase: *NewMetadataBase(RewardGOVProposalSubmitterMeta),
	}
}

func (rewardGOVProposalSubmitterMetadata *RewardGOVProposalSubmitterMetadata) Hash() *common.Hash {
	record := rewardGOVProposalSubmitterMetadata.MetadataBase.Hash().String()
	hash := common.DoubleHashH([]byte(record))
	return &hash
}

func (rewardGOVProposalSubmitterMetadata *RewardGOVProposalSubmitterMetadata) ValidateTxWithBlockChain(tx Transaction, bcr BlockchainRetriever, b byte, db database.DatabaseInterface) (bool, error) {
	return common.TrueValue, nil
}

func (rewardGOVProposalSubmitterMetadata *RewardGOVProposalSubmitterMetadata) ValidateSanityData(bcr BlockchainRetriever, tx Transaction) (bool, bool, error) {
	return common.TrueValue, common.TrueValue, nil
}

func (rewardGOVProposalSubmitterMetadata *RewardGOVProposalSubmitterMetadata) ValidateMetadataByItself() bool {
	return common.TrueValue
}
