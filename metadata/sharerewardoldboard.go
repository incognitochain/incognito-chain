package metadata

import (
	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/database"
	"github.com/ninjadotorg/constant/privacy"
)

type ShareRewardOldBoardMetadata struct {
	chairPaymentAddress privacy.PaymentAddress
	voterPaymentAddress privacy.PaymentAddress

	MetadataBase
}

func NewShareRewardOldBoardMetadata(
	candidatePaymentAddress privacy.PaymentAddress,
	voterPaymentAddress privacy.PaymentAddress,
	boardType BoardType,
) *ShareRewardOldBoardMetadata {
	metadataType := 0
	if boardType == DCBBoard {
		metadataType = ShareRewardOldDCBBoardMeta
	} else {
		metadataType = ShareRewardOldGOVBoardMeta
	}

	return &ShareRewardOldBoardMetadata{
		chairPaymentAddress: candidatePaymentAddress,
		voterPaymentAddress: voterPaymentAddress,
		MetadataBase: MetadataBase{
			Type: metadataType,
		},
	}
}

func (rewardShareOldBoardMetadata *ShareRewardOldBoardMetadata) Hash() *common.Hash {
	record := rewardShareOldBoardMetadata.voterPaymentAddress.String()
	record += rewardShareOldBoardMetadata.chairPaymentAddress.String()
	record += rewardShareOldBoardMetadata.MetadataBase.Hash().String()
	hash := common.DoubleHashH([]byte(record))
	return &hash
}

func (rewardShareOldBoardMetadata *ShareRewardOldBoardMetadata) ValidateTxWithBlockChain(tx Transaction, bcr BlockchainRetriever, b byte, db database.DatabaseInterface) (bool, error) {
	return true, nil
}

func (rewardShareOldBoardMetadata *ShareRewardOldBoardMetadata) ValidateSanityData(bcr BlockchainRetriever, tx Transaction) (bool, bool, error) {
	return true, true, nil
}

func (rewardShareOldBoardMetadata *ShareRewardOldBoardMetadata) ValidateMetadataByItself() bool {
	return true
}
