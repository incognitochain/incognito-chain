package bridgeagg

import (
	"fmt"

	"github.com/pkg/errors"
)

const (
	OtherError = iota
	NotFoundTokenIDInNetwork
	NotFoundNetworkID
	FailToBuildModifyListToken
)

var ErrCodeMessage = map[int]struct {
	Code    uint
	Message string
}{
	OtherError:                 {1, "Not found token id in network"},
	NotFoundTokenIDInNetwork:   {1000, "Not found token id in network"},
	NotFoundNetworkID:          {1001, "Not found networkID"},
	FailToBuildModifyListToken: {1100, "Fail to build modify list token instruction"},
}

type BridgeAggError struct {
	Code    uint   // The code to send with reject messages
	Message string // Human readable message of the issue
	Err     error
}

// Error satisfies the error interface and prints human-readable errors.
func (e BridgeAggError) Error() string {
	return fmt.Sprintf("%d: %s %+v", e.Code, e.Message, e.Err)
}

func NewBridgeAggErrorWithValue(key int, err error, params ...interface{}) *BridgeAggError {
	return &BridgeAggError{
		Code:    ErrCodeMessage[key].Code,
		Message: fmt.Sprintf(ErrCodeMessage[key].Message, params),
		Err:     errors.Wrap(err, ErrCodeMessage[key].Message),
	}
}
