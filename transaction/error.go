package transaction

import (
	"fmt"
	"github.com/pkg/errors"
)

const (
	UnexpectedErr = iota
	WrongTokenTxType
	CustomTokenExisted
	WrongInput
	WrongSig
	DoubleSpend
	TxNotExist
	RandomCommitmentErr
	InvalidSanityDataPRV
	InvalidSanityDataPrivacyToken
	InvalidDoubleSpendPRV
	InvalidDoubleSpendPrivacyToken
)

var ErrCodeMessage = map[int]struct {
	code    int
	message string
}{
	// for common
	UnexpectedErr:       {-1000, "Unexpected error"},
	WrongTokenTxType:    {-1001, "Can't handle this TokenTxType"},
	CustomTokenExisted:  {-1002, "This token is existed in network"},
	WrongInput:          {-1003, "Wrong input transaction"},
	WrongSig:            {-1004, "Wrong signature"},
	DoubleSpend:         {-1005, "Double spend"},
	TxNotExist:          {-1006, "Not exist tx for this"},
	RandomCommitmentErr: {-1007, "Number of list commitments indices must be corresponding with number of input coins"},

	// for PRV
	InvalidSanityDataPRV:  {-2000, "Invalid sanity data for PRV"},
	InvalidDoubleSpendPRV: {-2001, "Double spend PRV in blockchain"},

	// for privacy token
	InvalidSanityDataPrivacyToken:  {-3000, "Invalid sanity data for privacy Token"},
	InvalidDoubleSpendPrivacyToken: {-3001, "Double spend privacy Token in blockchain"},
}

type TransactionError struct {
	code    int
	message string
	err     error
}

func (e TransactionError) Error() string {
	return fmt.Sprintf("%+v: %+v %+v", e.code, e.message, e.err)
}

func NewTransactionErr(key int, err error) *TransactionError {
	return &TransactionError{
		err:     errors.Wrap(err, ErrCodeMessage[key].message),
		code:    ErrCodeMessage[key].code,
		message: ErrCodeMessage[key].message,
	}
}
