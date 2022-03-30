package bridgeagg

import (
	"fmt"

	"github.com/pkg/errors"
)

const (
	OtherError = iota
	NotFoundTokenIDInNetworkError
	NotFoundNetworkIDError
	InvalidRewardReserveError
	CalculateShieldAmountError
	CalculateUnshieldAmountError
	InvalidConvertAmountError
	FailToExtractDataError
	FailToVerifyTokenPairError
	OutOfRangeUni64Error
	FailToBuildModifyListTokenError
	FailToConvertTokenError
	FailToShieldError
	FailToUnshieldError
)

var ErrCodeMessage = map[int]struct {
	Code    uint
	Message string
}{
	OtherError:                      {1, "Other error"},
	NotFoundTokenIDInNetworkError:   {1000, "Not found token id in network"},
	NotFoundNetworkIDError:          {1001, "Not found networkID"},
	InvalidRewardReserveError:       {1003, "Invalid reward reserve"},
	CalculateShieldAmountError:      {1004, "Calculate shield amount"},
	CalculateUnshieldAmountError:    {1005, "Calculate unshield amount"},
	InvalidConvertAmountError:       {1006, "Invalid convert amount"},
	FailToExtractDataError:          {1007, "Fail to extract data"},
	FailToVerifyTokenPairError:      {1008, "Fail to verify token pair"},
	OutOfRangeUni64Error:            {1009, "Out of range uint64"},
	FailToBuildModifyListTokenError: {1100, "Fail to build modify list token instruction"},
	FailToConvertTokenError:         {1200, "Fail to convert token instruction"},
	FailToShieldError:               {1100, "Fail to shield instruction"},
	FailToUnshieldError:             {1100, "Fail to unshield instruction"},
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
