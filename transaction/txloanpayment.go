package transaction

import (
	"github.com/ninjadotorg/constant/common"
	"encoding/hex"
)

type LoanPayment struct {
	LoanID []byte
}

func NewLoanPayment(data map[string]interface{}) *LoanPayment {
	result := LoanPayment{}
	s, _ := hex.DecodeString(data["LoanID"].(string))
	result.LoanID = s
	return &result
}

type TxLoanPayment struct {
	Tx
	*LoanPayment // data for a loan response
}

func CreateTxLoanPayment(
	feeArgs FeeArgs,
	loanPayment *LoanPayment,
) (*TxLoanPayment, error) {
	// Create tx for fee
	tx, err := CreateTx(
		feeArgs.SenderKey,
		feeArgs.PaymentInfo,
		feeArgs.Rts,
		feeArgs.UsableTx,
		feeArgs.Commitments,
		feeArgs.Fee,
		feeArgs.SenderChainID,
		true,
	)
	if err != nil {
		return nil, err
	}

	txLoanPayment := &TxLoanPayment{
		Tx:          *tx,
		LoanPayment: loanPayment,
	}

	return txLoanPayment, nil
}

func (tx *TxLoanPayment) Hash() *common.Hash {
	// get hash of tx
	record := tx.Tx.Hash().String()

	// add more hash of loan response data
	record += string(tx.LoanID)

	// final hash
	hash := common.DoubleHashH([]byte(record))
	return &hash
}

func (tx *TxLoanPayment) ValidateTransaction() bool {
	// validate for normal tx
	if !tx.Tx.ValidateTransaction() {
		return false
	}

	for _, desc := range tx.Tx.Descs {
		if desc.Note == nil {
			// TODO(@0xbunyip): check if payment is sent to DCB
			return false // Loan payment tx must be non-privacy-protocol
		}
	}
	return true
}

func (tx *TxLoanPayment) GetType() string {
	return common.TxLoanPayment
}
