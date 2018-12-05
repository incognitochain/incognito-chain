package transaction

import (
	"github.com/ninjadotorg/constant/common"
)

type TxLoanUnlock struct {
	Tx
	LoanID []byte
}

func (tx *TxLoanUnlock) Hash() *common.Hash {
	// get hash of tx
	record := tx.Tx.Hash().String()
	record += string(tx.LoanID)
	// final hash
	hash := common.DoubleHashH([]byte(record))
	return &hash
}

func (tx *TxLoanUnlock) ValidateTransaction() bool {
	// validate for normal tx
	if !tx.Tx.ValidateTransaction() {
		return false
	}
	// TODO(@0xbunyip): validate that there's a corresponding TxLoanWithdraw in the same block

	return true
}

func (tx *TxLoanUnlock) GetType() string {
	return common.TxLoanUnlock
}
