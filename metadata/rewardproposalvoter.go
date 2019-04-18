package metadata

import (
	"github.com/constant-money/constant-chain/common"
	"github.com/constant-money/constant-chain/database"
)

//validate by checking vout address of this tx equal to vin address of winning proposal
type RewardDCBProposalVoterMetadata struct {
	MetadataBase
}

func (rewardDCBProposalVoterMetadata *RewardDCBProposalVoterMetadata) ProcessWhenInsertBlockShard(tx Transaction, bcr BlockchainRetriever) error {
	// bcr.UpdateDCBFund(tx)
	return nil
}

func NewRewardDCBProposalVoterMetadata() *RewardDCBProposalVoterMetadata {
	return &RewardDCBProposalVoterMetadata{
		MetadataBase: *NewMetadataBase(RewardDCBProposalVoterMeta),
	}
}

func (rewardDCBProposalVoterMetadata *RewardDCBProposalVoterMetadata) Hash() *common.Hash {
	record := rewardDCBProposalVoterMetadata.MetadataBase.Hash().String()
	hash := common.HashH([]byte(record))
	return &hash
}

func (rewardDCBProposalVoterMetadata *RewardDCBProposalVoterMetadata) ValidateTxWithBlockChain(tx Transaction, bcr BlockchainRetriever, b byte, db database.DatabaseInterface) (bool, error) {
	return true, nil
}

func (rewardDCBProposalVoterMetadata *RewardDCBProposalVoterMetadata) ValidateSanityData(bcr BlockchainRetriever, tx Transaction) (bool, bool, error) {
	return true, true, nil
}

func (rewardDCBProposalVoterMetadata *RewardDCBProposalVoterMetadata) ValidateMetadataByItself() bool {
	return true
}

type RewardGOVProposalVoterMetadata struct {
	MetadataBase
}

func (rewardGOVProposalVoterMetadata *RewardGOVProposalVoterMetadata) ProcessWhenInsertBlockShard(tx Transaction, bcr BlockchainRetriever) error {
	// bcr.UpdateDCBFund(tx)
	return nil
}

func NewRewardGOVProposalVoterMetadata() *RewardGOVProposalVoterMetadata {
	return &RewardGOVProposalVoterMetadata{
		MetadataBase: *NewMetadataBase(RewardGOVProposalVoterMeta),
	}
}

func (rewardGOVProposalVoterMetadata *RewardGOVProposalVoterMetadata) Hash() *common.Hash {
	record := rewardGOVProposalVoterMetadata.MetadataBase.Hash().String()
	hash := common.HashH([]byte(record))
	return &hash
}

func (rewardGOVProposalVoterMetadata *RewardGOVProposalVoterMetadata) ValidateTxWithBlockChain(tx Transaction, bcr BlockchainRetriever, b byte, db database.DatabaseInterface) (bool, error) {
	return true, nil
}

func (rewardGOVProposalVoterMetadata *RewardGOVProposalVoterMetadata) ValidateSanityData(bcr BlockchainRetriever, tx Transaction) (bool, bool, error) {
	return true, true, nil
}

func (rewardGOVProposalVoterMetadata *RewardGOVProposalVoterMetadata) ValidateMetadataByItself() bool {
	return true
}

func (rewardGOVProposalVoterMetadata *RewardGOVProposalVoterMetadata) CalculateSize() uint64 {
	return calculateSize(rewardGOVProposalVoterMetadata)
}
