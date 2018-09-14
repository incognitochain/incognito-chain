package mempool

import (
	"github.com/ninjadotorg/cash-prototype/transaction"
	"github.com/ninjadotorg/cash-prototype/common"
)

// Policy houses the policy (configuration parameters) which is used to control the mempool.
type Policy struct {
	// MaxTxVersion is the transaction version that the mempool should
	// accept.  All transactions above this version are rejected as
	// non-standard.
	MaxTxVersion int8
}

func (self *Policy) CheckTxVersion(tx *transaction.Transaction) bool {
	txType := (*tx).GetType()
	switch txType {
	case common.TxNormalType:
		{
			temp := (*tx).(*transaction.Tx)
			if temp.Version > self.MaxTxVersion {
				return false
			}
		}
	case common.TxActionParamsType:
		{
			temp := (*tx).(*transaction.ActionParamTx)
			if temp.Version > self.MaxTxVersion {
				return false
			}
		}
	}
	return true
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
