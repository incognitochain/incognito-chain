package incognitokey

import (
	"fmt"

	"github.com/pkg/errors"
)

const (
	InvalidPrivateKeyErr = iota
	B58DecodePubKeyErr
	B58DecodeSigErr
	B58ValidateErr
	InvalidDataValidateErr
	SignDataB58Err
	InvalidDataSignErr
	InvalidVerificationKeyErr
	DecodeFromStringErr
	SignError
	JSONError
)

var ErrCodeMessage = map[int]struct {
	Code    int
	Message string
}{
	InvalidPrivateKeyErr:      {-201, "Private key is invalid"},
	B58DecodePubKeyErr:        {-202, "Base58 decode pub key error"},
	B58DecodeSigErr:           {-203, "Base58 decode signature error"},
	B58ValidateErr:            {-204, "Base58 validate data error"},
	InvalidDataValidateErr:    {-205, "Validated base58 data is invalid"},
	SignDataB58Err:            {-206, "Signing B58 data error"},
	InvalidDataSignErr:        {-207, "Signed data is invalid"},
	InvalidVerificationKeyErr: {-208, "Verification key is invalid"},
	DecodeFromStringErr:       {-209, "Decode key set from string error"},
	SignError:                 {-210, "Can not sign data"},
	JSONError:                 {-211, "JSON Marshal, Unmarshal error"},
}

type CashecError struct {
	Code    int
	Message string
	err     error
}

func (e CashecError) Error() string {
	return fmt.Sprintf("%d: %s %+v", e.Code, e.Message, e.err)
}

func (e CashecError) GetCode() int {
	return e.Code
}

func NewCashecError(key int, err error) *CashecError {
	return &CashecError{
		err:     errors.Wrap(err, ErrCodeMessage[key].Message),
		Code:    ErrCodeMessage[key].Code,
		Message: ErrCodeMessage[key].Message,
	}
}
