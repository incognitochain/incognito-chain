package mempool

import (
	"fmt"

	"github.com/pkg/errors"
)

const (
	RejectDuplicateTx = iota
	RejectInvalidTx
	RejectSansityTx
	RejectSalaryTx
	RejectDuplicateStakeTx
	RejectDuplicateInitTokenTx
	RejectVersion
	RejectInvalidTxType
	RejectDoubleSpendWithMempoolTx
	RejectInvalidFee
	RejectInvalidSize
	CanNotCheckDoubleSpend
	DatabaseError
	ShardToBeaconBoolError
	DuplicateBlockError
	OldBlockError
	MaxPoolSizeError
	UnexpectedTransactionError
	TransactionNotFoundError
	RejectTestTransactionError
)

var ErrCodeMessage = map[int]struct {
	Code    int
	Message string
}{
	RejectDuplicateTx:          {-1000, "Reject duplicate tx"},
	RejectInvalidTx:            {-1001, "Reject invalid tx"},
	RejectSansityTx:            {-1002, "Reject not sansity tx"},
	RejectSalaryTx:             {-1003, "Reject salary tx"},
	RejectInvalidFee:           {-1004, "Reject invalid fee"},
	RejectVersion:              {-1005, "Reject invalid version"},
	CanNotCheckDoubleSpend:     {-1006, "Can not check double spend"},
	DatabaseError:              {-1007, "Database Error"},
	ShardToBeaconBoolError:     {-1007, "ShardToBeaconBool Error"},
	RejectDuplicateStakeTx:     {-1008, "Reject Duplicate Stake Error"},
	DuplicateBlockError:        {-1009, "Duplicate Block Error"},
	OldBlockError:              {-1010, "Old Block Error"},
	MaxPoolSizeError:           {-1011, "Max Pool Size Error"},
	UnexpectedTransactionError: {-1012, "Unexpected Transaction Error"},
	TransactionNotFoundError:   {-1013, "Transaction Not Found Error"},
	RejectTestTransactionError: {-1014, "Reject Test Transaction Error"},
	RejectInvalidTxType: {-1015, "Reject Invalid Tx Type"},
	RejectDoubleSpendWithMempoolTx: {-1016, "Reject Double Spend With Other Tx in mempool"},
}

type MempoolTxError struct {
	Code    int    // The code to send with reject messages
	Message string // Human readable message of the issue
	Err     error
}

// Error satisfies the error interface and prints human-readable errors.
func (e MempoolTxError) Error() string {
	return fmt.Sprintf("%d: %s %+v", e.Code, e.Message, e.Err)
}

// txRuleError creates an underlying MempoolTxError with the given a set of
// arguments and returns a RuleError that encapsulates it.
func (e *MempoolTxError) Init(key int, err error) {
	e.Code = ErrCodeMessage[key].Code
	e.Message = ErrCodeMessage[key].Message
	e.Err = errors.Wrap(err, e.Message)
}

type BlockPoolError struct {
	Code    int
	Message string
	Err     error
}

func (e *BlockPoolError) Error() string {
	return fmt.Sprintf("%d: %s %+v", e.Code, e.Message, e.Err)
}

func (e *BlockPoolError) Init(key int, err error) {
	e.Code = ErrCodeMessage[key].Code
	e.Message = ErrCodeMessage[key].Message
	e.Err = errors.Wrap(err, e.Message)
}

func NewBlockPoolError(key int, err error) *BlockPoolError {
	return &BlockPoolError{
		Code:    ErrCodeMessage[key].Code,
		Message: ErrCodeMessage[key].Message,
		Err:     errors.Wrap(err, ErrCodeMessage[key].Message),
	}
}
