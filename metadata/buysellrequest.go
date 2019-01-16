package metadata

import (
	"bytes"
	"errors"
	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/database"
	"github.com/ninjadotorg/constant/privacy"
)

type BuySellRequest struct {
	PaymentAddress privacy.PaymentAddress
	TokenID        common.Hash
	Amount         uint64
	BuyPrice       uint64 // in Constant unit

	SaleID []byte // only when requesting to DCB

	MetadataBase
}

func NewBuySellRequest(
	paymentAddress privacy.PaymentAddress,
	tokenID common.Hash,
	amount uint64,
	buyPrice uint64,
	metaType int,
) *BuySellRequest {
	metadataBase := MetadataBase{
		Type: metaType,
	}
	result := &BuySellRequest{
		PaymentAddress: paymentAddress,
		Amount:         amount,
		BuyPrice:       buyPrice,
		MetadataBase:   metadataBase,
		TokenID:        tokenID,
	}
	return result
}

func (bsReq *BuySellRequest) ValidateTxWithBlockChain(txr Transaction, bcr BlockchainRetriever, chainID byte, db database.DatabaseInterface) (bool, error) {

	// TODO: support and validate for either bonds or govs buy requests

	govParams := bcr.GetGOVParams()
	sellingBondsParams := govParams.SellingBonds
	if sellingBondsParams == nil {
		return common.FalseValue, errors.New("SellingBonds params are not existed.")
	}

	bondID := sellingBondsParams.GetID()
	if !bytes.Equal(bondID[:], bsReq.TokenID[:]) {
		return common.FalseValue, errors.New("Requested tokenID has not been selling yet.")
	}

	// check if buy price againsts SellingBonds params' BondPrice is correct or not
	if bsReq.BuyPrice < sellingBondsParams.BondPrice {
		return common.FalseValue, errors.New("Requested buy price is under SellingBonds params' buy price.")
	}
	return common.TrueValue, nil
}

func (bsReq *BuySellRequest) ValidateSanityData(bcr BlockchainRetriever, txr Transaction) (bool, bool, error) {
	if len(bsReq.PaymentAddress.Pk) == 0 {
		return common.FalseValue, common.FalseValue, errors.New("Wrong request info's payment address")
	}
	if len(bsReq.PaymentAddress.Tk) == 0 {
		return common.FalseValue, common.FalseValue, errors.New("Wrong request info's payment address")
	}
	if bsReq.BuyPrice == 0 {
		return common.FalseValue, common.FalseValue, errors.New("Wrong request info's buy price")
	}
	if bsReq.Amount == 0 {
		return common.FalseValue, common.FalseValue, errors.New("Wrong request info's amount")
	}
	if len(bsReq.TokenID) != common.HashSize {
		return common.FalseValue, common.FalseValue, errors.New("Wrong request info's asset type")
	}
	return common.TrueValue, common.TrueValue, nil
}

func (bsReq *BuySellRequest) ValidateMetadataByItself() bool {
	// The validation just need to check at tx level, so returning common.TrueValue here
	return common.TrueValue
}

func (bsReq *BuySellRequest) Hash() *common.Hash {
	record := string(bsReq.PaymentAddress.Bytes())
	record += bsReq.TokenID.String()
	record += string(bsReq.Amount)
	record += string(bsReq.BuyPrice)
	record += string(bsReq.SaleID)
	record += bsReq.MetadataBase.Hash().String()

	// final hash
	hash := common.DoubleHashH([]byte(record))
	return &hash
}
