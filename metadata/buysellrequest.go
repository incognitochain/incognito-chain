package metadata

import (
	"github.com/ninjadotorg/constant/common"
	privacy "github.com/ninjadotorg/constant/privacy-protocol"
)

type BuySellRequest struct {
	PaymentAddress privacy.PaymentAddress
	AssetType      common.Hash // token id (note: for bond, this one is just bond token id prefix)
	Amount         uint64
	BuyPrice       uint64 // in Constant unit

	SaleID []byte // only when requesting to DCB

	MetadataBase
}

func NewBuySellRequest(bsReqData map[string]interface{}) *BuySellRequest {
	return &BuySellRequest{
		PaymentAddress: bsReqData["paymentAddress"].(privacy.PaymentAddress),
		AssetType:      bsReqData["assetType"].(common.Hash),
		Amount:         uint64(bsReqData["amount"].(float64)),
		BuyPrice:       uint64(bsReqData["buyPrice"].(float64)),
		SaleID:         bsReqData["saleId"].([]byte),
	}
}
