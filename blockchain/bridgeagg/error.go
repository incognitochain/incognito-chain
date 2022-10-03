package bridgeagg

import (
	"fmt"

	"github.com/pkg/errors"
)

const (
	OtherError = iota
	NotFoundUnifiedTokenIDError
	InvalidPTokenIDError
	InvalidNetworkIDError
	InvalidStatusError
	InvalidConvertAmountError
	CalRewardError
	ProducerUpdateStateError
	ProcessUpdateStateError
	UnmarshalShieldProofError
	ValidateDoubleShieldProofError
	ShieldProofIsSubmittedError
	ExtractDataFromReceiptError
	InvalidTokenPairError
	CheckVaultUnshieldError
	StoreShieldExtTxError
	CheckBridgeTokenExistedError
	StoreBridgeTokenError
	InsufficientFundsVaultError
	NoValidShieldEventError
)

var ErrCodeMessage = map[int]struct {
	Code    int
	Message string
}{
	OtherError:                     {1, "Other error"},
	NotFoundUnifiedTokenIDError:    {1000, "Not found punified token id in network"},
	InvalidPTokenIDError:           {1001, "Invalid ptoken id in punifiedToken"},
	InvalidNetworkIDError:          {1002, "Invalid networkID"},
	InvalidStatusError:             {1003, "Invalid instruction status error"},
	InvalidConvertAmountError:      {1004, "Invalid convert amount"},
	CalRewardError:                 {1005, "Calculate reward error"},
	ProducerUpdateStateError:       {1006, "Beacon producer update state error"},
	ProcessUpdateStateError:        {1007, "Beacon process update state error"},
	UnmarshalShieldProofError:      {1008, "Unmarshal shielding proof error"},
	ValidateDoubleShieldProofError: {1009, "Validate double shielding proof error"},
	ShieldProofIsSubmittedError:    {1010, "Shield proof was submitted"},
	ExtractDataFromReceiptError:    {1011, "Extract data from shielding receipt error"},
	InvalidTokenPairError:          {1012, "Invalid token pair"},
	CheckVaultUnshieldError:        {1013, "Check vaults for new unshielding request error"},
	StoreShieldExtTxError:          {1014, "Store shield external tx error"},
	CheckBridgeTokenExistedError:   {1015, "Check bridge token existed error"},
	StoreBridgeTokenError:          {1016, "Store bridge token error"},
	InsufficientFundsVaultError:    {1017, "Insufficient funds in Vault"},
	NoValidShieldEventError:        {1018, "Shielding receipt contains no valid shield event"},
}

type BridgeAggError struct {
	Code    int    // The code to send with reject messages
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
