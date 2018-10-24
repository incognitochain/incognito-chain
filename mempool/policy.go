package mempool

import (
	"fmt"

	"github.com/ninjadotorg/cash/common"
	"github.com/ninjadotorg/cash/transaction"
	"errors"
)

// Policy houses the policy (configuration parameters) which is used to control the mempool.
type Policy struct {
	// MaxTxVersion is the transaction version that the mempool should
	// accept.  All transactions above this version are rejected as
	// non-standard.
	MaxTxVersion int8
}

/*

 */
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
	case common.TxVotingType:
		{
			temp := (*tx).(*transaction.TxVoting)
			if temp.Version > self.MaxTxVersion {
				return false
			}
		}
	}
	return true
}

// return min transacton fee required for a transaction that we accepted into the memmory pool and replayed.
func (self *Policy) calcMinFeeTxAccepted(tx *transaction.Tx) uint64 {
	//@todo we will create rules of calc here later.
	return 0
}

// return min transacton fee required for a transaction that we accepted into the memmory pool and replayed.
func (self *Policy) calcMinFeeVotingTxAccepted(tx *transaction.TxVoting) uint64 {
	//@todo we will create rules of calc here later.
	return 0
}

/*

 */
func (self *Policy) CheckTransactionFee(tx *transaction.Tx) error {
	minFee := self.calcMinFeeTxAccepted(tx)
	if tx.Fee < minFee {
		str := fmt.Sprintf("transaction %+v has %d fees which is under the required amount of %d", tx.Hash().String(), tx.Fee, minFee)
		err := MempoolTxError{}
		err.Init(RejectInvalidFee, errors.New(str))
		return err
	}
	return nil
}

func (self *Policy) CheckVotingTransactionFee(tx *transaction.TxVoting) error {
	minFee := self.calcMinFeeVotingTxAccepted(tx)
	if tx.Fee < minFee {
		str := fmt.Sprintf("transaction %+v has %d fees which is under "+
			"the required amount of %d", tx.Hash().String(), tx.Fee,
			minFee)
		err := MempoolTxError{}
		err.Init(RejectInvalidFee, errors.New(str))
		return err
	}
	return nil
}
