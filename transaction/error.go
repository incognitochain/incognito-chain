package transaction

import (
	"fmt"
	"github.com/pkg/errors"
)

const (
	UnexpectedErr = iota
)

var ErrCodeMessage = map[int]struct {
	code    int
	message string
}{
	UnexpectedErr: {-1, "Unexpected error"},
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
