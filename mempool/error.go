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
	RejectDuplicateStakePubkey
	RejectDuplicateInitTokenTx
	RejectVersion
	RejectInvalidTxType
	RejectDoubleSpendWithMempoolTx
	RejectDoubleSpendWithBlockchainTx
	RejectInvalidFee
	RejectInvalidSize
	CanNotCheckDoubleSpend
	DatabaseError
	MarshalError
	UnmarshalError
	ShardToBeaconBoolError
	DuplicateBlockError
	OldBlockError
	MaxPoolSizeError
	UnexpectedTransactionError
	TransactionNotFoundError
	RejectTestTransactionError
	WrongShardIDError
	HashError
	ReplacementError
)

var ErrCodeMessage = map[int]struct {
	Code    int
	Message string
}{
	RejectDuplicateTx:                 {-1000, "Reject duplicate tx"},
	RejectInvalidTx:                   {-1001, "Reject invalid tx"},
	RejectSansityTx:                   {-1002, "Reject not sansity tx"},
	RejectSalaryTx:                    {-1003, "Reject salary tx"},
	RejectInvalidFee:                  {-1004, "Reject invalid fee"},
	RejectVersion:                     {-1005, "Reject invalid version"},
	CanNotCheckDoubleSpend:            {-1006, "Can not check double spend"},
	DatabaseError:                     {-1007, "Database Error"},
	ShardToBeaconBoolError:            {-1007, "ShardToBeaconBool Error"},
	RejectDuplicateStakePubkey:        {-1008, "Reject Duplicate Stake Error"},
	DuplicateBlockError:               {-1009, "Duplicate Block Error"},
	OldBlockError:                     {-1010, "Old Block Error"},
	MaxPoolSizeError:                  {-1011, "Max Pool Size Error"},
	UnexpectedTransactionError:        {-1012, "Unexpected Transaction Error"},
	TransactionNotFoundError:          {-1013, "Transaction Not Found Error"},
	RejectTestTransactionError:        {-1014, "Reject Test Transaction Error"},
	RejectInvalidTxType:               {-1015, "Reject Invalid Tx Type"},
	RejectDoubleSpendWithMempoolTx:    {-1016, "Reject Double Spend With Other Tx in mempool"},
	RejectDoubleSpendWithBlockchainTx: {-1017, "Reject Double Spend With Current Blockchain"},
	WrongShardIDError:                 {-1018, "Reject Cross Shard Block With Same ShardID in Pool"},
	MarshalError:                      {-1019, "Marshal Error"},
	UnmarshalError:                    {-1020, "Unmarshal Error"},
	HashError:                    {-1021, "Hash Error"},
	ReplacementError:                    {-1022, "Replacement or Cancel Tx Error"},
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
func NewMempoolTxError(key int, err error) *MempoolTxError {
	return &MempoolTxError{
		Code:    ErrCodeMessage[key].Code,
		Message: ErrCodeMessage[key].Message,
		Err:     errors.Wrap(err, ErrCodeMessage[key].Message),
	}
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
