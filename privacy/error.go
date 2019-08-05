package privacy

import (
	"fmt"
	"github.com/pkg/errors"
)

const (
	UnexpectedErr = iota
	InvalidOutputValue
	ProvingErr
	VerificationErr
	MarshalErr
	UnmarshalErr
	SetBytesProofErr
	EncryptOutputCoinErr
	DecryptOutputCoinErr
	DecompressTransmissionKeyErr
)

var ErrCodeMessage = map[int]struct {
	Code    int
	Message string
}{
	UnexpectedErr: {-1, "Unexpected error"},

	InvalidOutputValue:           {-2, "Invalid output value"},
	ProvingErr:                   {-3, "Zero knowledge proving error"},
	VerificationErr:              {-4, "Zero knowledge verification error"},
	MarshalErr:                   {-5, "Marshal payment proof error"},
	UnmarshalErr:                 {-6, "Unmarshal payment proof error"},
	SetBytesProofErr:             {-6, "Set bytes payment proof error"},
	EncryptOutputCoinErr:         {-7, "Encrypt output coins error"},
	DecryptOutputCoinErr:         {-8, "Decrypt output coins error"},
	DecompressTransmissionKeyErr: {-7, "Can not decompress transmission key error"},
}

type PrivacyError struct {
	Code    int
	Message string
	err     error
}

func (e PrivacyError) Error() string {
	return fmt.Sprintf("%+v: %+v %+v", e.Code, e.Message, e.err)
}

func (e PrivacyError) GetCode() int {
	return e.Code
}

func NewPrivacyErr(key int, err error) *PrivacyError {
	return &PrivacyError{
		err:     errors.Wrap(err, ErrCodeMessage[key].Message),
		Code:    ErrCodeMessage[key].Code,
		Message: ErrCodeMessage[key].Message,
	}
}
