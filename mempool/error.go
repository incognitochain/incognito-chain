package mempool

import (
	"github.com/pkg/errors"
	"fmt"
)

const (
	RejectDuplicateTx      = iota
	RejectInvalidTx
	RejectSansityTx
	RejectSalaryTx
	RejectVersion
	RejectInvalidFee
	CanNotCheckDoubleSpend
)

var ErrCodeMessage = map[int]struct {
	code    int
	message string
}{
	RejectDuplicateTx:      {-1000, "Reject duplicate tx"},
	RejectInvalidTx:        {-1001, "Reject invalid tx"},
	RejectSansityTx:        {-1002, "Reject not sansity tx"},
	RejectSalaryTx:         {-1003, "Reject salary tx"},
	RejectInvalidFee:       {-1004, "Reject invalid fee"},
	RejectVersion:          {-1005, "Reject invalid version"},
	CanNotCheckDoubleSpend: {-1006, "Can not check double spend"},
}

type MempoolTxError struct {
	code    int    // The code to send with reject messages
	message string // Human readable message of the issue
	err     error
}

// Error satisfies the error interface and prints human-readable errors.
func (e MempoolTxError) Error() string {
	return fmt.Sprintf("%d: %s %+v", e.code, e.message, e.err)
}

// txRuleError creates an underlying MempoolTxError with the given a set of
// arguments and returns a RuleError that encapsulates it.
func (e *MempoolTxError) Init(key int, err error) {
	e.code = ErrCodeMessage[key].code
	e.message = ErrCodeMessage[key].message
	e.err = errors.Wrap(err, e.message)
}
