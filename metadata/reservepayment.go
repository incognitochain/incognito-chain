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

type ReservePayment struct {
	RequestedTxID *common.Hash
	SaleID        []byte

	MetadataBase
}

func NewReservePayment(rpayData map[string]interface{}) *ReservePayment {
	s, err := hex.DecodeString(rpayData["RequestedTxID"].(string))
	if err != nil {
		return nil
	}
	saleID, err := hex.DecodeString(rpayData["SaleId"].(string))
	if err != nil {
		return nil
	}
	result := &ReservePayment{
		RequestedTxID: &common.Hash{},
		SaleID:        saleID,
	}
	result.Type = ReservePaymentMeta
	copy(result.RequestedTxID[:], s)
	return result
}

func (rpay *ReservePayment) ValidateTxWithBlockChain(txr Transaction, bcr BlockchainRetriever, shardID byte, db database.DatabaseInterface) (bool, error) {
	// TODO: check if there's a corresponding request in the same block
	// Check if sale exists
	saleData, err := bcr.GetCrowdsaleData(rpay.SaleID)
	if err != nil {
		return false, err
	}

	// Check if sending address is DCB's
	keyWalletDCBAccount, _ := wallet.Base58CheckDeserialize(common.DCBAddress)
	if saleData.SellingAsset.IsEqual(&common.ConstantID) {
		if !bytes.Equal(txr.GetSigPubKey(), keyWalletDCBAccount.KeySet.PaymentAddress.Pk[:]) {
			return false, fmt.Errorf("Crowdsale payment must send Constant from DCB address")
		}
	} else if bytes.Equal(saleData.SellingAsset[:8], common.BondTokenID[:8]) {
		// check double spending if selling bond
		return true, nil
	}

	// TODO(@0xbunyip): validate amount of asset sent
	return false, nil
}

func (rpay *ReservePayment) ValidateSanityData(bcr BlockchainRetriever, txr Transaction) (bool, bool, error) {
	ok, err := txr.ValidateSanityData(bcr)
	if err != nil || !ok {
		return false, ok, err
	}
	if len(rpay.SaleID) == 0 {
		return false, false, errors.New("Wrong request info's SaleID")
	}
	return false, true, nil
}

func (rpay *ReservePayment) ValidateMetadataByItself() bool {
	// The validation just need to check at tx level, so returning true here
	return true
}

func (rpay *ReservePayment) Hash() *common.Hash {
	record := rpay.RequestedTxID.String()
	record += string(rpay.SaleID)

	// final hash
	hash := common.DoubleHashH([]byte(record))
	return &hash
}

func (rpay *ReservePayment) CalculateSize() uint64 {
	return calculateSize(rpay)
}
