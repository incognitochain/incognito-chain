package metadata

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/constant-money/constant-chain/common"
	"github.com/constant-money/constant-chain/database"
	"github.com/constant-money/constant-chain/wallet"
)

type CrowdsalePayment struct {
	SaleID []byte

	MetadataBase
}

func (csRes *CrowdsalePayment) ValidateTxWithBlockChain(txr Transaction, bcr BlockchainRetriever, shardID byte, db database.DatabaseInterface) (bool, error) {
	// Check if sale exists
	sale, err := bcr.GetSaleData(csRes.SaleID) // okay to use unsynced data since we only use immutable fields
	if err != nil {
		return false, err
	}

	// Check if sending address is DCB's
	keyWalletDCBAccount, _ := wallet.Base58CheckDeserialize(common.DCBAddress)
	if !bytes.Equal(txr.GetSigPubKey(), keyWalletDCBAccount.KeySet.PaymentAddress.Pk[:]) {
		return false, fmt.Errorf("Crowdsale payment must send asset from DCB address")
	}

	// TODO(@0xbunyip): check double spending for coinbase CST tx?
	if !sale.Buy {
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
	hash := common.HashH([]byte(record))
	return &hash
}

func (csRes *CrowdsalePayment) CalculateSize() uint64 {
	return calculateSize(csRes)
}
