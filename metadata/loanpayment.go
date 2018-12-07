package metadata

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"strconv"

	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/wallet"
)

type LoanPayment struct {
	LoanID       []byte
	PayPrinciple bool
	MetadataBase
}

func NewLoanPayment(data map[string]interface{}) *LoanPayment {
	result := LoanPayment{}
	s, _ := hex.DecodeString(data["LoanID"].(string))
	result.LoanID = s
	result.PayPrinciple = data["PayPrinciple"].(bool)
	return &result
}

func (lp *LoanPayment) GetType() int {
	return LoanPaymentMeta
}

func (lp *LoanPayment) Hash() *common.Hash {
	record := string(lp.LoanID)
	record += string(strconv.FormatBool(lp.PayPrinciple))

	// final hash
	hash := common.DoubleHashH([]byte(record))
	return &hash
}

func (lp *LoanPayment) ValidateTxWithBlockChain(txr Transaction, bcr BlockchainRetriever, chainID byte) (bool, error) {
	receivers, _ := txr.GetReceivers()
	if len(receivers) == 0 {
		return false, fmt.Errorf("Loan payment must be sent to DCB address")
	}

	accountDCB, _ := wallet.Base58CheckDeserialize(common.DCBAddress)
	dcbPk := accountDCB.KeySet.PaymentAddress.Pk
	sender := txr.GetJSPubKey()
	for _, receiver := range receivers {
		if !bytes.Equal(receiver, sender) && !bytes.Equal(receiver, dcbPk) {
			return false, fmt.Errorf("Loan payment can only be sent to DCB address")
		}
	}

	// Check if payment amount is correct
	requestMeta, err := bcr.GetLoanRequestMeta(lp.LoanID)
	if err != nil {
		return false, err
	}
	if requestMeta == nil {
		return false, fmt.Errorf("Found no loan request for this loan payment")
	}
	_, _, deadline, err := bcr.GetLoanPayment(lp.LoanID)
	if err != nil {
		return false, err
	}
	if lp.PayPrinciple && uint32(bcr.GetHeight())+requestMeta.Params.Maturity >= deadline {
		return false, fmt.Errorf("Interest must be fully paid before paying principle")
	}
	return true, nil
}

func (lp *LoanPayment) ValidateSanityData(bcr BlockchainRetriever, txr Transaction) (bool, bool, error) {
	return true, true, nil // continue checking for fee
}

func (lp *LoanPayment) ValidateMetadataByItself() bool {
	return true
}
