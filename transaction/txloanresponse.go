package transaction

import (
	"strconv"

	"github.com/ninjadotorg/constant/common"
)

type ValidLoanResponse int

const (
	Accept ValidLoanResponse = iota
	Reject
)

type LoanResponse struct {
	LoanID     []byte
	Response   ValidLoanResponse
	ValidUntil uint64
}

type TxLoanResponse struct {
	TxWithFee
	*LoanResponse // data for a loan response
}

func CreateTxLoanResponse(
	feeArgs FeeArgs,
	loanResponse *LoanResponse,
) (*TxLoanResponse, error) {
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

	txLoanResponse := &TxLoanResponse{
		TxWithFee:    TxWithFee{Tx: tx},
		LoanResponse: loanResponse,
	}

	return txLoanResponse, nil
}

func (tx *TxLoanResponse) Hash() *common.Hash {
	// get hash of tx
	record := tx.Tx.Hash().String()

	// add more hash of loan response data
	record += string(tx.LoanID)
	record += strconv.Itoa(int(tx.ValidUntil))

	// final hash
	hash := common.DoubleHashH([]byte(record))
	return &hash
}

func (tx *TxLoanResponse) ValidateTransaction() bool {
	// validate for normal tx
	if !tx.Tx.ValidateTransaction() {
		return false
	}

	// TODO(@0xbunyip): check if only board members created this tx
	if tx.Response != Accept || tx.Response != Reject {
		return false
	}

	return true
}

func (tx *TxLoanResponse) GetType() string {
	return common.TxLoanResponse
}
