package statedb

import (
	"fmt"

	"github.com/pkg/errors"
)

const (
	InvalidByteArrayTypeError = iota
	InvalidCommitteeStateTypeError
	InvalidPaymentAddressTypeError
	InvalidIncognitoPublicKeyTypeError
	InvalidRewardReceiverStateTypeError
)

var ErrCodeMessage = map[int]struct {
	Code    int
	message string
}{
	InvalidByteArrayTypeError:           {-1000, "invalid byte array type"},
	InvalidCommitteeStateTypeError:      {-1001, "invalid committee state type"},
	InvalidPaymentAddressTypeError:      {-1002, "invalid payment address type"},
	InvalidIncognitoPublicKeyTypeError:  {-1003, "invalid incognito public key type"},
	InvalidRewardReceiverStateTypeError: {-1004, "invalid reward receiver state type "},
}

type StatedbError struct {
	err     error
	Code    int
	Message string
}

func (e StatedbError) GetErrorCode() int {
	return e.Code
}
func (e StatedbError) GetError() error {
	return e.err
}
func (e StatedbError) GetMessage() string {
	return e.Message
}

func (e StatedbError) Error() string {
	return fmt.Sprintf("%d: %+v", e.Code, e.err)
}

func NewStatedbError(key int, err error, params ...interface{}) *StatedbError {
	return &StatedbError{
		err:     errors.Wrap(err, ErrCodeMessage[key].message),
		Code:    ErrCodeMessage[key].Code,
		Message: fmt.Sprintf(ErrCodeMessage[key].message, params),
	}
}
