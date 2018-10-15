package wallet

import "fmt"

const (
	InvalidChecksumErr    = "InvalidChecksumErr"
	WrongPassphraseErr    = "WrongPassphraseErr"
	ExistedAccountErr     = "ExistedAccountErr"
	ExistedAccountNameErr = "ExistedAccountNameErr"
	UnexpectedErr         = "UnexpectedErr"
)

var ErrCodeMessage = map[string]struct {
	code    int
	message string
}{
	UnexpectedErr: {-1, "Unexpected error"},


	InvalidChecksumErr:    {-1000, "Checksum does not match"},
	WrongPassphraseErr:    {-1001, "Wrong passphrase"},
	ExistedAccountErr:     {-1002, "Existed account"},
	ExistedAccountNameErr: {-1002, "Existed account name"},
}

type WalletError struct {
	code    int
	message string
	err     error
}

func (e WalletError) Error() string {
	return fmt.Sprintf("%v: %v", e.code, e.message)
}

func NewWalletError(key string, err error) *WalletError {
	return &WalletError{
		err:     err,
		code:    ErrCodeMessage[key].code,
		message: ErrCodeMessage[key].message,
	}
}
