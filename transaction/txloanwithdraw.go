package transaction

import (
	"github.com/ninjadotorg/constant/common"
	"encoding/hex"
)

type LoanWithdraw struct {
	LoanID []byte
	Key    []byte
}

func NewLoanWithdraw(data map[string]interface{}) *LoanWithdraw {
	result := LoanWithdraw{}
	s, _ := hex.DecodeString(data["LoanID"].(string))
	result.LoanID = s
	s, _ = hex.DecodeString(data["Key"].(string))
	result.Key = s

	return &result
}

type TxLoanWithdraw struct {
	Tx
	*LoanWithdraw // data for a loan response
}

func CreateTxLoanWithdraw(
	feeArgs FeeArgs,
	loanWithdraw *LoanWithdraw,
) (*TxLoanWithdraw, error) {
	// Create tx for fee
	tx, err := CreateTx(
		feeArgs.SenderKey,
		feeArgs.PaymentInfo,
		feeArgs.Rts,
		feeArgs.UsableTx,
		feeArgs.Commitments,
		feeArgs.Fee,
		feeArgs.SenderChainID,
		false,
	)
	if err != nil {
		return nil, err
	}

	txLoanWithdraw := &TxLoanWithdraw{
		Tx:           *tx,
		LoanWithdraw: loanWithdraw,
	}

	return txLoanWithdraw, nil
}

func (tx *TxLoanWithdraw) Hash() *common.Hash {
	// get hash of tx
	record := tx.Tx.Hash().String()

	// add more hash of loan response data
	record += string(tx.LoanID)
	record += string(tx.Key)

	// final hash
	hash := common.DoubleHashH([]byte(record))
	return &hash
}

func (tx *TxLoanWithdraw) ValidateTransaction() bool {
	// validate for normal tx
	if !tx.Tx.ValidateTransaction() {
		return false
	}

	if len(tx.Key) != LoanKeyLen {
		return false
	}
	return true
}

func (tx *TxLoanWithdraw) GetType() string {
	return common.TxLoanWithdraw
}
