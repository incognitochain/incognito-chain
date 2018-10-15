package wallet

const (
	ErrInvalidChecksum = "ErrInvalidChecksum"
)

var ErrCodeMessage = map[string]struct {
	code    int
	message string
}{
	ErrInvalidChecksum: {-1000, "Checksum does not match"},
}

type WalletError struct {
	code    int
	message string
	err     error
}

func NewWalletError(key string, err error) *WalletError {
	return &WalletError{
		err:     err,
		code:    ErrCodeMessage[key].code,
		message: ErrCodeMessage[key].message,
	}
}
