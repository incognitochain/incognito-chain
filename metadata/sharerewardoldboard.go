package metadata

import (
	"github.com/constant-money/constant-chain/common"
	"github.com/constant-money/constant-chain/database"
	"github.com/constant-money/constant-chain/privacy"
)

type ShareRewardOldBoardMetadata struct {
	ChairPaymentAddress privacy.PaymentAddress
	VoterPaymentAddress privacy.PaymentAddress

	MetadataBase
}

func NewShareRewardOldBoardMetadata(
	candidatePaymentAddress privacy.PaymentAddress,
	voterPaymentAddress privacy.PaymentAddress,
	boardType common.BoardType,
) *ShareRewardOldBoardMetadata {
	metadataType := 0
	if boardType == common.DCBBoard {
		metadataType = ShareRewardOldDCBBoardMeta
	} else {
		metadataType = ShareRewardOldGOVBoardMeta
	}

	return &ShareRewardOldBoardMetadata{
		ChairPaymentAddress: candidatePaymentAddress,
		VoterPaymentAddress: voterPaymentAddress,
		MetadataBase: MetadataBase{
			Type: metadataType,
		},
	}
}

func (rewardShareOldBoardMetadata *ShareRewardOldBoardMetadata) Hash() *common.Hash {
	record := rewardShareOldBoardMetadata.VoterPaymentAddress.String()
	record += rewardShareOldBoardMetadata.ChairPaymentAddress.String()
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

func (rewardShareOldBoardMetadata *ShareRewardOldBoardMetadata) CalculateSize() uint64 {
	return calculateSize(rewardShareOldBoardMetadata)
}
