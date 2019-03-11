package metadata

import (
	"github.com/big0t/constant-chain/common"
	"github.com/big0t/constant-chain/database"
)

//validate by checking vout address of this tx equal to vin address of winning proposal
type RewardDCBProposalSubmitterMetadata struct {
	MetadataBase
}

func (rewardDCBProposalSubmitterMetadata *RewardDCBProposalSubmitterMetadata) ProcessWhenInsertBlockShard(tx Transaction, bcr BlockchainRetriever) error {
	bcr.UpdateDCBFund(tx)
	return nil
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
	return true, nil
}

func (rewardDCBProposalSubmitterMetadata *RewardDCBProposalSubmitterMetadata) ValidateSanityData(bcr BlockchainRetriever, tx Transaction) (bool, bool, error) {
	return true, true, nil
}

func (rewardDCBProposalSubmitterMetadata *RewardDCBProposalSubmitterMetadata) ValidateMetadataByItself() bool {
	return true
}

type RewardGOVProposalSubmitterMetadata struct {
	MetadataBase
}

func (rewardGOVProposalSubmitterMetadata *RewardGOVProposalSubmitterMetadata) ProcessWhenInsertBlockShard(tx Transaction, bcr BlockchainRetriever) error {
	bcr.UpdateDCBFund(tx)
	return nil
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
	return true, nil
}

func (rewardGOVProposalSubmitterMetadata *RewardGOVProposalSubmitterMetadata) ValidateSanityData(bcr BlockchainRetriever, tx Transaction) (bool, bool, error) {
	return true, true, nil
}

func (rewardGOVProposalSubmitterMetadata *RewardGOVProposalSubmitterMetadata) ValidateMetadataByItself() bool {
	return true
}

func (rewardGOVProposalSubmitterMetadata *RewardGOVProposalSubmitterMetadata) CalculateSize() uint64 {
	return calculateSize(rewardGOVProposalSubmitterMetadata)
}
