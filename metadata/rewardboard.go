package metadata

import (
	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/database"
	"github.com/ninjadotorg/constant/privacy"
)

type RewardShareOldBoardMetadata struct {
	candidatePaymentAddress privacy.PaymentAddress
	voterPaymentAddress     privacy.PaymentAddress

	MetadataBase
}

func NewRewardShareOldBoardMetadata(
	candidatePaymentAddress privacy.PaymentAddress,
	voterPaymentAddress privacy.PaymentAddress,
	boardType byte,
) *RewardShareOldBoardMetadata {
	metadataType := 0
	if boardType == common.DCBBoard {
		metadataType = RewardShareOldDCBBoardMeta
	} else {
		metadataType = RewardShareOldGOVBoardMeta
	}

	return &RewardShareOldBoardMetadata{
		candidatePaymentAddress: candidatePaymentAddress,
		voterPaymentAddress:     voterPaymentAddress,
		MetadataBase: MetadataBase{
			Type: metadataType,
		},
	}
}

func (rewardShareOldBoardMetadata *RewardShareOldBoardMetadata) Hash() *common.Hash {
	record := rewardShareOldBoardMetadata.voterPaymentAddress.String()
	record += rewardShareOldBoardMetadata.candidatePaymentAddress.String()
	record += rewardShareOldBoardMetadata.MetadataBase.Hash().String()
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

func (rewardShareOldBoardMetadata *RewardShareOldBoardMetadata) CalculateSize() uint64 {
	return calculateSize(rewardShareOldBoardMetadata)
}
