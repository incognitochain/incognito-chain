package transaction

import (
	"fmt"
	"github.com/pkg/errors"
)

const (
	UnexpectedErr      = iota
	WrongTokenTxType
	CustomTokenExisted
	WrongInput
)

var ErrCodeMessage = map[int]struct {
	code    int
	message string
}{
	UnexpectedErr:      {-1, "Unexpected error"},
	WrongTokenTxType:   {-2, "Can't handle this TokenTxType"},
	CustomTokenExisted: {-3, "This token is existed in network"},
	WrongInput:         {-4, "Wrong input transaction"},
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
