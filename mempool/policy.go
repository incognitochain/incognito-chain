package mempool

import (
	"github.com/ninjadotorg/cash-prototype/transaction"
)

// Policy houses the policy (configuration parameters) which is used to control the mempool.
type Policy struct {
	// MaxTxVersion is the transaction version that the mempool should
	// accept.  All transactions above this version are rejected as
	// non-standard.
	MaxTxVersion int32
}

// return min transacton fee required for a transaction that we accpted into the memmory pool and replayed.
func calcMinFeeTxAccepted(tx transaction.Tx) int64 {
	//@todo we will create rules of calc here later.
	return 1
}

// it make surce tx is validate standard (it is a "standard" stransaction input)
func checkValidateStandardOfTx(tx transaction.Tx) (bool, error) {
	//@todo we will create rules of calc here later.
	return true, nil
}

// it make surce tx is validate
func CheckValidateTx(tx transaction.Tx) bool {
	//@todo we will create rules of calc here later.
	return true
}
