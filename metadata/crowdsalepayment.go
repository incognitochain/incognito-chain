package metadata

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/big0t/constant-chain/common"
	"github.com/big0t/constant-chain/database"
	"github.com/big0t/constant-chain/wallet"
)

type CrowdsalePayment struct {
	SaleID []byte

	MetadataBase
}

func (csRes *CrowdsalePayment) ValidateTxWithBlockChain(txr Transaction, bcr BlockchainRetriever, shardID byte, db database.DatabaseInterface) (bool, error) {
	// Check if sale exists
	saleData, err := bcr.GetCrowdsaleData(csRes.SaleID)
	if err != nil {
		return false, err
	}

	// Check if sending address is DCB's
	keyWalletDCBAccount, _ := wallet.Base58CheckDeserialize(common.DCBAddress)
	if !bytes.Equal(txr.GetSigPubKey(), keyWalletDCBAccount.KeySet.PaymentAddress.Pk[:]) {
		return false, fmt.Errorf("Crowdsale payment must send asset from DCB address")
	}

	// TODO(@0xbunyip): check double spending for coinbase CST tx?
	if common.IsBondAsset(&saleData.SellingAsset) {
		// Check if sent from DCB address
		// check double spending if selling bond
		return true, nil
	}
	return false, nil
}

func (csRes *CrowdsalePayment) ValidateSanityData(bcr BlockchainRetriever, txr Transaction) (bool, bool, error) {
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
	record := string(csRes.SaleID)

	// final hash
	hash := common.DoubleHashH([]byte(record))
	return &hash
}

func (csRes *CrowdsalePayment) CalculateSize() uint64 {
	return calculateSize(csRes)
}
