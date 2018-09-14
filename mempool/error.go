package mempool

// TxRuleError identifies a rule violation.  It is used to indicate that
// processing of a transaction failed due to one of the many validation
// rules.  The caller can use type assertions to determine if a failure was
// specifically due to a rule violation and access the ErrorCode field to
// ascertain the specific reason for the rule violation.
type TxRuleError struct {
	rejectCode  TxErrCode // The code to send with reject messages
	description string    // Human readable description of the issue
}

// Error satisfies the error interface and prints human-readable errors.
func (e TxRuleError) Error() string {
	return e.description
}

// txRuleError creates an underlying TxRuleError with the given a set of
// arguments and returns a RuleError that encapsulates it.
func (e *TxRuleError) Init(code TxErrCode, desc string) {
	e.rejectCode = code
	e.description = desc
}

// rejectCode represents a numeric value by which a remote peer indicates
// why a message was rejected.
type TxErrCode uint8

// These constants define the various supported reject codes.
const (
	RejectDuplicateTx TxErrCode = 1
	RejectInvalidTx   TxErrCode = 2
	SansityInvalidTx  TxErrCode = 3
	RejectCoinbaseTx  TxErrCode = 4
)
