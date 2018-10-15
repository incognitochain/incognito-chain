package wallet

import "errors"

var (
	ErrInvalidChecksum = errors.New("Checksum doesn't match")
)
