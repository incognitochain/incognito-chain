package metadata

import (
	"errors"

	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/database"
	"github.com/ninjadotorg/constant/privacy"
)

type BuyBackRequest struct {
	PaymentAddress privacy.PaymentAddress
	Amount         uint64
	TokenID        common.Hash
	MetadataBase
}

func NewBuyBackRequest(
	paymentAddress privacy.PaymentAddress,
	amount uint64,
	tokenID common.Hash,
	metaType int,
) *BuyBackRequest {
	metadataBase := MetadataBase{
		Type: metaType,
	}
	return &BuyBackRequest{
		PaymentAddress: paymentAddress,
		Amount:         amount,
		TokenID:        tokenID,
		MetadataBase:   metadataBase,
	}
}

func (bbReq *BuyBackRequest) ValidateTxWithBlockChain(
	txr Transaction,
	bcr BlockchainRetriever,
	chainID byte,
	db database.DatabaseInterface,
) (bool, error) {

	// TODO: need to check vin's amt and burning address in vout
	return common.TrueValue, nil
}

func (bbReq *BuyBackRequest) ValidateSanityData(
	bcr BlockchainRetriever,
	txr Transaction,
) (bool, bool, error) {
	if len(bbReq.PaymentAddress.Pk) == 0 {
		return common.FalseValue, common.FalseValue, errors.New("Wrong request info's payment address")
	}
	if len(bbReq.PaymentAddress.Tk) == 0 {
		return common.FalseValue, common.FalseValue, errors.New("Wrong request info's payment address")
	}
	if bbReq.Amount == 0 {
		return common.FalseValue, common.FalseValue, errors.New("Wrong request info's amount")
	}
	if len(bbReq.TokenID) != common.HashSize {
		return common.FalseValue, common.FalseValue, errors.New("Wrong request info's token id")
	}
	return common.TrueValue, common.TrueValue, nil
}

func (bbReq *BuyBackRequest) ValidateMetadataByItself() bool {
	// The validation just need to check at tx level, so returning common.TrueValue here
	return common.TrueValue
}

func (bbReq *BuyBackRequest) Hash() *common.Hash {
	record := string(bbReq.PaymentAddress.Bytes())
	record += string(bbReq.Amount)
	record += bbReq.TokenID.String()
	record += bbReq.MetadataBase.Hash().String()
	hash := common.DoubleHashH([]byte(record))
	return &hash
}
