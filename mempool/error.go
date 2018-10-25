package mempool

const (
	RejectDuplicateTx      = "RejectDuplicateTx"
	RejectInvalidTx        = "RejectInvalidTx"
	RejectSansityTx        = "RejectSansityTx"
	RejectSalaryTx         = "RejectSalaryTx"
	RejectVersion          = "RejectVersion"
	RejectInvalidFee       = "RejectInvalidFee"
	CanNotCheckDoubleSpend = "CanNotCheckDoubleSpend"
)

var ErrCodeMessage = map[string]struct {
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
	code        int    // The code to send with reject messages
	description string // Human readable description of the issue
	err         error
}

// Error satisfies the error interface and prints human-readable errors.
func (e MempoolTxError) Error() string {
	return e.description
}

// txRuleError creates an underlying MempoolTxError with the given a set of
// arguments and returns a RuleError that encapsulates it.
func (e *MempoolTxError) Init(key string, err error) {
	e.code = ErrCodeMessage[key].code
	e.description = ErrCodeMessage[key].message
	e.err = err
}
