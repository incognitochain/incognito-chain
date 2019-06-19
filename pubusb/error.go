package pubusb

import (
	"fmt"
	"github.com/pkg/errors"
)

const (
	UnexpectedError = iota
	UnmashallJsonError
	MashallJsonError
)

var ErrCodeMessage = map[int]struct {
	Code    int
	Message string
}{
	UnexpectedError: {-1000, "Unexpected Error"},
	UnmashallJsonError: {-1001, "Umarshall Json Error"},
	MashallJsonError: {-1002, "Marshall Json Error"},
}

type PubsubError struct {
	Code    int    // The code to send with reject messages
	Message string // Human readable message of the issue
	Err     error
}

// Error satisfies the error interface and prints human-readable errors.
func (e PubsubError) Error() string {
	return fmt.Sprintf("%d: %s %+v", e.Code, e.Message, e.Err)
}

// txRuleError creates an underlying MempoolTxError with the given a set of
// arguments and returns a RuleError that encapsulates it.
func (e *PubsubError) Init(key int, err error) {
	e.Code = ErrCodeMessage[key].Code
	e.Message = ErrCodeMessage[key].Message
	e.Err = errors.Wrap(err, e.Message)
}

func NewMetricError(key int, err error) *PubsubError {
	return &PubsubError{
		Code:    ErrCodeMessage[key].Code,
		Message: ErrCodeMessage[key].Message,
		Err:     errors.Wrap(err, ErrCodeMessage[key].Message),
	}
}
