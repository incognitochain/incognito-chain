package metadata

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/database"
	"github.com/ninjadotorg/constant/wallet"
)

type CrowdsalePayment struct {
	RequestedTxID *common.Hash
	SaleID        []byte

	MetadataBase
}

func NewCrowdsalePayment(csResData map[string]interface{}) *CrowdsalePayment {
	s, err := hex.DecodeString(csResData["RequestedTxID"].(string))
	if err != nil {
		return nil
	}
	saleID, err := hex.DecodeString(csResData["SaleId"].(string))
	if err != nil {
		return nil
	}
	result := &CrowdsalePayment{
		RequestedTxID: &common.Hash{},
		SaleID:        saleID,
	}
	result.Type = CrowdsalePaymentMeta
	copy(result.RequestedTxID[:], s)
	return result
}

func (csRes *CrowdsalePayment) ValidateTxWithBlockChain(txr Transaction, bcr BlockchainRetriever, chainID byte, db database.DatabaseInterface) (bool, error) {
	// TODO: check if there's a corresponding request in the same block
	// Check if sale exists
	saleData, err := bcr.GetCrowdsaleData(csRes.SaleID)
	if err != nil {
		return false, err
	}

	// Check if sending address is DCB's
	accountDCB, _ := wallet.Base58CheckDeserialize(common.DCBAddress)
	if bytes.Equal(saleData.SellingAsset, common.ConstantID[:]) {
		if !bytes.Equal(txr.GetJSPubKey(), accountDCB.KeySet.PaymentAddress.Pk[:]) {
			return false, fmt.Errorf("Crowdsale payment must send Constant from DCB address")
		}
	} else if bytes.Equal(saleData.SellingAsset[:8], common.BondTokenID[:8]) {
		// check double spending if selling bond
		return true, nil
	}

	// TODO(@0xbunyip): validate amount of asset sent
	return false, nil
}

func (csRes *CrowdsalePayment) ValidateSanityData(bcr BlockchainRetriever, txr Transaction) (bool, bool, error) {
	ok, err := txr.ValidateSanityData(bcr)
	if err != nil || !ok {
		return false, ok, err
	}
	if len(csRes.SaleID) == 0 {
		return false, false, errors.New("Wrong request info's SaleID")
	}
	return false, true, nil
}

func (csRes *CrowdsalePayment) ValidateMetadataByItself() bool {
	// The validation just need to check at tx level, so returning true here
	return true
}

func (csRes *CrowdsalePayment) Hash() *common.Hash {
	record := string(csRes.RequestedTxID[:])
	record += string(csRes.SaleID)

	// final hash
	hash := common.DoubleHashH([]byte(record))
	return &hash
}
