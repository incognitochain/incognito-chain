package grafana

import (
	"fmt"

	"github.com/pkg/errors"
)

const (
	UnexpectedError = iota
)

var ErrCodeMessage = map[int]struct {
	Code    int
	Message string
}{
	UnexpectedError: {-1000, "Unexpected Error"},
}

type MetricError struct {
	Code    int    // The code to send with reject messages
	Message string // Human readable message of the issue
	Err     error
}

// Error satisfies the error interface and prints human-readable errors.
func (e MetricError) Error() string {
	return fmt.Sprintf("%d: %s %+v", e.Code, e.Message, e.Err)
}

// txRuleError creates an underlying MempoolTxError with the given a set of
// arguments and returns a RuleError that encapsulates it.
func (e *MetricError) Init(key int, err error) {
	e.Code = ErrCodeMessage[key].Code
	e.Message = ErrCodeMessage[key].Message
	e.Err = errors.Wrap(err, e.Message)
}

func NewMetricError(key int, err error) *MetricError {
	return &MetricError{
		Code:    ErrCodeMessage[key].Code,
		Message: ErrCodeMessage[key].Message,
		Err:     errors.Wrap(err, ErrCodeMessage[key].Message),
	}
}
