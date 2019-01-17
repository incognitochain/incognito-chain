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
	result.Type = CrowdSalePaymentMeta
	copy(result.RequestedTxID[:], s)
	return result
}

func (csRes *CrowdsalePayment) ValidateTxWithBlockChain(txr Transaction, bcr BlockchainRetriever, chainID byte, db database.DatabaseInterface) (bool, error) {
	// TODO: check if there's a corresponding request in the same block
	// Check if sale exists
	saleData, err := bcr.GetCrowdsaleData(csRes.SaleID)
	if err != nil {
		return common.FalseValue, err
	}

	// Check if sending address is DCB's
	keyWalletDCBAccount, _ := wallet.Base58CheckDeserialize(common.DCBAddress)
	if saleData.SellingAsset.IsEqual(&common.ConstantID) {
		if !bytes.Equal(txr.GetSigPubKey(), keyWalletDCBAccount.KeySet.PaymentAddress.Pk[:]) {
			return common.FalseValue, fmt.Errorf("Crowdsale payment must send Constant from DCB address")
		}
	} else if bytes.Equal(saleData.SellingAsset[:8], common.BondTokenID[:8]) {
		// check double spending if selling bond
		return common.TrueValue, nil
	}

	// TODO(@0xbunyip): validate amount of asset sent
	return common.FalseValue, nil
}

func (csRes *CrowdsalePayment) ValidateSanityData(bcr BlockchainRetriever, txr Transaction) (bool, bool, error) {
	ok, err := txr.ValidateSanityData(bcr)
	if err != nil || !ok {
		return common.FalseValue, ok, err
	}
	if len(csRes.SaleID) == 0 {
		return common.FalseValue, common.FalseValue, errors.New("Wrong request info's SaleID")
	}
	return common.FalseValue, common.TrueValue, nil
}

func (csRes *CrowdsalePayment) ValidateMetadataByItself() bool {
	// The validation just need to check at tx level, so returning common.TrueValue here
	return common.TrueValue
}

func (csRes *CrowdsalePayment) Hash() *common.Hash {
	record := csRes.RequestedTxID.String()
	record += string(csRes.SaleID)

	// final hash
	hash := common.DoubleHashH([]byte(record))
	return &hash
}
