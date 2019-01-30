package metadata

import (
	"bytes"
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

func (csRes *CrowdsalePayment) ValidateTxWithBlockChain(txr Transaction, bcr BlockchainRetriever, shardID byte, db database.DatabaseInterface) (bool, error) {
	// TODO(@0xbunyip): check if there's a corresponding request in the same block
	// Check if sale exists
	saleData, err := bcr.GetCrowdsaleData(csRes.SaleID)
	if err != nil {
		return false, err
	}

	// TODO(@0xbunyip): validate amount of asset sent and if price limit is not violated

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
	record := csRes.RequestedTxID.String()
	record += string(csRes.SaleID)

	// final hash
	hash := common.DoubleHashH([]byte(record))
	return &hash
}
