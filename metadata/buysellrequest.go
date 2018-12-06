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

func (bsReq *BuySellRequest) CheckTransactionFee(tr Transaction, minFee uint64) bool {
	txFee := tr.GetTxFee()
	if txFee < minFee {
		return false
	}
	return true
}

func (bsReq *BuySellRequest) ValidateTxWithBlockChain(txr Transaction, bcr BlockchainRetriever, chainID byte) (bool, error) {
	return true, nil
}

func (bsReq *BuySellRequest) ValidateSanityData() (bool, bool, error) {
	return true, true, nil
}

func (bsReq *BuySellRequest) ValidateMetadataByItself() bool {
	return true
}

func (bsReq *BuySellRequest) GetType() int {
	return BuySellRequestMeta
}

func (bsReq *BuySellRequest) Hash() *common.Hash {
	return &common.Hash{}
}
