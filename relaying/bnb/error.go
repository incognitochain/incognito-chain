package relaying

import (
	"fmt"
	"github.com/pkg/errors"
)

const (
	UnexpectedErr = iota
	InvalidBasicSignedHeaderErr
	InvalidSignatureSignedHeaderErr
	InvalidNewHeaderErr
	InvalidBasicHeaderErr
)

var ErrCodeMessage = map[int]struct {
	Code    int
	Message string
}{
	UnexpectedErr: {-14000, "Unexpected error"},

	InvalidBasicSignedHeaderErr:     {-14001, "Invalid basic signed header error"},
	InvalidSignatureSignedHeaderErr: {-14002, "Invalid signature signed header error"},
	InvalidNewHeaderErr:             {-14003, "Invalid new header"},
	InvalidBasicHeaderErr:           {-14004, "Invalid basic header error"},
}

type BNBRelayingError struct {
	Code    int
	Message string
	err     error
}

func (e BNBRelayingError) Error() string {
	return fmt.Sprintf("%+v: %+v %+v", e.Code, e.Message, e.err)
}

func (e BNBRelayingError) GetCode() int {
	return e.Code
}

func NewBNBRelayingError(key int, err error) *BNBRelayingError {
	return &BNBRelayingError{
		err:     errors.Wrap(err, ErrCodeMessage[key].Message),
		Code:    ErrCodeMessage[key].Code,
		Message: ErrCodeMessage[key].Message,
	}
}
