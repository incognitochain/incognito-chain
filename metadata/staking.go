package metadata

import (
	"errors"

	"github.com/ninjadotorg/constant/database"
)

type StakingMetadata struct {
	*MetadataBase
}

func NewStakingMetadata(stakingType int) (*StakingMetadata, error) {
	if stakingType != ShardStakingMeta && stakingType != BeaconStakingMeta {
		return nil, errors.New("Invalid staking type")
	}
	metadataBase := NewMetadataBase(stakingType)

	return &StakingMetadata{metadataBase}, nil
}

func (sm *StakingMetadata) ValidateMetadataByItself() bool {
	return true
}
func (sm *StakingMetadata) ValidateTxWithBlockChain(tx Transaction, bcr BlockchainRetriever, b byte, db database.DatabaseInterface) (bool, error) {
	return true, nil
}

func (sm *StakingMetadata) ValidateSanityData(bcr BlockchainRetriever, tx Transaction) (bool, bool, error) {

	return true, true, nil
}
