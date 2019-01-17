package metadata

import (
	"bytes"
	"encoding/hex"
	"fmt"

	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/database"
	"github.com/ninjadotorg/constant/wallet"
	"github.com/pkg/errors"
)

const Decimals = uint64(10000) // Each float number is multiplied by this value to store as uint64

type LoanPayment struct {
	LoanID []byte
	MetadataBase
}

func NewLoanPayment(data map[string]interface{}) (Metadata, error) {
	result := LoanPayment{}
	s, _ := hex.DecodeString(data["LoanID"].(string))
	result.LoanID = s
	result.Type = LoanPaymentMeta
	return &result, nil
}

func (lp *LoanPayment) Hash() *common.Hash {
	record := string(lp.LoanID)

	// final hash
	record += lp.MetadataBase.Hash().String()
	hash := common.DoubleHashH([]byte(record))
	return &hash
}

func (lp *LoanPayment) ValidateTxWithBlockChain(txr Transaction, bcr BlockchainRetriever, chainID byte, db database.DatabaseInterface) (bool, error) {
	fmt.Println("Start validating LoanPayment tx with blockchain!!!")
	// Check if loan is withdrawed
	_, _, _, err := bcr.GetLoanPayment(lp.LoanID)
	if err != nil {
		return common.FalseValue, err
	}

	// Check loan payment
	accountBurn, _ := wallet.Base58CheckDeserialize(common.BurningAddress)
	burnPk := accountBurn.KeySet.PaymentAddress.Pk
	unique, receiver, amount := txr.GetUniqueReceiver()
	fmt.Printf("unique, receiver, amount: %v, %x, %v\n", unique, receiver, amount)
	if !unique || !bytes.Equal(receiver, burnPk) {
		return common.FalseValue, fmt.Errorf("Loan payment must be sent to burn address")
	}

	return common.TrueValue, nil
}

func (lp *LoanPayment) ValidateSanityData(bcr BlockchainRetriever, txr Transaction) (bool, bool, error) {
	proof := txr.GetProof()
	if proof == nil || len(proof.InputCoins) < 1 || len(proof.OutputCoins) < 1 {
		return common.FalseValue, common.FalseValue, errors.Errorf("Loan payment must send Constant")
	}
	return common.TrueValue, common.TrueValue, nil // continue checking for fee
}

func (lp *LoanPayment) ValidateMetadataByItself() bool {
	return common.TrueValue
}

func GetTotalInterest(principle, interest, interestRate, maturity, deadline, currentHeight uint64) uint64 {
	totalInterest := uint64(0)
	if currentHeight >= deadline {
		perTerm := GetInterestPerTerm(principle, interestRate)
		totalInterest = interest
		if perTerm > 0 {
			totalInterest += (currentHeight - deadline) / maturity * perTerm
		}
	}
	return totalInterest
}

func GetTotalDebt(principle, interest, interestRate, maturity, deadline, currentHeight uint64) uint64 {
	return principle + GetTotalInterest(principle, interest, interestRate, maturity, deadline, currentHeight)
}

func GetInterestPerTerm(principle, interestRate uint64) uint64 {
	return principle * interestRate / Decimals
}
