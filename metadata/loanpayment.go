package metadata

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strconv"

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

func (lp *LoanPayment) ValidateTxWithBlockChain(txr Transaction, bcr BlockchainRetriever, shardID byte, db database.DatabaseInterface) (bool, error) {
	fmt.Println("Start validating LoanPayment tx with blockchain!!!")
	// Check if loan is withdrawed
	_, _, _, err := bcr.GetLoanPayment(lp.LoanID)
	if err != nil {
		return false, err
	}

	// Check loan payment
	keyWalletBurningAdd, _ := wallet.Base58CheckDeserialize(common.BurningAddress)
	burnPk := keyWalletBurningAdd.KeySet.PaymentAddress.Pk
	unique, receiver, amount := txr.GetUniqueReceiver()
	fmt.Printf("unique, receiver, amount: %v, %x, %v\n", unique, receiver, amount)
	if !unique || !bytes.Equal(receiver, burnPk) {
		return false, errors.New("Loan payment must be sent to burn address")
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

type LoanPaymentAction struct {
	LoanID       []byte
	AmountSent   uint64
	InterestRate uint64
	Maturity     uint64
}

func (lp *LoanPayment) BuildReqActions(txr Transaction, bcr BlockchainRetriever, shardID byte) ([][]string, error) {
	lrActionValue, err := getLoanPaymentActionValue(txr, bcr, lp.LoanID)
	if err != nil {
		return nil, err
	}
	fmt.Printf("[db] LoanPay built action: %s\n", lrActionValue)
	lrAction := []string{strconv.Itoa(LoanPaymentMeta), lrActionValue}
	return [][]string{lrAction}, nil
}

func getLoanPaymentActionValue(txr Transaction, bcr BlockchainRetriever, loanID []byte) (string, error) {
	_, _, amountSent := txr.GetUniqueReceiver()

	// Get loan params
	reqMeta, err := bcr.GetLoanRequestMeta(loanID)
	if err != nil {
		return "", err
	}

	action := &LoanPaymentAction{
		LoanID:       loanID,
		AmountSent:   amountSent,
		InterestRate: reqMeta.Params.InterestRate,
		Maturity:     reqMeta.Params.Maturity,
	}
	value, _ := json.Marshal(action)
	return string(value), nil
}

func ParseLoanPaymentActionValue(value string) ([]byte, uint64, uint64, uint64, error) {
	action := &LoanPaymentAction{}
	err := json.Unmarshal([]byte(value), action)
	if err != nil {
		return nil, 0, 0, 0, err
	}
	return action.LoanID, action.AmountSent, action.InterestRate, action.Maturity, nil
}

func CalculateInterestPaid(amountSent, principle, interest, deadline, interestRate, maturity, currentHeight uint64) uint64 {
	// Only keep interest
	totalInterest := GetTotalInterest(
		principle,
		interest,
		interestRate,
		maturity,
		deadline,
		currentHeight,
	)
	interestPaid := amountSent
	if amountSent > totalInterest {
		interestPaid = totalInterest
	}
	fmt.Printf("[db] calcInterestPaid: %d %d %d %d %d %d\n", principle, interest, deadline, amountSent, totalInterest, interestPaid)
	return interestPaid
}
