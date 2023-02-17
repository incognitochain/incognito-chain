package bridgehub

import (
	"fmt"

	"github.com/pkg/errors"
)

const (
	OtherError = iota
	InvalidNumberValidatorError
	InvalidStakedAmountValidatorError
	BridgeIDExistedError
	InvalidStatusError
	InvalidBTCShieldStatus
	StoreShieldExtTxError
)

var ErrCodeMessage = map[int]struct {
	Code    int
	Message string
}{
	OtherError:                        {1, "Other error"},
	InvalidNumberValidatorError:       {2000, "Number of validators must meet the minimum number"},
	InvalidStakedAmountValidatorError: {2001, "Validator staked amount must meet the minimum amount"},
	BridgeIDExistedError:              {2002, "Bridge validator was registered before"},
	InvalidStatusError:                {2003, "Invalid instruction status error"},
	InvalidBTCShieldStatus:            {2004, "Invalid btc shield status"},
}

type BridgeHubError struct {
	Code    int    // The code to send with reject messages
	Message string // Human readable message of the issue
	Err     error
}

// Error satisfies the error interface and prints human-readable errors.
func (e BridgeHubError) Error() string {
	return fmt.Sprintf("%d: %s %+v", e.Code, e.Message, e.Err)
}

func NewBridgeHubErrorWithValue(key int, err error, params ...interface{}) *BridgeHubError {
	return &BridgeHubError{
		Code:    ErrCodeMessage[key].Code,
		Message: fmt.Sprintf(ErrCodeMessage[key].Message, params),
		Err:     errors.Wrap(err, ErrCodeMessage[key].Message),
	}
}
