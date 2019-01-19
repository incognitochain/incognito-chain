package metadata

import (
	"bytes"
	"encoding/hex"

	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/database"
	"github.com/ninjadotorg/constant/privacy"
	"github.com/ninjadotorg/constant/wallet"
	"github.com/pkg/errors"
)

// CrowdsaleRequest represents a buying request created by user to send to DCB
type CrowdsaleRequest struct {
	PaymentAddress privacy.PaymentAddress
	SaleID         []byte

	PriceLimit uint64 // max buy price set by user
	ValidUntil uint64
	MetadataBase
}

func NewCrowdsaleRequest(csReqData map[string]interface{}) (*CrowdsaleRequest, error) {
	errSaver := &ErrorSaver{}
	saleID, errSale := hex.DecodeString(csReqData["SaleId"].(string))
	priceLimit, okPrice := csReqData["PriceLimit"].(float64)
	validUntil, okValid := csReqData["ValidUntil"].(float64)
	paymentAddressStr, okAddr := csReqData["PaymentAddress"].(string)
	keyWallet, errPayment := wallet.Base58CheckDeserialize(paymentAddressStr)

	if !okPrice || !okValid || !okAddr {
		return nil, errors.Errorf("Error parsing crowdsale request data")
	}
	if errSaver.Save(errSale, errPayment) != nil {
		return nil, errSaver.Get()
	}

	result := &CrowdsaleRequest{
		PaymentAddress: keyWallet.KeySet.PaymentAddress,
		SaleID:         saleID,
		PriceLimit:     uint64(priceLimit),
		ValidUntil:     uint64(validUntil),
	}
	result.Type = CrowdsaleRequestMeta
	return result, nil
}

func (csReq *CrowdsaleRequest) ValidateTxWithBlockChain(txr Transaction, bcr BlockchainRetriever, chainID byte, db database.DatabaseInterface) (bool, error) {
	// Check if sale exists and ongoing
	saleData, err := bcr.GetCrowdsaleData(csReq.SaleID)
	if err != nil {
		return false, err
	}
	// TODO(@0xbunyip): get height of beacon chain on new consensus
	height, err := bcr.GetTxChainHeight(txr)
	if err != nil || saleData.EndBlock >= height {
		return false, errors.Errorf("Crowdsale ended")
	}

	// Check if request is still valid
	if height >= csReq.ValidUntil {
		return false, errors.Errorf("Crowdsale request is not valid anymore")
	}

	// Check if Payment address is DCB's
	keyWalletDCBAccount, _ := wallet.Base58CheckDeserialize(common.DCBAddress)
	if !bytes.Equal(csReq.PaymentAddress.Pk[:], keyWalletDCBAccount.KeySet.PaymentAddress.Pk[:]) || !bytes.Equal(csReq.PaymentAddress.Tk[:], keyWalletDCBAccount.KeySet.PaymentAddress.Tk[:]) {
		return false, errors.Errorf("Crowdsale request must send CST to DCBAddress")
	}
	return true, nil
}

func (csReq *CrowdsaleRequest) ValidateSanityData(bcr BlockchainRetriever, txr Transaction) (bool, bool, error) {
	if len(csReq.PaymentAddress.Pk) == 0 {
		return false, false, errors.New("Wrong request info's payment address")
	}
	return false, true, nil
}

func (csReq *CrowdsaleRequest) ValidateMetadataByItself() bool {
	// The validation just need to check at tx level, so returning true here
	// TODO(@0xbunyip): accept only some pairs of assets
	return true
}

func (csReq *CrowdsaleRequest) Hash() *common.Hash {
	record := csReq.PaymentAddress.String()
	record += string(csReq.SaleID)
	record += string(csReq.PriceLimit)
	record += string(csReq.ValidUntil)

	// final hash
	record += csReq.MetadataBase.Hash().String()
	hash := common.DoubleHashH([]byte(record))
	return &hash
}
