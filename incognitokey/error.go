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
)

var ErrCodeMessage = map[int]struct {
	code    int
	message string
}{
	InvalidPrivateKeyErr: {-201, "Private key is invalid"},
	B58DecodePubKeyErr:  {-202, "Base58 decode pub key error"},
	B58DecodeSigErr:  {-203, "Base58 decode signature error"},
	B58ValidateErr:  {-203, "Base58 validate data error"},
	InvalidDataValidateErr:  {-203, "Validated base58 data is invalid"},
	SignDataB58Err:  {-203, "Signing B58 data error"},
	InvalidDataSignErr:  {-203, "Signed data is invalid"},
	InvalidVerificationKeyErr:  {-203, "Verification key is invalid"},
	DecodeFromStringErr:  {-203, "Decode key set from string error"},
}

type CashecError struct {
	code    int
	message string
	err     error
}

func (e CashecError) Error() string {
	return fmt.Sprintf("%+v: %+v", e.code, e.message)
}

func (e CashecError) GetCode() int {
	return e.code
}

func NewCashecError(key int, err error) *CashecError {
	return &CashecError{
		err:     errors.Wrap(err, ErrCodeMessage[key].message),
		code:    ErrCodeMessage[key].code,
		message: ErrCodeMessage[key].message,
	}
}


