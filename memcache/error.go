package memcache

import (
	"fmt"
	"github.com/pkg/errors"
)

const (
	MemCacheClosedError = iota
	MemCacheNotFoundError
	ExpiredError
)

var ErrCodeMessage = map[int]struct {
	Code    int
	Message string
}{
	MemCacheClosedError:   {-1, "Database closed"},
	MemCacheNotFoundError: {-2, "Not found for key"},
	ExpiredError:          {-3, "expired key time"},
}

type MemCacheError struct {
	Code    int
	Message string
	err     error
}

func (e MemCacheError) Error() string {
	return fmt.Sprintf("%d: %s %+v", e.Code, e.Message, e.err)
}

func NewMemCacheError(key int, err error) *MemCacheError {
	return &MemCacheError{
		Code:    ErrCodeMessage[key].Code,
		Message: ErrCodeMessage[key].Message,
		err:     errors.Wrap(err, ErrCodeMessage[key].Message),
	}
}
