package rawdbv2

import (
	"fmt"

	"github.com/pkg/errors"
)

const (
	// Beacon
	StoreBeaconBlockError = iota
	HasBeaconBlockError
	GetBeaconBlockByHashError
	GetBeaconBlockByIndexError
	DeleteBeaconBlockError

	// Shard
	StoreShardBlockError = iota
	HasShardBlockError
	GetShardBlockByHashError
	GetShardBlockByIndexError
	DeleteShardBlockError

	// tx
	StoreTransactionIndexError
	GetTransactionByHashError
)

var ErrCodeMessage = map[int]struct {
	Code    int
	message string
}{}

type RawdbError struct {
	err     error
	Code    int
	Message string
}

func (e RawdbError) GetErrorCode() int {
	return e.Code
}
func (e RawdbError) GetError() error {
	return e.err
}
func (e RawdbError) GetMessage() string {
	return e.Message
}

func (e RawdbError) Error() string {
	return fmt.Sprintf("%d: %+v", e.Code, e.err)
}

func NewRawdbError(key int, err error, params ...interface{}) *RawdbError {
	return &RawdbError{
		err:     errors.Wrap(err, ErrCodeMessage[key].message),
		Code:    ErrCodeMessage[key].Code,
		Message: fmt.Sprintf(ErrCodeMessage[key].message, params),
	}
}
