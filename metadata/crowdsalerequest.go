package metadata

import (
	"bytes"
	"errors"

	"github.com/ninjadotorg/constant/common"
	privacy "github.com/ninjadotorg/constant/privacy-protocol"
	"github.com/ninjadotorg/constant/wallet"
)

type CrowdsaleRequest struct {
	PaymentAddress privacy.PaymentAddress
	Amount         uint64
	SaleID         []byte // only when requesting to DCB

	MetadataBase
}

func NewCrowdsaleRequest(csReqData map[string]interface{}) *CrowdsaleRequest {
	return &CrowdsaleRequest{
		PaymentAddress: csReqData["paymentAddress"].(privacy.PaymentAddress),
		Amount:         uint64(csReqData["amount"].(float64)),
		SaleID:         csReqData["saleId"].([]byte),
	}
}

func (csReq *CrowdsaleRequest) ValidateTxWithBlockChain(txr Transaction, bcr BlockchainRetriever, chainID byte) (bool, error) {
	// check double spending on fee + buy/sell amount tx
	err := txr.ValidateConstDoubleSpendWithBlockchain(bcr, chainID)
	if err != nil {
		return false, err
	}

	// Check if Payment address is DCB's
	accountDCB, _ := wallet.Base58CheckDeserialize(common.DCBAddress)
	if !bytes.Equal(csReq.PaymentAddress.Pk[:], accountDCB.KeySet.PaymentAddress.Pk[:]) || !bytes.Equal(csReq.PaymentAddress.Tk[:], accountDCB.KeySet.PaymentAddress.Tk[:]) {
		return false, err
	}

	// Check if sale exists and ongoing
	saleData, err := bcr.GetCrowdsaleData(csReq.SaleID)
	if err != nil {
		return false, err
	}
	if saleData.EndBlock >= bcr.GetHeight() {
		return false, err
	}
	return false, nil
}

func (csReq *CrowdsaleRequest) ValidateSanityData(bcr BlockchainRetriever, txr Transaction) (bool, bool, error) {
	ok, err := txr.ValidateSanityData(bcr)
	if err != nil || !ok {
		return false, ok, err
	}
	if len(csReq.PaymentAddress.Pk) == 0 {
		return false, false, errors.New("Wrong request info's payment address")
	}
	if csReq.Amount == 0 {
		return false, false, errors.New("Wrong request info's amount")
	}
	return false, true, nil
}

func (csReq *CrowdsaleRequest) ValidateMetadataByItself() bool {
	// The validation just need to check at tx level, so returning true here
	return true
}

func (csReq *CrowdsaleRequest) Hash() *common.Hash {
	record := string(csReq.PaymentAddress.ToBytes())
	record += string(csReq.Amount)
	record += string(csReq.SaleID)

	// final hash
	hash := common.DoubleHashH([]byte(record))
	return &hash
}
