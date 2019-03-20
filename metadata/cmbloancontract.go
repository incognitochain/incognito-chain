package metadata

import (
	"bytes"

	"github.com/constant-money/constant-chain/common"
	"github.com/constant-money/constant-chain/database"
	"github.com/constant-money/constant-chain/privacy"
	"github.com/constant-money/constant-chain/wallet"
	"github.com/pkg/errors"
)

type principleBuilder func(map[string]interface{}) (Principler, error)
type interestBuilder func(map[string]interface{}) (Interester, error)

var principleBuilderMap = map[string]principleBuilder{
	"SinglePrinciplePayment": BuildSinglePrinciplePayment,
}

var interestBuilderMap = map[string]interestBuilder{
	"SingleInterestPayment": BuildSingleInterestPayment,
}

// Principler contains logic related to the principle of a loan.
type Principler interface {
	Value() uint64    // Amount of principle left to repay
	Term() uint64     // Amount of principle to repay next term
	NextTerm() uint64 // Deadline of the next principle payment term (in block height)

	Hash() string // For tx hash
}

// SinglePrinciplePayment implements Principler, user repays all principle at
// the end of the loan.
type SinglePrinciplePayment struct {
	TotalValue uint64
	Deadline   uint64 // next principle deadline
}

func BuildSinglePrinciplePayment(data map[string]interface{}) (Principler, error) {
	tmpValue, ok := data["TotalValue"]
	totalValue, okType := tmpValue.(float64)
	if !ok || !okType {
		return nil, errors.Errorf("Error parsing Principle")
	}
	tmpDeadline, ok := data["Deadline"]
	deadline, okType := tmpDeadline.(float64)
	if !ok || !okType {
		return nil, errors.Errorf("Error parsing principle deadline")
	}
	principle := &SinglePrinciplePayment{
		TotalValue: uint64(totalValue),
		Deadline:   uint64(deadline),
	}
	return principle, nil
}

func (p *SinglePrinciplePayment) Value() uint64 {
	return p.TotalValue
}

func (p *SinglePrinciplePayment) Term() uint64 {
	return p.TotalValue
}

func (p *SinglePrinciplePayment) NextTerm() uint64 {
	return p.Deadline
}

func (p *SinglePrinciplePayment) Hash() string {
	return string(p.TotalValue) + string(p.Deadline)
}

// AmortizedPrinciplePayment implements Principler, user repays part of the
// principle for each term of the loan.
type AmortizedPrinciplePayment struct {
	SinglePrinciplePayment
	TermLength uint64 // #blocks between 2 payments
}

func (p *AmortizedPrinciplePayment) Term() uint64 {
	return 0 // calculate
}

func (p *AmortizedPrinciplePayment) Hash() string {
	return string(p.TotalValue) + string(p.Deadline) + string(p.TermLength)
}

// Principler contains logic related to the interest of a loan.
type Interester interface {
	Value() uint64    // Amount of interest left to repay
	Term() uint64     // Amount of interest to repay next term
	NextTerm() uint64 // Deadline of the next interest payment term (in block height)

	Hash() string
}

type SingleInterestPayment struct {
	TotalValue uint64
	Deadline   uint64
}

func BuildSingleInterestPayment(data map[string]interface{}) (Interester, error) {
	tmpValue, ok := data["TotalValue"]
	totalValue, errType := tmpValue.(float64)
	if !ok || !errType {
		return nil, errors.Errorf("Error parsing Interest")
	}
	tmpDeadline, ok := data["Deadline"]
	deadline, errType := tmpDeadline.(float64)
	if !ok || !errType {
		return nil, errors.Errorf("Error parsing interest Deadline")
	}
	principle := &SingleInterestPayment{
		TotalValue: uint64(totalValue),
		Deadline:   uint64(deadline),
	}
	return principle, nil
}

func (p *SingleInterestPayment) Value() uint64 {
	return p.TotalValue
}

func (p *SingleInterestPayment) Term() uint64 {
	return p.TotalValue
}

func (p *SingleInterestPayment) NextTerm() uint64 {
	return p.Deadline
}

