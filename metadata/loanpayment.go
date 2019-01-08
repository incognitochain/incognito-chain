package metadata

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"strconv"

	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/database"
	"github.com/ninjadotorg/constant/wallet"
	"github.com/pkg/errors"
)

const Decimals = uint64(10000) // Each float number is multiplied by this value to store as uint64

type LoanPayment struct {
	LoanID       []byte
	PayPrinciple bool
	MetadataBase
}

func NewLoanPayment(data map[string]interface{}) (Metadata, error) {
	result := LoanPayment{}
	s, _ := hex.DecodeString(data["LoanID"].(string))
	result.LoanID = s
	result.PayPrinciple = data["PayPrinciple"].(bool)
	result.Type = LoanPaymentMeta
	return &result, nil
}

func (lp *LoanPayment) Hash() *common.Hash {
	record := string(lp.LoanID)
	record += string(strconv.FormatBool(lp.PayPrinciple))

	// final hash
	record += string(lp.MetadataBase.Hash()[:])
	hash := common.DoubleHashH([]byte(record))
	return &hash
}

func (lp *LoanPayment) ValidateTxWithBlockChain(txr Transaction, bcr BlockchainRetriever, chainID byte, db database.DatabaseInterface) (bool, error) {
	fmt.Println("Start validating LoanPayment tx with blockchain!!!")
	accountDCB, _ := wallet.Base58CheckDeserialize(common.DCBAddress)
	dcbPk := accountDCB.KeySet.PaymentAddress.Pk
	// TODO(@0xbunyip); use unique receiver, ignore case DCB loan from itself
	//	unique, receiver, amount := txr.GetUniqueReceiver()
	//	fmt.Printf("unique, receiver, amount: %v, %x, %v\n", unique, receiver, amount)
	//	fmt.Printf("input, output coins: %d %d\n", len(txr.GetProof().InputCoins), len(txr.GetProof().OutputCoins))
	//	if !unique || !bytes.Equal(receiver, dcbPk) {
	//		return false, fmt.Errorf("Loan payment must be sent to DCB address")
	//	}
	amount := uint64(0)
	receivers, amounts := txr.GetReceivers()
	for i, pubkey := range receivers {
		if bytes.Equal(pubkey, dcbPk) {
			amount += amounts[i]
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
	principle, interest, deadline, err := bcr.GetLoanPayment(lp.LoanID)
	if err != nil {
		return false, err
	}
	totalDebt := GetTotalDebt(
		principle,
		interest,
		requestMeta.Params.InterestRate,
		requestMeta.Params.Maturity,
		deadline,
		uint32(bcr.GetHeight()),
	)
	if lp.PayPrinciple && amount < totalDebt {
		return false, fmt.Errorf("Interest must be fully paid before paying principle")
	}
	return true, nil
}

func (lp *LoanPayment) ValidateSanityData(bcr BlockchainRetriever, txr Transaction) (bool, bool, error) {
	proof := txr.GetProof()
	if proof == nil || len(proof.InputCoins) < 1 || len(proof.OutputCoins) < 1 {
		return false, false, errors.Errorf("Loan payment must send Constant")
	}
	return true, true, nil // continue checking for fee
}

func (lp *LoanPayment) ValidateMetadataByItself() bool {
	return true
}

func GetTotalDebt(principle, interest, interestRate uint64, maturity, deadline, currentHeight uint32) uint64 {
	totalInterest := uint64(0)
	if currentHeight >= deadline {
		totalInterest = interest + uint64(1+(currentHeight-deadline)/maturity)*GetInterestPerTerm(principle, interestRate)
	}
	return principle + totalInterest
}

func GetInterestPerTerm(principle, interestRate uint64) uint64 {
	return principle * interestRate / Decimals
}