func (p *SingleInterestPayment) Hash() string {
	return string(p.TotalValue) + string(p.Deadline)
}

type MultipleInterestPayment struct {
	SingleInterestPayment
	TermLength uint64 //#blocks between 2 payments
}

func (p *MultipleInterestPayment) Term() uint64 {
	return 0 // calculate
}

func (p *MultipleInterestPayment) Hash() string {
	return string(p.TotalValue) + string(p.Deadline) + string(p.TermLength)
}

// CMBLoanContract represents a loan created by the borrower; the contract must
// be accepted by CMB before considered as in effect
type CMBLoanContract struct {
	Principle Principler
	Interest  Interester

	Receiver   privacy.PaymentAddress // must be the same as the one creating this tx (to receive loan when CMB accept)
	CMBAddress privacy.PaymentAddress // to get the correct CMB from db

	ValidUntil uint64
	MetadataBase
}

func NewCMBLoanContract(data map[string]interface{}) (*CMBLoanContract, error) {
	keyRec, errRec := wallet.Base58CheckDeserialize(data["Receiver"].(string))
	keyCMB, errCMB := wallet.Base58CheckDeserialize(data["CMBAddress"].(string))
	if errRec != nil || errCMB != nil {
		return nil, errors.Errorf("Error parsing address")
	}

	validUntil := uint64(data["ValidUntil"].(float64))
	principleData := data["Principle"].(map[string]interface{})
	principleType := principleData["Type"].(string)
	principle, err := principleBuilderMap[principleType](data)
	if err != nil {
		return nil, err
	}
	interestData := data["Interest"].(map[string]interface{})
	interestType := interestData["Type"].(string)
	interest, err := interestBuilderMap[interestType](data)
	if err != nil {
		return nil, err
	}
	result := CMBLoanContract{
		Principle:  principle,
		Interest:   interest,
		Receiver:   keyRec.KeySet.PaymentAddress,
		CMBAddress: keyCMB.KeySet.PaymentAddress,
		ValidUntil: validUntil,
	}

	result.Type = CMBLoanContractMeta
	return &result, nil
}

func (lc *CMBLoanContract) Hash() *common.Hash {
	record := lc.Principle.Hash()
	record += lc.Interest.Hash()
	record += string(lc.Receiver.Bytes())
	record += string(lc.CMBAddress.Bytes())
	record += string(lc.ValidUntil)

	// final hash
	record += string(lc.MetadataBase.Hash().String())
	hash := common.DoubleHashH([]byte(record))
	return &hash
}

func (lc *CMBLoanContract) ValidateTxWithBlockChain(txr Transaction, bcr BlockchainRetriever, shardID byte, db database.DatabaseInterface) (bool, error) {
	cmbChainHeight := bcr.GetChainHeight(shardID)
	if cmbChainHeight+1 >= lc.ValidUntil {
		return false, errors.Errorf("ValidUntil must be bigger than current block height of CMB")
	}

	// Check if CMB existed
	_, _, _, _, _, _, err := bcr.GetCMB(lc.CMBAddress.Bytes())
	if err != nil {
		return false, err
	}

	// Receiver must be valid
	if !bytes.Equal(txr.GetSigPubKey(), lc.Receiver.Pk[:]) {
		return false, errors.Errorf("Receiver must be the one creating this tx")
	}

	// TODO(@0xbunyip): add validate methods to Principler and Interester and
	// call them
	return true, nil
}

func (lc *CMBLoanContract) ValidateSanityData(bcr BlockchainRetriever, txr Transaction) (bool, bool, error) {
	if len(lc.Receiver.Pk) <= 0 {
		return false, false, errors.Errorf("Receiver must be set")
	}
	if len(lc.CMBAddress.Pk) <= 0 {
		return false, false, errors.Errorf("CMBAddress must be set")
	}
	return true, true, nil // continue to check for fee
}

func (lc *CMBLoanContract) ValidateMetadataByItself() bool {
	return true
}

func (lc *CMBLoanContract) CalculateSize() uint64 {
	return calculateSize(lc)
}
